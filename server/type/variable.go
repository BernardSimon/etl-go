package _type

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	params2 "github.com/BernardSimon/etl-go/etl/core/params"
)

type NewVariableRequest struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	DataSourceID *string   `json:"datasource_id" binding:"required"`
	Name         string    `json:"name" binding:"required"`
	Description  string    `json:"description" binding:"required"`
	Value        KeyValues `json:"value" binding:"required"`
	Edit         string    `json:"edit" binding:"required"`
}

type DeleteVariableRequest struct {
	Id string `json:"id" binding:"required"`
}

type TestVariableRequest struct {
	Id string `json:"id" binding:"required"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KeyValues []KeyValue

func (ct *KeyValues) Value() (driver.Value, error) {
	if ct == nil {
		return nil, nil
	}
	return json.Marshal(ct)
}

func (ct *KeyValues) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		err := json.Unmarshal(v, ct)
		return err
	case string:
		err := json.Unmarshal([]byte(v), ct)
		return err
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

type GetVariableTypeListResponse struct {
	Type           string           `json:"type"`
	Params         []params2.Params `json:"params"`
	DatasourceList *[]struct {
		Name string
		ID   string
	} `json:"datasource_list"`
}
