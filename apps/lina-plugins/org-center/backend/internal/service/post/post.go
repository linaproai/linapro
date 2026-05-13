// Package post implements post management for the org-center source plugin.
// It owns plugin_org_center_post CRUD, select-option queries, and tree/export helpers without
// relying on host-internal post services.
package post

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/xuri/excelize/v2"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/excelutil"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-org-center/backend/internal/dao"
	"lina-plugin-org-center/backend/internal/model/do"
	entitymodel "lina-plugin-org-center/backend/internal/model/entity"
)

// Table and column constants for post management.
const (
	colPostID      = "id"
	colPostDeptID  = "dept_id"
	colPostCode    = "code"
	colPostName    = "name"
	colPostSort    = "sort"
	colPostStatus  = "status"
	colPostRemark  = "remark"
	colPostCreated = "created_at"

	colDeptID       = "id"
	colDeptParentID = "parent_id"
	colDeptName     = "name"
	colDeptOrderNum = "order_num"

	colUserPostPostID = "post_id"
)

// Runtime i18n keys for backend-owned post projections and exports.
const (
	postTreeUnassignedDeptKey     = "plugin.org-center.post.tree.unassignedDept"
	postExportHeaderCodeKey       = "plugin.org-center.post.export.headers.code"
	postExportHeaderNameKey       = "plugin.org-center.post.export.headers.name"
	postExportHeaderSortKey       = "plugin.org-center.post.export.headers.sort"
	postExportHeaderStatusKey     = "plugin.org-center.post.export.headers.status"
	postExportHeaderRemarkKey     = "plugin.org-center.post.export.headers.remark"
	postExportHeaderCreatedAtKey  = "plugin.org-center.post.export.headers.createdAt"
	postExportStatusEnabledKey    = "plugin.org-center.post.export.status.enabled"
	postExportStatusDisabledKey   = "plugin.org-center.post.export.status.disabled"
	postTreeUnassignedDeptDefault = "Unassigned"
)

// Service defines the post service contract.
type Service interface {
	// List queries the paged post list.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates one post record.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetByID retrieves one post detail by primary key.
	GetByID(ctx context.Context, id int) (*PostEntity, error)
	// Update updates one post record.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes one or more posts.
	Delete(ctx context.Context, ids string) error
	// DeptTree returns the department tree decorated with post counts.
	DeptTree(ctx context.Context) ([]*DeptTreeNode, error)
	// OptionSelect returns post options for one department subtree.
	OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error)
	// Export generates one Excel file for the filtered post set.
	Export(ctx context.Context, in ExportInput) ([]byte, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	i18nSvc      plugincontract.I18nService         // i18nSvc resolves plugin runtime translations.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned post rows.
}

// New creates and returns a new post service instance.
func New(i18nSvc plugincontract.I18nService, tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{
		i18nSvc:      i18nSvc,
		tenantFilter: tenantFilter,
	}
}

// PostEntity mirrors the plugin-local generated plugin_org_center_post entity.
type PostEntity = entitymodel.Post

// ListInput defines filters for post list queries.
type ListInput struct {
	PageNum  int
	PageSize int
	DeptId   *int
	Code     string
	Name     string
	Status   *int
}

// ListOutput defines the paged post result.
type ListOutput struct {
	List  []*PostEntity
	Total int
}

// CreateInput defines the create-post input.
type CreateInput struct {
	DeptId int
	Code   string
	Name   string
	Sort   int
	Status int
	Remark string
}

// UpdateInput defines the update-post input.
type UpdateInput struct {
	Id     int
	DeptId *int
	Code   *string
	Name   *string
	Sort   *int
	Status *int
	Remark *string
}

// DeptTreeNode represents one post-filter department node.
type DeptTreeNode struct {
	Id        int             `json:"id"`
	Label     string          `json:"label"`
	PostCount int             `json:"postCount"`
	Children  []*DeptTreeNode `json:"children"`
}

// PostOption represents one selectable post row.
type PostOption struct {
	PostId   int    `json:"postId"`
	PostName string `json:"postName"`
}

// OptionSelectInput defines the option-select input.
type OptionSelectInput struct {
	DeptId *int
}

// ExportInput defines the export filters.
type ExportInput struct {
	DeptId *int
	Code   string
	Name   string
	Status *int
}

// deptRow reuses the plugin-local generated plugin_org_center_dept entity.
type deptRow = entitymodel.Dept

// deptCountRow is the grouped post-count projection keyed by department.
type deptCountRow struct {
	DeptId int `json:"deptId"`
	Cnt    int `json:"cnt"`
}

// List queries the paged post list.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx))
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			model = model.Where(colPostDeptID, 0)
		} else {
			deptIDs, err := s.descendantDeptIDs(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			model = model.WhereIn(colPostDeptID, deptIDs)
		}
	}
	if in.Code != "" {
		model = model.WhereLike(colPostCode, "%"+in.Code+"%")
	}
	if in.Name != "" {
		model = model.WhereLike(colPostName, "%"+in.Name+"%")
	}
	if in.Status != nil {
		model = model.Where(colPostStatus, *in.Status)
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*PostEntity, 0)
	err = model.Page(in.PageNum, in.PageSize).OrderAsc(colPostSort).Scan(&list)
	if err != nil {
		return nil, err
	}
	return &ListOutput{List: list, Total: total}, nil
}

// Create creates one post record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if err := s.checkCodeUnique(ctx, in.Code, 0); err != nil {
		return 0, err
	}
	id, err := dao.Post.Ctx(ctx).Data(do.Post{
		TenantId: s.tenantFilter.Current(ctx),
		DeptId:   in.DeptId,
		Code:     in.Code,
		Name:     in.Name,
		Sort:     in.Sort,
		Status:   in.Status,
		Remark:   in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// GetByID retrieves one post detail by primary key.
func (s *serviceImpl) GetByID(ctx context.Context, id int) (*PostEntity, error) {
	var post *PostEntity
	err := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).Where(colPostID, id).Scan(&post)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, bizerr.NewCode(CodePostNotFound)
	}
	return post, nil
}

// Update updates one post record.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	if _, err := s.GetByID(ctx, in.Id); err != nil {
		return err
	}
	data := do.Post{}
	if in.DeptId != nil {
		data.DeptId = *in.DeptId
	}
	if in.Code != nil {
		if err := s.checkCodeUnique(ctx, *in.Code, in.Id); err != nil {
			return err
		}
		data.Code = *in.Code
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}
	_, err := dao.Post.Ctx(ctx).
		OmitNilData().
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		Where(colPostID, in.Id).
		Data(data).
		Update()
	return err
}

// Delete deletes one or more posts.
func (s *serviceImpl) Delete(ctx context.Context, ids string) error {
	idList := gstr.SplitAndTrim(ids, ",")
	if len(idList) == 0 {
		return bizerr.NewCode(CodePostDeleteRequired)
	}

	validIDs := make([]int, 0, len(idList))
	for _, idStr := range idList {
		id := gconv.Int(idStr)
		if id == 0 {
			continue
		}
		count, err := s.tenantFilter.Apply(ctx, dao.UserPost.Ctx(ctx)).
			Where(colUserPostPostID, id).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return bizerr.NewCode(CodePostAssignedDeleteDenied, bizerr.P("id", id))
		}
		validIDs = append(validIDs, id)
	}
	if len(validIDs) == 0 {
		return bizerr.NewCode(CodePostValidIDRequired)
	}
	_, err := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).WhereIn(colPostID, validIDs).Delete()
	return err
}

// DeptTree returns the department tree decorated with post counts.
func (s *serviceImpl) DeptTree(ctx context.Context) ([]*DeptTreeNode, error) {
	deptList := make([]*deptRow, 0)
	err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).OrderAsc(colDeptOrderNum).Scan(&deptList)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[int]*DeptTreeNode, len(deptList))
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		nodeMap[deptItem.Id] = &DeptTreeNode{Id: deptItem.Id, Label: deptItem.Name, Children: make([]*DeptTreeNode, 0)}
	}
	roots := make([]*DeptTreeNode, 0)
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		node := nodeMap[deptItem.Id]
		if parent, ok := nodeMap[deptItem.ParentId]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}

	unassignedLabel := s.translate(ctx, postTreeUnassignedDeptKey, postTreeUnassignedDeptDefault)
	unassignedNode := &DeptTreeNode{Id: 0, Label: unassignedLabel, Children: make([]*DeptTreeNode, 0)}
	roots = append(roots, unassignedNode)

	counts := make([]deptCountRow, 0)
	err = s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).Fields("dept_id, COUNT(*) as cnt").Group("dept_id").Scan(&counts)
	if err != nil {
		return nil, err
	}
	countMap := make(map[int]int, len(counts))
	for _, item := range counts {
		countMap[item.DeptId] = item.Cnt
	}

	var applyCount func(nodes []*DeptTreeNode)
	applyCount = func(nodes []*DeptTreeNode) {
		for _, node := range nodes {
			if node == nil {
				continue
			}
			applyCount(node.Children)
			node.PostCount = countMap[node.Id]
			for _, child := range node.Children {
				if child == nil {
					continue
				}
				node.PostCount += child.PostCount
			}
			node.Label = fmt.Sprintf("%s(%d)", node.Label, node.PostCount)
		}
	}
	applyCount(roots[:len(roots)-1])
	unassignedNode.PostCount = countMap[0]
	unassignedNode.Label = fmt.Sprintf("%s(%d)", unassignedLabel, unassignedNode.PostCount)
	return roots, nil
}

// OptionSelect returns post options for one department subtree.
func (s *serviceImpl) OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error) {
	model := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).Where(colPostStatus, 1)
	if in.DeptId != nil {
		deptIDs, err := s.descendantDeptIDs(ctx, *in.DeptId)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(colPostDeptID, deptIDs)
	}
	list := make([]*PostEntity, 0)
	if err := model.OrderAsc(colPostSort).Scan(&list); err != nil {
		return nil, err
	}
	options := make([]PostOption, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		options = append(options, PostOption{PostId: item.Id, PostName: item.Name})
	}
	return options, nil
}

// Export generates one Excel file for the filtered post set.
func (s *serviceImpl) Export(ctx context.Context, in ExportInput) (data []byte, err error) {
	model := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx))
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			model = model.Where(colPostDeptID, 0)
		} else {
			deptIDs, err := s.descendantDeptIDs(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			model = model.WhereIn(colPostDeptID, deptIDs)
		}
	}
	if in.Code != "" {
		model = model.WhereLike(colPostCode, "%"+in.Code+"%")
	}
	if in.Name != "" {
		model = model.WhereLike(colPostName, "%"+in.Name+"%")
	}
	if in.Status != nil {
		model = model.Where(colPostStatus, *in.Status)
	}

	list := make([]*PostEntity, 0)
	if err := model.OrderAsc(colPostSort).Scan(&list); err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	defer excelutil.CloseFile(ctx, file, &err)
	sheet := "Sheet1"
	headers := s.exportHeaders(ctx)
	for index, header := range headers {
		if err = excelutil.SetCellValue(file, sheet, index+1, 1, header); err != nil {
			return nil, err
		}
	}
	for index, item := range list {
		if item == nil {
			continue
		}
		row := index + 2
		if err = excelutil.SetCellValue(file, sheet, 1, row, item.Code); err != nil {
			return nil, err
		}
		if err = excelutil.SetCellValue(file, sheet, 2, row, item.Name); err != nil {
			return nil, err
		}
		if err = excelutil.SetCellValue(file, sheet, 3, row, item.Sort); err != nil {
			return nil, err
		}
		statusText := s.exportStatusText(ctx, item.Status)
		if err = excelutil.SetCellValue(file, sheet, 4, row, statusText); err != nil {
			return nil, err
		}
		if err = excelutil.SetCellValue(file, sheet, 5, row, item.Remark); err != nil {
			return nil, err
		}
		if item.CreatedAt != nil {
			if err = excelutil.SetCellValue(file, sheet, 6, row, item.CreatedAt.String()); err != nil {
				return nil, err
			}
		}
	}
	var buf bytes.Buffer
	if err = file.Write(&buf); err != nil {
		return nil, err
	}
	data = buf.Bytes()
	return data, nil
}

// exportHeaders returns localized Excel headers for post export.
func (s *serviceImpl) exportHeaders(ctx context.Context) []string {
	return []string{
		s.translate(ctx, postExportHeaderCodeKey, "Post Code"),
		s.translate(ctx, postExportHeaderNameKey, "Post Name"),
		s.translate(ctx, postExportHeaderSortKey, "Sort"),
		s.translate(ctx, postExportHeaderStatusKey, "Status"),
		s.translate(ctx, postExportHeaderRemarkKey, "Remark"),
		s.translate(ctx, postExportHeaderCreatedAtKey, "Created At"),
	}
}

// exportStatusText returns the localized export label for one post status.
func (s *serviceImpl) exportStatusText(ctx context.Context, status int) string {
	if status == 0 {
		return s.translate(ctx, postExportStatusDisabledKey, "Disabled")
	}
	return s.translate(ctx, postExportStatusEnabledKey, "Enabled")
}

// descendantDeptIDs returns the given department plus all descendants.
func (s *serviceImpl) descendantDeptIDs(ctx context.Context, deptID int) ([]int, error) {
	deptIDs := []int{deptID}
	parentIDs := []int{deptID}
	for len(parentIDs) > 0 {
		childValues, err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
			WhereIn(colDeptParentID, parentIDs).
			Fields(colDeptID).
			Array()
		if err != nil {
			return nil, err
		}
		childIDs := gconv.Ints(childValues)
		deptIDs = append(deptIDs, childIDs...)
		parentIDs = childIDs
	}
	return deptIDs, nil
}

// checkCodeUnique checks whether one post code already exists.
func (s *serviceImpl) checkCodeUnique(ctx context.Context, code string, excludeID int) error {
	model := s.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).Where(colPostCode, code)
	if excludeID > 0 {
		model = model.WhereNot(colPostID, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodePostCodeExists)
	}
	return nil
}

// translate resolves one plugin runtime i18n key and falls back to English
// source text when the current language bundle does not define it.
func (s *serviceImpl) translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.i18nSvc == nil || strings.TrimSpace(key) == "" {
		return fallback
	}
	return s.i18nSvc.Translate(ctx, key, fallback)
}
