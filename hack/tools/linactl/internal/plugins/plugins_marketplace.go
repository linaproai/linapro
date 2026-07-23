// This file installs plugin releases from the LinaPro marketplace distribution
// API while reusing the managed source-plugin workspace.

package plugins

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"linactl/internal/config"
	"linactl/internal/fileutil"
	"linactl/internal/toolutil"
)

const marketplaceAPIPathPrefix = "/x/linapro-plugin-marketplace/api/v1/market"

// marketplaceDistributionResponse mirrors the marketplace distribution endpoint.
type marketplaceDistributionResponse struct {
	Distribution *marketplaceDistribution `json:"distribution"`
}

// marketplaceDownloadSessionResponse mirrors the marketplace download endpoint.
type marketplaceDownloadSessionResponse struct {
	Session *marketplaceDownloadSession `json:"session"`
}

// marketplaceDistribution stores the CLI-facing release installation projection.
type marketplaceDistribution struct {
	Mode                    string `json:"mode"`
	PluginID                string `json:"pluginId"`
	Version                 string `json:"version"`
	PluginType              string `json:"pluginType"`
	RepoURL                 string `json:"repoUrl"`
	Ref                     string `json:"ref"`
	Path                    string `json:"path"`
	Provider                string `json:"provider"`
	RequiresAuth            bool   `json:"requiresAuth"`
	ArtifactType            string `json:"artifactType"`
	Sha256                  string `json:"sha256"`
	SizeBytes               int64  `json:"sizeBytes"`
	DownloadSessionRequired bool   `json:"downloadSessionRequired"`
}

// marketplaceDownloadSession stores the short-lived package download metadata.
type marketplaceDownloadSession struct {
	SessionID    string `json:"sessionId"`
	PluginID     string `json:"pluginId"`
	Version      string `json:"version"`
	ArtifactType string `json:"artifactType"`
	SizeBytes    int64  `json:"sizeBytes"`
	Sha256       string `json:"sha256"`
	Status       string `json:"status"`
	DownloadURL  string `json:"downloadUrl"`
}

// normalizeMarketplaceBaseURL validates and normalizes a marketplace service URL.
func normalizeMarketplaceBaseURL(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("marketplace url is empty")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse marketplace url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("marketplace url must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("marketplace url host is required")
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

// installPlanMarketplaceItems installs marketplace-origin plan items via distribution API.
func installPlanMarketplaceItems(
	ctx context.Context,
	runtime Runtime,
	input Input,
	items []pluginPlanItem,
	update bool,
	force bool,
) error {
	if len(items) == 0 {
		return nil
	}
	action := "Installing"
	if update {
		action = "Updating"
	}
	if _, err := fmt.Fprintf(runtime.Stdout, "Preparing marketplace plugin %s for %d item(s)...\n", strings.ToLower(action), len(items)); err != nil {
		return fmt.Errorf("write plugin progress: %w", err)
	}
	token := toolutil.FirstNonEmpty(strings.TrimSpace(input.Get("token")), toolutil.EnvValue(runtime.Env, "LINAPRO_MARKETPLACE_TOKEN"))
	client := &http.Client{Timeout: 60 * time.Second}
	for index, item := range items {
		if item.All {
			return fmt.Errorf("marketplace origin %s does not support wildcard items", item.Source)
		}
		if strings.TrimSpace(item.URL) == "" {
			return fmt.Errorf("marketplace origin %s is missing url", item.Source)
		}
		if strings.TrimSpace(item.Version) == "" {
			return fmt.Errorf("marketplace install for plugin %s requires version", item.ID)
		}
		if _, err := fmt.Fprintf(runtime.Stdout, "[%d/%d] %s plugin %s@%s from origin %s (%s)...\n", index+1, len(items), strings.ToLower(action), item.ID, item.Version, item.Source, item.URL); err != nil {
			return fmt.Errorf("write plugin progress: %w", err)
		}
		if err := installOneMarketplaceRelease(ctx, runtime, client, item.URL, token, item.Source, item.ID, item.Version, update, force); err != nil {
			return err
		}
	}
	return nil
}

func installOneMarketplaceRelease(
	ctx context.Context,
	runtime Runtime,
	client *http.Client,
	baseURL string,
	token string,
	originName string,
	pluginID string,
	version string,
	update bool,
	force bool,
) error {
	distribution, err := fetchMarketplaceDistribution(ctx, client, baseURL, pluginID, version, token)
	if err != nil {
		return err
	}
	if strings.TrimSpace(distribution.PluginID) == "" {
		return fmt.Errorf("marketplace distribution is missing pluginId")
	}
	if strings.TrimSpace(distribution.Version) == "" {
		return fmt.Errorf("marketplace distribution is missing version")
	}
	if distribution.PluginID != pluginID {
		return fmt.Errorf("marketplace distribution plugin id mismatch: got %s want %s", distribution.PluginID, pluginID)
	}
	if distribution.Version != version {
		return fmt.Errorf("marketplace distribution version mismatch: got %s want %s", distribution.Version, version)
	}

	switch strings.ToLower(strings.TrimSpace(distribution.Mode)) {
	case "git":
		return installMarketplaceGitRelease(ctx, runtime, originName, distribution, update, force)
	case "https":
		return installMarketplaceHTTPSRelease(ctx, runtime, client, baseURL, token, originName, distribution, update, force)
	default:
		return fmt.Errorf("unsupported marketplace distribution mode %q for %s@%s", distribution.Mode, pluginID, version)
	}
}

func fetchMarketplaceDistribution(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	pluginID string,
	version string,
	token string,
) (*marketplaceDistribution, error) {
	endpoint, err := marketplaceURL(baseURL, "plugins", pluginID, "releases", version, "distribution")
	if err != nil {
		return nil, err
	}
	var response marketplaceDistributionResponse
	if err = marketplaceJSON(ctx, client, http.MethodGet, endpoint, token, nil, &response); err != nil {
		return nil, err
	}
	if response.Distribution == nil {
		return nil, fmt.Errorf("marketplace distribution response is missing distribution")
	}
	return response.Distribution, nil
}

func installMarketplaceGitRelease(
	ctx context.Context,
	runtime Runtime,
	originName string,
	distribution *marketplaceDistribution,
	update bool,
	force bool,
) error {
	if strings.TrimSpace(distribution.RepoURL) == "" {
		return fmt.Errorf("marketplace git distribution is missing repoUrl")
	}
	if strings.TrimSpace(distribution.Ref) == "" {
		return fmt.Errorf("marketplace git distribution is missing ref")
	}
	root, err := marketplaceGitRoot(distribution.PluginID, distribution.Path)
	if err != nil {
		return err
	}
	if strings.TrimSpace(originName) == "" {
		return fmt.Errorf("marketplace install requires origin name")
	}
	item := pluginPlanItem{
		ID:      distribution.PluginID,
		Source:  originName,
		Type:    config.OriginTypeMarketplace,
		Repo:    distribution.RepoURL,
		Root:    root,
		Version: distribution.Version,
		GitRef:  distribution.Ref,
	}
	checkout, err := checkoutPluginSource(ctx, runtime, item)
	if err != nil {
		return err
	}
	lock, err := ReadLock(runtime.Root)
	if err != nil {
		return err
	}
	if err = applyPluginFromCheckout(ctx, runtime, item, checkout, &lock, update, force); err != nil {
		return err
	}
	return WriteLock(runtime.Root, lock)
}

func installMarketplaceHTTPSRelease(
	ctx context.Context,
	runtime Runtime,
	client *http.Client,
	baseURL string,
	token string,
	originName string,
	distribution *marketplaceDistribution,
	update bool,
	force bool,
) error {
	if strings.TrimSpace(originName) == "" {
		return fmt.Errorf("marketplace install requires origin name")
	}
	expectedSHA := strings.ToLower(strings.TrimSpace(distribution.Sha256))
	if expectedSHA == "" {
		return fmt.Errorf("marketplace https distribution is missing sha256")
	}
	session, err := createMarketplaceDownloadSession(ctx, client, baseURL, token, distribution)
	if err != nil {
		return err
	}
	if strings.TrimSpace(session.Sha256) == "" {
		return fmt.Errorf("marketplace download session is missing sha256")
	}
	if !strings.EqualFold(session.Sha256, expectedSHA) {
		return fmt.Errorf("marketplace download session sha256 mismatch: got %s want %s", session.Sha256, expectedSHA)
	}

	tempParent := filepath.Join(runtime.Root, "temp")
	if err = os.MkdirAll(tempParent, 0o755); err != nil {
		return fmt.Errorf("create marketplace temp parent: %w", err)
	}
	tempDir, err := os.MkdirTemp(tempParent, "marketplace-install-*")
	if err != nil {
		return fmt.Errorf("create marketplace temp dir: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			fmt.Fprintf(runtime.Stderr, "warning: remove marketplace temp dir: %v\n", removeErr)
		}
	}()

	packagePath := filepath.Join(tempDir, "package")
	downloadURL, err := marketplaceDownloadURL(baseURL, session)
	if err != nil {
		return err
	}
	actualSHA, err := downloadMarketplacePackage(ctx, client, downloadURL, token, packagePath)
	if err != nil {
		return err
	}
	if actualSHA != expectedSHA {
		return fmt.Errorf("marketplace package sha256 mismatch: got %s want %s", actualSHA, expectedSHA)
	}

	extractDir := filepath.Join(tempDir, "extract")
	artifactType := strings.TrimSpace(distribution.ArtifactType)
	if artifactType == "" {
		artifactType = strings.TrimSpace(session.ArtifactType)
	}
	if artifactType == "" {
		return fmt.Errorf("marketplace https distribution is missing artifactType")
	}
	if err = extractMarketplacePackage(packagePath, artifactType, extractDir); err != nil {
		return err
	}
	sourceDir, err := marketplaceExtractedPluginRoot(extractDir, distribution.PluginID)
	if err != nil {
		return err
	}
	lock, err := ReadLock(runtime.Root)
	if err != nil {
		return err
	}
	// HTTPS packages have no Git ref; leave GitRef empty so lock.ref stays empty.
	item := pluginPlanItem{
		ID:      distribution.PluginID,
		Source:  originName,
		Type:    config.OriginTypeMarketplace,
		Repo:    baseURL,
		Root:    ".",
		Version: distribution.Version,
		GitRef:  "",
	}
	if err = applyPluginDirectory(ctx, runtime, item, sourceDir, expectedSHA, &lock, update, force); err != nil {
		return err
	}
	return WriteLock(runtime.Root, lock)
}

func createMarketplaceDownloadSession(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token string,
	distribution *marketplaceDistribution,
) (*marketplaceDownloadSession, error) {
	endpoint, err := marketplaceURL(baseURL, "plugins", distribution.PluginID, "releases", distribution.Version, "downloads")
	if err != nil {
		return nil, err
	}
	payload := map[string]string{}
	if distribution.ArtifactType != "" {
		payload["artifactType"] = distribution.ArtifactType
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal marketplace download request: %w", err)
	}
	var response marketplaceDownloadSessionResponse
	if err = marketplaceJSON(ctx, client, http.MethodPost, endpoint, token, bytes.NewReader(body), &response); err != nil {
		return nil, err
	}
	if response.Session == nil {
		return nil, fmt.Errorf("marketplace download response is missing session")
	}
	return response.Session, nil
}

func marketplaceJSON(
	ctx context.Context,
	client *http.Client,
	method string,
	endpoint string,
	token string,
	body io.Reader,
	target interface{},
) error {
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return fmt.Errorf("create marketplace request: %w", err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request marketplace API %s: %w", endpoint, err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace response: %v\n", closeErr)
		}
	}()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		content, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("marketplace API %s returned %s: %s", endpoint, response.Status, strings.TrimSpace(string(content)))
	}
	if err = json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode marketplace response: %w", err)
	}
	return nil
}

func downloadMarketplacePackage(ctx context.Context, client *http.Client, endpoint string, token string, dst string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create marketplace package request: %w", err)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("download marketplace package: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace package response: %v\n", closeErr)
		}
	}()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		content, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return "", fmt.Errorf("marketplace package download returned %s: %s", response.Status, strings.TrimSpace(string(content)))
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("create marketplace package file: %w", err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace package file: %v\n", closeErr)
		}
	}()
	hasher := sha256.New()
	if _, err = io.Copy(io.MultiWriter(out, hasher), response.Body); err != nil {
		return "", fmt.Errorf("write marketplace package: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func marketplaceURL(baseURL string, parts ...string) (string, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse marketplace base URL: %w", err)
	}
	escaped := make([]string, 0, len(parts)+1)
	escaped = append(escaped, marketplaceAPIPathPrefix)
	for _, part := range parts {
		escaped = append(escaped, url.PathEscape(part))
	}
	parsed.Path = path.Join(append([]string{strings.TrimRight(parsed.Path, "/")}, escaped...)...)
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func marketplaceDownloadURL(baseURL string, session *marketplaceDownloadSession) (string, error) {
	if session == nil || strings.TrimSpace(session.SessionID) == "" {
		return "", fmt.Errorf("marketplace download session is missing sessionId")
	}
	value := strings.TrimSpace(session.DownloadURL)
	if value == "" {
		return marketplaceURL(baseURL, "download-sessions", session.SessionID, "content")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse marketplace download URL: %w", err)
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse marketplace base URL: %w", err)
	}
	return base.ResolveReference(parsed).String(), nil
}

func marketplaceGitRoot(pluginID string, distributionPath string) (string, error) {
	value := strings.TrimSpace(strings.ReplaceAll(distributionPath, "\\", "/"))
	if value == "" || value == "." {
		return ".", nil
	}
	if strings.Contains(value, ":") || path.IsAbs(value) {
		return "", fmt.Errorf("marketplace distribution path %q is not a repository-relative path", distributionPath)
	}
	cleaned := path.Clean(value)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return "", fmt.Errorf("marketplace distribution path %q escapes the repository root", distributionPath)
	}
	if path.Base(cleaned) != pluginID {
		return "", fmt.Errorf("marketplace distribution path %q does not end with plugin id %s", distributionPath, pluginID)
	}
	root := path.Dir(cleaned)
	if root == "." {
		return ".", nil
	}
	return root, nil
}

func extractMarketplacePackage(packagePath string, artifactType string, dst string) error {
	normalizedType := strings.ToLower(strings.TrimSpace(artifactType))
	if normalizedType == "plugin_wasm" {
		return fmt.Errorf("marketplace artifact type plugin_wasm cannot be installed as a source workspace package")
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create marketplace extract dir: %w", err)
	}
	if strings.Contains(normalizedType, "tar_gz") || strings.Contains(normalizedType, "tgz") {
		return extractMarketplaceTarGzip(packagePath, dst)
	}
	if strings.Contains(normalizedType, "zip") {
		return extractMarketplaceZip(packagePath, dst)
	}
	if err := extractMarketplaceZip(packagePath, dst); err == nil {
		return nil
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("reset marketplace extract dir: %w", err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("recreate marketplace extract dir: %w", err)
	}
	return extractMarketplaceTarGzip(packagePath, dst)
}

func extractMarketplaceZip(packagePath string, dst string) error {
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return fmt.Errorf("open marketplace zip package: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace zip package: %v\n", closeErr)
		}
	}()
	for _, file := range reader.File {
		target, err := safeMarketplaceExtractPath(dst, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create marketplace zip directory: %w", err)
			}
			continue
		}
		if file.FileInfo().Mode()&os.ModeType != 0 {
			return fmt.Errorf("marketplace zip package contains unsupported entry %s", file.Name)
		}
		if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create marketplace zip parent: %w", err)
		}
		source, err := file.Open()
		if err != nil {
			return fmt.Errorf("open marketplace zip entry: %w", err)
		}
		if err = writeExtractedFile(target, source); err != nil {
			_ = source.Close()
			return err
		}
		if closeErr := source.Close(); closeErr != nil {
			return fmt.Errorf("close marketplace zip entry: %w", closeErr)
		}
	}
	return nil
}

func extractMarketplaceTarGzip(packagePath string, dst string) error {
	file, err := os.Open(packagePath)
	if err != nil {
		return fmt.Errorf("open marketplace tar.gz package: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace tar.gz package: %v\n", closeErr)
		}
	}()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open marketplace gzip stream: %w", err)
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close marketplace gzip stream: %v\n", closeErr)
		}
	}()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read marketplace tar entry: %w", err)
		}
		target, err := safeMarketplaceExtractPath(dst, header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("create marketplace tar directory: %w", err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("create marketplace tar parent: %w", err)
			}
			if err = writeExtractedFile(target, tarReader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("marketplace tar package contains unsupported entry %s", header.Name)
		}
	}
	return nil
}

func writeExtractedFile(target string, source io.Reader) error {
	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create extracted marketplace file: %w", err)
	}
	defer func() {
		if closeErr := output.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close extracted marketplace file: %v\n", closeErr)
		}
	}()
	if _, err = io.Copy(output, source); err != nil {
		return fmt.Errorf("write extracted marketplace file: %w", err)
	}
	return nil
}

func safeMarketplaceExtractPath(root string, name string) (string, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	if normalized == "" {
		return "", fmt.Errorf("marketplace package contains empty path")
	}
	if strings.Contains(normalized, ":") || path.IsAbs(normalized) {
		return "", fmt.Errorf("marketplace package path %q is not relative", name)
	}
	cleaned := path.Clean(normalized)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("marketplace package path %q escapes extraction root", name)
	}
	target := filepath.Join(root, filepath.FromSlash(cleaned))
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("validate marketplace package path: %w", err)
	}
	if strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", fmt.Errorf("marketplace package path %q escapes extraction root", name)
	}
	return target, nil
}

func marketplaceExtractedPluginRoot(extractDir string, pluginID string) (string, error) {
	var found string
	err := filepath.WalkDir(extractDir, func(current string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if found != "" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() || entry.Name() != "plugin.yaml" {
			return nil
		}
		manifest, readErr := ReadManifest(current)
		if readErr != nil {
			return readErr
		}
		if manifest.ID == "" || manifest.ID == pluginID {
			found = filepath.Dir(current)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("scan marketplace package plugin root: %w", err)
	}
	if found == "" {
		return "", fmt.Errorf("marketplace package does not contain plugin.yaml for %s", pluginID)
	}
	if !fileutil.FileExists(filepath.Join(found, "plugin.yaml")) {
		return "", fmt.Errorf("marketplace package plugin root is invalid")
	}
	return found, nil
}
