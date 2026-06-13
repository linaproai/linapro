// This file implements dynamic-plugin storage upload sessions for chunked
// transfer. Sessions are runtime-local, backed by temporary files, and commit
// through storagecap.Service so provider and tenant scoping stay centralized.

package wasm

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/storagecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	// storageUploadMaxChunkBytes bounds one WASM host-call storage chunk payload.
	storageUploadMaxChunkBytes = 4 * 1024 * 1024
	// storageUploadSessionTTL bounds orphaned upload session lifetime.
	storageUploadSessionTTL = 15 * time.Minute
)

type storageUploadSessions struct {
	mu       sync.Mutex
	sessions map[string]*storageUploadSession
}

type storageUploadSession struct {
	pluginID    string
	path        string
	contentType string
	overwrite   bool
	tempDir     string
	tempPath    string
	offset      int64
	expiresAt   time.Time
}

func newStorageUploadSessions() *storageUploadSessions {
	return &storageUploadSessions{sessions: make(map[string]*storageUploadSession)}
}

// handleStoragePutInit starts one chunked storage upload session.
func handleStoragePutInit(
	hcc *hostCallContext,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutInitRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	uploads, err := storageUploadsForHostCall(hcc)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	uploadID, err := uploads.init(hcc.pluginID, objectPath, request.ContentType, request.Overwrite)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStoragePutInitResponse(&bridgehostservice.HostServiceStoragePutInitResponse{
			UploadID: uploadID,
		}),
	)
}

// handleStoragePutChunk appends one ordered storage upload chunk.
func handleStoragePutChunk(
	hcc *hostCallContext,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutChunkRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	uploads, err := storageUploadsForHostCall(hcc)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	nextOffset, err := uploads.chunk(hcc.pluginID, objectPath, request.UploadID, request.Offset, request.Body)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStoragePutChunkResponse(&bridgehostservice.HostServiceStoragePutChunkResponse{
			NextOffset: nextOffset,
		}),
	)
}

// handleStoragePutCommit commits one chunked storage upload session through storagecap.
func handleStoragePutCommit(
	ctx context.Context,
	hcc *hostCallContext,
	service storagecap.Service,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutCommitRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	uploads, err := storageUploadsForHostCall(hcc)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	output, err := uploads.commit(ctx, service, hcc.pluginID, objectPath, request.UploadID, request.Size)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(
		bridgehostservice.MarshalHostServiceStoragePutCommitResponse(&bridgehostservice.HostServiceStoragePutCommitResponse{
			Object: storageObjectResponse(outputObject(output)),
		}),
	)
}

// handleStoragePutAbort aborts one chunked storage upload session.
func handleStoragePutAbort(
	hcc *hostCallContext,
	targetPath string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	request, err := bridgehostservice.UnmarshalHostServiceStoragePutAbortRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	objectPath, err := normalizeStorageObjectPath(request.Path)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = validateStorageRequestTarget(targetPath, objectPath); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	uploads, err := storageUploadsForHostCall(hcc)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInternalError, err)
	}
	if err = uploads.abort(hcc.pluginID, objectPath, request.UploadID); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallEmptySuccessResponse()
}

// init creates one runtime-local temporary upload session.
func (s *storageUploadSessions) init(pluginID string, objectPath string, contentType string, overwrite bool) (string, error) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return "", gerror.New("storage upload plugin id is required")
	}
	tempDir, err := os.MkdirTemp("", "linapro-storage-upload-*")
	if err != nil {
		return "", gerror.Wrap(err, "create storage upload temp directory failed")
	}
	tempFile, err := os.CreateTemp(tempDir, "payload-*")
	if err != nil {
		cleanupErr := removeStorageUploadTemp("", tempDir)
		return "", errors.Join(gerror.Wrap(err, "create storage upload temp file failed"), cleanupErr)
	}
	tempPath := tempFile.Name()
	if err = tempFile.Close(); err != nil {
		removeStorageUploadTempBestEffort(tempPath, tempDir)
		return "", gerror.Wrap(err, "close storage upload temp file failed")
	}
	uploadID, err := newStorageUploadID()
	if err != nil {
		removeStorageUploadTempBestEffort(tempPath, tempDir)
		return "", err
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked(now)
	if s.sessions == nil {
		s.sessions = make(map[string]*storageUploadSession)
	}
	if _, exists := s.sessions[uploadID]; exists {
		removeStorageUploadTempBestEffort(tempPath, tempDir)
		return "", gerror.New("storage upload id collision")
	}
	s.sessions[uploadID] = &storageUploadSession{
		pluginID:    pluginID,
		path:        objectPath,
		contentType: contentType,
		overwrite:   overwrite,
		tempDir:     tempDir,
		tempPath:    tempPath,
		expiresAt:   now.Add(storageUploadSessionTTL),
	}
	return uploadID, nil
}

// chunk appends a sequential chunk and returns the next expected offset.
func (s *storageUploadSessions) chunk(pluginID string, objectPath string, uploadID string, offset int64, body []byte) (int64, error) {
	pluginID = strings.TrimSpace(pluginID)
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return 0, gerror.New("storage upload id is required")
	}
	if offset < 0 {
		return 0, gerror.New("storage upload chunk offset cannot be negative")
	}
	if len(body) == 0 {
		return 0, gerror.New("storage upload chunk body is required")
	}
	if len(body) > storageUploadMaxChunkBytes {
		return 0, gerror.Newf("storage upload chunk exceeds maximum size: %d", len(body))
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupExpiredLocked(now)
	session, err := s.sessionLocked(pluginID, objectPath, uploadID)
	if err != nil {
		return 0, err
	}
	if offset != session.offset {
		return 0, gerror.Newf("storage upload chunk offset mismatch: got %d want %d", offset, session.offset)
	}
	if err = appendStorageUploadChunk(session.tempPath, session.offset, body); err != nil {
		return 0, err
	}
	session.offset += int64(len(body))
	session.expiresAt = now.Add(storageUploadSessionTTL)
	return session.offset, nil
}

// commit streams the temporary upload file into the storage domain service.
func (s *storageUploadSessions) commit(
	ctx context.Context,
	service storagecap.Service,
	pluginID string,
	objectPath string,
	uploadID string,
	size int64,
) (*storagecap.PutOutput, error) {
	pluginID = strings.TrimSpace(pluginID)
	uploadID = strings.TrimSpace(uploadID)
	if service == nil {
		return nil, gerror.New("storage upload commit service is required")
	}
	if uploadID == "" {
		return nil, gerror.New("storage upload id is required")
	}
	if size < 0 {
		return nil, gerror.New("storage upload size cannot be negative")
	}

	now := time.Now()
	s.mu.Lock()
	s.cleanupExpiredLocked(now)
	session, err := s.sessionLocked(pluginID, objectPath, uploadID)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}
	if session.offset != size {
		s.mu.Unlock()
		return nil, gerror.Newf("storage upload size mismatch: got %d want %d", size, session.offset)
	}
	delete(s.sessions, uploadID)
	tempDir := session.tempDir
	tempPath := session.tempPath
	contentType := session.contentType
	overwrite := session.overwrite
	s.mu.Unlock()

	file, err := os.Open(tempPath)
	if err != nil {
		cleanupErr := removeStorageUploadTemp(tempPath, tempDir)
		if cleanupErr != nil {
			cleanupErr = gerror.Wrap(cleanupErr, "remove storage upload temp file failed")
		}
		return nil, errors.Join(gerror.Wrap(err, "open storage upload temp file failed"), cleanupErr)
	}
	output, putErr := service.Put(ctx, storagecap.PutInput{
		Path:        objectPath,
		Body:        file,
		Size:        size,
		ContentType: contentType,
		Overwrite:   overwrite,
	})
	closeErr := file.Close()
	cleanupErr := removeStorageUploadTemp(tempPath, tempDir)
	if closeErr != nil {
		closeErr = gerror.Wrap(closeErr, "close storage upload temp file failed")
	}
	if cleanupErr != nil {
		cleanupErr = gerror.Wrap(cleanupErr, "remove storage upload temp file failed")
	}
	if putErr != nil {
		return nil, errors.Join(putErr, closeErr, cleanupErr)
	}
	return output, errors.Join(closeErr, cleanupErr)
}

// abort removes one temporary upload session and its file.
func (s *storageUploadSessions) abort(pluginID string, objectPath string, uploadID string) error {
	pluginID = strings.TrimSpace(pluginID)
	uploadID = strings.TrimSpace(uploadID)
	if uploadID == "" {
		return gerror.New("storage upload id is required")
	}

	now := time.Now()
	s.mu.Lock()
	s.cleanupExpiredLocked(now)
	session, err := s.sessionLocked(pluginID, objectPath, uploadID)
	if err != nil {
		s.mu.Unlock()
		return err
	}
	delete(s.sessions, uploadID)
	tempDir := session.tempDir
	tempPath := session.tempPath
	s.mu.Unlock()
	if err = removeStorageUploadTemp(tempPath, tempDir); err != nil {
		return gerror.Wrap(err, "remove storage upload temp file failed")
	}
	return nil
}

func storageUploadsForHostCall(hcc *hostCallContext) (*storageUploadSessions, error) {
	if hcc == nil || hcc.runtime == nil || hcc.runtime.storageUploads == nil {
		return nil, gerror.New("storage upload runtime is not configured")
	}
	return hcc.runtime.storageUploads, nil
}

func (s *storageUploadSessions) sessionLocked(pluginID string, objectPath string, uploadID string) (*storageUploadSession, error) {
	if s == nil || s.sessions == nil {
		return nil, gerror.New("storage upload session not found")
	}
	session := s.sessions[uploadID]
	if session == nil {
		return nil, gerror.New("storage upload session not found")
	}
	if session.pluginID != pluginID {
		return nil, gerror.New("storage upload session plugin mismatch")
	}
	if session.path != objectPath {
		return nil, gerror.New("storage upload session path mismatch")
	}
	return session, nil
}

func (s *storageUploadSessions) cleanupExpiredLocked(now time.Time) {
	for uploadID, session := range s.sessions {
		if session == nil || session.expiresAt.After(now) {
			continue
		}
		delete(s.sessions, uploadID)
		removeStorageUploadTempBestEffort(session.tempPath, session.tempDir)
	}
}

func removeStorageUploadTempBestEffort(tempPath string, tempDir string) {
	if err := removeStorageUploadTemp(tempPath, tempDir); err != nil {
		return
	}
}

func removeStorageUploadTemp(tempPath string, tempDir string) error {
	if tempPath != "" {
		if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if tempDir != "" {
		if err := os.RemoveAll(tempDir); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func appendStorageUploadChunk(tempPath string, offset int64, body []byte) error {
	file, err := os.OpenFile(tempPath, os.O_RDWR, 0600)
	if err != nil {
		return gerror.Wrap(err, "open storage upload temp file failed")
	}
	stat, err := file.Stat()
	if err != nil {
		return closeStorageUploadFileWithError(file, gerror.Wrap(err, "stat storage upload temp file failed"))
	}
	if stat.Size() != offset {
		return closeStorageUploadFileWithError(
			file,
			gerror.Newf("storage upload temp file offset mismatch: got %d want %d", stat.Size(), offset),
		)
	}
	written, err := file.WriteAt(body, offset)
	if err != nil || written != len(body) {
		truncateErr := file.Truncate(offset)
		if err != nil {
			return closeStorageUploadFileWithError(
				file,
				errors.Join(gerror.Wrap(err, "write storage upload chunk failed"), truncateErr),
			)
		}
		return closeStorageUploadFileWithError(file, errors.Join(io.ErrShortWrite, truncateErr))
	}
	if err = file.Close(); err != nil {
		return gerror.Wrap(err, "close storage upload temp file failed")
	}
	return nil
}

func closeStorageUploadFileWithError(file *os.File, err error) error {
	closeErr := file.Close()
	if closeErr == nil {
		return err
	}
	return errors.Join(err, gerror.Wrap(closeErr, "close storage upload temp file failed"))
}

func newStorageUploadID() (string, error) {
	var content [16]byte
	if _, err := rand.Read(content[:]); err != nil {
		return "", gerror.Wrap(err, "generate storage upload id failed")
	}
	return hex.EncodeToString(content[:]), nil
}
