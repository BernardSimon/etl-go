package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
)

var name = "http"

// Source 实现了 core.Source 接口，用于从 HTTP API 读取数据。
//
// 支持从 RESTful API 获取 JSON 数据，支持分页（offset、page、cursor）、
// 自定义请求头和请求体，以及通过 data_path 提取嵌套 JSON 中的数据数组。
type Source struct {
	client  *http.Client
	url     string
	method  string
	headers map[string]string
	body    string

	// 分页相关
	paginationType string // none, offset, page, cursor
	pageSize       int
	currentOffset  int
	currentPage    int
	cursorField    string
	nextCursor     string

	// 数据提取
	dataPath string // 用于从嵌套 JSON 中提取数据数组的路径，如 "data.items"

	// 内部状态
	keys    []string
	buffer  []record.Record
	pos     int
	done    bool
	started bool
}

// SourceCreator 实现了源组件的创建接口
func SourceCreator() (string, source.Source, *string, []params.Params) {
	paramList := []params.Params{
		{
			Key:         "url",
			Required:    true,
			Description: "HTTP request URL",
			Example:     "https://api.example.com/data",
		},
		{
			Key:          "method",
			Required:     false,
			DefaultValue: "GET",
			Description:  "HTTP method (GET or POST)",
		},
		{
			Key:         "headers",
			Required:    false,
			Description: "HTTP request headers in JSON format, e.g. {\"Authorization\": \"Bearer token\"}",
			Example:     "{\"Authorization\": \"Bearer xxx\"}",
		},
		{
			Key:         "body",
			Required:    false,
			Description: "HTTP request body (for POST requests), JSON string",
		},
		{
			Key:          "pagination_type",
			Required:     false,
			DefaultValue: "none",
			Description:  "Pagination type: none, offset, page, cursor",
		},
		{
			Key:          "page_size",
			Required:     false,
			DefaultValue: "100",
			Description:  "Number of records per page (used with pagination)",
		},
		{
			Key:          "cursor_field",
			Required:     false,
			DefaultValue: "next_cursor",
			Description:  "Field name for cursor value in response (used with cursor pagination)",
		},
		{
			Key:         "data_path",
			Required:    false,
			Description: "Dot-separated path to extract data array from response, e.g. 'data.items'",
			Example:     "data.items",
		},
	}
	return name, &Source{}, nil, paramList
}

func (s *Source) Open(ctx context.Context, config map[string]string, _ datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	url, ok := config["url"]
	if !ok || url == "" {
		return fmt.Errorf("http source: config is missing required key 'url'")
	}
	s.url = url

	s.method = "GET"
	if m, ok := config["method"]; ok && m != "" {
		s.method = strings.ToUpper(m)
	}

	// 解析 headers
	s.headers = make(map[string]string)
	if h, ok := config["headers"]; ok && h != "" {
		if err := json.Unmarshal([]byte(h), &s.headers); err != nil {
			return fmt.Errorf("http source: failed to parse 'headers' as JSON: %w", err)
		}
	}

	s.body = config["body"]

	// 分页配置
	s.paginationType = "none"
	if pt, ok := config["pagination_type"]; ok && pt != "" {
		s.paginationType = pt
	}

	s.pageSize = 100
	if ps, ok := config["page_size"]; ok && ps != "" {
		var size int
		if _, err := fmt.Sscanf(ps, "%d", &size); err == nil && size > 0 {
			s.pageSize = size
		}
	}

	s.cursorField = "next_cursor"
	if cf, ok := config["cursor_field"]; ok && cf != "" {
		s.cursorField = cf
	}

	s.dataPath = config["data_path"]

	s.client = &http.Client{Timeout: 120 * time.Second}
	s.currentOffset = 0
	s.currentPage = 1
	s.pos = 0
	s.done = false
	s.started = false

	// 预取第一批数据以确定 keys
	if err := s.fetchPage(ctx); err != nil {
		return fmt.Errorf("http source: failed to fetch initial data: %w", err)
	}

	// 从第一批数据中提取所有 keys
	keysSet := make(map[string]bool)
	for _, r := range s.buffer {
		for k := range r {
			keysSet[k] = true
		}
	}
	s.keys = make([]string, 0, len(keysSet))
	for k := range keysSet {
		s.keys = append(s.keys, k)
	}
	sort.Strings(s.keys)

	s.started = true
	return nil
}

func (s *Source) Read(ctx context.Context) (record.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 当前缓冲区还有数据
	if s.pos < len(s.buffer) {
		r := s.buffer[s.pos]
		s.pos++
		return r, nil
	}

	// 缓冲区读完，尝试获取下一页
	if s.done {
		return nil, io.EOF
	}

	if err := s.fetchPage(ctx); err != nil {
		return nil, err
	}

	if len(s.buffer) == 0 {
		return nil, io.EOF
	}

	r := s.buffer[s.pos]
	s.pos++
	return r, nil
}

func (s *Source) Close() error {
	s.client = nil
	s.buffer = nil
	return nil
}

func (s *Source) Column() map[string]string {
	columns := make(map[string]string)
	for _, v := range s.keys {
		columns[v] = v
	}
	return columns
}

// fetchPage 发送 HTTP 请求获取一页数据
func (s *Source) fetchPage(ctx context.Context) error {
	url := s.buildURL()

	var bodyReader io.Reader
	if s.body != "" && s.method == "POST" {
		bodyReader = strings.NewReader(s.body)
	}

	req, err := http.NewRequestWithContext(ctx, s.method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("http source: failed to create request: %w", err)
	}

	for k, v := range s.headers {
		req.Header.Set(k, v)
	}
	if s.method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("http source: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http source: request returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("http source: failed to read response body: %w", err)
	}

	// 解析响应
	var rawData interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return fmt.Errorf("http source: failed to parse response as JSON: %w", err)
	}

	// 提取数据数组
	dataArray, fullResponse, err := s.extractData(rawData)
	if err != nil {
		return fmt.Errorf("http source: %w", err)
	}

	// 转换为 records
	s.buffer = make([]record.Record, 0, len(dataArray))
	for _, item := range dataArray {
		if obj, ok := item.(map[string]interface{}); ok {
			s.buffer = append(s.buffer, record.Record(obj))
		}
	}
	s.pos = 0

	// 处理分页
	pageLen := len(s.buffer)
	switch s.paginationType {
	case "offset":
		s.currentOffset += pageLen
		if pageLen < s.pageSize {
			s.done = true
		}
	case "page":
		s.currentPage++
		if pageLen < s.pageSize {
			s.done = true
		}
	case "cursor":
		cursor := s.extractCursor(fullResponse)
		if cursor == "" || pageLen == 0 {
			s.done = true
		} else {
			s.nextCursor = cursor
		}
	default:
		// none: 只请求一次
		s.done = true
	}

	return nil
}

// buildURL 根据分页类型构建请求 URL
func (s *Source) buildURL() string {
	url := s.url

	if !s.started {
		// 第一次请求，根据分页类型添加参数
		switch s.paginationType {
		case "offset":
			url = appendQueryParam(url, "offset", "0")
			url = appendQueryParam(url, "limit", fmt.Sprintf("%d", s.pageSize))
		case "page":
			url = appendQueryParam(url, "page", "1")
			url = appendQueryParam(url, "page_size", fmt.Sprintf("%d", s.pageSize))
		}
		return url
	}

	switch s.paginationType {
	case "offset":
		url = appendQueryParam(url, "offset", fmt.Sprintf("%d", s.currentOffset))
		url = appendQueryParam(url, "limit", fmt.Sprintf("%d", s.pageSize))
	case "page":
		url = appendQueryParam(url, "page", fmt.Sprintf("%d", s.currentPage))
		url = appendQueryParam(url, "page_size", fmt.Sprintf("%d", s.pageSize))
	case "cursor":
		if s.nextCursor != "" {
			url = appendQueryParam(url, "cursor", s.nextCursor)
		}
	}

	return url
}

// extractData 从响应中提取数据数组
func (s *Source) extractData(rawData interface{}) ([]interface{}, interface{}, error) {
	// 如果没有指定 data_path，尝试直接当数组处理
	if s.dataPath == "" {
		if arr, ok := rawData.([]interface{}); ok {
			return arr, rawData, nil
		}
		// 如果是对象，尝试当作单条记录
		if obj, ok := rawData.(map[string]interface{}); ok {
			return []interface{}{obj}, rawData, nil
		}
		return nil, rawData, fmt.Errorf("response is neither an array nor an object")
	}

	// 按路径提取嵌套数据
	parts := strings.Split(s.dataPath, ".")
	var current interface{} = rawData
	for _, part := range parts {
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, rawData, fmt.Errorf("cannot traverse path '%s': intermediate value is not an object", s.dataPath)
		}
		current, ok = obj[part]
		if !ok {
			return nil, rawData, fmt.Errorf("path '%s' not found in response (missing key '%s')", s.dataPath, part)
		}
	}

	if arr, ok := current.([]interface{}); ok {
		return arr, rawData, nil
	}
	return nil, rawData, fmt.Errorf("value at path '%s' is not an array", s.dataPath)
}

// extractCursor 从响应中提取 cursor 值
func (s *Source) extractCursor(rawData interface{}) string {
	obj, ok := rawData.(map[string]interface{})
	if !ok {
		return ""
	}

	// 支持点分路径
	parts := strings.Split(s.cursorField, ".")
	var current interface{} = obj
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current, ok = m[part]
		if !ok {
			return ""
		}
	}

	if str, ok := current.(string); ok {
		return str
	}
	// 数字类型的 cursor
	if num, ok := current.(float64); ok {
		return fmt.Sprintf("%.0f", num)
	}
	return ""
}

func appendQueryParam(rawURL, key, value string) string {
	if strings.Contains(rawURL, "?") {
		return rawURL + "&" + key + "=" + value
	}
	return rawURL + "?" + key + "=" + value
}
