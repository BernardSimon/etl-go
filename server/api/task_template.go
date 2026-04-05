package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetTaskTemplateList(_ *interface{}, _ string) (interface{}, error) {
	var templates []model.TaskTemplate
	if err := model.DB.Order("created_at desc").Find(&templates).Error; err != nil {
		return nil, errors.New("failed to get task template list")
	}
	return map[string]interface{}{
		"list": templates,
	}, nil
}

func SaveTaskTemplate(req *types.SaveTaskTemplateRequest, lang string) (interface{}, error) {
	var template model.TaskTemplate
	if req.ID != "" {
		if err := model.DB.Where("id = ?", req.ID).First(&template).Error; err != nil {
			return nil, errors.New("task template not found")
		}
	}

	template.Name = req.Name
	template.TaskType = req.TaskType
	template.Cron = req.Cron
	template.Data = &req.ParStr

	if err := model.DB.Save(&template).Error; err != nil {
		return nil, errors.New("failed to save task template")
	}
	return i18n.Translate(lang, "success"), nil
}

func DeleteTaskTemplate(req *types.DeleteTaskTemplateRequest, lang string) (interface{}, error) {
	var template model.TaskTemplate
	if err := model.DB.Where("id = ?", req.ID).First(&template).Error; err != nil {
		return nil, errors.New("task template not found")
	}
	if err := model.DB.Delete(&template).Error; err != nil {
		return nil, errors.New("failed to delete task template")
	}
	return i18n.Translate(lang, "success"), nil
}
