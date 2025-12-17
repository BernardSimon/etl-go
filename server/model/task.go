package model

import (
	_type "github.com/BernardSimon/etl-go/server/type"

	"gorm.io/gorm"
)

type Task struct {
	Model
	DeletedAt       gorm.DeletedAt  `gorm:"index"`
	Name            string          `json:"mission_name" gorm:"size:255"`
	Cron            string          `json:"cron" gorm:"size:40"`
	Data            *_type.TaskData `json:"data" gorm:"type:json"`
	Status          int             `json:"status" gorm:"default:0;size:2"` // 0:暂存 1:调度中 2:错误
	LastRunTime     *CustomTime     `json:"last_run_time"`
	LastSuccessTime *CustomTime     `json:"last_success_time"`
	LastEndTime     *CustomTime     `json:"last_end_time"`
	ErrMsg          string          `json:"err_msg"`
	IsRunning       bool
	EntryID         *int // cron.EntryID
}

type TaskRecord struct {
	Model
	CreatedAt *CustomTime
	RunBy     string          `json:"run_by"`
	TaskID    string          `json:"task_id"`
	Task      Task            `json:"task"`
	Status    int             `json:"status"` //0运行中；1运行成功；2运行失败
	StartTime *CustomTime     `json:"start_time"`
	EndTime   *CustomTime     `json:"end_time"`
	Message   string          `json:"message"`
	Data      *_type.TaskData `json:"data" gorm:"type:json"`
}
