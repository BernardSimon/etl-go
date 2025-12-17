package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/core/executor"
	"github.com/BernardSimon/etl-go/etl/core/procrssor"
	"github.com/BernardSimon/etl-go/etl/core/sink"
	"github.com/BernardSimon/etl-go/etl/core/source"
	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	"github.com/BernardSimon/etl-go/server/config"
	"github.com/BernardSimon/etl-go/server/model"
	_type "github.com/BernardSimon/etl-go/server/type"
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
		zap.L().Error("任务启动失败-数据库查询失败", zap.String("service", "system"), zap.String("name", config.Ip), zap.Error(err))
		os.Exit(1)
	}
	if len(missions) == 0 {
		zap.L().Info("系统任务已启动", zap.String("service", "system"), zap.String("name", config.Ip))
		return
	}
	for _, mission := range missions {
		if mission.Cron == "manual" {
			continue
		}
		err := ScheduleMission(&mission)
		if err != nil {
			panic(err)
		}
	}
	cr.Start()
	zap.L().Info("系统任务已启动", zap.String("service", "system"), zap.String("name", config.Ip))
}

func middleware(missionID string, runBy string) {
	var mission model.Task
	runtime := model.CustomTime{Time: time.Now()}
	model.DB.Where("id = ?", missionID).First(&mission)
	zap.L().Info(fmt.Sprintf("开始执行任务 %s", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID))
	if mission.Status != 1 && runBy == "system" {
		zap.L().Error("系统错误，执行未调度任务", zap.String("service", "task"), zap.String("name", mission.ID))
		return
	}
	defer model.DB.Save(&mission)
	if !mission.IsRunning {

		//记录开始状态时间
		mission.IsRunning = true
		mission.LastRunTime = &runtime
		model.DB.Save(&mission)
		// 初始化map
		variableList := make(map[string]string)
		// 或者更好的方式是直接处理变量替换
		rawData, _ := json.Marshal(mission.Data)
		stringData := string(rawData)

		re := regexp.MustCompile(`\$\{[^}]*}`)
		matches := re.FindAllString(stringData, -1)
		missionRun := mission
		if len(matches) != 0 {
			// 使用map去重并获取变量值
			for _, match := range matches {
				if _, exists := variableList[match]; !exists {
					vName := strings.TrimPrefix(match, "${")
					vName = strings.TrimSuffix(vName, "}")
					value, err := GetValueByName(vName)
					if err != nil {
						zap.L().Error("变量解析错误:"+match, zap.String("service", "task"), zap.String("name", mission.ID), zap.Error(err))
						return
						// 可以考虑是否继续执行或者返回错误
					}
					variableList[match] = value
				}
			}
			// 变量替换
			for placeholder, value := range variableList {
				stringData = strings.ReplaceAll(stringData, placeholder, value)
			}

			// 更新任务配置
			var replacedData _type.TaskData
			err := json.Unmarshal([]byte(stringData), &replacedData)
			if err != nil {
				zap.L().Error("任务变量配置解析错误", zap.String("service", "task"), zap.String("name", mission.ID), zap.Error(err))
				return
			}
			zap.L().Info(fmt.Sprintf("任务 %s 变量替换成功", mission.Name), zap.String("service", "task"), zap.String("name", mission.ID), zap.Any("content", variableList))
			missionRun.Data = &replacedData
		}
		//执行任务业务函数
		err := RunTask(missionRun, runBy)
		//记录结束时间
		endTime := model.CustomTime{Time: time.Now()}
		mission.LastEndTime = &endTime
		if err != nil {
			//记录错误
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
	} else {
		zap.L().Info("任务正在运行中,下个周期将再次尝试", zap.String("service", "task"), zap.String("name", mission.ID))
		return
	}
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
	model.DB.Save(mission)
	return nil
}
func RunMissionManual(missionID string) error {
	var isRunning bool
	model.DB.Model(&model.Task{}).Where("id = ?", missionID).Select("is_running").Find(&isRunning)
	if isRunning {
		return errors.New("任务正在运行中")
	}
	go middleware(missionID, "manual")
	return nil
}

var runCtxMap = make(map[string]context.CancelFunc)

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
			if _, exist := ManualCancelMap[missionRecord.ID]; exist {
				missionRecord.Status = 2
				missionRecord.Message = "任务被手动中止"
				delete(ManualCancelMap, mission.ID)
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

	var cfg pipeline.Config
	if err != nil {
		return
	}
	if mission.ID == "" {
		return errors.New("任务不存在")
	}
	var BeforeExecutorConfig *map[string]string
	var BeforeExecutor *executor.Executor
	var BeforeExecutorDatasource *datasource.Datasource
	if mission.Data.BeforeExecute != nil {
		beforeExecutorConfig := make(map[string]string)
		beforeExecutorStore, err := factory.CreateExecutor(mission.Data.BeforeExecute.Type)
		if err != nil {
			return err
		}
		BeforeExecutor = &beforeExecutorStore.Handle
		for _, param := range mission.Data.BeforeExecute.Params {
			beforeExecutorConfig[param.Key] = param.Value
		}
		if beforeExecutorStore.Datasource != nil {
			if mission.Data.BeforeExecute.DataSource == nil {
				return errors.New("数据源未指定")
			}
			var dataSourceData model.DataSource
			err = model.DB.Where("`id` = ?", mission.Data.BeforeExecute.DataSource).First(&dataSourceData).Error
			if err != nil {
				return errors.New("数据源不存在")
			}

			var dataSourceDataConfig = make(map[string]string)
			for _, param := range dataSourceData.Data {
				dataSourceDataConfig[param.Key] = param.Value
			}
			dsName := *beforeExecutorStore.Datasource
			if dataSourceData.Type != dsName {
				return errors.New("数据源类型错误")
			}
			dsStore, err := factory.CreateDataSource(dsName)
			if err != nil {
				return errors.New("数据源类型未找到")
			}
			_, err = pipeline.HandleInternalConfig(&dataSourceDataConfig)
			if err != nil {
				return err
			}
			err = dsStore.Handle.Init(dataSourceDataConfig)
			if err != nil {
				return err
			}
			BeforeExecutorDatasource = &dsStore.Handle
		}
		BeforeExecutorConfig = &beforeExecutorConfig
	}

	var SourceConfig = make(map[string]string)
	var Source source.Source
	var SourceDatasource *datasource.Datasource
	SourceStore, err := factory.CreateSource(mission.Data.Source.Type)
	if err != nil {
		return err
	}
	for _, param := range mission.Data.Source.Params {
		SourceConfig[param.Key] = param.Value
	}
	if SourceStore.Datasource != nil {
		if mission.Data.Source.DataSource == nil {
			return errors.New("数据源未指定")
		}
		var dataSourceData model.DataSource
		err = model.DB.Where("`id` = ?", mission.Data.Source.DataSource).First(&dataSourceData).Error
		if err != nil {
			return errors.New("数据源不存在")
		}

		var dataSourceDataConfig = make(map[string]string)
		for _, param := range dataSourceData.Data {
			dataSourceDataConfig[param.Key] = param.Value
		}
		dsName := *SourceStore.Datasource
		if dsName != dataSourceData.Type {
			return errors.New("数据源类型错误")
		}
		dsStore, err := factory.CreateDataSource(dsName)
		if err != nil {
			return errors.New("数据源类型未找到")
		}
		_, err = pipeline.HandleInternalConfig(&dataSourceDataConfig)
		if err != nil {
			return err
		}
		err = dsStore.Handle.Init(dataSourceDataConfig)
		if err != nil {
			return err
		}
		SourceDatasource = &dsStore.Handle
	}
	Source = SourceStore.Handle
	var SinkConfig = make(map[string]string)
	var Sink sink.Sink
	var SinkDatasource *datasource.Datasource
	SinkStore, err := factory.CreateSink(mission.Data.Sinks.Type)
	if err != nil {
		return err
	}
	for _, param := range mission.Data.Sinks.Params {
		SinkConfig[param.Key] = param.Value
	}
	if SinkStore.Datasource != nil {
		if mission.Data.Sinks.DataSource == nil {
			return errors.New("数据源未指定")
		}
		var dataSourceData model.DataSource
		err = model.DB.Where("`id` = ?", mission.Data.Sinks.DataSource).First(&dataSourceData).Error
		if err != nil {
			return errors.New("数据源不存在")
		}
		var dataSourceDataConfig = make(map[string]string)
		for _, param := range dataSourceData.Data {
			dataSourceDataConfig[param.Key] = param.Value
		}
		dsName := *SinkStore.Datasource
		if dsName != dataSourceData.Type {
			return errors.New("数据源类型错误")
		}
		dataSourceStore, err := factory.CreateDataSource(dsName)
		if err != nil {
			return errors.New("数据源类型未找到")
		}
		_, err = pipeline.HandleInternalConfig(&dataSourceDataConfig)
		if err != nil {
			return err
		}
		err = dataSourceStore.Handle.Init(dataSourceDataConfig)
		if err != nil {
			return err
		}
		SinkDatasource = &dataSourceStore.Handle
	}
	Sink = SinkStore.Handle

	processors := make([]procrssor.Processor, 0)
	processorsConfigs := make([]pipeline.ProcessorConfig, 0)
	for _, pConfig := range mission.Data.Processors {
		p, err := factory.CreateProcessor(pConfig.Type)
		if err != nil {
			return err
		}
		processors = append(processors, p.Handle)
		var ProcessConfig = make(map[string]string)
		for _, param := range pConfig.Params {
			ProcessConfig[param.Key] = param.Value
		}
		processorsConfigs = append(processorsConfigs, pipeline.ProcessorConfig{
			Type:   pConfig.Type,
			Params: ProcessConfig,
		})
	}

	var AfterExecutorConfig *map[string]string
	var AfterExecutor *executor.Executor
	var AfterExecutorDatasource *datasource.Datasource
	if mission.Data.AfterExecute != nil {
		afterExecuteConfig := make(map[string]string)
		afterExecuteStore, err := factory.CreateExecutor(mission.Data.AfterExecute.Type)
		if err != nil {
			return err
		}
		AfterExecutor = &afterExecuteStore.Handle
		for _, param := range mission.Data.AfterExecute.Params {
			afterExecuteConfig[param.Key] = param.Value
		}
		AfterExecutorConfig = &afterExecuteConfig
		if afterExecuteStore.Datasource != nil {
			if mission.Data.AfterExecute.DataSource == nil {
				return errors.New("数据源未指定")
			}
			var dataSourceData model.DataSource
			err = model.DB.Where("`id` = ?", mission.Data.AfterExecute.DataSource).First(&dataSourceData).Error
			if err != nil {
				return errors.New("数据源不存在")
			}

			var dataSourceDataConfig = make(map[string]string)
			for _, param := range dataSourceData.Data {
				dataSourceDataConfig[param.Key] = param.Value
			}
			dsName := *afterExecuteStore.Datasource
			if dsName != dataSourceData.Type {
				return errors.New("数据源类型错误")
			}
			dsStore, err := factory.CreateDataSource(dsName)
			if err != nil {
				return errors.New("数据源类型未找到")
			}
			_, err = pipeline.HandleInternalConfig(&dataSourceDataConfig)
			if err != nil {
				return err
			}
			err = dsStore.Handle.Init(dataSourceDataConfig)
			if err != nil {
				return err
			}
			AfterExecutorDatasource = &dsStore.Handle
		}
	}
	engine := pipeline.NewEngine(missionRecord.ID, BeforeExecutor, BeforeExecutorDatasource, Source, SourceDatasource, processors, Sink, SinkDatasource, cfg, AfterExecutor, AfterExecutorDatasource)
	ctx := context.Background()
	runCtx, cancel := context.WithCancel(ctx)
	defer delete(runCtxMap, missionRecord.ID)
	defer cancel()
	runCtxMap[missionRecord.ID] = cancel
	if err := engine.Run(missionRecord.ID, runCtx, BeforeExecutorConfig, SourceConfig, processorsConfigs, SinkConfig, AfterExecutorConfig); err != nil {
		return err
	}
	return nil
}

var ManualCancelMap = make(map[string]string)

func CancelMissionRecord(ID string) error {
	if cancel, ok := runCtxMap[ID]; ok {
		cancel()
		ManualCancelMap[ID] = "cancel"
		return nil
	} else {
		return errors.New("任务不存在或状态不可停止")
	}
}

func GetValueByName(name string) (string, error) {
	var variable model.Variable
	err := model.DB.Where("`name` = ?", name).Preload("DataSource").First(&variable).Error
	if err != nil {
		return "", errors.New("variable does not exist")
	}
	v, err := factory.CreateVariable(variable.DataSource.Type)
	if err != nil {
		return "", errors.New("variable type does not exist")
	}
	var variableConfig = make(map[string]string)
	for _, param := range *variable.Value {
		variableConfig[param.Key] = param.Value
	}
	var vDatasource *datasource.Datasource
	if v.Datasource != nil {
		if variable.DataSource.ID == "" {
			return "", errors.New("variable data source does not exist")
		}
		var dataSourceData model.DataSource
		err := model.DB.Where("`name` = ?", variable.DataSource.Name).Find(&dataSourceData).Error
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
		vDatasource = &dsStore.Handle
	}
	handle := v.Handle
	return handle.Get(variableConfig, vDatasource)
}
