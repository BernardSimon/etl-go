package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
	//_ "modernc.org/sqlite"  //项目本身已经引入sqlite，无需重复引用
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
	d.db, err = sql.Open("sqlite3", config["file_path"])
	if err != nil {
		return err
	}
	if err = d.db.Ping(); err != nil {
		return fmt.Errorf("sqlite executor: failed to connect to database: %w", err)
	}
	return nil
}

func (d *DataSource) Open() any {
	return d.db
}

func (d *DataSource) Close() error {
	return d.db.Close()
}
