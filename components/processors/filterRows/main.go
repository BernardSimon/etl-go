package filterRows

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

var name = "filterRows"

func SetCustomName(customName string) {
	name = customName
}

// Processor 实现了 core.Processor 接口，用于根据用户定义的条件筛选数据记录。
// 只有满足条件的记录会被保留并传递到下一个处理阶段。
type Processor struct {
	column   string      // 要进行比较列名。
	operator string      // 比较操作符。
	value    interface{} // 用于比较的配置值。
}

// ProcessorCreator 返回处理器名称、实例和参数定义
func ProcessorCreator() (string, procrssor.Processor, []params.Params) {
	return name, &Processor{}, []params.Params{
		{
			Key:          "column",
			Required:     true,
			DefaultValue: "",
			Description:  "column to filter on",
		},
		{
			Key:          "operator",
			Required:     true,
			DefaultValue: "",
			Description:  "comparison operator",
		},
		{
			Key:          "value",
			Required:     true,
			DefaultValue: "",
			Description:  "value to compare against",
		},
	}
}

// Open 从配置中解析过滤条件，包括列名、操作符和要比较的值。
func (p *Processor) Open(config map[string]string) error {
	column, ok := config["column"]
	if !ok || column == "" {
		return fmt.Errorf("filterRows processor: config is missing or has invalid 'column'")
	}
	p.column = column

	operator, ok := config["operator"]
	if !ok || operator == "" {
		return fmt.Errorf("filterRows processor: config is missing or has invalid 'operator'")
	}
	p.operator = operator

	value, ok := config["value"]
	if !ok {
		return fmt.Errorf("filterRows processor: config is missing 'value'")
	}
	p.value = value

	return nil
}

// Process 根据定义的条件对记录进行评估。
// 如果记录满足条件，则原样返回；如果不满足，则返回 (nil, nil) 将其丢弃。
// 如果指定的 [column](file:///Users/szy/Desktop/code/etl-go/components/processors/maskData/main.go#L22-L22) 在记录中不存在，该记录也会被丢弃。
func (p *Processor) Process(record record.Record) (record.Record, error) {
	recordVal, ok := record[p.column]
	if !ok {
		// 如果记录中不存在要比较列，我们选择直接将其过滤掉。
		return nil, nil
	}

	match, err := p.compare(recordVal, p.value)
	if err != nil {
		// 如果比较过程中发生错误（例如，对字符串使用不支持的 ">" 操作符），则中止管道。
		return nil, fmt.Errorf("filterRows processor: error comparing values for column '%s': %w", p.column, err)
	}

	if match {
		return record, nil // 条件满足，保留该记录。
	}

	return nil, nil // 条件不满足，通过返回 nil 来过滤掉该记录。
}

// compare 实现了一个"智能"比较逻辑，它会优先尝试进行数字比较。
//
// 比较策略如下:
//  1. **数字比较优先**: 它会首先尝试将记录中的值和配置中的值都转换为 `float64`。
//  2. **执行数字比较**: 如果两个值都能成功转换为数字，则进行数字比较。
//     支持的数字操作符: `=`, `==`, `!=`, `<>`, `>`, `>=`, `<`, `<=`
//  3. **回退到字符串比较**: 如果上述两个值中任何一个无法转换为数字，则回退到字符串比较。
//  4. **执行字符串比较**: 对于字符串，只支持等于和不等于比较。
//     支持的字符串操作符: `=`, `==`, `!=`, `<>`
//     在字符串上使用 `>`, `<`, 等操作符会返回一个错误。
func (p *Processor) compare(recordValue, configValue interface{}) (bool, error) {
	val1Float, err1 := p.toFloat(recordValue)
	val2Float, err2 := p.toFloat(configValue)

	// 如果两个值都能成功转换为数字，则进行数字比较。
	if err1 == nil && err2 == nil {
		switch p.operator {
		case "=", "==":
			return val1Float == val2Float, nil
		case "!=", "<>":
			return val1Float != val2Float, nil
		case ">":
			return val1Float > val2Float, nil
		case ">=":
			return val1Float >= val2Float, nil
		case "<":
			return val1Float < val2Float, nil
		case "<=":
			return val1Float <= val2Float, nil
		default:
			return false, fmt.Errorf("unsupported numeric operator: '%s'", p.operator)
		}
	}

	// 如果无法进行数字比较，则进行字符串比较。
	val1Str := fmt.Sprintf("%v", recordValue)
	val2Str := fmt.Sprintf("%v", configValue)

	switch p.operator {
	case "=", "==":
		return val1Str == val2Str, nil
	case "!=", "<>":
		return val1Str != val2Str, nil
	default:
		// 对于字符串，我们只支持等于和不等于操作。
		return false, fmt.Errorf("unsupported string operator: '%s' (only '=', '==', '!=', '<>' are supported for non-numeric comparisons)", p.operator)
	}
}

// toFloat 尝试将一个 interface{} 类型的值转换为 float64，以便进行数字比较。
// 它能处理常见的数字类型以及可以被解析为数字的字符串。
func (p *Processor) toFloat(v interface{}) (float64, error) {
	switch i := v.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case json.Number: // 来自 JSON source 的数字可能是这个类型
		return i.Float64()
	case string:
		return strconv.ParseFloat(i, 64)
	default:
		return 0, fmt.Errorf("cannot convert type %T to float64", v)
	}
}

// Close 是一个无操作（no-op）方法，因为 filterRows 处理器是无状态的。
func (p *Processor) Close() error {
	return nil
}

func (p *Processor) HandleColumns(_ *map[string]string) {
	return
}
