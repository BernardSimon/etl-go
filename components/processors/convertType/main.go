package convertType

import (
	"fmt"
	"strconv"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

var name = "convertType"

func SetCustomName(customName string) {
	name = customName
}

type Processor struct {
	column string // 要转换的列名。
	toType string // 目标数据类型。
}

func ProcessorCreator() (string, procrssor.Processor, []params.Params) {
	return name, &Processor{}, []params.Params{
		{
			Key:          "column",
			Required:     true,
			DefaultValue: "",
			Description:  "column to convert",
		},
		{
			Key:          "type",
			Required:     true,
			DefaultValue: "",
			Description:  "type to convert to",
		},
	}
}

func (p *Processor) Open(config map[string]string) error {
	column, ok := config["column"]
	if !ok || column == "" {
		return fmt.Errorf("convertType processor: config is missing or has invalid 'column'")
	}
	p.column = column

	toType, ok := config["type"]
	if !ok || toType == "" {
		return fmt.Errorf("convertType processor: config is missing or has invalid 'type'")
	}
	p.toType = toType

	return nil
}

// Process 对记录进行处理，转换指定列的类型。
//
// 设计说明:
//   - 如果记录中不存在要转换的列，或者该列的值为 nil，处理器会静默忽略，以保证管道的弹性。
//   - 为了统一处理来自不同源（如数据库的 int、json 的 float64、csv 的 string）的原始类型，
//     它首先将输入值格式化为字符串，然后再从字符串解析为目标类型。
//   - 如果类型转换失败（例如，试图将 "abc" 转换为 integer），将返回一个错误，导致整个管道中止。
func (p *Processor) Process(record record.Record) (record.Record, error) {
	originalValue, ok := record[p.column]
	if !ok {
		// 如果列不存在，静默忽略。
		return record, nil
	}

	if originalValue == nil {
		// 不对 nil 值进行转换。
		return record, nil
	}

	var convertedValue interface{}
	var err error

	// 统一转换为字符串，再进行解析，以获得最大的兼容性。
	valStr := fmt.Sprintf("%v", originalValue)

	switch p.toType {
	case "integer", "int":
		convertedValue, err = strconv.ParseInt(valStr, 10, 64)
	case "float", "double":
		convertedValue, err = strconv.ParseFloat(valStr, 64)
	case "string":
		convertedValue = valStr
	case "boolean", "bool":
		// strconv.ParseBool 能很好地处理 "1", "t", "T", "TRUE", "true", "True" 等情况。
		convertedValue, err = strconv.ParseBool(valStr)
	default:
		return nil, fmt.Errorf("convertType processor: unsupported target type: '%s'", p.toType)
	}

	if err != nil {
		// 如果转换失败，返回一个清晰的错误，这将导致整个管道停止。
		return nil, fmt.Errorf("convertType processor: failed to convert value '%v' to type '%s' for column '%s': %w", originalValue, p.toType, p.column, err)
	}

	// 使用转换后的新值更新记录。
	record[p.column] = convertedValue

	return record, nil
}

func (p *Processor) Close() error {
	return nil
}

func (p *Processor) HandleColumns(_ *map[string]string) {
	return
}
