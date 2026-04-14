package file

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/BernardSimon/etl-go/server/model"
	"gorm.io/gorm"
)

const (
	SessionStatusPending    = "pending"
	SessionStatusUploading  = "uploading"
	SessionStatusAssembling = "assembling"
	SessionStatusDone       = "done"
	SessionStatusError      = "error"
	SessionStatusCancelled  = "cancelled"
)

// ChunkUploadResult is returned after a chunk is saved successfully.
type ChunkUploadResult struct {
	ReceivedChunks int `json:"received_chunks"`
	TotalChunks    int `json:"total_chunks"`
}

// InitSession creates a new upload session and its temporary directory.
func InitSession(filename, exName string, totalSize, chunkSize int64, totalChunks int, expectedHash string) (*model.UploadSession, error) {
	if totalChunks < 1 {
		return nil, errors.New("total_chunks must be >= 1")
	}
	if chunkSize < 65536 {
		return nil, errors.New("chunk_size must be >= 65536 bytes")
	}
	if totalSize < 1 {
		return nil, errors.New("total_size must be > 0")
	}

	id := model.GenerateID()
	tempDir := filepath.Join("./file/chunks", id)
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	session := &model.UploadSession{
		Model:          model.Model{ID: id},
		Filename:       filename,
		ExName:         exName,
		TotalSize:      totalSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		ReceivedChunks: 0,
		ReceivedMap:    "[]",
		Status:         SessionStatusPending,
		TempDir:        tempDir,
		ExpectedHash:   expectedHash,
	}
	if err := model.DB.Create(session).Error; err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create upload session: %w", err)
	}
	return session, nil
}

// GetSession fetches an upload session by ID.
func GetSession(sessionID string) (*model.UploadSession, error) {
	var session model.UploadSession
	if err := model.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("upload session not found")
		}
		return nil, fmt.Errorf("failed to get upload session: %w", err)
	}
	return &session, nil
}

// SaveChunk streams chunk data from r to disk and updates the session.
// expectedHash is an optional hex-encoded SHA-256 of the chunk; pass empty string to skip.
func SaveChunk(sessionID string, chunkIndex int, r io.Reader, expectedHash string) (*ChunkUploadResult, error) {
	session, err := GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	switch session.Status {
	case SessionStatusDone, SessionStatusCancelled:
		return nil, fmt.Errorf("session is in terminal state: %s", session.Status)
	case SessionStatusAssembling:
		return nil, errors.New("session is currently assembling")
	}

	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return nil, fmt.Errorf("chunk_index %d out of range [0, %d)", chunkIndex, session.TotalChunks)
	}

	received, err := parseReceivedMap(session.ReceivedMap)
	if err != nil {
		return nil, fmt.Errorf("invalid received_map: %w", err)
	}
	receivedSet := toSet(received)

	// Chunk already received — idempotent, return current state
	if receivedSet[chunkIndex] {
		return &ChunkUploadResult{
			ReceivedChunks: session.ReceivedChunks,
			TotalChunks:    session.TotalChunks,
		}, nil
	}

	// Stream chunk to disk
	chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("%d.part", chunkIndex))
	dst, err := os.Create(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk file: %w", err)
	}

	h := sha256.New()
	tee := io.TeeReader(r, h)
	if _, err := io.Copy(dst, tee); err != nil {
		dst.Close()
		_ = os.Remove(chunkPath)
		return nil, fmt.Errorf("failed to write chunk data: %w", err)
	}
	dst.Close()

	if expectedHash != "" {
		if actualHash := hex.EncodeToString(h.Sum(nil)); actualHash != expectedHash {
			_ = os.Remove(chunkPath)
			return nil, errors.New("chunk hash mismatch")
		}
	}

	// Update session atomically
	received = append(received, chunkIndex)
	newMap, _ := json.Marshal(received)
	newCount := session.ReceivedChunks + 1

	if err := model.DB.Model(session).Updates(map[string]interface{}{
		"received_map":    string(newMap),
		"received_chunks": newCount,
		"status":          SessionStatusUploading,
	}).Error; err != nil {
		_ = os.Remove(chunkPath)
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return &ChunkUploadResult{
		ReceivedChunks: newCount,
		TotalChunks:    session.TotalChunks,
	}, nil
}

// AssembleFile concatenates all chunks into the final file and creates a model.File record.
func AssembleFile(sessionID string) (*model.File, error) {
	session, err := GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// If already done, return the existing file record
	if session.Status == SessionStatusDone {
		var f model.File
		if err := model.DB.Where("id = ?", session.FileID).First(&f).Error; err == nil {
			return &f, nil
		}
	}

	if session.Status != SessionStatusUploading && session.Status != SessionStatusPending {
		return nil, fmt.Errorf("session is not in uploadable state: %s", session.Status)
	}
	if session.ReceivedChunks < session.TotalChunks {
		return nil, fmt.Errorf("not all chunks received: %d/%d", session.ReceivedChunks, session.TotalChunks)
	}

	// Mark as assembling to prevent duplicate triggers
	if err := model.DB.Model(session).Update("status", SessionStatusAssembling).Error; err != nil {
		return nil, fmt.Errorf("failed to set assembling status: %w", err)
	}

	setErr := func(e error) error {
		model.DB.Model(session).Update("status", SessionStatusError)
		return e
	}

	// Prepare destination
	if err := os.MkdirAll("./file/input", os.ModePerm); err != nil {
		return nil, setErr(fmt.Errorf("failed to create input directory: %w", err))
	}

	fileID := model.GenerateID()
	finalPath := filepath.Join("./file/input", fileID+session.ExName)

	finalFile, err := os.Create(finalPath)
	if err != nil {
		return nil, setErr(fmt.Errorf("failed to create output file: %w", err))
	}

	// Wire up optional whole-file hash check
	var fileHasher hash.Hash
	var writer io.Writer = finalFile
	if session.ExpectedHash != "" {
		fileHasher = sha256.New()
		writer = io.MultiWriter(finalFile, fileHasher)
	}

	// Concatenate chunks
	for i := 0; i < session.TotalChunks; i++ {
		chunkPath := filepath.Join(session.TempDir, fmt.Sprintf("%d.part", i))
		chunk, err := os.Open(chunkPath)
		if err != nil {
			finalFile.Close()
			_ = os.Remove(finalPath)
			return nil, setErr(fmt.Errorf("failed to open chunk %d: %w", i, err))
		}
		_, copyErr := io.Copy(writer, chunk)
		chunk.Close()
		if copyErr != nil {
			finalFile.Close()
			_ = os.Remove(finalPath)
			return nil, setErr(fmt.Errorf("failed to assemble chunk %d: %w", i, copyErr))
		}
	}
	finalFile.Close()

	// Verify whole-file hash
	if session.ExpectedHash != "" && fileHasher != nil {
		if actualHash := hex.EncodeToString(fileHasher.Sum(nil)); actualHash != session.ExpectedHash {
			_ = os.Remove(finalPath)
			return nil, setErr(errors.New("assembled file hash mismatch"))
		}
	}

	// Get real file size
	info, err := os.Stat(finalPath)
	if err != nil {
		_ = os.Remove(finalPath)
		return nil, setErr(fmt.Errorf("failed to stat assembled file: %w", err))
	}

	// Create file record
	record := &model.File{
		Model:  model.Model{ID: fileID},
		Name:   session.Filename,
		Path:   "input",
		Size:   info.Size(),
		ExName: session.ExName,
	}
	if err := model.DB.Create(record).Error; err != nil {
		_ = os.Remove(finalPath)
		return nil, setErr(fmt.Errorf("failed to create file record: %w", err))
	}

	// Mark session done
	model.DB.Model(session).Updates(map[string]interface{}{
		"status":  SessionStatusDone,
		"file_id": fileID,
	})

	// Async cleanup of temp chunk directory
	go func() { _ = os.RemoveAll(session.TempDir) }()

	return record, nil
}

// CancelSession cancels an upload session and removes its temporary directory.
func CancelSession(sessionID string) error {
	session, err := GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.Status == SessionStatusDone {
		return errors.New("cannot cancel a completed session")
	}
	if err := model.DB.Model(session).Update("status", SessionStatusCancelled).Error; err != nil {
		return fmt.Errorf("failed to cancel session: %w", err)
	}
	go func() { _ = os.RemoveAll(session.TempDir) }()
	return nil
}

// DeleteStaleSessions cancels upload sessions older than the given duration
// that are still in pending, uploading, or error state.
func DeleteStaleSessions(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	var sessions []model.UploadSession
	if err := model.DB.Where(
		"status IN ? AND updated_at < ?",
		[]string{SessionStatusPending, SessionStatusUploading, SessionStatusError},
		cutoff,
	).Find(&sessions).Error; err != nil {
		return err
	}
	for _, s := range sessions {
		_ = CancelSession(s.ID)
	}
	return nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func parseReceivedMap(jsonStr string) ([]int, error) {
	if jsonStr == "" {
		return []int{}, nil
	}
	var result []int
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return result, nil
}

func toSet(indices []int) map[int]bool {
	set := make(map[int]bool, len(indices))
	for _, idx := range indices {
		set[idx] = true
	}
	return set
}
