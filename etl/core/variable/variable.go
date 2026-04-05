package variable

import (
	"context"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
)

type VariableCreator func() (name string, variable Variable, datasource *string, params []params.Params)
type Variable interface {
	Get(ctx context.Context, config map[string]string, datasource datasource.Datasource) (string, error)
}
