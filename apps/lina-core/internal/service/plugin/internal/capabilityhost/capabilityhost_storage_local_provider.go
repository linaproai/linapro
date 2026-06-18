// This file implements the host built-in local disk storage provider used as
// the zero-configuration fallback for plugin object storage. The provider
// receives only scoped provider object keys and never plugin authorization
// snapshots.

package capabilityhost

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/os/gfile"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

const (
	localStorageProviderRootDir = ".capability-storage"
)

// localStorageProvider stores plugin objects in a local directory tree.
type localStorageProvider struct {
	rootDir string
}

// NewLocalStorageProvider creates the host built-in local storage provider.
func NewLocalStorageProvider(rootDir string) storagecap.Provider {
	return &localStorageProvider{
		rootDir: strings.TrimSpace(rootDir),
	}
}

// Put writes one scoped object key.
func (p *localStorageProvider) Put(ctx context.Context, in storagecap.ProviderPutInput) (*storagecap.ProviderObject, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectPath, err := p.resolveKey(in.Key)
	if err != nil {
		return nil, err
	}
	if !in.Overwrite {
		if _, exists, statErr := localStorageFileInfo(objectPath); statErr != nil {
			return nil, statErr
		} else if exists {
			return nil, bizerr.NewCode(storagecap.CodeStorageObjectExists)
		}
	}
	if err = gfile.Mkdir(filepath.Dir(objectPath)); err != nil {
		return nil, err
	}
	if in.Body == nil {
		in.Body = strings.NewReader("")
	}
	target, err := os.OpenFile(objectPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	_, copyErr := io.Copy(target, in.Body)
	closeErr := target.Close()
	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	fileInfo, exists, err := localStorageFileInfo(objectPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, bizerr.NewCode(storagecap.CodeStorageProviderUnavailable)
	}
	return p.providerObject(in.Key, fileInfo, in.ContentType), nil
}

// Get reads one scoped object key.
func (p *localStorageProvider) Get(ctx context.Context, in storagecap.ProviderGetInput) (*storagecap.ProviderGetOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectPath, err := p.resolveKey(in.Key)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localStorageFileInfo(objectPath)
	if err != nil || !exists {
		return &storagecap.ProviderGetOutput{Found: false}, err
	}
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, err
	}
	return &storagecap.ProviderGetOutput{
		Object: p.providerObject(in.Key, fileInfo, ""),
		Body:   file,
		Found:  true,
	}, nil
}

// Delete removes one scoped object key.
func (p *localStorageProvider) Delete(ctx context.Context, in storagecap.ProviderDeleteInput) error {
	if err := p.ensureAvailable(); err != nil {
		return err
	}
	objectPath, err := p.resolveKey(in.Key)
	if err != nil {
		return err
	}
	if err = os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteMany removes explicit scoped object keys.
func (p *localStorageProvider) DeleteMany(ctx context.Context, in storagecap.ProviderDeleteManyInput) error {
	for _, key := range in.Keys {
		if err := p.Delete(ctx, storagecap.ProviderDeleteInput{Key: key}); err != nil {
			return err
		}
	}
	return nil
}

// List lists scoped object keys under one prefix.
func (p *localStorageProvider) List(ctx context.Context, in storagecap.ProviderListInput) (*storagecap.ProviderListOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	rootDir, err := p.rootPath()
	if err != nil {
		return nil, err
	}
	prefix, err := normalizePluginStoragePath(in.Prefix, false)
	if err != nil {
		return nil, err
	}
	limit := normalizeStorageListLimit(in.Limit)
	prefixPath, err := p.resolveKey(prefix)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localStoragePathInfo(prefixPath)
	if err != nil || !exists {
		return &storagecap.ProviderListOutput{Objects: []*storagecap.ProviderObject{}}, nil
	}
	if !fileInfo.IsDir() {
		return &storagecap.ProviderListOutput{Objects: []*storagecap.ProviderObject{
			p.providerObject(prefix, fileInfo, ""),
		}}, nil
	}

	objects := make([]*storagecap.ProviderObject, 0, limit)
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
			return errLocalStorageListLimitReached
		}
		fileInfo, statErr := entry.Info()
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return nil
			}
			return statErr
		}
		relativePath, relErr := filepath.Rel(rootDir, absolutePath)
		if relErr != nil {
			return relErr
		}
		key := filepath.ToSlash(relativePath)
		objects = append(objects, p.providerObject(key, fileInfo, ""))
		if len(objects) >= limit {
			return errLocalStorageListLimitReached
		}
		return nil
	})
	if err != nil && !errors.Is(err, errLocalStorageListLimitReached) {
		return nil, err
	}
	return &storagecap.ProviderListOutput{Objects: objects}, nil
}

// ListCursor lists scoped object keys under one prefix using deterministic key order.
func (p *localStorageProvider) ListCursor(ctx context.Context, in storagecap.ProviderListCursorInput) (*storagecap.ProviderListCursorOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	rootDir, err := p.rootPath()
	if err != nil {
		return nil, err
	}
	prefix, err := normalizePluginStoragePath(in.Prefix, false)
	if err != nil {
		return nil, err
	}
	limit := normalizeStorageListLimit(in.Limit)
	prefixPath, err := p.resolveKey(prefix)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localStoragePathInfo(prefixPath)
	if err != nil || !exists {
		return &storagecap.ProviderListCursorOutput{Objects: []*storagecap.ProviderObject{}}, nil
	}
	if !fileInfo.IsDir() {
		if strings.TrimSpace(prefix) <= strings.TrimSpace(in.Cursor) {
			return &storagecap.ProviderListCursorOutput{Objects: []*storagecap.ProviderObject{}}, nil
		}
		return &storagecap.ProviderListCursorOutput{Objects: []*storagecap.ProviderObject{
			p.providerObject(prefix, fileInfo, ""),
		}}, nil
	}

	cursor := strings.TrimSpace(in.Cursor)
	page := make([]*storagecap.ProviderObject, 0, limit+1)
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
		relativePath, relErr := filepath.Rel(rootDir, absolutePath)
		if relErr != nil {
			return relErr
		}
		key := filepath.ToSlash(relativePath)
		if cursor != "" && key <= cursor {
			return nil
		}
		page = append(page, p.providerObject(key, fileInfo, ""))
		if len(page) > limit {
			nextCursor = strings.TrimSpace(page[limit-1].Key)
			page = page[:limit]
			return errLocalStorageListLimitReached
		}
		return nil
	})
	if err != nil && !errors.Is(err, errLocalStorageListLimitReached) {
		return nil, err
	}
	return &storagecap.ProviderListCursorOutput{Objects: page, NextCursor: nextCursor}, nil
}

// Stat reads scoped object metadata.
func (p *localStorageProvider) Stat(ctx context.Context, in storagecap.ProviderStatInput) (*storagecap.ProviderStatOutput, error) {
	if err := p.ensureAvailable(); err != nil {
		return nil, err
	}
	objectPath, err := p.resolveKey(in.Key)
	if err != nil {
		return nil, err
	}
	fileInfo, exists, err := localStorageFileInfo(objectPath)
	if err != nil || !exists {
		return &storagecap.ProviderStatOutput{Found: false}, err
	}
	return &storagecap.ProviderStatOutput{Object: p.providerObject(in.Key, fileInfo, ""), Found: true}, nil
}

// BatchStat reads metadata for explicit scoped object keys.
func (p *localStorageProvider) BatchStat(ctx context.Context, in storagecap.ProviderBatchStatInput) (*storagecap.ProviderBatchStatOutput, error) {
	output := &storagecap.ProviderBatchStatOutput{Objects: []*storagecap.ProviderObject{}}
	for _, key := range in.Keys {
		statOutput, err := p.Stat(ctx, storagecap.ProviderStatInput{Key: key})
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

// ensureAvailable verifies the local provider has a configured storage root.
func (p *localStorageProvider) ensureAvailable() error {
	if p == nil || strings.TrimSpace(p.rootDir) == "" {
		return bizerr.NewCode(storagecap.CodeStorageProviderUnavailable)
	}
	return nil
}

// rootPath returns the absolute provider root directory.
func (p *localStorageProvider) rootPath() (string, error) {
	root := filepath.Join(p.rootDir, localStorageProviderRootDir)
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absoluteRoot), nil
}

// resolveKey maps a scoped provider object key into a local filesystem path.
func (p *localStorageProvider) resolveKey(key string) (string, error) {
	normalizedKey, err := normalizePluginStoragePath(key, false)
	if err != nil {
		return "", err
	}
	rootDir, err := p.rootPath()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Clean(filepath.Join(rootDir, filepath.FromSlash(normalizedKey)))
	if fullPath != rootDir && !strings.HasPrefix(fullPath, rootDir+string(filepath.Separator)) {
		return "", bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	return fullPath, nil
}

// providerObject maps filesystem metadata into provider metadata.
func (p *localStorageProvider) providerObject(
	key string,
	fileInfo os.FileInfo,
	contentType string,
) *storagecap.ProviderObject {
	if strings.TrimSpace(contentType) == "" {
		contentType = detectObjectContentType("", nil, key)
	}
	if fileInfo == nil {
		return &storagecap.ProviderObject{Key: key, ContentType: contentType, Visibility: storagecap.VisibilityPrivate}
	}
	updatedAt := fileInfo.ModTime().UTC()
	return &storagecap.ProviderObject{
		Key:         key,
		Size:        fileInfo.Size(),
		ContentType: strings.TrimSpace(contentType),
		ETag:        localStorageETag(key, fileInfo),
		UpdatedAt:   &updatedAt,
		Visibility:  storagecap.VisibilityPrivate,
	}
}

// localStorageFileInfo returns file metadata while rejecting directory targets.
func localStorageFileInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if fileInfo.IsDir() {
		return nil, false, bizerr.NewCode(storagecap.CodeStoragePathInvalid)
	}
	return fileInfo, true, nil
}

// errLocalStorageListLimitReached stops bounded local provider traversal.
var errLocalStorageListLimitReached = errors.New("local storage list limit reached")

// localStoragePathInfo returns path metadata and allows directory targets.
func localStoragePathInfo(absolutePath string) (os.FileInfo, bool, error) {
	fileInfo, err := os.Stat(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return fileInfo, true, nil
}

// localStorageETag builds stable weak metadata from key, size, and update time.
func localStorageETag(key string, fileInfo os.FileInfo) string {
	if fileInfo == nil {
		return ""
	}
	hash := sha1.Sum([]byte(key + "|" + strconv.FormatInt(fileInfo.Size(), 10) + "|" + fileInfo.ModTime().UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(hash[:])
}
