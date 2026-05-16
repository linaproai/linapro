// This file implements bounded output buffering for shell executions. The
// buffer accepts full writes from exec while retaining only the configured
// prefix plus a truncation marker for persistence.

package shellexec

import "bytes"

// limitedBuffer captures at most a fixed number of bytes and appends one
// truncation marker once the limit is exceeded.
type limitedBuffer struct {
	limit     int          // limit bounds the captured payload size.
	buffer    bytes.Buffer // buffer stores the captured output.
	truncated bool         // truncated reports whether the marker was already appended.
}

// newLimitedBuffer creates one bounded output buffer.
func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

// Write captures as much of the input as still fits in the buffer limit.
func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	if b == nil {
		return len(p), nil
	}
	if b.buffer.Len() >= b.limit {
		if !b.truncated {
			if _, err := b.buffer.WriteString(truncatedOutputMarker); err != nil {
				return 0, err
			}
			b.truncated = true
		}
		return len(p), nil
	}

	remaining := b.limit - b.buffer.Len()
	if len(p) <= remaining {
		if _, err := b.buffer.Write(p); err != nil {
			return 0, err
		}
		return len(p), nil
	}

	if _, err := b.buffer.Write(p[:remaining]); err != nil {
		return 0, err
	}
	if !b.truncated {
		if _, err := b.buffer.WriteString(truncatedOutputMarker); err != nil {
			return 0, err
		}
		b.truncated = true
	}
	return len(p), nil
}

// String returns the captured output.
func (b *limitedBuffer) String() string {
	if b == nil {
		return ""
	}
	return b.buffer.String()
}
