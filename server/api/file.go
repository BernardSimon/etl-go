package api

import (
	"errors"
	"strings"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetFileList(_ *struct{}, query *types.GetFileListRequest, _ string) (interface{}, error) {
	var fileList = make([]model.File, 0)
	var total int64
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	pageNo := query.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}

	q := model.DB.Model(&model.File{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if ids := strings.TrimSpace(query.IDs); ids != "" {
		parts := strings.Split(ids, ",")
		cleaned := make([]string, 0, len(parts))
		for _, id := range parts {
			id = strings.TrimSpace(id)
			if id != "" {
				cleaned = append(cleaned, id)
			}
		}
		if len(cleaned) > 0 {
			q = q.Where("id IN ?", cleaned)
		}
	}

	if err := q.Count(&total).Order("created_at desc").Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&fileList).Error; err != nil {
		return nil, errors.New("failed to get file list")
	}
	return map[string]interface{}{
		"total":     total,
		"page_no":   pageNo,
		"page_size": pageSize,
		"list":      fileList,
	}, nil
}

func UploadFile(_ *struct{}, body *types.UploadFileRequest, _ string) (interface{}, error) {
	f, err := file.SaveFileInput(&body.File)
	if err != nil {
		return nil, errors.New("failed to upload file")
	}
	return f, nil
}

func DeleteFile(uri *types.IDUri, _ *struct{}, lang string) (interface{}, error) {
	err := file.DeleteFile(uri.Id)
	if err != nil {
		return nil, errors.New("failed to delete file")
	}
	return i18n.Translate(lang, "success"), nil
}
