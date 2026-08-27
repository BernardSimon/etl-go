package datasource

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

const (
	defaultMaxOpenConns    = 5
	defaultMaxIdleConns    = 2
	defaultConnMaxLifetime = 5 * time.Minute
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

func ConfigureSQLDBPool(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("sql db is nil")
	}
	db.SetMaxOpenConns(defaultMaxOpenConns)
	db.SetMaxIdleConns(defaultMaxIdleConns)
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	return nil
}

func ConfigureSQLDBPoolFromConfig(db *sql.DB, config map[string]string) error {
	if db == nil {
		return fmt.Errorf("sql db is nil")
	}
	maxOpenConns, err := parsePositiveInt(config, "__pool_max_open_conns", defaultMaxOpenConns)
	if err != nil {
		return err
	}
	maxIdleConns, err := parsePositiveInt(config, "__pool_max_idle_conns", defaultMaxIdleConns)
	if err != nil {
		return err
	}
	connMaxLifetimeSeconds, err := parsePositiveInt(config, "__pool_conn_max_lifetime", int(defaultConnMaxLifetime/time.Second))
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(time.Duration(connMaxLifetimeSeconds) * time.Second)
	return nil
}

func parsePositiveInt(config map[string]string, key string, fallback int) (int, error) {
	if config == nil {
		return fallback, nil
	}
	raw := config[key]
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid sql pool config %s: %w", key, err)
	}
	if value <= 0 {
		return fallback, nil
	}
	return value, nil
}
