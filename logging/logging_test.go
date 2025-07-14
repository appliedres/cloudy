package logging

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

type testHandler struct {
	lastMsg string
	lastErr error
}

func (h *testHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testHandler) Handle(_ context.Context, r slog.Record) error {
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "error" {
			if err, ok := a.Value.Any().(error); ok {
				h.lastErr = err
			}
		}
		return true
	})
	h.lastMsg = r.Message
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}

func TestLogMsgAndErr(t *testing.T) {
	ctx := context.Background()
	handler := &testHandler{}
	logger := slog.New(handler)

	origErr := errors.New("original error")
	wrappedErr := LogMsgAndErr(ctx, logger, origErr, "first thing failed")
	doubleWrappedErr := LogMsgAndErr(ctx, logger, wrappedErr, "second thing failed")

	// Check error wrapping for single wrap
	if !strings.Contains(wrappedErr.Error(), "first thing failed") {
		t.Errorf("wrapped error missing message: %v", wrappedErr)
	}
	if !strings.Contains(wrappedErr.Error(), "original error") {
		t.Errorf("wrapped error missing original error: %v", wrappedErr)
	}

	// Optionally check log message for single wrap
	if handler.lastMsg != "first thing failed" {
		t.Errorf("log message not as expected: %v", handler.lastMsg)
	}

	// Check error wrapping for double wrap
	if !strings.Contains(doubleWrappedErr.Error(), "second thing failed") {
		t.Errorf("double wrapped error missing outer message: %v", doubleWrappedErr)
	}
	if !strings.Contains(doubleWrappedErr.Error(), "first thing failed") {
		t.Errorf("double wrapped error missing inner message: %v", doubleWrappedErr)
	}
	if !strings.Contains(doubleWrappedErr.Error(), "original error") {
		t.Errorf("double wrapped error missing original error: %v", doubleWrappedErr)
	}

	// Optionally check log message for double wrap
	if handler.lastMsg != "second thing failed" {
		t.Errorf("log message after double wrap not as expected: %v", handler.lastMsg)
	}
}
