package file

import (
	"context"

	v1 "lina-core/api/file/v1"
)

// FileSuffixes returns file suffix list
func (c *ControllerV1) FileSuffixes(ctx context.Context, req *v1.FileSuffixesReq) (res *v1.FileSuffixesRes, err error) {
	out, err := c.fileSvc.Suffixes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.FileSuffixItem, len(out))
	for i, item := range out {
		items[i] = &v1.FileSuffixItem{
			Value: item.Value,
			Label: item.Label,
		}
	}
	return &v1.FileSuffixesRes{
		List: items,
	}, nil
}
