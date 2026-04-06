package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkaDatasource "github.com/BernardSimon/etl-go/components/datasource/kafka"
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	kafkago "github.com/segmentio/kafka-go"
)

var name = "kafka"
var datasourceName = "kafka"

func SetCustomName(customName string, customDatasourceName string) {
	name = customName
	datasourceName = customDatasourceName
}

// Sink 实现了 core.Sink 接口，将 Record 作为消息写入 Kafka topic。
type Sink struct {
	writer        *kafkago.Writer
	topic         string
	valueFormat   string // json | string
	keyField      string // 非空时从 Record 中取该字段作为消息 key
	columnMapping map[string]string
}

func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return name, &Sink{}, &datasourceName, []params.Params{
		{
			Key:          "topic",
			Required:     true,
			DefaultValue: "",
			Description:  "Kafka topic to produce messages to",
		},
		{
			Key:          "value_format",
			Required:     false,
			DefaultValue: "json",
			Description:  "Message value format: json (serialize record as JSON object) or string (use 'value' field as raw string)",
		},
		{
			Key:          "key_field",
			Required:     false,
			DefaultValue: "",
			Description:  "If set, the value of this field will be used as the Kafka message key",
		},
	}
}

func (s *Sink) Open(ctx context.Context, config map[string]string, columnMapping map[string]string, ds datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	topic, ok := config["topic"]
	if !ok || topic == "" {
		return fmt.Errorf("kafka sink: 'topic' is required")
	}
	s.topic = topic

	s.valueFormat = "json"
	if vf, ok := config["value_format"]; ok && vf != "" {
		switch vf {
		case "json", "string":
			s.valueFormat = vf
		default:
			return fmt.Errorf("kafka sink: unsupported value_format '%s', expected 'json' or 'string'", vf)
		}
	}

	s.keyField = config["key_field"]
	s.columnMapping = columnMapping

	// 从 datasource 获取连接配置
	kds, err := kafkaDatasource.AsKafkaDataSource(ds)
	if err != nil {
		return fmt.Errorf("kafka sink: %w", err)
	}

	transport := &kafkago.Transport{
		Dial: kds.Dialer().DialFunc,
	}
	s.writer = &kafkago.Writer{
		Addr:      kafkago.TCP(kds.Brokers()...),
		Topic:     topic,
		Balancer:  &kafkago.LeastBytes{},
		Transport: transport,
	}

	return nil
}

func (s *Sink) Write(ctx context.Context, _ string, records []record.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	messages := make([]kafkago.Message, 0, len(records))
	for _, r := range records {
		msg, err := s.recordToMessage(r)
		if err != nil {
			return err
		}
		messages = append(messages, msg)
	}

	if err := s.writer.WriteMessages(ctx, messages...); err != nil {
		return fmt.Errorf("kafka sink: failed to write messages: %w", err)
	}
	return nil
}

func (s *Sink) recordToMessage(r record.Record) (kafkago.Message, error) {
	// 应用 column_mapping
	mapped := make(map[string]interface{})
	if len(s.columnMapping) > 0 {
		for recordKey, targetKey := range s.columnMapping {
			val, exists := r[recordKey]
			if !exists {
				val = nil
			}
			mapped[targetKey] = val
		}
	} else {
		for k, v := range r {
			mapped[k] = v
		}
	}

	var value []byte
	var err error

	switch s.valueFormat {
	case "json":
		value, err = json.Marshal(mapped)
		if err != nil {
			return kafkago.Message{}, fmt.Errorf("kafka sink: failed to marshal record as JSON: %w", err)
		}
	case "string":
		// 取 "value" 字段作为消息内容
		if v, ok := mapped["value"]; ok && v != nil {
			value = []byte(fmt.Sprintf("%v", v))
		}
	}

	msg := kafkago.Message{Value: value}

	// 设置消息 key
	if s.keyField != "" {
		if v, ok := r[s.keyField]; ok && v != nil {
			msg.Key = []byte(fmt.Sprintf("%v", v))
		}
	}

	return msg, nil
}

func (s *Sink) Close() error {
	if s.writer != nil {
		return s.writer.Close()
	}
	return nil
}
