// This file computes deterministic fingerprints for cached runtime i18n bundles.

package i18n

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

const (
	// runtimeBundleFingerprintHexLength keeps ETags compact while preserving
	// enough SHA-256 entropy for runtime translation cache validators.
	runtimeBundleFingerprintHexLength = 32
)

// runtimeBundleFingerprint returns a deterministic content digest for one flat
// message catalog. The caller owns the input map and must not mutate it while
// the digest is being computed.
func runtimeBundleFingerprint(messages map[string]string) string {
	var builder strings.Builder
	keys := make([]string, 0, len(messages))
	for key := range messages {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	builder.WriteString("flat{")
	for _, key := range keys {
		appendRuntimeBundleFingerprintString(&builder, key)
		appendRuntimeBundleFingerprintString(&builder, messages[key])
	}
	builder.WriteString("}")

	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:])[:runtimeBundleFingerprintHexLength]
}

// appendRuntimeBundleFingerprintString writes a length-prefixed string so
// adjacent keys and values cannot collide in the fingerprint input.
func appendRuntimeBundleFingerprintString(builder *strings.Builder, value string) {
	builder.WriteString("str:")
	builder.WriteString(strconv.Itoa(len(value)))
	builder.WriteString(":")
	builder.WriteString(value)
}
