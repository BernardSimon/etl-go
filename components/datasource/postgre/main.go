package postgre

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	_ "github.com/lib/pq"
)

type DataSource struct {
	db *sql.DB
}

var name = "postgre"

func SetCustomName(customName string) {
	name = customName
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "host",
			Required:     true,
			DefaultValue: "",
			Description:  "postgresql host",
		},
		{
			Key:          "port",
			Required:     true,
			DefaultValue: "5432",
			Description:  "postgresql port",
		},
		{
			Key:          "user",
			Required:     true,
			DefaultValue: "",
			Description:  "postgresql user",
		},
		{
			Key:          "password",
			Required:     true,
			DefaultValue: "",
			Description:  "postgresql password",
			Mask:         true,
		},
		{
			Key:          "database",
			Required:     true,
			DefaultValue: "",
			Description:  "postgresql database",
		},
		{
			Key:          "sslmode",
			Required:     false,
			DefaultValue: "disable",
			Description:  "postgresql ssl mode",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	var err error
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		config["host"],
		config["port"],
		config["user"],
		config["password"],
		config["database"],
	)
	if v, ok := config["sslmode"]; ok {
		connStr += " sslmode=" + v
	}
	d.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err = d.db.Ping(); err != nil {
		return fmt.Errorf("postgresql executor: failed to connect to database: %w", err)
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

// ListTables 实现 SchemaProvider 接口，通过 information_schema 获取所有表和列
func (d *DataSource) ListTables() ([]datasource.TableInfo, error) {
	rows, err := d.db.Query(`
		SELECT table_name, column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

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
