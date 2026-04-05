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

func (d *DataSource) Close() error {
	return d.db.Close()
}
