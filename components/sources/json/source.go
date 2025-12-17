package json

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
)

var name = "json"

// Source 实现了 core.Source 接口，用于从 JSON 文件读取数据。
//
// 它被设计用来流式处理一个包含对象数组的 JSON 文件。
// 示例格式: [{"id": 1, "name": "A"}, {"id": 2, "name": "B"}]
//
// 注意：此实现不支持 JSON Lines (.jsonl) 或 NDJSON 格式（即每行一个独立的JSON对象）。
// 它通过使用标准库的 json.Decoder，实现了高效的内存使用，可以处理G字节级别的大文件。
type Source struct {
	filePath string        // 要读取的JSON文件路径
	file     *os.File      // 文件句柄
	decoder  *json.Decoder // Go 标准库的流式 JSON 解码器
	keys     []string      // 所有可能的键集合
}

// SourceCreator 实现了源组件的创建接口，返回组件名称、实例和参数定义
func SourceCreator() (string, source.Source, *string, []params.Params) {
	paramList := []params.Params{
		{
			Key:          "file_id",
			DefaultValue: "",
			Required:     true,
			Description:  "The file_id to the JSON file",
		},
		{
			Key:          "keys_sample_rows",
			DefaultValue: "100",
			Required:     true,
			Description:  "Number of rows to sample for determining keys, default is 100",
		},
	}

	return name, &Source{}, nil, paramList
}

// Open 负责解析配置、打开文件，并验证 JSON 格式的起始部分。
// 一个关键步骤是它会立即尝试读取 JSON 数组的起始符'['。
// 这是一种"快速失败"策略，可以及早确认文件格式是否符合预期。
func (s *Source) Open(config map[string]string, dataSource *datasource.Datasource) error {
	filePath, ok := config["file_path"]
	if !ok {
		return fmt.Errorf("json source: config is missing required key 'file_path'")
	}
	s.filePath = filePath

	// 采样行数
	keysSampleRows := 100
	if keysSampleRowsStr, ok := config["keys_sample_rows"]; ok {
		if parsed, err := strconv.Atoi(keysSampleRowsStr); err == nil && parsed >= 0 {
			keysSampleRows = parsed
		}
	}
	var err error
	s.file, err = os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("json source: failed to open file %s: %w", s.filePath, err)
	}

	s.decoder = json.NewDecoder(s.file)

	// 验证文件是否以 `[` 开头。
	token, err := s.decoder.Token()
	if err != nil {
		return fmt.Errorf("json source: failed to read opening bracket of json array: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("json source: expected file to start with a json array '[', but got '%v'", token)
	}

	// 新增逻辑：预读取所有对象的 keys 并集
	keysSet := make(map[string]bool)
	sampleRows := 0

	for s.decoder.More() {
		sampleRows++
		var r record.Record
		if err := s.decoder.Decode(&r); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("json source: failed to decode json object: %w", err)
		}

		for key := range r {
			keysSet[key] = true
		}

		if sampleRows >= keysSampleRows && keysSampleRows > 0 {
			break
		}
	}

	// 将 keysSet 转换为 slice 存储到 s.keys
	s.keys = make([]string, 0, len(keysSet))
	for key := range keysSet {
		s.keys = append(s.keys, key)
	}

	// 回到文件开始位置以便后续正常读取
	_, err = s.file.Seek(0, 0)
	if err != nil {
		return err
	}
	s.decoder = json.NewDecoder(s.file)
	_, err = s.decoder.Token()
	if err != nil {
		return err
	} // Skip the opening '['

	return nil
}

// Read 从 JSON 数组流中解码下一个对象，并将其转换为 core.Record。
// 它依赖 `decoder.More()` 来判断数组中是否还有更多元素。
// 当 `More()` 返回 false 时，表示已到达数组末尾，此时方法会返回 io.EOF 来通知管道数据已耗尽。
func (s *Source) Read() (record.Record, error) {
	// `decoder.More()` 是驱动流式读取的核心。
	if !s.decoder.More() {
		// 当没有更多元素时，我们期望读到数组的结束符 ']'。
		// 这确认了 JSON 文件的结构是完整的。
		token, err := s.decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("json source: error reading closing bracket of json array: %w", err)
		}
		if delim, ok := token.(json.Delim); ok && delim == ']' {
			return nil, io.EOF // 成功到达数组末尾，这是数据流正常结束的信号。
		}
		return nil, fmt.Errorf("json source: expected end of json array ']', but got '%v'", token)
	}

	// 解码下一个 JSON 对象到 record 中。
	var r record.Record
	if err := s.decoder.Decode(&r); err != nil {
		// 如果在解码过程中遇到 io.EOF，说明文件在数组结束前被截断了。
		if err == io.EOF {
			return nil, fmt.Errorf("json source: unexpected end of file, json array not closed with ']'")
		}
		return nil, fmt.Errorf("json source: failed to decode json object: %w", err)
	}

	return r, nil
}

// Close 关闭文件句柄，释放资源。
func (s *Source) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Column 返回源数据的列映射关系
func (s *Source) Column() map[string]string {
	columns := make(map[string]string)
	for _, v := range s.keys {
		columns[v] = v
	}
	return columns
}
