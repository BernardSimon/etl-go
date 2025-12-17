package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/etl/factory"
	"github.com/BernardSimon/etl-go/server/model"
	"github.com/BernardSimon/etl-go/server/task"
	_type "github.com/BernardSimon/etl-go/server/type"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
	"gorm.io/gorm"
)

func GetVariableList(_ *interface{}, _ string) (interface{}, error) {
	var variableList []model.Variable
	model.DB.Order("created_at desc").Preload("DataSource", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name")
	}).Find(&variableList)
	return map[string]interface{}{
		"list": variableList,
	}, nil
}

func NewVariable(req *_type.NewVariableRequest, lang string) (interface{}, error) {
	// 检查变量名是否已存在（排除当前编辑的变量）
	var existingVariable model.Variable
	query := model.DB.Where("name = ?", req.Name)

	// 如果是编辑模式，排除当前变量ID
	if req.Edit == "true" {
		query = query.Where("id != ?", req.ID)
	}

	if err := query.Limit(1).Find(&existingVariable).Error; err != nil {
		return nil, errors.New("system error")
	}

	if existingVariable.ID != "" {
		return nil, errors.New("variable name already exists")
	}

	// 处理变量创建或更新
	var variable model.Variable
	if req.Edit == "true" {
		// 编辑模式：检查变量是否存在
		if err := model.DB.Where("id = ?", req.ID).First(&variable).Error; err != nil {
			return nil, errors.New("variable not exists")
		}
	}

	store, err := factory.CreateVariable(req.Type)
	if err != nil {
		return nil, errors.New("invalid variable type")
	}

	// 验证参数完整性
	for _, p := range store.Params {
		matched := false
		for _, param := range req.Value {
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

	// 验证数据源
	if store.Datasource != nil {
		if req.DataSourceID == nil {
			return nil, errors.New("variable type need datasource")
		}

		var dataSourceInstance model.DataSource
		if err := model.DB.Where("id = ?", req.DataSourceID).First(&dataSourceInstance).Error; err != nil {
			return nil, errors.New("datasource does not exist")
		}

		dsType := *store.Datasource
		if dataSourceInstance.Type != dsType {
			return nil, errors.New("datasource type is not match")
		}
	} else if req.DataSourceID != nil {
		return nil, errors.New("variable type do not support datasource")
	}

	// 更新变量属性
	variable.Name = req.Name
	variable.Type = req.Type
	variable.Description = req.Description
	variable.Value = &req.Value
	variable.DataSourceID = req.DataSourceID

	// 保存到数据库
	if err := model.DB.Save(&variable).Error; err != nil {
		return nil, errors.New("failed to save variable")
	}

	return i18n.Translate(lang, "success"), nil
}

func DeleteVariable(req *_type.DeleteVariableRequest, lang string) (interface{}, error) {
	var variable model.Variable
	model.DB.Where("id = ?", req.Id).First(&variable)
	if variable.ID == "" {
		return nil, errors.New("variable not exists")
	}
	if err := model.DB.Delete(&variable).Error; err != nil {
		return nil, errors.New("failed to delete variable")
	}
	return i18n.Translate(lang, "success"), nil
}

func TestVariable(req *_type.TestVariableRequest, _ string) (interface{}, error) {
	var variable model.Variable
	err := model.DB.Where("id = ?", req.Id).Limit(1).Preload("DataSource").First(&variable).Error
	if err != nil {
		return nil, errors.New("variable not exists")
	}
	return task.GetValueByName(variable.Name)
}

//	func GetVariableTypeList(_ *interface{}, _ string) (interface{}, error) {
//		return factory.GetVariableTypeList(), nil
//	}
func GetVariableTypeList(_ *interface{}, _ string) (interface{}, error) {
	list := factory.GetVariableTypeList()
	var resp = make([]_type.GetVariableTypeListResponse, 0)
	for _, v := range list {
		store, _ := factory.CreateVariable(v)
		r := _type.GetVariableTypeListResponse{
			Type:   v,
			Params: store.Params,
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
