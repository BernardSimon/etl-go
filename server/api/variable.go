package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/task"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
	"gorm.io/gorm"
)

func GetVariableList(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	var variableList []model.Variable
	model.DB.Order("created_at desc").Preload("DataSource", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	}).Find(&variableList)
	return map[string]interface{}{
		"list": variableList,
	}, nil
}

func NewVariable(_ *struct{}, body *types.NewVariableRequest, lang string) (interface{}, error) {
	var existingVariable model.Variable
	query := model.DB.Where("name = ?", body.Name)
	if body.Edit {
		query = query.Where("id != ?", body.ID)
	}
	if err := query.Limit(1).Find(&existingVariable).Error; err != nil {
		return nil, errors.New("system error")
	}
	if existingVariable.ID != "" {
		return nil, errors.New("variable name already exists")
	}

	var variable model.Variable
	if body.Edit {
		if err := model.DB.Where("id = ?", body.ID).First(&variable).Error; err != nil {
			return nil, errors.New("variable not exists")
		}
	}

	store, err := factory.CreateVariable(body.Type)
	if err != nil {
		return nil, errors.New("invalid variable type")
	}

	for _, p := range store.Params {
		matched := false
		for _, param := range body.Value {
			if param.Key == p.Key {
				if p.Required && param.Value == "" {
					return nil, errors.New("variable value is not complete")
				}
				matched = true
				break
			}
		}
		if !matched {
			return nil, errors.New("variable value is not complete")
		}
	}

	if store.Datasource != nil {
		if body.DataSourceID == nil {
			return nil, errors.New("variable type need datasource")
		}
		var dataSourceInstance model.DataSource
		if err := model.DB.Where("id = ?", body.DataSourceID).First(&dataSourceInstance).Error; err != nil {
			return nil, errors.New("datasource does not exist")
		}
		dsType := *store.Datasource
		if dataSourceInstance.Type != dsType {
			return nil, errors.New("datasource type is not match")
		}
	} else if body.DataSourceID != nil {
		return nil, errors.New("variable type do not support datasource")
	}

	variable.Name = body.Name
	variable.Type = body.Type
	variable.Description = body.Description
	variable.Value = &body.Value
	variable.DataSourceID = body.DataSourceID

	if err := model.DB.Save(&variable).Error; err != nil {
		return nil, errors.New("failed to save variable")
	}
	return i18n.Translate(lang, "success"), nil
}

func DeleteVariable(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var variable model.Variable
	model.DB.Where("id = ?", uri.Id).First(&variable)
	if variable.ID == "" {
		return nil, errors.New("variable not exists")
	}
	if err := model.DB.Delete(&variable).Error; err != nil {
		return nil, errors.New("failed to delete variable")
	}
	return i18n.Translate(lang, "success"), nil
}

func TestVariable(uri *types.IDUri, _ *struct{}, _ string) (interface{}, error) {
	var variable model.Variable
	err := model.DB.Where("id = ?", uri.Id).Limit(1).Preload("DataSource").First(&variable).Error
	if err != nil {
		return nil, errors.New("variable not exists")
	}
	return task.GetValueByName(variable.Name)
}

func GetVariableTypeList(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	list := factory.GetVariableTypeList()
	var resp = make([]types.GetVariableTypeListResponse, 0)
	for _, v := range list {
		store, _ := factory.CreateVariable(v)
		r := types.GetVariableTypeListResponse{
			Type:   v,
			Params: normalizeParams(store.Params),
		}
		if store.Datasource != nil {
			name := *store.Datasource
			var dataSourceList []model.DataSource
			model.DB.Where("type = ?", name).Find(&dataSourceList)
			dsL := make([]struct {
				Name string
				ID   string
			}, 0)
			for _, ds := range dataSourceList {
				dsL = append(dsL, struct {
					Name string
					ID   string
				}{
					Name: ds.Name,
					ID:   ds.ID,
				})
			}
			r.DatasourceList = &dsL
		}
		resp = append(resp, r)
	}
	return map[string]interface{}{
		"list": resp,
	}, nil
}
