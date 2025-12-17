package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/BernardSimon/etl-go/server/utils/file"

	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

// ProcessorConfig 定义了单个处理器的配置，供引擎的 Run 方法消费。
type ProcessorConfig struct {
	Type   string            `yaml:"type"`
	Params map[string]string `yaml:"params"`
}

// Config 包含了对管道性能进行微调的参数。
// BatchSize 控制了 Sink 批量写入的大小，增大此值可提高写入吞吐量，但会增加延迟和内存消耗。
// ChannelSize 定义了连接各阶段的通道缓冲区大小，更大的缓冲区可以减少阶段间的等待，但同样会增加内存占用。
type Config struct {
	BatchSize   int `yaml:"batch_size"`
	ChannelSize int `yaml:"channel_size"`
}

const (
	defaultBatchSize   = 1000
	defaultChannelSize = 10000
)

// Engine 是 ETL 管道的并发编排器。
// 它将 Source、Processors 和 Sink 组装成一个基于 Goroutine 和 Channel 的流水线，
// 并负责管理整个流水线的生命周期，包括启动、优雅关闭和错误传播。
type Engine struct {
	id                       string
	beforeExecutor           executor.Executor
	beforeExecutorDatasource *datasource.Datasource
	source                   source.Source
	sourceDatasource         *datasource.Datasource
	processors               []procrssor.Processor
	sink                     sink.Sink
	sinkDatasource           *datasource.Datasource
	afterExecutor            executor.Executor
	afterExecutorDatasource  *datasource.Datasource
	batchSize                int
	channelSize              int
	cancel                   context.CancelFunc
	wg                       sync.WaitGroup
}

// NewEngine 创建一个新的管道引擎实例。

func NewEngine(id string, beforeExecutor *executor.Executor, beforeExecutorDatasource *datasource.Datasource, source source.Source, sourceDatasource *datasource.Datasource, processors []procrssor.Processor, sink sink.Sink, sinkDatasource *datasource.Datasource, config Config, afterExecute *executor.Executor, afterExecuteDatasource *datasource.Datasource) *Engine {
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	channelSize := config.ChannelSize
	if channelSize <= 0 {
		channelSize = defaultChannelSize
	}
	zap.L().Info(fmt.Sprintf("Channel Config: BatchSize=%d, ChannelSize=%d", batchSize, channelSize), zap.String("service", "etl"), zap.String("name", id))
	engine := Engine{
		id:               id,
		source:           source,
		sourceDatasource: sourceDatasource,
		processors:       processors,
		sink:             sink,
		sinkDatasource:   sinkDatasource,
		batchSize:        batchSize,
		channelSize:      channelSize,
	}
	if beforeExecutor != nil {
		engine.beforeExecutor = *beforeExecutor
		engine.beforeExecutorDatasource = beforeExecutorDatasource
	}
	if afterExecute != nil {
		engine.afterExecutor = *afterExecute
		engine.afterExecutorDatasource = afterExecuteDatasource
	}
	return &engine
}

// Run 动态构建并启动整个并发 ETL 流水线。
func (e *Engine) Run(id string, ctx context.Context, beforeExecuteConfig *map[string]string, sourceConfig map[string]string, processorConfigs []ProcessorConfig, sinkConfig map[string]string, afterExecuteConfig *map[string]string) (err error) {
	// 1. 创建一个可取消的上下文，用于实现“一处失败，全体取消”的快速失败机制。
	runCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	defer e.cancel() // 确保在函数退出时，所有 goroutine 都能收到停止信号
	// 在函数返回前，确保所有组件的 Close() 方法都被调用。
	// 使用命名返回值和 defer 来优雅地处理组件的关闭逻辑。
	var fileIds []string
	// 处理保留config参数
	var fileId string
	fileId, err = HandleInternalConfig(beforeExecuteConfig)
	if err != nil {
		zap.L().Error("Failed to handle internal config", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to handle internal config: %w", err)
	}
	if fileId != "" {
		fileIds = append(fileIds, fileId)
	}
	fileId, err = HandleInternalConfig(afterExecuteConfig)
	if err != nil {
		zap.L().Error("Failed to handle internal config", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to handle internal config: %w", err)
	}
	if fileId != "" {
		fileIds = append(fileIds, fileId)
	}
	fileId, err = HandleInternalConfig(&sourceConfig)
	if err != nil {
		zap.L().Error("Failed to handle internal config", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to handle internal config: %w", err)
	}
	if fileId != "" {
		fileIds = append(fileIds, fileId)
	}
	fileId, err = HandleInternalConfig(&sinkConfig)
	if err != nil {
		zap.L().Error("Failed to handle internal config", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to handle internal config: %w", err)
	}
	if fileId != "" {
		fileIds = append(fileIds, fileId)
	}
	for i := range processorConfigs {
		fileId, err = HandleInternalConfig(&processorConfigs[i].Params)
		if err != nil {
			zap.L().Error("Failed to handle internal config", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			return fmt.Errorf("pipeline: failed to handle internal config: %w", err)
		}
		if fileId != "" {
			fileIds = append(fileIds, fileId)
		}
	}

	defer func() {
		zap.L().Info("Closing (Sink)...", zap.String("service", "etl"), zap.String("name", id))
		if closeErr := e.sink.Close(); closeErr != nil && err == nil {
			zap.L().Error("Failed to close Sink", zap.Error(closeErr), zap.String("service", "etl"), zap.String("name", id))
			err = errors.Join(err, fmt.Errorf("pipeline: failed to close sink: %w", closeErr))
		}
		for i := len(e.processors) - 1; i >= 0; i-- {
			zap.L().Info("Closing (Processor) #"+strconv.Itoa(i+1)+" ("+processorConfigs[i].Type+")...", zap.String("service", "etl"), zap.String("name", id))
			if closeErr := e.processors[i].Close(); closeErr != nil && err == nil {
				zap.L().Error("Failed to close Processor", zap.Error(closeErr), zap.String("service", "etl"), zap.String("name", id))
				err = errors.Join(err, fmt.Errorf("pipeline: failed to close processor #%d (%s): %w", i+1, processorConfigs[i].Type, closeErr))
			}
		}
		zap.L().Info("Closing (Source)...", zap.String("service", "etl"), zap.String("name", id))
		if closeErr := e.source.Close(); closeErr != nil && err == nil {
			zap.L().Error("Failed to close Source", zap.Error(closeErr), zap.String("service", "etl"), zap.String("name", id))
			err = errors.Join(err, fmt.Errorf("pipeline: failed to close source: %w", closeErr))
		}
		if len(fileIds) > 0 {
			zap.L().Info("Saving output file...", zap.String("service", "etl"), zap.String("name", id))
			isError := false
			if err != nil {
				isError = true
			}
			fileSaveErr := file.SaveOutputFile(id, fileIds, isError)
			if fileSaveErr != nil {
				zap.L().Error("Failed to save output file", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
				err = errors.Join(err, fmt.Errorf("pipeline: failed to save output file: %w", fileSaveErr))
			}
		}

	}()

	// 2. 按顺序打开所有组件，这是运行前的准备和验证阶段。
	if beforeExecuteConfig != nil {
		zap.L().Info("Opening (Before Executor)...", zap.String("service", "etl"), zap.String("name", id))
		if err = e.beforeExecutor.Open(*beforeExecuteConfig, e.beforeExecutorDatasource); err != nil {
			zap.L().Error("Failed to open Before Executor", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			return fmt.Errorf("pipeline: failed to open before executor: %w", err)
		} else {
			zap.L().Info("Before Executor Doing Success", zap.String("service", "etl"), zap.String("name", id))
		}
		zap.L().Info("Closing (Before Executor)...", zap.String("service", "etl"), zap.String("name", id))
		if err := e.beforeExecutor.Close(); err != nil {
			zap.L().Error("Failed to close Before Executor", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			return fmt.Errorf("pipeline: failed to close before executor: %w", err)
		}
	}
	zap.L().Info("正在打开数据源 (Source)...", zap.String("service", "etl"), zap.String("name", id))
	if err := e.source.Open(sourceConfig, e.sourceDatasource); err != nil {
		zap.L().Error("数据源打开失败", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to open source: %w", err)
	}
	column := e.source.Column()

	for i, p := range e.processors {
		zap.L().Info("正在打开处理器 (Processor) #"+strconv.Itoa(i+1)+" ("+processorConfigs[i].Type+")...", zap.String("service", "etl"), zap.String("name", id))
		p.HandleColumns(&column)
		if err := p.Open(processorConfigs[i].Params); err != nil {
			zap.L().Error("处理器打开失败", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			return fmt.Errorf("pipeline: failed to open processor #%d (%s): %w", i+1, processorConfigs[i].Type, err)
		}
	}
	zap.L().Info("正在打开数据汇 (Sink)...", zap.String("service", "etl"), zap.String("name", id))
	if err := e.sink.Open(sinkConfig, column, e.sinkDatasource); err != nil {
		zap.L().Error("数据汇打开失败", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
		return fmt.Errorf("pipeline: failed to open sink: %w", err)
	}

	// 3. 动态创建一系列 Channel，作为连接各个并发阶段的“传送带”。
	numWorkers := len(e.processors) + 2 // Source + Processors + Sink
	errChan := make(chan error, numWorkers)
	numChan := len(e.processors) + 1
	dataChan := make([]chan record.Record, numChan)
	for i := 0; i < numChan; i++ {
		dataChan[i] = make(chan record.Record, e.channelSize)
	}

	// 4. 为每个组件启动一个专属的 goroutine，并将它们通过 Channel 连接起来，形成流水线。
	e.wg.Add(numWorkers)
	go e.runSource(id, runCtx, dataChan[0], errChan)
	for i := range e.processors {
		go e.runProcessor(id, runCtx, e.processors[i], dataChan[i], dataChan[i+1], errChan, i+1, processorConfigs[i].Type)
	}
	go e.runSink(id, runCtx, dataChan[len(dataChan)-1], errChan)

	// 5. 等待所有 goroutine 执行结束。
	e.wg.Wait()
	close(errChan)

	// 6. 收集在运行过程中可能发生的任何错误，并统一返回。
	var finalErr error
	for runErr := range errChan {
		if finalErr == nil {
			finalErr = runErr
		} else {
			finalErr = fmt.Errorf("%v; %w", finalErr, runErr)
		}
	}
	if finalErr != nil {
		zap.L().Error("数据处理失败", zap.Error(finalErr), zap.String("service", "etl"), zap.String("name", id))
		return finalErr
	}
	if afterExecuteConfig != nil {
		zap.L().Info("正在打开后处理器 (Executor)...", zap.String("service", "etl"), zap.String("name", id))
		if err = e.afterExecutor.Open(*afterExecuteConfig, e.afterExecutorDatasource); err != nil {
			zap.L().Error("后置处理器执行失败", zap.Error(finalErr), zap.String("service", "etl"), zap.String("name", id))
			return fmt.Errorf("pipeline: failed to open after executor: %w", err)
		}
		zap.L().Info("后置处理器执行成功", zap.String("service", "etl"), zap.String("name", id))
		zap.L().Info("正在关闭后处理器 (Executor)...", zap.String("service", "etl"), zap.String("name", id))
		if err := e.afterExecutor.Close(); err != nil {
			zap.L().Error("后置处理器关闭失败", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			err = fmt.Errorf("pipeline: failed to close after executor: %w", err)
			return err
		}
	}
	return nil
}

// runSource 是 Source 的工作协程，负责从数据源读取数据并送入第一个通道。
func (e *Engine) runSource(id string, ctx context.Context, outChan chan<- record.Record, errChan chan<- error) {
	defer e.wg.Done()
	defer close(outChan) // 读取完成后关闭输出通道，这是通知下游数据已耗尽的关键信号。
	zap.L().Info("正在从数据源读取数据...", zap.String("service", "etl"), zap.String("name", id))

	for {
		// 优先检查上下文是否已被取消，实现快速失败。
		select {
		case <-ctx.Done():
			zap.L().Warn("Source worker 检测到取消信号，正在停止...", zap.String("service", "etl"), zap.String("name", id))
			return
		default:
			// 非阻塞地继续执行
		}

		readRecord, err := e.source.Read()
		if err != nil {
			if err == io.EOF {
				zap.L().Info("Source 已成功读取所有数据", zap.String("service", "etl"), zap.String("name", id))
				return // 数据流正常结束
			}
			// 发生不可恢复的读取错误
			zap.L().Error("Source 读取数据时发生错误", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			errChan <- fmt.Errorf("source error: %w", err)
			e.cancel() // 发生错误，立即取消所有其他 goroutine
			return
		}

		// 将记录发送到输出通道，同时监听取消信号。
		select {
		case outChan <- readRecord:
		case <-ctx.Done():
			logrus.Warn("Source worker 在发送数据时收到取消信号，正在停止...")
			return
		}
	}
}

// runProcessor 是流水线上的一个工作站，负责执行单个处理逻辑。
func (e *Engine) runProcessor(id string, ctx context.Context, p procrssor.Processor, inChan <-chan record.Record, outChan chan<- record.Record, errChan chan<- error, num int, pType string) {
	defer e.wg.Done()
	defer close(outChan) // 当前阶段处理完毕，关闭自己的输出通道，以通知下一阶段。
	zap.L().Info("正在启动处理器 (Processor) #"+strconv.Itoa(num)+" ("+pType+")...", zap.String("service", "etl"), zap.String("name", id))

	// for-range 会自动处理通道的关闭，是消费通道数据的优雅方式。
	for chanRecord := range inChan {
		// 每次循环开始时，都检查是否需要提前退出。
		select {
		case <-ctx.Done():
			zap.L().Warn(fmt.Sprintf("Processor #%d (%s) worker 收到取消信号，正在停止...", num, pType), zap.String("service", "etl"), zap.String("name", id))
			return
		default:
			processedRecord, err := p.Process(chanRecord)
			if err != nil {
				zap.L().Error(fmt.Sprintf("Processor #%d (%s) 处理记录时发生错误: %v", num, pType, err), zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
				errChan <- fmt.Errorf("processor #%d (%s) error: %w", num, pType, err)
				e.cancel()
				return
			}

			// 如果处理器返回 nil, 意味着该记录被过滤，我们通过 continue 跳过它。
			if processedRecord == nil {
				continue
			}

			select {
			case outChan <- processedRecord:
			case <-ctx.Done():
				zap.L().Warn(fmt.Sprintf("Processor #%d (%s) worker 在发送数据时收到取消信号，正在停止...", num, pType), zap.String("service", "etl"), zap.String("name", id))
				return
			}
		}
	}
	zap.L().Info("处理器 (Processor) #"+strconv.Itoa(num)+" ("+pType+") 启动成功", zap.String("service", "etl"), zap.String("name", id))
}

// runSink 是 Sink 的工作协程，负责从最后一个通道接收数据并批量写入目的地。
func (e *Engine) runSink(id string, ctx context.Context, inChan <-chan record.Record, errChan chan<- error) {
	defer e.wg.Done()
	zap.L().Info("正在启动 Sink...", zap.String("service", "etl"), zap.String("name", id))
	batch := make([]record.Record, 0, e.batchSize)

	for chanRecord := range inChan {
		// 同样，优先检查取消信号。
		select {
		case <-ctx.Done():
			zap.L().Warn("Sink worker 检测到取消信号，正在停止...", zap.String("service", "etl"), zap.String("name", id))
			return
		default:
			batch = append(batch, chanRecord)
			if len(batch) >= e.batchSize {
				zap.L().Info(fmt.Sprintf("正在刷入一批 %d 条记录...", len(batch)), zap.String("service", "etl"), zap.String("name", id))
				if err := e.flush(batch); err != nil {
					zap.L().Error("Sink 刷入批次时发生错误", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
					errChan <- fmt.Errorf("sink error: %w", err)
					e.cancel()
					// 注意：这里在出错后没有立即 return，是为了让循环自然结束，
					// 从而可以继续处理 defer 和最后的 flush。
					// 但因为 cancel() 被调用，其他 goroutine 会快速退出。
				}
				batch = make([]record.Record, 0, e.batchSize) // 重置批次
			}
		}
	}

	// 注意：循环结束后，必须处理最后一批可能不足一个 batchSize 的数据，否则会造成数据丢失。
	if len(batch) > 0 {
		zap.L().Info(fmt.Sprintf("正在刷入最后 %d 条记录...", len(batch)), zap.String("service", "etl"), zap.String("name", id))
		if err := e.flush(batch); err != nil {
			zap.L().Error("Sink 刷入最后批次时发生错误", zap.Error(err), zap.String("service", "etl"), zap.String("name", id))
			errChan <- fmt.Errorf("sink error on final flush: %w", err)
		}
	}
	zap.L().Info("Sink worker 输入通道已关闭，正常退出", zap.String("service", "etl"), zap.String("name", id))
}

// flush 将一个批次的数据写入 sink。
func (e *Engine) flush(batch []record.Record) error {
	if len(batch) == 0 {
		return nil
	}
	return e.sink.Write(e.id, batch)
}

func HandleInternalConfig(config *map[string]string) (string, error) {
	if config == nil {
		return "", nil
	}
	var fileId = ""
	for k, v := range *config {
		switch k {
		case "file_id":
			filePath, err := file.GetFilePath(v)
			if err != nil {
				return "", fmt.Errorf("file_id config is invalid: %w", k)
			}
			(*config)["file_path"] = filePath
			continue

		case "file_ids":
			fileIds := strings.Split(v, ",")
			if len(fileIds) == 0 {
				return "", fmt.Errorf("file_ids config is invalid: %w", k)
			}
			filePaths := make([]string, len(fileIds))
			for i, fileId := range fileIds {
				filePath, err := file.GetFilePath(fileId)
				if err != nil {
					return "", fmt.Errorf("file_ids config is invalid: %w", k)
				}
				filePaths[i] = filePath
			}
			(*config)["file_paths"] = strings.Join(filePaths, ",")
			continue
		case "file_name":
			fileExt, ok := (*config)["file_ext"]
			if !ok {
				fileExt = ""
			}
			id, filePath, err := file.CreateOutputFile(v, fileExt)
			if err != nil {
				return "", fmt.Errorf("file_name config is invalid: %w", k)
			}
			(*config)["file_path"] = filePath
			fileId = id
			continue
		default:
			continue
		}
	}
	return fileId, nil
}
