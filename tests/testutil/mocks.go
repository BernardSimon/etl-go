package testutil

import (
	"io"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/record"
)

// MockSource is a test double for the source.Source interface.
type MockSource struct {
	Records    []record.Record
	ColumnMap  map[string]string
	OpenErr    error
	ReadErr    error
	CloseErr   error
	readIndex  int
	OpenCalled bool
}

func (m *MockSource) Open(_ map[string]string, _ *datasource.Datasource) error {
	m.OpenCalled = true
	return m.OpenErr
}

func (m *MockSource) Read() (record.Record, error) {
	if m.ReadErr != nil {
		return nil, m.ReadErr
	}
	if m.readIndex >= len(m.Records) {
		return nil, io.EOF
	}
	r := m.Records[m.readIndex]
	m.readIndex++
	return r, nil
}

func (m *MockSource) Close() error {
	return m.CloseErr
}

func (m *MockSource) Column() map[string]string {
	if m.ColumnMap != nil {
		return m.ColumnMap
	}
	return map[string]string{}
}

// MockProcessor is a test double for the procrssor.Processor interface.
type MockProcessor struct {
	ProcessFunc func(record.Record) (record.Record, error)
	OpenErr     error
	CloseErr    error
}

func (m *MockProcessor) Open(_ map[string]string) error {
	return m.OpenErr
}

func (m *MockProcessor) Process(r record.Record) (record.Record, error) {
	if m.ProcessFunc != nil {
		return m.ProcessFunc(r)
	}
	return r, nil
}

func (m *MockProcessor) Close() error {
	return m.CloseErr
}

func (m *MockProcessor) HandleColumns(_ *map[string]string) {}

// MockSink is a test double for the sink.Sink interface.
type MockSink struct {
	WrittenBatches [][]record.Record
	OpenErr        error
	WriteErr       error
	CloseErr       error
}

func (m *MockSink) Open(_ map[string]string, _ map[string]string, _ *datasource.Datasource) error {
	return m.OpenErr
}

func (m *MockSink) Write(_ string, records []record.Record) error {
	if m.WriteErr != nil {
		return m.WriteErr
	}
	batch := make([]record.Record, len(records))
	copy(batch, records)
	m.WrittenBatches = append(m.WrittenBatches, batch)
	return nil
}

func (m *MockSink) Close() error {
	return m.CloseErr
}

// MockExecutor is a test double for the executor.Executor interface.
type MockExecutor struct {
	OpenErr  error
	CloseErr error
}

func (m *MockExecutor) Open(_ map[string]string, _ *datasource.Datasource) error {
	return m.OpenErr
}

func (m *MockExecutor) Close() error {
	return m.CloseErr
}

// MockDatasource is a test double for the datasource.Datasource interface.
type MockDatasource struct {
	InitErr  error
	OpenVal  any
	CloseErr error
}

func (m *MockDatasource) Init(_ map[string]string) error {
	return m.InitErr
}

func (m *MockDatasource) Open() any {
	return m.OpenVal
}

func (m *MockDatasource) Close() error {
	return m.CloseErr
}
