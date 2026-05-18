// This file contains shared job management helper methods for identity,
// ordering, group lookup, shell workdir validation, and executable checks.

package jobmgmt

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
)

// currentUserID returns the current operator ID or zero when unavailable.
func (s *serviceImpl) currentUserID(ctx context.Context) int64 {
	if s == nil {
		return 0
	}
	businessCtx := s.bizCtxSvc.Get(ctx)
	if businessCtx == nil || businessCtx.UserId <= 0 {
		return 0
	}
	return int64(businessCtx.UserId)
}

// parseInt64IDs parses one comma-separated identifier list.
func parseInt64IDs(ids string) []int64 {
	parts := gstr.SplitAndTrim(ids, ",")
	result := make([]int64, 0, len(parts))
	for _, part := range parts {
		currentID := gconv.Int64(strings.TrimSpace(part))
		if currentID == 0 {
			continue
		}
		result = append(result, currentID)
	}
	return result
}

// applySingleOrder applies one validated order field and direction to the model.
func applySingleOrder(
	model *gdb.Model,
	orderBy string,
	orderDirection string,
	allowed map[orderField]string,
	defaultField string,
	defaultDirection gdbutil.OrderDirection,
) *gdb.Model {
	if model == nil {
		return nil
	}
	field := allowed[orderField(strings.TrimSpace(orderBy))]
	if field == "" {
		field = defaultField
	}
	direction := gdbutil.NormalizeOrderDirectionOrDefault(orderDirection, defaultDirection)
	return gdbutil.ApplyModelOrder(model, field, direction)
}

// defaultGroup returns the current default scheduled-job group.
func (s *serviceImpl) defaultGroup(ctx context.Context) (*entity.SysJobGroup, error) {
	var group *entity.SysJobGroup
	err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{IsDefault: 1}).
		Scan(&group)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, bizerr.NewCode(CodeJobGroupDefaultNotFound)
	}
	return group, nil
}

// groupByID returns one job group by ID.
func (s *serviceImpl) groupByID(ctx context.Context, id int64) (*entity.SysJobGroup, error) {
	var group *entity.SysJobGroup
	err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{Id: id}).
		Scan(&group)
	return group, err
}

// validateWorkDir validates one optional shell working directory.
func validateWorkDir(workDir string) error {
	trimmed := strings.TrimSpace(workDir)
	if trimmed == "" {
		return nil
	}
	cleaned := filepath.Clean(trimmed)
	if cleaned == string(filepath.Separator) {
		return bizerr.NewCode(jobmeta.CodeJobShellWorkdirRootDenied)
	}
	info, err := os.Stat(cleaned)
	if err != nil {
		return bizerr.WrapCode(err, jobmeta.CodeJobShellWorkdirValidateFailed)
	}
	if !info.IsDir() {
		return bizerr.NewCode(jobmeta.CodeJobShellWorkdirNotDirectory)
	}
	return nil
}

// validateExecutableJob validates the runtime prerequisites for one persisted job definition.
func (s *serviceImpl) validateExecutableJob(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	switch jobmeta.NormalizeTaskType(job.TaskType) {
	case jobmeta.TaskTypeHandler:
		def, ok := s.registry.Lookup(job.HandlerRef)
		if !ok {
			return bizerr.NewCode(jobhandler.CodeJobHandlerNotFound)
		}
		return jobhandler.ValidateParams(def.ParamsSchema, json.RawMessage(job.Params))
	case jobmeta.TaskTypeShell:
		enabled, err := s.configSvc.IsCronShellEnabled(ctx)
		if err != nil {
			return err
		}
		if !enabled {
			return bizerr.NewCode(jobmeta.CodeJobShellDisabled)
		}
		return validateWorkDir(job.WorkDir)
	}
	return bizerr.NewCode(jobmeta.CodeJobTaskTypeUnsupported)
}
