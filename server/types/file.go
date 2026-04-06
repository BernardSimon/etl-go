package types

import "mime/multipart"

type UploadFileRequest struct {
	File multipart.FileHeader `form:"file"`
}

type GetFileListRequest struct {
	PageSize int    `form:"page_size"`
	PageNo   int    `form:"page_no"`
	Keyword  string `form:"keyword"`
	IDs      string `form:"ids"`
}
