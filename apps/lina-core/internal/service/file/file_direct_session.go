// This file implements in-memory direct-upload sessions for the file center.
// Sessions bind tenant, actor, storage key, size, and expiry before complete
// writes sys_file metadata.

package file

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// directUploadSession stores host-owned state for one client direct upload.
type directUploadSession struct {
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
	ProviderKey  string
	ExpiresAt    time.Time
	Completed    bool
	FileID       int64
	CompletedURL string
	CompletedAt  time.Time
}

// directUploadSessionStore is a process-local session map with lazy expiry purge.
type directUploadSessionStore struct {
	mu       sync.Mutex
	sessions map[string]*directUploadSession
}

func newDirectUploadSessionStore() *directUploadSessionStore {
	return &directUploadSessionStore{sessions: make(map[string]*directUploadSession)}
}

func (s *directUploadSessionStore) put(session *directUploadSession) {
	if s == nil || session == nil || strings.TrimSpace(session.ID) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeExpiredLocked(time.Now().UTC())
	cp := *session
	s.sessions[session.ID] = &cp
}

func (s *directUploadSessionStore) get(id string) (*directUploadSession, error) {
	if s == nil {
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	session, ok := s.sessions[id]
	if !ok || session == nil {
		s.purgeExpiredLocked(now)
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	if !session.Completed && !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
		delete(s.sessions, id)
		s.purgeExpiredLocked(now)
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectSessionExpired)
	}
	s.purgeExpiredLocked(now)
	// Re-read after purge in case purge removed unrelated sessions only.
	session, ok = s.sessions[id]
	if !ok || session == nil {
		return nil, bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	cp := *session
	return &cp, nil
}

func (s *directUploadSessionStore) markCompleted(id string, fileID int64, url string) error {
	if s == nil {
		return bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[strings.TrimSpace(id)]
	if !ok || session == nil {
		return bizerr.NewCode(storagecap.CodeStorageDirectSessionInvalid)
	}
	session.Completed = true
	session.FileID = fileID
	session.CompletedURL = url
	session.CompletedAt = time.Now().UTC()
	return nil
}

func (s *directUploadSessionStore) delete(id string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, strings.TrimSpace(id))
}

func (s *directUploadSessionStore) purgeExpiredLocked(now time.Time) {
	for id, session := range s.sessions {
		if session == nil {
			delete(s.sessions, id)
			continue
		}
		if session.Completed {
			// Keep completed sessions briefly for idempotent complete; drop after 2x TTL window.
			// Drop completed sessions after twice the maximum access lifetime.
			if !session.CompletedAt.IsZero() && now.Sub(session.CompletedAt) > 2*time.Hour {
				delete(s.sessions, id)
			}
			continue
		}
		if !session.ExpiresAt.IsZero() && !session.ExpiresAt.After(now) {
			delete(s.sessions, id)
		}
	}
}

func newDirectUploadSessionID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

// Ensure package-level store is shared by the process file service graph when
// a service instance does not own a private store (for example some tests).
func defaultDirectUploadSessions() *directUploadSessionStore {
	return processDirectUploadSessions
}

var processDirectUploadSessions = newDirectUploadSessionStore()
