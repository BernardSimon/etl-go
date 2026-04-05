package task

import (
	"errors"
	"fmt"

	"github.com/BernardSimon/etl-go/etl/core/datasource"
	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	"github.com/BernardSimon/etl-go/server/model"
)

type datasourceResolver struct {
	cache map[string]*datasource.Shared
}

func newDatasourceResolver() *datasourceResolver {
	return &datasourceResolver{
		cache: make(map[string]*datasource.Shared),
	}
}

// buildConfig 从 Params 切片构建 map[string]string 配置。
func buildConfig(params []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}) map[string]string {
	config := make(map[string]string, len(params))
	for _, p := range params {
		config[p.Key] = p.Value
	}
	return config
}

// initDatasource 统一处理组件的数据源初始化流程：
// 查询DB → 构建config → 类型校验 → 工厂创建 → HandleInternalConfig → Init
func (r *datasourceResolver) initDatasource(dsID interface{}, expectedType string) (datasource.Datasource, error) {
	if dsID == nil {
		return nil, errors.New("数据源未指定")
	}
	cacheKey := fmt.Sprintf("%s:%v", expectedType, dsID)
	if shared, ok := r.cache[cacheKey]; ok {
		return shared.Lease(), nil
	}

	var dataSourceData model.DataSource
	if err := model.DB.Where("`id` = ?", dsID).First(&dataSourceData).Error; err != nil {
		return nil, errors.New("数据源不存在")
	}

	if dataSourceData.Type != expectedType {
		return nil, errors.New("数据源类型错误")
	}

	dsConfig := make(map[string]string, len(dataSourceData.Data))
	for _, p := range dataSourceData.Data {
		dsConfig[p.Key] = p.Value
	}

	dsStore, err := factory.CreateDataSource(expectedType)
	if err != nil {
		return nil, errors.New("数据源类型未找到")
	}

	if _, err := pipeline.HandleInternalConfig(&dsConfig); err != nil {
		return nil, err
	}

	if err := dsStore.Handle.Init(dsConfig); err != nil {
		return nil, err
	}

	shared, err := datasource.NewShared(dsStore.Handle)
	if err != nil {
		return nil, err
	}
	r.cache[cacheKey] = shared
	return shared.Lease(), nil
}
