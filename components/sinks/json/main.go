package jsonSink

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
)

// Sink 实现了 core.Sink 接口，用于将数据以 JSON 格式写入文件。
type Sink struct {
	ID       string
	filePath string        // 输出文件路径。
	file     *os.File      // 文件句柄。
	encoder  *json.Encoder // JSON 编码器。
}

func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return "json", &Sink{}, nil, []params.Params{
		{
			Key:         "file_name",
			Description: "The name of the output file",
			Required:    true,
		},
		{
			Key:          "file_ext",
			Description:  "The extension of the output file",
			DefaultValue: "json",
			Required:     true,
		},
	}
}

// Open 打开输出文件并初始化编码器
func (s *Sink) Open(config map[string]string, columnMapping map[string]string, _ *datasource.Datasource) error {
	filePath, ok := config["file_path"]
	if !ok {
		return fmt.Errorf("json sink: config is missing or has invalid 'file_name'")
	}
	s.filePath = filePath

	var err error
	s.file, err = os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("json sink: failed to create/open file: %w", err)
	}

	s.encoder = json.NewEncoder(s.file)

	return nil
}

// Write 将一批记录以 JSON 对象的形式写入文件。
func (s *Sink) Write(ID string, records []record.Record) error {
	s.ID = ID
	if len(records) == 0 {
		return nil
	}

	if s.encoder == nil {
		return fmt.Errorf("json sink: encoder is not initialized")
	}

	// 写入JSON数组开始符号
	if _, err := s.file.WriteString("[\n"); err != nil {
		return fmt.Errorf("json sink: failed to write array start: %w", err)
	}

	// 写入每条记录
	for i, r := range records {
		if err := s.encoder.Encode(r); err != nil {
			return fmt.Errorf("json sink: failed to encode/write record: %w", err)
		}
		// 如果不是最后一条记录，添加逗号
		if i < len(records)-1 {
			if _, err := s.file.WriteString(",\n"); err != nil {
				return fmt.Errorf("json sink: failed to write separator: %w", err)
			}
		}
	}

	// 写入JSON数组结束符号
	if _, err := s.file.WriteString("\n]"); err != nil {
		return fmt.Errorf("json sink: failed to write array end: %w", err)
	}

	return nil
}

// Close 负责关闭文件句柄并保存元信息。
func (s *Sink) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}
