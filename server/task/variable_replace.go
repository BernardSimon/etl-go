package task

import (
	"encoding/json"
	"fmt"
	"regexp"

	_type "github.com/BernardSimon/etl-go/server/type"
)

var variablePattern = regexp.MustCompile(`\$\{([^}]+)}`)

// ReplaceVariables 对 TaskData 中所有组件的 Params.Value 执行变量替换。
// 只替换 Value 字段中的 ${...} 占位符，不修改 Key 和其他结构字段。
// 返回深拷贝后的 TaskData，不修改原始数据。
func ReplaceVariables(data *_type.TaskData) (*_type.TaskData, error) {
	if data == nil {
		return nil, nil
	}

	// 深拷贝：序列化再反序列化
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task data: %w", err)
	}
	var copied _type.TaskData
	if err := json.Unmarshal(raw, &copied); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	// 收集所有需要替换的变量名，批量解析避免重复查询
	allValues := collectParamValues(&copied)
	variableNames := make(map[string]struct{})
	for _, v := range allValues {
		for _, match := range variablePattern.FindAllStringSubmatch(*v, -1) {
			variableNames[match[1]] = struct{}{}
		}
	}

	if len(variableNames) == 0 {
		return &copied, nil
	}

	// 一次性解析所有变量值
	resolved := make(map[string]string, len(variableNames))
	for name := range variableNames {
		value, err := GetValueByName(name)
		if err != nil {
			return nil, fmt.Errorf("变量解析错误: ${%s}: %w", name, err)
		}
		resolved[name] = value
	}

	// 对每个 Param.Value 执行替换
	for _, v := range allValues {
		*v = variablePattern.ReplaceAllStringFunc(*v, func(match string) string {
			name := match[2 : len(match)-1] // 去掉 ${ 和 }
			if val, ok := resolved[name]; ok {
				return val
			}
			return match
		})
	}

	return &copied, nil
}

// collectParamValues 收集 TaskData 中所有 Params 的 Value 指针，用于原地替换。
func collectParamValues(data *_type.TaskData) []*string {
	var ptrs []*string

	if data.BeforeExecute != nil {
		for i := range data.BeforeExecute.Params {
			ptrs = append(ptrs, &data.BeforeExecute.Params[i].Value)
		}
	}
	for i := range data.Source.Params {
		ptrs = append(ptrs, &data.Source.Params[i].Value)
	}
	for pi := range data.Processors {
		for i := range data.Processors[pi].Params {
			ptrs = append(ptrs, &data.Processors[pi].Params[i].Value)
		}
	}
	for i := range data.Sinks.Params {
		ptrs = append(ptrs, &data.Sinks.Params[i].Value)
	}
	if data.AfterExecute != nil {
		for i := range data.AfterExecute.Params {
			ptrs = append(ptrs, &data.AfterExecute.Params[i].Value)
		}
	}

	return ptrs
}
