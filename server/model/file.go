package model

type File struct {
	Model
	Name   string `json:"name"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	ExName string `json:"ex_name"`
}

type TaskRecordFile struct {
	Model
	TaskRecordID string `json:"task_record_id"`
	FileID       string `json:"file_id"`
	File         File
}
