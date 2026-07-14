// This file provides registration-only capability stubs for plugin service
// tests that load real source plugins without executing their business paths.

package testutil

import (
	"context"

	jobv1 "lina-core/api/job/v1"
	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap/extlogin"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/routecap"
	capabilitysessioncap "lina-core/pkg/plugin/capability/sessioncap"
)

type testNoopAPIDoc struct{}

func (testNoopAPIDoc) ResolveRouteText(_ context.Context, input apidoccap.RouteTextInput) apidoccap.RouteTextOutput {
	return apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
}

func (testNoopAPIDoc) ResolveRouteTexts(_ context.Context, inputs []apidoccap.RouteTextInput) []apidoccap.RouteTextOutput {
	outputs := make([]apidoccap.RouteTextOutput, 0, len(inputs))
	for _, input := range inputs {
		outputs = append(outputs, apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
	}
	return outputs
}

func (testNoopAPIDoc) FindRouteTitleOperationKeys(context.Context, string) []string {
	return nil
}

type testNoopAuth struct{}

func (testNoopAuth) SelectTenant(context.Context, token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

func (testNoopAuth) SwitchTenant(context.Context, token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	return &token.TenantTokenOutput{}, nil
}

func (testNoopAuth) IssueImpersonationToken(context.Context, token.ImpersonationTokenIssueInput) (*token.ImpersonationTokenOutput, error) {
	return &token.ImpersonationTokenOutput{}, nil
}

func (testNoopAuth) RevokeImpersonationToken(context.Context, token.ImpersonationTokenRevokeInput) error {
	return nil
}

// testNoopExternalLogin is a registration-safe external-login stub so LDAP/OIDC
// plugins can wire routes without a full host external-login implementation.
type testNoopExternalLogin struct{}

func (testNoopExternalLogin) LoginByVerifiedIdentity(
	context.Context,
	extlogin.LoginInput,
) (*extlogin.LoginOutput, error) {
	return &extlogin.LoginOutput{}, nil
}

type testNoopI18n struct{}

func (testNoopI18n) GetLocale(context.Context) string {
	return "zh-CN"
}

func (testNoopI18n) Translate(_ context.Context, _ string, fallback string) string {
	return fallback
}

type testNoopPluginLifecycle struct{}

func (testNoopPluginLifecycle) EnsureTenantPluginDisableAllowed(context.Context, string, int) error {
	return nil
}

func (testNoopPluginLifecycle) NotifyTenantPluginDisabled(context.Context, string, int) {}

func (testNoopPluginLifecycle) EnsureTenantDeleteAllowed(context.Context, int) error {
	return nil
}

func (testNoopPluginLifecycle) NotifyTenantDeleted(context.Context, int) {}

type testNoopRoute struct{}

func (testNoopRoute) GetMetadata(context.Context) *routecap.Metadata {
	return nil
}

type testNoopFiles struct{}

func (testNoopFiles) BatchGet(_ context.Context, ids []capabilityfilecap.FileID) (*capmodel.BatchResult[*capabilityfilecap.FileInfo, capabilityfilecap.FileID], error) {
	return &capmodel.BatchResult[*capabilityfilecap.FileInfo, capabilityfilecap.FileID]{
		Items:      map[capabilityfilecap.FileID]*capabilityfilecap.FileInfo{},
		MissingIDs: append([]capabilityfilecap.FileID(nil), ids...),
	}, nil
}

func (s testNoopFiles) Get(ctx context.Context, id capabilityfilecap.FileID) (*capabilityfilecap.FileInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityfilecap.FileID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

func (testNoopFiles) Detail(context.Context, capabilityfilecap.FileID) (*capabilityfilecap.DetailInfo, error) {
	return nil, nil
}

func (testNoopFiles) List(context.Context, capabilityfilecap.ListInput) (*capmodel.PageResult[*capabilityfilecap.FileInfo], error) {
	return &capmodel.PageResult[*capabilityfilecap.FileInfo]{Items: []*capabilityfilecap.FileInfo{}}, nil
}

func (testNoopFiles) ListScenes(context.Context) ([]*capabilityfilecap.Option, error) {
	return []*capabilityfilecap.Option{}, nil
}

func (testNoopFiles) ListSuffixes(context.Context) ([]*capabilityfilecap.Option, error) {
	return []*capabilityfilecap.Option{}, nil
}

func (testNoopFiles) Open(context.Context, capabilityfilecap.FileID) (*capabilityfilecap.FileContent, error) {
	return nil, nil
}

func (testNoopFiles) EnsureVisible(context.Context, []capabilityfilecap.FileID) error {
	return nil
}

func (testNoopFiles) Upload(context.Context, capabilityfilecap.UploadInput) (*capabilityfilecap.FileInfo, error) {
	return nil, nil
}

func (testNoopFiles) CreateFromStorage(context.Context, capabilityfilecap.CreateFromStorageInput) (*capabilityfilecap.FileInfo, error) {
	return nil, nil
}

func (testNoopFiles) UpdateMetadata(context.Context, capabilityfilecap.UpdateMetadataInput) error {
	return nil
}

func (testNoopFiles) Delete(context.Context, capabilityfilecap.FileID) error {
	return nil
}

func (testNoopFiles) DeleteMany(context.Context, []capabilityfilecap.FileID) error {
	return nil
}

type testNoopSysConfig struct{}

func (s testNoopSysConfig) Get(ctx context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	result, err := s.BatchGet(ctx, []hostconfigcap.SysConfigKey{key})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[key], nil
}

func (testNoopSysConfig) BatchGet(_ context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	return &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items:      map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
		MissingIDs: append([]hostconfigcap.SysConfigKey(nil), keys...),
	}, nil
}

func (testNoopSysConfig) List(context.Context, hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return &capmodel.PageResult[*hostconfigcap.SysConfigInfo]{Items: []*hostconfigcap.SysConfigInfo{}}, nil
}

func (testNoopSysConfig) SetValue(context.Context, hostconfigcap.SysConfigKey, string, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

func (testNoopSysConfig) BatchSetValue(context.Context, []hostconfigcap.SetSysConfigValueItem, *hostconfigcap.SetSysConfigValueOptions) error {
	return nil
}

func (testNoopSysConfig) Reset(context.Context, hostconfigcap.SysConfigKey) error {
	return nil
}

func (testNoopSysConfig) EnsureVisible(context.Context, []hostconfigcap.SysConfigKey) error {
	return nil
}

type testNoopNotifications struct{}

func (testNoopNotifications) BatchGet(_ context.Context, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID], error) {
	return &capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID]{
		Items:      map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageInfo{},
		MissingIDs: append([]capabilitynotifycap.MessageID(nil), ids...),
	}, nil
}

func (s testNoopNotifications) Get(ctx context.Context, id capabilitynotifycap.MessageID) (*capabilitynotifycap.MessageInfo, error) {
	result, err := s.BatchGet(ctx, []capabilitynotifycap.MessageID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

func (testNoopNotifications) List(context.Context, capabilitynotifycap.ListInput) (*capmodel.PageResult[*capabilitynotifycap.MessageInfo], error) {
	return &capmodel.PageResult[*capabilitynotifycap.MessageInfo]{Items: []*capabilitynotifycap.MessageInfo{}}, nil
}

func (testNoopNotifications) BatchGetBySource(_ context.Context, input capabilitynotifycap.BatchGetBySourceInput) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	return &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageInfo{},
		MissingIDs: append([]string(nil), input.SourceIDs...),
	}, nil
}

func (testNoopNotifications) EnsureVisible(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

func (testNoopNotifications) Send(context.Context, capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	return &capabilitynotifycap.SendResult{}, nil
}

func (testNoopNotifications) Delete(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

func (testNoopNotifications) DeleteBySource(context.Context, usermsgv1.SourceType, []string) error {
	return nil
}

func (testNoopNotifications) MarkRead(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

func (testNoopNotifications) MarkUnread(context.Context, []capabilitynotifycap.MessageID) error {
	return nil
}

type testNoopSessions struct{}

func (testNoopSessions) Current(context.Context) (*capabilitysessioncap.SessionInfo, error) {
	return nil, nil
}

func (s testNoopSessions) Get(ctx context.Context, id capabilitysessioncap.SessionID) (*capabilitysessioncap.SessionInfo, error) {
	result, err := s.BatchGet(ctx, []capabilitysessioncap.SessionID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

func (testNoopSessions) List(context.Context, capabilitysessioncap.ListInput) (*capmodel.PageResult[*capabilitysessioncap.SessionInfo], error) {
	return &capmodel.PageResult[*capabilitysessioncap.SessionInfo]{Items: []*capabilitysessioncap.SessionInfo{}}, nil
}

func (testNoopSessions) BatchGet(_ context.Context, ids []capabilitysessioncap.SessionID) (*capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.SessionInfo, capabilitysessioncap.SessionID]{
		Items:      map[capabilitysessioncap.SessionID]*capabilitysessioncap.SessionInfo{},
		MissingIDs: append([]capabilitysessioncap.SessionID(nil), ids...),
	}, nil
}

func (testNoopSessions) BatchGetUserOnlineStatus(_ context.Context, userIDs []string) (*capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string], error) {
	return &capmodel.BatchResult[*capabilitysessioncap.UserOnlineStatus, string]{
		Items:      map[string]*capabilitysessioncap.UserOnlineStatus{},
		MissingIDs: append([]string(nil), userIDs...),
	}, nil
}

func (testNoopSessions) EnsureVisible(context.Context, []capabilitysessioncap.SessionID) error {
	return nil
}

func (testNoopSessions) Revoke(context.Context, capabilitysessioncap.SessionID) error {
	return nil
}

func (testNoopSessions) RevokeMany(context.Context, []capabilitysessioncap.SessionID) error {
	return nil
}

type testNoopJobs struct{}

func (testNoopJobs) BatchGet(_ context.Context, ids []capabilityjobcap.JobID) (*capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID], error) {
	return &capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID]{
		Items:      map[capabilityjobcap.JobID]*capabilityjobcap.JobInfo{},
		MissingIDs: append([]capabilityjobcap.JobID(nil), ids...),
	}, nil
}

func (s testNoopJobs) Get(ctx context.Context, id capabilityjobcap.JobID) (*capabilityjobcap.JobInfo, error) {
	result, err := s.BatchGet(ctx, []capabilityjobcap.JobID{id})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[id], nil
}

func (testNoopJobs) List(context.Context, capabilityjobcap.ListInput) (*capmodel.PageResult[*capabilityjobcap.JobInfo], error) {
	return &capmodel.PageResult[*capabilityjobcap.JobInfo]{Items: []*capabilityjobcap.JobInfo{}}, nil
}

func (testNoopJobs) EnsureVisible(context.Context, []capabilityjobcap.JobID) error {
	return nil
}

func (testNoopJobs) Create(context.Context, capabilityjobcap.SaveInput) (capabilityjobcap.JobID, error) {
	return "", nil
}

func (testNoopJobs) Update(context.Context, capabilityjobcap.UpdateInput) error {
	return nil
}

func (testNoopJobs) Delete(context.Context, capabilityjobcap.JobID) error {
	return nil
}

func (testNoopJobs) Run(context.Context, capabilityjobcap.JobID) error {
	return nil
}

func (testNoopJobs) SetStatus(context.Context, capabilityjobcap.JobID, jobv1.Status) error {
	return nil
}
