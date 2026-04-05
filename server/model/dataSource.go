package model

import types "github.com/BernardSimon/etl-go/server/types"

type DataSource struct {
	Model
	Name string          `json:"name" gorm:"size:255"`
	Type string          `json:"type" gorm:"size:255"`
	Data types.KeyValues `json:"data" gorm:"type:text;serializer:encryption"`
}
