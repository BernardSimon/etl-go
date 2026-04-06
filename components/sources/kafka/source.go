package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	kafkaDatasource "github.com/BernardSimon/etl-go/components/datasource/kafka"
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
	kafkago "github.com/segmentio/kafka-go"
)

var name = "kafka"
var datasourceName = "kafka"

func SetCustomName(customName string, customDatasourceName string) {
	name = customName
	datasourceName = customDatasourceName
}

// Source 实现了 core.Source 接口，从 Kafka topic 消费消息。
type Source struct {
	reader      *kafkago.Reader
	topic       string
	valueFormat string // json | string
	keyField    string // 非空时，将消息 key 存入该字段
	maxMessages int    // 0 表示不限制（读到 timeout 为止）
	timeout     time.Duration
	count       int    // 已读取消息数
	columns     map[string]string
}

func SourceCreator() (string, source.Source, *string, []params.Params) {
	return name, &Source{}, &datasourceName, []params.Params{
		{
			Key:          "topic",
			Required:     true,
			DefaultValue: "",
			Description:  "Kafka topic to consume",
		},
		{
			Key:          "group_id",
			Required:     false,
			DefaultValue: "etl-go-consumer",
			Description:  "Kafka consumer group ID",
		},
		{
			Key:          "value_format",
			Required:     false,
			DefaultValue: "json",
			Description:  "Message value format: json (parse JSON into fields) or string (store raw value in 'value' field)",
		},
		{
			Key:          "key_field",
			Required:     false,
			DefaultValue: "",
			Description:  "If set, message key will be stored in this field name",
		},
		{
			Key:          "max_messages",
			Required:     false,
			DefaultValue: "0",
			Description:  "Maximum number of messages to read. 0 means read until timeout.",
		},
		{
			Key:          "timeout_seconds",
			Required:     false,
			DefaultValue: "30",
			Description:  "Timeout in seconds waiting for each message. Source ends when timeout is reached.",
		},
		{
			Key:          "start_offset",
			Required:     false,
			DefaultValue: "earliest",
			Description:  "Where to start consuming: earliest (from beginning) or latest (new messages only)",
		},
	}
}

func (s *Source) Open(ctx context.Context, config map[string]string, ds datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	topic, ok := config["topic"]
	if !ok || topic == "" {
		return fmt.Errorf("kafka source: 'topic' is required")
	}
	s.topic = topic

	groupID := "etl-go-consumer"
	if g, ok := config["group_id"]; ok && g != "" {
		groupID = g
	}

	s.valueFormat = "json"
	if vf, ok := config["value_format"]; ok && vf != "" {
		switch vf {
		case "json", "string":
			s.valueFormat = vf
		default:
			return fmt.Errorf("kafka source: unsupported value_format '%s', expected 'json' or 'string'", vf)
		}
	}

	s.keyField = config["key_field"]

	s.maxMessages = 0
	if mm, ok := config["max_messages"]; ok && mm != "" {
		n, err := strconv.Atoi(mm)
		if err != nil || n < 0 {
			return fmt.Errorf("kafka source: invalid max_messages '%s'", mm)
		}
		s.maxMessages = n
	}

	s.timeout = 30 * time.Second
	if ts, ok := config["timeout_seconds"]; ok && ts != "" {
		n, err := strconv.Atoi(ts)
		if err != nil || n <= 0 {
			return fmt.Errorf("kafka source: invalid timeout_seconds '%s'", ts)
		}
		s.timeout = time.Duration(n) * time.Second
	}

	startOffset := kafkago.FirstOffset
	if so, ok := config["start_offset"]; ok && so == "latest" {
		startOffset = kafkago.LastOffset
	}

	// 从 datasource 获取连接配置
	kds, err := kafkaDatasource.AsKafkaDataSource(ds)
	if err != nil {
		return fmt.Errorf("kafka source: %w", err)
	}

	readerCfg := kafkago.ReaderConfig{
		Brokers:     kds.Brokers(),
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: startOffset,
		Dialer:      kds.Dialer(),
		MinBytes:    1,
		MaxBytes:    10e6, // 10 MB
		MaxWait:     s.timeout,
	}

	s.reader = kafkago.NewReader(readerCfg)
	s.count = 0
	s.columns = make(map[string]string)

	return nil
}

func (s *Source) Read(ctx context.Context) (record.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// 已达最大消息数
	if s.maxMessages > 0 && s.count >= s.maxMessages {
		return nil, io.EOF
	}

	// 设置读取超时
	readCtx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	msg, err := s.reader.ReadMessage(readCtx)
	if err != nil {
		// 超时视为正常结束
		if err == context.DeadlineExceeded {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("kafka source: failed to read message: %w", err)
	}

	r, err := s.messageToRecord(msg)
	if err != nil {
		return nil, err
	}

	// 更新已知列
	for k := range r {
		s.columns[k] = k
	}

	s.count++
	return r, nil
}

func (s *Source) messageToRecord(msg kafkago.Message) (record.Record, error) {
	r := make(record.Record)

	switch s.valueFormat {
	case "json":
		var obj map[string]interface{}
		if err := json.Unmarshal(msg.Value, &obj); err != nil {
			return nil, fmt.Errorf("kafka source: failed to parse message as JSON (offset %d): %w", msg.Offset, err)
		}
		for k, v := range obj {
			r[k] = v
		}
	case "string":
		r["value"] = string(msg.Value)
	}

	if s.keyField != "" && len(msg.Key) > 0 {
		r[s.keyField] = string(msg.Key)
	}

	return r, nil
}

func (s *Source) Column() map[string]string {
	return s.columns
}

func (s *Source) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return nil
}
