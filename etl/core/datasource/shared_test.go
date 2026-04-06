package datasource

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type trackingDatasource struct {
	closeCalls int
	db         *sql.DB
	cfg        map[string]string
}

func (d *trackingDatasource) Init(map[string]string) error { return nil }
func (d *trackingDatasource) Close() error {
	d.closeCalls++
	return nil
}
func (d *trackingDatasource) DB() *sql.DB {
	return d.db
}
func (d *trackingDatasource) ConfigMap() map[string]string {
	return d.cfg
}
func (d *trackingDatasource) ListTables() ([]TableInfo, error) {
	return nil, nil
}

func TestSharedLease_ClosesBaseOnlyAfterLastRelease(t *testing.T) {
	base := &trackingDatasource{}
	shared, err := NewShared(base)
	require.NoError(t, err)

	lease1 := shared.Lease()
	lease2 := shared.Lease()

	require.NoError(t, lease1.Close())
	assert.Equal(t, 0, base.closeCalls)

	require.NoError(t, lease2.Close())
	assert.Equal(t, 1, base.closeCalls)

	require.NoError(t, lease2.Close())
	assert.Equal(t, 1, base.closeCalls)
}

func TestSharedLease_DelegatesTypedProviders(t *testing.T) {
	db := &sql.DB{}
	base := &trackingDatasource{
		db:  db,
		cfg: map[string]string{"host": "localhost"},
	}
	shared, err := NewShared(base)
	require.NoError(t, err)

	lease := shared.Lease()
	gotDB, err := AsSQLDB(lease)
	require.NoError(t, err)
	assert.Same(t, db, gotDB)

	gotCfg, err := AsConfigMap(lease)
	require.NoError(t, err)
	assert.Equal(t, "localhost", gotCfg["host"])

	require.NoError(t, lease.Close())
}
