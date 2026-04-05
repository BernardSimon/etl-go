package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/security"
)

var mysqlName = "mysql"
var mysqlDatasourceName = "mysql"

func SetCustomNameMysql(name string, datasourceName string) {
	mysqlName = name
	mysqlDatasourceName = datasourceName
}

var postgreName = "postgre"
var postgreDatasourceName = "postgre"

func SetCustomNamePostgresql(name string, datasourceName string) {
	postgreName = name
	postgreDatasourceName = datasourceName
}

var sqliteName = "sqlite"
var sqliteDatasourceName = "sqlite"

func SetCustomNameSqlite(name string, datasourceName string) {
	sqliteName = name
	sqliteDatasourceName = datasourceName
}

type Executor struct {
	db         *sql.DB
	results    sql.Result
	datasource datasource.Datasource
}

func ExecutorCreatorMysql() (string, executor.Executor, *string, []params.Params) {
	return mysqlName, &Executor{}, &mysqlDatasourceName, []params.Params{
		{
			Key:          "sql",
			Required:     true,
			DefaultValue: "",
			Description:  "sql query",
		},
	}
}
func ExecutorCreatorPostgre() (string, executor.Executor, *string, []params.Params) {
	return postgreName, &Executor{}, &postgreDatasourceName, []params.Params{
		{
			Key:          "sql",
			Required:     true,
			DefaultValue: "",
			Description:  "sql query",
		},
	}
}
func ExecutorCreatorSqlite() (string, executor.Executor, *string, []params.Params) {
	return sqliteName, &Executor{}, &sqliteDatasourceName, []params.Params{
		{
			Key:          "sql",
			Required:     true,
			DefaultValue: "",
			Description:  "sql query",
		},
	}
}

func (s *Executor) Open(ctx context.Context, config map[string]string, ds datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	query, ok := config["sql"]
	if !ok || query == "" {
		return fmt.Errorf("sql executor: config is missing or has invalid 'sql'")
	}
	// Validate SQL for dangerous statements
	allowDangerous := config["allow_dangerous"] == "true"
	if err := security.ValidateExecutorSQL(query, allowDangerous); err != nil {
		return fmt.Errorf("sql executor: %w", err)
	}
	s.datasource = ds
	if s.datasource == nil {
		return fmt.Errorf("sql executor: datasource is required")
	}
	dbInstance, dbErr := datasource.AsSQLDB(s.datasource)
	if dbErr != nil {
		return fmt.Errorf("sql executor: failed to get database connection from datasource: %w", dbErr)
	}
	s.db = dbInstance
	var err error
	s.results, err = s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("sql executor: failed to executor sql: %w", err)
	}
	return nil
}
func (s *Executor) Close() error {
	if s.datasource == nil {
		return nil
	}
	return s.datasource.Close()
}
