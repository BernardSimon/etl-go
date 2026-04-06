package datasource

import (
	"database/sql"

	"github.com/BernardSimon/etl-go/etl/core/params"
)

type DatasourceCreator func() (name string, datasource Datasource, params []params.Params)

type Datasource interface {
	Init(map[string]string) error
	Close() error
	// DB 返回底层 *sql.DB，不支持 SQL 的实现返回 nil。
	DB() *sql.DB
	// ConfigMap 返回配置键值对，不需要配置映射的实现返回 nil。
	ConfigMap() map[string]string
	// ListTables 返回所有表及列信息，不支持 schema 发现的实现返回空切片。
	ListTables() ([]TableInfo, error)
}
