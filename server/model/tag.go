package model

import "gorm.io/gorm"

// Tag 标签
type Tag struct {
	Model
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Color     string         `json:"color" gorm:"size:20"` // 标签颜色，如 #1890ff
}

// TaskTag 任务-标签关联表
type TaskTag struct {
	TaskID string `gorm:"size:36;primarykey" json:"task_id"`
	TagID  string `gorm:"size:36;primarykey" json:"tag_id"`
	Tag    Tag    `gorm:"foreignKey:TagID" json:"tag"`
}
