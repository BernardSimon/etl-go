package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	_ "github.com/go-sql-driver/mysql"
)

type DataSource struct {
	db *sql.DB
}

var name = "mysql"

func SetCustomName(customName string) {
	name = customName
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "host",
			Required:     true,
			DefaultValue: "",
			Description:  "sql host",
		},
		{
			Key:          "port",
			Required:     true,
			DefaultValue: "3306",
			Description:  "sql port",
		},
		{
			Key:          "user",
			Required:     true,
			DefaultValue: "",
			Description:  "sql user",
		},
		{
			Key:          "password",
			Required:     true,
			DefaultValue: "",
			Description:  "sql password",
		},
		{
			Key:          "database",
			Required:     true,
			DefaultValue: "",
			Description:  "sql database",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	var err error
	d.db, err = sql.Open("mysql", config["user"]+":"+config["password"]+"@tcp("+config["host"]+":"+config["port"]+")/"+config["database"])
	if err != nil {
		return err
	}
	if err = d.db.Ping(); err != nil {
		return fmt.Errorf("sql executor: failed to connect to database: %w", err)
	}
	return nil
}

func (d *DataSource) DB() *sql.DB {
	return d.db
}

func (d *DataSource) ConfigMap() map[string]string {
	return nil
}

func (d *DataSource) Close() error {
	return d.db.Close()
}

// ListTables 实现 SchemaProvider 接口，通过 INFORMATION_SCHEMA 获取所有表和列
func (d *DataSource) ListTables() ([]datasource.TableInfo, error) {
	rows, err := d.db.Query(`
		SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		ORDER BY TABLE_NAME, ORDINAL_POSITION`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()
	return scanTableColumns(rows)
}

func scanTableColumns(rows *sql.Rows) ([]datasource.TableInfo, error) {
	tableMap := make(map[string]*datasource.TableInfo)
	var tableOrder []string

	for rows.Next() {
		var tableName, colName, colType, isNullable string
		if err := rows.Scan(&tableName, &colName, &colType, &isNullable); err != nil {
			return nil, err
		}
		if _, exists := tableMap[tableName]; !exists {
			tableMap[tableName] = &datasource.TableInfo{Name: tableName}
			tableOrder = append(tableOrder, tableName)
		}
		tableMap[tableName].Columns = append(tableMap[tableName].Columns, datasource.ColumnInfo{
			Name:     colName,
			Type:     colType,
			Nullable: strings.EqualFold(isNullable, "yes"),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]datasource.TableInfo, 0, len(tableOrder))
	for _, n := range tableOrder {
		result = append(result, *tableMap[n])
	}
	return result, nil
}
