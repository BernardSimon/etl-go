package datasource

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSQLDatasource struct {
	db *sql.DB
}

func (m *mockSQLDatasource) Init(map[string]string) error     { return nil }
func (m *mockSQLDatasource) Close() error                     { return nil }
func (m *mockSQLDatasource) DB() *sql.DB                      { return m.db }
func (m *mockSQLDatasource) ConfigMap() map[string]string     { return nil }
func (m *mockSQLDatasource) ListTables() ([]TableInfo, error) { return nil, nil }

type mockMapDatasource struct {
	cfg map[string]string
}

func (m *mockMapDatasource) Init(map[string]string) error     { return nil }
func (m *mockMapDatasource) Close() error                     { return nil }
func (m *mockMapDatasource) DB() *sql.DB                      { return nil }
func (m *mockMapDatasource) ConfigMap() map[string]string     { return m.cfg }
func (m *mockMapDatasource) ListTables() ([]TableInfo, error) { return nil, nil }

type mockInvalidDatasource struct{}

func (m *mockInvalidDatasource) Init(map[string]string) error     { return nil }
func (m *mockInvalidDatasource) Close() error                     { return nil }
func (m *mockInvalidDatasource) DB() *sql.DB                      { return nil }
func (m *mockInvalidDatasource) ConfigMap() map[string]string     { return nil }
func (m *mockInvalidDatasource) ListTables() ([]TableInfo, error) { return nil, nil }

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

func TestConfigureSQLDBPool_SetsDefaults(t *testing.T) {
	db := &sql.DB{}

	err := ConfigureSQLDBPool(db)

	require.NoError(t, err)
	stats := db.Stats()
	assert.Equal(t, defaultMaxOpenConns, stats.MaxOpenConnections)
	assert.GreaterOrEqual(t, db.Stats().Idle, 0)
}

func TestConfigureSQLDBPool_ReturnsErrorForNilDB(t *testing.T) {
	err := ConfigureSQLDBPool(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sql db is nil")
}

func TestDefaultConnMaxLifetime_IsPositive(t *testing.T) {
	assert.Equal(t, 5*time.Minute, defaultConnMaxLifetime)
}

func TestConfigureSQLDBPoolFromConfig_UsesOverrides(t *testing.T) {
	db := &sql.DB{}

	err := ConfigureSQLDBPoolFromConfig(db, map[string]string{
		"__pool_max_open_conns":     "7",
		"__pool_max_idle_conns":     "3",
		"__pool_conn_max_lifetime":  "120",
	})

	require.NoError(t, err)
	assert.Equal(t, 7, db.Stats().MaxOpenConnections)
}

func TestConfigureSQLDBPoolFromConfig_ReturnsErrorForInvalidOverride(t *testing.T) {
	db := &sql.DB{}

	err := ConfigureSQLDBPoolFromConfig(db, map[string]string{
		"__pool_max_open_conns": "abc",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "__pool_max_open_conns")
}
