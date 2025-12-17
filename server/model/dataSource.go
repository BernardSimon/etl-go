package model

import _type "github.com/BernardSimon/etl-go/server/type"

type DataSource struct {
	Model
	Name string          `json:"name" gorm:"size:255"`
	Type string          `json:"type" gorm:"size:255"`
	Data _type.KeyValues `json:"data" gorm:"type:text;serializer:encryption"`
}
