package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	_ "github.com/glebarez/go-sqlite"
)

type DataSource struct {
	db *sql.DB
}

var name = "sqlite"

func SetCustomName(customName string) {
	name = customName
}

func DatasourceCreator() (string, datasource.Datasource, []params.Params) {
	return name, &DataSource{}, []params.Params{
		{
			Key:          "file_id",
			Required:     true,
			DefaultValue: "",
			Description:  "sqlite database file id",
		},
	}
}

func (d *DataSource) Init(config map[string]string) error {
	var err error
	d.db, err = sql.Open("sqlite", config["file_path"])
	if err != nil {
		return err
	}
	if err = d.db.Ping(); err != nil {
		return fmt.Errorf("sqlite executor: failed to connect to database: %w", err)
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

// ListTables 实现 SchemaProvider 接口，通过 sqlite_master + PRAGMA 获取所有表和列
func (d *DataSource) ListTables() ([]datasource.TableInfo, error) {
	tableRows, err := d.db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer tableRows.Close()

	var tableNames []string
	for tableRows.Next() {
		var name string
		if err := tableRows.Scan(&name); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}

	var tables []datasource.TableInfo
	for _, tableName := range tableNames {
		colRows, err := d.db.Query(fmt.Sprintf("PRAGMA table_info(%q)", tableName))
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
		}
		var columns []datasource.ColumnInfo
		for colRows.Next() {
			var cid int
			var colName, colType string
			var notNull int
			var dfltValue sql.NullString
			var pk int
			if err := colRows.Scan(&cid, &colName, &colType, &notNull, &dfltValue, &pk); err != nil {
				colRows.Close()
				return nil, err
			}
			columns = append(columns, datasource.ColumnInfo{
				Name:     colName,
				Type:     colType,
				Nullable: notNull == 0,
			})
		}
		colRows.Close()
		tables = append(tables, datasource.TableInfo{Name: tableName, Columns: columns})
	}
	return tables, nil
}
