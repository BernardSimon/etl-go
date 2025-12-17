package postgre

import (
	"database/sql"
	"fmt"

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
	// PostgreSQL连接字符串格式
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

func (d *DataSource) Open() any {
	return d.db
}

func (d *DataSource) Close() error {
	return d.db.Close()
}
