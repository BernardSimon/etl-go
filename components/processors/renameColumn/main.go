package renameColumn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/processor"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

var name = "renameColumn"

func SetCustomName(customName string) {
	name = customName
}

// Processor 实现了 core.Processor 接口，用于重命名记录中的一个或多个列。
// 这个处理器常用于统一来自不同数据源的列名，或使列名更具可读性。
type Processor struct {
	mapping map[string]string // 存储列名映射关系，键是旧列名，值是新列名。
}

// ProcessorCreator 返回处理器名称、实例和参数定义
func ProcessorCreator() (string, processor.Processor, []params.Params) {
	return name, &Processor{}, []params.Params{
		{
			Key:          "mapping",
			Required:     true,
			DefaultValue: "",
			Description:  "column mapping from old name to new name",
		},
	}
}

// Open 从配置中解析列的重命名映射。
//
// 它期望配置中有一个名为mapping的键，其值为一个从旧列名到新列名的映射。
func (p *Processor) Open(ctx context.Context, config map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	mappingVal, ok := config["mapping"]
	if !ok {
		return fmt.Errorf("renameColumn processor: config is missing required key 'mapping'")
	}

	var mapping map[string]string
	err := json.Unmarshal([]byte(mappingVal), &mapping)
	if err != nil {
		return fmt.Errorf("renameColumn processor: 'mapping' must be a map (key-value pairs)")
	}

	if len(mapping) == 0 {
		return fmt.Errorf("renameColumn processor: 'mapping' cannot be empty")
	}

	cleanedMapping := make(map[string]string)
	for k, v := range mapping {
		cleanedMapping[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	p.mapping = cleanedMapping

	return nil
}

// Process 对记录进行处理，根据映射重命名指定的列。
//
// 如果在映射中指定的旧列名不存在于记录中，它将被静默忽略。
//
// Process 采用两阶段写入，保证行为确定性：
//  1. 第一阶段：将所有不在 mapping 中的列（去 BOM 后）写入新记录。
//  2. 第二阶段：将 mapping 中的列执行重命名后写入，若新列名与已有列冲突，重命名结果优先（覆盖）。
//
// 注意：若配置了重命名 A→B，且 B 列本身也存在于原记录中，B 的原始值将被丢弃。
// 请确保新列名不与其他未被重命名的列冲突。
func (p *Processor) Process(ctx context.Context, r record.Record) (record.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	newRecord := make(record.Record, len(r))

	// 第一阶段：复制不需要重命名的列（同时去除 BOM）
	for oldKey, value := range r {
		cleanOldKey := strings.TrimLeft(strings.TrimSpace(oldKey), "\uFEFF")
		if _, willBeRenamed := p.mapping[cleanOldKey]; !willBeRenamed {
			newRecord[cleanOldKey] = value
		}
	}

	// 第二阶段：执行重命名，rename 结果优先
	for oldKey, value := range r {
		cleanOldKey := strings.TrimLeft(strings.TrimSpace(oldKey), "\uFEFF")
		if newKey, ok := p.mapping[cleanOldKey]; ok {
			newRecord[newKey] = value
		}
	}

	return newRecord, nil
}

// ... existing code ...

// Close 是一个无操作（no-op）方法，因为 renameColumn 处理器是无状态的，不需要在处理结束后清理任何资源。
func (p *Processor) Close() error {
	return nil
}

// HandleColumns 与 Process 保持相同的两阶段逻辑，确保传递给 Sink 的列 schema 与运行时记录完全一致。
// 同时去除所有列名的 BOM，避免 BOM 传播到 Sink 导致 SQL 列名错误。
func (p *Processor) HandleColumns(columns *map[string]string) {
	newColumns := make(map[string]string, len(*columns))
	mapping := p.mapping
	// 第一阶段：不需要重命名的列（去 BOM）
	for k := range *columns {
		cleanKey := strings.TrimLeft(strings.TrimSpace(k), "\uFEFF")
		if _, willBeRenamed := mapping[cleanKey]; !willBeRenamed {
			newColumns[cleanKey] = cleanKey
		}
	}
	// 第二阶段：重命名的列
	for k := range *columns {
		cleanKey := strings.TrimLeft(strings.TrimSpace(k), "\uFEFF")
		if newKey, ok := mapping[cleanKey]; ok {
			newColumns[newKey] = newKey
		}
	}
	*columns = newColumns
}
