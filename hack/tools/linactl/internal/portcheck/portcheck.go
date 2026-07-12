// Package portcheck verifies that the backend listening port, the frontend
// vite proxy target, and the port supplied to linactl all agree before
// development services are started. It exists to make implicit cross-file port
// assumptions explicit and to fail fast with an actionable error message
// instead of letting `make dev` start mismatched services.
//
// portcheck 子组件用于在 linactl 启动开发服务之前，校验三处端口配置一致：
//  1. 命令侧传入的 backend_port（默认来自 Makefile 的 LINA_CORE_PORT）
//  2. 后端配置 apps/lina-core/manifest/config/config.yaml 中的 server.address
//  3. 前端 apps/lina-vben/apps/web-antd/vite.config.mts 中所有 proxy target
//
// 任一处不一致都会返回带有具体来源、期望值与实际值的错误，避免 dev 服务以错配
// 端口启动后再以"探活超时"或"接口 404"等间接形式暴露问题。
package portcheck

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// 后端配置文件相对仓库根的路径。
// Backend manifest config path relative to the repository root.
const backendConfigRelPath = "apps/lina-core/manifest/config/config.yaml"

// 前端 vite 配置文件相对仓库根的路径。
// Frontend vite config path relative to the repository root.
const frontendViteConfigRelPath = "apps/lina-vben/apps/web-antd/vite.config.mts"

// 命令传入端口在错误消息中显示的来源标识。
// Source label used in error messages for the linactl-supplied port value.
const commandSource = "linactl backend_port (Makefile LINA_CORE_PORT)"

// PortFinding stores one discovered port with its human-readable source.
// PortFinding 用来记录一处端口配置及其可读来源信息。
type PortFinding struct {
	// Source 是用户可识别的来源描述，例如 "config.yaml server.address"。
	Source string
	// Port 是从该来源解析得到的端口值。
	Port int
	// Detail 是该来源附加的可选定位信息，例如行号或匹配片段。
	Detail string
}

// Verify ensures the backend config port and all frontend vite proxy targets
// match the supplied expected port. Any mismatch is returned as a single error
// that lists every disagreeing finding with its source so the operator can fix
// the misaligned file directly.
//
// Verify 校验后端配置端口与前端 vite proxy target 端口是否与命令传入的
// expectedPort 一致。任一不一致都会以单条错误返回，错误中列出全部不匹配项，
// 便于排查。
func Verify(root string, expectedPort int) error {
	backendFinding, err := readBackendConfigPort(root)
	if err != nil {
		return err
	}
	frontendFindings, err := readFrontendProxyPorts(root)
	if err != nil {
		return err
	}

	all := append([]PortFinding{backendFinding}, frontendFindings...)
	var mismatches []PortFinding
	for _, finding := range all {
		if finding.Port != expectedPort {
			mismatches = append(mismatches, finding)
		}
	}
	if len(mismatches) == 0 {
		return nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "backend port mismatch: %s expects :%d, but:", commandSource, expectedPort)
	for _, finding := range mismatches {
		fmt.Fprintf(&b, "\n  - %s reports :%d", finding.Source, finding.Port)
		if finding.Detail != "" {
			fmt.Fprintf(&b, " (%s)", finding.Detail)
		}
	}
	b.WriteString("\nalign these sources to a single port, or pass backend_port=<port> to linactl dev")
	return errors.New(b.String())
}

// backendConfigShape mirrors the minimal subset of the backend manifest needed
// to extract the HTTP listen address.
// backendConfigShape 仅映射后端 manifest 中需要的最小字段，用于提取监听地址。
type backendConfigShape struct {
	Server struct {
		Address string `yaml:"address"`
	} `yaml:"server"`
}

// readBackendConfigPort parses the backend manifest config and extracts the
// port portion of server.address.
//
// readBackendConfigPort 解析后端 manifest 配置，从 server.address 中提取端口。
func readBackendConfigPort(root string) (PortFinding, error) {
	configPath := filepath.Join(root, filepath.FromSlash(backendConfigRelPath))
	source := fmt.Sprintf("%s server.address", backendConfigRelPath)

	content, err := os.ReadFile(configPath)
	if err != nil {
		return PortFinding{}, fmt.Errorf("read backend config %s: %w", configPath, err)
	}

	var shape backendConfigShape
	if err = yaml.Unmarshal(content, &shape); err != nil {
		return PortFinding{}, fmt.Errorf("parse backend config %s: %w", configPath, err)
	}
	address := strings.TrimSpace(shape.Server.Address)
	if address == "" {
		return PortFinding{}, fmt.Errorf("backend config %s is missing server.address", configPath)
	}
	port, err := parsePortFromAddress(address)
	if err != nil {
		return PortFinding{}, fmt.Errorf("parse server.address %q in %s: %w", address, configPath, err)
	}
	return PortFinding{
		Source: source,
		Port:   port,
		Detail: fmt.Sprintf("address=%q", address),
	}, nil
}

// addressPortPattern matches the trailing :<port> in a GoFrame server address
// such as ":9120", "127.0.0.1:9120", or "[::1]:9120".
// addressPortPattern 用于从形如 ":9120"、"127.0.0.1:9120"、"[::1]:9120" 的
// GoFrame 监听地址中提取末尾端口。
var addressPortPattern = regexp.MustCompile(`:(\d+)\s*$`)

// parsePortFromAddress extracts the numeric port from a GoFrame style listen
// address. It accepts a leading host or empty host (":9120" form).
//
// parsePortFromAddress 从 GoFrame 风格监听地址中解析出端口，支持带 host 与
// 仅冒号端口（":9120"）两种形式。
func parsePortFromAddress(address string) (int, error) {
	match := addressPortPattern.FindStringSubmatch(address)
	if len(match) != 2 {
		return 0, fmt.Errorf("address %q does not contain a trailing :port", address)
	}
	port, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", match[1], err)
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("port %d out of range 1-65535", port)
	}
	return port, nil
}

// viteProxyTargetPattern matches `target: 'http(s)?://host:<port>'` literals
// inside vite.config.mts proxy entries. The host portion is intentionally
// permissive (alphanumerics, dots, hyphens, brackets for IPv6) so future host
// changes (such as switching from "localhost" to "127.0.0.1") do not require
// updating this regex.
//
// viteProxyTargetPattern 匹配 vite.config.mts 中 proxy 项的 target 字段，
// 形如 `target: 'http://localhost:9120'`。host 段刻意写得宽松，未来切换
// localhost/127.0.0.1 不需要修改正则。
var viteProxyTargetPattern = regexp.MustCompile(
	`target:\s*['"]https?://[A-Za-z0-9\.\-\[\]]+:(\d+)['"]`,
)

// readFrontendProxyPorts scans vite.config.mts for every proxy target and
// returns one PortFinding per occurrence. Returning each occurrence (instead
// of a deduplicated set) lets Verify report a precise location when only a
// subset of targets drifts out of sync.
//
// readFrontendProxyPorts 扫描 vite.config.mts 中所有 proxy target 字面量，
// 每一次出现都返回一条 PortFinding。逐条返回（而非去重）便于在仅部分 target
// 不一致时给出精确的定位信息。
func readFrontendProxyPorts(root string) ([]PortFinding, error) {
	configPath := filepath.Join(root, filepath.FromSlash(frontendViteConfigRelPath))
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read frontend vite config %s: %w", configPath, err)
	}

	matches := viteProxyTargetPattern.FindAllSubmatchIndex(content, -1)
	if len(matches) == 0 {
		// 找不到 proxy target 时不视为错误，可能是用户裁剪了 dev proxy 块。
		// 此时无法做前端代理校验，返回空集合让 Verify 继续校验其他来源。
		// When no proxy target is found, treat the frontend as not in scope of
		// this check rather than returning an error, since a custom config may
		// legitimately omit the dev proxy block.
		return nil, nil
	}

	findings := make([]PortFinding, 0, len(matches))
	for _, match := range matches {
		// match[0:1] 是整个匹配；match[2:3] 是端口子捕获组的字节区间。
		// match[0:1] is the full match span; match[2:3] is the port capture group.
		fullStart, fullEnd := match[0], match[1]
		portStart, portEnd := match[2], match[3]
		portText := string(content[portStart:portEnd])
		port, err := strconv.Atoi(portText)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy target port %q in %s: %w", portText, configPath, err)
		}
		if port <= 0 || port > 65535 {
			return nil, fmt.Errorf("proxy target port %d out of range 1-65535 in %s", port, configPath)
		}
		findings = append(findings, PortFinding{
			Source: fmt.Sprintf("%s proxy target", frontendViteConfigRelPath),
			Port:   port,
			Detail: fmt.Sprintf("line %d: %s", lineNumber(content, fullStart), strings.TrimSpace(string(content[fullStart:fullEnd]))),
		})
	}
	return findings, nil
}

// lineNumber returns the 1-based line number that contains the given byte
// offset, used to enrich port mismatch error messages with file locations.
//
// lineNumber 根据字节偏移返回 1 起始的行号，用于在错误消息中增强定位信息。
func lineNumber(content []byte, offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}
	// 行号 = 偏移之前出现的换行符数 + 1。
	// Line number is one plus the count of newline characters before the offset.
	count := 1
	for _, b := range content[:offset] {
		if b == '\n' {
			count++
		}
	}
	return count
}
