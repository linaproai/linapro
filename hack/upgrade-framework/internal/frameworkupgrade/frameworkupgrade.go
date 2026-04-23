// Package frameworkupgrade implements the development-only source-upgrade tool
// that lives under the repository-root hack workspace for LinaPro framework releases.
package frameworkupgrade

import (
	"context"
)

// Service defines the framework source-upgrade service contract.
type Service interface {
	// BuildPlan resolves the target framework release and calculates the upgrade work.
	BuildPlan(ctx context.Context, input BuildPlanInput) (*Plan, error)
	// ExecutePlan performs the planned code synchronization and host SQL replay.
	ExecutePlan(ctx context.Context, plan *Plan) (*ExecuteResult, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// BuildPlanInput describes one framework-upgrade planning request.
type BuildPlanInput struct {
	RepoURL   string // RepoURL overrides the upstream framework repository URL.
	TargetRef string // TargetRef overrides the target tag or git reference to upgrade to.
}

// Plan stores the resolved framework-upgrade execution plan.
type Plan struct {
	RepoRoot         string        // RepoRoot is the current project repository root.
	RepoURL          string        // RepoURL is the upstream framework repository URL.
	TargetRef        string        // TargetRef is the resolved target tag or git reference.
	TargetCloneDir   string        // TargetCloneDir is the temporary checkout of the target framework release.
	CurrentFramework UpgradeConfig // CurrentFramework stores upgrade metadata from the current project hack config.
	TargetFramework  UpgradeConfig // TargetFramework stores upgrade metadata from the target release hack config.
	CurrentVersion   string        // CurrentVersion is the current project framework version from hack/config.yaml.
	TargetVersion    string        // TargetVersion is the target framework version from hack/config.yaml.
	LatestSQLFile    string        // LatestSQLFile is the newest host SQL file present in the target release.
	SQLFiles         []string      // SQLFiles stores absolute target-clone SQL paths replayed during upgrade.
	UpgradeNeeded    bool          // UpgradeNeeded reports whether the target version is newer than the current version.
	DirtyGitFiles    []string      // DirtyGitFiles stores current git working-tree changes when precheck fails.
}

// ExecuteResult stores the final outcome of one framework-upgrade execution.
type ExecuteResult struct {
	TargetVersion    string   // TargetVersion is the applied framework version.
	ExecutedSQLFiles []string // ExecutedSQLFiles stores the SQL file basenames executed during this upgrade.
	UpgradeNeeded    bool     // UpgradeNeeded reports whether any upgrade work actually ran.
}

// New creates and returns a new framework-upgrade service instance.
func New() Service {
	return &serviceImpl{}
}
