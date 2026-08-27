package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
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
	postgreDatasourceName = customDatasourceName
}

type Sink struct {
	db            *sql.DB           // 数据库连接池
	table         string            // 目标表名
	columnMapping map[string]string // 列映射关系 (Record 中的键 -> 数据库中的列名)
	datasource    datasource.Datasource
	dialect       sqlDialect
	mode          writeMode
	conflictCols  []string
	updateCols    []string
}

type sqlDialect string

const (
	dialectMySQL     sqlDialect = "mysql"
	dialectPostgre   sqlDialect = "postgre"
	dialectSQLite    sqlDialect = "sqlite"
	defaultWriteMode writeMode  = writeModeInsert
)

type writeMode string

const (
	writeModeInsert writeMode = "insert"
	writeModeUpsert writeMode = "upsert"
)

func SinkCreatorMysql() (string, sink.Sink, *string, []params.Params) {
	return mysqlName, &Sink{dialect: dialectMySQL, mode: defaultWriteMode}, &mysqlDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
		{
			Key:          "mode",
			Required:     false,
			DefaultValue: string(defaultWriteMode),
			Description:  "Write mode: insert or upsert",
		},
		{
			Key:          "conflict_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns used to detect conflicts in upsert mode; defaults to table primary key columns when omitted",
		},
		{
			Key:          "update_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns to update in upsert mode; defaults to all mapped non-conflict columns",
		},
	}
}
func SinkCreatorPostgre() (string, sink.Sink, *string, []params.Params) {
	return postgreName, &Sink{dialect: dialectPostgre, mode: defaultWriteMode}, &postgreDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
		{
			Key:          "mode",
			Required:     false,
			DefaultValue: string(defaultWriteMode),
			Description:  "Write mode: insert or upsert",
		},
		{
			Key:          "conflict_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns used to detect conflicts in upsert mode; defaults to table primary key columns when omitted",
		},
		{
			Key:          "update_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns to update in upsert mode; defaults to all mapped non-conflict columns",
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
	return sqliteName, &Sink{dialect: dialectSQLite, mode: defaultWriteMode}, &sqliteDatasourceName, []params.Params{
		{
			Key:          "table",
			Required:     true,
			DefaultValue: "",
			Description:  "sql table name",
		},
		{
			Key:          "mode",
			Required:     false,
			DefaultValue: string(defaultWriteMode),
			Description:  "Write mode: insert or upsert",
		},
		{
			Key:          "conflict_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns used to detect conflicts in upsert mode; defaults to table primary key columns when omitted",
		},
		{
			Key:          "update_columns",
			Required:     false,
			DefaultValue: "",
			Description:  "Comma-separated columns to update in upsert mode; defaults to all mapped non-conflict columns",
		},
	}
}

// Open 负责解析配置并初始化数据库连接设置
func (s *Sink) Open(ctx context.Context, config map[string]string, columnMapping map[string]string, dataSource datasource.Datasource) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// 处理 column_mapping
	if len(columnMapping) == 0 {
		return fmt.Errorf("sql sink: 'column_mapping' cannot be empty")
	}
	s.columnMapping = columnMapping

	// 从 datasource 获取数据库连接
	if dataSource != nil {
		dbInstance, err := datasource.AsSQLDB(dataSource)
		if err != nil {
			return fmt.Errorf("sql sink: failed to get database connection from datasource: %w", err)
		}
		s.db = dbInstance
		s.datasource = dataSource
	}

	// 从 params 获取表名
	if t, ok := config["table"]; ok && t != "" {
		s.table = t
	} else {
		return fmt.Errorf("sql sink: config is missing required key 'table'")
	}

	s.mode = defaultWriteMode
	if mode, ok := config["mode"]; ok && mode != "" {
		switch writeMode(strings.ToLower(strings.TrimSpace(mode))) {
		case writeModeInsert:
			s.mode = writeModeInsert
		case writeModeUpsert:
			s.mode = writeModeUpsert
		default:
			return fmt.Errorf("sql sink: unsupported mode '%s', expected 'insert' or 'upsert'", mode)
		}
	}

	var err error
	s.conflictCols, err = s.resolveConfiguredColumns(config["conflict_columns"])
	if err != nil {
		return fmt.Errorf("sql sink: invalid conflict_columns: %w", err)
	}
	s.updateCols, err = s.resolveConfiguredColumns(config["update_columns"])
	if err != nil {
		return fmt.Errorf("sql sink: invalid update_columns: %w", err)
	}

	// 验证数据库连接是否存在
	if s.db == nil {
		return fmt.Errorf("sql sink: database connection is not available")
	}

	if s.mode == writeModeUpsert {
		if len(s.conflictCols) == 0 {
			s.conflictCols, err = s.detectPrimaryKeyColumns(ctx)
			if err != nil {
				return fmt.Errorf("sql sink: failed to detect primary key columns for upsert: %w", err)
			}
		}
		if len(s.conflictCols) == 0 {
			return fmt.Errorf("sql sink: upsert mode requires conflict_columns or a primary key on table '%s'", s.table)
		}
	}

	return nil
}

// Write 将一批记录通过构建一个大的 INSERT 语句在事务中批量写入数据库
func (s *Sink) Write(ctx context.Context, _ string, records []record.Record) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	if s.db == nil {
		return fmt.Errorf("sql sink: database connection is not open")
	}

	// 启动事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sql sink: failed to begin transaction: %w", err)
	}
	// 如果出现任何错误，确保事务被回滚
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// 根据 column_mapping 准备数据库列名和占位符顺序
	dbColumns := make([]string, 0, len(s.columnMapping))
	recordKeysInOrder := make([]string, 0, len(s.columnMapping))

	recordKeys := make([]string, 0, len(s.columnMapping))
	for recordKey := range s.columnMapping {
		recordKeys = append(recordKeys, recordKey)
	}
	sort.Slice(recordKeys, func(i, j int) bool {
		leftCol := s.columnMapping[recordKeys[i]]
		rightCol := s.columnMapping[recordKeys[j]]
		if leftCol == rightCol {
			return recordKeys[i] < recordKeys[j]
		}
		return leftCol < rightCol
	})

	for _, recordKey := range recordKeys {
		dbCol := s.columnMapping[recordKey]
		dbColumns = append(dbColumns, s.quoteIdentifier(dbCol))
		recordKeysInOrder = append(recordKeysInOrder, recordKey)
	}

	query := s.buildInsertQuery(len(records), dbColumns, recordKeysInOrder)

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
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("sql sink: failed to execute batch insert: %w", err)
	}

	// 如果一切顺利，提交事务
	return tx.Commit()
}

// Close 负责关闭数据库连接池，释放所有底层连接
func (s *Sink) Close() error {
	if s.datasource == nil {
		return nil
	}
	return s.datasource.Close()
}

func (s *Sink) buildInsertQuery(recordCount int, dbColumns []string, recordKeysInOrder []string) string {
	valueGroups := make([]string, 0, recordCount)
	argIndex := 1
	for i := 0; i < recordCount; i++ {
		placeholders := make([]string, 0, len(recordKeysInOrder))
		for range recordKeysInOrder {
			placeholders = append(placeholders, s.placeholder(argIndex))
			argIndex++
		}
		valueGroups = append(valueGroups, "("+strings.Join(placeholders, ", ")+")")
	}

	baseQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		s.quoteTableName(s.table), strings.Join(dbColumns, ", "), strings.Join(valueGroups, ", "))

	if s.mode != writeModeUpsert {
		return baseQuery
	}

	updateCols := s.updateCols
	if len(updateCols) == 0 {
		updateCols = make([]string, 0, len(dbColumns))
		conflictSet := make(map[string]struct{}, len(s.conflictCols))
		for _, col := range s.conflictCols {
			conflictSet[col] = struct{}{}
		}
		for _, col := range s.orderedMappedDBColumns(recordKeysInOrder) {
			if _, ok := conflictSet[col]; !ok {
				updateCols = append(updateCols, col)
			}
		}
	}

	switch s.dialect {
	case dialectMySQL:
		return baseQuery + " " + s.mysqlUpsertClause(updateCols)
	case dialectPostgre, dialectSQLite:
		return baseQuery + " " + s.standardUpsertClause(updateCols)
	default:
		return baseQuery
	}
}

func (s *Sink) mysqlUpsertClause(updateCols []string) string {
	if len(updateCols) == 0 {
		col := s.quoteIdentifier(s.conflictCols[0])
		return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s=%s", col, col)
	}
	assignments := make([]string, 0, len(updateCols))
	for _, col := range updateCols {
		quoted := s.quoteIdentifier(col)
		assignments = append(assignments, fmt.Sprintf("%s=VALUES(%s)", quoted, quoted))
	}
	return "ON DUPLICATE KEY UPDATE " + strings.Join(assignments, ", ")
}

func (s *Sink) standardUpsertClause(updateCols []string) string {
	conflictCols := make([]string, 0, len(s.conflictCols))
	for _, col := range s.conflictCols {
		conflictCols = append(conflictCols, s.quoteIdentifier(col))
	}
	if len(updateCols) == 0 {
		return "ON CONFLICT (" + strings.Join(conflictCols, ", ") + ") DO NOTHING"
	}
	assignments := make([]string, 0, len(updateCols))
	for _, col := range updateCols {
		quoted := s.quoteIdentifier(col)
		assignments = append(assignments, fmt.Sprintf("%s=excluded.%s", quoted, quoted))
	}
	return "ON CONFLICT (" + strings.Join(conflictCols, ", ") + ") DO UPDATE SET " + strings.Join(assignments, ", ")
}

func (s *Sink) orderedMappedDBColumns(recordKeysInOrder []string) []string {
	cols := make([]string, 0, len(recordKeysInOrder))
	for _, key := range recordKeysInOrder {
		cols = append(cols, s.columnMapping[key])
	}
	return cols
}

func (s *Sink) resolveConfiguredColumns(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	items := strings.Split(raw, ",")
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		if mapped, ok := s.columnMapping[name]; ok {
			name = mapped
		}
		if !s.isMappedDBColumn(name) {
			return nil, fmt.Errorf("column '%s' is not present in column_mapping", item)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return result, nil
}

func (s *Sink) isMappedDBColumn(name string) bool {
	for _, col := range s.columnMapping {
		if col == name {
			return true
		}
	}
	return false
}

func (s *Sink) detectPrimaryKeyColumns(ctx context.Context) ([]string, error) {
	schemaName, tableName := splitSchemaAndTable(s.table)

	switch s.dialect {
	case dialectMySQL:
		query := `
SELECT k.COLUMN_NAME
FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS t
JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE k
  ON t.CONSTRAINT_NAME = k.CONSTRAINT_NAME
 AND t.TABLE_SCHEMA = k.TABLE_SCHEMA
 AND t.TABLE_NAME = k.TABLE_NAME
WHERE t.CONSTRAINT_TYPE = 'PRIMARY KEY'
  AND t.TABLE_NAME = ?
  AND t.TABLE_SCHEMA = COALESCE(NULLIF(?, ''), DATABASE())
ORDER BY k.ORDINAL_POSITION`
		return s.queryPrimaryKeyColumns(ctx, query, tableName, schemaName)
	case dialectPostgre:
		query := `
SELECT kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
WHERE tc.constraint_type = 'PRIMARY KEY'
  AND tc.table_name = $1
  AND tc.table_schema = COALESCE(NULLIF($2, ''), current_schema())
ORDER BY kcu.ordinal_position`
		return s.queryPrimaryKeyColumns(ctx, query, tableName, schemaName)
	case dialectSQLite:
		return s.detectSQLitePrimaryKeyColumns(ctx, schemaName, tableName)
	default:
		return nil, fmt.Errorf("unsupported sql dialect '%s'", s.dialect)
	}
}

func (s *Sink) queryPrimaryKeyColumns(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if s.isMappedDBColumn(name) {
			result = append(result, name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Sink) detectSQLitePrimaryKeyColumns(ctx context.Context, schemaName, tableName string) ([]string, error) {
	pragmaTarget := s.quoteSQLitePragmaTarget(tableName)
	query := "PRAGMA table_info(" + pragmaTarget + ")"
	if schemaName != "" {
		query = "PRAGMA " + s.quoteIdentifier(schemaName) + ".table_info(" + pragmaTarget + ")"
	}

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type sqlitePKCol struct {
		name string
		pk   int
	}
	var cols []sqlitePKCol
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		if pk > 0 && s.isMappedDBColumn(name) {
			cols = append(cols, sqlitePKCol{name: name, pk: pk})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(cols, func(i, j int) bool {
		return cols[i].pk < cols[j].pk
	})
	result := make([]string, 0, len(cols))
	for _, col := range cols {
		result = append(result, col.name)
	}
	return result, nil
}

func (s *Sink) quoteIdentifier(name string) string {
	parts := strings.Split(name, ".")
	quoted := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		switch s.dialect {
		case dialectPostgre, dialectSQLite:
			quoted = append(quoted, `"`+strings.ReplaceAll(part, `"`, `""`)+`"`)
		default:
			quoted = append(quoted, "`"+strings.ReplaceAll(part, "`", "``")+"`")
		}
	}
	return strings.Join(quoted, ".")
}

func (s *Sink) quoteTableName(name string) string {
	return s.quoteIdentifier(name)
}

func (s *Sink) placeholder(index int) string {
	if s.dialect == dialectPostgre {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

func (s *Sink) quoteSQLitePragmaTarget(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}

func splitSchemaAndTable(name string) (schema string, table string) {
	parts := strings.Split(name, ".")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", strings.TrimSpace(name)
}
