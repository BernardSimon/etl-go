package source

import (
	"context"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

type SourceCreator func() (name string, source Source, datasource *string, params []params.Params)
type Source interface {
	Column() map[string]string
	Open(ctx context.Context, config map[string]string, dataSource datasource.Datasource) error
	Read(ctx context.Context) (record.Record, error)
	Close() error
}
