// This file exposes shared WASM custom-section readers for dynamic plugin
// artifact discovery.

package pluginbridge

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// wasmCustomSectionID identifies custom sections in a WASM module.
	wasmCustomSectionID byte = 0
	// wasmHeaderLength is the fixed byte length of the WASM magic and version header.
	wasmHeaderLength = 8
	// wasmMagic stores the canonical WASM binary magic prefix.
	wasmMagic = "\x00asm"
	// wasmVersion1 stores the canonical version bytes supported by LinaPro plugin artifacts.
	wasmVersion1 = "\x01\x00\x00\x00"
)

// ReadCustomSection returns one named WASM custom section payload.
func ReadCustomSection(content []byte, name string) ([]byte, bool, error) {
	sections, err := ListCustomSections(content)
	if err != nil {
		return nil, false, err
	}
	payload, ok := sections[strings.TrimSpace(name)]
	if !ok {
		return nil, false, nil
	}
	return payload, true, nil
}

// ListCustomSections extracts all WASM custom section payloads by section name.
func ListCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < wasmHeaderLength {
		return nil, gerror.New("wasm file is too short")
	}
	if string(content[:4]) != wasmMagic {
		return nil, gerror.New("invalid wasm header")
	}
	if string(content[4:wasmHeaderLength]) != wasmVersion1 {
		return nil, gerror.New("invalid wasm version")
	}

	sections := make(map[string][]byte)
	cursor := wasmHeaderLength
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readWasmULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, gerror.New("wasm section length exceeds content")
		}
		if sectionID == wasmCustomSectionID {
			nameLength, nameCursor, err := readWasmULEB128(content, cursor)
			if err != nil {
				return nil, err
			}
			nameEnd := nameCursor + int(nameLength)
			if nameEnd > end {
				return nil, gerror.New("wasm custom section name exceeds content")
			}
			sectionName := string(content[nameCursor:nameEnd])
			sectionPayload := make([]byte, end-nameEnd)
			copy(sectionPayload, content[nameEnd:end])
			sections[sectionName] = sectionPayload
		}

		cursor = end
	}
	return sections, nil
}

// readWasmULEB128 decodes one unsigned LEB128 integer from a WASM byte stream.
func readWasmULEB128(content []byte, start int) (uint32, int, error) {
	var (
		value uint32
		shift uint
	)

	cursor := start
	for {
		if cursor >= len(content) {
			return 0, cursor, gerror.New("wasm ULEB128 data exceeds content")
		}
		current := content[cursor]
		cursor++

		value |= uint32(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, cursor, nil
		}

		shift += 7
		if shift >= 32 {
			return 0, cursor, gerror.New("wasm ULEB128 value is too large")
		}
	}
}
