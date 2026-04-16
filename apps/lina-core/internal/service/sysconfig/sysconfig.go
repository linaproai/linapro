// Package sysconfig implements system-configuration query, mutation, import,
// and export services for the Lina core host service.
package sysconfig

import (
	"bytes"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/xuri/excelize/v2"

	"io"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Service defines the sysconfig service contract.
type Service interface {
	// List queries config list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves config by ID.
	GetById(ctx context.Context, id int) (*entity.SysConfig, error)
	// Create creates a new config record.
	Create(ctx context.Context, in CreateInput) (int, error)
	// Update updates config information.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes a config record using GoFrame's auto soft-delete feature.
	Delete(ctx context.Context, id int) error
	// GetByKey retrieves config by key name.
	GetByKey(ctx context.Context, key string) (*entity.SysConfig, error)
	// Export generates an Excel file with config data.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
	// Import reads an Excel file and creates configs from it.
	// If updateSupport is true, existing records (matched by key) will be updated; otherwise, they will be skipped.
	Import(ctx context.Context, fileReader io.Reader, updateSupport bool) (result *ImportResult, err error)
	// GenerateImportTemplate creates an Excel template for config import.
	GenerateImportTemplate() (data []byte, err error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

func New() Service {
	return &serviceImpl{}
}

type ListInput struct {
	PageNum   int
	PageSize  int
	Name      string
	Key       string
	BeginTime string
	EndTime   string
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysConfig // Config list
	Total int                 // Total count
}

// List queries config list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysConfig.Columns()
		m    = dao.SysConfig.Ctx(ctx)
	)

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

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysConfig
	err = m.Page(in.PageNum, in.PageSize).
		OrderDesc(cols.Id).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// GetById retrieves config by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: id}).
		Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, gerror.New("参数设置不存在")
	}
	return cfg, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Name   string // Parameter name
	Key    string // Parameter key
	Value  string // Parameter value
	Remark string // Remark
}

// Create creates a new config record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check key uniqueness (GoFrame auto-adds deleted_at IS NULL)
	count, err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Key: in.Key}).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("参数键名已存在")
	}

	// Insert config (GoFrame auto-fills created_at and updated_at)
	id, err := dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
		Name:   in.Name,
		Key:    in.Key,
		Value:  in.Value,
		Remark: in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int     // Parameter ID
	Name   *string // Parameter name
	Key    *string // Parameter key
	Value  *string // Parameter value
	Remark *string // Remark
}

// Update updates config information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Check config exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	// Check key uniqueness (exclude self) - GoFrame auto-adds deleted_at IS NULL
	if in.Key != nil {
		cols := dao.SysConfig.Columns()
		count, err := dao.SysConfig.Ctx(ctx).
			Where(do.SysConfig{Key: *in.Key}).
			WhereNot(cols.Id, in.Id).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.New("参数键名已存在")
		}
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

	_, err := dao.SysConfig.Ctx(ctx).Where(do.SysConfig{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a config record using GoFrame's auto soft-delete feature.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	// Check config exists
	if _, err := s.GetById(ctx, id); err != nil {
		return err
	}

	// Soft delete
	_, err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Id: id}).
		Delete()
	return err
}

// GetByKey retrieves config by key name.
func (s *serviceImpl) GetByKey(ctx context.Context, key string) (*entity.SysConfig, error) {
	var cfg *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{Key: key}).
		Scan(&cfg)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, gerror.New("参数键名不存在")
	}
	return cfg, nil
}

// ExportInput defines input for Export function.
type ExportInput struct {
	Name      string // Parameter name, supports fuzzy search
	Key       string // Parameter key, supports fuzzy search
	BeginTime string // Creation time start
	EndTime   string // Creation time end
	Ids       []int  // Specific IDs to export; if empty, export all matching records
}

// Export generates an Excel file with config data.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	cols := dao.SysConfig.Columns()
	m := dao.SysConfig.Ctx(ctx)

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

	var list []*entity.SysConfig
	err = m.OrderAsc(cols.Id).Scan(&list)
	if err != nil {
		return nil, err
	}

	// Create Excel file
	f := excelize.NewFile()
	defer closeExcelFile(f, &err)
	sheet := "Sheet1"

	headers := []string{"参数名称", "参数键名", "参数键值", "备注", "创建时间", "修改时间"}
	for i, h := range headers {
		if err = setCellValue(f, sheet, i+1, 1, h); err != nil {
			return nil, err
		}
	}

	for i, c := range list {
		row := i + 2
		if err = setCellValueByName(f, sheet, cellName(1, row), c.Name); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(2, row), c.Key); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(3, row), c.Value); err != nil {
			return nil, err
		}
		if err = setCellValueByName(f, sheet, cellName(4, row), c.Remark); err != nil {
			return nil, err
		}
		if c.CreatedAt != nil {
			if err = setCellValueByName(f, sheet, cellName(5, row), c.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
		if c.UpdatedAt != nil {
			if err = setCellValueByName(f, sheet, cellName(6, row), c.UpdatedAt.String()); err != nil {
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
