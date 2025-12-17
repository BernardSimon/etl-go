package csvSink

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
)

// Sink 实现了 core.Sink 接口，用于将数据以 CSV 格式写入文件。
type Sink struct {
	ID       string
	filePath string // 输出文件路径。
	file     *os.File
	writer   *csv.Writer // CSV 写入器。
	header   []string    // 可选的表头字段。
	written  bool        // 是否已经写入过数据（防止重复写header）
}

// NewSink 是 csv.Sink 的构造函数，由工厂调用。
func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return "csv", &Sink{}, nil, []params.Params{
		{
			Key:         "file_name",
			Description: "The name of the output file",
			Required:    true,
		},
		{
			Key:          "file_ext",
			Description:  "The extension of the output file",
			DefaultValue: "csv",
			Required:     true,
		},
	}
}

func (s *Sink) Open(config map[string]string, columnMapping map[string]string, _ *datasource.Datasource) error {
	filePath, ok := config["file_path"]
	if !ok {
		return fmt.Errorf("csv sink: config is missing or has invalid 'file_name'")
	}
	s.filePath = filePath
	var err error
	s.file, err = os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("csv sink: failed to create/open file: %w", err)
	}
	s.writer = csv.NewWriter(s.file)

	for _, key := range columnMapping {
		s.header = append(s.header, key)
	}

	return nil
}

// Write 将一批记录以 CSV 行的形式写入文件。
func (s *Sink) Write(ID string, records []record.Record) error {
	s.ID = ID
	if len(records) == 0 {
		return nil
	}

	if s.writer == nil {
		return fmt.Errorf("csv sink: writer is not initialized")
	}

	// 第一次写入时先写 header（如果有）
	if !s.written && len(s.header) > 0 {
		if err := s.writer.Write(s.header); err != nil {
			return fmt.Errorf("csv sink: failed to write header: %w", err)
		}
		s.written = true
	}

	for _, r := range records {
		// 假设每个 record 都是 map[string]interface{} 或可转换为字符串数组
		row := make([]string, 0, len(r))
		for _, key := range s.header {
			value := ""
			if v, ok := r[key]; ok {
				value = fmt.Sprintf("%v", v)
			}
			row = append(row, value)
		}
		if err := s.writer.Write(row); err != nil {
			return fmt.Errorf("csv sink: failed to write row: %w", err)
		}
	}

	s.writer.Flush() // 每批数据后刷新缓冲区
	if err := s.writer.Error(); err != nil {
		return fmt.Errorf("csv sink: flush error: %w", err)
	}

	return nil
}

// Close 负责关闭文件句柄并保存元信息。
func (s *Sink) Close() error {
	if s.writer != nil {
		s.writer.Flush()
	}
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
