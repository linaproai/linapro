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
)

// jobmgmtI18nTranslator defines the narrow source-text translation capability jobmgmt needs.
type jobmgmtI18nTranslator interface {
	// TranslateSourceText returns one source-text-backed key with source text fallback.
	TranslateSourceText(ctx context.Context, key string, sourceText string) string
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
	if job == nil || job.IsBuiltin != 1 {
		return
	}
	job.Name = s.localizeBuiltinJobName(ctx, job.HandlerRef, job.Name, job.IsBuiltin)
	job.Description = s.localizeBuiltinJobDescription(ctx, job.HandlerRef, job.Description, job.IsBuiltin)
}

// localizeBuiltinJobName translates one built-in job name by handler ref.
func (s *serviceImpl) localizeBuiltinJobName(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
) string {
	if isBuiltin != 1 {
		return fallback
	}
	return s.translateSourceText(ctx, jobmeta.HandlerI18nKey(handlerRef, jobNameI18nField), fallback)
}

// localizeBuiltinJobDescription translates one built-in job description by handler ref.
func (s *serviceImpl) localizeBuiltinJobDescription(
	ctx context.Context,
	handlerRef string,
	fallback string,
	isBuiltin int,
) string {
	if isBuiltin != 1 {
		return fallback
	}
	return s.translateSourceText(ctx, jobmeta.HandlerI18nKey(handlerRef, jobDescriptionI18nField), fallback)
}

// localizedHandlerRefsMatchingKeyword returns handler refs whose localized
// backend-owned display metadata matches a list keyword.
func (s *serviceImpl) localizedHandlerRefsMatchingKeyword(ctx context.Context, keyword string) ([]string, error) {
	normalizedKeyword := strings.ToLower(strings.TrimSpace(keyword))
	if normalizedKeyword == "" {
		return []string{}, nil
	}

	refSet := make(map[string]struct{})
	if s != nil && s.registry != nil {
		for _, handler := range s.registry.List() {
			displayName := s.translateSourceText(ctx, jobmeta.HandlerI18nKey(handler.Ref, jobNameI18nField), handler.DisplayName)
			description := s.translateSourceText(ctx, jobmeta.HandlerI18nKey(handler.Ref, jobDescriptionI18nField), handler.Description)
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
		displayName := s.localizeBuiltinJobName(ctx, job.HandlerRef, job.Name, job.IsBuiltin)
		description := s.localizeBuiltinJobDescription(ctx, job.HandlerRef, job.Description, job.IsBuiltin)
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

// translateSourceText resolves code-owned display metadata with source-text
// fallback so the selected locale never silently falls back to another language.
func (s *serviceImpl) translateSourceText(ctx context.Context, key string, sourceText string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return sourceText
	}
	return s.i18nSvc.TranslateSourceText(ctx, key, sourceText)
}
