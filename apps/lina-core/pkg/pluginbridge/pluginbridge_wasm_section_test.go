// This file tests shared WASM custom-section parsing helpers.

package pluginbridge

import (
	"bytes"
	"strings"
	"testing"
)

// TestListCustomSectionsReturnsPayloads verifies named custom-section payloads
// are copied out of a valid WASM module.
func TestListCustomSectionsReturnsPayloads(t *testing.T) {
	t.Parallel()

	content := []byte(wasmMagic + wasmVersion1)
	content = appendTestPluginBridgeWasmCustomSection(content, WasmSectionManifest, []byte(`{"pluginId":"demo"}`))
	content = appendTestPluginBridgeWasmCustomSection(content, WasmSectionI18NAssets, []byte(`[{"locale":"en-US"}]`))

	sections, err := ListCustomSections(content)
	if err != nil {
		t.Fatalf("expected custom sections to parse: %v", err)
	}
	if actual := string(sections[WasmSectionManifest]); actual != `{"pluginId":"demo"}` {
		t.Fatalf("unexpected manifest section payload: %s", actual)
	}

	payload, ok, err := ReadCustomSection(content, WasmSectionI18NAssets)
	if err != nil {
		t.Fatalf("expected named custom section read to succeed: %v", err)
	}
	if !ok || !bytes.Equal(payload, []byte(`[{"locale":"en-US"}]`)) {
		t.Fatalf("unexpected named section payload: exists=%v payload=%s", ok, string(payload))
	}
}

// TestReadCustomSectionRejectsInvalidHeader verifies malformed WASM files fail
// before custom-section lookup.
func TestReadCustomSectionRejectsInvalidHeader(t *testing.T) {
	t.Parallel()

	_, _, err := ReadCustomSection([]byte("not-wasm"), WasmSectionManifest)
	if err == nil || !strings.Contains(err.Error(), "invalid wasm header") {
		t.Fatalf("expected invalid header error, got %v", err)
	}
}

// TestListCustomSectionsRejectsULEBOverflow verifies malformed ULEB128 lengths
// fail with a bounded parse error.
func TestListCustomSectionsRejectsULEBOverflow(t *testing.T) {
	t.Parallel()

	content := []byte(wasmMagic + wasmVersion1)
	content = append(content, wasmCustomSectionID)
	content = append(content, 0x80, 0x80, 0x80, 0x80, 0x80)

	_, err := ListCustomSections(content)
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Fatalf("expected ULEB128 overflow error, got %v", err)
	}
}

// appendTestPluginBridgeWasmCustomSection appends one custom section to test WASM content.
func appendTestPluginBridgeWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeTestPluginBridgeWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, wasmCustomSectionID)
	result = append(result, encodeTestPluginBridgeWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodeTestPluginBridgeWasmULEB128 encodes one uint32 with WASM ULEB128 encoding.
func encodeTestPluginBridgeWasmULEB128(value uint32) []byte {
	if value == 0 {
		return []byte{0}
	}
	result := make([]byte, 0)
	for value > 0 {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
	}
	return result
}
