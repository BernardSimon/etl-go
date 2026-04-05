package jsonSink

import (
	"context"
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
	filePath string   // 输出文件路径。
	file     *os.File // 文件句柄。
	written  bool
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
func (s *Sink) Open(ctx context.Context, config map[string]string, columnMapping map[string]string, _ datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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

	if _, err := s.file.WriteString("["); err != nil {
		_ = s.file.Close()
		s.file = nil
		return fmt.Errorf("json sink: failed to initialize array start: %w", err)
	}
	s.written = false

	return nil
}

// Write 将一批记录以 JSON 对象的形式写入文件。
func (s *Sink) Write(ctx context.Context, ID string, records []record.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.ID = ID
	if len(records) == 0 {
		return nil
	}

	if s.file == nil {
		return fmt.Errorf("json sink: file is not initialized")
	}

	for _, r := range records {
		payload, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("json sink: failed to encode/write record: %w", err)
		}
		if s.written {
			if _, err := s.file.WriteString(",\n"); err != nil {
				return fmt.Errorf("json sink: failed to write separator: %w", err)
			}
		}
		if _, err := s.file.Write(payload); err != nil {
			return fmt.Errorf("json sink: failed to write record payload: %w", err)
		}
		s.written = true
	}

	return nil
}

// Close 负责关闭文件句柄并保存元信息。
func (s *Sink) Close() error {
	if s.file == nil {
		return nil
	}
	if s.written {
		if _, err := s.file.WriteString("\n]"); err != nil {
			_ = s.file.Close()
			return fmt.Errorf("json sink: failed to finalize array: %w", err)
		}
	} else {
		if _, err := s.file.WriteString("]"); err != nil {
			_ = s.file.Close()
			return fmt.Errorf("json sink: failed to finalize empty array: %w", err)
		}
	}
	return s.file.Close()
}
