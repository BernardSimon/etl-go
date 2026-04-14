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

// ── Chunked upload types ──────────────────────────────────────────────────────

type InitUploadSessionRequest struct {
	Filename     string `json:"filename"     binding:"required"`
	TotalSize    int64  `json:"total_size"   binding:"required,min=1"`
	ChunkSize    int64  `json:"chunk_size"   binding:"required,min=65536"`
	TotalChunks  int    `json:"total_chunks" binding:"required,min=1"`
	ExpectedHash string `json:"expected_hash"`
}

// SessionIDUri is used for routes that only have a :session_id path param.
type SessionIDUri struct {
	SessionID string `uri:"session_id"`
}
