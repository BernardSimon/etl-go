package pipeline

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/processor"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/stretchr/testify/assert"
)

// --- Mock implementations (internal to pipeline tests) ---

type mockSource struct {
	records   []record.Record
	columns   map[string]string
	openErr   error
	readErr   error
	closeErr  error
	readIndex int
}

func (m *mockSource) Open(_ context.Context, _ map[string]string, _ datasource.Datasource) error {
	return m.openErr
}
func (m *mockSource) Read(_ context.Context) (record.Record, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if m.readIndex >= len(m.records) {
		return nil, io.EOF
	}
	r := m.records[m.readIndex]
	m.readIndex++
	return r, nil
}
func (m *mockSource) Close() error              { return m.closeErr }
func (m *mockSource) Column() map[string]string { return m.columns }

type mockProcessor struct {
	processFunc func(record.Record) (record.Record, error)
	openErr     error
	closeErr    error
}

func (m *mockProcessor) Open(_ context.Context, _ map[string]string) error { return m.openErr }
func (m *mockProcessor) Process(_ context.Context, r record.Record) (record.Record, error) {
	if m.processFunc != nil {
		return m.processFunc(r)
	}
	return r, nil
}
func (m *mockProcessor) Close() error                       { return m.closeErr }
func (m *mockProcessor) HandleColumns(_ *map[string]string) {}

type mockSink struct {
	writtenBatches [][]record.Record
	openErr        error
	writeErr       error
	closeErr       error
}

func (m *mockSink) Open(_ context.Context, _ map[string]string, _ map[string]string, _ datasource.Datasource) error {
	return m.openErr
}
func (m *mockSink) Write(_ context.Context, _ string, records []record.Record) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	batch := make([]record.Record, len(records))
	copy(batch, records)
	m.writtenBatches = append(m.writtenBatches, batch)
	return nil
}
func (m *mockSink) Close() error { return m.closeErr }

// --- Test cases ---

func TestEngine_Run_SimplePassThrough(t *testing.T) {
	records := []record.Record{
		{"id": "1", "name": "alice"},
		{"id": "2", "name": "bob"},
	}
	columns := map[string]string{"id": "id", "name": "name"}

	src := &mockSource{records: records, columns: columns}
	snk := &mockSink{}

	engine := newTestEngine("test-1", src, nil, snk)

	ctx := context.Background()
	sourceConfig := map[string]string{}
	sinkConfig := map[string]string{}

	err := engine.Run("test-1", ctx, nil, sourceConfig, nil, sinkConfig, nil)
	assert.NoError(t, err)

	// Verify all records were written
	totalWritten := 0
	for _, batch := range snk.writtenBatches {
		totalWritten += len(batch)
	}
	assert.Equal(t, 2, totalWritten)
}

func TestEngine_Run_WithProcessor(t *testing.T) {
	records := []record.Record{
		{"id": "1", "name": "alice"},
		{"id": "2", "name": "bob"},
		{"id": "3", "name": "charlie"},
	}
	columns := map[string]string{"id": "id", "name": "name"}

	src := &mockSource{records: records, columns: columns}
	snk := &mockSink{}

	// Processor that filters out records with id "2"
	proc := &mockProcessor{
		processFunc: func(r record.Record) (record.Record, error) {
			if r["id"] == "2" {
				return nil, nil // filter out
			}
			return r, nil
		},
	}

	engine := newTestEngine("test-2", src, []processor.Processor{proc}, snk)

	ctx := context.Background()
	err := engine.Run("test-2", ctx, nil, map[string]string{}, []ProcessorConfig{{Type: "mock", Params: map[string]string{}}}, map[string]string{}, nil)
	assert.NoError(t, err)

	totalWritten := 0
	for _, batch := range snk.writtenBatches {
		totalWritten += len(batch)
	}
	assert.Equal(t, 2, totalWritten)
}

func TestEngine_Run_SourceError(t *testing.T) {
	src := &mockSource{readErr: errors.New("read failure")}
	snk := &mockSink{}

	engine := newTestEngine("test-err", src, nil, snk)

	ctx := context.Background()
	err := engine.Run("test-err", ctx, nil, map[string]string{}, nil, map[string]string{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source error")
}

func TestEngine_Run_SinkError(t *testing.T) {
	records := []record.Record{
		{"id": "1"},
	}
	src := &mockSource{records: records, columns: map[string]string{"id": "id"}}
	snk := &mockSink{writeErr: errors.New("write failure")}

	engine := newTestEngine("test-sink-err", src, nil, snk)

	ctx := context.Background()
	err := engine.Run("test-sink-err", ctx, nil, map[string]string{}, nil, map[string]string{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sink error")
}

func TestEngine_Run_ContextCancellation(t *testing.T) {
	// Source that generates many records
	manyRecords := make([]record.Record, 10000)
	for i := range manyRecords {
		manyRecords[i] = record.Record{"id": i}
	}
	src := &mockSource{records: manyRecords, columns: map[string]string{"id": "id"}}
	snk := &mockSink{}

	engine := newTestEngine("test-cancel", src, nil, snk)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := engine.Run("test-cancel", ctx, nil, map[string]string{}, nil, map[string]string{}, nil)
	// May or may not error depending on timing, but should not hang
	_ = err
}

// newTestEngine creates a minimal Engine for testing with small batch/channel sizes.
func newTestEngine(id string, src source.Source, processors []processor.Processor, snk sink.Sink) *Engine {
	if processors == nil {
		processors = []processor.Processor{}
	}
	return &Engine{
		id:          id,
		source:      src,
		processors:  processors,
		sink:        snk,
		batchSize:   10,
		channelSize: 100,
	}
}
