// This file implements media node, device-node, and tenant stream config CRUD.

package media

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	entitymodel "lina-plugin-media/backend/internal/model/entity"
)

// ListNodesInput defines media node list filters.
type ListNodesInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches node number, name, or gateway URL.
}

// ListNodesOutput defines paged media node entries.
type ListNodesOutput struct {
	List  []*NodeOutput // List contains current page nodes.
	Total int           // Total is total matched row count.
}

// NodeOutput defines one media node response.
type NodeOutput struct {
	Id         int    // Id is the generated primary key.
	NodeNum    int    // NodeNum is the node number.
	Name       string // Name is the node name.
	QnUrl      string // QnUrl is the node gateway URL.
	BasicUrl   string // BasicUrl is the basic platform gateway URL.
	DnUrl      string // DnUrl is the district gateway URL.
	CreatorId  int    // CreatorId is the creator user ID.
	CreateTime string // CreateTime is the formatted creation time.
	UpdaterId  int    // UpdaterId is the last updater user ID.
	UpdateTime string // UpdateTime is the formatted update time.
}

// NodeMutationInput defines media node create/update input.
type NodeMutationInput struct {
	NodeNum  int    // NodeNum is the node number.
	Name     string // Name is the node name.
	QnUrl    string // QnUrl is the node gateway URL.
	BasicUrl string // BasicUrl is the basic platform gateway URL.
	DnUrl    string // DnUrl is the district gateway URL.
}

// NodeMutationOutput defines media node mutation result.
type NodeMutationOutput struct {
	NodeNum int // NodeNum is the node number.
}

// ListDeviceNodesInput defines device-node list filters.
type ListDeviceNodesInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches device ID or node number.
}

// ListDeviceNodesOutput defines paged device-node mappings.
type ListDeviceNodesOutput struct {
	List  []*DeviceNodeOutput // List contains current page mappings.
	Total int                 // Total is total matched row count.
}

// DeviceNodeOutput defines one device-node mapping response.
type DeviceNodeOutput struct {
	DeviceId string // DeviceId is the GB device ID.
	NodeNum  int    // NodeNum is the linked node number.
	NodeName string // NodeName is the linked node name.
}

// DeviceNodeMutationInput defines device-node create/update input.
type DeviceNodeMutationInput struct {
	DeviceId string // DeviceId is the GB device ID.
	NodeNum  int    // NodeNum is the linked node number.
}

// DeviceNodeMutationOutput defines device-node mutation result.
type DeviceNodeMutationOutput struct {
	DeviceId string // DeviceId is the GB device ID.
}

// ListTenantStreamConfigsInput defines tenant stream config list filters.
type ListTenantStreamConfigsInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches tenant ID or node number.
	Enable   *int   // Enable filters status when set.
}

// ListTenantStreamConfigsOutput defines paged tenant stream configs.
type ListTenantStreamConfigsOutput struct {
	List  []*TenantStreamConfigOutput // List contains current page configs.
	Total int                         // Total is total matched row count.
}

// TenantStreamConfigOutput defines one tenant stream config response.
type TenantStreamConfigOutput struct {
	TenantId      string // TenantId is the media tenant ID.
	MaxConcurrent int    // MaxConcurrent is the max stream concurrency.
	NodeNum       int    // NodeNum is the linked node number.
	NodeName      string // NodeName is the linked node name.
	Enable        int    // Enable marks whether the config is active.
	CreatorId     int    // CreatorId is the creator user ID.
	CreateTime    string // CreateTime is the formatted creation time.
	UpdaterId     int    // UpdaterId is the last updater user ID.
	UpdateTime    string // UpdateTime is the formatted update time.
}

// TenantStreamConfigMutationInput defines tenant stream config create/update input.
type TenantStreamConfigMutationInput struct {
	TenantId      string // TenantId is the media tenant ID.
	MaxConcurrent int    // MaxConcurrent is the max stream concurrency.
	NodeNum       int    // NodeNum is the linked node number.
	Enable        int    // Enable marks whether the config is active.
}

// TenantStreamConfigMutationOutput defines tenant stream config mutation result.
type TenantStreamConfigMutationOutput struct {
	TenantId string // TenantId is the media tenant ID.
}

// Generated entity aliases for media node config tables.
type (
	nodeEntity               = entitymodel.HgNode
	deviceNodeEntity         = entitymodel.HgDeviceNode
	tenantStreamConfigEntity = entitymodel.HgTenantStreamConfig
)

// ListNodes returns paged media nodes.
func (s *serviceImpl) ListNodes(ctx context.Context, in ListNodesInput) (*ListNodesOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.HgNode.Columns()
	model := dao.HgNode.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.NodeNum+"::text LIKE ? OR "+columns.Name+" LIKE ? OR "+columns.QnUrl+" LIKE ? OR "+columns.BasicUrl+" LIKE ? OR "+columns.DnUrl+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeCountQueryFailed)
	}

	items := make([]*nodeEntity, 0)
	err = model.
		OrderAsc(columns.NodeNum).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeListQueryFailed)
	}

	list := make([]*NodeOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildNodeOutput(item))
	}
	return &ListNodesOutput{List: list, Total: total}, nil
}

// GetNode returns one media node by node number.
func (s *serviceImpl) GetNode(ctx context.Context, nodeNum int) (*NodeOutput, error) {
	record, err := s.getNodeEntity(ctx, nodeNum)
	if err != nil {
		return nil, err
	}
	return buildNodeOutput(record), nil
}

// CreateNode creates one media node.
func (s *serviceImpl) CreateNode(ctx context.Context, in NodeMutationInput) (*NodeMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalized, err := normalizeNodeMutationInput(in)
	if err != nil {
		return nil, err
	}
	exists, err := s.nodeExists(ctx, normalized.NodeNum)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, bizerr.NewCode(CodeMediaNodeDuplicate)
	}

	actorID := int(s.currentActorID(ctx))
	now := gtime.Now()
	_, err = dao.HgNode.Ctx(ctx).Data(do.HgNode{
		NodeNum:    normalized.NodeNum,
		Name:       normalized.Name,
		QnUrl:      normalized.QnUrl,
		BasicUrl:   normalized.BasicUrl,
		DnUrl:      normalized.DnUrl,
		CreatorId:  actorID,
		CreateTime: now,
		UpdaterId:  actorID,
		UpdateTime: now,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeCreateFailed)
	}
	return &NodeMutationOutput{NodeNum: normalized.NodeNum}, nil
}

// UpdateNode updates one media node by old node number.
func (s *serviceImpl) UpdateNode(ctx context.Context, oldNodeNum int, in NodeMutationInput) (*NodeMutationOutput, error) {
	normalizedOldNodeNum, err := normalizeNodeNum(oldNodeNum)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalizedOldNodeNum); err != nil {
		return nil, err
	}
	normalized, err := normalizeNodeMutationInput(in)
	if err != nil {
		return nil, err
	}
	if normalized.NodeNum != normalizedOldNodeNum {
		exists, err := s.nodeExists(ctx, normalized.NodeNum)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, bizerr.NewCode(CodeMediaNodeDuplicate)
		}
	}

	_, err = dao.HgNode.Ctx(ctx).
		Where(do.HgNode{NodeNum: normalizedOldNodeNum}).
		Data(do.HgNode{
			NodeNum:    normalized.NodeNum,
			Name:       normalized.Name,
			QnUrl:      normalized.QnUrl,
			BasicUrl:   normalized.BasicUrl,
			DnUrl:      normalized.DnUrl,
			UpdaterId:  int(s.currentActorID(ctx)),
			UpdateTime: gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeUpdateFailed)
	}
	return &NodeMutationOutput{NodeNum: normalized.NodeNum}, nil
}

// DeleteNode deletes one unreferenced media node.
func (s *serviceImpl) DeleteNode(ctx context.Context, nodeNum int) (*NodeMutationOutput, error) {
	normalizedNodeNum, err := normalizeNodeNum(nodeNum)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalizedNodeNum); err != nil {
		return nil, err
	}
	referenced, err := s.nodeReferenced(ctx, normalizedNodeNum)
	if err != nil {
		return nil, err
	}
	if referenced {
		return nil, bizerr.NewCode(CodeMediaNodeReferenced)
	}

	_, err = dao.HgNode.Ctx(ctx).
		Where(do.HgNode{NodeNum: normalizedNodeNum}).
		Delete()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeDeleteFailed)
	}
	return &NodeMutationOutput{NodeNum: normalizedNodeNum}, nil
}

// ListDeviceNodes returns paged device-node mappings.
func (s *serviceImpl) ListDeviceNodes(ctx context.Context, in ListDeviceNodesInput) (*ListDeviceNodesOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.HgDeviceNode.Columns()
	model := dao.HgDeviceNode.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.DeviceId+" LIKE ? OR "+columns.NodeNum+"::text LIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeCountQueryFailed)
	}

	items := make([]*deviceNodeEntity, 0)
	err = model.
		OrderAsc(columns.DeviceId).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeListQueryFailed)
	}

	nodeNames, err := s.nodeNameMap(ctx, collectDeviceNodeNums(items))
	if err != nil {
		return nil, err
	}
	list := make([]*DeviceNodeOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildDeviceNodeOutput(item, nodeNames))
	}
	return &ListDeviceNodesOutput{List: list, Total: total}, nil
}

// GetDeviceNode returns one device-node mapping by device ID.
func (s *serviceImpl) GetDeviceNode(ctx context.Context, deviceID string) (*DeviceNodeOutput, error) {
	record, err := s.getDeviceNodeEntity(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	nodeNames, err := s.nodeNameMap(ctx, []int{record.NodeNum})
	if err != nil {
		return nil, err
	}
	return buildDeviceNodeOutput(record, nodeNames), nil
}

// CreateDeviceNode creates one device-node mapping.
func (s *serviceImpl) CreateDeviceNode(ctx context.Context, in DeviceNodeMutationInput) (*DeviceNodeMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalized, err := normalizeDeviceNodeMutationInput(in)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalized.NodeNum); err != nil {
		return nil, err
	}
	exists, err := s.deviceNodeExists(ctx, normalized.DeviceId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, bizerr.NewCode(CodeMediaDeviceNodeDuplicate)
	}

	_, err = dao.HgDeviceNode.Ctx(ctx).Data(do.HgDeviceNode{
		DeviceId: normalized.DeviceId,
		NodeNum:  normalized.NodeNum,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeCreateFailed)
	}
	return &DeviceNodeMutationOutput{DeviceId: normalized.DeviceId}, nil
}

// UpdateDeviceNode updates one device-node mapping by old device ID.
func (s *serviceImpl) UpdateDeviceNode(ctx context.Context, oldDeviceID string, in DeviceNodeMutationInput) (*DeviceNodeMutationOutput, error) {
	normalizedOldDeviceID, err := normalizeDeviceNodeKey(oldDeviceID)
	if err != nil {
		return nil, err
	}
	if _, err = s.getDeviceNodeEntity(ctx, normalizedOldDeviceID); err != nil {
		return nil, err
	}
	normalized, err := normalizeDeviceNodeMutationInput(in)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalized.NodeNum); err != nil {
		return nil, err
	}
	if normalized.DeviceId != normalizedOldDeviceID {
		exists, err := s.deviceNodeExists(ctx, normalized.DeviceId)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, bizerr.NewCode(CodeMediaDeviceNodeDuplicate)
		}
	}

	_, err = dao.HgDeviceNode.Ctx(ctx).
		Where(do.HgDeviceNode{DeviceId: normalizedOldDeviceID}).
		Data(do.HgDeviceNode{
			DeviceId: normalized.DeviceId,
			NodeNum:  normalized.NodeNum,
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeUpdateFailed)
	}
	return &DeviceNodeMutationOutput{DeviceId: normalized.DeviceId}, nil
}

// DeleteDeviceNode deletes one device-node mapping.
func (s *serviceImpl) DeleteDeviceNode(ctx context.Context, deviceID string) (*DeviceNodeMutationOutput, error) {
	normalizedDeviceID, err := normalizeDeviceNodeKey(deviceID)
	if err != nil {
		return nil, err
	}
	if _, err = s.getDeviceNodeEntity(ctx, normalizedDeviceID); err != nil {
		return nil, err
	}

	_, err = dao.HgDeviceNode.Ctx(ctx).
		Where(do.HgDeviceNode{DeviceId: normalizedDeviceID}).
		Delete()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeDeleteFailed)
	}
	return &DeviceNodeMutationOutput{DeviceId: normalizedDeviceID}, nil
}

// ListTenantStreamConfigs returns paged tenant stream configs.
func (s *serviceImpl) ListTenantStreamConfigs(ctx context.Context, in ListTenantStreamConfigsInput) (*ListTenantStreamConfigsOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.HgTenantStreamConfig.Columns()
	model := dao.HgTenantStreamConfig.Ctx(ctx)
	if keyword := strings.TrimSpace(in.Keyword); keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.TenantId+" LIKE ? OR "+columns.NodeNum+"::text LIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}
	if in.Enable != nil {
		enable, err := normalizeTenantStreamEnableValue(*in.Enable, TenantStreamEnabled)
		if err != nil {
			return nil, err
		}
		model = model.Where(columns.Enable, enable)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamCountQueryFailed)
	}

	items := make([]*tenantStreamConfigEntity, 0)
	err = model.
		OrderAsc(columns.TenantId).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamListQueryFailed)
	}

	nodeNames, err := s.nodeNameMap(ctx, collectTenantStreamNodeNums(items))
	if err != nil {
		return nil, err
	}
	list := make([]*TenantStreamConfigOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildTenantStreamConfigOutput(item, nodeNames))
	}
	return &ListTenantStreamConfigsOutput{List: list, Total: total}, nil
}

// GetTenantStreamConfig returns one tenant stream config by tenant ID.
func (s *serviceImpl) GetTenantStreamConfig(ctx context.Context, tenantID string) (*TenantStreamConfigOutput, error) {
	record, err := s.getTenantStreamConfigEntity(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	nodeNames, err := s.nodeNameMap(ctx, []int{record.NodeNum})
	if err != nil {
		return nil, err
	}
	return buildTenantStreamConfigOutput(record, nodeNames), nil
}

// CreateTenantStreamConfig creates one tenant stream config.
func (s *serviceImpl) CreateTenantStreamConfig(ctx context.Context, in TenantStreamConfigMutationInput) (*TenantStreamConfigMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalized, err := normalizeTenantStreamConfigMutationInput(in)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalized.NodeNum); err != nil {
		return nil, err
	}
	exists, err := s.tenantStreamConfigExists(ctx, normalized.TenantId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, bizerr.NewCode(CodeMediaTenantStreamDuplicate)
	}

	actorID := int(s.currentActorID(ctx))
	now := gtime.Now()
	_, err = dao.HgTenantStreamConfig.Ctx(ctx).Data(do.HgTenantStreamConfig{
		TenantId:      normalized.TenantId,
		MaxConcurrent: normalized.MaxConcurrent,
		NodeNum:       normalized.NodeNum,
		Enable:        normalized.Enable,
		CreatorId:     actorID,
		CreateTime:    now,
		UpdaterId:     actorID,
		UpdateTime:    now,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamCreateFailed)
	}
	return &TenantStreamConfigMutationOutput{TenantId: normalized.TenantId}, nil
}

// UpdateTenantStreamConfig updates one tenant stream config by old tenant ID.
func (s *serviceImpl) UpdateTenantStreamConfig(ctx context.Context, oldTenantID string, in TenantStreamConfigMutationInput) (*TenantStreamConfigMutationOutput, error) {
	normalizedOldTenantID, err := normalizeTenantStreamConfigKey(oldTenantID)
	if err != nil {
		return nil, err
	}
	if _, err = s.getTenantStreamConfigEntity(ctx, normalizedOldTenantID); err != nil {
		return nil, err
	}
	normalized, err := normalizeTenantStreamConfigMutationInput(in)
	if err != nil {
		return nil, err
	}
	if _, err = s.getNodeEntity(ctx, normalized.NodeNum); err != nil {
		return nil, err
	}
	if normalized.TenantId != normalizedOldTenantID {
		exists, err := s.tenantStreamConfigExists(ctx, normalized.TenantId)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, bizerr.NewCode(CodeMediaTenantStreamDuplicate)
		}
	}

	_, err = dao.HgTenantStreamConfig.Ctx(ctx).
		Where(do.HgTenantStreamConfig{TenantId: normalizedOldTenantID}).
		Data(do.HgTenantStreamConfig{
			TenantId:      normalized.TenantId,
			MaxConcurrent: normalized.MaxConcurrent,
			NodeNum:       normalized.NodeNum,
			Enable:        normalized.Enable,
			UpdaterId:     int(s.currentActorID(ctx)),
			UpdateTime:    gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamUpdateFailed)
	}
	return &TenantStreamConfigMutationOutput{TenantId: normalized.TenantId}, nil
}

// DeleteTenantStreamConfig deletes one tenant stream config.
func (s *serviceImpl) DeleteTenantStreamConfig(ctx context.Context, tenantID string) (*TenantStreamConfigMutationOutput, error) {
	normalizedTenantID, err := normalizeTenantStreamConfigKey(tenantID)
	if err != nil {
		return nil, err
	}
	if _, err = s.getTenantStreamConfigEntity(ctx, normalizedTenantID); err != nil {
		return nil, err
	}

	_, err = dao.HgTenantStreamConfig.Ctx(ctx).
		Where(do.HgTenantStreamConfig{TenantId: normalizedTenantID}).
		Delete()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamDeleteFailed)
	}
	return &TenantStreamConfigMutationOutput{TenantId: normalizedTenantID}, nil
}

// nodeExists reports whether one media node exists.
func (s *serviceImpl) nodeExists(ctx context.Context, nodeNum int) (bool, error) {
	count, err := dao.HgNode.Ctx(ctx).
		Where(do.HgNode{NodeNum: nodeNum}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaNodeDetailQueryFailed)
	}
	return count > 0, nil
}

// getNodeEntity returns one media node entity by node number.
func (s *serviceImpl) getNodeEntity(ctx context.Context, nodeNum int) (*nodeEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedNodeNum, err := normalizeNodeNum(nodeNum)
	if err != nil {
		return nil, err
	}

	var record *nodeEntity
	err = dao.HgNode.Ctx(ctx).
		Where(do.HgNode{NodeNum: normalizedNodeNum}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaNodeNotFound)
	}
	return record, nil
}

// nodeReferenced reports whether one node number is referenced by dependent tables.
func (s *serviceImpl) nodeReferenced(ctx context.Context, nodeNum int) (bool, error) {
	deviceCount, err := dao.HgDeviceNode.Ctx(ctx).
		Where(do.HgDeviceNode{NodeNum: nodeNum}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaDeviceNodeCountQueryFailed)
	}
	if deviceCount > 0 {
		return true, nil
	}

	streamCount, err := dao.HgTenantStreamConfig.Ctx(ctx).
		Where(do.HgTenantStreamConfig{NodeNum: nodeNum}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaTenantStreamCountQueryFailed)
	}
	return streamCount > 0, nil
}

// deviceNodeExists reports whether one device-node mapping exists.
func (s *serviceImpl) deviceNodeExists(ctx context.Context, deviceID string) (bool, error) {
	count, err := dao.HgDeviceNode.Ctx(ctx).
		Where(do.HgDeviceNode{DeviceId: deviceID}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaDeviceNodeDetailQueryFailed)
	}
	return count > 0, nil
}

// getDeviceNodeEntity returns one device-node entity by device ID.
func (s *serviceImpl) getDeviceNodeEntity(ctx context.Context, deviceID string) (*deviceNodeEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedDeviceID, err := normalizeDeviceNodeKey(deviceID)
	if err != nil {
		return nil, err
	}

	var record *deviceNodeEntity
	err = dao.HgDeviceNode.Ctx(ctx).
		Where(do.HgDeviceNode{DeviceId: normalizedDeviceID}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaDeviceNodeDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaDeviceNodeNotFound)
	}
	return record, nil
}

// tenantStreamConfigExists reports whether one tenant stream config exists.
func (s *serviceImpl) tenantStreamConfigExists(ctx context.Context, tenantID string) (bool, error) {
	count, err := dao.HgTenantStreamConfig.Ctx(ctx).
		Where(do.HgTenantStreamConfig{TenantId: tenantID}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaTenantStreamDetailQueryFailed)
	}
	return count > 0, nil
}

// getTenantStreamConfigEntity returns one tenant stream config entity by tenant ID.
func (s *serviceImpl) getTenantStreamConfigEntity(ctx context.Context, tenantID string) (*tenantStreamConfigEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedTenantID, err := normalizeTenantStreamConfigKey(tenantID)
	if err != nil {
		return nil, err
	}

	var record *tenantStreamConfigEntity
	err = dao.HgTenantStreamConfig.Ctx(ctx).
		Where(do.HgTenantStreamConfig{TenantId: normalizedTenantID}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantStreamDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaTenantStreamNotFound)
	}
	return record, nil
}

// nodeNameMap returns node names by node number.
func (s *serviceImpl) nodeNameMap(ctx context.Context, nodeNums []int) (map[int]string, error) {
	result := make(map[int]string)
	uniqueNodeNums := uniqueInts(nodeNums)
	if len(uniqueNodeNums) == 0 {
		return result, nil
	}

	nodes := make([]*nodeEntity, 0)
	err := dao.HgNode.Ctx(ctx).
		Fields(dao.HgNode.Columns().NodeNum, dao.HgNode.Columns().Name).
		WhereIn(dao.HgNode.Columns().NodeNum, uniqueNodeNums).
		Scan(&nodes)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaNodeListQueryFailed)
	}
	for _, node := range nodes {
		if node == nil {
			continue
		}
		result[node.NodeNum] = node.Name
	}
	return result, nil
}

// normalizeNodeMutationInput validates media node mutation input.
func normalizeNodeMutationInput(in NodeMutationInput) (NodeMutationInput, error) {
	nodeNum, err := normalizeNodeNum(in.NodeNum)
	if err != nil {
		return NodeMutationInput{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return NodeMutationInput{}, bizerr.NewCode(CodeMediaNodeNameRequired)
	}
	if utf8.RuneCountInString(name) > 32 {
		return NodeMutationInput{}, bizerr.NewCode(CodeMediaNodeNameTooLong)
	}

	qnURL, err := normalizeNodeURL(in.QnUrl)
	if err != nil {
		return NodeMutationInput{}, err
	}
	basicURL, err := normalizeNodeURL(in.BasicUrl)
	if err != nil {
		return NodeMutationInput{}, err
	}
	dnURL, err := normalizeNodeURL(in.DnUrl)
	if err != nil {
		return NodeMutationInput{}, err
	}
	return NodeMutationInput{
		NodeNum:  nodeNum,
		Name:     name,
		QnUrl:    qnURL,
		BasicUrl: basicURL,
		DnUrl:    dnURL,
	}, nil
}

// normalizeNodeNum validates one node number.
func normalizeNodeNum(nodeNum int) (int, error) {
	if nodeNum < 0 || nodeNum > 255 {
		return 0, bizerr.NewCode(CodeMediaNodeNumInvalid)
	}
	return nodeNum, nil
}

// normalizeNodeURL validates and trims one node gateway URL field.
func normalizeNodeURL(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaNodeURLRequired)
	}
	if utf8.RuneCountInString(normalized) > 255 {
		return "", bizerr.NewCode(CodeMediaNodeURLTooLong)
	}
	return normalized, nil
}

// normalizeDeviceNodeMutationInput validates device-node mutation input.
func normalizeDeviceNodeMutationInput(in DeviceNodeMutationInput) (DeviceNodeMutationInput, error) {
	deviceID, err := normalizeDeviceNodeKey(in.DeviceId)
	if err != nil {
		return DeviceNodeMutationInput{}, err
	}
	nodeNum, err := normalizeNodeNum(in.NodeNum)
	if err != nil {
		return DeviceNodeMutationInput{}, err
	}
	return DeviceNodeMutationInput{DeviceId: deviceID, NodeNum: nodeNum}, nil
}

// normalizeDeviceNodeKey validates one device-node natural key.
func normalizeDeviceNodeKey(deviceID string) (string, error) {
	normalized := strings.TrimSpace(deviceID)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaDeviceNodeDeviceRequired)
	}
	if utf8.RuneCountInString(normalized) > 64 {
		return "", bizerr.NewCode(CodeMediaDeviceNodeDeviceTooLong)
	}
	return normalized, nil
}

// normalizeTenantStreamConfigMutationInput validates tenant stream config mutation input.
func normalizeTenantStreamConfigMutationInput(in TenantStreamConfigMutationInput) (TenantStreamConfigMutationInput, error) {
	tenantID, err := normalizeTenantStreamConfigKey(in.TenantId)
	if err != nil {
		return TenantStreamConfigMutationInput{}, err
	}
	if in.MaxConcurrent < 0 {
		return TenantStreamConfigMutationInput{}, bizerr.NewCode(CodeMediaTenantStreamMaxConcurrentInvalid)
	}
	nodeNum, err := normalizeNodeNum(in.NodeNum)
	if err != nil {
		return TenantStreamConfigMutationInput{}, err
	}
	enable, err := normalizeTenantStreamEnableValue(in.Enable, TenantStreamEnabled)
	if err != nil {
		return TenantStreamConfigMutationInput{}, err
	}
	return TenantStreamConfigMutationInput{
		TenantId:      tenantID,
		MaxConcurrent: in.MaxConcurrent,
		NodeNum:       nodeNum,
		Enable:        enable,
	}, nil
}

// normalizeTenantStreamConfigKey validates one tenant stream config natural key.
func normalizeTenantStreamConfigKey(tenantID string) (string, error) {
	normalized := strings.TrimSpace(tenantID)
	if normalized == "" {
		return "", bizerr.NewCode(CodeMediaTenantStreamTenantRequired)
	}
	if utf8.RuneCountInString(normalized) > 64 {
		return "", bizerr.NewCode(CodeMediaTenantStreamTenantTooLong)
	}
	return normalized, nil
}

// normalizeTenantStreamEnableValue validates and normalizes tenant stream enable value.
func normalizeTenantStreamEnableValue(value int, defaultValue TenantStreamEnableValue) (int, error) {
	if value == 0 && defaultValue == TenantStreamEnabled {
		return int(TenantStreamDisabled), nil
	}
	switch TenantStreamEnableValue(value) {
	case TenantStreamEnabled, TenantStreamDisabled:
		return value, nil
	default:
		return 0, bizerr.NewCode(CodeMediaTenantStreamEnableInvalid)
	}
}

// buildNodeOutput converts one generated node entity into service output.
func buildNodeOutput(item *nodeEntity) *NodeOutput {
	if item == nil {
		return &NodeOutput{}
	}
	return &NodeOutput{
		Id:         item.Id,
		NodeNum:    item.NodeNum,
		Name:       item.Name,
		QnUrl:      item.QnUrl,
		BasicUrl:   item.BasicUrl,
		DnUrl:      item.DnUrl,
		CreatorId:  item.CreatorId,
		CreateTime: formatTime(item.CreateTime),
		UpdaterId:  item.UpdaterId,
		UpdateTime: formatTime(item.UpdateTime),
	}
}

// buildDeviceNodeOutput converts one generated device-node entity into service output.
func buildDeviceNodeOutput(item *deviceNodeEntity, nodeNames map[int]string) *DeviceNodeOutput {
	if item == nil {
		return &DeviceNodeOutput{}
	}
	return &DeviceNodeOutput{
		DeviceId: item.DeviceId,
		NodeNum:  item.NodeNum,
		NodeName: nodeNames[item.NodeNum],
	}
}

// buildTenantStreamConfigOutput converts one generated tenant stream config entity into service output.
func buildTenantStreamConfigOutput(item *tenantStreamConfigEntity, nodeNames map[int]string) *TenantStreamConfigOutput {
	if item == nil {
		return &TenantStreamConfigOutput{}
	}
	return &TenantStreamConfigOutput{
		TenantId:      item.TenantId,
		MaxConcurrent: item.MaxConcurrent,
		NodeNum:       item.NodeNum,
		NodeName:      nodeNames[item.NodeNum],
		Enable:        item.Enable,
		CreatorId:     item.CreatorId,
		CreateTime:    formatTime(item.CreateTime),
		UpdaterId:     item.UpdaterId,
		UpdateTime:    formatTime(item.UpdateTime),
	}
}

// collectDeviceNodeNums collects unique node numbers from device-node rows.
func collectDeviceNodeNums(rows []*deviceNodeEntity) []int {
	nodeNums := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		nodeNums = append(nodeNums, row.NodeNum)
	}
	return uniqueInts(nodeNums)
}

// collectTenantStreamNodeNums collects unique node numbers from tenant stream rows.
func collectTenantStreamNodeNums(rows []*tenantStreamConfigEntity) []int {
	nodeNums := make([]int, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		nodeNums = append(nodeNums, row.NodeNum)
	}
	return uniqueInts(nodeNums)
}

// uniqueInts removes duplicate integer values while preserving first-seen order.
func uniqueInts(values []int) []int {
	result := make([]int, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
