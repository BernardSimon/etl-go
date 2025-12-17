package factory

import (
	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/BernardSimon/etl-go/etl/core/variable"
)

type SourceStore struct {
	Name       string
	Handle     source.Source
	Datasource *string
	Params     []params.Params
}

type SinkStore struct {
	Name       string
	Handle     sink.Sink
	Datasource *string
	Params     []params.Params
}

type ExecutorStore struct {
	Name       string
	Handle     executor.Executor
	Datasource *string
	Params     []params.Params
}
type VariableStore struct {
	Name       string
	Handle     variable.Variable
	Datasource *string
	Params     []params.Params
}
type ProcessorStore struct {
	Name   string
	Handle procrssor.Processor
	Params []params.Params
}
type DatasourceStore struct {
	Name   string
	Handle datasource.Datasource
	Params []params.Params
}
