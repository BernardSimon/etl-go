package redis

import (
	"context"
	"encoding/json"
	"fmt"

	redisDatasource "github.com/BernardSimon/etl-go/components/datasource/redis"
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/redis/go-redis/v9"
)

var name = "redis"
var datasourceName = "redis"

func SetCustomName(customName string, customDatasourceName string) {
	name = customName
	datasourceName = customDatasourceName
}

// Sink 实现了 core.Sink 接口，将 Record 写入 Redis。
//
// 支持三种模式：
//   - hash: 每条 record 写为一个 Hash，key = key_prefix + record[key_field]
//   - list: 每条 record 序列化为 JSON 后 RPUSH 到指定 list key
//   - string: 每条 record 以 SET key value 写入，key = key_prefix + record[key_field]，value = record[value_field]
type Sink struct {
	client        *redis.Client
	mode          string
	keyField      string
	keyPrefix     string
	listKey       string
	valueField    string
	columnMapping map[string]string
}

func SinkCreator() (string, sink.Sink, *string, []params.Params) {
	return name, &Sink{}, &datasourceName, []params.Params{
		{
			Key:          "mode",
			Required:     false,
			DefaultValue: "hash",
			Description:  "Write mode: hash (HSET), list (RPUSH as JSON), string (SET key value)",
		},
		{
			Key:          "key_field",
			Required:     false,
			DefaultValue: "",
			Description:  "Field name whose value is used as the Redis key (hash and string modes). Required for hash/string modes.",
		},
		{
			Key:          "key_prefix",
			Required:     false,
			DefaultValue: "",
			Description:  "Prefix prepended to the Redis key (hash and string modes), e.g. 'user:'",
		},
		{
			Key:          "key",
			Required:     false,
			DefaultValue: "",
			Description:  "Target Redis list key (list mode). Required for list mode.",
		},
		{
			Key:          "value_field",
			Required:     false,
			DefaultValue: "value",
			Description:  "Field name whose value is used as the Redis string value (string mode).",
		},
	}
}

func (s *Sink) Open(ctx context.Context, config map[string]string, columnMapping map[string]string, ds datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	rds, err := redisDatasource.AsRedisDataSource(ds)
	if err != nil {
		return fmt.Errorf("redis sink: %w", err)
	}
	s.client = rds.Client()
	s.columnMapping = columnMapping

	s.mode = "hash"
	if m, ok := config["mode"]; ok && m != "" {
		switch m {
		case "hash", "list", "string":
			s.mode = m
		default:
			return fmt.Errorf("redis sink: unsupported mode '%s'", m)
		}
	}

	s.keyPrefix = config["key_prefix"]
	s.keyField = config["key_field"]
	s.listKey = config["key"]
	s.valueField = "value"
	if vf, ok := config["value_field"]; ok && vf != "" {
		s.valueField = vf
	}

	switch s.mode {
	case "hash", "string":
		if s.keyField == "" {
			return fmt.Errorf("redis sink: 'key_field' is required for mode '%s'", s.mode)
		}
	case "list":
		if s.listKey == "" {
			return fmt.Errorf("redis sink: 'key' is required for list mode")
		}
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

	pipe := s.client.Pipeline()

	for _, r := range records {
		mapped := s.applyMapping(r)

		switch s.mode {
		case "hash":
			if err := s.writeHash(ctx, pipe, mapped); err != nil {
				return err
			}
		case "list":
			if err := s.writeList(ctx, pipe, mapped); err != nil {
				return err
			}
		case "string":
			if err := s.writeString(ctx, pipe, mapped); err != nil {
				return err
			}
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis sink: pipeline execution failed: %w", err)
	}
	return nil
}

func (s *Sink) writeHash(ctx context.Context, pipe redis.Pipeliner, r map[string]interface{}) error {
	keyVal, ok := r[s.keyField]
	if !ok || keyVal == nil {
		return fmt.Errorf("redis sink: key_field '%s' not found in record", s.keyField)
	}
	redisKey := s.keyPrefix + fmt.Sprintf("%v", keyVal)

	fields := make(map[string]interface{}, len(r))
	for k, v := range r {
		if k == s.keyField {
			continue
		}
		fields[k] = v
	}
	if len(fields) == 0 {
		return nil
	}
	pipe.HSet(ctx, redisKey, fields)
	return nil
}

func (s *Sink) writeList(ctx context.Context, pipe redis.Pipeliner, r map[string]interface{}) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("redis sink: failed to marshal record as JSON: %w", err)
	}
	pipe.RPush(ctx, s.listKey, string(data))
	return nil
}

func (s *Sink) writeString(ctx context.Context, pipe redis.Pipeliner, r map[string]interface{}) error {
	keyVal, ok := r[s.keyField]
	if !ok || keyVal == nil {
		return fmt.Errorf("redis sink: key_field '%s' not found in record", s.keyField)
	}
	redisKey := s.keyPrefix + fmt.Sprintf("%v", keyVal)

	val, ok := r[s.valueField]
	if !ok {
		val = ""
	}
	pipe.Set(ctx, redisKey, fmt.Sprintf("%v", val), 0)
	return nil
}

func (s *Sink) applyMapping(r record.Record) map[string]interface{} {
	result := make(map[string]interface{}, len(r))
	if len(s.columnMapping) > 0 {
		for recordKey, targetKey := range s.columnMapping {
			val, exists := r[recordKey]
			if !exists {
				val = nil
			}
			result[targetKey] = val
		}
	} else {
		for k, v := range r {
			result[k] = v
		}
	}
	return result
}

func (s *Sink) Close() error {
	return nil
}
