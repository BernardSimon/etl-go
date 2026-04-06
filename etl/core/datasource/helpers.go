package datasource

import (
	"database/sql"
	"fmt"
)

func AsSQLDB(ds Datasource) (*sql.DB, error) {
	if ds == nil {
		return nil, fmt.Errorf("datasource is nil")
	}
	db := ds.DB()
	if db == nil {
		return nil, fmt.Errorf("datasource does not implement SQLDBProvider")
	}
	return db, nil
}

func AsConfigMap(ds Datasource) (map[string]string, error) {
	if ds == nil {
		return nil, fmt.Errorf("datasource is nil")
	}
	cfg := ds.ConfigMap()
	if cfg == nil {
		return nil, fmt.Errorf("datasource does not implement ConfigMapProvider")
	}
	return cfg, nil
}
