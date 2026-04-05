package datasource

import (
	"database/sql"

	"github.com/BernardSimon/etl-go/etl/core/params"
)

type DatasourceCreator func() (name string, datasource Datasource, params []params.Params)

type Datasource interface {
	Init(map[string]string) error
	Close() error
}

type SQLDBProvider interface {
	DB() *sql.DB
}

type ConfigMapProvider interface {
	ConfigMap() map[string]string
}
