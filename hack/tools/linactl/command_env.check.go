// This file implements the env.check command for local development prerequisites.
// Probes are intentionally non-fatal: each tool reports its own status and
// remark so one missing dependency does not hide the rest of the environment.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"linactl/internal/toolutil"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// envVersionPattern extracts the first semantic version from common tool output.
var envVersionPattern = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

const (
	// envCoreConfigPath is the Lina core runtime config used for database probes.
	envCoreConfigPath = "apps/lina-core/manifest/config/config.yaml"
	// envDatabaseTypePostgreSQL is the GoFrame link prefix for PostgreSQL.
	envDatabaseTypePostgreSQL = "pgsql"
	// envPostgreSQLDefaultPort is used when the configured tcp address omits a port.
	envPostgreSQLDefaultPort = "5432"
	// envPostgreSQLProbeTimeout bounds server-version checks against unreachable databases.
	envPostgreSQLProbeTimeout = 5 * time.Second
)

// envProbeKind selects special probe behavior for tools that need more than a
// simple version command. The zero value runs the declared command and parses
// its version output.
type envProbeKind string

const (
	// envProbeKindPostgreSQLServer connects to the configured database and reads
	// the server version instead of reporting a local client version.
	envProbeKindPostgreSQLServer envProbeKind = "postgresql-server"
)

// envTool describes one local development prerequisite checked by env.check.
type envTool struct {
	Name          string
	ProbeKind     envProbeKind
	Command       string
	Args          []string
	Dir           string
	RequiredPath  string
	Required      string
	MinVersion    string
	MissingRemark string
	FailureRemark string
	SuccessRemark string
}

// envProbeResult stores one raw tool probe result before version evaluation.
type envProbeResult struct {
	Output  string
	Missing bool
	Err     error
	Remark  string
}

// envCheckRow stores one rendered table row for a prerequisite.
type envCheckRow struct {
	Name     string
	Current  string
	Required string
	OK       bool
	Remark   string
}

// envProbeFunc probes one local tool and returns its version command output.
type envProbeFunc func(context.Context, *app, envTool) envProbeResult

// semanticVersion stores a normalized comparable tool version.
type semanticVersion struct {
	Major int
	Minor int
	Patch int
}

// envCoreConfig stores only the runtime config fields needed by env.check.
type envCoreConfig struct {
	Database envDatabaseSection `yaml:"database"`
}

// envDatabaseSection stores the default GoFrame database connection settings.
type envDatabaseSection struct {
	Default envDatabaseConfig `yaml:"default"`
}

// envDatabaseConfig stores one database connection link from config.yaml.
type envDatabaseConfig struct {
	Link string `yaml:"link"`
}

// envPostgreSQLConnection stores database/sql fields parsed from a GoFrame link.
type envPostgreSQLConnection struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// runEnvCheck prints a local development prerequisite table.
func runEnvCheck(ctx context.Context, a *app, _ commandInput) error {
	rows := collectEnvCheckRows(ctx, a, defaultEnvTools(a.root), probeEnvTool)
	return printEnvCheckTable(a.stdout, rows)
}

// defaultEnvTools returns the local tools required by the default development workflow.
func defaultEnvTools(root string) []envTool {
	return []envTool{
		{
			Name:          "Go",
			Command:       "go",
			Args:          []string{"version"},
			Required:      ">= 1.25.0",
			MinVersion:    "1.25.0",
			MissingRemark: "Install Go 1.25 or newer.",
			SuccessRemark: "Go toolchain is available.",
		},
		{
			Name:          "Node.js",
			Command:       "node",
			Args:          []string{"--version"},
			Required:      ">= 20.19.0",
			MinVersion:    "20.19.0",
			MissingRemark: "Install Node.js 20.19 or newer.",
			SuccessRemark: "Node.js runtime is available.",
		},
		{
			Name:          "pnpm",
			Command:       "pnpm",
			Args:          []string{"--version"},
			Required:      ">= 10.0.0",
			MinVersion:    "10.0.0",
			MissingRemark: "Install pnpm 10 or newer.",
			SuccessRemark: "pnpm package manager is available.",
		},
		{
			Name:          "Vite",
			Command:       "node",
			Args:          []string{envViteCLIPath(root), "--version"},
			RequiredPath:  envViteCLIPath(root),
			Required:      ">= 7.3.1",
			MinVersion:    "7.3.1",
			MissingRemark: "Frontend dependencies are missing; run make env.setup.",
			FailureRemark: "Vite could not run; run make env.setup.",
			SuccessRemark: "Project-local Vite binary is available.",
		},
		{
			Name:          "Playwright",
			Command:       "node",
			Args:          []string{envPlaywrightCLIPath(root), "--version"},
			Dir:           filepath.Join(root, "hack", "tests"),
			RequiredPath:  envPlaywrightCLIPath(root),
			Required:      ">= 1.58.2",
			MinVersion:    "1.58.2",
			MissingRemark: "Playwright dependencies are missing; run make env.setup.",
			FailureRemark: "Playwright CLI is unavailable; run make env.setup.",
			SuccessRemark: "Playwright CLI is available.",
		},
		{
			Name:          "PostgreSQL",
			ProbeKind:     envProbeKindPostgreSQLServer,
			Required:      ">= 14.0.0",
			MinVersion:    "14.0.0",
			FailureRemark: "Could not query PostgreSQL server version.",
			SuccessRemark: "PostgreSQL server version detected via Go database connection.",
		},
	}
}

// collectEnvCheckRows converts raw tool probes into table rows.
func collectEnvCheckRows(ctx context.Context, a *app, tools []envTool, probe envProbeFunc) []envCheckRow {
	rows := make([]envCheckRow, 0, len(tools))
	for _, tool := range tools {
		result := probe(ctx, a, tool)
		rows = append(rows, evaluateEnvTool(tool, result))
	}
	return rows
}

// probeEnvTool runs one tool version command and captures its output.
func probeEnvTool(ctx context.Context, a *app, tool envTool) envProbeResult {
	if tool.ProbeKind == envProbeKindPostgreSQLServer {
		return probePostgreSQLServerVersion(ctx, a, tool)
	}

	if tool.RequiredPath != "" && !envFileAvailable(tool.RequiredPath) {
		return envProbeResult{Missing: true}
	}

	if !envCommandAvailable(tool.Command) {
		return envProbeResult{Missing: true, Remark: fmt.Sprintf("Required command %q is not available in PATH.", tool.Command)}
	}

	options := commandOptions{Dir: tool.Dir, Quiet: true}
	output, err := a.runCommandOutput(ctx, options, tool.Command, tool.Args...)
	return envProbeResult{Output: output, Err: err}
}

// evaluateEnvTool compares one probe result with the declared minimum version.
func evaluateEnvTool(tool envTool, result envProbeResult) envCheckRow {
	row := envCheckRow{Name: tool.Name, Required: tool.Required}
	if result.Missing {
		row.Current = "not found"
		row.Remark = toolutil.FirstNonEmpty(result.Remark, tool.MissingRemark)
		return row
	}
	if result.Err != nil {
		row.Current = "unavailable"
		row.Remark = toolutil.FirstNonEmpty(result.Remark, tool.FailureRemark, shortEnvOutput(result.Err.Error()))
		return row
	}

	current, ok := extractSemanticVersion(result.Output)
	if !ok {
		row.Current = "unknown"
		row.Remark = toolutil.FirstNonEmpty(result.Remark, "could not parse version from: "+shortEnvOutput(result.Output))
		return row
	}
	minimum, ok := parseSemanticVersion(tool.MinVersion)
	if !ok {
		row.Current = current.String()
		row.Remark = "invalid required version " + tool.MinVersion
		return row
	}

	row.Current = current.String()
	if compareSemanticVersion(current, minimum) < 0 {
		row.Remark = "upgrade required"
		return row
	}
	row.OK = true
	row.Remark = toolutil.FirstNonEmpty(result.Remark, tool.SuccessRemark)
	return row
}

// probePostgreSQLServerVersion connects with Go's database/sql and returns SHOW server_version.
func probePostgreSQLServerVersion(ctx context.Context, a *app, tool envTool) envProbeResult {
	connection, err := loadEnvPostgreSQLConnection(a.root)
	if err != nil {
		return envProbeResult{
			Err:    err,
			Remark: "could not load PostgreSQL database link from " + envCoreConfigPath + ": " + shortEnvOutput(err.Error()),
		}
	}

	probeCtx, cancel := context.WithTimeout(ctx, envPostgreSQLProbeTimeout)
	defer cancel()

	output, err := queryPostgreSQLServerVersion(probeCtx, connection)
	if err != nil {
		reason := shortEnvOutput(err.Error())
		if probeCtx.Err() != nil {
			reason = "timed out after " + envPostgreSQLProbeTimeout.String()
		}
		return envProbeResult{
			Err:    err,
			Remark: "could not query PostgreSQL server version using " + envCoreConfigPath + " database.default.link: " + reason,
		}
	}
	return envProbeResult{Output: output}
}

// queryPostgreSQLServerVersion opens the configured PostgreSQL database and
// queries the server version without relying on an external client binary.
func queryPostgreSQLServerVersion(ctx context.Context, connection envPostgreSQLConnection) (string, error) {
	return queryPostgreSQLServerVersionWithDriver(ctx, "postgres", connection)
}

// printEnvCheckTable writes the environment check table in a stable format.
func printEnvCheckTable(out io.Writer, rows []envCheckRow) error {
	headers := []string{"Name", "Current Version", "Required Version", "Satisfied", "Remark"}
	tableRows := make([][]string, 0, len(rows)+1)
	tableRows = append(tableRows, headers)
	for _, row := range rows {
		satisfied := "No"
		if row.OK {
			satisfied = "Yes"
		}
		tableRows = append(tableRows, []string{row.Name, row.Current, row.Required, satisfied, row.Remark})
	}

	widths := envTableWidths(tableRows)
	border := envTableBorder(widths)
	if _, err := fmt.Fprintln(out, border); err != nil {
		return fmt.Errorf("write environment check border: %w", err)
	}
	for index, cells := range tableRows {
		if _, err := fmt.Fprintln(out, envTableRow(cells, widths)); err != nil {
			return fmt.Errorf("write environment check row: %w", err)
		}
		if index == 0 {
			if _, err := fmt.Fprintln(out, border); err != nil {
				return fmt.Errorf("write environment check header border: %w", err)
			}
		}
	}
	if _, err := fmt.Fprintln(out, border); err != nil {
		return fmt.Errorf("write environment check closing border: %w", err)
	}
	return nil
}

// envCommandAvailable reports whether a command can be executed by linactl.
func envCommandAvailable(command string) bool {
	if filepath.IsAbs(command) {
		return envFileAvailable(command)
	}
	_, err := exec.LookPath(command)
	return err == nil
}

// envFileAvailable reports whether a local command or CLI script exists.
func envFileAvailable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// envViteCLIPath returns Vite's project-local JavaScript CLI entrypoint.
func envViteCLIPath(root string) string {
	return filepath.Join(root, "apps", "lina-vben", "node_modules", "vite", "bin", "vite.js")
}

// envPlaywrightCLIPath returns the E2E workspace Playwright CLI entrypoint.
func envPlaywrightCLIPath(root string) string {
	return filepath.Join(root, "hack", "tests", "node_modules", "@playwright", "test", "cli.js")
}

// loadEnvPostgreSQLConnection reads and parses the default core database link.
func loadEnvPostgreSQLConnection(root string) (envPostgreSQLConnection, error) {
	configPath := filepath.Join(root, envCoreConfigPath)
	content, err := os.ReadFile(configPath)
	if err != nil {
		return envPostgreSQLConnection{}, fmt.Errorf("read %s: %w", envCoreConfigPath, err)
	}

	var cfg envCoreConfig
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return envPostgreSQLConnection{}, fmt.Errorf("parse %s: %w", envCoreConfigPath, err)
	}
	link := strings.TrimSpace(cfg.Database.Default.Link)
	if link == "" {
		return envPostgreSQLConnection{}, fmt.Errorf("database.default.link is empty in %s", envCoreConfigPath)
	}
	connection, err := parseEnvPostgreSQLLink(link)
	if err != nil {
		return envPostgreSQLConnection{}, fmt.Errorf("parse database.default.link in %s: %w", envCoreConfigPath, err)
	}
	return connection, nil
}

// queryPostgreSQLServerVersionWithDriver allows tests to inject a database/sql
// driver while production uses the PostgreSQL driver registered by lib/pq.
func queryPostgreSQLServerVersionWithDriver(ctx context.Context, driverName string, connection envPostgreSQLConnection) (version string, err error) {
	db, err := sql.Open(driverName, connection.dsn())
	if err != nil {
		return "", fmt.Errorf("open PostgreSQL connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("close PostgreSQL connection: %w", closeErr)
			}
		}
	}()

	if err = db.QueryRowContext(ctx, "SHOW server_version").Scan(&version); err != nil {
		return "", fmt.Errorf("query PostgreSQL server version: %w", err)
	}
	return version, nil
}

// parseEnvPostgreSQLLink converts a GoFrame pgsql link into database/sql fields.
func parseEnvPostgreSQLLink(link string) (envPostgreSQLConnection, error) {
	driver, remainder, ok := strings.Cut(strings.TrimSpace(link), ":")
	if !ok {
		return envPostgreSQLConnection{}, fmt.Errorf("missing database type prefix")
	}
	if driver != envDatabaseTypePostgreSQL {
		return envPostgreSQLConnection{}, fmt.Errorf("configured database type %q is not PostgreSQL", driver)
	}

	credentials, remainder, ok := strings.Cut(remainder, "@tcp(")
	if !ok {
		return envPostgreSQLConnection{}, fmt.Errorf("missing tcp address")
	}
	address, databasePart, ok := strings.Cut(remainder, ")/")
	if !ok {
		return envPostgreSQLConnection{}, fmt.Errorf("missing database name")
	}

	user, password, err := parseEnvPostgreSQLCredentials(credentials)
	if err != nil {
		return envPostgreSQLConnection{}, err
	}
	host, port, err := parseEnvPostgreSQLAddress(address)
	if err != nil {
		return envPostgreSQLConnection{}, err
	}
	database, sslMode, err := parseEnvPostgreSQLDatabase(databasePart)
	if err != nil {
		return envPostgreSQLConnection{}, err
	}

	return envPostgreSQLConnection{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
	}, nil
}

// parseEnvPostgreSQLCredentials parses user and password from a GoFrame link.
func parseEnvPostgreSQLCredentials(credentials string) (string, string, error) {
	userPart, passwordPart, _ := strings.Cut(credentials, ":")
	user, err := envURLUnescape("database user", userPart)
	if err != nil {
		return "", "", err
	}
	password, err := envURLUnescape("database password", passwordPart)
	if err != nil {
		return "", "", err
	}
	return user, password, nil
}

// parseEnvPostgreSQLAddress parses host and port from tcp(host:port).
func parseEnvPostgreSQLAddress(address string) (string, string, error) {
	trimmed := strings.TrimSpace(address)
	if trimmed == "" {
		return "", "", fmt.Errorf("database tcp address is empty")
	}

	host, port, err := net.SplitHostPort(trimmed)
	if err == nil {
		return normalizeEnvPostgreSQLHost(host), toolutil.FirstNonEmpty(port, envPostgreSQLDefaultPort), nil
	}
	if strings.Count(trimmed, ":") == 1 {
		hostPart, portPart, _ := strings.Cut(trimmed, ":")
		if strings.TrimSpace(hostPart) == "" {
			return "", "", fmt.Errorf("database host is empty")
		}
		return normalizeEnvPostgreSQLHost(hostPart), toolutil.FirstNonEmpty(strings.TrimSpace(portPart), envPostgreSQLDefaultPort), nil
	}
	return normalizeEnvPostgreSQLHost(trimmed), envPostgreSQLDefaultPort, nil
}

// parseEnvPostgreSQLDatabase parses the database name and query options.
func parseEnvPostgreSQLDatabase(databasePart string) (string, string, error) {
	databaseRaw, queryRaw, _ := strings.Cut(databasePart, "?")
	database, err := envURLUnescape("database name", strings.TrimPrefix(databaseRaw, "/"))
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(database) == "" {
		return "", "", fmt.Errorf("database name is empty")
	}

	query, err := url.ParseQuery(queryRaw)
	if err != nil {
		return "", "", fmt.Errorf("parse database query options: %w", err)
	}
	return database, strings.TrimSpace(query.Get("sslmode")), nil
}

// envURLUnescape decodes connection fields that may be URL escaped.
func envURLUnescape(label string, value string) (string, error) {
	decoded, err := url.PathUnescape(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("decode %s: %w", label, err)
	}
	return decoded, nil
}

// normalizeEnvPostgreSQLHost trims connection-string brackets from IPv6 hosts.
func normalizeEnvPostgreSQLHost(host string) string {
	return strings.Trim(strings.TrimSpace(host), "[]")
}

// dsn returns a lib/pq connection string for PostgreSQL server-version checks.
func (connection envPostgreSQLConnection) dsn() string {
	values := url.Values{}
	if connection.SSLMode != "" {
		values.Set("sslmode", connection.SSLMode)
	}

	dsn := url.URL{
		Scheme:   "postgres",
		Host:     net.JoinHostPort(connection.Host, connection.Port),
		Path:     connection.Database,
		RawQuery: values.Encode(),
	}
	if connection.User != "" {
		dsn.User = url.UserPassword(connection.User, connection.Password)
	}
	return dsn.String()
}

// envTableWidths computes the minimum width for each rendered table column.
func envTableWidths(rows [][]string) []int {
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for index, cell := range row {
			width := len(envTableCell(cell))
			if width > widths[index] {
				widths[index] = width
			}
		}
	}
	return widths
}

// envTableBorder renders one ASCII border for the boxed environment table.
func envTableBorder(widths []int) string {
	var builder strings.Builder
	builder.WriteByte('+')
	for _, width := range widths {
		builder.WriteString(strings.Repeat("-", width+2))
		builder.WriteByte('+')
	}
	return builder.String()
}

// envTableRow renders one padded table row.
func envTableRow(cells []string, widths []int) string {
	var builder strings.Builder
	builder.WriteByte('|')
	for index, cell := range cells {
		normalized := envTableCell(cell)
		builder.WriteByte(' ')
		builder.WriteString(normalized)
		builder.WriteString(strings.Repeat(" ", widths[index]-len(normalized)+1))
		builder.WriteByte('|')
	}
	return builder.String()
}

// envTableCell normalizes one table cell to a single-line value.
func envTableCell(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// extractSemanticVersion returns the first semantic version found in command output.
func extractSemanticVersion(output string) (semanticVersion, bool) {
	match := envVersionPattern.FindString(output)
	if match == "" {
		return semanticVersion{}, false
	}
	return parseSemanticVersion(match)
}

// parseSemanticVersion parses a two- or three-part semantic version.
func parseSemanticVersion(value string) (semanticVersion, bool) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) < 2 || len(parts) > 3 {
		return semanticVersion{}, false
	}

	numbers := []int{0, 0, 0}
	for index, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil {
			return semanticVersion{}, false
		}
		numbers[index] = parsed
	}
	return semanticVersion{Major: numbers[0], Minor: numbers[1], Patch: numbers[2]}, true
}

// compareSemanticVersion compares two semantic versions.
func compareSemanticVersion(left semanticVersion, right semanticVersion) int {
	if left.Major != right.Major {
		return left.Major - right.Major
	}
	if left.Minor != right.Minor {
		return left.Minor - right.Minor
	}
	return left.Patch - right.Patch
}

// String returns a normalized three-part version string.
func (v semanticVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// shortEnvOutput normalizes command output for a single-line table remark.
func shortEnvOutput(value string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(value, "\n", " "))
	const maxLength = 220
	if len(trimmed) <= maxLength {
		return trimmed
	}
	return trimmed[:maxLength] + "..."
}
