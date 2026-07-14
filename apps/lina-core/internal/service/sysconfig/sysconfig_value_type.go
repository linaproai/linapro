// This file normalizes and validates sys_config value types and options for
// create, update, and import write paths.

package sysconfig

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/configvaluetype"
)

// resolveCreateTypeMetadata resolves value type and options for create, with
// optional inheritance from the platform row when the tenant creates an override.
func resolveCreateTypeMetadata(
	ctx context.Context,
	key string,
	valueTypeRaw string,
	options []configvaluetype.Option,
) (valueType configvaluetype.Code, optionsRaw string, err error) {
	valueType, err = configvaluetype.ResolveCode(valueTypeRaw)
	if err != nil {
		return "", "", bizerr.WrapCode(err, CodeSysConfigValueTypeInvalid)
	}
	optionsRaw, err = configvaluetype.EncodeOptions(options)
	if err != nil {
		return "", "", bizerr.WrapCode(err, CodeSysConfigOptionsInvalid)
	}

	// Inherit from platform default when tenant creates a same-key override and
	// did not provide type/options explicitly.
	if strings.TrimSpace(valueTypeRaw) == "" || (len(options) == 0 && optionsRaw == "") {
		platform, lookupErr := lookupPlatformConfigByKey(ctx, key)
		if lookupErr != nil {
			return "", "", lookupErr
		}
		if platform != nil {
			if strings.TrimSpace(valueTypeRaw) == "" {
				inherited, resolveErr := configvaluetype.ResolveCode(platform.ValueType)
				if resolveErr != nil {
					valueType = configvaluetype.Text
				} else {
					valueType = inherited
				}
			}
			if len(options) == 0 {
				optionsRaw = platform.Options
			}
		}
	}
	return valueType, optionsRaw, nil
}

// lookupPlatformConfigByKey returns the platform-scope row for key when present.
func lookupPlatformConfigByKey(ctx context.Context, key string) (*entity.SysConfig, error) {
	var row *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Where(do.SysConfig{
			TenantId: datascope.PlatformTenantID,
			Key:      key,
		}).
		Scan(&row)
	if err != nil {
		return nil, err
	}
	return row, nil
}

// validateTypedConfigValue validates value against type and options, wrapping
// domain error codes for API callers.
func validateTypedConfigValue(valueType configvaluetype.Code, optionsRaw string, value string) error {
	finalType := valueType
	if valueType == configvaluetype.MultiSelect {
		value = configvaluetype.NormalizeMultiSelectValue(value)
	}
	if err := configvaluetype.ValidateTypedValue(finalType, optionsRaw, value); err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "unsupported config value type"):
			return bizerr.WrapCode(err, CodeSysConfigValueTypeInvalid)
		case strings.Contains(msg, "invalid config options JSON"),
			strings.Contains(msg, "requires non-empty options"),
			strings.Contains(msg, "options[") && strings.Contains(msg, "value is required"):
			return bizerr.WrapCode(err, CodeSysConfigOptionsInvalid)
		default:
			return bizerr.WrapCode(err, CodeSysConfigTypedValueInvalid)
		}
	}
	return nil
}

// normalizePersistedValue applies type-specific value normalization before write.
func normalizePersistedValue(valueType configvaluetype.Code, value string) string {
	if valueType == configvaluetype.MultiSelect {
		return configvaluetype.NormalizeMultiSelectValue(value)
	}
	if valueType == configvaluetype.Number {
		return strings.TrimSpace(value)
	}
	return value
}

// parseEntityOptions converts the stored options JSON into API-friendly options.
func parseEntityOptions(raw string) []configvaluetype.Option {
	options, err := configvaluetype.ParseOptions(raw)
	if err != nil || len(options) == 0 {
		return []configvaluetype.Option{}
	}
	return options
}

// entityValueType returns a supported value type, defaulting unknown to text.
func entityValueType(raw string) configvaluetype.Code {
	return configvaluetype.Normalize(raw)
}
