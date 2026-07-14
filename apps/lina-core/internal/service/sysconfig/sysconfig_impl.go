// This file contains sysconfig query and mutation methods, including tenant
// fallback scoping, protected config validation, and runtime snapshot refresh.

package sysconfig

import (
	"bytes"
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/configvaluetype"
)

// List queries config list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysConfig.Columns()
		m    = dao.SysConfig.Ctx(ctx)
	)
	m = applySysconfigFallbackScope(ctx, m)
	// Management surface only lists system-manageable rows; plugin closed-loop
	// settings (system_manageable=0) stay out of this page.
	m = m.Where(cols.SystemManageable, 1)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Key != "" {
		m = m.WhereLike(cols.Key, "%"+in.Key+"%")
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(cols.CreatedAt, in.BeginTime+" 00:00:00")
	}
	if in.EndTime != "" {
		m = m.WhereLTE(cols.CreatedAt, in.EndTime+" 23:59:59")
	}

	var rows []*entity.SysConfig
	err := m.OrderDesc(cols.Id).Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := visibleConfigs(ctx, rows)
	sort.SliceStable(list, func(i int, j int) bool {
		return list[i].Id > list[j].Id
	})
	total := len(list)
	list = paginateConfigs(list, in.PageNum, in.PageSize)
	s.localizeConfigEntities(ctx, list)

	return &ListOutput{
		List:  projectConfigs(ctx, list),
		Total: total,
	}, nil
}

// withConfigMutation runs one sysconfig mutation inside the shared transaction
// boundary used for runtime-param refresh coordination.
func (s *serviceImpl) withConfigMutation(ctx context.Context, handler func(ctx context.Context) error) error {
	return dao.SysConfig.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		return handler(ctx)
	})
}

// refreshRuntimeParamSnapshotIfNeeded marks the sys_config snapshot dirty when
// any runtime configuration value changes.
func (s *serviceImpl) refreshRuntimeParamSnapshotIfNeeded(
	ctx context.Context,
	key string,
	previousValue string,
	currentValue string,
	forceRefresh bool,
) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	if !forceRefresh && previousValue == currentValue {
		return nil
	}
	if s == nil || s.configSvc == nil {
		return nil
	}
	return s.configSvc.MarkRuntimeParamsChanged(ctx)
}

// GetById retrieves config by ID for edit/detail display. Name and remark are
// localized for the request language; value stays as the stored raw text.
// Non system-manageable rows are not found on the management surface.
func (s *serviceImpl) GetById(ctx context.Context, id int64) (*entity.SysConfig, error) {
	cfg, err := s.getByIdRawForManagement(ctx, id)
	if err != nil {
		return nil, err
	}
	s.localizeConfigEntityMetadata(ctx, cfg)
	return cfg, nil
}

// getByIdRaw loads one config row by ID without i18n projection. Mutation paths
// must use this helper so localized name/remark never drive write-back or
// built-in protection checks against display text.
func (s *serviceImpl) getByIdRaw(ctx context.Context, id int64) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	model := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: id})
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, bizerr.NewCode(CodeSysConfigNotFound)
	}
	return cfg, nil
}

// getByIdRawForManagement loads one management-surface config row. Non
// system-manageable rows are treated as not found so plugin closed-loop
// settings cannot be inspected or mutated through system settings APIs.
func (s *serviceImpl) getByIdRawForManagement(ctx context.Context, id int64) (*entity.SysConfig, error) {
	cfg, err := s.getByIdRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if !isSystemManageableRecord(cfg) {
		return nil, bizerr.NewCode(CodeSysConfigNotFound)
	}
	return cfg, nil
}

// isSystemManageableRecord reports whether a sys_config row may appear and be
// mutated on the system config management page (system_manageable = 1).
func isSystemManageableRecord(record *entity.SysConfig) bool {
	return record != nil && record.SystemManageable == 1
}

// Create creates a new config record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int64, error) {
	valueType, optionsRaw, err := resolveCreateTypeMetadata(ctx, in.Key, in.ValueType, in.Options)
	if err != nil {
		return 0, err
	}
	value := normalizePersistedValue(valueType, in.Value)
	if err = validateTypedConfigValue(valueType, optionsRaw, value); err != nil {
		return 0, err
	}
	if err = validateManagedConfigValue(in.Key, value); err != nil {
		return 0, err
	}

	var createdID int64
	err = s.withConfigMutation(ctx, func(ctx context.Context) error {
		// Check key uniqueness (GoFrame auto-adds deleted_at IS NULL)
		model := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Key: in.Key})
		model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
		count, countErr := model.Count()
		if countErr != nil {
			return countErr
		}
		if count > 0 {
			return bizerr.NewCode(CodeSysConfigKeyExists, bizerr.P("key", in.Key))
		}

		// Insert config (GoFrame auto-fills created_at and updated_at).
		// Management-surface creates are always system-manageable.
		data := currentTenantConfigDO(ctx)
		data.Name = in.Name
		data.Key = in.Key
		data.Value = value
		data.ValueType = valueType.String()
		data.Options = optionsRaw
		data.IsBuiltin = builtInConfigFlag(in.Key)
		data.SystemManageable = 1
		data.Remark = in.Remark

		insertedID, insertErr := dao.SysConfig.Ctx(ctx).Data(data).InsertAndGetId()
		if insertErr != nil {
			return insertErr
		}
		createdID = insertedID
		return nil
	})
	if err != nil {
		return 0, err
	}

	if err = s.refreshRuntimeParamSnapshotIfNeeded(ctx, in.Key, "", value, true); err != nil {
		return 0, err
	}

	return createdID, nil
}

// Update updates config information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	var (
		previousKey   string
		previousValue string
		finalKey      string
		finalValue    string
	)
	if err := s.withConfigMutation(ctx, func(ctx context.Context) error {
		// Check config exists using raw storage values (not display projections).
		// Non system-manageable rows are outside the management surface.
		existing, err := s.getByIdRawForManagement(ctx, in.Id)
		if err != nil {
			return err
		}
		if isBuiltInConfigRecord(existing) && in.Key != nil && *in.Key != existing.Key {
			return bizerr.NewCode(CodeSysConfigBuiltinKeyRenameDenied)
		}

		// Check key uniqueness (exclude self) - GoFrame auto-adds deleted_at IS NULL
		if in.Key != nil {
			cols := dao.SysConfig.Columns()
			model := dao.SysConfig.Ctx(ctx).
				Where(do.SysConfig{Key: *in.Key}).
				WhereNot(cols.Id, in.Id)
			model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
			count, countErr := model.Count()
			if countErr != nil {
				return countErr
			}
			if count > 0 {
				return bizerr.NewCode(CodeSysConfigKeyExists, bizerr.P("key", *in.Key))
			}
		}

		previousKey = existing.Key
		finalKey = existing.Key
		if in.Key != nil {
			finalKey = *in.Key
		}
		previousValue = existing.Value
		finalValue = existing.Value
		if in.Value != nil {
			finalValue = *in.Value
		}

		finalValueType := entityValueType(existing.ValueType)
		finalOptionsRaw := existing.Options
		if isBuiltInConfigRecord(existing) {
			if in.ValueType != nil && strings.TrimSpace(*in.ValueType) != "" &&
				entityValueType(*in.ValueType) != finalValueType {
				return bizerr.NewCode(CodeSysConfigBuiltinTypeChangeDenied)
			}
			if in.Options != nil {
				encoded, encodeErr := configvaluetype.EncodeOptions(*in.Options)
				if encodeErr != nil {
					return bizerr.WrapCode(encodeErr, CodeSysConfigOptionsInvalid)
				}
				if encoded != strings.TrimSpace(existing.Options) {
					return bizerr.NewCode(CodeSysConfigBuiltinTypeChangeDenied)
				}
			}
		} else {
			if in.ValueType != nil {
				resolved, resolveErr := configvaluetype.ResolveCode(*in.ValueType)
				if resolveErr != nil {
					return bizerr.WrapCode(resolveErr, CodeSysConfigValueTypeInvalid)
				}
				finalValueType = resolved
			}
			if in.Options != nil {
				encoded, encodeErr := configvaluetype.EncodeOptions(*in.Options)
				if encodeErr != nil {
					return bizerr.WrapCode(encodeErr, CodeSysConfigOptionsInvalid)
				}
				finalOptionsRaw = encoded
			}
		}

		finalValue = normalizePersistedValue(finalValueType, finalValue)
		if err = validateTypedConfigValue(finalValueType, finalOptionsRaw, finalValue); err != nil {
			return err
		}
		if err = validateManagedConfigValue(finalKey, finalValue); err != nil {
			return err
		}

		data := do.SysConfig{}
		// Built-in name/remark are framework display metadata owned by i18n
		// resources; never persist edit-form projections back into sys_config.
		if !isBuiltInConfigRecord(existing) {
			if in.Name != nil {
				data.Name = *in.Name
			}
			if in.Remark != nil {
				data.Remark = *in.Remark
			}
			if in.ValueType != nil {
				data.ValueType = finalValueType.String()
			}
			if in.Options != nil {
				data.Options = finalOptionsRaw
			}
		}
		if in.Key != nil {
			data.Key = *in.Key
		}
		if in.Value != nil {
			data.Value = finalValue
		}

		_, err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return s.refreshRuntimeParamSnapshotIfNeeded(ctx, finalKey, previousValue, finalValue, previousKey != finalKey)
}

// Delete soft-deletes a config record using GoFrame's auto soft-delete feature.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	// Check config exists using raw storage values; non system-manageable rows
	// are outside the management surface.
	existing, err := s.getByIdRawForManagement(ctx, id)
	if err != nil {
		return err
	}
	if isBuiltInConfigRecord(existing) {
		return bizerr.NewCode(CodeSysConfigBuiltinDeleteDenied)
	}

	// Soft delete
	_, err = dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: id}).
		Delete()
	if err != nil {
		return err
	}
	return s.refreshRuntimeParamSnapshotIfNeeded(ctx, existing.Key, existing.Value, "", true)
}

// validateManagedConfigValue delegates validation for protected runtime and
// public-frontend config keys to the host config service rules.
func validateManagedConfigValue(key string, value string) error {
	if err := hostconfig.ValidateProtectedConfigValue(key, value); err != nil {
		return bizerr.WrapCode(err, CodeSysConfigProtectedValueInvalid)
	}
	return nil
}

// isBuiltInConfigRecord reports whether one sys_config record is delivered by
// the host as a built-in system parameter.
func isBuiltInConfigRecord(record *entity.SysConfig) bool {
	if record == nil {
		return false
	}
	return record.IsBuiltin == 1 || hostconfig.IsManagedSysConfigKey(record.Key)
}

// builtInConfigFlag returns the persisted built-in marker for protected
// system parameters that are created through management or import paths.
func builtInConfigFlag(key string) int {
	if hostconfig.IsManagedSysConfigKey(key) {
		return 1
	}
	return 0
}

// GetByKey retrieves config by key name for the management surface.
// Non system-manageable rows are reported as key-not-found.
func (s *serviceImpl) GetByKey(ctx context.Context, key string) (*ConfigProjection, error) {
	var cfg *entity.SysConfig
	model := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Key: key, SystemManageable: 1})
	model = applySysconfigFallbackScope(ctx, model).
		OrderDesc(datascope.TenantColumn)
	err := model.Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, bizerr.NewCode(CodeSysConfigKeyNotFound)
	}
	s.localizeConfigEntity(ctx, cfg)
	return ProjectConfig(ctx, cfg), nil
}

// Export generates an Excel file with config data.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysConfig.Columns()
	m := dao.SysConfig.Ctx(ctx)
	m = applySysconfigFallbackScope(ctx, m)
	// Keep export aligned with the management list: only system-manageable rows.
	m = m.Where(cols.SystemManageable, 1)

	if len(in.Ids) > 0 {
		m = m.WhereIn(cols.Id, in.Ids)
	} else {
		if in.Name != "" {
			m = m.WhereLike(cols.Name, "%"+in.Name+"%")
		}
		if in.Key != "" {
			m = m.WhereLike(cols.Key, "%"+in.Key+"%")
		}
		if in.BeginTime != "" {
			m = m.WhereGTE(cols.CreatedAt, in.BeginTime+" 00:00:00")
		}
		if in.EndTime != "" {
			m = m.WhereLTE(cols.CreatedAt, in.EndTime+" 23:59:59")
		}
	}

	var rows []*entity.SysConfig
	err = m.OrderAsc(cols.Id).Scan(&rows)
	if err != nil {
		return nil, err
	}
	list := visibleConfigs(ctx, rows)

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(ctx, f, &err)
	sheet := "Sheet1"

	headers := s.buildLocalizedExportHeaders(ctx)
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, c := range list {
		row := i + 2
		if err = setCellValue(f, sheet, 1, row, c.Name); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 2, row, c.Key); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 3, row, c.Value); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 4, row, entityValueType(c.ValueType).String()); err != nil {
			return nil, err
		}
		// Export options in simple line format for human-friendly Excel editing.
		exportedOptions := c.Options
		if parsed, parseErr := configvaluetype.ParseOptions(c.Options); parseErr == nil {
			exportedOptions = configvaluetype.FormatOptionsSimple(parsed)
		}
		if err = setCellValue(f, sheet, 5, row, exportedOptions); err != nil {
			return nil, err
		}
		if err = setCellValue(f, sheet, 6, row, c.Remark); err != nil {
			return nil, err
		}
		if c.CreatedAt != nil {
			if err = setCellValue(f, sheet, 7, row, c.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
		if c.UpdatedAt != nil {
			if err = setCellValue(f, sheet, 8, row, c.UpdatedAt.String()); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}
