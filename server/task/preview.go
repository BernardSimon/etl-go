package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/processor"
	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	"github.com/BernardSimon/etl-go/server/config"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/types"
)

const previewLimit = 20

// PreviewTask 用任务的当前配置跑一次预览：
//   - SQL source：自动在查询外层包 LIMIT，避免全表扫描
//   - 所有 sink 替换为内存 PreviewSink，不向任何外部系统写入
//   - 返回前 N 行数据及列结构，供前端展示
func PreviewTask(missionID string) (*types.PreviewResponse, error) {
	var mission model.Task
	if err := model.DB.Where("id = ?", missionID).First(&mission).Error; err != nil {
		return nil, fmt.Errorf("task not found")
	}
	if mission.Data == nil {
		return nil, fmt.Errorf("task has no configuration")
	}

	// 变量替换（与正式运行保持一致）
	replacedData, err := ReplaceVariables(mission.Data)
	if err != nil {
		return nil, fmt.Errorf("variable replacement failed: %w", err)
	}
	if replacedData != nil {
		mission.Data = replacedData
	}

	cfg := pipeline.Config{
		BatchSize:   config.Config.Pipeline.BatchSize,
		ChannelSize: config.Config.Pipeline.ChannelSize,
	}

	dsResolver := newDatasourceResolver()

	// ── Source ──────────────────────────────────────────────────────────────
	sourceStore, err := factory.CreateSource(mission.Data.Source.Type)
	if err != nil {
		return nil, fmt.Errorf("invalid source type: %w", err)
	}
	sourceConfig := buildConfig(mission.Data.Source.Params)
	injectPreviewLimit(sourceConfig, mission.Data.Source.Type)

	var sourceDatasource datasource.Datasource
	if sourceStore.Datasource != nil {
		ds, err := dsResolver.initDatasource(mission.Data.Source.DataSource, *sourceStore.Datasource)
		if err != nil {
			return nil, fmt.Errorf("failed to init source datasource: %w", err)
		}
		sourceDatasource = ds
	}

	// ── Processors ──────────────────────────────────────────────────────────
	processors := make([]processor.Processor, 0, len(mission.Data.Processors))
	processorConfigs := make([]pipeline.ProcessorConfig, 0, len(mission.Data.Processors))
	for _, pConfig := range mission.Data.Processors {
		p, err := factory.CreateProcessor(pConfig.Type)
		if err != nil {
			return nil, fmt.Errorf("invalid processor type %s: %w", pConfig.Type, err)
		}
		processors = append(processors, p.Handle)
		processorConfigs = append(processorConfigs, pipeline.ProcessorConfig{
			Type:   pConfig.Type,
			Params: buildConfig(pConfig.Params),
		})
	}

	// ── PreviewSink（替换真实 sink）──────────────────────────────────────────
	previewSink := &pipeline.PreviewSink{Limit: previewLimit}

	engine := pipeline.NewEngine(
		"preview-"+missionID,
		nil, nil,
		sourceStore.Handle, sourceDatasource,
		processors,
		previewSink, nil,
		cfg,
		nil, nil,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := engine.Run("preview-"+missionID, ctx, nil, sourceConfig, processorConfigs, map[string]string{}, nil); err != nil {
		return nil, fmt.Errorf("preview execution failed: %w", err)
	}

	return buildPreviewResponse(previewSink), nil
}

// injectPreviewLimit 对 SQL 类 source 在 query 外层包一个子查询加 LIMIT，
// 避免全表扫描。非 SQL source 不做处理（PreviewSink 自身会限制记录数）。
func injectPreviewLimit(config map[string]string, sourceType string) {
	sqlTypes := map[string]bool{"mysql": true, "postgre": true, "sqlite": true, "doris": true}
	if !sqlTypes[strings.ToLower(sourceType)] {
		return
	}
	if q, ok := config["query"]; ok && q != "" {
		config["query"] = fmt.Sprintf(
			"SELECT * FROM (%s) AS _etl_preview LIMIT %d",
			q, previewLimit,
		)
	}
}

// buildPreviewResponse 将 PreviewSink 收集到的记录转换为前端友好的格式
func buildPreviewResponse(s *pipeline.PreviewSink) *types.PreviewResponse {
	resp := &types.PreviewResponse{
		Columns: s.Columns,
		Rows:    make([]map[string]interface{}, 0, len(s.Records)),
	}
	for _, r := range s.Records {
		row := make(map[string]interface{}, len(r))
		for k, v := range r {
			row[k] = v
		}
		resp.Rows = append(resp.Rows, row)
	}
	return resp
}
