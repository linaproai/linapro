// This file contains sysconfig query and mutation methods, including tenant
// fallback scoping, protected config validation, and runtime snapshot refresh.

package sysconfig

import (
	"bytes"
	"context"
	"sort"

	"github.com/xuri/excelize/v2"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
)

// List queries config list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysConfig.Columns()
		m    = dao.SysConfig.Ctx(ctx)
	)
	m = applySysconfigFallbackScope(ctx, m)

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
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves config by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysConfig, error) {
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

// Create creates a new config record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if err := validateManagedConfigValue(in.Key, in.Value); err != nil {
		return 0, err
	}

	var createdID int64
	err := s.withConfigMutation(ctx, func(ctx context.Context) error {
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

		// Insert config (GoFrame auto-fills created_at and updated_at)
		data := currentTenantConfigDO(ctx)
		data.Name = in.Name
		data.Key = in.Key
		data.Value = in.Value
		data.IsBuiltin = builtInConfigFlag(in.Key)
		data.Remark = in.Remark

		insertedID, insertErr := dao.SysConfig.Ctx(ctx).Data(data).InsertAndGetId()
		if insertErr != nil {
			return insertErr
		}
		createdID = insertedID

		return s.refreshRuntimeParamSnapshotIfNeeded(ctx, in.Key, "", in.Value, true)
	})
	if err != nil {
		return 0, err
	}

	return int(createdID), nil
}

// Update updates config information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	return s.withConfigMutation(ctx, func(ctx context.Context) error {
		// Check config exists
		existing, err := s.GetById(ctx, in.Id)
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

		finalKey := existing.Key
		if in.Key != nil {
			finalKey = *in.Key
		}
		finalValue := existing.Value
		if in.Value != nil {
			finalValue = *in.Value
		}
		if err = validateManagedConfigValue(finalKey, finalValue); err != nil {
			return err
		}

		data := do.SysConfig{}
		if in.Name != nil {
			data.Name = *in.Name
		}
		if in.Key != nil {
			data.Key = *in.Key
		}
		if in.Value != nil {
			data.Value = *in.Value
		}
		if in.Remark != nil {
			data.Remark = *in.Remark
		}

		_, err = dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: in.Id}).Data(data).Update()
		if err != nil {
			return err
		}

		return s.refreshRuntimeParamSnapshotIfNeeded(ctx, finalKey, existing.Value, finalValue, false)
	})
}

// Delete soft-deletes a config record using GoFrame's auto soft-delete feature.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	// Check config exists
	existing, err := s.GetById(ctx, id)
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
	return err
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
	return record.IsBuiltin == 1 || hostconfig.IsProtectedConfigParam(record.Key)
}

// builtInConfigFlag returns the persisted built-in marker for protected
// system parameters that are created through management or import paths.
func builtInConfigFlag(key string) int {
	if hostconfig.IsProtectedConfigParam(key) {
		return 1
	}
	return 0
}

// GetByKey retrieves config by key name.
func (s *serviceImpl) GetByKey(ctx context.Context, key string) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	model := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Key: key})
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
	return cfg, nil
}

// Export generates an Excel file with config data.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysConfig.Columns()
	m := dao.SysConfig.Ctx(ctx)
	m = applySysconfigFallbackScope(ctx, m)

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
		if err = setCellValue(f, sheet, 4, row, c.Remark); err != nil {
			return nil, err
		}
		if c.CreatedAt != nil {
			if err = setCellValue(f, sheet, 5, row, c.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
		if c.UpdatedAt != nil {
			if err = setCellValue(f, sheet, 6, row, c.UpdatedAt.String()); err != nil {
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
