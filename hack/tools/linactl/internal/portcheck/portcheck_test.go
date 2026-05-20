// Tests for portcheck.Verify and helper parsers. Each test builds its own
// repository fixture under t.TempDir() so cases are self-contained and order
// independent. The fixtures only include the files portcheck reads, keeping
// setup minimal and the failure messages easy to interpret.
//
// portcheck 的单元测试。每个用例都使用 t.TempDir() 构造一份独立的仓库目录
// 雏形，仅包含 portcheck 实际读取的两个文件，保证测试自包含、顺序无关。

package portcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRepoFixture 在临时根目录下写入后端 config.yaml 和前端 vite.config.mts
// 两个测试夹具。任意一个传空字符串则不创建该文件，便于覆盖 "缺失文件" 用例。
// writeRepoFixture writes the backend config.yaml and the frontend
// vite.config.mts fixtures under a temporary repository root. Passing an empty
// string for either file skips creation, enabling negative tests for missing
// inputs.
func writeRepoFixture(t *testing.T, backendYAML string, viteConfig string) string {
	t.Helper()
	root := t.TempDir()

	if backendYAML != "" {
		backendPath := filepath.Join(root, filepath.FromSlash(backendConfigRelPath))
		if err := os.MkdirAll(filepath.Dir(backendPath), 0o755); err != nil {
			t.Fatalf("create backend config dir: %v", err)
		}
		if err := os.WriteFile(backendPath, []byte(backendYAML), 0o644); err != nil {
			t.Fatalf("write backend config: %v", err)
		}
	}

	if viteConfig != "" {
		vitePath := filepath.Join(root, filepath.FromSlash(frontendViteConfigRelPath))
		if err := os.MkdirAll(filepath.Dir(vitePath), 0o755); err != nil {
			t.Fatalf("create vite config dir: %v", err)
		}
		if err := os.WriteFile(vitePath, []byte(viteConfig), 0o644); err != nil {
			t.Fatalf("write vite config: %v", err)
		}
	}

	return root
}

// TestVerifyAcceptsAlignedPorts 验证三处端口一致时 Verify 返回 nil。
// TestVerifyAcceptsAlignedPorts checks Verify returns nil when every source
// agrees on the same port.
func TestVerifyAcceptsAlignedPorts(t *testing.T) {
	backendYAML := "server:\n  address: \":9120\"\n"
	viteConfig := `
		proxy: {
			'/api': { target: 'http://localhost:9120' },
			'/plugin-assets': { target: 'http://localhost:9120' },
		}
	`
	root := writeRepoFixture(t, backendYAML, viteConfig)

	if err := Verify(root, 9120); err != nil {
		t.Fatalf("expected aligned ports to pass verification, got error: %v", err)
	}
}

// TestVerifyReportsBackendMismatch 验证当 backend config 端口与命令传入值
// 不一致时，Verify 返回的错误中包含后端来源与实际端口。
// TestVerifyReportsBackendMismatch checks Verify returns an error mentioning
// the backend source and observed port when only the backend config drifts.
func TestVerifyReportsBackendMismatch(t *testing.T) {
	backendYAML := "server:\n  address: \":8080\"\n"
	viteConfig := `proxy: { '/api': { target: 'http://localhost:9120' } }`
	root := writeRepoFixture(t, backendYAML, viteConfig)

	err := Verify(root, 9120)
	if err == nil {
		t.Fatal("expected mismatched backend port to fail verification")
	}
	message := err.Error()
	if !strings.Contains(message, "server.address") {
		t.Fatalf("expected error to mention backend server.address source, got: %s", message)
	}
	if !strings.Contains(message, ":8080") {
		t.Fatalf("expected error to mention observed backend port :8080, got: %s", message)
	}
	if !strings.Contains(message, ":9120") {
		t.Fatalf("expected error to mention expected port :9120, got: %s", message)
	}
}

// TestVerifyReportsFrontendMismatch 验证仅前端 vite proxy 端口不一致时，
// Verify 返回的错误中包含 vite 来源与不一致端口。
// TestVerifyReportsFrontendMismatch checks Verify reports the vite proxy
// source and the drifting port when only the frontend config disagrees.
func TestVerifyReportsFrontendMismatch(t *testing.T) {
	backendYAML := "server:\n  address: \":9120\"\n"
	viteConfig := `proxy: { '/api': { target: 'http://localhost:8080' } }`
	root := writeRepoFixture(t, backendYAML, viteConfig)

	err := Verify(root, 9120)
	if err == nil {
		t.Fatal("expected mismatched frontend proxy port to fail verification")
	}
	message := err.Error()
	if !strings.Contains(message, "vite.config.mts proxy target") {
		t.Fatalf("expected error to mention vite proxy source, got: %s", message)
	}
	if !strings.Contains(message, ":8080") {
		t.Fatalf("expected error to mention observed frontend port :8080, got: %s", message)
	}
}

// TestVerifyReportsAllMismatchedFrontendTargets 验证多个 proxy target 中仅
// 部分不一致时，Verify 错误信息逐条列出每个不一致的 target。
// TestVerifyReportsAllMismatchedFrontendTargets checks each diverging proxy
// target appears in the error when only a subset is misaligned.
func TestVerifyReportsAllMismatchedFrontendTargets(t *testing.T) {
	backendYAML := "server:\n  address: \":9120\"\n"
	viteConfig := `
		proxy: {
			'/api': { target: 'http://localhost:9120' },
			'/plugin-assets': { target: 'http://localhost:8080' },
			'/stoplight/apidocs.html': { target: 'http://localhost:7000' },
		}
	`
	root := writeRepoFixture(t, backendYAML, viteConfig)

	err := Verify(root, 9120)
	if err == nil {
		t.Fatal("expected partially aligned frontend proxy targets to fail verification")
	}
	message := err.Error()
	if !strings.Contains(message, ":8080") || !strings.Contains(message, ":7000") {
		t.Fatalf("expected error to list both diverging frontend ports, got: %s", message)
	}
	// 与 backend 一致的那一项不应出现在不一致清单中（即不应作为 "reports :9120" 出现）。
	// The aligned proxy entry must not be flagged as a mismatch (e.g., it
	// should not appear as "reports :9120" in the error).
	if strings.Contains(message, "reports :9120") {
		t.Fatalf("expected aligned proxy target to be excluded from mismatch list, got: %s", message)
	}
}

// TestVerifyFailsWhenBackendConfigMissing 验证后端配置文件缺失时 Verify 返回
// 明确的读文件错误，不静默通过。
// TestVerifyFailsWhenBackendConfigMissing ensures Verify surfaces a clear file
// read error rather than silently passing when the backend manifest is absent.
func TestVerifyFailsWhenBackendConfigMissing(t *testing.T) {
	root := writeRepoFixture(t, "", `proxy: { '/api': { target: 'http://localhost:9120' } }`)

	err := Verify(root, 9120)
	if err == nil {
		t.Fatal("expected missing backend config to produce an error")
	}
	if !strings.Contains(err.Error(), "read backend config") {
		t.Fatalf("expected error to mention reading backend config, got: %v", err)
	}
}

// TestVerifyAcceptsMissingFrontendProxyBlock 验证前端 vite 配置中没有 proxy
// target 块时不视为错误（用户可能裁剪了 dev proxy），只校验后端来源。
// TestVerifyAcceptsMissingFrontendProxyBlock confirms an empty proxy block in
// vite.config.mts is not treated as a verification failure, since an operator
// may legitimately strip the dev proxy section while still using linactl dev.
func TestVerifyAcceptsMissingFrontendProxyBlock(t *testing.T) {
	backendYAML := "server:\n  address: \":9120\"\n"
	viteConfig := `// vite config without a proxy block at all.`
	root := writeRepoFixture(t, backendYAML, viteConfig)

	if err := Verify(root, 9120); err != nil {
		t.Fatalf("expected verification to pass without a frontend proxy block, got: %v", err)
	}
}

// TestParsePortFromAddressAcceptsCommonShapes 验证常见 GoFrame 监听地址格式
// 都能解析出端口。
// TestParsePortFromAddressAcceptsCommonShapes checks the common GoFrame
// listen-address shapes are accepted by the parser.
func TestParsePortFromAddressAcceptsCommonShapes(t *testing.T) {
	cases := map[string]int{
		":9120":             9120,
		"127.0.0.1:9120":    9120,
		"0.0.0.0:18080":     18080,
		"[::1]:9120":        9120,
		"  127.0.0.1:9120 ": 9120,
	}
	for address, expected := range cases {
		got, err := parsePortFromAddress(address)
		if err != nil {
			t.Fatalf("parsePortFromAddress(%q) returned error: %v", address, err)
		}
		if got != expected {
			t.Fatalf("parsePortFromAddress(%q) = %d, want %d", address, got, expected)
		}
	}
}

// TestParsePortFromAddressRejectsInvalidShapes 验证不合法地址会返回错误。
// TestParsePortFromAddressRejectsInvalidShapes ensures malformed addresses
// produce an error rather than a zero port.
func TestParsePortFromAddressRejectsInvalidShapes(t *testing.T) {
	cases := []string{
		"",
		"localhost",
		":0",
		":99999",
		":notaport",
	}
	for _, address := range cases {
		if _, err := parsePortFromAddress(address); err == nil {
			t.Fatalf("expected parsePortFromAddress(%q) to return an error", address)
		}
	}
}
