// This file implements independent media strategy binding resources and effective strategy resolution.

package media

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	entitymodel "lina-plugin-media/backend/internal/model/entity"
)

// ListBindingsInput defines media binding list filters.
type ListBindingsInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches tenant ID or device ID.
}

// ListBindingsOutput defines paged media strategy bindings.
type ListBindingsOutput struct {
	List  []*BindingOutput // List contains current page bindings.
	Total int              // Total is total matched row count.
}

// BindingOutput defines one media strategy binding row.
type BindingOutput struct {
	RowKey       string // RowKey is the stable frontend table key.
	TenantId     string // TenantId is the media tenant ID.
	DeviceId     string // DeviceId is the GB device ID.
	StrategyId   int64  // StrategyId is the linked media strategy ID.
	StrategyName string // StrategyName is the linked media strategy name.
}

// DeviceBindingMutationInput defines device binding save input.
type DeviceBindingMutationInput struct {
	DeviceId   string // DeviceId is the GB device ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// DeviceBindingMutationOutput defines device binding mutation result.
type DeviceBindingMutationOutput struct {
	DeviceId   string // DeviceId is the GB device ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// TenantBindingMutationInput defines tenant binding save input.
type TenantBindingMutationInput struct {
	TenantId   string // TenantId is the media tenant ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// TenantBindingMutationOutput defines tenant binding mutation result.
type TenantBindingMutationOutput struct {
	TenantId   string // TenantId is the media tenant ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// TenantDeviceBindingMutationInput defines tenant-device binding save input.
type TenantDeviceBindingMutationInput struct {
	TenantId   string // TenantId is the media tenant ID.
	DeviceId   string // DeviceId is the GB device ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// TenantDeviceBindingMutationOutput defines tenant-device binding mutation result.
type TenantDeviceBindingMutationOutput struct {
	TenantId   string // TenantId is the media tenant ID.
	DeviceId   string // DeviceId is the GB device ID.
	StrategyId int64  // StrategyId is the linked media strategy ID.
}

// ResolveStrategyInput defines effective strategy resolution input.
type ResolveStrategyInput struct {
	TenantId string // TenantId is the media tenant ID.
	DeviceId string // DeviceId is the GB device ID.
}

// ResolveStrategyOutput defines effective strategy resolution output.
type ResolveStrategyOutput struct {
	Matched      bool   // Matched reports whether a strategy matched.
	Source       string // Source is the matching strategy source.
	SourceLabel  string // SourceLabel is the Chinese source label.
	StrategyId   int64  // StrategyId is the matched strategy ID.
	StrategyName string // StrategyName is the matched strategy name.
	Strategy     string // Strategy is the matched YAML strategy body.
}

// Row-key prefixes used by frontend tables.
const (
	rowKeyPrefixDevice       = "device:"
	rowKeyPrefixTenant       = "tenant:"
	rowKeyPrefixTenantDevice = "tenantDevice:"
)

// binding strategy entity aliases.
type (
	deviceBindingEntity       = entitymodel.MediaStrategyDevice
	tenantBindingEntity       = entitymodel.MediaStrategyTenant
	tenantDeviceBindingEntity = entitymodel.MediaStrategyDeviceTenant
)

// ListDeviceBindings returns paged device strategy bindings.
func (s *serviceImpl) ListDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	return s.listDeviceBindings(ctx, in)
}

// SaveDeviceBinding creates or updates one device strategy binding.
func (s *serviceImpl) SaveDeviceBinding(ctx context.Context, in DeviceBindingMutationInput) (*DeviceBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	deviceID, err := normalizeDeviceID(in.DeviceId)
	if err != nil {
		return nil, err
	}
	if _, err = s.getStrategyEntity(ctx, in.StrategyId); err != nil {
		return nil, err
	}

	err = dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.deleteDeviceBindingRecord(ctx, deviceID); err != nil {
			return err
		}
		if err := s.insertDeviceBindingRecord(ctx, DeviceBindingMutationInput{
			DeviceId:   deviceID,
			StrategyId: in.StrategyId,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &DeviceBindingMutationOutput{DeviceId: deviceID, StrategyId: in.StrategyId}, nil
}

// DeleteDeviceBinding deletes one device strategy binding.
func (s *serviceImpl) DeleteDeviceBinding(ctx context.Context, deviceID string) (*DeviceBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedDeviceID, err := normalizeDeviceID(deviceID)
	if err != nil {
		return nil, err
	}
	if err := s.deleteDeviceBindingRecord(ctx, normalizedDeviceID); err != nil {
		return nil, err
	}
	return &DeviceBindingMutationOutput{DeviceId: normalizedDeviceID}, nil
}

// ListTenantBindings returns paged tenant strategy bindings.
func (s *serviceImpl) ListTenantBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	return s.listTenantBindings(ctx, in)
}

// SaveTenantBinding creates or updates one tenant strategy binding.
func (s *serviceImpl) SaveTenantBinding(ctx context.Context, in TenantBindingMutationInput) (*TenantBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	tenantID, err := normalizeTenantID(in.TenantId)
	if err != nil {
		return nil, err
	}
	if _, err = s.getStrategyEntity(ctx, in.StrategyId); err != nil {
		return nil, err
	}

	err = dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.deleteTenantBindingRecord(ctx, tenantID); err != nil {
			return err
		}
		if err := s.insertTenantBindingRecord(ctx, TenantBindingMutationInput{
			TenantId:   tenantID,
			StrategyId: in.StrategyId,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &TenantBindingMutationOutput{TenantId: tenantID, StrategyId: in.StrategyId}, nil
}

// DeleteTenantBinding deletes one tenant strategy binding.
func (s *serviceImpl) DeleteTenantBinding(ctx context.Context, tenantID string) (*TenantBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedTenantID, err := normalizeTenantID(tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.deleteTenantBindingRecord(ctx, normalizedTenantID); err != nil {
		return nil, err
	}
	return &TenantBindingMutationOutput{TenantId: normalizedTenantID}, nil
}

// ListTenantDeviceBindings returns paged tenant-device strategy bindings.
func (s *serviceImpl) ListTenantDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	return s.listTenantDeviceBindings(ctx, in)
}

// SaveTenantDeviceBinding creates or updates one tenant-device strategy binding.
func (s *serviceImpl) SaveTenantDeviceBinding(ctx context.Context, in TenantDeviceBindingMutationInput) (*TenantDeviceBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	tenantID, err := normalizeTenantID(in.TenantId)
	if err != nil {
		return nil, err
	}
	deviceID, err := normalizeDeviceID(in.DeviceId)
	if err != nil {
		return nil, err
	}
	if _, err = s.getStrategyEntity(ctx, in.StrategyId); err != nil {
		return nil, err
	}

	err = dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.deleteTenantDeviceBindingRecord(ctx, tenantID, deviceID); err != nil {
			return err
		}
		if err := s.insertTenantDeviceBindingRecord(ctx, TenantDeviceBindingMutationInput{
			TenantId:   tenantID,
			DeviceId:   deviceID,
			StrategyId: in.StrategyId,
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &TenantDeviceBindingMutationOutput{
		TenantId:   tenantID,
		DeviceId:   deviceID,
		StrategyId: in.StrategyId,
	}, nil
}

// DeleteTenantDeviceBinding deletes one tenant-device strategy binding.
func (s *serviceImpl) DeleteTenantDeviceBinding(ctx context.Context, tenantID string, deviceID string) (*TenantDeviceBindingMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedTenantID, err := normalizeTenantID(tenantID)
	if err != nil {
		return nil, err
	}
	normalizedDeviceID, err := normalizeDeviceID(deviceID)
	if err != nil {
		return nil, err
	}
	if err := s.deleteTenantDeviceBindingRecord(ctx, normalizedTenantID, normalizedDeviceID); err != nil {
		return nil, err
	}
	return &TenantDeviceBindingMutationOutput{TenantId: normalizedTenantID, DeviceId: normalizedDeviceID}, nil
}

// ResolveStrategy resolves the effective strategy for one tenant/device pair.
func (s *serviceImpl) ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	tenantID := strings.TrimSpace(in.TenantId)
	deviceID := strings.TrimSpace(in.DeviceId)

	if tenantID != "" && deviceID != "" {
		strategy, err := s.strategyFromTenantDeviceBinding(ctx, tenantID, deviceID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceTenantDevice, strategy), nil
		}
	}
	if deviceID != "" {
		strategy, err := s.strategyFromDeviceBinding(ctx, deviceID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceDevice, strategy), nil
		}
	}
	if tenantID != "" {
		strategy, err := s.strategyFromTenantBinding(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceTenant, strategy), nil
		}
	}
	strategy, err := s.globalStrategy(ctx)
	if err != nil {
		return nil, err
	}
	if strategy != nil {
		return buildResolveOutput(StrategySourceGlobal, strategy), nil
	}
	return buildResolveOutput(StrategySourceNone, nil), nil
}

// listDeviceBindings returns paged device bindings.
func (s *serviceImpl) listDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaStrategyDevice.Columns()
	model := dao.MediaStrategyDevice.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		model = model.WhereLike(columns.DeviceId, "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	rows := make([]*deviceBindingEntity, 0)
	err = model.OrderAsc(columns.DeviceId).Page(pageNum, pageSize).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	strategyNames, err := s.strategyNameMap(ctx, collectDeviceBindingStrategyIDs(rows))
	if err != nil {
		return nil, err
	}
	list := make([]*BindingOutput, 0, len(rows))
	for _, row := range rows {
		list = append(list, &BindingOutput{
			RowKey:       rowKeyPrefixDevice + row.DeviceId,
			DeviceId:     row.DeviceId,
			StrategyId:   row.StrategyId,
			StrategyName: strategyNames[row.StrategyId],
		})
	}
	return &ListBindingsOutput{List: list, Total: total}, nil
}

// listTenantBindings returns paged tenant bindings.
func (s *serviceImpl) listTenantBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaStrategyTenant.Columns()
	model := dao.MediaStrategyTenant.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		model = model.WhereLike(columns.TenantId, "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	rows := make([]*tenantBindingEntity, 0)
	err = model.OrderAsc(columns.TenantId).Page(pageNum, pageSize).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	strategyNames, err := s.strategyNameMap(ctx, collectTenantBindingStrategyIDs(rows))
	if err != nil {
		return nil, err
	}
	list := make([]*BindingOutput, 0, len(rows))
	for _, row := range rows {
		list = append(list, &BindingOutput{
			RowKey:       rowKeyPrefixTenant + row.TenantId,
			TenantId:     row.TenantId,
			StrategyId:   row.StrategyId,
			StrategyName: strategyNames[row.StrategyId],
		})
	}
	return &ListBindingsOutput{List: list, Total: total}, nil
}

// listTenantDeviceBindings returns paged tenant-device bindings.
func (s *serviceImpl) listTenantDeviceBindings(ctx context.Context, in ListBindingsInput) (*ListBindingsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaStrategyDeviceTenant.Columns()
	model := dao.MediaStrategyDeviceTenant.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.TenantId+" LIKE ? OR "+columns.DeviceId+" LIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	rows := make([]*tenantDeviceBindingEntity, 0)
	err = model.OrderAsc(columns.TenantId).OrderAsc(columns.DeviceId).Page(pageNum, pageSize).Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	strategyNames, err := s.strategyNameMap(ctx, collectTenantDeviceBindingStrategyIDs(rows))
	if err != nil {
		return nil, err
	}
	list := make([]*BindingOutput, 0, len(rows))
	for _, row := range rows {
		list = append(list, &BindingOutput{
			RowKey:       rowKeyPrefixTenantDevice + row.TenantId + ":" + row.DeviceId,
			TenantId:     row.TenantId,
			DeviceId:     row.DeviceId,
			StrategyId:   row.StrategyId,
			StrategyName: strategyNames[row.StrategyId],
		})
	}
	return &ListBindingsOutput{List: list, Total: total}, nil
}

// insertDeviceBindingRecord inserts one device binding record.
func (s *serviceImpl) insertDeviceBindingRecord(ctx context.Context, in DeviceBindingMutationInput) error {
	_, err := dao.MediaStrategyDevice.Ctx(ctx).Data(do.MediaStrategyDevice{
		DeviceId:   in.DeviceId,
		StrategyId: in.StrategyId,
	}).Insert()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingSaveFailed)
	}
	return nil
}

// insertTenantBindingRecord inserts one tenant binding record.
func (s *serviceImpl) insertTenantBindingRecord(ctx context.Context, in TenantBindingMutationInput) error {
	_, err := dao.MediaStrategyTenant.Ctx(ctx).Data(do.MediaStrategyTenant{
		TenantId:   in.TenantId,
		StrategyId: in.StrategyId,
	}).Insert()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingSaveFailed)
	}
	return nil
}

// insertTenantDeviceBindingRecord inserts one tenant-device binding record.
func (s *serviceImpl) insertTenantDeviceBindingRecord(ctx context.Context, in TenantDeviceBindingMutationInput) error {
	_, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).Data(do.MediaStrategyDeviceTenant{
		TenantId:   in.TenantId,
		DeviceId:   in.DeviceId,
		StrategyId: in.StrategyId,
	}).Insert()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingSaveFailed)
	}
	return nil
}

// deleteDeviceBindingRecord deletes one device binding record by natural key.
func (s *serviceImpl) deleteDeviceBindingRecord(ctx context.Context, deviceID string) error {
	_, err := dao.MediaStrategyDevice.Ctx(ctx).
		Where(do.MediaStrategyDevice{DeviceId: deviceID}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingDeleteFailed)
	}
	return nil
}

// deleteTenantBindingRecord deletes one tenant binding record by natural key.
func (s *serviceImpl) deleteTenantBindingRecord(ctx context.Context, tenantID string) error {
	_, err := dao.MediaStrategyTenant.Ctx(ctx).
		Where(do.MediaStrategyTenant{TenantId: tenantID}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingDeleteFailed)
	}
	return nil
}

// deleteTenantDeviceBindingRecord deletes one tenant-device binding record by natural key.
func (s *serviceImpl) deleteTenantDeviceBindingRecord(ctx context.Context, tenantID string, deviceID string) error {
	_, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).
		Where(do.MediaStrategyDeviceTenant{TenantId: tenantID, DeviceId: deviceID}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaBindingDeleteFailed)
	}
	return nil
}

// strategyFromTenantDeviceBinding returns the enabled strategy bound to a tenant-device pair.
func (s *serviceImpl) strategyFromTenantDeviceBinding(ctx context.Context, tenantID string, deviceID string) (*strategyEntity, error) {
	var binding *tenantDeviceBindingEntity
	err := dao.MediaStrategyDeviceTenant.Ctx(ctx).
		Where(do.MediaStrategyDeviceTenant{TenantId: tenantID, DeviceId: deviceID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return s.enabledStrategyByID(ctx, binding.StrategyId)
}

// strategyFromDeviceBinding returns the enabled strategy bound to a device.
func (s *serviceImpl) strategyFromDeviceBinding(ctx context.Context, deviceID string) (*strategyEntity, error) {
	var binding *deviceBindingEntity
	err := dao.MediaStrategyDevice.Ctx(ctx).
		Where(do.MediaStrategyDevice{DeviceId: deviceID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return s.enabledStrategyByID(ctx, binding.StrategyId)
}

// strategyFromTenantBinding returns the enabled strategy bound to a tenant.
func (s *serviceImpl) strategyFromTenantBinding(ctx context.Context, tenantID string) (*strategyEntity, error) {
	var binding *tenantBindingEntity
	err := dao.MediaStrategyTenant.Ctx(ctx).
		Where(do.MediaStrategyTenant{TenantId: tenantID}).
		Scan(&binding)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaBindingListQueryFailed)
	}
	if binding == nil {
		return nil, nil
	}
	return s.enabledStrategyByID(ctx, binding.StrategyId)
}

// globalStrategy returns the enabled global strategy.
func (s *serviceImpl) globalStrategy(ctx context.Context) (*strategyEntity, error) {
	var strategy *strategyEntity
	err := dao.MediaStrategy.Ctx(ctx).
		Where(dao.MediaStrategy.Columns().Global, int(SwitchOn)).
		Where(dao.MediaStrategy.Columns().Enable, int(SwitchOn)).
		Scan(&strategy)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyDetailQueryFailed)
	}
	return strategy, nil
}

// enabledStrategyByID returns one enabled strategy by ID.
func (s *serviceImpl) enabledStrategyByID(ctx context.Context, id int64) (*strategyEntity, error) {
	var strategy *strategyEntity
	err := dao.MediaStrategy.Ctx(ctx).
		Where(dao.MediaStrategy.Columns().Enable, int(SwitchOn)).
		Where(do.MediaStrategy{Id: id}).
		Scan(&strategy)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyDetailQueryFailed)
	}
	return strategy, nil
}

// strategyNameMap returns strategy names by ID for list rendering.
func (s *serviceImpl) strategyNameMap(ctx context.Context, ids []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(ids) == 0 {
		return result, nil
	}
	strategies := make([]*strategyEntity, 0)
	err := dao.MediaStrategy.Ctx(ctx).
		Fields(dao.MediaStrategy.Columns().Id, dao.MediaStrategy.Columns().Name).
		WhereIn(dao.MediaStrategy.Columns().Id, ids).
		Scan(&strategies)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyListQueryFailed)
	}
	for _, strategy := range strategies {
		result[strategy.Id] = strategy.Name
	}
	return result, nil
}

// normalizeDeviceID validates and trims one device ID.
func normalizeDeviceID(deviceID string) (string, error) {
	normalized := strings.TrimSpace(deviceID)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaBindingDeviceRequired)
	}
	return normalized, nil
}

// normalizeTenantID validates and trims one media tenant ID.
func normalizeTenantID(tenantID string) (string, error) {
	normalized := strings.TrimSpace(tenantID)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaBindingTenantRequired)
	}
	return normalized, nil
}

// buildResolveOutput converts a strategy source and entity into resolution output.
func buildResolveOutput(source StrategySource, strategy *strategyEntity) *ResolveStrategyOutput {
	out := &ResolveStrategyOutput{
		Matched:     strategy != nil,
		Source:      string(source),
		SourceLabel: strategySourceLabel(source),
	}
	if strategy != nil {
		out.StrategyId = strategy.Id
		out.StrategyName = strategy.Name
		out.Strategy = strategy.Strategy
	}
	return out
}

// strategySourceLabel returns the Chinese label for one strategy source.
func strategySourceLabel(source StrategySource) string {
	switch source {
	case StrategySourceTenantDevice:
		return "租户设备策略"
	case StrategySourceDevice:
		return "设备策略"
	case StrategySourceTenant:
		return "租户策略"
	case StrategySourceGlobal:
		return "全局策略"
	default:
		return "未匹配"
	}
}

// collectDeviceBindingStrategyIDs collects unique strategy IDs from device bindings.
func collectDeviceBindingStrategyIDs(rows []*deviceBindingEntity) []int64 {
	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if _, ok := seen[row.StrategyId]; ok {
			continue
		}
		seen[row.StrategyId] = struct{}{}
		ids = append(ids, row.StrategyId)
	}
	return ids
}

// collectTenantBindingStrategyIDs collects unique strategy IDs from tenant bindings.
func collectTenantBindingStrategyIDs(rows []*tenantBindingEntity) []int64 {
	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if _, ok := seen[row.StrategyId]; ok {
			continue
		}
		seen[row.StrategyId] = struct{}{}
		ids = append(ids, row.StrategyId)
	}
	return ids
}

// collectTenantDeviceBindingStrategyIDs collects unique strategy IDs from tenant-device bindings.
func collectTenantDeviceBindingStrategyIDs(rows []*tenantDeviceBindingEntity) []int64 {
	ids := make([]int64, 0, len(rows))
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		if _, ok := seen[row.StrategyId]; ok {
			continue
		}
		seen[row.StrategyId] = struct{}{}
		ids = append(ids, row.StrategyId)
	}
	return ids
}
