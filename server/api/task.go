package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/task"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"

	"github.com/robfig/cron/v3"
)

func AddTask(req *types.AddTaskRequest, _ string) (interface{}, error) {
	if req.Cron != "manual" {
		if _, err := cron.ParseStandard(req.Cron); err != nil {
			return nil, errors.New("invalid cron expression")
		}
	}
	Mission := model.Task{
		Name:   req.Name,
		Cron:   req.Cron,
		Status: 0,
		Data:   &req.ParStr,
	}
	if err := model.DB.Create(&Mission).Error; err != nil {
		return nil, errors.New("failed to create task")
	}
	return "success", nil
}
func DeleteTask(req *types.DeleteTaskRequest, lang string) (interface{}, error) {
	var m model.Task
	result := model.DB.Where("id = ?", req.Id).First(&m)
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

func GetTaskAll(req *types.GetTaskAllRequest, _ string) (interface{}, error) {
	var missionList []model.Task
	var total int64
	pageNo := req.PageNo
	pageSize := req.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	tx := model.DB.Model(&model.Task{})
	if req.Search != "" {
		tx = tx.Where("name LIKE ?", "%"+req.Search+"%")
	}
	if req.MissionName != "" {
		tx = tx.Where("name LIKE ?", "%"+req.MissionName+"%")
	}
	if req.Status != nil {
		tx = tx.Where("status = ?", *req.Status)
	}
	switch req.TaskType {
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

func GetTaskById(req *types.GetTaskByIdRequest, _ string) (interface{}, error) {
	var m model.Task
	model.DB.First(&m, req.Id)
	return m, nil
}
func UpdateTask(req *types.UpdateTaskRequest, lang string) (interface{}, error) {
	if req.Cron != "manual" {
		if _, err := cron.ParseStandard(req.Cron); err != nil {
			return nil, errors.New("invalid cron expression")
		}
	}
	var m model.Task
	model.DB.Where("id = ?", req.Id).First(&m)
	if m.ID == "" {
		return nil, errors.New("task not found")
	}
	if m.Status == 1 {
		return nil, errors.New("cannot edit in task scheduling")
	}
	m.Name = req.Name
	m.Cron = req.Cron
	m.Data = &req.ParStr
	m.Status = 0
	if err := model.DB.Save(&m).Error; err != nil {
		return nil, errors.New("failed to edit task")
	}
	return i18n.Translate(lang, "success"), nil
}

func RunTask(req *types.RunTaskRequest, lang string) (interface{}, error) {
	var m model.Task

	// 开启一个数据库事务
	tx := model.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查询并锁定任务记录
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", req.Id).First(&m).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("task not found")
	}
	// 查询Corn
	if m.Cron == "manual" {
		return nil, errors.New("manual task cannot be scheduled")
	}

	// 检查任务状态
	if m.Status == 1 {
		tx.Rollback()
		return nil, errors.New("task already scheduling")
	}

	// 更新任务状态
	m.Status = 1
	if err := tx.Save(&m).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update task status")
	}

	// 调度任务
	err := task.ScheduleMission(&m)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("system error")
	}

	return i18n.Translate(lang, "success"), nil
}

func StopTask(req *types.StopTaskRequest, lang string) (interface{}, error) {
	var m model.Task
	model.DB.Where("id = ?", req.Id).Find(&m)
	defer model.DB.Save(&m)
	if m.Status != 1 {
		return nil, errors.New("unable to stop scheduling task has not started yet")
	}
	task.CancelMission(&m)
	return i18n.Translate(lang, "success"), nil
}
func RunTaskOnce(req *types.RunTaskOnceRequest, _ string) (interface{}, error) {
	var m model.Task
	model.DB.Where("id = ?", req.Id).Find(&m)
	err := task.RunMissionManual(m.ID)
	if err != nil {
		return nil, err
	}
	return "task has started running, please check the results", nil
}

// Task 信息获取

func GetTypeByComponent(_ *interface{}, _ string) (interface{}, error) {
	var response types.GetTypeByComponentResponse

	// Query all datasources once to avoid N+1 queries
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

func GetTaskRecordList(req *types.GetTaskRecordListRequest, _ string) (interface{}, error) {
	var missionRecordList []model.TaskRecord
	var total int64
	tx := model.DB.Model(&model.TaskRecord{}).Preload("Task")
	pageNo := req.PageNo
	pageSize := req.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if req.ID != "" {
		tx = tx.Where("id = ?", req.ID)
	}
	if req.TaskID != "" {
		tx = tx.Where("task_id = ?", req.TaskID)
	}
	if req.MissionName != "" {
		tx = tx.Joins("Task").Where("Task.name LIKE ?", "%"+req.MissionName+"%")
	}
	if req.Status != -1 {
		tx = tx.Where("status = ?", req.Status)
	}
	tx.Count(&total).Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("created_at desc").Find(&missionRecordList)
	return map[string]interface{}{
		"total":     total,
		"list":      missionRecordList,
		"page_no":   pageNo,
		"page_size": pageSize,
	}, nil
}

func GetTaskRecordsByTaskID(req *types.GetTaskRecordListRequest, lang string) (interface{}, error) {
	return GetTaskRecordList(req, lang)
}

func GetTaskLatestLog(req *types.GetTaskRecordListRequest, _ string) (interface{}, error) {
	var taskRecord model.TaskRecord
	if err := model.DB.Where("task_id = ?", req.TaskID).Order("created_at desc").First(&taskRecord).Error; err != nil {
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

func CancelTaskRecord(req *types.CancelTaskRecord, lang string) (interface{}, error) {
	var missionRecord model.TaskRecord
	err := model.DB.Where("id = ?", req.ID).First(&missionRecord).Error
	if err != nil {
		return nil, errors.New("task record not found")
	}
	if missionRecord.Status != 0 {
		return nil, errors.New("task record already finish")
	}
	err = task.CancelMissionRecord(req.ID)
	if err != nil {
		return nil, err
	}
	return i18n.Translate(lang, "the task is being forcibly terminated. Please refresh later to check the status"), nil
}

func GetFileListByTaskRecordID(req *types.CancelTaskRecord, _ string) (interface{}, error) {
	var fileList []model.TaskRecordFile
	if req.ID == "" {
		return nil, errors.New("task record id is required")
	}
	err := model.DB.Model(&model.TaskRecordFile{}).Where("task_record_id = ?", req.ID).Preload("File").Find(&fileList).Error
	if err != nil {
		return nil, errors.New("failed to get file list")
	}
	var files = make([]model.File, 0)
	for _, file := range fileList {
		files = append(files, file.File)
	}
	return files, nil
}

func GetTaskRecordParams(req *types.CancelTaskRecord, _ string) (interface{}, error) {
	var missionRecord model.TaskRecord
	if err := model.DB.Preload("Task").Where("id = ?", req.ID).First(&missionRecord).Error; err != nil {
		return nil, errors.New("task record not found")
	}
	return map[string]interface{}{
		"id":           missionRecord.ID,
		"task_id":      missionRecord.TaskID,
		"mission_name": missionRecord.Task.Name,
		"params":       missionRecord.Data,
	}, nil
}

func GetTaskRecordLogs(req *types.CancelTaskRecord, _ string) (interface{}, error) {
	var missionRecord model.TaskRecord
	if err := model.DB.Preload("Task").Where("id = ?", req.ID).First(&missionRecord).Error; err != nil {
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
