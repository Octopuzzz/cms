// Package logger provides structured logging with correlation IDs,
// request tracing, and support for microservices integration.
package logger

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// contextKey is a type for context keys
type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	RequestIDKey     contextKey = "request_id"
	TraceIDKey       contextKey = "trace_id"
	SpanIDKey        contextKey = "span_id"
	UserIDKey        contextKey = "user_id"
)

// Logger is the application logger interface
type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	WithContext(ctx context.Context) Logger
	WithFields(fields map[string]any) Logger
	GetWriter() io.Writer
}

type logrusLogger struct {
	logger *logrus.Logger
	fields logrus.Fields
	mu     sync.RWMutex
}

var (
	instance Logger
	once     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger() Logger {
	once.Do(func() {
		instance = newLogger()
	})
	return instance
}

func newLogger() Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)
	return &logrusLogger{logger: log, fields: make(logrus.Fields)}
}

// InitLogger initializes the logger with configuration
func InitLogger(level string, format string) {
	log := logrus.New()
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)

	switch format {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	default:
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	}
	log.SetOutput(os.Stdout)
	instance = &logrusLogger{logger: log, fields: make(logrus.Fields)}
}

func (l *logrusLogger) Debug(msg string, fields ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry := l.logger.WithFields(l.fields)
	if len(fields) > 0 {
		entry = entry.WithFields(kvToFields(fields...))
	}
	entry.Debug(msg)
}

func (l *logrusLogger) Info(msg string, fields ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry := l.logger.WithFields(l.fields)
	if len(fields) > 0 {
		entry = entry.WithFields(kvToFields(fields...))
	}
	entry.Info(msg)
}

func (l *logrusLogger) Warn(msg string, fields ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry := l.logger.WithFields(l.fields)
	if len(fields) > 0 {
		entry = entry.WithFields(kvToFields(fields...))
	}
	entry.Warn(msg)
}

func (l *logrusLogger) Error(msg string, fields ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry := l.logger.WithFields(l.fields)
	if len(fields) > 0 {
		entry = entry.WithFields(kvToFields(fields...))
	}
	entry.Error(msg)
}

func (l *logrusLogger) Fatal(msg string, fields ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	entry := l.logger.WithFields(l.fields)
	if len(fields) > 0 {
		entry = entry.WithFields(kvToFields(fields...))
	}
	entry.Fatal(msg)
}

func (l *logrusLogger) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}
	newFields := make(logrus.Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	if v := GetCorrelationID(ctx); v != "" {
		newFields["correlation_id"] = v
	}
	if v := GetRequestID(ctx); v != "" {
		newFields["request_id"] = v
	}
	if v := GetTraceID(ctx); v != "" {
		newFields["trace_id"] = v
	}
	if v := GetUserID(ctx); v != "" {
		newFields["user_id"] = v
	}
	return &logrusLogger{logger: l.logger, fields: newFields}
}

func (l *logrusLogger) WithFields(fields map[string]any) Logger {
	newFields := make(logrus.Fields)
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &logrusLogger{logger: l.logger, fields: newFields}
}

func (l *logrusLogger) GetWriter() io.Writer {
	return l.logger.Out
}

func kvToFields(kv ...any) logrus.Fields {
	fields := make(logrus.Fields)
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			fields[k] = kv[i+1]
		}
	}
	return fields
}

// ---- Context helpers ----

func WithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, CorrelationIDKey, id)
}

func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(CorrelationIDKey).(string)
	return v
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, RequestIDKey, id)
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(RequestIDKey).(string)
	return v
}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, TraceIDKey, id)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(TraceIDKey).(string)
	return v
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, UserIDKey, id)
}

func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

func GenerateCorrelationID() string { return uuid.New().String() }
func GenerateRequestID() string     { return uuid.New().String() }

// GinMiddleware creates a Gin middleware for logging with correlation IDs
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = GenerateCorrelationID()
		}
		requestID := GenerateRequestID()

		ctx := WithCorrelationID(c.Request.Context(), correlationID)
		ctx = WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Header("X-Correlation-ID", correlationID)
		c.Header("X-Request-ID", requestID)

		c.Next()

		duration := time.Since(start)
		log := GetLogger().WithContext(ctx)
		fields := map[string]any{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"status":    c.Writer.Status(),
			"duration":  duration.Milliseconds(),
			"client_ip": c.ClientIP(),
			"size":      c.Writer.Size(),
		}
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
			log.Error("Request failed", toKV(fields)...)
		} else {
			log.Info("Request completed", toKV(fields)...)
		}
	}
}

func toKV(m map[string]any) []any {
	out := make([]any, 0, len(m)*2)
	for k, v := range m {
		out = append(out, k, v)
	}
	return out
}
