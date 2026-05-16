// This file verifies that the root pluginbridge facade remains behaviorally
// identical to the focused subcomponents it forwards to.

package pluginbridge

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"

	"lina-core/pkg/pluginbridge/artifact"
	"lina-core/pkg/pluginbridge/codec"
	"lina-core/pkg/pluginbridge/hostservice"
)

// TestFacadeRequestEnvelopeMatchesCodec verifies root request envelope helpers
// are direct behavioral equivalents of the codec subcomponent.
func TestFacadeRequestEnvelopeMatchesCodec(t *testing.T) {
	t.Parallel()

	request := &BridgeRequestEnvelopeV1{
		PluginID: "plugin-demo-dynamic",
		Route: &RouteMatchSnapshotV1{
			Method:       http.MethodPost,
			PublicPath:   "/api/v1/extensions/plugin-demo-dynamic/items/42",
			InternalPath: "/items/:id",
			RoutePath:    "/items/{id}",
			Access:       AccessLogin,
			Permission:   "plugin-demo-dynamic:item:update",
			RequestType:  "UpdateItemReq",
			PathParams: map[string]string{
				"id": "42",
			},
			QueryValues: map[string][]string{
				"verbose": {"true"},
			},
		},
		Request: &HTTPRequestSnapshotV1{
			Method:      http.MethodPost,
			PublicPath:  "/api/v1/extensions/plugin-demo-dynamic/items/42",
			RawPath:     "/api/v1/extensions/plugin-demo-dynamic/items/42",
			RawQuery:    "verbose=true",
			Host:        "localhost:8080",
			ContentType: "application/json",
			Headers: map[string][]string{
				"Accept": {"application/json"},
			},
			Body: []byte(`{"name":"demo"}`),
		},
		Identity: &IdentitySnapshotV1{
			UserID:      1,
			Username:    "admin",
			Permissions: []string{"plugin-demo-dynamic:item:update"},
		},
		RequestID: "req-facade",
	}

	facadeContent, err := EncodeRequestEnvelope(request)
	if err != nil {
		t.Fatalf("facade encode failed: %v", err)
	}
	componentContent, err := codec.EncodeRequestEnvelope(request)
	if err != nil {
		t.Fatalf("component encode failed: %v", err)
	}
	if !bytes.Equal(facadeContent, componentContent) {
		t.Fatalf("facade request bytes differ from codec bytes")
	}

	facadeDecoded, err := DecodeRequestEnvelope(facadeContent)
	if err != nil {
		t.Fatalf("facade decode failed: %v", err)
	}
	componentDecoded, err := codec.DecodeRequestEnvelope(componentContent)
	if err != nil {
		t.Fatalf("component decode failed: %v", err)
	}
	if !reflect.DeepEqual(facadeDecoded, componentDecoded) {
		t.Fatalf("facade decoded request differs from codec decoded request")
	}
}

// TestFacadeWasmSectionMatchesArtifact verifies root WASM custom-section
// helpers return the same payloads as the artifact subcomponent.
func TestFacadeWasmSectionMatchesArtifact(t *testing.T) {
	t.Parallel()

	content := []byte("\x00asm\x01\x00\x00\x00")
	content = appendFacadeTestWasmCustomSection(content, WasmSectionManifest, []byte(`{"id":"demo"}`))
	content = appendFacadeTestWasmCustomSection(content, WasmSectionI18NAssets, []byte(`{"zh-CN":{}}`))
	content = appendFacadeTestWasmCustomSection(content, WasmSectionBackendLifecycle, []byte(`[]`))

	facadeSections, err := ListCustomSections(content)
	if err != nil {
		t.Fatalf("facade list sections failed: %v", err)
	}
	componentSections, err := artifact.ListCustomSections(content)
	if err != nil {
		t.Fatalf("artifact list sections failed: %v", err)
	}
	if !reflect.DeepEqual(facadeSections, componentSections) {
		t.Fatalf("facade sections differ from artifact sections")
	}

	facadePayload, facadeFound, err := ReadCustomSection(content, WasmSectionI18NAssets)
	if err != nil {
		t.Fatalf("facade read section failed: %v", err)
	}
	componentPayload, componentFound, err := artifact.ReadCustomSection(content, artifact.WasmSectionI18NAssets)
	if err != nil {
		t.Fatalf("artifact read section failed: %v", err)
	}
	if facadeFound != componentFound || !bytes.Equal(facadePayload, componentPayload) {
		t.Fatalf("facade section payload differs from artifact payload")
	}
}

// TestFacadeHostServicePayloadMatchesSubcomponent verifies representative
// host-service payload helpers remain equivalent through the root facade.
func TestFacadeHostServicePayloadMatchesSubcomponent(t *testing.T) {
	t.Parallel()

	request := &HostServiceStoragePutRequest{
		Path:        "reports/facade.json",
		Body:        []byte(`{"ok":true}`),
		ContentType: "application/json",
		Overwrite:   true,
	}
	facadeContent := MarshalHostServiceStoragePutRequest(request)
	componentContent := hostservice.MarshalHostServiceStoragePutRequest(request)
	if !bytes.Equal(facadeContent, componentContent) {
		t.Fatalf("facade storage request bytes differ from hostservice bytes")
	}

	facadeDecoded, err := UnmarshalHostServiceStoragePutRequest(facadeContent)
	if err != nil {
		t.Fatalf("facade storage request decode failed: %v", err)
	}
	componentDecoded, err := hostservice.UnmarshalHostServiceStoragePutRequest(componentContent)
	if err != nil {
		t.Fatalf("hostservice storage request decode failed: %v", err)
	}
	if !reflect.DeepEqual(facadeDecoded, componentDecoded) {
		t.Fatalf("facade storage request differs from hostservice request")
	}
}

// appendFacadeTestWasmCustomSection appends one custom section to a minimal
// WASM module used by facade consistency tests.
func appendFacadeTestWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeFacadeTestWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0)
	result = append(result, encodeFacadeTestWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

// encodeFacadeTestWasmULEB128 encodes one uint32 using WASM ULEB128 encoding.
func encodeFacadeTestWasmULEB128(value uint32) []byte {
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
