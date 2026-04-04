package _type

import "mime/multipart"

type UploadFileRequest struct {
	File multipart.FileHeader `form:"file"`
}

type GetFileListRequest struct {
	PageSize int `json:"page_size" form:"page_size"`
	PageNo   int `json:"page_no" form:"page_no"`
}

type DeleteFileRequest struct {
	ID string `json:"id" uri:"id" binding:"required"`
}
