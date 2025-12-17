package maskData

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

var name = "maskData"

func SetCustomName(customName string) {
	name = customName
}

// Processor 实现了 core.Processor 接口，用于对指定列的敏感数据进行脱敏。
// 它通过单向哈希算法将原始值替换为一个固定长度的哈希值，实现数据的假名化（pseudonymization）。
// 这对于在保留数据分析能力的同时保护用户隐私至关重要。
type Processor struct {
	column string // 要脱敏的列名。
	method string // 脱敏方法（哈希算法）。
}

// ProcessorCreator 返回处理器名称、实例和参数定义
func ProcessorCreator() (string, procrssor.Processor, []params.Params) {
	return name, &Processor{}, []params.Params{
		{
			Key:          "column",
			Required:     true,
			DefaultValue: "",
			Description:  "column to mask",
		},
		{
			Key:          "method",
			Required:     true,
			DefaultValue: "sha256",
			Description:  "hashing method (md5 or sha256)",
		},
	}
}

// Open 从配置中解析需要脱敏的列和所用的哈希方法。
// 支持的脱敏方法 ([method](file:///Users/szy/Desktop/code/etl-go/components/processors/maskData/main.go#L23-L23)):
// - "md5"
// - "sha256"
func (p *Processor) Open(config map[string]string) error {
	column, ok := config["column"]
	if !ok || column == "" {
		return fmt.Errorf("maskData processor: config is missing or has invalid 'column'")
	}
	p.column = column

	method, ok := config["method"]
	if !ok || method == "" {
		return fmt.Errorf("maskData processor: config is missing or has invalid 'method'")
	}
	p.method = method

	return nil
}

// Process 对记录进行处理，使用指定的哈希算法替换原始值。
//
// 设计与安全说明:
//   - 如果列不存在或值为 nil，处理器会静默忽略，以保证管道的弹性。
//   - 任何类型的值都会先被转换为字符串，然后再进行哈希计算。
//   - **安全警告**: 这是一种单向哈希，不是加密。相同地输入值将始终产生相同地输出哈希值。
//     这对于保持数据关联性很有用，但也意味着如果攻击者能够猜到原始值（例如，对于低复杂度的输入），
//     他们可以通过彩虹表攻击来反查。请勿将其用于需要可逆加密的场景。
func (p *Processor) Process(record record.Record) (record.Record, error) {
	originalValue, ok := record[p.column]
	if !ok {
		// 如果列不存在，静默忽略。
		return record, nil
	}

	if originalValue == nil {
		// 不处理 nil 值。
		return record, nil
	}

	// 将原始值转换为字符串以进行哈希。
	valStr := fmt.Sprintf("%v", originalValue)
	var maskedValue string

	switch p.method {
	case "md5":
		hasher := md5.New()
		hasher.Write([]byte(valStr))
		maskedValue = hex.EncodeToString(hasher.Sum(nil))
	case "sha256":
		hasher := sha256.New()
		hasher.Write([]byte(valStr))
		maskedValue = hex.EncodeToString(hasher.Sum(nil))
	default:
		return nil, fmt.Errorf("maskData processor: unsupported hashing method: '%s'", p.method)
	}

	// 使用脱敏后的新值更新记录。
	record[p.column] = maskedValue

	return record, nil
}

// Close 是一个无操作（no-op）方法，因为此处理器是无状态的。
func (p *Processor) Close() error {
	return nil
}

func (p *Processor) HandleColumns(_ *map[string]string) {
	return
}
