package procrssor

import (
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

type ProcessorCreator func() (name string, processor Processor, params []params.Params)

type Processor interface {
	Open(config map[string]string) error
	Process(record record.Record) (record.Record, error)
	Close() error
	HandleColumns(columns *map[string]string)
}
