package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetTaskTemplateList(_ *struct{}, _ *struct{}, _ string) (interface{}, error) {
	var templates []model.TaskTemplate
	if err := model.DB.Order("created_at desc").Find(&templates).Error; err != nil {
		return nil, errors.New("failed to get task template list")
	}
	return map[string]interface{}{
		"list": templates,
	}, nil
}

func SaveTaskTemplate(_ *struct{}, body *types.SaveTaskTemplateRequest, lang string) (interface{}, error) {
	var template model.TaskTemplate
	if body.ID != "" {
		if err := model.DB.Where("id = ?", body.ID).First(&template).Error; err != nil {
			return nil, errors.New("task template not found")
		}
	}
	template.Name = body.Name
	template.TaskType = body.TaskType
	template.Cron = body.Cron
	template.Data = &body.ParStr
	if err := model.DB.Save(&template).Error; err != nil {
		return nil, errors.New("failed to save task template")
	}
	return i18n.Translate(lang, "success"), nil
}

func DeleteTaskTemplate(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var template model.TaskTemplate
	if err := model.DB.Where("id = ?", uri.Id).First(&template).Error; err != nil {
		return nil, errors.New("task template not found")
	}
	if err := model.DB.Delete(&template).Error; err != nil {
		return nil, errors.New("failed to delete task template")
	}
	return i18n.Translate(lang, "success"), nil
}
