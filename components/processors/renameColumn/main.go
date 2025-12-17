package renameColumn

import (
	"encoding/json"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
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
func ProcessorCreator() (string, procrssor.Processor, []params.Params) {
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
func (p *Processor) Open(config map[string]string) error {
	mappingVal, ok := config["mapping"]
	if !ok {
		return fmt.Errorf("renameColumn processor: config is missing required key 'mapping'")
	}

	// YAML 解析器会将 map 解析为 map[interface{}]interface{}，这里需要进行类型转换。
	var mapping map[string]string
	err := json.Unmarshal([]byte(mappingVal), &mapping)
	if err != nil {
		return fmt.Errorf("renameColumn processor: 'mapping' must be a map (key-value pairs)")
	}

	if len(mapping) == 0 {
		return fmt.Errorf("renameColumn processor: 'mapping' cannot be empty")
	}

	p.mapping = mapping

	return nil
}

// Process 对记录进行处理，根据映射重命名指定的列。
//
// 如果在映射中指定的旧列名不存在于记录中，它将被静默忽略。
//
// **[CRITICAL] 警告：命名冲突与数据丢失风险**
// 当前实现通过遍历输入记录来构建一个新记录。如果一个重命名操作（例如 `A` -> `B`）
// 与一个已存在且未被重命名的列（`B`）发生冲突，那么输出记录中 `B` 的最终值将
// 取决于 Go 语言不确定的 map 迭代顺序。这可能导致不可预测的数据丢失。
//
// **强烈建议**：确保配置的 [mapping](file:///Users/szy/Desktop/code/etl-go/components/processors/renameColumn/main.go#L19-L19) 中，新列名不会与任何未被重命名的现有列名冲突。
func (p *Processor) Process(r record.Record) (record.Record, error) {
	newRecord := make(record.Record, len(r))

	for oldKey, value := range r {
		// 检查这个 key 是否在我们的重命名映射中。
		if newKey, ok := p.mapping[oldKey]; ok {
			// 如果是，使用新的 key。
			newRecord[newKey] = value
		} else {
			// 如果不是，则保留原样。
			newRecord[oldKey] = value
		}
	}

	return newRecord, nil
}

// Close 是一个无操作（no-op）方法，因为 renameColumn 处理器是无状态的，不需要在处理结束后清理任何资源。
func (p *Processor) Close() error {
	return nil
}

func (p *Processor) HandleColumns(columns *map[string]string) {
	for k := range *columns {
		if newKey, ok := p.mapping[k]; ok {
			(*columns)[newKey] = newKey
			delete(*columns, k)
		}
	}
}
