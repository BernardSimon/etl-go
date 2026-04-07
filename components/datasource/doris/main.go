package doris

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	_ "github.com/go-sql-driver/mysql"
)

var name = "doris"

func SetCustomName(customName string) {
	name = customName
}

type DataSource struct {
	Host      string `json:"host"`
	Port      string `json:"port"`       // HTTP 协议端口（用于 Stream Load 等）
	MysqlPort string `json:"mysql_port"` // MySQL 协议端口（用于 SQL 查询 / schema 发现）
	User      string `json:"user"`
	Password  string `json:"password"`
	Database  string `json:"database"`
	db        *sql.DB
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "host",
			Required:     true,
			DefaultValue: "",
			Description:  "doris host",
		},
		{
			Key:          "port",
			Required:     true,
			DefaultValue: "8030",
			Description:  "doris HTTP protocol port (used for Stream Load)",
		},
		{
			Key:          "mysql_port",
			Required:     true,
			DefaultValue: "9030",
			Description:  "doris MySQL protocol port (used for SQL queries and schema discovery)",
		},
		{
			Key:          "user",
			Required:     true,
			DefaultValue: "",
			Description:  "doris user",
		},
		{
			Key:          "password",
			Required:     true,
			DefaultValue: "",
			Description:  "doris password",
			Mask:         true,
		},
		{
			Key:          "database",
			Required:     true,
			DefaultValue: "",
			Description:  "doris database",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	d.Host = config["host"]
	d.Port = config["port"]
	d.MysqlPort = config["mysql_port"]
	d.User = config["user"]
	d.Password = config["password"]
	d.Database = config["database"]
	if !strings.HasPrefix(d.Host, "http://") && !strings.HasPrefix(d.Host, "https://") {
		d.Host = "http://" + d.Host
	}

	// 通过 MySQL 协议建立连接，用于 SQL 查询和 schema 发现
	host := strings.TrimPrefix(strings.TrimPrefix(d.Host, "https://"), "http://")
	var err error
	d.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		d.User, d.Password, host, d.MysqlPort, d.Database))
	if err != nil {
		return fmt.Errorf("doris: failed to open mysql connection: %w", err)
	}
	if err = d.db.Ping(); err != nil {
		return fmt.Errorf("doris: failed to connect via mysql protocol: %w", err)
	}
	return nil
}

func (d *DataSource) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// ConfigMap 实现 ConfigMapProvider，供 Doris Sink（Stream Load）使用
func (d *DataSource) ConfigMap() map[string]string {
	return map[string]string{
		"host":       d.Host,
		"port":       d.Port,
		"mysql_port": d.MysqlPort,
		"user":       d.User,
		"password":   d.Password,
		"database":   d.Database,
	}
}

// DB 实现 SQLDBProvider，供需要 SQL 连接的组件使用
func (d *DataSource) DB() *sql.DB {
	return d.db
}

// ListTables 实现 SchemaProvider，通过 MySQL 协议查询 Doris 表结构
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
