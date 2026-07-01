// This file handles the runtime i18n locale-list endpoint.

package i18n

import (
	"context"

	v1 "lina-core/api/i18n/v1"
)

// RuntimeLocales returns the runtime locale descriptors exposed by the host.
func (c *ControllerV1) RuntimeLocales(ctx context.Context, req *v1.RuntimeLocalesReq) (res *v1.RuntimeLocalesRes, err error) {
	var (
		locale = c.i18nSvc.ResolveLocale(ctx, req.Lang)
		output = c.i18nSvc.RuntimeLocales(ctx, locale)
		items  = make([]v1.RuntimeLocaleItem, 0, len(output.Items))
	)
	for _, descriptor := range output.Items {
		items = append(items, v1.RuntimeLocaleItem{
			Locale:     descriptor.Locale,
			Name:       descriptor.Name,
			NativeName: descriptor.NativeName,
			Direction:  v1.LocaleDirection(descriptor.Direction),
			IsDefault:  descriptor.IsDefault,
		})
	}

	return &v1.RuntimeLocalesRes{
		Locale:  locale,
		Enabled: output.Enabled,
		Items:   items,
	}, nil
}
