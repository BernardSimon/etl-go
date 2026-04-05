package http

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
)

var name = "http"

// Sink 实现了 core.Sink 接口，用于将数据批量推送到 HTTP API。
//
// 核心能力：
//   - 支持 POST/PUT/PATCH 方法
//   - 支持 Bearer Token、Basic Auth、API Key 认证
//   - 支持 body_template 自定义请求体结构（Go template 语法）
//   - 支持 send_mode 选择批量发送（数组）或逐条发送（单条对象）
//   - 模板内置签名函数：hmacSHA256 / md5 / sha256
//
// body_template 示例：
//
//	简单包装：
//	  {"code": 0, "data": {{.DataJSON}}}
//
//	带时间戳和签名：
//	  {"timestamp": {{.Timestamp}}, "sign": "{{hmacSHA256 .DataJSON "secret-key"}}", "data": {{.DataJSON}}}
//
//	自定义额外字段：
//	  {"app_id": "myapp", "batch_id": "{{.ID}}", "count": {{.Count}}, "items": {{.DataJSON}}}
type Sink struct {
	client        *http.Client
	url           string
	method        string
	headers       map[string]string
	columnMapping map[string]string

	// 认证
	authType   string // none, bearer, basic, api_key
	authValue  string // token / user:pass / key value
	apiKeyName string // API Key header name

	// 模板
	bodyTpl  *template.Template // 编译后的 body 模板，nil 表示直接发送 JSON 数组
	sendMode string             // batch 或 single
}

// TemplateData 是 body_template 中可用的变量集合。
type TemplateData struct {
	DataJSON    string // 数据的 JSON 字符串（batch 模式下是数组，single 模式下是对象）
	Timestamp   int64  // 当前 Unix 时间戳（秒）
	TimestampMs int64  // 当前 Unix 时间戳（毫秒）
	ID          string // 批次 ID
	Count       int    // 当前批次的记录数
}

// templateFuncs 是 body_template 中可用的内置函数。
var templateFuncs = template.FuncMap{
	// hmacSHA256 "data" "secret" -> HMAC-SHA256 签名的十六进制字符串
	"hmacSHA256": func(data, secret string) string {
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(data))
		return hex.EncodeToString(h.Sum(nil))
	},
	// md5 "data" -> MD5 哈希的十六进制字符串
	"md5": func(data string) string {
		h := md5.Sum([]byte(data))
		return hex.EncodeToString(h[:])
	},
	// sha256 "data" -> SHA-256 哈希的十六进制字符串
	"sha256": func(data string) string {
		h := sha256.Sum256([]byte(data))
		return hex.EncodeToString(h[:])
	},
	// concat 拼接多个字符串
	"concat": func(parts ...string) string {
		return strings.Join(parts, "")
	},
	// toString 将任意值转为字符串
	"toString": func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	},
}

func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return name, &Sink{}, nil, []params.Params{
		{
			Key:         "url",
			Required:    true,
			Description: "Target HTTP API URL",
			Example:     "https://api.example.com/data",
		},
		{
			Key:          "method",
			Required:     false,
			DefaultValue: "POST",
			Description:  "HTTP method (POST, PUT, PATCH)",
		},
		{
			Key:         "headers",
			Required:    false,
			Description: "Custom HTTP headers in JSON format, e.g. {\"X-Custom\": \"value\"}",
		},
		{
			Key:          "auth_type",
			Required:     false,
			DefaultValue: "none",
			Description:  "Authentication type: none, bearer, basic, api_key",
		},
		{
			Key:         "auth_value",
			Required:    false,
			Description: "Auth credential: Bearer token string / basic auth as 'user:password' / API key value",
		},
		{
			Key:          "api_key_name",
			Required:     false,
			DefaultValue: "X-API-Key",
			Description:  "Header name for API key authentication",
		},
		{
			Key:     "body_template",
			Required: false,
			Description: "Go template for custom request body structure. " +
				"Available variables: .DataJSON (JSON string), .Timestamp (unix sec), " +
				".TimestampMs (unix ms), .ID (batch id), .Count (record count). " +
				"Built-in functions: hmacSHA256, md5, sha256, concat, toString. " +
				"If empty, sends raw JSON array.",
			Example: `{"timestamp": {{.Timestamp}}, "sign": "{{hmacSHA256 .DataJSON "secret"}}", "data": {{.DataJSON}}}`,
		},
		{
			Key:          "send_mode",
			Required:     false,
			DefaultValue: "batch",
			Description:  "Send mode: 'batch' sends all records as JSON array per batch, 'single' sends one request per record",
		},
	}
}

func (s *Sink) Open(ctx context.Context, config map[string]string, columnMapping map[string]string, _ datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	url, ok := config["url"]
	if !ok || url == "" {
		return fmt.Errorf("http sink: config is missing required key 'url'")
	}
	s.url = url

	s.method = "POST"
	if m, ok := config["method"]; ok && m != "" {
		s.method = strings.ToUpper(m)
	}

	// 解析自定义 headers
	s.headers = make(map[string]string)
	if h, ok := config["headers"]; ok && h != "" {
		if err := json.Unmarshal([]byte(h), &s.headers); err != nil {
			return fmt.Errorf("http sink: failed to parse 'headers' as JSON: %w", err)
		}
	}

	s.columnMapping = columnMapping

	// 认证配置
	s.authType = "none"
	if at, ok := config["auth_type"]; ok && at != "" {
		s.authType = at
	}
	s.authValue = config["auth_value"]

	s.apiKeyName = "X-API-Key"
	if akn, ok := config["api_key_name"]; ok && akn != "" {
		s.apiKeyName = akn
	}

	// body_template 编译
	if tplStr, ok := config["body_template"]; ok && tplStr != "" {
		tpl, err := template.New("body").Funcs(templateFuncs).Parse(tplStr)
		if err != nil {
			return fmt.Errorf("http sink: failed to parse body_template: %w", err)
		}
		s.bodyTpl = tpl
	}

	// send_mode
	s.sendMode = "batch"
	if sm, ok := config["send_mode"]; ok && sm != "" {
		switch sm {
		case "batch", "single":
			s.sendMode = sm
		default:
			return fmt.Errorf("http sink: invalid send_mode '%s', expected 'batch' or 'single'", sm)
		}
	}

	s.client = &http.Client{Timeout: 120 * time.Second}

	return nil
}

func (s *Sink) Write(ctx context.Context, id string, records []record.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	// 应用 column_mapping，构建映射后的数据
	mappedData := s.applyMapping(records)

	if s.sendMode == "single" {
		// 逐条发送
		for i, row := range mappedData {
			if err := ctx.Err(); err != nil {
				return err
			}
			body, err := s.buildBody([]map[string]interface{}{row}, id, 1, true)
			if err != nil {
				return fmt.Errorf("http sink: failed to build body for record %d: %w", i, err)
			}
			if err := s.doRequest(ctx, body); err != nil {
				return fmt.Errorf("http sink: failed to send record %d: %w", i, err)
			}
		}
		return nil
	}

	// batch 模式
	body, err := s.buildBody(mappedData, id, len(mappedData), false)
	if err != nil {
		return fmt.Errorf("http sink: failed to build body: %w", err)
	}
	return s.doRequest(ctx, body)
}

func (s *Sink) Close() error {
	s.client = nil
	return nil
}

// applyMapping 对所有 records 应用 column_mapping。
func (s *Sink) applyMapping(records []record.Record) []map[string]interface{} {
	data := make([]map[string]interface{}, 0, len(records))
	for _, r := range records {
		row := make(map[string]interface{})
		if len(s.columnMapping) > 0 {
			for recordKey, apiField := range s.columnMapping {
				val, exists := r[recordKey]
				if !exists || val == nil {
					row[apiField] = nil
				} else {
					row[apiField] = val
				}
			}
		} else {
			for k, v := range r {
				row[k] = v
			}
		}
		data = append(data, row)
	}
	return data
}

// buildBody 根据是否配置了 body_template 来构建请求体。
// single=true 时 data 只有一个元素，DataJSON 序列化为对象而非数组。
func (s *Sink) buildBody(data []map[string]interface{}, id string, count int, single bool) ([]byte, error) {
	// 序列化数据
	var dataJSON []byte
	var err error
	if single && len(data) == 1 {
		dataJSON, err = json.Marshal(data[0])
	} else {
		dataJSON, err = json.Marshal(data)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	// 没有模板，直接返回原始 JSON
	if s.bodyTpl == nil {
		return dataJSON, nil
	}

	// 渲染模板
	now := time.Now()
	td := TemplateData{
		DataJSON:    string(dataJSON),
		Timestamp:   now.Unix(),
		TimestampMs: now.UnixMilli(),
		ID:          id,
		Count:       count,
	}

	var buf bytes.Buffer
	if err := s.bodyTpl.Execute(&buf, td); err != nil {
		return nil, fmt.Errorf("failed to render body_template: %w", err)
	}

	return buf.Bytes(), nil
}

// doRequest 发送 HTTP 请求并检查响应。
func (s *Sink) doRequest(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, s.method, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	switch s.authType {
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+s.authValue)
	case "basic":
		parts := strings.SplitN(s.authValue, ":", 2)
		if len(parts) == 2 {
			req.SetBasicAuth(parts[0], parts[1])
		}
	case "api_key":
		req.Header.Set(s.apiKeyName, s.authValue)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
