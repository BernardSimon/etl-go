package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/etl/pipeline"
	serverConfig "github.com/BernardSimon/etl-go/server/config"
	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	fileUtil "github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

const maskedValue = "****"

// maskSensitiveData 根据数据源类型的 Params.Mask 标记，将敏感字段值替换为 ****
func maskSensitiveData(dsType string, data types.KeyValues) types.KeyValues {
	maskSet := make(map[string]bool)
	if store, err := factory.CreateDataSource(dsType); err == nil {
		for _, p := range store.Params {
			if p.Mask {
				maskSet[p.Key] = true
			}
		}
	}
	masked := make(types.KeyValues, len(data))
	for i, kv := range data {
		masked[i] = kv
		if maskSet[kv.Key] && kv.Value != "" {
			masked[i].Value = maskedValue
		}
	}
	return masked
}

// restoreMaskedFields 将 body.Data 中值为 **** 的敏感字段替换回数据库中的原始值
func restoreMaskedFields(bodyData types.KeyValues, originalData types.KeyValues) {
	originalMap := make(map[string]string, len(originalData))
	for _, kv := range originalData {
		originalMap[kv.Key] = kv.Value
	}
	for i, kv := range bodyData {
		if kv.Value == maskedValue {
			bodyData[i].Value = originalMap[kv.Key]
		}
	}
}


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

	// 编辑模式下，若用户未修改密码（值为 ****），从数据库还原真实值
	if body.ID != "" {
		var ds model.DataSource
		if err := model.DB.Where("id = ?", body.ID).First(&ds).Error; err == nil {
			restoreMaskedFields(body.Data, ds.Data)
		}
	}

	dsConfig := keyValuesToMap(body.Data)
	if err := resolveDatasourceConfig(dsConfig); err != nil {
		return nil, err
	}
	serverConfig.ApplyDatasourcePoolConfig(dsConfig)
	for _, v := range store.Params {
		if v.Required && dsConfig[v.Key] == "" {
			return nil, errors.New("datasource params error")
		}
	}
	if err := store.Handle.Init(dsConfig); err != nil {
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
		// 用户未修改的密码字段值为 ****，还原为数据库中的原始值
		restoreMaskedFields(body.Data, existingRecord1.Data)
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
	err := model.DB.Select("id", "name", "type", "updated_at").Order("created_at desc").Find(&dataSourceList).Error
	if err != nil {
		return nil, errors.New("failed to get datasource list")
	}
	return map[string]interface{}{
		"list": dataSourceList,
	}, nil
}

// GetDataSourceById 返回单个数据源配置，敏感字段（密码等）以 **** 脱敏
func GetDataSourceById(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var ds model.DataSource
	if err := model.DB.Where("id = ?", uri.Id).First(&ds).Error; err != nil {
		return nil, errors.New("datasource not found")
	}
	return map[string]interface{}{
		"id":         ds.ID,
		"name":       ds.Name,
		"type":       ds.Type,
		"updated_at": ds.UpdatedAt,
		"data":       maskSensitiveData(ds.Type, ds.Data),
	}, nil
}

// GetDataSourceSchema 初始化 datasource 并查询其表结构（仅支持 SQL 类型）
func GetDataSourceSchema(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var ds model.DataSource
	if err := model.DB.Where("id = ?", uri.Id).First(&ds).Error; err != nil {
		return nil, errors.New("datasource not found")
	}

	// 构建配置 map
	dsConfig := make(map[string]string, len(ds.Data))
	for _, kv := range ds.Data {
		dsConfig[kv.Key] = kv.Value
	}
	// 解析 file_id → file_path（SQLite 需要）
	if err := resolveDatasourceConfig(dsConfig); err != nil {
		return nil, err
	}
	// 处理 HandleInternalConfig 中的 file_id 逻辑（与 pipeline 保持一致）
	if _, err := pipeline.HandleInternalConfig(&dsConfig); err != nil {
		return nil, err
	}
	serverConfig.ApplyDatasourcePoolConfig(dsConfig)

	// 创建并初始化 datasource 实例
	store, err := factory.CreateDataSource(ds.Type)
	if err != nil {
		return nil, errors.New("unsupported datasource type")
	}
	if err := store.Handle.Init(dsConfig); err != nil {
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
