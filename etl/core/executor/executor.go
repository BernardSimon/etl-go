package executor

import (
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
)

type ExecutorCreator func() (name string, executor Executor, datasource *string, params []params.Params)
type Executor interface {
	Open(config map[string]string, dataSource *datasource.Datasource) error
	Close() error
}
