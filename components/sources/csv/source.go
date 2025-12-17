package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
)

var name = "csv"

func SetCustomName(customName string) {
	name = customName
}

// Source 实现了 core.Source 接口，用于从 CSV 文件读取数据。
type Source struct {
	filePath  string      // 要读取的CSV文件路径
	file      *os.File    // 文件句柄
	reader    *csv.Reader // Go 标准库的 CSV 读取器
	header    []string    // 如果有表头，存储表头列名
	delimiter rune        // CSV文件的分隔符，默认为逗号
	line      int         // 当前已读取的行数，用于精确的错误报告
}

func SourceCreator() (string, source.Source, *string, []params.Params) {
	// 定义参数
	paramList := []params.Params{
		{
			Key:          "file_id",
			DefaultValue: "",
			Required:     true,
			Description:  "The file_id to the CSV file",
		},
		{
			Key:          "delimiter",
			DefaultValue: ",",
			Required:     true,
			Description:  "The delimiter used in the CSV file, default is comma",
		},
	}

	return name, &Source{}, nil, paramList
}

// Open 负责解析配置、打开 CSV 文件并准备读取。
// 它会处理文件路径、是否包含表头、自定义分隔符等配置项。
// 如果配置了 has_header: true，它会预先读取第一行作为后续 Record 的键。
func (s *Source) Open(config map[string]string, dataSource *datasource.Datasource) error {
	// 直接从 config 中获取 file_path，而不是通过 file_id 获取
	filePath, ok := config["file_path"]
	if !ok {
		return fmt.Errorf("csv source: config is missing required key 'file_path'")
	}
	s.filePath = filePath

	// 'delimiter' 是可选配置，默认为 ','。
	if delimiterStr, ok := config["delimiter"]; ok && len(delimiterStr) > 0 {
		// 只取第一个字符作为分隔符。
		s.delimiter = []rune(delimiterStr)[0]
	}

	var err error
	s.file, err = os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("csv source: failed to open file %s: %w", s.filePath, err)
	}

	s.reader = csv.NewReader(s.file)
	s.reader.Comma = s.delimiter

	s.line = 0

	// 如果有表头，立即读取并存储，以便后续的 Read() 调用使用。
	s.header, err = s.reader.Read()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("csv source: file is empty or contains only a header")
		}
		return fmt.Errorf("csv source: failed to read header: %w", err)
	}
	s.line++

	return nil
}

// Read 读取 CSV 文件中的下一行，并将其转换为一个 core.Record。
// 如果文件定义了表头，则使用表头作为键；否则，自动生成 "column_1", "column_2", ... 作为键。
// 它还会校验每行数据的列数是否与表头匹配，以确保数据规整。
func (s *Source) Read() (record.Record, error) {
	row, err := s.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF // 这是数据流正常结束的信号
		}
		return nil, fmt.Errorf("csv source: error reading data at line %d: %w", s.line+1, err)
	}
	s.line++

	r := make(record.Record)

	// 关键的数据完整性校验：确保每行数据的列数与表头一致。
	if len(row) != len(s.header) {
		return nil, fmt.Errorf("csv source: column count mismatch at line %d. Expected %d, got %d", s.line, len(s.header), len(row))
	}
	for i, value := range row {
		r[s.header[i]] = value
	}

	return r, nil
}

// Close 实现了 core.Source 接口，负责关闭已打开的文件句柄，释放资源。
func (s *Source) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

// Column 返回源数据的列映射关系
func (s *Source) Column() map[string]string {
	columns := make(map[string]string)
	for _, v := range s.header {
		columns[v] = v
	}
	return columns
}
