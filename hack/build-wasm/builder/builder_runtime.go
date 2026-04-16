// This file builds the optional guest runtime wasm module and derives the
// bridge ABI contract advertised by the final artifact.

package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"lina-core/pkg/pluginbridge"
)

func buildGuestRuntimeWasm(pluginDir string, pluginID string, outputDir string) (string, error) {
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
	buildDir := pluginDir
	buildTarget := "."
	buildEnv := append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
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
			_ = os.Remove(goModPath)
		}
		if removeTempGoSum {
			_ = os.Remove(goSumPath)
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

func buildBridgeSpec(runtimePath string) *pluginbridge.BridgeSpec {
	spec := &pluginbridge.BridgeSpec{
		ABIVersion:  pluginbridge.ABIVersionV1,
		RuntimeKind: pluginbridge.RuntimeKindWasm,
	}
	if strings.TrimSpace(runtimePath) != "" {
		spec.RouteExecution = true
		spec.RequestCodec = pluginbridge.CodecProtobuf
		spec.ResponseCodec = pluginbridge.CodecProtobuf
		spec.AllocExport = pluginbridge.DefaultGuestAllocExport
		spec.ExecuteExport = pluginbridge.DefaultGuestExecuteExport
	}
	pluginbridge.NormalizeBridgeSpec(spec)
	return spec
}
