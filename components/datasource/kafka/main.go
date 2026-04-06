package kafka

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// DataSource 实现了 core.Datasource 接口，用于管理 Kafka 连接配置。
type DataSource struct {
	brokers   []string
	dialer    *kafkago.Dialer
	configMap map[string]string
}

var name = "kafka"

func SetCustomName(customName string) {
	name = customName
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "brokers",
			Required:     true,
			DefaultValue: "localhost:9092",
			Description:  "Kafka broker addresses, comma-separated, e.g. host1:9092,host2:9092",
		},
		{
			Key:          "sasl_mechanism",
			Required:     false,
			DefaultValue: "",
			Description:  "SASL mechanism: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512. Leave empty to disable SASL.",
		},
		{
			Key:          "sasl_username",
			Required:     false,
			DefaultValue: "",
			Description:  "SASL username",
		},
		{
			Key:          "sasl_password",
			Required:     false,
			DefaultValue: "",
			Description:  "SASL password",
		},
		{
			Key:          "tls_enabled",
			Required:     false,
			DefaultValue: "false",
			Description:  "Enable TLS for Kafka connection (true/false)",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	brokersStr, ok := config["brokers"]
	if !ok || brokersStr == "" {
		return fmt.Errorf("kafka datasource: 'brokers' is required")
	}
	d.brokers = strings.Split(brokersStr, ",")
	for i, b := range d.brokers {
		d.brokers[i] = strings.TrimSpace(b)
	}

	dialer := &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// TLS 配置
	if v := config["tls_enabled"]; strings.EqualFold(v, "true") {
		dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	// SASL 配置
	mechanism := strings.ToUpper(config["sasl_mechanism"])
	username := config["sasl_username"]
	password := config["sasl_password"]

	switch mechanism {
	case "PLAIN":
		dialer.SASLMechanism = plain.Mechanism{
			Username: username,
			Password: password,
		}
	case "SCRAM-SHA-256":
		m, err := scram.Mechanism(scram.SHA256, username, password)
		if err != nil {
			return fmt.Errorf("kafka datasource: failed to create SCRAM-SHA-256 mechanism: %w", err)
		}
		dialer.SASLMechanism = m
	case "SCRAM-SHA-512":
		m, err := scram.Mechanism(scram.SHA512, username, password)
		if err != nil {
			return fmt.Errorf("kafka datasource: failed to create SCRAM-SHA-512 mechanism: %w", err)
		}
		dialer.SASLMechanism = m
	case "":
		// 不使用 SASL
	default:
		return fmt.Errorf("kafka datasource: unsupported SASL mechanism '%s'", mechanism)
	}

	d.dialer = dialer

	// 验证连接
	conn, err := dialer.DialContext(context.Background(), "tcp", d.brokers[0])
	if err != nil {
		return fmt.Errorf("kafka datasource: failed to connect to broker '%s': %w", d.brokers[0], err)
	}
	_ = conn.Close()

	d.configMap = map[string]string{
		"brokers": brokersStr,
	}
	return nil
}

// Dialer 返回配置好的 Kafka dialer，供 source/sink 使用。
func (d *DataSource) Dialer() *kafkago.Dialer {
	return d.dialer
}

// Brokers 返回 broker 地址列表。
func (d *DataSource) Brokers() []string {
	return d.brokers
}

func (d *DataSource) DB() *sql.DB {
	return nil
}

func (d *DataSource) ConfigMap() map[string]string {
	return d.configMap
}

func (d *DataSource) Close() error {
	return nil
}

// ListTables 不适用于 Kafka，返回空列表。
func (d *DataSource) ListTables() ([]datasource.TableInfo, error) {
	return []datasource.TableInfo{}, nil
}

// AsKafkaDataSource 从 core.Datasource 接口转换为 *DataSource。
func AsKafkaDataSource(ds datasource.Datasource) (*DataSource, error) {
	kds, ok := ds.(*DataSource)
	if !ok {
		return nil, fmt.Errorf("kafka: expected *kafka.DataSource, got %T", ds)
	}
	return kds, nil
}

// NewDialerFromBrokers 从 broker 地址字符串创建一个基础 Dialer（无 SASL/TLS）。
func NewDialerFromBrokers(brokers string) ([]string, *kafkago.Dialer) {
	list := strings.Split(brokers, ",")
	for i, b := range list {
		list[i] = strings.TrimSpace(b)
	}
	return list, &kafkago.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
}
