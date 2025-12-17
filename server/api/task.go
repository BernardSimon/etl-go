package api

import (
	"errors"

	params2 "github.com/BernardSimon/etl-go/etl/core/params"
	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/task"
	_type "github.com/BernardSimon/etl-go/server/type"
	"github.com/BernardSimon/etl-go/server/utils/i18n"

	"github.com/robfig/cron/v3"
)

func AddTask(req *_type.AddTaskRequest, _ string) (interface{}, error) {
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
func DeleteTask(req *_type.DeleteTaskRequest, lang string) (interface{}, error) {
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

func GetTaskAll(_ *interface{}, _ string) (interface{}, error) {
	var missionList []model.Task
	model.DB.Model(&model.Task{}).Order("created_at desc").Find(&missionList)
	return missionList, nil
}

func GetTaskById(req *_type.GetTaskByIdRequest, _ string) (interface{}, error) {
	var m model.Task
	model.DB.First(&m, req.Id)
	return m, nil
}
func UpdateTask(req *_type.UpdateTaskRequest, lang string) (interface{}, error) {
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

func RunTask(req *_type.RunTaskRequest, lang string) (interface{}, error) {
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

func StopTask(req *_type.StopTaskRequest, lang string) (interface{}, error) {
	var m model.Task
	model.DB.Where("id = ?", req.Id).Find(&m)
	defer model.DB.Save(&m)
	if m.Status != 1 {
		return nil, errors.New("unable to stop scheduling task has not started yet")
	}
	task.CancelMission(&m)
	return i18n.Translate(lang, "success"), nil
}
func RunTaskOnce(req *_type.RunTaskOnceRequest, _ string) (interface{}, error) {
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
	var response _type.GetTypeByComponentResponse
	var execute, source, sink []_type.TypeDataSource
	var processor []_type.TypeNoDataSource
	for _, typeName := range factory.GetExecutorTypeList() {
		var dataSource []model.DataSource
		store, _ := factory.CreateExecutor(typeName)
		params := store.Params
		var exItem _type.TypeDataSource
		if store.Datasource != nil {
			name := *store.Datasource
			model.DB.Model(&model.DataSource{}).Select("name", "id").Where("type = ?", name).Find(&dataSource)
			dsL := make([]struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}, 0)
			for _, ds := range dataSource {
				dsL = append(dsL, struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{
					Name: ds.Name,
					ID:   ds.ID,
				})
			}
			exItem.DataSource = &dsL
		}
		exItem.Type = typeName
		exItem.Params = params
		execute = append(execute, exItem)
	}
	for _, typeName := range factory.GetSourceTypeList() {
		store, _ := factory.CreateSource(typeName)
		var dataSource []model.DataSource
		params := store.Params
		var sourceItem _type.TypeDataSource
		if store.Datasource != nil {
			name := *store.Datasource
			model.DB.Model(&model.DataSource{}).Select("name", "id").Where("type = ?", name).Find(&dataSource)
			dsL := make([]struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}, 0)
			for _, ds := range dataSource {
				dsL = append(dsL, struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{
					Name: ds.Name,
					ID:   ds.ID,
				})
			}
			sourceItem.DataSource = &dsL
		}
		sourceItem.Type = typeName
		sourceItem.Params = params
		source = append(source, sourceItem)
	}
	for _, typeName := range factory.GetProcessorTypeList() {
		store, _ := factory.CreateProcessor(typeName)
		params := store.Params
		processor = append(processor, struct {
			Type   string           `json:"type"`
			Params []params2.Params `json:"params"`
		}{
			Type:   typeName,
			Params: params,
		})
	}
	for _, typeName := range factory.GetSinkTypeList() {
		var dataSource []model.DataSource
		store, _ := factory.CreateSink(typeName)
		params := store.Params
		var sinkItem _type.TypeDataSource
		if store.Datasource != nil {
			name := *store.Datasource
			model.DB.Model(&model.DataSource{}).Select("name", "id").Where("type = ?", name).Find(&dataSource)
			dsL := make([]struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}, 0)
			for _, ds := range dataSource {
				dsL = append(dsL, struct {
					Name string `json:"name"`
					ID   string `json:"id"`
				}{
					Name: ds.Name,
					ID:   ds.ID,
				})
			}
			sinkItem.DataSource = &dsL
		}
		sinkItem.Type = typeName
		sinkItem.Params = params
		sink = append(sink, sinkItem)
	}
	response.Executor = execute
	response.Source = source
	response.Processor = processor
	response.Sink = sink
	return response, nil
}

func GetTaskRecordList(req *_type.GetTaskRecordListRequest, _ string) (interface{}, error) {
	var missionRecordList []model.TaskRecord
	var total int64
	tx := model.DB.Model(&model.TaskRecord{}).Preload("Task")
	if req.ID != "" {
		tx = tx.Where("id = ?", req.ID)
	}
	if req.MissionName != "" {
		tx = tx.Joins("left join missions on missions.id = mission_records.mission_id").Where("missions.name like ?", "%"+req.MissionName+"%")
	}
	if req.Status != -1 {
		tx = tx.Where("status = ?", req.Status)
	}
	tx.Count(&total).Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Order("created_at desc").Find(&missionRecordList)
	return map[string]interface{}{
		"total": total,
		"list":  missionRecordList,
	}, nil
}

func CancelTaskRecord(req *_type.CancelTaskRecord, lang string) (interface{}, error) {
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

func GetFileListByTaskRecordID(req *_type.CancelTaskRecord, _ string) (interface{}, error) {
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
