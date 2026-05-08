// Package wasmsection provides internal WASM custom-section scanning helpers
// for the plugin bridge artifact component.
package wasmsection

import "github.com/gogf/gf/v2/errors/gerror"

const (
	// CustomSectionID identifies custom sections in a WASM module.
	CustomSectionID byte = 0
	// HeaderLength is the fixed byte length of the WASM magic and version header.
	HeaderLength = 8
	// Magic stores the canonical WASM binary magic prefix.
	Magic = "\x00asm"
	// Version1 stores the canonical version bytes supported by LinaPro plugin artifacts.
	Version1 = "\x01\x00\x00\x00"
)

// ListCustomSections extracts all WASM custom section payloads by section name.
func ListCustomSections(content []byte) (map[string][]byte, error) {
	if len(content) < HeaderLength {
		return nil, gerror.New("wasm file is too short")
	}
	if string(content[:4]) != Magic {
		return nil, gerror.New("invalid wasm header")
	}
	if string(content[4:HeaderLength]) != Version1 {
		return nil, gerror.New("invalid wasm version")
	}

	sections := make(map[string][]byte)
	cursor := HeaderLength
	for cursor < len(content) {
		sectionID := content[cursor]
		cursor++

		sectionSize, nextCursor, err := readULEB128(content, cursor)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor

		end := cursor + int(sectionSize)
		if end > len(content) {
			return nil, gerror.New("wasm section length exceeds content")
		}
		if sectionID == CustomSectionID {
			nameLength, nameCursor, err := readULEB128(content, cursor)
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

// readULEB128 decodes one unsigned LEB128 integer from a WASM byte stream.
func readULEB128(content []byte, start int) (uint32, int, error) {
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
