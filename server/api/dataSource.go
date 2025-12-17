package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	_type "github.com/BernardSimon/etl-go/server/type"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetDataSourceTypeList(_ *interface{}, _ string) (interface{}, error) {
	list := factory.GetDatasourceTypeList()
	var resp = make([]_type.GetDataSourceTypeListResponse, 0)
	for _, v := range list {
		store, _ := factory.CreateDataSource(v)
		resp = append(resp, _type.GetDataSourceTypeListResponse{
			Type:   v,
			Params: store.Params,
		})
	}
	return map[string]interface{}{
		"list": resp,
	}, nil
}

func NewDataSource(req *_type.NewDataSourceRequest, lang string) (interface{}, error) {
	store, exists := factory.CreateDataSource(req.Type)
	if exists != nil {
		return nil, errors.New("invalid Datasource type")
	}
	var existingRecord model.DataSource
	err := model.DB.Where("name = ?", req.Name).Find(&existingRecord).Error
	if err != nil {
		return nil, errors.New("illegal command")
	}
	if req.Edit == "true" {
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
	if req.Edit == "true" {
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

func DeleteDataSource(req *_type.DeleteDataSourceRequest, lang string) (interface{}, error) {
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
