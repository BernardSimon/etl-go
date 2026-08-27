package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/params"
)

type TaskData struct {
	BeforeExecute *struct {
		Type       string  `json:"type"`
		DataSource *string `json:"data_source"`
		Params     []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"before_execute"`
	Source struct {
		Type       string  `json:"type"`
		DataSource *string `json:"data_source"`
		Params     []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"source"`
	Processors []struct {
		Type   string `json:"type"`
		Params []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"processors"`
	Sinks struct {
		Type       string  `json:"type"`
		DataSource *string `json:"data_source"`
		Params     []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"sink"`
	AfterExecute *struct {
		Type       string  `json:"type"`
		DataSource *string `json:"data_source"`
		Params     []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"params"`
	} `json:"after_execute"`
}

func (ct *TaskData) Value() (driver.Value, error) {
	if ct == nil {
		return nil, nil
	}
	return json.Marshal(ct)
}

func (ct *TaskData) Scan(value interface{}) error {
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

// ── 任务 ──────────────────────────────────────────────────────────────────────

type AddTaskRequest struct {
	Name   string   `json:"mission_name" binding:"required"`
	ParStr TaskData `json:"params" binding:"required"`
	Cron   string   `json:"cron" binding:"required"`
	TagIDs []string `json:"tag_ids"`
}

type GetTaskAllRequest struct {
	PageNo      int    `form:"page_no"`
	PageSize    int    `form:"page_size"`
	MissionName string `form:"mission_name"`
	Status      *int   `form:"status"`
	Search      string `form:"search"`
	TaskType    string `form:"tasktypes"`
	TagID       string `form:"tag_id"`       // 按标签ID筛选, "none" 表示无标签
}

// UpdateTaskBody 是 PUT /tasks/:id 的请求体；路径中的 id 通过 IDUri 传入
type UpdateTaskBody struct {
	Name   string   `json:"mission_name" binding:"required"`
	ParStr TaskData `json:"params" binding:"required"`
	Cron   string   `json:"cron" binding:"required"`
	TagIDs []string `json:"tag_ids"`
}

// ── 任务执行记录 ────────────────────────────────────────────────────────────────

type GetTaskRecordListRequest struct {
	PageNo      int    `form:"page_no"`
	PageSize    int    `form:"page_size"`
	MissionName string `form:"mission_name"`
	Status      int    `form:"status"`
	ID          string `form:"id"`
	TaskID      string `form:"task_id"`
}

type CleanTaskRecordsRequest struct {
	Status *int   `form:"status"` // 1=成功 2=失败 nil=全部已完成
	Before string `form:"before"` // RFC3339 或 2006-01-02，清理该时间点之前的记录
}

// ── 组件元数据 ─────────────────────────────────────────────────────────────────

type GetTypeByComponentResponse struct {
	Executor  []TypeDataSource   `json:"executor"`
	Source    []TypeDataSource   `json:"source"`
	Processor []TypeNoDataSource `json:"processor"`
	Sink      []TypeDataSource   `json:"sink"`
}
type TypeDataSource struct {
	Type       string `json:"type"`
	DataSource *[]struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"data_source"`
	Params []params.Params `json:"params"`
}
type TypeNoDataSource struct {
	Type   string          `json:"type"`
	Params []params.Params `json:"params"`
}

// ── 预览 ──────────────────────────────────────────────────────────────────────

// PreviewResponse 是 POST /tasks/:id/preview 的响应体
type PreviewResponse struct {
	// Columns 是输出列名（按字母序稳定排列）
	Columns []string                 `json:"columns"`
	// Rows 是最多 20 行的预览数据，每行为 map[列名]值
	Rows    []map[string]interface{} `json:"rows"`
}

// ── 任务模板 ──────────────────────────────────────────────────────────────────

// SaveTaskTemplateRequest：POST /task-templates，ID 为可选 body 字段（编辑时传入）
type SaveTaskTemplateRequest struct {
	ID       string   `json:"id,omitempty"`
	Name     string   `json:"name" binding:"required"`
	ParStr   TaskData `json:"params" binding:"required"`
	Cron     string   `json:"cron" binding:"required"`
	TaskType string   `json:"tasktypes" binding:"required"`
}
