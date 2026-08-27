package api

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BernardSimon/etl-go/server/model"
	types "github.com/BernardSimon/etl-go/server/types"
	"github.com/BernardSimon/etl-go/server/utils/file"
	"github.com/BernardSimon/etl-go/server/utils/i18n"
	"github.com/gin-gonic/gin"
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

// ── Chunked upload handlers ───────────────────────────────────────────────────

func InitUploadSession(_ *struct{}, body *types.InitUploadSessionRequest, _ string) (interface{}, error) {
	exName := filepath.Ext(body.Filename)
	session, err := file.InitSession(body.Filename, exName, body.TotalSize, body.ChunkSize, body.TotalChunks, body.ExpectedHash)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"session_id":   session.ID,
		"chunk_size":   session.ChunkSize,
		"total_chunks": session.TotalChunks,
	}, nil
}

func GetUploadStatus(uri *types.SessionIDUri, _ *struct{}, _ string) (interface{}, error) {
	session, err := file.GetSession(uri.SessionID)
	if err != nil {
		return nil, err
	}
	var receivedMap []int
	_ = json.Unmarshal([]byte(session.ReceivedMap), &receivedMap)
	if receivedMap == nil {
		receivedMap = []int{}
	}
	return map[string]interface{}{
		"session_id":      session.ID,
		"status":          session.Status,
		"total_chunks":    session.TotalChunks,
		"received_chunks": session.ReceivedChunks,
		"received_map":    receivedMap,
		"file_id":         session.FileID,
	}, nil
}

func CompleteUpload(uri *types.SessionIDUri, _ *struct{}, _ string) (interface{}, error) {
	record, err := file.AssembleFile(uri.SessionID)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func CancelUpload(uri *types.SessionIDUri, _ *struct{}, lang string) (interface{}, error) {
	if err := file.CancelSession(uri.SessionID); err != nil {
		return nil, err
	}
	return i18n.Translate(lang, "success"), nil
}

// UploadChunkRaw is a plain gin.HandlerFunc that streams the binary request
// body directly to disk. It intentionally bypasses RequestResponseMiddleware
// (which would pre-read the whole body into memory) and handles auth itself.
func UploadChunkRaw(c *gin.Context) {
	// Manual auth — RequestResponseMiddleware has NOT run for this route
	token := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if _, err := DecodeToken(token); err != nil {
		c.JSON(401, types.ResponseModel{Code: 3, Message: "invalid token"})
		return
	}

	sessionID := c.Param("session_id")
	chunkIndex, err := strconv.Atoi(c.Param("chunk_index"))
	if err != nil || chunkIndex < 0 {
		c.JSON(400, types.ResponseModel{Code: 1, Message: "invalid chunk_index"})
		return
	}

	chunkHash := c.GetHeader("X-Chunk-SHA256")

	result, err := file.SaveChunk(sessionID, chunkIndex, c.Request.Body, chunkHash)
	if err != nil {
		c.JSON(422, types.ResponseWithData{Code: 2, Message: err.Error()})
		return
	}
	c.JSON(200, types.ResponseWithData{Code: 0, Message: "ok", Data: result})
}
