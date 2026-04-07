package task

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/processor"
	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	"github.com/BernardSimon/etl-go/server/config"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var cr *cron.Cron

func SetMissions() {
	cr = cron.New()
	var missions []model.Task
	err := model.DB.Model(&model.Task{}).Where("is_running != 0").UpdateColumn("is_running", 0).Error
	if err != nil {
		zap.L().Error("任务启动失败-数据库查询失败", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
		os.Exit(1)
	}
	err = model.DB.Where("status = ?", 1).Find(&missions).Error
	if err != nil {
		zap.L().Error("任务启动失败-数据库查询失败", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
		os.Exit(1)
	}
	tx := model.DB.Model(&model.TaskRecord{}).
		Where("status = ?", 0).
		UpdateColumns(map[string]interface{}{
			"status":  2,
			"message": "任务执行被中断，请重新执行",
		})
	if tx.RowsAffected > 0 {
		zap.L().Error("发现被中断任务，请查看任务运行记录", zap.String("service", "system"), zap.String("name", config.Ip))
	}
	if tx.Error != nil {
		zap.L().Error("任务启动失败-数据库查询失败", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(tx.Error))
		os.Exit(1)
	}
	cr.Start()
	if len(missions) == 0 {
		zap.L().Info("系统任务已启动", zap.String("service", "system"), zap.String("name", config.Ip))
		return
	}
	for _, mission := range missions {
		if mission.Cron == "manual" {
			continue
		}
		if err := ScheduleMission(&mission); err != nil {
			zap.L().Error("任务调度失败，已跳过", zap.String("service", "system"), zap.String("name", mission.ID), zap.Error(err))
			continue
		}
	}
	zap.L().Info("系统任务已启动", zap.String("service", "system"), zap.String("name", config.Ip))
}

func middleware(missionID string, runBy string) {
	var mission model.Task
	runtime := model.CustomTime{Time: time.Now()}
	if err := model.DB.Where("id = ?", missionID).First(&mission).Error; err != nil {
		zap.L().Error("任务查询失败", zap.String("service", "task"), zap.String("name", missionID), zap.Error(err))
		return
	}
	zap.L().Info(fmt.Sprintf("开始执行任务 %s", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID))
	if mission.Status != 1 && runBy == "system" {
		zap.L().Error("系统错误，执行未调度任务", zap.String("service", "task"), zap.String("name", mission.ID))
		return
	}
	if mission.IsRunning && runBy == "system" {
		zap.L().Info("任务正在运行中,下个周期将再次尝试", zap.String("service", "task"), zap.String("name", mission.ID))
		return
	}

	// 标记任务开始运行（manual 模式下 is_running 已由 RunMissionManual 原子写入，此处写入无副作用）
	mission.IsRunning = true
	mission.LastRunTime = &runtime
	if err := model.DB.Save(&mission).Error; err != nil {
		zap.L().Error("任务状态更新失败", zap.String("service", "task"), zap.String("name", mission.ID), zap.Error(err))
		return
	}

	// 结构化变量替换（只替换 Params.Value，不破坏 JSON 结构）
	missionRun := mission
	replacedData, varErr := ReplaceVariables(mission.Data)
	if varErr != nil {
		zap.L().Error("变量解析错误", zap.String("service", "task"), zap.String("name", mission.ID), zap.Error(varErr))
		mission.IsRunning = false
		mission.ErrMsg = varErr.Error()
		model.DB.Save(&mission)
		return
	}
	if replacedData != nil {
		missionRun.Data = replacedData
		zap.L().Info(fmt.Sprintf("任务 %s 变量替换成功", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID))
	}

	// 执行任务业务函数
	err := RunTask(missionRun, runBy)

	// 记录结束时间并更新状态
	endTime := model.CustomTime{Time: time.Now()}
	mission.LastEndTime = &endTime
	if err != nil {
		mission.ErrMsg = err.Error()
		if runBy == "system" {
			cancelMission(&mission, 2)
			zap.L().Error(fmt.Sprintf("任务 %s 执行失败,已自动暂停", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID), zap.Error(err))
		}
	} else {
		mission.LastSuccessTime = &runtime
		mission.ErrMsg = "Success"
		zap.L().Info(fmt.Sprintf("任务 %s 执行成功", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID))
	}
	mission.IsRunning = false
	model.DB.Save(&mission)
}

func CancelMission(mission *model.Task) {
	cancelMission(mission, 0)
}
func cancelMission(mission *model.Task, status int) {
	mission.Status = status
	if mission.EntryID != nil {
		cr.Remove(cron.EntryID(*mission.EntryID))
		mission.EntryID = nil
	}
}

func ScheduleMission(mission *model.Task) error {
	if mission.Cron == "manual" {
		return errors.New("手动任务不能被调度")
	}
	if _, err := cron.ParseStandard(mission.Cron); err != nil {
		return errors.New("任务的表达式无效")
	}
	EntryID, err := cr.AddFunc(mission.Cron, func() {
		middleware(mission.ID, "system")
	})
	if err != nil {
		return err
	}
	mission.Status = 1
	mission.IsRunning = false
	eId := int(EntryID)
	mission.EntryID = &eId
	if err := model.DB.Save(mission).Error; err != nil {
		cr.Remove(cron.EntryID(eId))
		return err
	}
	return nil
}
func RunMissionManual(missionID string) error {
	// 用条件更新原子地抢占 is_running 标志，避免并发重复执行
	tx := model.DB.Model(&model.Task{}).
		Where("id = ? AND is_running = ?", missionID, false).
		UpdateColumn("is_running", true)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("任务正在运行中")
	}
	go middleware(missionID, "manual")
	return nil
}

func RunTask(mission model.Task, runBy string) (err error) {
	var missionRecord = model.TaskRecord{
		RunBy:  runBy,
		TaskID: mission.ID,
		Status: 0,
		StartTime: &model.CustomTime{
			Time: time.Now(),
		},
		Message: "",
		Data:    mission.Data,
	}
	err = model.DB.Model(&model.TaskRecord{}).Create(&missionRecord).Error
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			if runState.IsManualCancelled(missionRecord.ID) {
				missionRecord.Status = 2
				missionRecord.Message = "任务被手动中止"
			} else {
				missionRecord.Status = 1
				missionRecord.Message = "ok"
			}
		} else {
			missionRecord.Status = 2
			missionRecord.Message = err.Error()
		}
		missionRecord.EndTime = &model.CustomTime{
			Time: time.Now(),
		}
		model.DB.Save(&missionRecord)
	}()

	cfg := pipeline.Config{
		BatchSize:   config.Config.Pipeline.BatchSize,
		ChannelSize: config.Config.Pipeline.ChannelSize,
	}
	if mission.ID == "" {
		return errors.New("任务不存在")
	}
	dsResolver := newDatasourceResolver()

	// BeforeExecutor
	var BeforeExecutorConfig *map[string]string
	var BeforeExecutor *executor.Executor
	var BeforeExecutorDatasource datasource.Datasource
	if mission.Data.BeforeExecute != nil {
		beforeExecutorStore, err := factory.CreateExecutor(mission.Data.BeforeExecute.Type)
		if err != nil {
			return err
		}
		BeforeExecutor = &beforeExecutorStore.Handle
		cfg := buildConfig(mission.Data.BeforeExecute.Params)
		BeforeExecutorConfig = &cfg
		if beforeExecutorStore.Datasource != nil {
			ds, err := dsResolver.initDatasource(mission.Data.BeforeExecute.DataSource, *beforeExecutorStore.Datasource)
			if err != nil {
				return err
			}
			BeforeExecutorDatasource = ds
		}
	}

	// Source
	SourceStore, err := factory.CreateSource(mission.Data.Source.Type)
	if err != nil {
		return err
	}
	SourceConfig := buildConfig(mission.Data.Source.Params)
	var SourceDatasource datasource.Datasource
	if SourceStore.Datasource != nil {
		ds, err := dsResolver.initDatasource(mission.Data.Source.DataSource, *SourceStore.Datasource)
		if err != nil {
			return err
		}
		SourceDatasource = ds
	}

	// Sink
	SinkStore, err := factory.CreateSink(mission.Data.Sinks.Type)
	if err != nil {
		return err
	}
	SinkConfig := buildConfig(mission.Data.Sinks.Params)
	var SinkDatasource datasource.Datasource
	if SinkStore.Datasource != nil {
		ds, err := dsResolver.initDatasource(mission.Data.Sinks.DataSource, *SinkStore.Datasource)
		if err != nil {
			return err
		}
		SinkDatasource = ds
	}

	// Processors
	processors := make([]processor.Processor, 0, len(mission.Data.Processors))
	processorsConfigs := make([]pipeline.ProcessorConfig, 0, len(mission.Data.Processors))
	for _, pConfig := range mission.Data.Processors {
		p, err := factory.CreateProcessor(pConfig.Type)
		if err != nil {
			return err
		}
		processors = append(processors, p.Handle)
		processorsConfigs = append(processorsConfigs, pipeline.ProcessorConfig{
			Type:   pConfig.Type,
			Params: buildConfig(pConfig.Params),
		})
	}

	// AfterExecutor
	var AfterExecutorConfig *map[string]string
	var AfterExecutor *executor.Executor
	var AfterExecutorDatasource datasource.Datasource
	if mission.Data.AfterExecute != nil {
		afterExecuteStore, err := factory.CreateExecutor(mission.Data.AfterExecute.Type)
		if err != nil {
			return err
		}
		AfterExecutor = &afterExecuteStore.Handle
		cfg := buildConfig(mission.Data.AfterExecute.Params)
		AfterExecutorConfig = &cfg
		if afterExecuteStore.Datasource != nil {
			ds, err := dsResolver.initDatasource(mission.Data.AfterExecute.DataSource, *afterExecuteStore.Datasource)
			if err != nil {
				return err
			}
			AfterExecutorDatasource = ds
		}
	}

	engine := pipeline.NewEngine(missionRecord.ID, BeforeExecutor, BeforeExecutorDatasource, SourceStore.Handle, SourceDatasource, processors, SinkStore.Handle, SinkDatasource, cfg, AfterExecutor, AfterExecutorDatasource)
	ctx := context.Background()
	runCtx, cancel := context.WithCancel(ctx)
	runState.SetCancel(missionRecord.ID, cancel)
	defer runState.RemoveCancel(missionRecord.ID)
	defer cancel()
	if err := engine.Run(missionRecord.ID, runCtx, BeforeExecutorConfig, SourceConfig, processorsConfigs, SinkConfig, AfterExecutorConfig); err != nil {
		return err
	}
	return nil
}

func CancelMissionRecord(ID string) error {
	return runState.CancelRecord(ID)
}

func GetValueByName(name string) (string, error) {
	var variable model.Variable
	err := model.DB.Where("`name` = ?", name).First(&variable).Error
	if err != nil {
		return "", errors.New("variable does not exist")
	}
	v, err := factory.CreateVariable(variable.Type)
	if err != nil {
		return "", errors.New("variable type does not exist")
	}
	var variableConfig = make(map[string]string)
	if variable.Value != nil {
		for _, param := range *variable.Value {
			variableConfig[param.Key] = param.Value
		}
	}
	var vDatasource datasource.Datasource
	if v.Datasource != nil {
		if variable.DataSourceID == nil || *variable.DataSourceID == "" {
			return "", errors.New("variable data source does not exist")
		}
		var dataSourceData model.DataSource
		err := model.DB.Where("`id` = ?", variable.DataSourceID).First(&dataSourceData).Error
		if err != nil {
			return "", errors.New("variable data source does not exist")
		}
		var dataSourceDataConfig = make(map[string]string)
		for _, param := range dataSourceData.Data {
			dataSourceDataConfig[param.Key] = param.Value
		}
		dsName := *v.Datasource
		if dsName != dataSourceData.Type {
			return "", errors.New("variable data source type error")
		}
		dsStore, err := factory.CreateDataSource(dsName)
		if err != nil {
			return "", errors.New("variable data source type does not exist")
		}
		_, err = pipeline.HandleInternalConfig(&dataSourceDataConfig)
		if err != nil {
			return "", err
		}
		err = dsStore.Handle.Init(dataSourceDataConfig)
		if err != nil {
			return "", errors.New("variable data source open error")
		}
		vDatasource = dsStore.Handle
	}
	handle := v.Handle
	return handle.Get(context.Background(), variableConfig, vDatasource)
}
