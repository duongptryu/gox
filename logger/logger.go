package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync"

	pkgContext "github.com/duongptryu/gox/context"
	"github.com/duongptryu/gox/syserr"
)

type Config struct {
	Level       slog.Level
	Output      io.Writer
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

var (
	logger *slog.Logger
	once   sync.Once
)

func Init(cfg *Config) {
	once.Do(func() {
		if cfg == nil {
			cfg = &Config{
				Level:     slog.LevelInfo,
				Output:    os.Stdout,
				AddSource: false,
			}
		}

		opts := &slog.HandlerOptions{
			Level:       cfg.Level,
			AddSource:   cfg.AddSource,
			ReplaceAttr: cfg.ReplaceAttr,
		}

		handler := slog.NewJSONHandler(cfg.Output, opts)

		logger = slog.New(handler)
	})
}

type Field struct {
	key   string
	value any
}

func F(key string, value any) *Field {
	return &Field{
		key:   key,
		value: value,
	}
}

func Warning(ctx context.Context, message string, fields ...*Field) {
	logger.Warn(message, convertFields(extractContextFields(ctx, fields))...)
}

func Error(ctx context.Context, message string, fields ...*Field) {
	logger.Error(message, convertFields(extractContextFields(ctx, fields))...)
}

func Info(ctx context.Context, message string, fields ...*Field) {
	logger.Info(message, convertFields(extractContextFields(ctx, fields))...)
}

func Debug(ctx context.Context, message string, fields ...*Field) {
	logger.Debug(message, convertFields(extractContextFields(ctx, fields))...)
}

func Fatal(ctx context.Context, message string, fields ...*Field) {
	logger.Error(message, convertFields(extractContextFields(ctx, fields))...)
	os.Exit(1)
}

func LogError(ctx context.Context, err error, fields ...*Field) {
	code := syserr.GetCodeFromGenericError(err)

	fields = append(fields, convertErrorFieldsToLoggerFields(syserr.GetFieldsFromGenericError(err))...)
	fields = append(fields, F("stack", syserr.GetStackFormattedFromGenericError(err)), F("code", code))

	Error(ctx, err.Error(), fields...)
}

func extractContextFields(ctx context.Context, fields []*Field) []*Field {
	if ctx == nil {
		return fields
	}

	operationID := pkgContext.GetOperationID(ctx)
	if operationID != "" {
		fields = append(fields, F("operation_id", operationID))
	}

	requestID := pkgContext.GetRequestID(ctx)
	if requestID != "" {
		fields = append(fields, F("request_id", requestID))
	}

	userID := pkgContext.GetUserIDFromContext(ctx)
	if userID != "" {
		fields = append(fields, F("user_id", userID))
	}

	userType := pkgContext.GetUserTypeFromContext(ctx)
	if userType != "" {
		fields = append(fields, F("user_type", userType))
	}

	return fields
}

const (
	slotsPerField = 2
)

func convertFields(fields []*Field) []any {
	result := make([]any, len(fields)*slotsPerField)

	index := 0
	for _, field := range fields {
		result[index] = field.key
		result[index+1] = field.value
		index += slotsPerField
	}

	return result
}

func convertErrorFieldsToLoggerFields(fields []*syserr.Field) []*Field {
	result := make([]*Field, len(fields))

	for index, field := range fields {
		result[index] = F(field.Key, field.Value)
	}

	return result
}
