// Package jobmgmt implements persistent scheduled-job CRUD, group management,
// log management, and cron-preview services.
package jobmgmt

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cluster"
	configsvc "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobhandler"
	"lina-core/internal/service/jobmeta"
	internalscheduler "lina-core/internal/service/jobmgmt/internal/scheduler"
	internalshellexec "lina-core/internal/service/jobmgmt/internal/shellexec"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/logger"
)

// GroupService defines the scheduled-job group management contract.
type GroupService interface {
	// ListGroups returns scheduled-job groups with pagination and job counts.
	ListGroups(ctx context.Context, in ListGroupsInput) (*ListGroupsOutput, error)
	// CreateGroup persists one new scheduled-job group.
	CreateGroup(ctx context.Context, in SaveGroupInput) (uint64, error)
	// UpdateGroup updates one existing scheduled-job group.
	UpdateGroup(ctx context.Context, in UpdateGroupInput) error
	// DeleteGroups removes one or more groups and migrates their jobs to the default group.
	DeleteGroups(ctx context.Context, ids string) error
}

// JobService defines the scheduled-job task management contract.
type JobService interface {
	// ListJobs returns scheduled jobs with pagination and group metadata.
	ListJobs(ctx context.Context, in ListJobsInput) (*ListJobsOutput, error)
	// GetJob returns one scheduled-job detail snapshot.
	GetJob(ctx context.Context, id uint64) (*JobDetailOutput, error)
	// CreateJob persists one new scheduled job and refreshes the scheduler when needed.
	CreateJob(ctx context.Context, in SaveJobInput) (uint64, error)
	// UpdateJob updates one scheduled job and refreshes the scheduler when needed.
	UpdateJob(ctx context.Context, in UpdateJobInput) error
	// DeleteJobs removes one or more non-built-in scheduled jobs.
	DeleteJobs(ctx context.Context, ids string) error
	// UpdateJobStatus toggles one job between enabled and disabled states.
	UpdateJobStatus(ctx context.Context, id uint64, status jobmeta.JobStatus) error
	// ResetJob resets executed_count and stop_reason for one scheduled job.
	ResetJob(ctx context.Context, id uint64) error
	// TriggerJob starts one manual execution and returns the created log ID.
	TriggerJob(ctx context.Context, id uint64) (uint64, error)
	// PreviewCron returns the next five fire times for one cron expression.
	PreviewCron(ctx context.Context, expr string, timezone string) ([]time.Time, error)
	// SyncBuiltinJobs upserts code-owned scheduled jobs into sys_job.
	SyncBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) error
	// ReconcileBuiltinJobs refreshes the full code-owned job projection and
	// prunes removed built-ins from sys_job.
	ReconcileBuiltinJobs(ctx context.Context, jobs []BuiltinJobDef) error
}

// LogService defines the scheduled-job execution log management contract.
type LogService interface {
	// ListLogs returns scheduled-job execution logs with pagination and job metadata.
	ListLogs(ctx context.Context, in ListLogsInput) (*ListLogsOutput, error)
	// GetLog returns one execution-log detail snapshot.
	GetLog(ctx context.Context, id uint64) (*LogDetailOutput, error)
	// ClearLogs deletes matching execution logs by selected IDs, job, or all rows.
	ClearLogs(ctx context.Context, jobID *uint64, ids string) error
	// CancelLog cancels one currently running execution instance.
	CancelLog(ctx context.Context, id uint64) error
	// CleanupDueLogs removes logs that exceed the effective retention policies.
	CleanupDueLogs(ctx context.Context) (int64, error)
}

// Service defines the complete scheduled-job management contract.
type Service interface {
	JobService
	LogService
	GroupService
}

// Scheduler defines the persistent scheduled-job runner contract exported to
// host wiring code while the concrete implementation stays internal.
type Scheduler interface {
	// LoadAndRegister registers all currently enabled persistent jobs at startup.
	LoadAndRegister(ctx context.Context) error
	// Refresh removes and re-registers one job according to its latest persisted state.
	Refresh(ctx context.Context, jobID uint64) error
	// Remove unregisters one persistent job from gcron.
	Remove(jobID uint64)
	// Trigger starts one manual execution and returns the created log ID.
	Trigger(ctx context.Context, jobID uint64) (uint64, error)
	// CancelLog cancels one currently running job-log instance.
	CancelLog(ctx context.Context, logID uint64) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// orderField identifies one public list sorting field accepted by jobmgmt APIs.
type orderField string

// Supported public order fields for scheduled-job management lists.
const (
	// orderFieldID sorts by the persisted numeric identifier.
	orderFieldID orderField = "id"
	// orderFieldName sorts by the display name.
	orderFieldName orderField = "name"
	// orderFieldGroupID sorts jobs by their owning group identifier.
	orderFieldGroupID orderField = "group_id"
	// orderFieldStatus sorts by the persisted status value.
	orderFieldStatus orderField = "status"
	// orderFieldTaskType sorts by the persisted job task type.
	orderFieldTaskType orderField = "task_type"
	// orderFieldSortOrder sorts groups by their configured display order.
	orderFieldSortOrder orderField = "sort_order"
	// orderFieldCode sorts groups by their stable code.
	orderFieldCode orderField = "code"
	// orderFieldCreatedAt sorts by creation time.
	orderFieldCreatedAt orderField = "created_at"
	// orderFieldUpdatedAt sorts by last update time.
	orderFieldUpdatedAt orderField = "updated_at"
	// orderFieldStartAt sorts logs by execution start time.
	orderFieldStartAt orderField = "start_at"
	// orderFieldEndAt sorts logs by execution end time.
	orderFieldEndAt orderField = "end_at"
	// orderFieldDurationMs sorts logs by execution duration.
	orderFieldDurationMs orderField = "duration_ms"
)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service        // bizCtxSvc resolves the current operator identity.
	configSvc configsvc.Service     // configSvc exposes runtime cron-management parameters.
	i18nSvc   jobmgmtI18nTranslator // i18nSvc localizes backend-owned display metadata.
	registry  jobhandler.Registry   // registry resolves handler definitions and validation schemas.
	scheduler Scheduler             // scheduler keeps persistent jobs registered with gcron.
}

// NewScheduler creates the persistent scheduler plus its internal shell
// executor so callers only depend on the jobmgmt component boundary.
func NewScheduler(
	clusterSvc cluster.Service,
	registry jobhandler.Registry,
	configSvc configsvc.Service,
) Scheduler {
	return internalscheduler.New(
		clusterSvc,
		registry,
		internalshellexec.New(configSvc),
	)
}

// New creates and returns one scheduled-job management service.
func New(
	configSvc configsvc.Service,
	registry jobhandler.Registry,
	scheduler Scheduler,
) Service {
	if configSvc == nil {
		configSvc = configsvc.New()
	}
	i18nSvc := i18nsvc.New()
	svc := &serviceImpl{
		bizCtxSvc: bizctx.New(),
		configSvc: configSvc,
		i18nSvc:   i18nSvc,
		registry:  registry,
		scheduler: scheduler,
	}
	if registry != nil {
		registry.SubscribeChanges(func(ref string, exists bool) {
			if err := svc.syncHandlerAvailability(context.Background(), ref, exists); err != nil {
				logger.Warningf(context.Background(), "sync scheduled job handler availability failed ref=%s exists=%t err=%v", ref, exists, err)
			}
		})
	}
	return svc
}

// SaveGroupInput stores mutable scheduled-job group fields.
type SaveGroupInput struct {
	Code      string // Code is the globally unique group code.
	Name      string // Name is the display name shown in the UI.
	Remark    string // Remark stores optional operator notes.
	SortOrder int    // SortOrder controls display ordering.
}

// UpdateGroupInput stores one group update request.
type UpdateGroupInput struct {
	ID uint64 // ID identifies the target group.
	SaveGroupInput
}

// ListGroupsInput stores group list filters and pagination.
type ListGroupsInput struct {
	PageNum        int    // PageNum is the 1-based page index.
	PageSize       int    // PageSize is the number of rows per page.
	Code           string // Code filters by group code.
	Name           string // Name filters by group name.
	OrderBy        string // OrderBy selects one supported sort field.
	OrderDirection string // OrderDirection selects asc or desc ordering.
}

// GroupListItem defines one group row returned to controllers.
type GroupListItem struct {
	*entity.SysJobGroup
	JobCount int64 // JobCount stores the number of jobs currently assigned to the group.
}

// ListGroupsOutput stores one paged group list response.
type ListGroupsOutput struct {
	List  []*GroupListItem // List stores paged group rows.
	Total int              // Total stores the total number of matching groups.
}

// SaveJobInput stores mutable scheduled-job fields.
type SaveJobInput struct {
	GroupID              uint64                   // GroupID identifies the owning group.
	Name                 string                   // Name is unique within the group.
	Description          string                   // Description explains the job purpose.
	TaskType             jobmeta.TaskType         // TaskType selects handler or shell execution.
	HandlerRef           string                   // HandlerRef selects the registered handler for handler jobs.
	Params               map[string]any           // Params stores handler parameters.
	Timeout              time.Duration            // Timeout bounds each execution.
	ShellCmd             string                   // ShellCmd stores the shell script for shell jobs.
	WorkDir              string                   // WorkDir stores the optional shell working directory.
	Env                  map[string]string        // Env stores shell environment overrides.
	CronExpr             string                   // CronExpr stores the cron expression.
	Timezone             string                   // Timezone stores the cron timezone identifier.
	Scope                jobmeta.JobScope         // Scope selects master-only or all-node execution.
	Concurrency          jobmeta.JobConcurrency   // Concurrency selects singleton or parallel execution.
	MaxConcurrency       int                      // MaxConcurrency caps parallel overlap per node.
	MaxExecutions        int                      // MaxExecutions caps cron-triggered runs.
	Status               jobmeta.JobStatus        // Status selects enabled or disabled persistence state.
	LogRetentionOverride *jobmeta.RetentionOption // LogRetentionOverride stores the optional per-job policy.
}

// BuiltinJobDef stores one code-owned scheduled-job definition projected into sys_job.
type BuiltinJobDef struct {
	GroupCode       string                 // GroupCode identifies the owning group by stable code.
	Name            string                 // Name is the human-readable job name.
	Description     string                 // Description explains the job purpose.
	TaskType        jobmeta.TaskType       // TaskType selects handler or shell execution.
	HandlerRef      string                 // HandlerRef selects the registered handler for handler jobs.
	Params          map[string]any         // Params stores the handler payload snapshot.
	Timeout         time.Duration          // Timeout bounds each execution.
	Pattern         string                 // Pattern stores the raw scheduler expression.
	Timezone        string                 // Timezone stores the display timezone for cron patterns.
	Scope           jobmeta.JobScope       // Scope selects master-only or all-node execution.
	Concurrency     jobmeta.JobConcurrency // Concurrency selects singleton or parallel execution.
	MaxConcurrency  int                    // MaxConcurrency caps parallel overlap.
	MaxExecutions   int                    // MaxExecutions caps cron-triggered runs.
	Status          jobmeta.JobStatus      // Status stores the desired steady-state status.
	LogRetentionRaw string                 // LogRetentionRaw stores the optional retention override JSON.
}

// UpdateJobInput stores one job update request.
type UpdateJobInput struct {
	ID uint64 // ID identifies the target job.
	SaveJobInput
}

// ListJobsInput stores job list filters and pagination.
type ListJobsInput struct {
	PageNum        int                    // PageNum is the 1-based page index.
	PageSize       int                    // PageSize is the number of rows per page.
	GroupID        *uint64                // GroupID filters by group ID.
	Status         jobmeta.JobStatus      // Status filters by job status.
	TaskType       jobmeta.TaskType       // TaskType filters by job type.
	Keyword        string                 // Keyword matches job name or description.
	Scope          jobmeta.JobScope       // Scope filters by job scope.
	Concurrency    jobmeta.JobConcurrency // Concurrency filters by concurrency policy.
	OrderBy        string                 // OrderBy selects one supported sort field.
	OrderDirection string                 // OrderDirection selects asc or desc ordering.
}

// JobListItem defines one job row returned to controllers.
type JobListItem struct {
	*entity.SysJob
	GroupCode string // GroupCode stores the owning group code.
	GroupName string // GroupName stores the owning group name.
}

// ListJobsOutput stores one paged job list response.
type ListJobsOutput struct {
	List  []*JobListItem // List stores paged job rows.
	Total int            // Total stores the total number of matching jobs.
}

// JobDetailOutput stores one job detail snapshot.
type JobDetailOutput struct {
	*entity.SysJob
	GroupCode string // GroupCode stores the owning group code.
	GroupName string // GroupName stores the owning group name.
}

// ListLogsInput stores log list filters and pagination.
type ListLogsInput struct {
	PageNum        int                 // PageNum is the 1-based page index.
	PageSize       int                 // PageSize is the number of rows per page.
	JobID          *uint64             // JobID filters by job identifier.
	Status         jobmeta.LogStatus   // Status filters by log status.
	Trigger        jobmeta.TriggerType // Trigger filters by trigger type.
	NodeID         string              // NodeID filters by execution node.
	BeginTime      string              // BeginTime filters by start_at lower bound.
	EndTime        string              // EndTime filters by start_at upper bound.
	OrderBy        string              // OrderBy selects one supported sort field.
	OrderDirection string              // OrderDirection selects asc or desc ordering.
}

// LogListItem defines one log row returned to controllers.
type LogListItem struct {
	*entity.SysJobLog
	JobName string // JobName stores the owning job name.
}

// ListLogsOutput stores one paged log list response.
type ListLogsOutput struct {
	List  []*LogListItem // List stores paged log rows.
	Total int            // Total stores the total number of matching logs.
}

// LogDetailOutput stores one execution-log detail snapshot.
type LogDetailOutput struct {
	*entity.SysJobLog
	JobName string // JobName stores the owning job name.
}

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

// parseUint64IDs parses one comma-separated identifier list.
func parseUint64IDs(ids string) []uint64 {
	parts := gstr.SplitAndTrim(ids, ",")
	result := make([]uint64, 0, len(parts))
	for _, part := range parts {
		currentID := gconv.Uint64(strings.TrimSpace(part))
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
		return nil, gerror.New("默认任务分组不存在")
	}
	return group, nil
}

// groupByID returns one job group by ID.
func (s *serviceImpl) groupByID(ctx context.Context, id uint64) (*entity.SysJobGroup, error) {
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
		return gerror.New("Shell 工作目录不能为根路径")
	}
	info, err := os.Stat(cleaned)
	if err != nil {
		return gerror.Wrap(err, "校验 Shell 工作目录失败")
	}
	if !info.IsDir() {
		return gerror.New("Shell 工作目录必须为目录")
	}
	return nil
}

// validateExecutableJob validates the runtime prerequisites for one persisted job definition.
func (s *serviceImpl) validateExecutableJob(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return gerror.New("定时任务不存在")
	}
	switch jobmeta.NormalizeTaskType(job.TaskType) {
	case jobmeta.TaskTypeHandler:
		def, ok := s.registry.Lookup(job.HandlerRef)
		if !ok {
			return gerror.New("任务处理器不存在")
		}
		return jobhandler.ValidateParams(def.ParamsSchema, json.RawMessage(job.Params))

	case jobmeta.TaskTypeShell:
		if !s.configSvc.IsCronShellEnabled(ctx) {
			return gerror.New("当前环境未启用 Shell 任务")
		}
		return validateWorkDir(job.WorkDir)
	}
	return gerror.New("任务类型不受支持")
}
