package sql

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/variable"
)

var mysqlName = "mysql"
var mysqlDatasourceName = "mysql"

func SetCustomNameMysql(customName string, datasourceName string) {
	mysqlName = customName
	mysqlDatasourceName = datasourceName
}

type Variable struct {
}

func VariableCreatorMysql() (string, variable.Variable, *string, []params.Params) {
	return mysqlName, &Variable{}, &mysqlDatasourceName, []params.Params{
		{
			Key:          "query",
			Required:     true,
			DefaultValue: "",
		},
	}
}

var postgreName = "postgre"
var postgreDatasourceName = "postgre"

func SetCustomNamePostgre(customName string, datasourceName string) {
	postgreName = customName
	postgreDatasourceName = datasourceName
}

func VariableCreatorPostgre() (string, variable.Variable, *string, []params.Params) {
	return postgreName, &Variable{}, &postgreDatasourceName, []params.Params{
		{
			Key:          "query",
			Required:     true,
			DefaultValue: "",
		},
	}
}

var sqliteName = "sqlite"
var sqliteDatasourceName = "sqlite"

func SetCustomNameSqlite(customName string, datasourceName string) {
	sqliteName = customName
	sqliteDatasourceName = datasourceName
}
func VariableCreatorSqlite() (string, variable.Variable, *string, []params.Params) {
	return sqliteName, &Variable{}, &sqliteDatasourceName, []params.Params{
		{
			Key:          "query",
			Required:     true,
			DefaultValue: "",
		},
	}
}

func (s *Variable) Get(ctx context.Context, config map[string]string, ds datasource.Datasource) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	query, exist := config["query"]
	if !exist {
		return "", errors.New("variable query is required")
	}
	err := validVariable(config)
	if err != nil {
		return "", err
	}
	if ds == nil {
		return "", errors.New("variable datasource is required")
	}
	db, err := datasource.AsSQLDB(ds)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = ds.Close()
	}()
	var result string
	err = db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func validVariable(config map[string]string) error {
	query := config["query"]
	trimmedSql := strings.TrimSpace(query)
	upperSql := strings.ToUpper(trimmedSql)

	// 必须以SELECT开头
	if !strings.HasPrefix(upperSql, "SELECT") {
		return errors.New("variable Should Has SELECT Prefix")
	}

	// 检查是否包含危险关键字
	dangerousKeywords := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "EXEC"}
	for _, keyword := range dangerousKeywords {
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		matched, _ := regexp.MatchString(pattern, upperSql)
		if matched {
			return errors.New("variable Should Not Contains Dangerous Keywords")
		}
	}
	return nil
}
