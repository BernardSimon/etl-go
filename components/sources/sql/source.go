package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"io"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/source"
)

// Source 实现了 core.Source 接口，用于从 MySQL 数据库读取数据。
type Source struct {
	db          *sql.DB   // 数据库连接池。它被设计为长期存活且线程安全。
	rows        *sql.Rows // SQL 查询结果集的前向只读迭代器。
	datasource  *datasource.Datasource
	columnNames []string // 预先获取的查询结果列名。
}

var mysqlName = "mysql"
var mysqlDatasourceName = "mysql"

func SetCustomNameMysql(customName string, customDatasourceName string) {
	mysqlName = customName
	mysqlDatasourceName = customDatasourceName
}
func SourceCreatorMysql() (string, source.Source, *string, []params.Params) {
	paramList := []params.Params{
		{
			Key:          "query",
			DefaultValue: "",
			Required:     true,
			Description:  "",
		},
	}

	return mysqlName, &Source{}, &mysqlDatasourceName, paramList
}

var postgresqlName = "postgre"
var postgresqlDatasourceName = "postgre"

func SetCustomNamePostgresql(customName string, customDatasourceName string) {
	postgresqlName = customName
	postgresqlDatasourceName = customDatasourceName
}
func SourceCreatorPostgre() (string, source.Source, *string, []params.Params) {
	paramList := []params.Params{
		{
			Key:          "query",
			DefaultValue: "",
			Required:     true,
			Description:  "",
		},
	}

	return postgresqlName, &Source{}, &postgresqlDatasourceName, paramList
}

var sqliteName = "sqlite"
var sqliteDatasourceName = "sqlite"

func SetCustomNameSqlite(customName string, customDatasourceName string) {
	sqliteName = customName
	sqliteDatasourceName = customDatasourceName
}
func SourceCreatorSqlite() (string, source.Source, *string, []params.Params) {
	paramList := []params.Params{
		{
			Key:          "query",
			DefaultValue: "",
			Required:     true,
			Description:  "",
		},
	}

	return sqliteName, &Source{}, &sqliteDatasourceName, paramList
}

func (s *Source) Open(config map[string]string, dataSource *datasource.Datasource) error {
	s.datasource = dataSource
	// 'query' 是必需配置。
	query, ok := config["query"]
	if !ok || query == "" {
		return fmt.Errorf("sql source: config is missing or has invalid 'query'")
	}

	var err error
	s.db = (*dataSource).Open().(*sql.DB)
	// 执行查询，获取结果集迭代器。
	s.rows, err = s.db.Query(query)
	if err != nil {
		return fmt.Errorf("sql source: failed to executor query: %w", err)
	}
	// 预先获取并存储列名，这将在 Read 方法中用于构建 map[string]interface{} 格式的 Record。
	s.columnNames, err = s.rows.Columns()
	if err != nil {
		return fmt.Errorf("sql source: failed to get column names from result set: %w", err)
	}

	return nil
}

// Read 读取查询结果的下一行，并将其转换为一个 `core.Record`。
func (s *Source) Read() (record.Record, error) {
	// 检查结果集中是否还有下一行。
	if !s.rows.Next() {
		// 在迭代结束后，必须调用 .Err() 来检查循环期间是否发生错误。
		if err := s.rows.Err(); err != nil {
			return nil, fmt.Errorf("sql source: error during row iteration: %w", err)
		}
		// 如果没有错误，说明已成功到达结果集末尾，返回 EOF 信号。
		return nil, io.EOF
	}

	// 核心技巧：为了实现最大的通用性和健壮性，我们不直接扫描到具体的 Go 类型（如 int, string, time.Time），
	// 而是使用 `sql.RawBytes`。这有几个好处：
	// 1. 避免因数据库类型（如可为 NULL 的整数）与 Go 类型不匹配而导致的扫描错误。
	// 2. 优雅地处理 NULL 值（此时 RawBytes 切片为 nil）。
	// 3. 将所有值作为原始字节切片来处理，将类型转换的责任推迟到下游的 Processor，
	//    这使得 Source 组件更通用，更符合 ETL 的分阶段处理思想。
	values := make([]sql.RawBytes, len(s.columnNames))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// 将当前行的数据扫描到 `values` 中。
	if err := s.rows.Scan(scanArgs...); err != nil {
		return nil, fmt.Errorf("sql source: failed to scan row: %w", err)
	}

	r := make(record.Record)
	for i, colName := range s.columnNames {
		// 如果数据库中的值是 NULL，对应的 RawBytes 切片将是 nil。
		if values[i] == nil {
			r[colName] = nil
		} else {
			// 否则，将其作为字符串存入 record。下游的 processor 可以根据需要再进行解析。
			r[colName] = string(values[i])
		}
	}

	return r, nil
}

// Close 负责优雅地关闭数据库资源。
// 它会先尝试关闭结果集迭代器，再关闭数据库连接池。
func (s *Source) Close() error {
	var errs []error

	// 必须先关闭 rows。
	if s.rows != nil {
		if err := s.rows.Close(); err != nil {
			errs = append(errs, fmt.Errorf("sql source: failed to close rows: %w", err))
		}
	}
	err := (*s.datasource).Close()
	if err != nil {
		errs = append(errs, fmt.Errorf("sql source: failed to close db: %w", err))
	}

	return errors.Join(errs...)
}

func (s *Source) Column() map[string]string {
	k := make(map[string]string)
	for _, v := range s.columnNames {
		k[v] = v
	}
	return k
}
