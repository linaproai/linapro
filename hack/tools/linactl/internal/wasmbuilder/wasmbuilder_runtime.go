// This file builds the optional guest runtime wasm module and derives the
// bridge ABI contract advertised by the final artifact.

package wasmbuilder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

const (
	guestRuntimeBuildLockRetryInterval = 100 * time.Millisecond
	guestRuntimeBuildLockStaleAfter    = 30 * time.Minute
	guestRuntimeBuildLockTimeout       = 2 * time.Minute
)

func buildGuestRuntimeWasm(
	pluginDir string,
	pluginID string,
	outputDir string,
	routeSources []*routeContractSource,
	lifecycleSpecs []*protocol.LifecycleContract,
) (runtimePath string, err error) {
	// The WASM guest runtime entry (main.go) lives at the plugin root
	// directory.
	mainGoPath := filepath.Join(pluginDir, "main.go")
	if _, err := os.Stat(mainGoPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	outputPath, err := resolveGuestRuntimeOutputPath(pluginDir, pluginID, outputDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}
	releaseBuildLock, err := acquireGuestRuntimeBuildLock(pluginDir)
	if err != nil {
		return "", err
	}
	defer func() {
		if releaseBuildLock != nil {
			if releaseErr := releaseBuildLock(); releaseErr != nil && err == nil {
				err = releaseErr
			}
		}
	}()
	cleanupDispatcher, err := prepareGeneratedWasmDispatcher(pluginDir, pluginID, routeSources, lifecycleSpecs)
	if err != nil {
		return "", err
	}
	defer func() {
		if cleanupDispatcher != nil {
			if cleanupErr := cleanupDispatcher(); cleanupErr != nil && err == nil {
				err = cleanupErr
			}
		}
	}()
	buildDir := pluginDir
	buildTarget := "."
	buildEnv := append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	if workMode := selectGuestRuntimeGoWork(pluginDir); workMode != "" {
		buildEnv = append(buildEnv, "GOWORK="+workMode)
	}
	var (
		goModPath       = filepath.Join(pluginDir, "go.mod")
		goSumPath       = filepath.Join(pluginDir, "go.sum")
		removeTempGoMod bool
		removeTempGoSum bool
	)
	if _, goModErr := os.Stat(goModPath); os.IsNotExist(goModErr) {
		// When the plugin root has no go.mod (e.g. synthetic test directories),
		// create a minimal one so that 'go build' can proceed.
		goModContent := "module lina-plugin-runtime-guest\n\ngo 1.25.0\n"
		if _, goSumErr := os.Stat(goSumPath); os.IsNotExist(goSumErr) {
			removeTempGoSum = true
		}
		if writeErr := os.WriteFile(goModPath, []byte(goModContent), 0o644); writeErr != nil {
			return "", writeErr
		}
		removeTempGoMod = true
		buildEnv = append(buildEnv, "GOWORK=off")
	}
	defer func() {
		if removeTempGoMod {
			if removeErr := os.Remove(goModPath); removeErr != nil && err == nil && !os.IsNotExist(removeErr) {
				err = removeErr
			}
		}
		if removeTempGoSum {
			if removeErr := os.Remove(goSumPath); removeErr != nil && err == nil && !os.IsNotExist(removeErr) {
				err = removeErr
			}
		}
	}()
	cmd := exec.Command("go", "build", "-buildmode=c-shared", "-o", outputPath, buildTarget)
	cmd.Dir = buildDir
	cmd.Env = buildEnv
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to build dynamic guest runtime: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return outputPath, nil
}

func acquireGuestRuntimeBuildLock(pluginDir string) (func() error, error) {
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return nil, err
	}
	normalizedPluginDir := filepath.Clean(absPluginDir)
	lockHash := sha256.Sum256([]byte(normalizedPluginDir))
	lockRoot := filepath.Join(os.TempDir(), "linapro-wasm-build-locks")
	if err = os.MkdirAll(lockRoot, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create wasm build lock root: %w", err)
	}
	lockDir := filepath.Join(lockRoot, hex.EncodeToString(lockHash[:]))
	deadline := time.Now().Add(guestRuntimeBuildLockTimeout)

	for {
		if err = os.Mkdir(lockDir, 0o700); err == nil {
			ownerPath := filepath.Join(lockDir, "owner")
			owner := fmt.Sprintf("pid=%d\npluginDir=%s\n", os.Getpid(), normalizedPluginDir)
			if writeErr := os.WriteFile(ownerPath, []byte(owner), 0o600); writeErr != nil {
				if removeErr := os.RemoveAll(lockDir); removeErr != nil {
					return nil, fmt.Errorf("failed to write wasm build lock owner: %w; cleanup failed: %v", writeErr, removeErr)
				}
				return nil, fmt.Errorf("failed to write wasm build lock owner: %w", writeErr)
			}
			return func() error {
				if removeErr := os.RemoveAll(lockDir); removeErr != nil && !os.IsNotExist(removeErr) {
					return fmt.Errorf("failed to release wasm build lock: %w", removeErr)
				}
				return nil
			}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf("failed to acquire wasm build lock: %w", err)
		}
		if stale, staleErr := guestRuntimeBuildLockStale(lockDir); staleErr != nil {
			return nil, staleErr
		} else if stale {
			if removeErr := os.RemoveAll(lockDir); removeErr != nil && !os.IsNotExist(removeErr) {
				return nil, fmt.Errorf("failed to remove stale wasm build lock: %w", removeErr)
			}
			continue
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for wasm build lock for %s", normalizedPluginDir)
		}
		time.Sleep(guestRuntimeBuildLockRetryInterval)
	}
}

func guestRuntimeBuildLockStale(lockDir string) (bool, error) {
	info, err := os.Stat(lockDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat wasm build lock: %w", err)
	}
	return time.Since(info.ModTime()) > guestRuntimeBuildLockStaleAfter, nil
}

// selectGuestRuntimeGoWork chooses the Go workspace mode used to build the
// guest runtime module. Official plugin runtime modules use the temporary
// plugin workspace generated by linactl, keeping the root workspace host-only.
func selectGuestRuntimeGoWork(pluginDir string) string {
	absolutePluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		return "off"
	}
	normalizedPluginDir := filepath.Clean(absolutePluginDir)

	repoRoot, ok := findRuntimeBuildRepoRoot(normalizedPluginDir)
	if !ok {
		return "off"
	}
	officialPluginsRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	relativePath, err := filepath.Rel(officialPluginsRoot, normalizedPluginDir)
	if err == nil && relativePath != "." && relativePath != "" && !strings.HasPrefix(relativePath, "..") {
		return filepath.Join(repoRoot, "temp", "go.work.plugins")
	}
	return "off"
}

func buildBridgeSpec(runtimePath string) *protocol.BridgeSpec {
	spec := &protocol.BridgeSpec{
		ABIVersion:  protocol.ABIVersionV1,
		RuntimeKind: protocol.RuntimeKindWasm,
	}
	if strings.TrimSpace(runtimePath) != "" {
		spec.RouteExecution = true
		spec.RequestCodec = protocol.CodecProtobuf
		spec.ResponseCodec = protocol.CodecProtobuf
		spec.AllocExport = protocol.DefaultGuestAllocExport
		spec.ExecuteExport = protocol.DefaultGuestExecuteExport
	}
	protocol.NormalizeBridgeSpec(spec)
	return spec
}
