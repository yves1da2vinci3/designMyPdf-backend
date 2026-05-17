package logs

import (
	"errors"
	"strings"
	"testing"
)

func TestFormatErrorWithBacktrace_nil(t *testing.T) {
	if got := FormatErrorWithBacktrace(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestFormatErrorWithBacktrace_includesMessageAndMarker(t *testing.T) {
	err := errors.New("pdf generation failed")
	got := FormatErrorWithBacktrace(err)

	if !strings.Contains(got, "pdf generation failed") {
		t.Fatalf("expected error message in output, got %q", got)
	}
	if !strings.Contains(got, "--- backtrace ---") {
		t.Fatalf("expected backtrace marker in output, got %q", got)
	}
	if !strings.Contains(got, "goroutine") {
		t.Fatalf("expected stack trace in output, got %q", got)
	}
}
