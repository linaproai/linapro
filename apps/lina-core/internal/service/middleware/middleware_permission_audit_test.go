// This file audits permission metadata coverage for static host API request DTOs.

package middleware

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// staticPermissionExemptionAllowlist documents the static routes that are
// intentionally exempt from request-level permission metadata.
var staticPermissionExemptionAllowlist = map[string]string{
	"GET /health":                            "public host health probe for container probes and load balancers",
	"POST /auth/login":                       "public login entrypoint",
	"GET /i18n/runtime/locales":              "public runtime i18n locale bootstrap",
	"GET /i18n/runtime/messages":             "public runtime i18n message bootstrap",
	"GET /config/public/frontend":            "public frontend bootstrap whitelist",
	"GET /plugins/dynamic":                   "public shell plugin state bootstrap",
	"GET /dict/data/type/{dictType}":         "login-bound dictionary option lookup reused by authorized business modules",
	"POST /auth/logout":                      "login-bound session logout",
	"GET /menus/all":                         "login-bound menu bootstrap",
	"GET /user/info":                         "login-bound permission bootstrap",
	"GET /user/profile":                      "login-bound current-user self profile query",
	"PUT /user/profile":                      "login-bound current-user self profile update",
	"PUT /user/profile/avatar":               "login-bound current-user self avatar update",
	"GET /user/message":                      "login-bound current-user inbox list",
	"GET /user/message/count":                "login-bound current-user inbox badge count",
	"GET /user/message/{id}":                 "login-bound current-user inbox detail query",
	"GET /plugins/{id}/resources/{resource}": "login-bound plugin resource query with controller-level plugin permission enforcement",
	"PUT /user/message/{id}/read":            "login-bound current-user inbox single read",
	"PUT /user/message/read-all":             "login-bound current-user inbox bulk read",
	"DELETE /user/message/{id}":              "login-bound current-user inbox single delete",
	"DELETE /user/message/clear":             "login-bound current-user inbox clear",
}

// staticPermissionAuditRoute captures the metadata needed to audit one static
// host API request DTO.
type staticPermissionAuditRoute struct {
	File       string
	Line       int
	TypeName   string
	RouteKey   string
	Permission string
	Summary    string
}

// TestStaticHostAPIRequestsDeclarePermissionOrAllowlistedExemption guards the
// declarative permission policy for static host API request DTOs.
func TestStaticHostAPIRequestsDeclarePermissionOrAllowlistedExemption(t *testing.T) {
	t.Parallel()

	routes := loadStaticPermissionAuditRoutes(t)
	if len(routes) == 0 {
		t.Fatal("expected at least one static host API request route to audit")
	}

	var (
		failures       []string
		seenExemptions = make(map[string]struct{}, len(staticPermissionExemptionAllowlist))
		seenRoutes     = make(map[string]staticPermissionAuditRoute, len(routes))
	)

	for _, route := range routes {
		if previous, ok := seenRoutes[route.RouteKey]; ok {
			failures = append(
				failures,
				fmt.Sprintf(
					"duplicate static API route %s declared by %s:%d (%s) and %s:%d (%s)",
					route.RouteKey,
					previous.File,
					previous.Line,
					previous.TypeName,
					route.File,
					route.Line,
					route.TypeName,
				),
			)
			continue
		}
		seenRoutes[route.RouteKey] = route

		exemptionReason, exempted := staticPermissionExemptionAllowlist[route.RouteKey]
		if exempted {
			seenExemptions[route.RouteKey] = struct{}{}
			if route.Permission != "" {
				failures = append(
					failures,
					fmt.Sprintf(
						"%s is allowlisted as %q but already declares permission %q at %s:%d; remove the stale exemption",
						route.RouteKey,
						exemptionReason,
						route.Permission,
						route.File,
						route.Line,
					),
				)
			}
			continue
		}

		if route.Permission == "" {
			failures = append(
				failures,
				fmt.Sprintf(
					"%s (%s:%d %s) is missing permission metadata and is not in the exemption allowlist",
					route.RouteKey,
					route.File,
					route.Line,
					route.TypeName,
				),
			)
		}
	}

	for routeKey, reason := range staticPermissionExemptionAllowlist {
		if _, ok := seenExemptions[routeKey]; ok {
			continue
		}
		failures = append(
			failures,
			fmt.Sprintf("allowlisted exemption %s (%s) no longer matches any request DTO", routeKey, reason),
		)
	}

	if len(failures) == 0 {
		return
	}
	sort.Strings(failures)
	t.Fatalf("static API permission audit failed:\n- %s", strings.Join(failures, "\n- "))
}

// loadStaticPermissionAuditRoutes parses all static API request DTO files and
// collects the route metadata that participates in the permission audit.
func loadStaticPermissionAuditRoutes(t *testing.T) []staticPermissionAuditRoute {
	t.Helper()

	moduleRoot := resolveMiddlewareModuleRoot(t)
	pattern := filepath.Join(moduleRoot, "api", "*", "v1", "*.go")

	paths, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob static API request files: %v", err)
	}
	sort.Strings(paths)

	fset := token.NewFileSet()
	routes := make([]staticPermissionAuditRoute, 0, len(paths))
	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		routes = append(routes, collectStaticPermissionAuditRoutes(t, fset, file, path)...)
	}
	return routes
}

// collectStaticPermissionAuditRoutes extracts audited request DTO metadata from
// one parsed API source file.
func collectStaticPermissionAuditRoutes(
	t *testing.T,
	fset *token.FileSet,
	file *ast.File,
	path string,
) []staticPermissionAuditRoute {
	t.Helper()

	routes := make([]staticPermissionAuditRoute, 0)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Req") {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			metaField := findEmbeddedMetaField(structType)
			if metaField == nil || metaField.Tag == nil {
				continue
			}

			rawTag := strings.Trim(metaField.Tag.Value, "`")
			tag := reflect.StructTag(rawTag)
			pathValue := tag.Get("path")
			methodValue := strings.ToUpper(strings.TrimSpace(tag.Get("method")))
			if pathValue == "" || methodValue == "" {
				t.Fatalf(
					"%s:%d %s must declare both path and method in g.Meta",
					path,
					fset.Position(metaField.Pos()).Line,
					typeSpec.Name.Name,
				)
			}

			permissionValue := strings.TrimSpace(tag.Get(staticPermissionMetaTag))
			if permissionValue == "" {
				permissionValue = strings.TrimSpace(tag.Get(staticPermissionMetaTagAlias))
			}
			routes = append(routes, staticPermissionAuditRoute{
				File:       path,
				Line:       fset.Position(metaField.Pos()).Line,
				TypeName:   typeSpec.Name.Name,
				RouteKey:   methodValue + " " + pathValue,
				Permission: permissionValue,
				Summary:    tag.Get("summary"),
			})
		}
	}
	return routes
}

// findEmbeddedMetaField returns the anonymous g.Meta field embedded in the DTO
// struct, if present.
func findEmbeddedMetaField(structType *ast.StructType) *ast.Field {
	if structType == nil || structType.Fields == nil {
		return nil
	}
	for _, field := range structType.Fields.List {
		if len(field.Names) != 0 {
			continue
		}
		selectorExpr, ok := field.Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if selectorExpr.Sel == nil || selectorExpr.Sel.Name != "Meta" {
			continue
		}
		return field
	}
	return nil
}

// resolveMiddlewareModuleRoot locates the lina-core module root from the test
// file location so the audit can scan API DTO files without hard-coded paths.
func resolveMiddlewareModuleRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", ".."))
}
