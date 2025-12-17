package factory

import (
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/BernardSimon/etl-go/etl/core/variable"
	"go.uber.org/zap"
)

var (
	sourceRegistry     = make(map[string]*SourceStore)
	processorRegistry  = make(map[string]*ProcessorStore)
	sinkRegistry       = make(map[string]*SinkStore)
	executorRegistry   = make(map[string]*ExecutorStore)
	variableRegistry   = make(map[string]*VariableStore)
	datasourceRegistry = make(map[string]*DatasourceStore)
)

func RegisterSource(creator source.SourceCreator) {
	n, s, d, p := creator()
	if _, exists := sourceRegistry[n]; exists {
		zap.L().Fatal(fmt.Sprintf("FATAL: Source with name '%s' is already registered", n), zap.String("service", "etl"), zap.String("name", n))
	}
	ss := SourceStore{
		Name:   n,
		Handle: s,
		Params: p,
	}
	if d != nil {
		_, err := CreateDataSource(*d)
		if err != nil {
			zap.L().Fatal("FATAL: DataSource with name '%s' has not registered", zap.String("service", "etl"), zap.String("name", n))
		}
		ss.Datasource = d
	}
	sourceRegistry[n] = &ss
}
func RegisterProcessor(creator procrssor.ProcessorCreator) {
	n, p, pa := creator()
	if _, exists := processorRegistry[n]; exists {
		zap.L().Fatal("FATAL: Processor with name '%s' is already registered", zap.String("service", "etl"), zap.String("name", n))
	}
	processorRegistry[n] = &ProcessorStore{
		Name:   n,
		Handle: p,
		Params: pa,
	}
}
func RegisterSink(creator sink.SinkCreator) {
	n, s, d, p := creator()
	if _, exists := sinkRegistry[n]; exists {
		zap.L().Fatal("FATAL: Sink with name '%s' is already registered", zap.String("service", "etl"), zap.String("name", n))
	}
	ss := SinkStore{
		Name:   n,
		Handle: s,
		Params: p,
	}
	if d != nil {
		_, err := CreateDataSource(*d)
		if err != nil {
			zap.L().Fatal("FATAL: DataSource with name '%s' has not registered", zap.String("service", "etl"), zap.String("name", n))
		}
		ss.Datasource = d
	}
	sinkRegistry[n] = &ss
}
func RegisterExecutor(creator executor.ExecutorCreator) {
	n, e, d, p := creator()
	if _, exists := executorRegistry[n]; exists {
		zap.L().Fatal("FATAL: Executor with name '%s' is already registered", zap.String("service", "etl"), zap.String("name", n))
	}
	es := ExecutorStore{
		Name:   n,
		Handle: e,
		Params: p,
	}
	if d != nil {
		_, err := CreateDataSource(*d)
		if err != nil {
			zap.L().Fatal("FATAL: DataSource with name '%s' has not registered", zap.String("service", "etl"), zap.String("name", n))
		}
		es.Datasource = d
	}
	executorRegistry[n] = &es
}
func RegisterVariable(creator variable.VariableCreator) {
	n, v, d, p := creator()
	if _, exists := variableRegistry[n]; exists {
		zap.L().Fatal("FATAL: Variable with name '%s' is already registered", zap.String("service", "etl"), zap.String("name", n))
	}
	vs := VariableStore{
		Name:   n,
		Handle: v,
		Params: p,
	}
	if d != nil {
		_, err := CreateDataSource(*d)
		if err != nil {
			zap.L().Fatal("FATAL: DataSource with name '%s' has not registered", zap.String("service", "etl"), zap.String("name", n))
		}
		vs.Datasource = d
	}
	variableRegistry[n] = &vs
}

func RegisterDataSource(creator datasource.DatasourceCreator) {
	n, d, p := creator()
	if _, exists := datasourceRegistry[n]; exists {
		zap.L().Fatal("FATAL: DataSource with name '%s' is already registered", zap.String("service", "etl"), zap.String("name", n))
	}
	datasourceRegistry[n] = &DatasourceStore{
		Name:   n,
		Handle: d,
		Params: p,
	}
}
func CreateSource(name string) (SourceStore, error) {
	creator, ok := sourceRegistry[name]
	if !ok {
		return SourceStore{}, fmt.Errorf("factory error: no source registered with name: %s", name)
	}
	return *creator, nil
}

func CreateProcessor(name string) (ProcessorStore, error) {
	store, ok := processorRegistry[name]
	if !ok {
		return ProcessorStore{}, fmt.Errorf("factory error: no processor registered with name: %s", name)
	}
	return *store, nil
}

func CreateSink(name string) (SinkStore, error) {
	store, ok := sinkRegistry[name]
	if !ok {
		return SinkStore{}, fmt.Errorf("factory error: no sink registered with name: %s", name)
	}
	return *store, nil
}

func CreateExecutor(name string) (ExecutorStore, error) {
	store, ok := executorRegistry[name]
	if !ok {
		return ExecutorStore{}, fmt.Errorf("factory error: no Executor registered with name: %s", name)
	}
	return *store, nil
}

func CreateVariable(name string) (VariableStore, error) {
	store, ok := variableRegistry[name]
	if !ok {
		return VariableStore{}, fmt.Errorf("factory error: no Variable registered with name: %s", name)
	}
	return *store, nil
}

func CreateDataSource(name string) (DatasourceStore, error) {
	store, ok := datasourceRegistry[name]
	if !ok {
		return DatasourceStore{}, fmt.Errorf("factory error: no DataSource registered with name: %s", name)
	}
	return *store, nil
}

func GetDatasourceTypeList() []string {
	var types []string
	for k := range datasourceRegistry {
		types = append(types, k)
	}
	return types
}
func GetExecutorTypeList() []string {
	var types []string
	for k := range executorRegistry {
		types = append(types, k)
	}
	return types
}
func GetProcessorTypeList() []string {
	var types []string
	for k := range processorRegistry {
		types = append(types, k)
	}
	return types
}
func GetSinkTypeList() []string {
	var types []string
	for k := range sinkRegistry {
		types = append(types, k)
	}
	return types
}
func GetSourceTypeList() []string {
	var types []string
	for k := range sourceRegistry {
		types = append(types, k)
	}
	return types
}
func GetVariableTypeList() []string {
	var types []string
	for k := range variableRegistry {
		types = append(types, k)
	}
	return types
}
