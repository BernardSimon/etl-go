package selectColumns

import (
	"encoding/json"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

var name = "selectColumns"

func SetCustomName(customName string) {
	name = customName
}

// Processor 实现了 core.Processor 接口，用于从每条记录中筛选出指定的列，并丢弃所有其他列。
// 这是一个典型的"白名单"模式，常用于减少数据体积、移除不必要的字段，或在数据发送到外部系统前进行清理。
type Processor struct {
	columnsToKeep []string // 需要保留的列名列表。
}

// ProcessorCreator 返回处理器名称、实例和参数定义
func ProcessorCreator() (string, procrssor.Processor, []params.Params) {
	return name, &Processor{}, []params.Params{
		{
			Key:          "columns",
			Required:     true,
			DefaultValue: "",
			Description:  "columns to keep as [array]",
		},
	}
}

// Open 从配置中解析需要保留的列名列表。
func (p *Processor) Open(config map[string]string) error {
	columnsVal, ok := config["columns"]
	if !ok {
		return fmt.Errorf("selectColumns processor: config is missing required key 'columns'")
	}

	// YAML 解析器会将字符串数组解析为 []interface{}，这里需要进行类型断言和转换。
	var columns []string
	err := json.Unmarshal([]byte(columnsVal), &columns)
	if err != nil {
		return fmt.Errorf("selectColumns processor: 'columns' must be an array of strings")
	}
	if len(columns) == 0 {
		return fmt.Errorf("selectColumns processor: 'columns' array cannot be empty")
	}

	p.columnsToKeep = columns

	return nil
}

// Process 方法根据 [columnsToKeep](file:///Users/szy/Desktop/code/etl-go/components/processors/selectColumns/main.go#L19-L19) 列表对输入的记录进行转换，生成一个只包含指定列的新记录。
//
// 一个关键的设计决策是：如果在输入记录中找不到配置中指定的某个列名，它会被静默忽略，而不会报错。
// 这使得管道对上游数据模式的微小变化（例如，某列被移除）更具弹性。
func (p *Processor) Process(r record.Record) (record.Record, error) {
	newRecord := make(record.Record, len(p.columnsToKeep))

	for _, colName := range p.columnsToKeep {
		// 检查原始记录中是否存在该列。
		if val, ok := r[colName]; ok {
			// 如果存在，则将其添加到新记录中。
			newRecord[colName] = val
		}
	}

	return newRecord, nil
}

// Close 是一个无操作（no-op）方法，因为 selectColumns 处理器是无状态的，不需要在处理结束后清理任何资源。
func (p *Processor) Close() error {
	return nil
}

func (p *Processor) HandleColumns(columns *map[string]string) {
	for k := range *columns {
		match := false
		for _, colName := range p.columnsToKeep {
			if k == colName {
				match = true
				break
			}
		}
		if !match {
			delete(*columns, k)
		}
	}
}
