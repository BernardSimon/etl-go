package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	fileUtil "github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)


func GetDataSourceTypeList(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	list := factory.GetDatasourceTypeList()
	var resp = make([]types.GetDataSourceTypeListResponse, 0)
	for _, v := range list {
		store, _ := factory.CreateDataSource(v)
		resp = append(resp, types.GetDataSourceTypeListResponse{
			Type:   v,
			Params: normalizeParams(store.Params),
		})
	}
	return map[string]interface{}{
		"list": resp,
	}, nil
}

func keyValuesToMap(values types.KeyValues) map[string]string {
	config := make(map[string]string, len(values))
	for _, item := range values {
		config[item.Key] = item.Value
	}
	return config
}

func resolveDatasourceConfig(config map[string]string) error {
	if fileID, ok := config["file_id"]; ok && fileID != "" {
		filePath, err := fileUtil.GetFilePath(fileID)
		if err != nil {
			return err
		}
		config["file_path"] = filePath
	}
	return nil
}

func TestDataSource(_ *struct{}, body *types.TestDataSourceRequest, lang string) (interface{}, error) {
	store, exists := factory.CreateDataSource(body.Type)
	if exists != nil {
		return nil, errors.New("invalid Datasource type")
	}

	config := keyValuesToMap(body.Data)
	if err := resolveDatasourceConfig(config); err != nil {
		return nil, err
	}
	for _, v := range store.Params {
		if v.Required && config[v.Key] == "" {
			return nil, errors.New("datasource params error")
		}
	}
	if err := store.Handle.Init(config); err != nil {
		return nil, errors.New("failed to test datasource connection")
	}
	_ = store.Handle.Close()

	return i18n.Translate(lang, "datasource connection test success"), nil
}

func NewDataSource(_ *struct{}, body *types.NewDataSourceRequest, lang string) (interface{}, error) {
	store, exists := factory.CreateDataSource(body.Type)
	if exists != nil {
		return nil, errors.New("invalid Datasource type")
	}
	var existingRecord model.DataSource
	err := model.DB.Where("name = ?", body.Name).Find(&existingRecord).Error
	if err != nil {
		return nil, errors.New("illegal command")
	}
	if body.Edit {
		if existingRecord.ID != body.ID && existingRecord.ID != "" {
			return nil, errors.New("datasource name already exist")
		}
	} else {
		if existingRecord.ID != "" {
			return nil, errors.New("datasource name already exist")
		}
	}
	for _, v := range store.Params {
		match := false
		for _, v1 := range body.Data {
			if v.Key == v1.Key {
				if v.Required {
					if v1.Value == "" {
						return nil, errors.New("datasource params error")
					}
				}
				match = true
				break
			}
		}
		if !match {
			return nil, errors.New("datasource params error")
		}
	}
	var existingRecord1 model.DataSource
	if body.Edit {
		if err := model.DB.Where("id = ?", body.ID).First(&existingRecord1).Error; err != nil {
			return nil, errors.New("illegal command")
		}
	}
	existingRecord1.Data = body.Data
	existingRecord1.Name = body.Name
	existingRecord1.Type = body.Type
	if err := model.DB.Save(&existingRecord1).Error; err != nil {
		return nil, errors.New("failed to save datasource")
	}
	return i18n.Translate(lang, "success"), nil
}

func GetDataSourceList(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	var dataSourceList []model.DataSource
	err := model.DB.Select("id", "name", "type", "updated_at", "data").Order("created_at desc").Find(&dataSourceList).Error
	if err != nil {
		return nil, errors.New("failed to get datasource list")
	}
	return map[string]interface{}{
		"list": dataSourceList,
	}, nil
}

// GetDataSourceSchema 初始化 datasource 并查询其表结构（仅支持 SQL 类型）
func GetDataSourceSchema(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var ds model.DataSource
	if err := model.DB.Where("id = ?", uri.Id).First(&ds).Error; err != nil {
		return nil, errors.New("datasource not found")
	}

	// 构建配置 map
	config := make(map[string]string, len(ds.Data))
	for _, kv := range ds.Data {
		config[kv.Key] = kv.Value
	}
	// 解析 file_id → file_path（SQLite 需要）
	if err := resolveDatasourceConfig(config); err != nil {
		return nil, err
	}
	// 处理 HandleInternalConfig 中的 file_id 逻辑（与 pipeline 保持一致）
	if _, err := pipeline.HandleInternalConfig(&config); err != nil {
		return nil, err
	}

	// 创建并初始化 datasource 实例
	store, err := factory.CreateDataSource(ds.Type)
	if err != nil {
		return nil, errors.New("unsupported datasource type")
	}
	if err := store.Handle.Init(config); err != nil {
		return nil, errors.New("failed to connect to datasource")
	}
	defer store.Handle.Close()

	// 直接调用 ListTables，不支持 schema 的实现返回空切片
	tables, err := store.Handle.ListTables()
	if err != nil {
		return nil, err
	}

	// 转换为响应类型
	resp := types.GetDataSourceSchemaResponse{
		Tables: make([]types.DataSourceTable, 0, len(tables)),
	}
	for _, t := range tables {
		cols := make([]types.DataSourceColumn, 0, len(t.Columns))
		for _, c := range t.Columns {
			cols = append(cols, types.DataSourceColumn{
				Name:     c.Name,
				Type:     c.Type,
				Nullable: c.Nullable,
			})
		}
		resp.Tables = append(resp.Tables, types.DataSourceTable{
			Name:    t.Name,
			Columns: cols,
		})
	}
	return resp, nil
}

func DeleteDataSource(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var dataSourceRecord model.DataSource
	model.DB.Where("id = ?", uri.Id).First(&dataSourceRecord)
	if dataSourceRecord.ID == "" {
		return nil, errors.New("datasource handle not found")
	}
	if err := model.DB.Delete(&dataSourceRecord).Error; err != nil {
		return nil, errors.New("failed to delete datasource record")
	}
	return i18n.Translate(lang, "success"), nil
}
