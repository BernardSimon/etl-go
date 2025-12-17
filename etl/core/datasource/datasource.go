package datasource

import "github.com/BernardSimon/etl-go/etl/core/params"

type DatasourceCreator func() (name string, datasource Datasource, params []params.Params)

type Datasource interface {
	Init(map[string]string) error
	Open() any
	Close() error
}
