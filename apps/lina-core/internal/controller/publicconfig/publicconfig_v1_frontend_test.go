// This file verifies public frontend config localization boundaries.

package publicconfig

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/os/gctx"

	v1 "lina-core/api/publicconfig/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	hostconfig "lina-core/internal/service/config"
	i18nsvc "lina-core/internal/service/i18n"
)

// TestFrontendProjectsLocalizedSeedCopy verifies built-in public frontend seed
// copy is projected to the requested locale at the consumption endpoint.
func TestFrontendProjectsLocalizedSeedCopy(t *testing.T) {
	ctx := newEnglishPublicConfigCtx()
	backgroundCtx := context.Background()
	ensureConfigRecordState(
		t,
		backgroundCtx,
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"面向可持续交付的 AI 驱动全栈框架",
		"控制登录页顶部主标题文案。",
	)
	ensureConfigRecordState(
		t,
		backgroundCtx,
		hostconfig.PublicFrontendSettingKeyAuthPageDesc,
		"登录展示-页面说明",
		"帮助团队快速交付生产级应用，同时保持架构、权限、测试与扩展边界可治理",
		"控制登录页顶部说明文案。",
	)
	ensureConfigRecordState(
		t,
		backgroundCtx,
		hostconfig.PublicFrontendSettingKeyAuthLoginSubtitle,
		"登录展示-登录副标题",
		"请输入您的帐户信息以开始管理您的项目",
		"控制登录表单上方的提示说明文案。",
	)

	controller := &ControllerV1{
		configSvc: hostconfig.New(),
		i18nSvc:   i18nsvc.New(),
	}

	res, err := controller.Frontend(ctx, &v1.FrontendReq{})
	if err != nil {
		t.Fatalf("load public frontend config: %v", err)
	}
	if res.Auth.PageTitle != "AI-driven full-stack development framework" {
		t.Fatalf("expected localized page title, got %q", res.Auth.PageTitle)
	}
	if res.Auth.PageDesc != "Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model" {
		t.Fatalf("expected localized page description, got %q", res.Auth.PageDesc)
	}
	if res.Auth.LoginSubtitle != "Enter your account credentials to start managing your projects" {
		t.Fatalf("expected localized login subtitle, got %q", res.Auth.LoginSubtitle)
	}
}

// TestFrontendKeepsCustomizedCopy verifies customized public frontend copy is
// not replaced by runtime baseline translations.
func TestFrontendKeepsCustomizedCopy(t *testing.T) {
	ctx := newEnglishPublicConfigCtx()
	backgroundCtx := context.Background()
	ensureConfigRecordState(
		t,
		backgroundCtx,
		hostconfig.PublicFrontendSettingKeyAuthPageTitle,
		"登录展示-页面标题",
		"自定义中文标题",
		"控制登录页顶部主标题文案。",
	)

	controller := &ControllerV1{
		configSvc: hostconfig.New(),
		i18nSvc:   i18nsvc.New(),
	}

	res, err := controller.Frontend(ctx, &v1.FrontendReq{})
	if err != nil {
		t.Fatalf("load customized public frontend config: %v", err)
	}
	if res.Auth.PageTitle != "自定义中文标题" {
		t.Fatalf("expected customized page title to remain raw, got %q", res.Auth.PageTitle)
	}
}

// newEnglishPublicConfigCtx builds one English request context for controller tests.
func newEnglishPublicConfigCtx() context.Context {
	return context.WithValue(
		context.Background(),
		gctx.StrKey("BizCtx"),
		&model.Context{Locale: i18nsvc.EnglishLocale},
	)
}

// ensureConfigRecordState forces one protected config row into the requested
// raw state and restores the original value after the test completes.
func ensureConfigRecordState(
	t *testing.T,
	ctx context.Context,
	key string,
	name string,
	value string,
	remark string,
) {
	t.Helper()

	existing, err := queryConfigRecord(ctx, key)
	if err != nil {
		t.Fatalf("query config record %s: %v", key, err)
	}
	if existing == nil {
		_, err = dao.SysConfig.Ctx(ctx).Data(do.SysConfig{
			Name:   name,
			Key:    key,
			Value:  value,
			Remark: remark,
		}).Insert()
		if err != nil {
			t.Fatalf("insert config record %s: %v", key, err)
		}
		if err = hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
			t.Fatalf("mark runtime params changed: %v", err)
		}
		t.Cleanup(func() {
			_, cleanupErr := dao.SysConfig.Ctx(ctx).Unscoped().Where(do.SysConfig{Key: key}).Delete()
			if cleanupErr != nil {
				t.Fatalf("delete config record %s: %v", key, cleanupErr)
			}
			if cleanupErr = hostconfig.New().MarkRuntimeParamsChanged(ctx); cleanupErr != nil {
				t.Fatalf("mark runtime params changed after delete: %v", cleanupErr)
			}
		})
		return
	}

	original := *existing
	_, err = dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Id: existing.Id}).
		Data(do.SysConfig{
			Name:   name,
			Value:  value,
			Remark: remark,
		}).
		Update()
	if err != nil {
		t.Fatalf("update config record %s: %v", key, err)
	}
	if err = hostconfig.New().MarkRuntimeParamsChanged(ctx); err != nil {
		t.Fatalf("mark runtime params changed: %v", err)
	}
	t.Cleanup(func() {
		_, cleanupErr := dao.SysConfig.Ctx(ctx).
			Unscoped().
			Where(do.SysConfig{Id: original.Id}).
			Data(do.SysConfig{
				Name:   original.Name,
				Key:    original.Key,
				Value:  original.Value,
				Remark: original.Remark,
			}).
			Update()
		if cleanupErr != nil {
			t.Fatalf("restore config record %s: %v", key, cleanupErr)
		}
		if cleanupErr = hostconfig.New().MarkRuntimeParamsChanged(ctx); cleanupErr != nil {
			t.Fatalf("mark runtime params changed after restore: %v", cleanupErr)
		}
	})
}

// queryConfigRecord loads one config row by key without soft-delete filtering.
func queryConfigRecord(ctx context.Context, key string) (*entity.SysConfig, error) {
	var item *entity.SysConfig
	err := dao.SysConfig.Ctx(ctx).
		Unscoped().
		Where(do.SysConfig{Key: key}).
		Scan(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}
