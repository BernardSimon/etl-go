package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

// GetTagList 获取所有标签
func GetTagList(_ *struct{}, query *types.GetTagListRequest, _ string) (interface{}, error) {
	var tags []model.Tag
	tx := model.DB.Model(&model.Tag{})
	if query.Search != "" {
		tx = tx.Where("name LIKE ?", "%"+query.Search+"%")
	}
	if err := tx.Order("created_at desc").Find(&tags).Error; err != nil {
		return nil, errors.New("failed to get tag list")
	}
	return tags, nil
}

// AddTag 新增标签
func AddTag(_ *struct{}, body *types.AddTagRequest, lang string) (interface{}, error) {
	// 检查名称是否重复
	var count int64
	model.DB.Model(&model.Tag{}).Where("name = ?", body.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("tag name already exists")
	}

	tag := model.Tag{
		Name:  body.Name,
		Color: body.Color,
	}
	if err := model.DB.Create(&tag).Error; err != nil {
		return nil, errors.New("failed to create tag")
	}
	return tag, nil
}

// UpdateTag 更新标签
func UpdateTag(uri *types.IDUri, body *types.UpdateTagBody, lang string) (interface{}, error) {
	var tag model.Tag
	if err := model.DB.Where("id = ?", uri.Id).First(&tag).Error; err != nil {
		return nil, errors.New("tag not found")
	}

	// 检查名称是否与其他标签重复
	var count int64
	model.DB.Model(&model.Tag{}).Where("name = ? AND id <> ?", body.Name, uri.Id).Count(&count)
	if count > 0 {
		return nil, errors.New("tag name already exists")
	}

	tag.Name = body.Name
	tag.Color = body.Color
	if err := model.DB.Save(&tag).Error; err != nil {
		return nil, errors.New("failed to update tag")
	}
	return i18n.Translate(lang, "success"), nil
}

// DeleteTag 删除标签
func DeleteTag(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	var tag model.Tag
	if err := model.DB.Where("id = ?", uri.Id).First(&tag).Error; err != nil {
		return nil, errors.New("tag not found")
	}

	// 同时删除关联关系
	model.DB.Where("tag_id = ?", uri.Id).Delete(&model.TaskTag{})

	if err := model.DB.Delete(&tag).Error; err != nil {
		return nil, errors.New("failed to delete tag")
	}
	return i18n.Translate(lang, "success"), nil
}
