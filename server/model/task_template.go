package model

import types "github.com/BernardSimon/etl-go/server/types"

type TaskTemplate struct {
	Model
	Name     string          `json:"name" gorm:"size:255;index"`
	TaskType string          `json:"tasktypes" gorm:"size:32"`
	Cron     string          `json:"cron" gorm:"size:40"`
	Data     *types.TaskData `json:"data" gorm:"type:json"`
}
