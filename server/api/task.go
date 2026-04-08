package api

import (
	"errors"
	"time"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/task"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"

	"github.com/robfig/cron/v3"
)

func AddTask(_ *struct{}, body *types.AddTaskRequest, _ string) (interface{}, error) {
	if body.Cron != "manual" {
		if _, err := cron.ParseStandard(body.Cron); err != nil {
			return nil, errors.New("invalid cron expression")
		}
	}
	Mission := model.Task{
		Name:   body.Name,
		Cron:   body.Cron,
		Status: 0,
		Data:   &body.ParStr,
	}
	if err := model.DB.Create(&Mission).Error; err != nil {
		return nil, errors.New("failed to create task")
	}
	return "success", nil
}

func DeleteTask(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var m model.Task
	result := model.DB.Where("id = ?", uri.Id).First(&m)
	if result.Error != nil {
		return false, errors.New("task not found")
	}
	if m.Status == 1 {
		return false, errors.New("cannot delete in task scheduling")
	}
	err := model.DB.Model(&model.Task{}).Where("id = ?", m.ID).Delete(&m).Error
	if err != nil {
		return false, errors.New("failed to delete task")
	}
	return i18n.Translate(lang, "success"), nil
}

func GetTaskAll(_ *struct{}, query *types.GetTaskAllRequest, _ string) (interface{}, error) {
	var missionList []model.Task
	var total int64
	pageNo := query.PageNo
	pageSize := query.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	tx := model.DB.Model(&model.Task{})
	if query.Search != "" {
		tx = tx.Where("name LIKE ?", "%"+query.Search+"%")
	}
	if query.MissionName != "" {
		tx = tx.Where("name LIKE ?", "%"+query.MissionName+"%")
	}
	if query.Status != nil {
		tx = tx.Where("status = ?", *query.Status)
	}
	switch query.TaskType {
	case "manual":
		tx = tx.Where("cron = ?", "manual")
	case "scheduled":
		tx = tx.Where("cron <> ?", "manual")
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, errors.New("failed to get task list")
	}
	if err := tx.Order("created_at desc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&missionList).Error; err != nil {
		return nil, errors.New("failed to get task list")
	}

	return map[string]interface{}{
		"list":      missionList,
		"total":     total,
		"page_no":   pageNo,
		"page_size": pageSize,
	}, nil
}

func GetTaskById(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var m model.Task
	model.DB.First(&m, uri.Id)
	return m, nil
}

func UpdateTask(uri *types.IDUri, body *types.UpdateTaskBody, lang string) (interface{}, error) {
	if body.Cron != "manual" {
		if _, err := cron.ParseStandard(body.Cron); err != nil {
			return nil, errors.New("invalid cron expression")
		}
	}
	var m model.Task
	model.DB.Where("id = ?", uri.Id).First(&m)
	if m.ID == "" {
		return nil, errors.New("task not found")
	}
	if m.Status == 1 {
		return nil, errors.New("cannot edit in task scheduling")
	}
	m.Name = body.Name
	m.Cron = body.Cron
	m.Data = &body.ParStr
	m.Status = 0
	if err := model.DB.Save(&m).Error; err != nil {
		return nil, errors.New("failed to edit task")
	}
	return i18n.Translate(lang, "success"), nil
}

func RunTask(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var m model.Task

	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", uri.Id).First(&m).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("task not found")
	}
	if m.Cron == "manual" {
		return nil, errors.New("manual task cannot be scheduled")
	}
	if m.Status == 1 {
		tx.Rollback()
		return nil, errors.New("task already scheduling")
	}

	if err := tx.Model(&m).Updates(map[string]interface{}{
		"status":     1,
		"is_running": false,
		"entry_id":   nil,
	}).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("system error")
	}

	m.Status = 1
	m.IsRunning = false
	m.EntryID = nil

	// 先在持锁事务里写入 status=1，避免 cron 在整分钟边界触发时读到未调度状态。
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("system error")
	}

	if err := task.ScheduleMission(&m); err != nil {
		return nil, err
	}

	return i18n.Translate(lang, "success"), nil
}

func StopTask(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var m model.Task
	model.DB.Where("id = ?", uri.Id).Find(&m)
	if m.Status != 1 {
		return nil, errors.New("unable to stop scheduling task has not started yet")
	}
	if err := task.CancelMission(&m); err != nil {
		return nil, errors.New("failed to stop task")
	}
	return i18n.Translate(lang, "success"), nil
}

func RunTaskOnce(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var m model.Task
	model.DB.Where("id = ?", uri.Id).Find(&m)
	err := task.RunMissionManual(m.ID)
	if err != nil {
		return nil, err
	}
	return "task has started running, please check the results", nil
}

func PreviewTask(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	result, err := task.PreviewTask(uri.Id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ── 组件元数据 ────────────────────────────────────────────────────────────────

func GetTypeByComponent(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	var response types.GetTypeByComponentResponse

	var allDataSources []model.DataSource
	model.DB.Model(&model.DataSource{}).Select("name", "id", "type").Find(&allDataSources)
	dsByType := make(map[string][]struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	})
	for _, ds := range allDataSources {
		dsByType[ds.Type] = append(dsByType[ds.Type], struct {
			Name string `json:"name"`
			ID   string `json:"id"`
		}{Name: ds.Name, ID: ds.ID})
	}

	var execute, source, sink []types.TypeDataSource
	var processor []types.TypeNoDataSource

	for _, typeName := range factory.GetExecutorTypeList() {
		store, _ := factory.CreateExecutor(typeName)
		exItem := types.TypeDataSource{Type: typeName, Params: normalizeParams(store.Params)}
		if store.Datasource != nil {
			dsL := dsByType[*store.Datasource]
			exItem.DataSource = &dsL
		}
		execute = append(execute, exItem)
	}
	for _, typeName := range factory.GetSourceTypeList() {
		store, _ := factory.CreateSource(typeName)
		sourceItem := types.TypeDataSource{Type: typeName, Params: normalizeParams(store.Params)}
		if store.Datasource != nil {
			dsL := dsByType[*store.Datasource]
			sourceItem.DataSource = &dsL
		}
		source = append(source, sourceItem)
	}
	for _, typeName := range factory.GetProcessorTypeList() {
		store, _ := factory.CreateProcessor(typeName)
		processor = append(processor, types.TypeNoDataSource{
			Type:   typeName,
			Params: normalizeParams(store.Params),
		})
	}
	for _, typeName := range factory.GetSinkTypeList() {
		store, _ := factory.CreateSink(typeName)
		sinkItem := types.TypeDataSource{Type: typeName, Params: normalizeParams(store.Params)}
		if store.Datasource != nil {
			dsL := dsByType[*store.Datasource]
			sinkItem.DataSource = &dsL
		}
		sink = append(sink, sinkItem)
	}

	response.Executor = execute
	response.Source = source
	response.Processor = processor
	response.Sink = sink
	return response, nil
}

// ── 任务执行记录 ──────────────────────────────────────────────────────────────

func GetTaskRecordList(_ *struct{}, query *types.GetTaskRecordListRequest, _ string) (interface{}, error) {
	var missionRecordList []model.TaskRecord
	var total int64
	tx := model.DB.Model(&model.TaskRecord{}).Preload("Task")
	pageNo := query.PageNo
	pageSize := query.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if query.ID != "" {
		tx = tx.Where("id = ?", query.ID)
	}
	if query.TaskID != "" {
		tx = tx.Where("task_id = ?", query.TaskID)
	}
	if query.MissionName != "" {
		tx = tx.Joins("Task").Where("Task.name LIKE ?", "%"+query.MissionName+"%")
	}
	if query.Status != -1 {
		tx = tx.Where("status = ?", query.Status)
	}
	tx.Count(&total).Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("created_at desc").Find(&missionRecordList)
	return map[string]interface{}{
		"total":     total,
		"list":      missionRecordList,
		"page_no":   pageNo,
		"page_size": pageSize,
	}, nil
}

func GetTaskRecordsByTaskID(uri *types.IDUri, query *types.GetTaskRecordListRequest, lang string) (interface{}, error) {
	query.TaskID = uri.Id
	return GetTaskRecordList(nil, query, lang)
}

func GetTaskLatestLog(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var taskRecord model.TaskRecord
	if err := model.DB.Where("task_id = ?", uri.Id).Order("created_at desc").First(&taskRecord).Error; err != nil {
		return nil, errors.New("task record not found")
	}
	return map[string]interface{}{
		"task_id":    taskRecord.TaskID,
		"record_id":  taskRecord.ID,
		"status":     taskRecord.Status,
		"start_time": taskRecord.StartTime,
		"end_time":   taskRecord.EndTime,
		"message":    taskRecord.Message,
	}, nil
}

func CleanTaskRecords(_ *struct{}, body *types.CleanTaskRecordsRequest, _ string) (interface{}, error) {
	tx := model.DB.Model(&model.TaskRecord{}).Where("status != 0") // 运行中的记录不清理
	if body.Status != nil {
		tx = tx.Where("status = ?", *body.Status)
	}
	if body.Before != "" {
		var t time.Time
		var err error
		t, err = time.Parse(time.RFC3339, body.Before)
		if err != nil {
			t, err = time.Parse("2006-01-02", body.Before)
			if err != nil {
				return nil, errors.New("invalid before date format, use RFC3339 or YYYY-MM-DD")
			}
		}
		tx = tx.Where("created_at < ?", t)
	}
	result := tx.Delete(&model.TaskRecord{})
	if result.Error != nil {
		return nil, errors.New("failed to clean records")
	}
	return map[string]interface{}{"deleted": result.RowsAffected}, nil
}

func CancelTaskRecord(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var missionRecord model.TaskRecord
	err := model.DB.Where("id = ?", uri.Id).First(&missionRecord).Error
	if err != nil {
		return nil, errors.New("task record not found")
	}
	if missionRecord.Status != 0 {
		return nil, errors.New("task record already finish")
	}
	err = task.CancelMissionRecord(uri.Id)
	if err != nil {
		return nil, err
	}
	return i18n.Translate(lang, "the task is being forcibly terminated. Please refresh later to check the status"), nil
}

func GetFileListByTaskRecordID(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var fileList []model.TaskRecordFile
	if uri.Id == "" {
		return nil, errors.New("task record id is required")
	}
	err := model.DB.Model(&model.TaskRecordFile{}).Where("task_record_id = ?", uri.Id).Preload("File").Find(&fileList).Error
	if err != nil {
		return nil, errors.New("failed to get file list")
	}
	var files = make([]model.File, 0)
	for _, file := range fileList {
		files = append(files, file.File)
	}
	return files, nil
}

func GetTaskRecordParams(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var missionRecord model.TaskRecord
	if err := model.DB.Preload("Task").Where("id = ?", uri.Id).First(&missionRecord).Error; err != nil {
		return nil, errors.New("task record not found")
	}
	return map[string]interface{}{
		"id":           missionRecord.ID,
		"task_id":      missionRecord.TaskID,
		"mission_name": missionRecord.Task.Name,
		"params":       missionRecord.Data,
	}, nil
}

func GetTaskRecordLogs(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var missionRecord model.TaskRecord
	if err := model.DB.Preload("Task").Where("id = ?", uri.Id).First(&missionRecord).Error; err != nil {
		return nil, errors.New("task record not found")
	}
	return map[string]interface{}{
		"id":           missionRecord.ID,
		"task_id":      missionRecord.TaskID,
		"mission_name": missionRecord.Task.Name,
		"status":       missionRecord.Status,
		"start_time":   missionRecord.StartTime,
		"end_time":     missionRecord.EndTime,
		"message":      missionRecord.Message,
		"log":          missionRecord.Message,
	}, nil
}
