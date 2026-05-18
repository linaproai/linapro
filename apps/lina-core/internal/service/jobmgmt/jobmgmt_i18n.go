// This file localizes backend-owned scheduled-job display metadata before API
// responses leave the job management service.

package jobmgmt

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
)

const (
	// defaultGroupNameI18nKey localizes the code-owned default job-group name.
	defaultGroupNameI18nKey = "job.group.default.name"
	// defaultGroupRemarkI18nKey localizes the code-owned default job-group remark.
	defaultGroupRemarkI18nKey = "job.group.default.remark"
	// jobNameI18nField identifies the built-in job display-name i18n field.
	jobNameI18nField = "name"
	// jobDescriptionI18nField identifies the built-in job description i18n field.
	jobDescriptionI18nField = "description"
	// pluginHandlerRefPrefix identifies plugin-owned scheduled-job handler refs.
	pluginHandlerRefPrefix = "plugin:"
)

// handlerSourceTextCache stores request-local handler metadata translations.
type handlerSourceTextCache map[string]string

// jobmgmtI18nTranslator defines the narrow source-text translation capabilities jobmgmt needs.
type jobmgmtI18nTranslator interface {
	// TranslateSourceText returns one source-text-backed key with source text fallback.
	TranslateSourceText(ctx context.Context, key string, sourceText string) string
	// TranslateDynamicPluginSourceText returns one source-text-backed key from
	// a dynamic-plugin artifact, falling back to sourceText when unavailable.
	TranslateDynamicPluginSourceText(ctx context.Context, pluginID string, key string, sourceText string) string
}

// localizeGroupForDisplay translates the code-owned default group display
// fields while preserving custom group records exactly as stored.
func (s *serviceImpl) localizeGroupForDisplay(ctx context.Context, group *entity.SysJobGroup) {
	if group == nil || strings.TrimSpace(group.Code) != defaultBuiltinGroupCode {
		return
	}
	group.Name = s.translateSourceText(ctx, defaultGroupNameI18nKey, group.Name)
	group.Remark = s.translateSourceText(ctx, defaultGroupRemarkI18nKey, group.Remark)
}

// defaultGroupMatchesKeyword reports whether the localized default group
// metadata matches one list keyword.
func (s *serviceImpl) defaultGroupMatchesKeyword(ctx context.Context, keyword string) bool {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return false
	}
	name := s.translateSourceText(ctx, defaultGroupNameI18nKey, "")
	remark := s.translateSourceText(ctx, defaultGroupRemarkI18nKey, "")
	return strings.Contains(strings.ToLower(name), normalizedKeyword) ||
		strings.Contains(strings.ToLower(remark), normalizedKeyword)
}

// localizeBuiltinJobForDisplay translates code-owned job display fields while
// preserving operator-created jobs exactly as stored.
func (s *serviceImpl) localizeBuiltinJobForDisplay(ctx context.Context, job *entity.SysJob) {
	s.localizeBuiltinJobForDisplayWithCache(ctx, job, nil)
}

// localizeBuiltinJobForDisplayWithCache translates one built-in job using the
// caller's request-local translation cache when many jobs share handler refs.
func (s *serviceImpl) localizeBuiltinJobForDisplayWithCache(
	ctx context.Context,
	job *entity.SysJob,
	cache handlerSourceTextCache,
) {
	if job == nil || job.IsBuiltin != 1 {
		return
	}
	job.Name = s.localizeBuiltinJobNameWithCache(ctx, job.HandlerRef, job.Name, job.IsBuiltin, cache)
	job.Description = s.localizeBuiltinJobDescriptionWithCache(ctx, job.HandlerRef, job.Description, job.IsBuiltin, cache)
}

// localizeBuiltinJobName translates one built-in job name by handler ref.
func (s *serviceImpl) localizeBuiltinJobName(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
) string {
	return s.localizeBuiltinJobNameWithCache(ctx, handlerRef, fallback, isBuiltin, nil)
}

// localizeBuiltinJobNameWithCache translates one built-in job name using a
// request-local cache to avoid repeated dynamic-plugin artifact lookups.
func (s *serviceImpl) localizeBuiltinJobNameWithCache(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
	cache handlerSourceTextCache,
) string {
	if isBuiltin != 1 {
		return fallback
	}
	return s.translateHandlerSourceTextWithCache(ctx, handlerRef, jobNameI18nField, fallback, cache)
}

// localizeBuiltinJobDescription translates one built-in job description by handler ref.
func (s *serviceImpl) localizeBuiltinJobDescription(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
) string {
	return s.localizeBuiltinJobDescriptionWithCache(ctx, handlerRef, fallback, isBuiltin, nil)
}

// localizeBuiltinJobDescriptionWithCache translates one built-in job
// description using a request-local cache to avoid repeated metadata reads.
func (s *serviceImpl) localizeBuiltinJobDescriptionWithCache(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
	cache handlerSourceTextCache,
) string {
	if isBuiltin != 1 {
		return fallback
	}
	return s.translateHandlerSourceTextWithCache(ctx, handlerRef, jobDescriptionI18nField, fallback, cache)
}

// localizedHandlerRefsMatchingKeyword returns handler refs whose localized
// backend-owned display metadata matches a list keyword.
func (s *serviceImpl) localizedHandlerRefsMatchingKeyword(ctx context.Context, keyword string) ([]string, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return []string{}, nil
	}

	refSet := make(map[string]struct{})
	cache := make(handlerSourceTextCache)
	if s != nil && s.registry != nil {
		for _, handler := range s.registry.List() {
			displayName := s.translateHandlerSourceTextWithCache(ctx, handler.Ref, jobNameI18nField, handler.DisplayName, cache)
			description := s.translateHandlerSourceTextWithCache(ctx, handler.Ref, jobDescriptionI18nField, handler.Description, cache)
			if localizedTextMatchesKeyword(displayName, description, normalizedKeyword) {
				refSet[handler.Ref] = struct{}{}
			}
		}
	}

	var jobs []*entity.SysJob
	cols := dao.SysJob.Columns()
	err := dao.SysJob.Ctx(ctx).
		Fields(cols.Name, cols.Description, cols.HandlerRef, cols.IsBuiltin).
		Where(cols.IsBuiltin, 1).
		Scan(&jobs)
	if err != nil {
		return nil, err
	}
	for _, job := range jobs {
		if job == nil || strings.TrimSpace(job.HandlerRef) == "" {
			continue
		}
		displayName := s.localizeBuiltinJobNameWithCache(ctx, job.HandlerRef, job.Name, job.IsBuiltin, cache)
		description := s.localizeBuiltinJobDescriptionWithCache(ctx, job.HandlerRef, job.Description, job.IsBuiltin, cache)
		if localizedTextMatchesKeyword(displayName, description, normalizedKeyword) {
			refSet[job.HandlerRef] = struct{}{}
		}
	}

	refs := make([]string, 0, len(refSet))
	for ref := range refSet {
		refs = append(refs, ref)
	}
	return refs, nil
}

// localizedTextMatchesKeyword reports whether either localized text contains
// the already-normalized keyword.
func localizedTextMatchesKeyword(name string, description string, normalizedKeyword string) bool {
	return strings.Contains(strings.ToLower(name), normalizedKeyword) ||
		strings.Contains(strings.ToLower(description), normalizedKeyword)
}

// translateHandlerSourceText resolves handler-owned display metadata, including
// dynamic-plugin artifact-local resources before a plugin is enabled.
func (s *serviceImpl) translateHandlerSourceText(ctx context.Context, handlerRef string, field string, sourceText string) string {
	return s.translateHandlerSourceTextWithCache(ctx, handlerRef, field, sourceText, nil)
}

// translateHandlerSourceTextWithCache resolves handler metadata and reuses the
// result for duplicate handler-ref/field/fallback triples in one request.
func (s *serviceImpl) translateHandlerSourceTextWithCache(
	ctx context.Context,
	handlerRef string,
	field string,
	sourceText string,
	cache handlerSourceTextCache,
) string {
	if cache != nil {
		cacheKey := handlerSourceTextCacheKey(handlerRef, field, sourceText)
		if value, ok := cache[cacheKey]; ok {
			return value
		}
		translated := s.translateHandlerSourceTextWithCache(ctx, handlerRef, field, sourceText, nil)
		cache[cacheKey] = translated
		return translated
	}

	key := jobmeta.HandlerI18nKey(handlerRef, field)
	translated := s.translateSourceText(ctx, key, sourceText)
	if translated != sourceText {
		return translated
	}
	pluginID := pluginIDFromHandlerRef(handlerRef)
	if pluginID == "" || s == nil || s.i18nSvc == nil {
		return translated
	}
	return s.i18nSvc.TranslateDynamicPluginSourceText(ctx, pluginID, key, translated)
}

// handlerSourceTextCacheKey builds a stable key for request-local handler
// metadata translation caching.
func handlerSourceTextCacheKey(handlerRef string, field string, sourceText string) string {
	return strings.TrimSpace(handlerRef) + "\x00" + strings.TrimSpace(field) + "\x00" + sourceText
}

// translateSourceText resolves code-owned display metadata with source-text
// fallback so the selected locale never silently falls back to another language.
func (s *serviceImpl) translateSourceText(ctx context.Context, key string, sourceText string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return sourceText
	}
	return s.i18nSvc.TranslateSourceText(ctx, key, sourceText)
}

// pluginIDFromHandlerRef extracts the plugin identifier from a plugin-owned
// handler ref such as plugin:demo-plugin/cron:heartbeat.
func pluginIDFromHandlerRef(handlerRef string) string {
	trimmedRef := strings.TrimSpace(handlerRef)
	if !strings.HasPrefix(trimmedRef, pluginHandlerRefPrefix) {
		return ""
	}
	withoutPrefix := strings.TrimPrefix(trimmedRef, pluginHandlerRefPrefix)
	separatorIndex := strings.Index(withoutPrefix, "/")
	if separatorIndex <= 0 {
		return ""
	}
	return strings.TrimSpace(withoutPrefix[:separatorIndex])
}
