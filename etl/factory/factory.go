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
	sourceRegistry     = make(map[string]source.SourceCreator)
	processorRegistry  = make(map[string]processor.ProcessorCreator)
	sinkRegistry       = make(map[string]sink.SinkCreator)
	executorRegistry   = make(map[string]executor.ExecutorCreator)
	variableRegistry   = make(map[string]variable.VariableCreator)
	datasourceRegistry = make(map[string]datasource.DatasourceCreator)
	registryMu         sync.RWMutex
)

func RegisterSource(creator source.SourceCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, d, _ := creator()
	if _, exists := sourceRegistry[n]; exists {
		return fmt.Errorf("source with name '%s' is already registered", n)
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by source '%s' has not been registered", *d, n)
		}
	}
	sourceRegistry[n] = creator
	return nil
}
func RegisterProcessor(creator processor.ProcessorCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, _ := creator()
	if _, exists := processorRegistry[n]; exists {
		return fmt.Errorf("processor with name '%s' is already registered", n)
	}
	processorRegistry[n] = creator
	return nil
}
func RegisterSink(creator sink.SinkCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, d, _ := creator()
	if _, exists := sinkRegistry[n]; exists {
		return fmt.Errorf("sink with name '%s' is already registered", n)
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by sink '%s' has not been registered", *d, n)
		}
	}
	sinkRegistry[n] = creator
	return nil
}
func RegisterExecutor(creator executor.ExecutorCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, d, _ := creator()
	if _, exists := executorRegistry[n]; exists {
		return fmt.Errorf("executor with name '%s' is already registered", n)
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by executor '%s' has not been registered", *d, n)
		}
	}
	executorRegistry[n] = creator
	return nil
}
func RegisterVariable(creator variable.VariableCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, d, _ := creator()
	if _, exists := variableRegistry[n]; exists {
		return fmt.Errorf("variable with name '%s' is already registered", n)
	}
	if d != nil {
		if _, exists := datasourceRegistry[*d]; !exists {
			return fmt.Errorf("datasource '%s' required by variable '%s' has not been registered", *d, n)
		}
	}
	variableRegistry[n] = creator
	return nil
}

func RegisterDataSource(creator datasource.DatasourceCreator) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	n, _, _ := creator()
	if _, exists := datasourceRegistry[n]; exists {
		return fmt.Errorf("datasource with name '%s' is already registered", n)
	}
	datasourceRegistry[n] = creator
	return nil
}
func CreateSource(name string) (SourceStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := sourceRegistry[name]
	if !ok {
		return SourceStore{}, fmt.Errorf("factory error: no source registered with name: %s", name)
	}
	n, handle, ds, p := creator()
	return SourceStore{Name: n, Handle: handle, Datasource: ds, Params: p}, nil
}

func CreateProcessor(name string) (ProcessorStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := processorRegistry[name]
	if !ok {
		return ProcessorStore{}, fmt.Errorf("factory error: no processor registered with name: %s", name)
	}
	n, handle, p := creator()
	return ProcessorStore{Name: n, Handle: handle, Params: p}, nil
}

func CreateSink(name string) (SinkStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := sinkRegistry[name]
	if !ok {
		return SinkStore{}, fmt.Errorf("factory error: no sink registered with name: %s", name)
	}
	n, handle, ds, p := creator()
	return SinkStore{Name: n, Handle: handle, Datasource: ds, Params: p}, nil
}

func CreateExecutor(name string) (ExecutorStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := executorRegistry[name]
	if !ok {
		return ExecutorStore{}, fmt.Errorf("factory error: no Executor registered with name: %s", name)
	}
	n, handle, ds, p := creator()
	return ExecutorStore{Name: n, Handle: handle, Datasource: ds, Params: p}, nil
}

func CreateVariable(name string) (VariableStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := variableRegistry[name]
	if !ok {
		return VariableStore{}, fmt.Errorf("factory error: no Variable registered with name: %s", name)
	}
	n, handle, ds, p := creator()
	return VariableStore{Name: n, Handle: handle, Datasource: ds, Params: p}, nil
}

func CreateDataSource(name string) (DatasourceStore, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	creator, ok := datasourceRegistry[name]
	if !ok {
		return DatasourceStore{}, fmt.Errorf("factory error: no DataSource registered with name: %s", name)
	}
	n, handle, p := creator()
	return DatasourceStore{Name: n, Handle: handle, Params: p}, nil
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
