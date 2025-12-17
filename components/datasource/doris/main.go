package doris

import (
	"strings"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/params"
)

var name = "doris"

func SetCustomName(customName string) {
	name = customName
}

type DataSource struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
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
			DefaultValue: "",
			Description:  "doris port",
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
		},
		{
			Key:          "database",
			Required:     true,
			DefaultValue: "",
			Description:  "doris database",
		},
	}
}

func (d *DataSource) Open() any {
	return map[string]string{
		"host":     d.Host,
		"port":     d.Port,
		"user":     d.User,
		"password": d.Password,
		"database": d.Database,
	}
}

func (d *DataSource) Init(config map[string]string) error {
	d.Host = config["host"]
	d.Port = config["port"]
	d.User = config["user"]
	d.Password = config["password"]
	d.Database = config["database"]
	if !strings.HasPrefix(d.Host, "http://") {
		d.Host = "http://" + d.Host
	}
	return nil
}
func (d *DataSource) Close() error {
	return nil
}
