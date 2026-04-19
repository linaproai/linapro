// This file covers request-body limit calculation for multipart and non-multipart requests.

package middleware

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
)

// TestRequestBodyLimitForContentType verifies body-limit selection for
// multipart and non-multipart requests.
func TestRequestBodyLimitForContentType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		contentType     string
		uploadMaxSizeMB int64
		expected        int64
	}{
		{
			name:            "json keeps default request ceiling",
			contentType:     "application/json",
			uploadMaxSizeMB: 10,
			expected:        defaultRequestBodyLimitBytes,
		},
		{
			name:            "multipart reserves upload envelope overhead",
			contentType:     "multipart/form-data; boundary=----WebKitFormBoundary",
			uploadMaxSizeMB: 10,
			expected:        11 * bytesPerMegabyte,
		},
		{
			name:            "invalid upload size falls back to default ceiling",
			contentType:     "multipart/form-data",
			uploadMaxSizeMB: 0,
			expected:        defaultRequestBodyLimitBytes,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actual := requestBodyLimitForContentType(testCase.contentType, testCase.uploadMaxSizeMB)
			if actual != testCase.expected {
				t.Fatalf("expected request-body limit %d, got %d", testCase.expected, actual)
			}
		})
	}
}

// TestIsMultipartContentType verifies multipart media types are detected
// regardless of casing.
func TestIsMultipartContentType(t *testing.T) {
	t.Parallel()

	if !isMultipartContentType("Multipart/Form-Data; boundary=abc") {
		t.Fatal("expected multipart content type to be detected")
	}
	if isMultipartContentType("application/json") {
		t.Fatal("expected non-multipart content type not to match")
	}
}

// TestRequestBodyLimitFriendlyError verifies multipart overflows are converted
// into user-facing upload-size validation errors.
func TestRequestBodyLimitFriendlyError(t *testing.T) {
	t.Parallel()

	err := requestBodyLimitFriendlyError(
		"multipart/form-data; boundary=abc",
		gerror.Wrap(&http.MaxBytesError{Limit: 11 * bytesPerMegabyte}, "r.ParseMultipartForm failed"),
		10,
	)
	if err == nil {
		t.Fatal("expected multipart size overflow to map to friendly error")
	}
	if !strings.Contains(err.Error(), "文件大小不能超过10MB") {
		t.Fatalf("expected friendly size error, got %v", err)
	}
}

// TestRequestBodyLimitFriendlyErrorIgnoresNonMultipartRequests verifies
// friendly overflow translation does not affect non-multipart requests.
func TestRequestBodyLimitFriendlyErrorIgnoresNonMultipartRequests(t *testing.T) {
	t.Parallel()

	err := requestBodyLimitFriendlyError(
		"application/json",
		&http.MaxBytesError{Limit: defaultRequestBodyLimitBytes},
		10,
	)
	if err != nil {
		t.Fatalf("expected non-multipart overflow to remain unhandled, got %v", err)
	}
}
