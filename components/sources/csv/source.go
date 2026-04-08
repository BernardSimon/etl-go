package csv

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
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
	firstRow  []string    // 无表头场景下缓存第一行数据
	delimiter rune        // CSV文件的分隔符，默认为逗号
	line      int         // 当前已读取的行数，用于精确的错误报告
	hasHeader bool
	encoding  string
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
		{
			Key:          "has_header",
			DefaultValue: "true",
			Required:     true,
			Description:  "Whether the CSV file has a header row",
		},
		{
			Key:          "encoding",
			DefaultValue: "utf-8",
			Required:     false,
			Description:  "File encoding: utf-8, gbk, or gb18030",
		},
	}

	return name, &Source{}, nil, paramList
}

// Open 负责解析配置、打开 CSV 文件并准备读取。
// 它会处理文件路径、是否包含表头、自定义分隔符等配置项。
// 如果配置了 has_header: true，它会预先读取第一行作为后续 Record 的键。
func (s *Source) Open(ctx context.Context, config map[string]string, dataSource datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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
	} else {
		s.delimiter = ','
	}
	s.hasHeader = config["has_header"] != "false"
	s.encoding = normalizeEncoding(config["encoding"])
	s.header = nil
	s.firstRow = nil

	closeOnError := true
	defer func() {
		if closeOnError && s.file != nil {
			_ = s.file.Close()
			s.file = nil
		}
	}()

	var err error
	s.file, err = os.Open(s.filePath)
	if err != nil {
		return fmt.Errorf("csv source: failed to open file %s: %w", s.filePath, err)
	}

	csvInput, err := wrapReaderWithEncoding(s.file, s.encoding)
	if err != nil {
		return err
	}

	s.reader = csv.NewReader(csvInput)
	s.reader.Comma = s.delimiter

	s.line = 0

	row, err := s.reader.Read()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("csv source: file is empty")
		}
		return fmt.Errorf("csv source: failed to read first row: %w", err)
	}
	s.line++
	if s.hasHeader {
		s.header = row
	} else {
		s.firstRow = row
		s.header = make([]string, len(row))
		for i := range row {
			s.header[i] = fmt.Sprintf("column_%d", i+1)
		}
	}

	closeOnError = false
	return nil
}

// Read 读取 CSV 文件中的下一行，并将其转换为一个 core.Record。
// 如果文件定义了表头，则使用表头作为键；否则，自动生成 "column_1", "column_2", ... 作为键。
// 它还会校验每行数据的列数是否与表头匹配，以确保数据规整。
func (s *Source) Read(ctx context.Context) (record.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	row := s.firstRow
	if row != nil {
		s.firstRow = nil
	} else {
		var err error
		row, err = s.reader.Read()
		if err != nil {
			if err == io.EOF {
				return nil, io.EOF // 这是数据流正常结束的信号
			}
			return nil, fmt.Errorf("csv source: error reading data at line %d: %w", s.line+1, err)
		}
		s.line++
	}

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

func normalizeEncoding(raw string) string {
	encoding := strings.ToLower(strings.TrimSpace(raw))
	if encoding == "" {
		return "utf-8"
	}
	switch encoding {
	case "utf8":
		return "utf-8"
	case "gb2312":
		return "gbk"
	default:
		return encoding
	}
}

func wrapReaderWithEncoding(r io.Reader, encoding string) (io.Reader, error) {
	switch encoding {
	case "utf-8":
		return r, nil
	case "gbk":
		return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder()), nil
	case "gb18030":
		return transform.NewReader(r, simplifiedchinese.GB18030.NewDecoder()), nil
	default:
		return nil, fmt.Errorf("csv source: unsupported encoding %q, expected utf-8, gbk, or gb18030", encoding)
	}
}
