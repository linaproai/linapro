// This file implements the local filesystem backend for the host-internal
// object storage service, including namespace-root resolution and safe key
// normalization.

package storage

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/closeutil"
	"lina-core/pkg/plugin/capability/storagecap"
)

var errListLimitReached = errors.New("storage list limit reached")

// Put writes an object into the selected namespace root.
func (s *serviceImpl) Put(ctx context.Context, in PutInput) (out *PutOutput, err error) {
	objectPath, normalizedKey, err := s.resolveObjectPath(in.Namespace, in.Key, false)
	if err != nil {
		return nil, err
	}
	if !in.Overwrite {
		if _, exists, statErr := localFileInfo(objectPath); statErr != nil {
			return nil, statErr
		} else if exists {
			return nil, ErrObjectExist
		}
	}
	if err = gfile.Mkdir(filepath.Dir(objectPath)); err != nil {
		return nil, err
	}
	body := in.Body
	if body == nil {
		body = strings.NewReader("")
	}
	target, err := os.OpenFile(objectPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	defer closeutil.Close(ctx, target, &err, "close storage object file failed")
	if _, err = io.Copy(target, body); err != nil {
		if removeErr := os.Remove(objectPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return nil, errors.Join(err, removeErr)
		}
		return nil, err
	}
	fileInfo, exists, err := localFileInfo(objectPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUnavailable
	}
	return &PutOutput{Object: storageObject(normalizedKey, fileInfo, in.ContentType)}, nil
}

// Get opens an object stream from the selected namespace root.
func (s *serviceImpl) Get(_ context.Context, in GetInput) (*GetOutput, error) {
	objectPath, normalizedKey, err := s.resolveObjectPath(in.Namespace, in.Key, false)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localFileInfo(objectPath)
	if err != nil || !exists {
		return &GetOutput{Found: false}, err
	}
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, err
	}
	return &GetOutput{
		Object: storageObject(normalizedKey, fileInfo, ""),
		Body:   file,
		Found:  true,
	}, nil
}

// Delete removes one object from the selected namespace root.
func (s *serviceImpl) Delete(_ context.Context, in DeleteInput) error {
	objectPath, _, err := s.resolveObjectPath(in.Namespace, in.Key, false)
	if err != nil {
		return err
	}
	if err = os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteMany removes explicit object keys from the selected namespace root.
func (s *serviceImpl) DeleteMany(ctx context.Context, in DeleteManyInput) error {
	for _, key := range in.Keys {
		if err := s.Delete(ctx, DeleteInput{Namespace: in.Namespace, Key: key}); err != nil {
			return err
		}
	}
	return nil
}

// List returns objects under a prefix using a bounded traversal.
func (s *serviceImpl) List(_ context.Context, in ListInput) (*ListOutput, error) {
	limit := normalizeListLimit(in.Limit)
	rootPath, err := s.namespaceRoot(in.Namespace)
	if err != nil {
		return nil, err
	}
	prefixPath, normalizedPrefix, err := s.resolveObjectPath(in.Namespace, in.Prefix, true)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localPathInfo(prefixPath)
	if err != nil || !exists {
		return &ListOutput{Objects: []*Object{}, Limit: limit}, err
	}
	if !fileInfo.IsDir() {
		return &ListOutput{Objects: []*Object{storageObject(normalizedPrefix, fileInfo, "")}, Limit: limit}, nil
	}

	objects := make([]*Object, 0, limit)
	err = filepath.WalkDir(prefixPath, func(absolutePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if len(objects) >= limit {
			return errListLimitReached
		}
		fileInfo, statErr := entry.Info()
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return nil
			}
			return statErr
		}
		key, relErr := relativeObjectKey(rootPath, absolutePath)
		if relErr != nil {
			return relErr
		}
		objects = append(objects, storageObject(key, fileInfo, ""))
		if len(objects) >= limit {
			return errListLimitReached
		}
		return nil
	})
	if err != nil && !errors.Is(err, errListLimitReached) {
		return nil, err
	}
	return &ListOutput{Objects: objects, Limit: limit}, nil
}

// ListCursor returns objects under a prefix after an optional cursor key.
func (s *serviceImpl) ListCursor(_ context.Context, in ListCursorInput) (*ListCursorOutput, error) {
	limit := normalizeListLimit(in.Limit)
	rootPath, err := s.namespaceRoot(in.Namespace)
	if err != nil {
		return nil, err
	}
	prefixPath, normalizedPrefix, err := s.resolveObjectPath(in.Namespace, in.Prefix, true)
	if err != nil {
		return nil, err
	}
	cursor := strings.TrimSpace(strings.ReplaceAll(in.Cursor, "\\", "/"))
	if cursor != "" {
		cursor, err = normalizeObjectKey(cursor, false)
		if err != nil {
			return nil, err
		}
	}
	fileInfo, exists, err := localPathInfo(prefixPath)
	if err != nil || !exists {
		return &ListCursorOutput{Objects: []*Object{}, Limit: limit}, err
	}
	if !fileInfo.IsDir() {
		if cursor != "" && normalizedPrefix <= cursor {
			return &ListCursorOutput{Objects: []*Object{}, Limit: limit}, nil
		}
		return &ListCursorOutput{Objects: []*Object{storageObject(normalizedPrefix, fileInfo, "")}, Limit: limit}, nil
	}

	page := make([]*Object, 0, limit+1)
	nextCursor := ""
	err = filepath.WalkDir(prefixPath, func(absolutePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsNotExist(walkErr) {
				return nil
			}
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		fileInfo, statErr := entry.Info()
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return nil
			}
			return statErr
		}
		key, relErr := relativeObjectKey(rootPath, absolutePath)
		if relErr != nil {
			return relErr
		}
		if cursor != "" && key <= cursor {
			return nil
		}
		page = append(page, storageObject(key, fileInfo, ""))
		if len(page) > limit {
			nextCursor = strings.TrimSpace(page[limit-1].Key)
			page = page[:limit]
			return errListLimitReached
		}
		return nil
	})
	if err != nil && !errors.Is(err, errListLimitReached) {
		return nil, err
	}
	return &ListCursorOutput{Objects: page, NextCursor: nextCursor, Limit: limit}, nil
}

// Stat returns metadata for one object in the selected namespace root.
func (s *serviceImpl) Stat(_ context.Context, in StatInput) (*StatOutput, error) {
	objectPath, normalizedKey, err := s.resolveObjectPath(in.Namespace, in.Key, false)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localFileInfo(objectPath)
	if err != nil || !exists {
		return &StatOutput{Found: false}, err
	}
	return &StatOutput{Object: storageObject(normalizedKey, fileInfo, ""), Found: true}, nil
}

// BatchStat returns metadata for explicit object keys in one namespace.
func (s *serviceImpl) BatchStat(ctx context.Context, in BatchStatInput) (*BatchStatOutput, error) {
	output := &BatchStatOutput{Objects: []*Object{}}
	for _, key := range in.Keys {
		statOutput, err := s.Stat(ctx, StatInput{Namespace: in.Namespace, Key: key})
		if err != nil {
			return nil, err
		}
		if statOutput == nil || !statOutput.Found {
			output.MissingKeys = append(output.MissingKeys, key)
			continue
		}
		output.Objects = append(output.Objects, statOutput.Object)
	}
	return output, nil
}

func (s *serviceImpl) resolveObjectPath(namespace string, key string, allowEmpty bool) (string, string, error) {
	rootPath, err := s.namespaceRoot(namespace)
	if err != nil {
		return "", "", err
	}
	normalizedKey, err := normalizeObjectKey(key, allowEmpty)
	if err != nil {
		return "", "", err
	}
	fullPath := filepath.Clean(filepath.Join(rootPath, filepath.FromSlash(normalizedKey)))
	if normalizedKey == "" {
		fullPath = rootPath
	}
	if fullPath != rootPath && !strings.HasPrefix(fullPath, rootPath+string(filepath.Separator)) {
		return "", "", ErrPathInvalid
	}
	return fullPath, normalizedKey, nil
}

func (s *serviceImpl) namespaceRoot(namespace string) (string, error) {
	normalizedNamespace := strings.TrimSpace(namespace)
	if normalizedNamespace == "" {
		return "", ErrPathInvalid
	}
	if s == nil {
		return "", ErrUnavailable
	}
	root := strings.TrimSpace(s.rootDir)
	if s != nil && s.namespaceRoots != nil {
		if namespaceRoot := strings.TrimSpace(s.namespaceRoots[normalizedNamespace]); namespaceRoot != "" {
			root = namespaceRoot
		}
	}
	if root == "" {
		return "", ErrUnavailable
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absoluteRoot), nil
}

func normalizeObjectKey(rawKey string, allowEmpty bool) (string, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(rawKey), "\\", "/")
	if trimmed == "" || trimmed == "." {
		if allowEmpty {
			return "", nil
		}
		return "", ErrPathInvalid
	}
	if strings.Contains(trimmed, "://") || strings.HasPrefix(trimmed, "/") {
		return "", ErrPathInvalid
	}
	if len(trimmed) >= 2 && ((trimmed[0] >= 'A' && trimmed[0] <= 'Z') || (trimmed[0] >= 'a' && trimmed[0] <= 'z')) && trimmed[1] == ':' {
		return "", ErrPathInvalid
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", ErrPathInvalid
	}
	return normalized, nil
}

func normalizeListLimit(limit int) int {
	if limit <= 0 {
		return DefaultListLimit
	}
	if limit > MaxListLimit {
		return MaxListLimit
	}
	return limit
}

func relativeObjectKey(rootPath string, absolutePath string) (string, error) {
	relativePath, err := filepath.Rel(rootPath, absolutePath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(relativePath), nil
}

func localFileInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if fileInfo.IsDir() {
		return nil, false, ErrPathInvalid
	}
	return fileInfo, true, nil
}

func localPathInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return fileInfo, true, nil
}

func storageObject(key string, fileInfo os.FileInfo, contentType string) *Object {
	if fileInfo == nil {
		return &Object{Key: key, ContentType: strings.TrimSpace(contentType)}
	}
	updatedAt := fileInfo.ModTime().UTC()
	return &Object{
		Key:         key,
		Size:        fileInfo.Size(),
		ContentType: strings.TrimSpace(contentType),
		ETag:        localETag(key, fileInfo),
		UpdatedAt:   &updatedAt,
	}
}

func localETag(key string, fileInfo os.FileInfo) string {
	if fileInfo == nil {
		return ""
	}
	hash := sha1.Sum([]byte(key + "|" + strconv.FormatInt(fileInfo.Size(), 10) + "|" + fileInfo.ModTime().UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(hash[:])
}

// CreateDirectAccess always returns proxy mode for the local filesystem backend.
func (s *serviceImpl) CreateDirectAccess(_ context.Context, in DirectAccessInput) (*DirectAccessOutput, error) {
	op := storagecap.NormalizeDirectAccessOperation(in.Operation)
	if op == "" {
		op = storagecap.DirectAccessOpPut
	}
	return &DirectAccessOutput{
		Access: &storagecap.DirectAccess{
			Mode:       storagecap.DirectAccessModeProxy,
			Operation:  op,
			ProviderID: storagecap.LocalProviderID,
		},
		ProviderID: storagecap.LocalProviderID,
	}, nil
}
