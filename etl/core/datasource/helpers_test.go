package datasource

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSQLDatasource struct {
	db *sql.DB
}

func (m *mockSQLDatasource) Init(map[string]string) error { return nil }
func (m *mockSQLDatasource) Close() error                 { return nil }
func (m *mockSQLDatasource) DB() *sql.DB                  { return m.db }

type mockMapDatasource struct {
	cfg map[string]string
}

func (m *mockMapDatasource) Init(map[string]string) error { return nil }
func (m *mockMapDatasource) Close() error                 { return nil }
func (m *mockMapDatasource) ConfigMap() map[string]string { return m.cfg }

type mockInvalidDatasource struct{}

func (m *mockInvalidDatasource) Init(map[string]string) error { return nil }
func (m *mockInvalidDatasource) Close() error                 { return nil }

func TestAsSQLDB_UsesTypedProvider(t *testing.T) {
	ds := &mockSQLDatasource{db: &sql.DB{}}

	db, err := AsSQLDB(ds)

	require.NoError(t, err)
	assert.NotNil(t, db)
}

func TestAsSQLDB_ReturnsErrorForInvalidDatasource(t *testing.T) {
	_, err := AsSQLDB(&mockInvalidDatasource{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SQLDBProvider")
}

func TestAsConfigMap_UsesTypedProvider(t *testing.T) {
	ds := &mockMapDatasource{cfg: map[string]string{"host": "localhost"}}

	cfg, err := AsConfigMap(ds)

	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg["host"])
}

func TestAsConfigMap_ReturnsErrorForInvalidDatasource(t *testing.T) {
	_, err := AsConfigMap(&mockInvalidDatasource{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ConfigMapProvider")
}
