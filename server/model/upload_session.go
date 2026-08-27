package model

// UploadSession tracks one client-initiated chunked upload.
// Status state machine:
//
//	pending → uploading → assembling → done
//	                              ↘ error
//	↘ cancelled (from any state before done)
type UploadSession struct {
	Model
	Filename       string `json:"filename"`
	ExName         string `json:"ex_name"`
	TotalSize      int64  `json:"total_size"`
	ChunkSize      int64  `json:"chunk_size"`
	TotalChunks    int    `json:"total_chunks"`
	ReceivedChunks int    `json:"received_chunks"`
	// ReceivedMap is a JSON array of received chunk indices, e.g. "[0,1,3]"
	ReceivedMap  string `json:"received_map" gorm:"type:text"`
	Status       string `json:"status"`
	FileID       string `json:"file_id"`
	TempDir      string `json:"temp_dir"`
	ExpectedHash string `json:"expected_hash"`
}
