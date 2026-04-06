package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	redisDatasource "github.com/BernardSimon/etl-go/components/datasource/redis"
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/redis/go-redis/v9"
)

var name = "redis"
var datasourceName = "redis"

func SetCustomName(customName string, customDatasourceName string) {
	name = customName
	datasourceName = customDatasourceName
}

// Source 实现了 core.Source 接口，从 Redis 读取数据。
//
// 支持三种模式：
//   - hash_scan: 按 key 模式扫描，HGETALL 每个 Hash key，每个 key 生成一条 record
//   - list:      LRANGE 指定 list key，每个元素解析为 JSON record（或原始字符串）
//   - string_scan: 按 key 模式扫描，GET 每个 key，生成 {"key": k, "value": v} record
type Source struct {
	client  *redis.Client
	mode    string
	buffer  []record.Record
	pos     int
	columns map[string]string
}

func SourceCreator() (string, source.Source, *string, []params.Params) {
	return name, &Source{}, &datasourceName, []params.Params{
		{
			Key:          "mode",
			Required:     false,
			DefaultValue: "hash_scan",
			Description:  "Read mode: hash_scan (scan keys + HGETALL), list (LRANGE as JSON), string_scan (scan keys + GET)",
		},
		{
			Key:          "scan_match",
			Required:     false,
			DefaultValue: "*",
			Description:  "Key pattern for scan modes (glob-style, e.g. 'user:*'). Used by hash_scan and string_scan.",
		},
		{
			Key:          "scan_count",
			Required:     false,
			DefaultValue: "100",
			Description:  "Hint for number of keys to return per SCAN call (used by hash_scan and string_scan).",
		},
		{
			Key:          "key",
			Required:     false,
			DefaultValue: "",
			Description:  "Redis key to read from. Required for list mode.",
		},
		{
			Key:          "list_start",
			Required:     false,
			DefaultValue: "0",
			Description:  "LRANGE start index (list mode). 0 = beginning.",
		},
		{
			Key:          "list_stop",
			Required:     false,
			DefaultValue: "-1",
			Description:  "LRANGE stop index (list mode). -1 = end.",
		},
		{
			Key:          "value_format",
			Required:     false,
			DefaultValue: "json",
			Description:  "Value format for list mode: json (parse each element as JSON) or string (store raw string in 'value' field).",
		},
	}
}

func (s *Source) Open(ctx context.Context, config map[string]string, ds datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	rds, err := redisDatasource.AsRedisDataSource(ds)
	if err != nil {
		return fmt.Errorf("redis source: %w", err)
	}
	s.client = rds.Client()

	mode := "hash_scan"
	if m, ok := config["mode"]; ok && m != "" {
		switch m {
		case "hash_scan", "list", "string_scan":
			mode = m
		default:
			return fmt.Errorf("redis source: unsupported mode '%s'", m)
		}
	}
	s.mode = mode
	s.columns = make(map[string]string)

	switch mode {
	case "hash_scan":
		if err := s.loadHashScan(ctx, config); err != nil {
			return err
		}
	case "string_scan":
		if err := s.loadStringScan(ctx, config); err != nil {
			return err
		}
	case "list":
		if err := s.loadList(ctx, config); err != nil {
			return err
		}
	}

	// 收集列名
	for _, r := range s.buffer {
		for k := range r {
			s.columns[k] = k
		}
	}

	s.pos = 0
	return nil
}

func (s *Source) loadHashScan(ctx context.Context, config map[string]string) error {
	match := "*"
	if m, ok := config["scan_match"]; ok && m != "" {
		match = m
	}
	scanCount := int64(100)
	if sc, ok := config["scan_count"]; ok && sc != "" {
		if n, err := strconv.ParseInt(sc, 10, 64); err == nil && n > 0 {
			scanCount = n
		}
	}

	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := s.client.Scan(ctx, cursor, match, scanCount).Result()
		if err != nil {
			return fmt.Errorf("redis source: SCAN failed: %w", err)
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	s.buffer = make([]record.Record, 0, len(keys))
	for _, key := range keys {
		fields, err := s.client.HGetAll(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis source: HGETALL '%s' failed: %w", key, err)
		}
		r := make(record.Record, len(fields)+1)
		r["_key"] = key
		for f, v := range fields {
			r[f] = v
		}
		s.buffer = append(s.buffer, r)
	}
	return nil
}

func (s *Source) loadStringScan(ctx context.Context, config map[string]string) error {
	match := "*"
	if m, ok := config["scan_match"]; ok && m != "" {
		match = m
	}
	scanCount := int64(100)
	if sc, ok := config["scan_count"]; ok && sc != "" {
		if n, err := strconv.ParseInt(sc, 10, 64); err == nil && n > 0 {
			scanCount = n
		}
	}

	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := s.client.Scan(ctx, cursor, match, scanCount).Result()
		if err != nil {
			return fmt.Errorf("redis source: SCAN failed: %w", err)
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	s.buffer = make([]record.Record, 0, len(keys))
	for _, key := range keys {
		val, err := s.client.Get(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis source: GET '%s' failed: %w", key, err)
		}
		s.buffer = append(s.buffer, record.Record{
			"key":   key,
			"value": val,
		})
	}
	return nil
}

func (s *Source) loadList(ctx context.Context, config map[string]string) error {
	key, ok := config["key"]
	if !ok || key == "" {
		return fmt.Errorf("redis source: 'key' is required for list mode")
	}

	start := int64(0)
	if sv, ok := config["list_start"]; ok && sv != "" {
		if n, err := strconv.ParseInt(sv, 10, 64); err == nil {
			start = n
		}
	}
	stop := int64(-1)
	if sv, ok := config["list_stop"]; ok && sv != "" {
		if n, err := strconv.ParseInt(sv, 10, 64); err == nil {
			stop = n
		}
	}

	valueFormat := "json"
	if vf, ok := config["value_format"]; ok && vf != "" {
		valueFormat = vf
	}

	elements, err := s.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return fmt.Errorf("redis source: LRANGE '%s' failed: %w", key, err)
	}

	s.buffer = make([]record.Record, 0, len(elements))
	for _, elem := range elements {
		var r record.Record
		if valueFormat == "json" {
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(elem), &obj); err != nil {
				return fmt.Errorf("redis source: failed to parse list element as JSON: %w", err)
			}
			r = record.Record(obj)
		} else {
			r = record.Record{"value": elem}
		}
		s.buffer = append(s.buffer, r)
	}
	return nil
}

func (s *Source) Read(ctx context.Context) (record.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if s.pos >= len(s.buffer) {
		return nil, io.EOF
	}
	r := s.buffer[s.pos]
	s.pos++
	return r, nil
}

func (s *Source) Column() map[string]string {
	return s.columns
}

func (s *Source) Close() error {
	s.buffer = nil
	return nil
}
