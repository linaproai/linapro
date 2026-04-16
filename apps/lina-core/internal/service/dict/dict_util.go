package dict

import (
	"context"
	"strconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// GetLabelByValueInput defines input for GetLabelByValue function.
type GetLabelByValueInput struct {
	DictType string // Dictionary type
	Value    string // Dictionary value
}

// GetLabelByValue retrieves the label for a given dict type and value.
// Returns the value itself if not found or on error.
func (s *serviceImpl) GetLabelByValue(ctx context.Context, in GetLabelByValueInput) string {
	if in.DictType == "" || in.Value == "" {
		return in.Value
	}

	var dictData *entity.SysDictData
	err := dao.SysDictData.Ctx(ctx).
		Where(do.SysDictData{DictType: in.DictType, Value: in.Value, Status: 1}).
		Scan(&dictData)
	if err != nil || dictData == nil {
		return in.Value
	}
	return dictData.Label
}

// GetLabelByIntValue retrieves the label for a given dict type and integer value.
// Converts the integer to string and calls GetLabelByValue.
func (s *serviceImpl) GetLabelByIntValue(ctx context.Context, dictType string, value int) string {
	return s.GetLabelByValue(ctx, GetLabelByValueInput{
		DictType: dictType,
		Value:    strconv.Itoa(value),
	})
}

// BuildLabelMap builds a map of value to label for a given dict type.
// Useful for batch lookups to avoid repeated database queries.
func (s *serviceImpl) BuildLabelMap(ctx context.Context, dictType string) map[string]string {
	list, err := s.DataByType(ctx, dictType)
	if err != nil || len(list) == 0 {
		return make(map[string]string)
	}

	labelMap := make(map[string]string, len(list))
	for _, item := range list {
		labelMap[item.Value] = item.Label
	}
	return labelMap
}

// BuildIntLabelMap builds a map of integer value to label for a given dict type.
func (s *serviceImpl) BuildIntLabelMap(ctx context.Context, dictType string) map[int]string {
	list, err := s.DataByType(ctx, dictType)
	if err != nil || len(list) == 0 {
		return make(map[int]string)
	}

	labelMap := make(map[int]string, len(list))
	for _, item := range list {
		if val, err := strconv.Atoi(item.Value); err == nil {
			labelMap[val] = item.Label
		}
	}
	return labelMap
}
