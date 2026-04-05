package model

import types "github.com/BernardSimon/etl-go/server/types"

type Variable struct {
	Model
	Name         string           `json:"name" gorm:"size:255"`
	Type         string           `json:"type" gorm:"size:255"`
	Description  string           `json:"description" gorm:"size:255"`
	DataSourceID *string          `json:"datasource_id" gorm:"size:36"`
	DataSource   *DataSource      `json:"datasource"`
	Value        *types.KeyValues `json:"value" gorm:"type:json"`
}
