package sql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/core/record"
	"github.com/BernardSimon/etl-go/etl/core/sink"
)

var mysqlName = "mysql"
var mysqlDatasourceName = "mysql"

func SetCustomNameMysql(customName string, customDatasourceName string) {
	mysqlName = customName
	mysqlDatasourceName = customDatasourceName
}

var postgreName = "postgre"
var postgreDatasourceName = "postgre"

func SetCustomNamePostgre(customName string, customDatasourceName string) {
	postgreName = customName
	mysqlDatasourceName = customDatasourceName
}

type Sink struct {
	db            *sql.DB           // 数据库连接池
	table         string            // 目标表名
	columnMapping map[string]string // 列映射关系 (Record 中的键 -> 数据库中的列名)
	datasource    *datasource.Datasource
}

func SinkCreatorMysql() (string, sink.Sink, *string, []params.Params) {
	return mysqlName, &Sink{}, &mysqlDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
	}
}
func SinkCreatorPostgre() (string, sink.Sink, *string, []params.Params) {
	return postgreName, &Sink{}, &postgreDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
	}
}

var sqliteName = "sqlite"
var sqliteDatasourceName = "sqlite"

func SetCustomNameSqlite(customName string, customDatasourceName string) {
	sqliteName = customName
	sqliteDatasourceName = customDatasourceName
}
func SinkCreatorSqlite() (string, sink.Sink, *string, []params.Params) {
	return sqliteName, &Sink{}, &sqliteDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
	}
}

// Open 负责解析配置并初始化数据库连接设置
func (s *Sink) Open(config map[string]string, columnMapping map[string]string, dataSource *datasource.Datasource) error {
	// 处理 column_mapping
	if len(columnMapping) == 0 {
		return fmt.Errorf("sql sink: 'column_mapping' cannot be empty")
	}
	s.columnMapping = columnMapping

	// 从 datasource 获取数据库连接
	if dataSource != nil {
		db := (*dataSource).Open()
		if dbInstance, ok := db.(*sql.DB); ok {
			s.db = dbInstance
		} else {
			return fmt.Errorf("sql sink: failed to get database connection from datasource")
		}
		s.datasource = dataSource
	}

	// 从 params 获取表名
	if t, ok := config["table"]; ok && t != "" {
		s.table = t
	} else {
		return fmt.Errorf("sql sink: config is missing required key 'table'")
	}

	// 验证数据库连接是否存在
	if s.db == nil {
		return fmt.Errorf("sql sink: database connection is not available")
	}

	return nil
}

// Write 将一批记录通过构建一个大的 INSERT 语句在事务中批量写入数据库
func (s *Sink) Write(_ string, records []record.Record) error {
	if len(records) == 0 {
		return nil
	}

	if s.db == nil {
		return fmt.Errorf("sql sink: database connection is not open")
	}

	// 启动事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("sql sink: failed to begin transaction: %w", err)
	}
	// 如果出现任何错误，确保事务被回滚
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// 根据 column_mapping 准备数据库列名和占位符顺序
	dbColumns := make([]string, 0, len(s.columnMapping))
	placeholders := make([]string, 0, len(s.columnMapping))
	recordKeysInOrder := make([]string, 0, len(s.columnMapping))

	for recordKey, dbCol := range s.columnMapping {
		dbColumns = append(dbColumns, "`"+dbCol+"`") // 为列名加上反引号以处理保留字
		placeholders = append(placeholders, "?")
		recordKeysInOrder = append(recordKeysInOrder, recordKey)
	}

	// 构建 SQL 语句
	// 最终形式: "INSERT INTO [table](file:///Users/szy/Desktop/code/etl-go/components/sinks/sql/main.go#L13-L13) (`col1`, `col2`) VALUES (?, ?), (?, ?), ..."
	valuePlaceholderGroup := "(" + strings.Join(placeholders, ", ") + ")"
	allValuePlaceholders := strings.Repeat(valuePlaceholderGroup+",", len(records)-1) + valuePlaceholderGroup

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", s.table, strings.Join(dbColumns, ", "), allValuePlaceholders)

	// 准备所有参数
	args := make([]interface{}, 0, len(records)*len(s.columnMapping))
	for _, r := range records {
		for _, recordKey := range recordKeysInOrder {
			val, exists := r[recordKey]
			if !exists {
				val = nil // 如果记录中缺少该键，则插入 NULL
			}
			args = append(args, val)
		}
	}

	// 执行批量插入
	_, err = tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("sql sink: failed to execute batch insert: %w", err)
	}

	// 如果一切顺利，提交事务
	return tx.Commit()
}

// Close 负责关闭数据库连接池，释放所有底层连接
func (s *Sink) Close() error {
	return (*s.datasource).Close()
}
