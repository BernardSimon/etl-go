package datasource

import (
	"database/sql"
	"fmt"
)

func AsSQLDB(ds Datasource) (*sql.DB, error) {
	if ds == nil {
		return nil, fmt.Errorf("datasource is nil")
	}
	if provider, ok := ds.(SQLDBProvider); ok {
		db := provider.DB()
		if db == nil {
			return nil, fmt.Errorf("datasource returned a nil *sql.DB")
		}
		return db, nil
	}
	return nil, fmt.Errorf("datasource does not implement SQLDBProvider")
}

func AsConfigMap(ds Datasource) (map[string]string, error) {
	if ds == nil {
		return nil, fmt.Errorf("datasource is nil")
	}
	if provider, ok := ds.(ConfigMapProvider); ok {
		cfg := provider.ConfigMap()
		if cfg == nil {
			return nil, fmt.Errorf("datasource returned a nil config map")
		}
		return cfg, nil
	}
	return nil, fmt.Errorf("datasource does not implement ConfigMapProvider")
}
