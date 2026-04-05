package factory

import (
	"fmt"
	"sort"
	"sync"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/processor"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/BernardSimon/etl-go/etl/core/variable"
)

var (
	sourceRegistry     = make(map[string]*SourceStore)
	processorRegistry  = make(map[string]*ProcessorStore)
	sinkRegistry       = make(map[string]*SinkStore)
	executorRegistry   = make(map[string]*ExecutorStore)
	variableRegistry   = make(map[string]*VariableStore)
	datasourceRegistry = make(map[string]*DatasourceStore)
	registryMu         sync.RWMutex
)

func RegisterSource(creator source.SourceCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, s, d, p := creator()
	if _, exists := sourceRegistry[n]; exists {
		return fmt.Errorf("source with name '%s' is already registered", n)
	}
	ss := SourceStore{
		Name:   n,
		Handle: s,
		Params: p,
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by source '%s' has not been registered", *d, n)
		}
		ss.Datasource = d
	}
	sourceRegistry[n] = &ss
	return nil
}
func RegisterProcessor(creator processor.ProcessorCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, p, pa := creator()
	if _, exists := processorRegistry[n]; exists {
		return fmt.Errorf("processor with name '%s' is already registered", n)
	}
	processorRegistry[n] = &ProcessorStore{
		Name:   n,
		Handle: p,
		Params: pa,
	}
	return nil
}
func RegisterSink(creator sink.SinkCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, s, d, p := creator()
	if _, exists := sinkRegistry[n]; exists {
		return fmt.Errorf("sink with name '%s' is already registered", n)
	}
	ss := SinkStore{
		Name:   n,
		Handle: s,
		Params: p,
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by sink '%s' has not been registered", *d, n)
		}
		ss.Datasource = d
	}
	sinkRegistry[n] = &ss
	return nil
}
func RegisterExecutor(creator executor.ExecutorCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, e, d, p := creator()
	if _, exists := executorRegistry[n]; exists {
		return fmt.Errorf("executor with name '%s' is already registered", n)
	}
	es := ExecutorStore{
		Name:   n,
		Handle: e,
		Params: p,
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by executor '%s' has not been registered", *d, n)
		}
		es.Datasource = d
	}
	executorRegistry[n] = &es
	return nil
}
func RegisterVariable(creator variable.VariableCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, v, d, p := creator()
	if _, exists := variableRegistry[n]; exists {
		return fmt.Errorf("variable with name '%s' is already registered", n)
	}
	vs := VariableStore{
		Name:   n,
		Handle: v,
		Params: p,
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by variable '%s' has not been registered", *d, n)
		}
		vs.Datasource = d
	}
	variableRegistry[n] = &vs
	return nil
}

func RegisterDataSource(creator datasource.DatasourceCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, d, p := creator()
	if _, exists := datasourceRegistry[n]; exists {
		return fmt.Errorf("datasource with name '%s' is already registered", n)
	}
	datasourceRegistry[n] = &DatasourceStore{
		Name:   n,
		Handle: d,
		Params: p,
	}
	return nil
}
func CreateSource(name string) (SourceStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := sourceRegistry[name]
	if !ok {
		return SourceStore{}, fmt.Errorf("factory error: no source registered with name: %s", name)
	}
	return *creator, nil
}

func CreateProcessor(name string) (ProcessorStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	store, ok := processorRegistry[name]
	if !ok {
		return ProcessorStore{}, fmt.Errorf("factory error: no processor registered with name: %s", name)
	}
	return *store, nil
}

func CreateSink(name string) (SinkStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	store, ok := sinkRegistry[name]
	if !ok {
		return SinkStore{}, fmt.Errorf("factory error: no sink registered with name: %s", name)
	}
	return *store, nil
}

func CreateExecutor(name string) (ExecutorStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	store, ok := executorRegistry[name]
	if !ok {
		return ExecutorStore{}, fmt.Errorf("factory error: no Executor registered with name: %s", name)
	}
	return *store, nil
}

func CreateVariable(name string) (VariableStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	store, ok := variableRegistry[name]
	if !ok {
		return VariableStore{}, fmt.Errorf("factory error: no Variable registered with name: %s", name)
	}
	return *store, nil
}

func CreateDataSource(name string) (DatasourceStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	store, ok := datasourceRegistry[name]
	if !ok {
		return DatasourceStore{}, fmt.Errorf("factory error: no DataSource registered with name: %s", name)
	}
	return *store, nil
}

func GetDatasourceTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range datasourceRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
func GetExecutorTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range executorRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
func GetProcessorTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range processorRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
func GetSinkTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range sinkRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
func GetSourceTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range sourceRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
func GetVariableTypeList() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	var types []string
	for k := range variableRegistry {
		types = append(types, k)
	}
	sort.Strings(types)
	return types
}
