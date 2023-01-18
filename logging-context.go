package cloudy

import (
	"bytes"
	"context"
	"fmt"
	"log"
)

type logkey string
type userkey string

var Log logkey
var UserKey userkey = userkey("userkey")
var IDKey userkey = userkey("id")

type Logs struct {
	Info   *log.Logger
	Warn   *log.Logger
	Error  *log.Logger
	Buffer *bytes.Buffer
}

func NewContext(ctx context.Context) context.Context {
	return WithLogging(ctx)
}

func StartContext() context.Context {
	return WithLogging(context.Background())
}

func WithLogging(ctx context.Context) context.Context {
	var memory bytes.Buffer
	InfoLogger := log.New(&memory, "INFO: ", log.Ldate|log.Ltime)
	WarningLogger := log.New(&memory, "WARNING: ", log.Ldate|log.Ltime)
	ErrorLogger := log.New(&memory, "ERROR: ", log.Ldate|log.Ltime)

	logs := &Logs{
		Info:   InfoLogger,
		Warn:   WarningLogger,
		Error:  ErrorLogger,
		Buffer: &memory,
	}

	Log = logkey("logkey")

	return context.WithValue(ctx, Log, logs)
}

func WithUser(ctx context.Context, user *UserJWT) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

func WithID(ctx context.Context, ID string) context.Context {
	return context.WithValue(ctx, IDKey, ID)
}

func GetLog(ctx context.Context) string {
	logs := ctx.Value(Log).(*Logs)
	buffer := logs.Buffer
	value := buffer.String()
	return value
}

func PrintLog(ctx context.Context) {
	fmt.Println(GetLog(ctx))
}

func GetUser(ctx context.Context) *UserJWT {
	item := ctx.Value(UserKey)
	if item == nil {
		return &UserJWT{
			UPN:   "Nobody",
			Email: "None",
		}
	}
	return item.(*UserJWT)
}

func GetID(ctx context.Context) string {
	ID := ctx.Value(IDKey).(string)
	return ID
}

func Info(ctx context.Context, msg string, args ...interface{}) {
	var logger *log.Logger
	if logs := getLogs(ctx); logs != nil {
		logger = logs.Info
	}
	logme(logger, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	var logger *log.Logger
	if logs := getLogs(ctx); logs != nil {
		logger = logs.Warn
	}
	logme(logger, msg, args...)
}

func Error(ctx context.Context, msg string, args ...interface{}) error {
	var logger *log.Logger
	if logs := getLogs(ctx); logs != nil {
		logger = logs.Error
	}
	logme(logger, msg, args...)
	return fmt.Errorf(msg, args...)
}

func getLogs(ctx context.Context) *Logs {
	item := ctx.Value(Log)
	if item == nil {
		return nil
	}
	return item.(*Logs)
}

func logme(logger *log.Logger, msg string, args ...interface{}) {
	log.Printf(msg, args...)

	if logger != nil {
		logger.Printf(msg+"\n", args...)
	}
}
