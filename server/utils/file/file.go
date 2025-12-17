package file

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/BernardSimon/etl-go/server/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func saveFile(path string, file *multipart.FileHeader) (*model.File, error) {
	// 确保目录存在
	dir := filepath.Dir("./file/" + path + "/")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}
	id := uuid.New().String()
	fileExName := filepath.Ext(file.Filename)

	// 创建目标文件
	dst, err := os.Create("./file/" + path + "/" + id + fileExName)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = dst.Close()
	}()

	// 打开上传的文件
	content, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func(content multipart.File) {
		_ = content.Close()
	}(content)

	// 将上传文件内容复制到目标文件
	_, err = io.Copy(dst, content)
	if err != nil {
		return nil, err
	}
	record := &model.File{
		Model: model.Model{
			ID: id,
		},
		Name:   file.Filename,
		Path:   path,
		Size:   file.Size,
		ExName: fileExName,
	}
	if err := model.DB.Create(&record).Error; err != nil {
		zap.L().Error("Failed To Create File Record", zap.Error(err), zap.String("name", "file"), zap.Any("content", record))
		//删除本地文件
		_ = os.Remove("./file/" + path + "/" + id + fileExName)
		return nil, err
	}

	return record, nil
}

func SaveFileInput(file *multipart.FileHeader) (*model.File, error) {
	return saveFile("input", file)
}

func DeleteFile(id string) error {
	var file model.File
	if err := model.DB.Where("id = ?", id).First(&file).Error; err != nil {
		return errors.New("file record does not exist")
	}

	// 检查文件是否存在
	filePath := "./file/" + file.Path + "/" + file.ID + file.ExName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，直接删除记录
		if err = model.DB.Delete(&file).Error; err != nil {
			return errors.New("failed to delete file record")
		}
		return nil
	}

	// 文件存在，删除本地文件
	err := os.Remove(filePath)
	if err != nil {
		return errors.New("failed to delete file")
	}

	// 删除数据库记录
	tx := model.DB.Begin()
	if err = tx.Where("file_id = ?", id).Delete(&model.TaskRecordFile{}).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to delete task record file")
	}
	if err = tx.Delete(&file).Error; err != nil {
		tx.Rollback()
		return errors.New("failed to delete file record")
	}
	tx.Commit()
	return nil
}

func GetFilePath(id string) (string, error) {
	var files model.File
	if err := model.DB.Where("id = ?", id).First(&files).Error; err != nil {
		return "", errors.New("file record does not exist")
	}
	absPath, err := filepath.Abs("./file/" + files.Path + "/" + files.ID + files.ExName)
	if err != nil {
		return "", errors.New("failed to get file path")
	}
	return absPath, nil
}

func SetOutputFile(fileName string, exName string) model.File {
	var file = model.File{
		Model: model.Model{
			ID: uuid.New().String(),
		},
		Name:   fileName,
		Path:   "output",
		Size:   0,
		ExName: exName,
	}
	return file
}

func SaveOutputFileRecord(file model.File) {
	if err := model.DB.Create(&file).Error; err != nil {
		zap.L().Error("Failed To Create File Record", zap.Error(err), zap.String("name", "file"), zap.Any("content", file))
	}
}

func SaveOutputTaskRecord(taskRecordId string, fileId string) {
	if err := model.DB.Create(&model.TaskRecordFile{
		TaskRecordID: taskRecordId,
		FileID:       fileId,
	}).Error; err != nil {
		zap.L().Error("Failed To Create TaskRecordFile Record", zap.Error(err), zap.String("name", "file"), zap.Any("content", model.TaskRecordFile{
			TaskRecordID: taskRecordId,
			FileID:       fileId,
		}))
	}
}

func CreateOutputFile(fileName string, exName string) (string, string, error) {
	if !strings.HasPrefix(exName, ".") {
		exName = "." + exName
	}
	id := uuid.New().String()
	fullName := id + exName
	if strings.HasSuffix(fileName, exName) {
		fileName = strings.TrimSuffix(fileName, exName)
	}
	var file = model.File{
		Model: model.Model{
			ID: id,
		},
		Name:   fileName,
		Path:   "output",
		Size:   0,
		ExName: exName,
	}
	err := model.DB.Create(&file).Error
	if err != nil {
		return "", "", errors.New("failed to create file record")
	}
	absPath, err := filepath.Abs("./file/output/" + fullName)
	if err != nil {
		return "", "", errors.New("failed to get file path")
	}
	return id, absPath, nil
}

func SaveOutputFile(recordID string, ids []string, isError bool) error {
	var ers []error
	for _, id := range ids {
		if isError {
			if err := DeleteFile(id); err != nil {
				ers = append(ers, err)
				continue
			}
		}
		var file model.File
		if err := model.DB.Where("id = ?", id).First(&file).Error; err != nil {
			ers = append(ers, errors.New("file record does not exist"))
			continue
		}
		if fileInfo, err := os.Stat("./file/" + file.Path + "/" + file.ID + file.ExName); os.IsNotExist(err) {
			// 文件不存在，直接删除记录
			if err = model.DB.Delete(&file).Error; err != nil {
				ers = append(ers, errors.New("failed to delete file record"))
				continue
			}
			continue
		} else {
			file.Size = fileInfo.Size()
		}
		if err := model.DB.Save(&file).Error; err != nil {
			ers = append(ers, errors.New("failed to save file record"))
			continue
		}
		err := model.DB.Create(&model.TaskRecordFile{TaskRecordID: recordID, FileID: id}).Error
		if err != nil {
			ers = append(ers, errors.New("failed to create task record file"))
			continue
		}

	}
	if len(ers) > 0 {
		return errors.Join(ers...)
	}
	return nil
}
