package api

import (
	"errors"

	"github.com/BernardSimon/etl-go/server/model"
	_type "github.com/BernardSimon/etl-go/server/type"
	"github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
)

func GetFileList(req *_type.GetFileListRequest, _ string) (interface{}, error) {
	var fileList = make([]model.File, 0)
	var total int64
	if err := model.DB.Model(&model.File{}).Order("created_at desc").Count(&total).Limit(req.PageSize).Offset((req.PageNo - 1) * req.PageSize).Find(&fileList).Error; err != nil {
		return nil, errors.New("failed to get file list")
	}
	return map[string]interface{}{
		"total": total,
		"list":  fileList,
	}, nil
}

func UploadFile(req *_type.UploadFileRequest, _ string) (interface{}, error) {
	f, err := file.SaveFileInput(&req.File)
	if err != nil {
		return nil, errors.New("failed to upload file")
	}
	return f, nil
}

func DeleteFile(req *_type.DeleteFileRequest, lang string) (interface{}, error) {
	err := file.DeleteFile(req.ID)
	if err != nil {
		return nil, errors.New("failed to delete file")
	}
	return i18n.Translate(lang, "success"), nil
}
