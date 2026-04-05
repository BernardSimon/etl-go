package api

import (
	"errors"
	"strings"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetFileList(req *types.GetFileListRequest, _ string) (interface{}, error) {
	var fileList = make([]model.File, 0)
	var total int64
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}

	query := model.DB.Model(&model.File{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}
	if ids := strings.TrimSpace(req.IDs); ids != "" {
		parts := strings.Split(ids, ",")
		cleaned := make([]string, 0, len(parts))
		for _, id := range parts {
			id = strings.TrimSpace(id)
			if id != "" {
				cleaned = append(cleaned, id)
			}
		}
		if len(cleaned) > 0 {
			query = query.Where("id IN ?", cleaned)
		}
	}

	if err := query.Count(&total).Order("created_at desc").Limit(pageSize).Offset((pageNo - 1) * pageSize).Find(&fileList).Error; err != nil {
		return nil, errors.New("failed to get file list")
	}
	return map[string]interface{}{
		"total":     total,
		"page_no":   pageNo,
		"page_size": pageSize,
		"list":      fileList,
	}, nil
}

func UploadFile(req *types.UploadFileRequest, _ string) (interface{}, error) {
	f, err := file.SaveFileInput(&req.File)
	if err != nil {
		return nil, errors.New("failed to upload file")
	}
	return f, nil
}

func DeleteFile(req *types.DeleteFileRequest, lang string) (interface{}, error) {
	err := file.DeleteFile(req.ID)
	if err != nil {
		return nil, errors.New("failed to delete file")
	}
	return i18n.Translate(lang, "success"), nil
}
