package sql

import (
	"database/sql"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/params"
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
	datasource *datasource.Datasource
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

func (s *Executor) Open(config map[string]string, datasource *datasource.Datasource) error {
	query, ok := config["sql"]
	if !ok || query == "" {
		return fmt.Errorf("sql executor: config is missing or has invalid 'sql'")
	}
	var err error
	s.datasource = datasource
	s.db = (*s.datasource).Open().(*sql.DB)
	s.results, err = s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("sql executor: failed to executor sql: %w", err)
	}
	return nil
}
func (s *Executor) Close() error {
	// 然后关闭 db 连接池。
	err := (*s.datasource).Close()
	if err != nil {
		return err
	}
	return nil
}
