// This file implements file-center client direct upload init/complete/abort and
// direct download access issuance. Object keys remain host-assigned; permanent
// credentials are never returned to clients.

package file

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	storagesvc "lina-core/internal/service/storage"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// DirectUploadInitInput starts one file-center direct upload session.
type DirectUploadInitInput struct {
	Scene       string
	FileName    string
	Size        int64
	ContentType string
	// ContentHash is optional SHA-256 hex; when set, enables instant reuse lookup.
	ContentHash string
}

// DirectUploadInitOutput is the domain result of direct upload init.
type DirectUploadInitOutput struct {
	// InstantReuse is true when ContentHash matched an existing file and no upload is needed.
	InstantReuse bool
	// UploadSessionID is set when a new session was opened (absent on instant reuse).
	UploadSessionID string
	// Access is the neutral client transfer description; proxy mode means host-mediated upload.
	Access *storagecap.DirectAccess
	// File is populated only when InstantReuse is true.
	File *UploadOutput
	// Strategy is the planned channel/encoding for non-instant uploads.
	Strategy *UploadStrategy
	// Multipart is populated when Strategy.Encoding is multipart.
	Multipart *UploadMultipartPlan
}

// DirectUploadCompleteInput finishes one direct upload session.
type DirectUploadCompleteInput struct {
	UploadSessionID string
	// ETag is optional for single-object put sessions.
	ETag string
	// Parts is required when the session encoding is multipart.
	Parts []MultipartPartRef
}

// DirectUploadPartURLInput issues client access for one direct multipart part.
type DirectUploadPartURLInput struct {
	UploadSessionID string
	PartNumber      int32
	// Size is the expected part size when known; zero means unspecified.
	Size int64
}

// DirectUploadPartURLOutput returns neutral part access.
type DirectUploadPartURLOutput struct {
	Access *storagecap.DirectAccess
}

// DirectUploadAbortInput aborts one direct upload session.
type DirectUploadAbortInput struct {
	UploadSessionID string
}

// DirectDownloadInput requests client get access for one file id.
type DirectDownloadInput struct {
	ID int64
}

// DirectDownloadOutput returns direct get access or proxy indication.
type DirectDownloadOutput struct {
	Access *storagecap.DirectAccess
	// ProxyURL is the host download path when Access is proxy mode.
	ProxyURL string
}

// DirectUploadInit creates a session and issues put access, or reuses by hash.
func (s *serviceImpl) DirectUploadInit(ctx context.Context, in *DirectUploadInitInput) (*DirectUploadInitOutput, error) {
	if in == nil {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	sanitizedFilename := sanitizeFilename(in.FileName)
	if sanitizedFilename == "" || sanitizedFilename == "." {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	uploadMaxSize, err := s.configSvc.GetUploadMaxSize(ctx)
	if err != nil {
		return nil, err
	}
	if in.Size <= 0 {
		return nil, bizerr.NewCode(CodeFileUploadRequired)
	}
	if in.Size > uploadLimitBytes(uploadMaxSize) {
		return nil, bizerr.NewCode(CodeFileTooLarge, bizerr.P("maxSizeMB", uploadMaxSize))
	}
	if err = s.ensureFilesBackendReady(ctx); err != nil {
		return nil, err
	}

	scene := strings.TrimSpace(in.Scene)
	if scene == "" {
		scene = DefaultFileSceneOther
	}
	suffix := gstr.ToLower(gfile.ExtName(sanitizedFilename))
	contentHash := strings.ToLower(strings.TrimSpace(in.ContentHash))
	tenantID := int64(datascope.CurrentTenantID(ctx))
	userID := s.currentUserID(ctx)

	if contentHash != "" {
		reused, reuseErr := s.tryInstantReuseByHash(ctx, tenantID, userID, sanitizedFilename, suffix, scene, in.Size, contentHash)
		if reuseErr != nil {
			return nil, reuseErr
		}
		if reused != nil {
			return &DirectUploadInitOutput{InstantReuse: true, File: reused}, nil
		}
	}

	storagePath := buildStorageKey(ctx, sanitizedFilename)
	ttl, err := s.resolveDirectUploadTTL(ctx)
	if err != nil {
		return nil, err
	}
	directOut, err := s.storage.CreateDirectAccess(ctx, storagesvc.DirectAccessInput{
		Namespace:   storagesvc.NamespaceFiles,
		Key:         storagePath,
		Operation:   storagecap.DirectAccessOpPut,
		Size:        in.Size,
		ContentType: strings.TrimSpace(in.ContentType),
		TTL:         ttl,
		Overwrite:   true,
	})
	if err != nil {
		return nil, mapStorageInitError(err)
	}
	access := directOut.Access
	if access == nil {
		access = &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpPut}
	}

	strategy, multipartPlan, planErr := s.planDirectUpload(ctx, in.Size, uploadMaxSize, !storagecap.IsProxyDirectAccess(access))
	if planErr != nil {
		return nil, planErr
	}
	if strategy.Channel == UploadChannelProxy {
		return &DirectUploadInitOutput{
			Access:    &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpPut},
			Strategy:  &strategy,
			Multipart: multipartPlan,
		}, nil
	}

	return s.openDirectUploadSession(ctx, openDirectUploadSessionInput{
		tenantID:      tenantID,
		userID:        userID,
		scene:         scene,
		filename:      sanitizedFilename,
		suffix:        suffix,
		contentType:   strings.TrimSpace(in.ContentType),
		size:          in.Size,
		contentHash:   contentHash,
		storagePath:   storagePath,
		ttl:           ttl,
		access:        access,
		directOut:     directOut,
		strategy:      strategy,
		multipartPlan: multipartPlan,
	})
}

type openDirectUploadSessionInput struct {
	tenantID      int64
	userID        int64
	scene         string
	filename      string
	suffix        string
	contentType   string
	size          int64
	contentHash   string
	storagePath   string
	ttl           time.Duration
	access        *storagecap.DirectAccess
	directOut     *storagesvc.DirectAccessOutput
	strategy      UploadStrategy
	multipartPlan *UploadMultipartPlan
}

func (s *serviceImpl) planDirectUpload(
	ctx context.Context,
	sizeBytes int64,
	uploadMaxSize int64,
	supportsDirect bool,
) (UploadStrategy, *UploadMultipartPlan, error) {
	supportsCloudMP, err := s.resolveCloudMultipartSupport(ctx)
	if err != nil {
		return UploadStrategy{}, nil, mapStorageInitError(err)
	}
	multipartEnabled, err := s.configSvc.GetUploadMultipartEnabled(ctx)
	if err != nil {
		return UploadStrategy{}, nil, err
	}
	thresholdMB, err := s.configSvc.GetUploadMultipartThresholdMB(ctx)
	if err != nil {
		return UploadStrategy{}, nil, err
	}
	if thresholdMB >= uploadMaxSize {
		multipartEnabled = false
	}
	strategy := planUploadStrategy(sizeBytes, multipartEnabled, thresholdMB, supportsDirect, supportsCloudMP)
	if strategy.Encoding != UploadEncodingMultipart {
		return strategy, nil, nil
	}
	partSizeMB, err := s.configSvc.GetUploadMultipartPartSizeMB(ctx)
	if err != nil {
		return UploadStrategy{}, nil, err
	}
	concurrency, err := s.configSvc.GetUploadMultipartMaxConcurrency(ctx)
	if err != nil {
		return UploadStrategy{}, nil, err
	}
	return strategy, buildMultipartPlan(partSizeMB, concurrency), nil
}

func (s *serviceImpl) openDirectUploadSession(ctx context.Context, in openDirectUploadSessionInput) (*DirectUploadInitOutput, error) {
	sessionID, err := newDirectUploadSessionID()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileDirectInitFailed)
	}
	expiresAt := time.Now().UTC().Add(in.ttl)
	if !in.access.ExpiresAt.IsZero() && in.access.ExpiresAt.Before(expiresAt) {
		expiresAt = in.access.ExpiresAt.UTC()
	}
	providerID := strings.TrimSpace(in.directOut.ProviderID)
	if providerID == "" {
		providerID = strings.TrimSpace(in.access.ProviderID)
	}
	session := &directUploadSession{
		ID:           sessionID,
		TenantID:     in.tenantID,
		UserID:       in.userID,
		Scene:        in.scene,
		OriginalName: in.filename,
		Suffix:       in.suffix,
		ContentType:  in.contentType,
		Size:         in.size,
		ContentHash:  in.contentHash,
		StoragePath:  in.storagePath,
		ProviderID:   providerID,
		ProviderKey:  in.directOut.ProviderKey,
		ExpiresAt:    expiresAt,
		Encoding:     in.strategy.Encoding,
	}
	outAccess := in.access
	if in.strategy.Encoding == UploadEncodingMultipart {
		created, createErr := s.storage.CreateMultipart(ctx, storagesvc.MultipartCreateInput{
			Namespace:   storagesvc.NamespaceFiles,
			Key:         in.storagePath,
			ContentType: in.contentType,
			Overwrite:   true,
		})
		if createErr != nil {
			if bizerr.Is(createErr, storagecap.CodeStorageProviderConflict) {
				return nil, bizerr.WrapCode(createErr, CodeFileStorageConflict)
			}
			return &DirectUploadInitOutput{
				Access:    &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpPut},
				Strategy:  &UploadStrategy{Channel: UploadChannelProxy, Encoding: UploadEncodingMultipart},
				Multipart: in.multipartPlan,
			}, nil
		}
		if created == nil || strings.TrimSpace(created.UploadID) == "" {
			return nil, bizerr.NewCode(CodeFileDirectInitFailed)
		}
		session.CloudUploadID = created.UploadID
		if strings.TrimSpace(created.ProviderID) != "" {
			session.ProviderID = created.ProviderID
			providerID = created.ProviderID
		}
		if strings.TrimSpace(created.ProviderKey) != "" {
			session.ProviderKey = created.ProviderKey
		}
		outAccess = &storagecap.DirectAccess{
			Mode:       storagecap.DirectAccessModePresignedURL,
			Operation:  storagecap.DirectAccessOpPut,
			ProviderID: providerID,
		}
	}
	s.sessionStore().put(session)
	return &DirectUploadInitOutput{
		UploadSessionID: sessionID,
		Access:          outAccess,
		Strategy:        &in.strategy,
		Multipart:       in.multipartPlan,
	}, nil
}

func mapStorageInitError(err error) error {
	if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
		return bizerr.WrapCode(err, CodeFileStorageConflict)
	}
	return bizerr.WrapCode(err, CodeFileDirectInitFailed)
}

// DirectUploadComplete validates the uploaded object and creates sys_file.
func (s *serviceImpl) DirectUploadComplete(ctx context.Context, in *DirectUploadCompleteInput) (*UploadOutput, error) {
	if in == nil || strings.TrimSpace(in.UploadSessionID) == "" {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	session, err := s.sessionStore().get(in.UploadSessionID)
	if err != nil {
		return s.mapSessionError(err)
	}
	if session.TenantID != int64(datascope.CurrentTenantID(ctx)) {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	if session.Completed {
		return &UploadOutput{
			Id:       session.FileID,
			Name:     gfile.Basename(session.StoragePath),
			Original: session.OriginalName,
			Url:      s.getBaseUrl(ctx) + session.CompletedURL,
			Suffix:   session.Suffix,
			Size:     session.Size,
		}, nil
	}

	if strings.EqualFold(strings.TrimSpace(session.Encoding), UploadEncodingMultipart) {
		parts := in.Parts
		if len(parts) == 0 {
			return nil, bizerr.NewCode(CodeFileDirectCompleteFailed)
		}
		sorted := sortMultipartParts(parts)
		completeParts := make([]storagesvc.MultipartCompletedPart, 0, len(sorted))
		for _, part := range sorted {
			if part.PartNumber < 1 || strings.TrimSpace(part.ETag) == "" {
				return nil, bizerr.NewCode(CodeFileDirectCompleteFailed)
			}
			completeParts = append(completeParts, storagesvc.MultipartCompletedPart{
				PartNumber: part.PartNumber,
				ETag:       strings.TrimSpace(part.ETag),
			})
		}
		if _, err = s.storage.CompleteMultipart(ctx, storagesvc.MultipartCompleteInput{
			Namespace: storagesvc.NamespaceFiles,
			Key:       session.StoragePath,
			UploadID:  session.CloudUploadID,
			Parts:     completeParts,
		}); err != nil {
			return nil, bizerr.WrapCode(err, CodeFileDirectCompleteFailed)
		}
	}

	stat, err := s.storage.Stat(ctx, storagesvc.StatInput{
		Namespace: storagesvc.NamespaceFiles,
		Key:       session.StoragePath,
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileDirectCompleteFailed)
	}
	if stat == nil || !stat.Found || stat.Object == nil {
		return nil, bizerr.NewCode(CodeFileDirectCompleteFailed)
	}
	if session.Size > 0 && stat.Object.Size > 0 && stat.Object.Size != session.Size {
		return nil, bizerr.NewCode(CodeFileDirectCompleteFailed)
	}

	engine := strings.TrimSpace(session.ProviderID)
	if engine == "" {
		engine = storagesvc.FilesProviderID(ctx, s.storage)
	}
	if engine == "" {
		engine = EngineLocal
	}
	url := storageURL(session.StoragePath)
	storedName := gfile.Basename(session.StoragePath)
	hash := session.ContentHash
	if hash == "" {
		// Content hash is optional for direct upload; empty hash disables future
		// hash-based reuse for this record until a future backfill job.
		hash = ""
	}

	var output *UploadOutput
	err = dao.SysFile.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		result, insertErr := dao.SysFile.Ctx(ctx).Data(do.SysFile{
			TenantId:  session.TenantID,
			Name:      storedName,
			Original:  session.OriginalName,
			Suffix:    session.Suffix,
			Scene:     session.Scene,
			Size:      session.Size,
			Hash:      hash,
			Url:       url,
			Path:      session.StoragePath,
			Engine:    engine,
			CreatedBy: session.UserID,
		}).Insert()
		if insertErr != nil {
			return bizerr.WrapCode(insertErr, CodeFileRecordSaveFailed)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			return bizerr.WrapCode(idErr, CodeFileRecordIDReadFailed)
		}
		output = &UploadOutput{
			Id:       id,
			Name:     storedName,
			Original: session.OriginalName,
			Url:      s.getBaseUrl(ctx) + url,
			Suffix:   session.Suffix,
			Size:     session.Size,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.sessionStore().markCompleted(session.ID, output.Id, url)
	return output, nil
}

// DirectUploadAbort discards one in-flight direct upload session.
func (s *serviceImpl) DirectUploadAbort(ctx context.Context, in *DirectUploadAbortInput) error {
	if in == nil || strings.TrimSpace(in.UploadSessionID) == "" {
		return nil
	}
	session, err := s.sessionStore().get(in.UploadSessionID)
	if err == nil && session != nil &&
		strings.EqualFold(strings.TrimSpace(session.Encoding), UploadEncodingMultipart) &&
		strings.TrimSpace(session.CloudUploadID) != "" {
		_ = s.storage.AbortMultipart(ctx, storagesvc.MultipartAbortInput{
			Namespace: storagesvc.NamespaceFiles,
			Key:       session.StoragePath,
			UploadID:  session.CloudUploadID,
		})
	}
	s.sessionStore().delete(in.UploadSessionID)
	return nil
}

// DirectUploadPartURL issues short-lived access for one direct multipart part.
func (s *serviceImpl) DirectUploadPartURL(ctx context.Context, in *DirectUploadPartURLInput) (*DirectUploadPartURLOutput, error) {
	if in == nil || strings.TrimSpace(in.UploadSessionID) == "" || in.PartNumber < 1 {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	session, err := s.sessionStore().get(in.UploadSessionID)
	if err != nil {
		_, mapped := s.mapSessionError(err)
		return nil, mapped
	}
	if session.TenantID != int64(datascope.CurrentTenantID(ctx)) {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	if !strings.EqualFold(strings.TrimSpace(session.Encoding), UploadEncodingMultipart) {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	if strings.TrimSpace(session.CloudUploadID) == "" {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	ttl, err := s.resolveDirectUploadTTL(ctx)
	if err != nil {
		return nil, err
	}
	accessOut, err := s.storage.CreateMultipartPartAccess(ctx, storagesvc.MultipartPartAccessInput{
		Namespace:   storagesvc.NamespaceFiles,
		Key:         session.StoragePath,
		UploadID:    session.CloudUploadID,
		PartNumber:  in.PartNumber,
		Size:        in.Size,
		ContentType: session.ContentType,
		TTL:         ttl,
	})
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileDirectInitFailed)
	}
	if accessOut == nil || accessOut.Access == nil {
		return nil, bizerr.NewCode(CodeFileDirectInitFailed)
	}
	return &DirectUploadPartURLOutput{Access: accessOut.Access}, nil
}

// DirectDownload issues get access for one visible file, or proxy mode.
func (s *serviceImpl) DirectDownload(ctx context.Context, in *DirectDownloadInput) (*DirectDownloadOutput, error) {
	if in == nil || in.ID <= 0 {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	record, err := s.Info(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if record == nil || strings.TrimSpace(record.Path) == "" {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	ttl, err := s.resolveDirectUploadTTL(ctx)
	if err != nil {
		return nil, err
	}
	directOut, err := s.storage.CreateDirectAccess(ctx, storagesvc.DirectAccessInput{
		Namespace: storagesvc.NamespaceFiles,
		Key:       record.Path,
		Operation: storagecap.DirectAccessOpGet,
		TTL:       ttl,
	})
	if err != nil {
		if bizerr.Is(err, storagecap.CodeStorageProviderConflict) {
			return nil, bizerr.WrapCode(err, CodeFileStorageConflict)
		}
		return &DirectDownloadOutput{
			Access:   &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpGet},
			ProxyURL: downloadProxyURL(record.Id),
		}, nil
	}
	access := directOut.Access
	if access == nil || storagecap.IsProxyDirectAccess(access) {
		return &DirectDownloadOutput{
			Access:   &storagecap.DirectAccess{Mode: storagecap.DirectAccessModeProxy, Operation: storagecap.DirectAccessOpGet},
			ProxyURL: downloadProxyURL(record.Id),
		}, nil
	}
	return &DirectDownloadOutput{Access: access}, nil
}

func downloadProxyURL(id int64) string {
	return fmt.Sprintf("/api/v1/file/%d/download", id)
}

func (s *serviceImpl) tryInstantReuseByHash(
	ctx context.Context,
	tenantID int64,
	userID int64,
	sanitizedFilename string,
	suffix string,
	scene string,
	size int64,
	fileHash string,
) (*UploadOutput, error) {
	var existing *entity.SysFile
	err := dao.SysFile.Ctx(ctx).
		Where(dao.SysFile.Columns().Hash, fileHash).
		Where(datascope.TenantColumn, tenantID).
		Scan(&existing)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileHashQueryFailed)
	}
	if existing == nil {
		return nil, nil
	}
	result, err := dao.SysFile.Ctx(ctx).Data(do.SysFile{
		TenantId:  tenantID,
		Name:      existing.Name,
		Original:  sanitizedFilename,
		Suffix:    suffix,
		Scene:     scene,
		Size:      size,
		Hash:      fileHash,
		Url:       existing.Url,
		Path:      existing.Path,
		Engine:    existing.Engine,
		CreatedBy: userID,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordSaveFailed)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordIDReadFailed)
	}
	return &UploadOutput{
		Id:       id,
		Name:     existing.Name,
		Original: sanitizedFilename,
		Url:      s.getBaseUrl(ctx) + existing.Url,
		Suffix:   suffix,
		Size:     size,
	}, nil
}

// resolveDirectUploadTTL loads the effective direct access TTL from system
// configuration (sys.upload.directUrlTTL, default 1h).
func (s *serviceImpl) resolveDirectUploadTTL(ctx context.Context) (time.Duration, error) {
	if s == nil || s.configSvc == nil {
		return 0, bizerr.NewCode(CodeFileDirectInitFailed)
	}
	return s.configSvc.GetUploadDirectUrlTTL(ctx)
}

func (s *serviceImpl) currentUserID(ctx context.Context) int64 {
	if s == nil || s.bizCtxSvc == nil {
		return 0
	}
	if bizCtx := s.bizCtxSvc.Get(ctx); bizCtx != nil {
		return int64(bizCtx.UserId)
	}
	return int64(s.bizCtxSvc.Current(ctx).UserID)
}

func (s *serviceImpl) sessionStore() *directUploadSessionStore {
	if s != nil && s.directSessions != nil {
		return s.directSessions
	}
	return defaultDirectUploadSessions()
}

func (s *serviceImpl) mapSessionError(err error) (*UploadOutput, error) {
	if err == nil {
		return nil, nil
	}
	if bizerr.Is(err, storagecap.CodeStorageDirectSessionExpired) {
		return nil, bizerr.NewCode(CodeFileDirectSessionExpired)
	}
	if bizerr.Is(err, storagecap.CodeStorageDirectSessionInvalid) {
		return nil, bizerr.NewCode(CodeFileDirectSessionInvalid)
	}
	return nil, err
}
