package sink

import (
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

type SinkCreator func() (name string, sink Sink, datasource *string, params []params.Params)

type Sink interface {
	Open(config map[string]string, columnMapping map[string]string, dataSource *datasource.Datasource) error
	Write(id string, records []record.Record) error
	Close() error
}
