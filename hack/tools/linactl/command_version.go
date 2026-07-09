// This file implements the version command for release metadata updates.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	markdownImagePattern = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)([^)]*)\)`)
	htmlImageSrcPattern  = regexp.MustCompile(`(?i)(<img\b[^>]*\bsrc\s*=\s*)(["'])([^"']+)(["'])`)
	yamlVersionKeyLine   = regexp.MustCompile(`^(\s*version\s*:\s*)`)
)

// runVersion updates repository release metadata and README image cache keys.
func runVersion(_ context.Context, a *app, input commandInput) error {
	version := strings.TrimSpace(input.Get("to"))
	if version == "" {
		return fmt.Errorf("version target is required; pass to=vMAJOR.MINOR.PATCH")
	}
	if err := validateFrameworkReleaseVersion(version); err != nil {
		return err
	}

	metadataPath := filepath.Join(a.root, "apps", "lina-core", "manifest", "config", "metadata.yaml")
	if err := updateFrameworkVersion(metadataPath, version); err != nil {
		return err
	}

	readmePaths, err := rootReadmeFiles(a.root)
	if err != nil {
		return err
	}
	cacheVersion := strings.TrimPrefix(version, "v")
	totalImages := 0
	for _, readmePath := range readmePaths {
		count, updateErr := updateReadmeImageVersions(readmePath, cacheVersion)
		if updateErr != nil {
			return updateErr
		}
		totalImages += count
	}

	fmt.Fprintf(a.stdout, "Updated framework.version to %s and refreshed %d README image cache keys\n", version, totalImages)
	return nil
}

// updateFrameworkVersion updates framework.version while preserving comments and
// surrounding YAML formatting.
func updateFrameworkVersion(path string, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read metadata %s: %w", path, err)
	}

	line, err := frameworkVersionLine(content, path)
	if err != nil {
		return err
	}
	updated, err := replaceYAMLScalarLine(string(content), line, version)
	if err != nil {
		return fmt.Errorf("update framework.version in %s: %w", path, err)
	}
	if updated == string(content) {
		return nil
	}
	if err = os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write metadata %s: %w", path, err)
	}
	return nil
}

// frameworkVersionLine uses the YAML parser to locate framework.version.
func frameworkVersionLine(content []byte, path string) (int, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return 0, fmt.Errorf("parse metadata %s: %w", path, err)
	}
	root := yamlDocumentContent(&doc)
	if root == nil || root.Kind != yaml.MappingNode {
		return 0, fmt.Errorf("metadata %s must be a YAML mapping", path)
	}
	framework := yamlMappingValue(root, "framework")
	if framework == nil || framework.Kind != yaml.MappingNode {
		return 0, fmt.Errorf("metadata %s is missing framework mapping", path)
	}
	version := yamlMappingValue(framework, "version")
	if version == nil {
		return 0, fmt.Errorf("metadata framework.version is empty in %s", path)
	}
	if strings.TrimSpace(version.Value) == "" {
		return 0, fmt.Errorf("metadata framework.version is empty in %s", path)
	}
	if version.Line <= 0 {
		return 0, fmt.Errorf("metadata framework.version line is unavailable in %s", path)
	}
	return version.Line, nil
}

func yamlDocumentContent(node *yaml.Node) *yaml.Node {
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}
	return node
}

func yamlMappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func replaceYAMLScalarLine(content string, oneBasedLine int, version string) (string, error) {
	lines := strings.SplitAfter(content, "\n")
	if oneBasedLine < 1 || oneBasedLine > len(lines) {
		return "", fmt.Errorf("line %d is outside file", oneBasedLine)
	}

	line := lines[oneBasedLine-1]
	body, ending := splitLineEnding(line)
	matches := yamlVersionKeyLine.FindStringSubmatch(body)
	if len(matches) != 2 {
		return "", fmt.Errorf("line %d does not contain a version key", oneBasedLine)
	}

	suffix := ""
	if commentAt := strings.Index(body, "#"); commentAt >= 0 {
		beforeComment := body[:commentAt]
		spacing := beforeComment[len(strings.TrimRight(beforeComment, " \t")):]
		suffix = spacing + body[commentAt:]
	}
	lines[oneBasedLine-1] = matches[1] + strconv.Quote(version) + suffix + ending
	return strings.Join(lines, ""), nil
}

func splitLineEnding(line string) (string, string) {
	if strings.HasSuffix(line, "\r\n") {
		return strings.TrimSuffix(line, "\r\n"), "\r\n"
	}
	if strings.HasSuffix(line, "\n") {
		return strings.TrimSuffix(line, "\n"), "\n"
	}
	return line, ""
}

func rootReadmeFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read repository root %s: %w", root, err)
	}
	paths := make([]string, 0, 2)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "readme") && strings.HasSuffix(lower, ".md") {
			paths = append(paths, filepath.Join(root, name))
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no root README markdown files found")
	}
	return paths, nil
}

func updateReadmeImageVersions(path string, version string) (int, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read README %s: %w", path, err)
	}
	updated, count := updateReadmeImageContent(string(content), version)
	if updated == string(content) {
		return count, nil
	}
	if err = os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return 0, fmt.Errorf("write README %s: %w", path, err)
	}
	return count, nil
}

func updateReadmeImageContent(content string, version string) (string, int) {
	count := 0
	updated := markdownImagePattern.ReplaceAllStringFunc(content, func(match string) string {
		parts := markdownImagePattern.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		count++
		return "![" + parts[1] + "](" + withVersionQuery(parts[2], version) + parts[3] + ")"
	})
	updated = htmlImageSrcPattern.ReplaceAllStringFunc(updated, func(match string) string {
		parts := htmlImageSrcPattern.FindStringSubmatch(match)
		if len(parts) != 5 || parts[2] != parts[4] {
			return match
		}
		count++
		return parts[1] + parts[2] + withVersionQuery(parts[3], version) + parts[4]
	})
	return updated, count
}

func withVersionQuery(raw string, version string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := parsed.Query()
	query.Set("v", version)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
