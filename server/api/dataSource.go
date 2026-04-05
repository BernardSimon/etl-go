package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	fileUtil "github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetDataSourceTypeList(_ *interface{}, _ string) (interface{}, error) {
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

func TestDataSource(req *types.TestDataSourceRequest, lang string) (interface{}, error) {
	store, exists := factory.CreateDataSource(req.Type)
	if exists != nil {
		return nil, errors.New("invalid Datasource type")
	}

	config := keyValuesToMap(req.Data)
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

func NewDataSource(req *types.NewDataSourceRequest, lang string) (interface{}, error) {
	store, exists := factory.CreateDataSource(req.Type)
	if exists != nil {
		return nil, errors.New("invalid Datasource type")
	}
	var existingRecord model.DataSource
	err := model.DB.Where("name = ?", req.Name).Find(&existingRecord).Error
	if err != nil {
		return nil, errors.New("illegal command")
	}
	if req.Edit {
		if existingRecord.ID != req.ID && existingRecord.ID != "" {
			return nil, errors.New("datasource name already exist")
		}
	} else {
		if existingRecord.ID != "" {
			return nil, errors.New("datasource name already exist")
		}
	}
	for _, v := range store.Params {
		match := false
		for _, v1 := range req.Data {
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
	if req.Edit {
		if err := model.DB.Where("id = ?", req.ID).First(&existingRecord1).Error; err != nil {
			return nil, errors.New("illegal command")
		}
	}
	existingRecord1.Data = req.Data
	existingRecord1.Name = req.Name
	existingRecord1.Type = req.Type
	if err := model.DB.Save(&existingRecord1).Error; err != nil {
		return nil, errors.New("failed to save datasource")
	}
	return i18n.Translate(lang, "success"), nil
}

func GetDataSourceList(_ *interface{}, _ string) (interface{}, error) {
	var dataSourceList []model.DataSource
	err := model.DB.Select("id", "name", "type", "updated_at", "data").Order("created_at desc").Find(&dataSourceList).Error
	if err != nil {
		return nil, errors.New("failed to get datasource list")
	}
	return map[string]interface{}{
		"list": dataSourceList,
	}, nil
}

func DeleteDataSource(req *types.DeleteDataSourceRequest, lang string) (interface{}, error) {
	var dataSourceRecord model.DataSource
	model.DB.Where("id = ?", req.Id).First(&dataSourceRecord)
	if dataSourceRecord.ID == "" {
		return nil, errors.New("datasource handle not found")
	}
	if err := model.DB.Delete(&dataSourceRecord).Error; err != nil {
		return nil, errors.New("failed to delete datasource record")
	}
	return i18n.Translate(lang, "success"), nil
}
