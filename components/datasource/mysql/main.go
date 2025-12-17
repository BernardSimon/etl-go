package mysql

import (
	"database/sql"
	"fmt"

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

func (d *DataSource) Open() any {
	return d.db
}

func (d *DataSource) Close() error {
	return d.db.Close()
}
