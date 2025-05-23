package logging

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type contextLoggerType string

var ContextLoggerKey = contextLoggerType("arkloud-context-logger")
var ContextTracingKey = contextLoggerType("arkloud-context-tracing")

func GetLogger(ctx context.Context) *slog.Logger {
	val := ctx.Value(ContextLoggerKey)
	logger, isLogger := val.(*slog.Logger)
	if isLogger {
		return logger
	}
	return slog.Default()
}

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	return logger
}

func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ContextLoggerKey, logger)
}

func WithError(err error) slog.Attr {
	var stack []string

	for err != nil {
		stack = append(stack, err.Error())
		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			break
		}
		err = unwrappedErr
	}

	return slog.String("error", fmt.Sprintf("%s", stack))
}

func AddRequest(ctx context.Context, logger *slog.Logger, req *http.Request) {

}

func NewContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Set up Tracking
	tracingId := uuid.NewString()
	tracingCtx := context.WithValue(ctx, ContextTracingKey, tracingId)

	// Create Logger
	log := NewLogger()
	log = log.With("tracing", tracingId)
	ctxWithLogging := CtxWithLogger(tracingCtx, log)

	return ctxWithLogging
}

// logAndWrap logs the 'msg' plus the original err, then returns errors.Wrap(err, msg).
func LogAndWrapErr(ctx context.Context, logger *slog.Logger, err error, msg string) error {
	logger.ErrorContext(ctx, msg, WithError(err))
	return errors.Wrap(err, msg)
}
