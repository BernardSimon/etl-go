package sql

import (
	"database/sql"
	"errors"
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
func (s *Variable) Get(config map[string]string, datasource *datasource.Datasource) (string, error) {
	query, exist := config["query"]
	if !exist {
		return "", errors.New("variable query is required")
	}
	err := validVariable(config)
	if err != nil {
		return "", err
	}
	db := (*datasource).Open().(*sql.DB)
	defer (*datasource).Close()
	var result string
	err = db.QueryRow(query).Scan(&result)
	if err != nil {
		err := (*datasource).Close()
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
		if strings.Contains(upperSql, keyword) {
			return errors.New("variable Should Not Contains Dangerous Keywords")
		}
	}
	return nil
}
