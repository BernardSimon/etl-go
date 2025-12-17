package doris

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
)

var name = "doris"
var datasourceName = "doris"

func SetCustomName(customName string, customDatasourceName string) {
	name = customName
	datasourceName = customDatasourceName
}

// Sink 实现了 core.Sink 接口，用于将数据批量写入到 Doris 数据库。
type Sink struct {
	client        *http.Client      // HTTP 客户端
	url           string            // Doris Stream Load URL
	user          string            // 用户名
	password      string            // 密码
	columnMapping map[string]string // 列映射关系 (Record 中的键 -> 数据库中的列名)
	table         string            // 表名
}

func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return name, &Sink{
			client: &http.Client{
				Timeout: 600 * time.Second,
			},
		}, &datasourceName, []params.Params{
			{
				Key:          "table",
				Required:     true,
				DefaultValue: "",
				Description:  "doris table name",
			},
		}
}

// Open 负责解析配置并初始化 Doris Stream Load 设置。
func (s *Sink) Open(config map[string]string, columnMapping map[string]string, dataSource *datasource.Datasource) error {
	// 处理 column_mapping
	if len(columnMapping) == 0 {
		return fmt.Errorf("doris sink: 'column_mapping' cannot be empty")
	}
	s.columnMapping = columnMapping

	// 从 datasource 获取基础配置
	var host, port, user, password, database string
	if dataSource != nil {
		dsConfig := (*dataSource).Open()
		if dsConfigMap, ok := dsConfig.(map[string]string); ok {
			host = dsConfigMap["host"]
			port = dsConfigMap["port"]
			user = dsConfigMap["user"]
			password = dsConfigMap["password"]
			database = dsConfigMap["database"]
		}
	}

	// 从 params 获取表名和其他可选配置
	if t, ok := config["table"]; ok && t != "" {
		s.table = t
	} else {
		return fmt.Errorf("doris sink: config is missing required key 'table'")
	}

	// 验证必要配置是否存在
	if host == "" {
		return fmt.Errorf("doris sink: missing required configuration 'host' from datasource")
	}
	if user == "" {
		return fmt.Errorf("doris sink: missing required configuration 'user' from datasource")
	}
	if database == "" {
		return fmt.Errorf("doris sink: missing required configuration 'database' from datasource")
	}
	if port == "" {
		return fmt.Errorf("doris sink: missing required configuration 'port' from datasource")
	}
	// 构建 URL
	s.url = fmt.Sprintf("%s:%s/api/%s/%s/_stream_load", host, port, database, s.table)
	s.user = user
	s.password = password

	return nil
}

type StreamLoadResponse struct {
	TxnId                  int64  `json:"TxnId"`                  // 导入的事务 ID，用户可不感知
	Label                  string `json:"Label"`                  // 导入 Label，由用户指定或系统自动生成
	Status                 string `json:"Status"`                 // 导入状态："Success"表示成功，"Publish Timeout"表示延迟可见，"Label Already Exists"表示Label重复，"Fail"表示失败
	ExistingJobStatus      string `json:"ExistingJobStatus"`      // 已存在的 Label 对应的导入作业状态："RUNNING"表示执行中，"FINISHED"表示成功
	Message                string `json:"Message"`                // 错误信息提示
	NumberTotalRows        int64  `json:"NumberTotalRows"`        // 导入总处理的行数
	NumberLoadedRows       int64  `json:"NumberLoadedRows"`       // 成功导入的行数
	NumberFilteredRows     int64  `json:"NumberFilteredRows"`     // 数据质量不合格的行数
	NumberUnselectedRows   int64  `json:"NumberUnselectedRows"`   // 被 where 条件过滤的行数
	LoadBytes              int64  `json:"LoadBytes"`              // 导入的字节数
	LoadTimeMs             int64  `json:"LoadTimeMs"`             // 导入完成时间(毫秒)
	BeginTxnTimeMs         int64  `json:"BeginTxnTimeMs"`         // 开始事务耗时(毫秒)
	StreamLoadPutTimeMs    int64  `json:"StreamLoadPutTimeMs"`    // 获取执行计划耗时(毫秒)
	ReadDataTimeMs         int64  `json:"ReadDataTimeMs"`         // 读取数据耗时(毫秒)
	WriteDataTimeMs        int64  `json:"WriteDataTimeMs"`        // 写入数据耗时(毫秒)
	CommitAndPublishTimeMs int64  `json:"CommitAndPublishTimeMs"` // 提交并发布事务耗时(毫秒)
	ErrorURL               string `json:"ErrorURL"`               // 数据质量问题详情URL
}

// Write 将一批记录通过 Doris Stream Load 方式导入。
func (s *Sink) Write(id string, records []record.Record) error {
	if len(records) == 0 {
		return nil
	}

	// 准备 JSON 数据
	var jsonData []byte
	var err error

	// 构造要发送的记录数组
	data := make([]map[string]interface{}, 0, len(records))

	for _, r := range records {
		row := make(map[string]interface{})
		for recordKey, dbCol := range s.columnMapping {
			val, exists := r[recordKey]
			if !exists || val == nil {
				row[dbCol] = nil
			} else {
				row[dbCol] = val
			}
		}
		data = append(data, row)
	}

	jsonData, err = json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("PUT", s.url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 添加认证信息和请求头
	req.SetBasicAuth(s.user, s.password)
	req.Header.Set("format", "JSON")
	req.Header.Set("Expect", "100-continue")
	req.Header.Set("strip_outer_array", "TRUE")
	req.Header.Set("label", id+"_"+strconv.FormatInt(time.Now().UnixMicro(), 10)) // 唯一标识本次导入任务

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send stream load request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream load failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, _ := io.ReadAll(resp.Body)
	var streamLoadResponse StreamLoadResponse
	err = json.Unmarshal(body, &streamLoadResponse)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stream load response: %w", err)
	}
	if streamLoadResponse.Status != "Success" {
		return fmt.Errorf("stream load failed: %s", string(body))
	}
	return nil
}

// Close 在此实现中不执行任何操作，因为没有持久连接需要关闭。
func (s *Sink) Close() error {
	return nil
}
