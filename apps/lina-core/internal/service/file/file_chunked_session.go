// This file implements process-local proxy chunked-upload sessions for the file
// center. Sessions bind tenant, actor, storage key, and either a temporary part
// assembly file or a cloud multipart upload id.

package file

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"time"

	"lina-core/pkg/bizerr"
)

// chunkedUploadSession stores host-owned state for one proxy chunked upload.
type chunkedUploadSession struct {
	ID           string
	TenantID     int64
	UserID       int64
	Scene        string
	OriginalName string
	Suffix       string
	ContentType  string
	Size         int64
	ContentHash  string
	StoragePath  string
	ProviderID   string
	// PartSize is the planned intermediate part size in bytes for this session.
	PartSize int64
	// UseCloudMultipart reports whether parts are uploaded via storage Multipart.
	UseCloudMultipart bool
	// CloudUploadID is the provider multipart upload id when UseCloudMultipart is true.
	CloudUploadID string
	// TempDir/TempPath hold assembled payload when not using cloud multipart.
	TempDir  string
	TempPath string
	// ReceivedParts maps partNumber -> etag (cloud) or placeholder (local assembly).
	ReceivedParts map[int32]string
	// ReceivedBytes is cumulative accepted payload size for local assembly progress.
	ReceivedBytes int64
	ExpiresAt     time.Time
	// Completed marks an idempotent complete; FileID/CompletedURL are then valid.
	Completed    bool
	FileID       int64
	CompletedURL string
}

// chunkedUploadSessionStore is a process-local session map with lazy expiry purge.
type chunkedUploadSessionStore struct {
	mu       sync.Mutex
	sessions map[string]*chunkedUploadSession
}

func newChunkedUploadSessionStore() *chunkedUploadSessionStore {
	return &chunkedUploadSessionStore{sessions: make(map[string]*chunkedUploadSession)}
}

func (s *chunkedUploadSessionStore) put(session *chunkedUploadSession) {
	if s == nil || session == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions == nil {
		s.sessions = make(map[string]*chunkedUploadSession)
	}
	s.purgeExpiredLocked(time.Now().UTC())
	s.sessions[session.ID] = session
}

func (s *chunkedUploadSessionStore) get(id string) (*chunkedUploadSession, error) {
	if s == nil {
		return nil, bizerr.NewCode(CodeFileChunkedSessionInvalid)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, bizerr.NewCode(CodeFileChunkedSessionInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(time.Now().UTC())
	session := s.sessions[id]
	if session == nil {
		return nil, bizerr.NewCode(CodeFileChunkedSessionInvalid)
	}
	if !session.Completed && !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		s.removeLocked(id)
		return nil, bizerr.NewCode(CodeFileChunkedSessionExpired)
	}
	// Return a shallow copy so callers can mutate maps under separate locks carefully.
	// Callers that mutate session fields should re-put or use store methods.
	return session, nil
}

func (s *chunkedUploadSessionStore) update(id string, fn func(*chunkedUploadSession) error) error {
	if s == nil {
		return bizerr.NewCode(CodeFileChunkedSessionInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(time.Now().UTC())
	session := s.sessions[strings.TrimSpace(id)]
	if session == nil {
		return bizerr.NewCode(CodeFileChunkedSessionInvalid)
	}
	if !session.Completed && !session.ExpiresAt.IsZero() && time.Now().UTC().After(session.ExpiresAt) {
		s.removeLocked(id)
		return bizerr.NewCode(CodeFileChunkedSessionExpired)
	}
	return fn(session)
}

func (s *chunkedUploadSessionStore) markCompleted(id string, fileID int64, url string) error {
	return s.update(id, func(session *chunkedUploadSession) error {
		session.Completed = true
		session.FileID = fileID
		session.CompletedURL = url
		removeChunkedTempBestEffort(session.TempPath, session.TempDir)
		session.TempPath = ""
		session.TempDir = ""
		return nil
	})
}

func (s *chunkedUploadSessionStore) delete(id string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removeLocked(strings.TrimSpace(id))
}

func (s *chunkedUploadSessionStore) removeLocked(id string) {
	if s.sessions == nil {
		return
	}
	session := s.sessions[id]
	if session == nil {
		return
	}
	delete(s.sessions, id)
	removeChunkedTempBestEffort(session.TempPath, session.TempDir)
}

func (s *chunkedUploadSessionStore) purgeExpiredLocked(now time.Time) {
	for id, session := range s.sessions {
		if session == nil || session.Completed {
			continue
		}
		if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
			s.removeLocked(id)
		}
	}
}

func newChunkedUploadSessionID() (string, error) {
	var content [16]byte
	if _, err := rand.Read(content[:]); err != nil {
		return "", bizerr.WrapCode(err, CodeFileDirectInitFailed)
	}
	return hex.EncodeToString(content[:]), nil
}

func removeChunkedTempBestEffort(tempPath string, tempDir string) {
	if tempPath != "" {
		_ = os.Remove(tempPath)
	}
	if tempDir != "" {
		_ = os.RemoveAll(tempDir)
	}
}

var processChunkedUploadSessions = newChunkedUploadSessionStore()

func defaultChunkedUploadSessions() *chunkedUploadSessionStore {
	return processChunkedUploadSessions
}
