package log

import (
	"context"
	"fmt"
	"runtime"

	"github.com/google/uuid"
)

const (
	ErrorType = 0
	InfoType  = 1
	DebugType = 2
)

type Err struct {
	Type        int32
	ServiceName string
	SpanID      uuid.UUID
	Source      string
	Message     string
}

const (
	SpanKeyName = "clever_span_id"
)

var (
	c           *Client
	serviceName string
)

type Options struct {
	ServiceName string
	Host        string
	Port        string
}

func init() {
	c = NewClient()
}

func SetServiceName(name string) {
	serviceName = name
}

func Error(ctx context.Context, message string) {
	_, file, line, _ := runtime.Caller(1)

	span, ok := ctx.Value(SpanKeyName).(uuid.UUID)
	if !ok {
		span = uuid.New()
	}

	c.Send(&Err{
		Type:        ErrorType,
		ServiceName: serviceName,
		SpanID:      span,
		Source:      fmt.Sprintf("file: %s, line: %s", file, line),
		Message:     message,
	})
}

func Info(ctx context.Context, message string) {
	_, file, line, _ := runtime.Caller(0)

	span, ok := ctx.Value(SpanKeyName).(uuid.UUID)
	if !ok {
		span = uuid.New()
	}

	c.Send(&Err{
		Type:        InfoType,
		ServiceName: serviceName,
		SpanID:      span,
		Source:      fmt.Sprintf("file: %s, line: %s", file, line),
		Message:     message,
	})
}

func Debug(ctx context.Context, message string) {
	_, file, line, _ := runtime.Caller(0)

	span, ok := ctx.Value(SpanKeyName).(uuid.UUID)
	if !ok {
		span = uuid.New()
	}

	c.Send(&Err{
		Type:        DebugType,
		ServiceName: serviceName,
		SpanID:      span,
		Source:      fmt.Sprintf("file: %s, line: %s", file, line),
		Message:     message,
	})
}
