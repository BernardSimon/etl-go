package redis

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/redis/go-redis/v9"
)

// DataSource 实现了 core.Datasource 接口，管理 Redis 连接。
type DataSource struct {
	client *redis.Client
}

var name = "redis"

func SetCustomName(customName string) {
	name = customName
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "addr",
			Required:     true,
			DefaultValue: "localhost:6379",
			Description:  "Redis server address (host:port)",
		},
		{
			Key:          "password",
			Required:     false,
			DefaultValue: "",
			Description:  "Redis password (AUTH). Leave empty if not required.",
			Mask:         true,
		},
		{
			Key:          "db",
			Required:     false,
			DefaultValue: "0",
			Description:  "Redis database number (0-15)",
		},
		{
			Key:          "tls_enabled",
			Required:     false,
			DefaultValue: "false",
			Description:  "Enable TLS for Redis connection (true/false)",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	addr := config["addr"]
	if addr == "" {
		return fmt.Errorf("redis datasource: 'addr' is required")
	}

	dbNum := 0
	if dbStr := config["db"]; dbStr != "" {
		n, err := strconv.Atoi(dbStr)
		if err != nil {
			return fmt.Errorf("redis datasource: invalid 'db' value '%s'", dbStr)
		}
		dbNum = n
	}

	opt := &redis.Options{
		Addr:     addr,
		Password: config["password"],
		DB:       dbNum,
	}

	if strings.EqualFold(config["tls_enabled"], "true") {
		opt.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	d.client = redis.NewClient(opt)

	// 验证连接
	if err := d.client.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis datasource: failed to connect to '%s': %w", addr, err)
	}
	return nil
}

// Client 返回底层 *redis.Client，供 source/sink 使用。
func (d *DataSource) Client() *redis.Client {
	return d.client
}

func (d *DataSource) DB() *sql.DB {
	return nil
}

func (d *DataSource) ConfigMap() map[string]string {
	return nil
}

func (d *DataSource) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// ListTables 不适用于 Redis，返回空列表。
func (d *DataSource) ListTables() ([]datasource.TableInfo, error) {
	return []datasource.TableInfo{}, nil
}

// AsRedisDataSource 从 core.Datasource 接口转换为 *DataSource。
func AsRedisDataSource(ds datasource.Datasource) (*DataSource, error) {
	rds, ok := ds.(*DataSource)
	if !ok {
		return nil, fmt.Errorf("redis: expected *redis.DataSource, got %T", ds)
	}
	return rds, nil
}
