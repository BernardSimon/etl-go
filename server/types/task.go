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

type AddTaskRequest struct {
	Name   string   `json:"mission_name" binding:"required"`
	ParStr TaskData `json:"params" binding:"required"`
	Cron   string   `json:"cron" binding:"required"`
}

type DeleteTaskRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

type GetTaskByIdRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

type GetTaskAllRequest struct {
	PageNo      int    `json:"page_no" form:"page_no"`
	PageSize    int    `json:"page_size" form:"page_size"`
	MissionName string `json:"mission_name" form:"mission_name"`
	Status      *int   `json:"status" form:"status"`
	Search      string `json:"search" form:"search"`
	TaskType    string `json:"tasktypes" form:"tasktypes"`
}

type UpdateTaskRequest struct {
	Id     string   `json:"id" uri:"id" binding:"required"`
	Name   string   `json:"mission_name" binding:"required"`
	ParStr TaskData `json:"params" binding:"required"`
	Cron   string   `json:"cron" binding:"required"`
}

type RunTaskRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

type StopTaskRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}
type RunTaskOnceRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

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

type GetTaskRecordListRequest struct {
	PageNo      int    `json:"page_no" form:"page_no"`
	PageSize    int    `json:"page_size" form:"page_size"`
	MissionName string `json:"mission_name" form:"mission_name"`
	Status      int    `json:"status" form:"status"`
	ID          string `json:"id" form:"id"`
	TaskID      string `json:"task_id" form:"task_id" uri:"id"`
}

type CancelTaskRecord struct {
	ID string `json:"id" uri:"id"`
}

type SaveTaskTemplateRequest struct {
	ID       string   `json:"id,omitempty" uri:"id"`
	Name     string   `json:"name" binding:"required"`
	ParStr   TaskData `json:"params" binding:"required"`
	Cron     string   `json:"cron" binding:"required"`
	TaskType string   `json:"tasktypes" binding:"required"`
}

type DeleteTaskTemplateRequest struct {
	ID string `json:"id" uri:"id" binding:"required"`
}
