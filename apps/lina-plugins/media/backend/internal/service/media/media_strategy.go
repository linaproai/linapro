// This file implements media strategy CRUD and global strategy selection.

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

// ListStrategiesInput defines media strategy list filters.
type ListStrategiesInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches strategy name.
	Enable   int    // Enable filters enable status when non-zero.
	Global   int    // Global filters global status when non-zero.
}

// ListStrategiesOutput defines paged media strategies.
type ListStrategiesOutput struct {
	List  []*StrategyOutput // List contains the current page.
	Total int               // Total is the total matched row count.
}

// StrategyOutput defines one media strategy response.
type StrategyOutput struct {
	Id        int64  // Id is the strategy primary key.
	Name      string // Name is the strategy name.
	Strategy  string // Strategy is the YAML strategy body.
	Global    int    // Global marks whether the strategy is global.
	Enable    int    // Enable marks whether the strategy is enabled.
	CreatorId int64  // CreatorId is the creator user ID.
	UpdaterId int64  // UpdaterId is the last updater user ID.
	CreateTime string // CreateTime is the formatted creation time.
	UpdateTime string // UpdateTime is the formatted update time.
}

// StrategyMutationInput defines strategy create/update input.
type StrategyMutationInput struct {
	Name     string // Name is the strategy name.
	Strategy string // Strategy is the YAML strategy body.
	Enable   int    // Enable marks whether the strategy is enabled.
	Global   int    // Global marks whether the strategy is global.
}

// strategyEntity reuses the plugin-local generated strategy entity.
type strategyEntity = entitymodel.MediaStrategy

// ListStrategies returns paged media strategies.
func (s *serviceImpl) ListStrategies(ctx context.Context, in ListStrategiesInput) (*ListStrategiesOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaStrategy.Columns()
	model := dao.MediaStrategy.Ctx(ctx)

	keyword := strings.TrimSpace(in.Keyword)
	if keyword != "" {
		model = model.WhereLike(columns.Name, "%"+keyword+"%")
	}
	if in.Enable > 0 {
		enable, err := normalizeSwitchValue(in.Enable, SwitchOn)
		if err != nil {
			return nil, err
		}
		model = model.Where(columns.Enable, enable)
	}
	if in.Global > 0 {
		global, err := normalizeSwitchValue(in.Global, SwitchOff)
		if err != nil {
			return nil, err
		}
		model = model.Where(columns.Global, global)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyCountQueryFailed)
	}

	items := make([]*strategyEntity, 0)
	err = model.
		OrderDesc(columns.UpdateTime).
		OrderDesc(columns.Id).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyListQueryFailed)
	}

	list := make([]*StrategyOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildStrategyOutput(item))
	}
	return &ListStrategiesOutput{List: list, Total: total}, nil
}

// GetStrategy returns one media strategy by ID.
func (s *serviceImpl) GetStrategy(ctx context.Context, id int64) (*StrategyOutput, error) {
	record, err := s.getStrategyEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	return buildStrategyOutput(record), nil
}

// CreateStrategy creates one media strategy.
func (s *serviceImpl) CreateStrategy(ctx context.Context, in StrategyMutationInput) (int64, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return 0, err
	}
	normalized, err := normalizeStrategyMutationInput(in)
	if err != nil {
		return 0, err
	}

	var id int64
	err = dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if normalized.Global == int(SwitchOn) {
			if err := s.clearGlobalStrategies(ctx); err != nil {
				return err
			}
		}
		createdID, insertErr := dao.MediaStrategy.Ctx(ctx).Data(do.MediaStrategy{
			Name:      normalized.Name,
			Strategy:  normalized.Strategy,
			Global:    normalized.Global,
			Enable:    normalized.Enable,
			CreatorId: s.currentActorID(ctx),
			UpdaterId: s.currentActorID(ctx),
		}).InsertAndGetId()
		if insertErr != nil {
			return bizerr.WrapCode(insertErr, CodeMediaStrategyCreateFailed)
		}
		id = createdID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateStrategy updates one media strategy.
func (s *serviceImpl) UpdateStrategy(ctx context.Context, id int64, in StrategyMutationInput) error {
	if err := validateMediaTablesReady(ctx); err != nil {
		return err
	}
	if _, err := s.getStrategyEntity(ctx, id); err != nil {
		return err
	}
	normalized, err := normalizeStrategyMutationInput(in)
	if err != nil {
		return err
	}

	return dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if normalized.Global == int(SwitchOn) {
			if err := s.clearGlobalStrategies(ctx); err != nil {
				return err
			}
		}
		_, updateErr := dao.MediaStrategy.Ctx(ctx).
			Where(do.MediaStrategy{Id: id}).
			Data(do.MediaStrategy{
				Name:      normalized.Name,
				Strategy:  normalized.Strategy,
				Global:    normalized.Global,
				Enable:    normalized.Enable,
				UpdaterId: s.currentActorID(ctx),
			}).
			Update()
		if updateErr != nil {
			return bizerr.WrapCode(updateErr, CodeMediaStrategyUpdateFailed)
		}
		return nil
	})
}

// UpdateStrategyEnable changes one media strategy enable status.
func (s *serviceImpl) UpdateStrategyEnable(ctx context.Context, id int64, enable int) error {
	if err := validateMediaTablesReady(ctx); err != nil {
		return err
	}
	if _, err := s.getStrategyEntity(ctx, id); err != nil {
		return err
	}
	normalizedEnable, err := normalizeSwitchValue(enable, SwitchOn)
	if err != nil {
		return err
	}

	_, err = dao.MediaStrategy.Ctx(ctx).
		Where(do.MediaStrategy{Id: id}).
		Data(do.MediaStrategy{
			Enable:    normalizedEnable,
			UpdaterId: s.currentActorID(ctx),
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaStrategyUpdateFailed)
	}
	return nil
}

// SetGlobalStrategy sets one media strategy as the active global strategy.
func (s *serviceImpl) SetGlobalStrategy(ctx context.Context, id int64) error {
	if err := validateMediaTablesReady(ctx); err != nil {
		return err
	}
	if _, err := s.getStrategyEntity(ctx, id); err != nil {
		return err
	}

	return dao.MediaStrategy.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.clearGlobalStrategies(ctx); err != nil {
			return err
		}
		_, err := dao.MediaStrategy.Ctx(ctx).
			Where(do.MediaStrategy{Id: id}).
			Data(do.MediaStrategy{
				Global:    int(SwitchOn),
				Enable:    int(SwitchOn),
				UpdaterId: s.currentActorID(ctx),
			}).
			Update()
		if err != nil {
			return bizerr.WrapCode(err, CodeMediaStrategyUpdateFailed)
		}
		return nil
	})
}

// DeleteStrategy deletes one unreferenced media strategy.
func (s *serviceImpl) DeleteStrategy(ctx context.Context, id int64) error {
	if _, err := s.getStrategyEntity(ctx, id); err != nil {
		return err
	}
	referenced, err := s.strategyReferenced(ctx, id)
	if err != nil {
		return err
	}
	if referenced {
		return bizerr.NewCode(CodeMediaStrategyReferenced)
	}

	_, err = dao.MediaStrategy.Ctx(ctx).
		Where(do.MediaStrategy{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaStrategyDeleteFailed)
	}
	return nil
}

// getStrategyEntity returns one media strategy entity by ID.
func (s *serviceImpl) getStrategyEntity(ctx context.Context, id int64) (*strategyEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	var record *strategyEntity
	err := dao.MediaStrategy.Ctx(ctx).
		Where(do.MediaStrategy{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaStrategyDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaStrategyNotFound)
	}
	return record, nil
}

// clearGlobalStrategies turns off all global strategies.
func (s *serviceImpl) clearGlobalStrategies(ctx context.Context) error {
	_, err := dao.MediaStrategy.Ctx(ctx).
		Where(dao.MediaStrategy.Columns().Global, int(SwitchOn)).
		Data(do.MediaStrategy{
			Global:    int(SwitchOff),
			UpdaterId: s.currentActorID(ctx),
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaStrategyUpdateFailed)
	}
	return nil
}

// strategyReferenced reports whether any binding references the strategy.
func (s *serviceImpl) strategyReferenced(ctx context.Context, id int64) (bool, error) {
	deviceCount, err := dao.MediaStrategyDevice.Ctx(ctx).
		Where(dao.MediaStrategyDevice.Columns().StrategyId, id).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	if deviceCount > 0 {
		return true, nil
	}

	tenantDeviceCount, err := dao.MediaStrategyDeviceTenant.Ctx(ctx).
		Where(dao.MediaStrategyDeviceTenant.Columns().StrategyId, id).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	if tenantDeviceCount > 0 {
		return true, nil
	}

	tenantCount, err := dao.MediaStrategyTenant.Ctx(ctx).
		Where(dao.MediaStrategyTenant.Columns().StrategyId, id).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaBindingCountQueryFailed)
	}
	return tenantCount > 0, nil
}

// normalizeStrategyMutationInput validates strategy mutation input.
func normalizeStrategyMutationInput(in StrategyMutationInput) (StrategyMutationInput, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return StrategyMutationInput{}, bizerr.NewCode(CodeMediaStrategyNameRequired)
	}
	strategy := strings.TrimSpace(in.Strategy)
	if strategy == "" {
		return StrategyMutationInput{}, bizerr.NewCode(CodeMediaStrategyContentRequired)
	}
	enable, err := normalizeSwitchValue(in.Enable, SwitchOn)
	if err != nil {
		return StrategyMutationInput{}, err
	}
	global, err := normalizeSwitchValue(in.Global, SwitchOff)
	if err != nil {
		return StrategyMutationInput{}, err
	}
	return StrategyMutationInput{
		Name:     name,
		Strategy: strategy,
		Enable:   enable,
		Global:   global,
	}, nil
}

// buildStrategyOutput converts one generated strategy entity into service output.
func buildStrategyOutput(item *strategyEntity) *StrategyOutput {
	if item == nil {
		return &StrategyOutput{}
	}
	return &StrategyOutput{
		Id:        item.Id,
		Name:      item.Name,
		Strategy:  item.Strategy,
		Global:    item.Global,
		Enable:    item.Enable,
		CreatorId: item.CreatorId,
		UpdaterId: item.UpdaterId,
		CreateTime: formatTime(item.CreateTime),
		UpdateTime: formatTime(item.UpdateTime),
	}
}
