// Package log provides ...
// @author: xincili
// -------------------------------------------
package log

import (
	"context"
	"etrpc-go/alarm"
	"etrpc-go/config"
	"fmt"
	"os"
	"sync"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	LogTrace = "TRPC_LOG_TRACE"
)

var (
	traceEnabled = traceEnableFromEnv()
	alarmerOnce  sync.Once
	loggerOnce   sync.Once
	alarmer      alarm.Alarmer
)

// SetDefaultLogger 设置默认 logger
func SetDefaultLogger() {
	logger := log.GetDefaultLogger().With(getCommonFields()...)
	// 注册默认普通日志
	log.SetLogger(logger)
}

// traceEnableFromEnv checks whether trace is enabled by reading from environment.
// Enable trace if env is empty or zero,  disable trace if env is not zero, default as disabled.
func traceEnableFromEnv() bool {
	if e := os.Getenv(LogTrace); e == "" || e == "0" {
		return false
	}
	return true
}

// Trace logs to TRACE log. Arguments are handled in the manner of fmt.Println.
func Trace(args ...interface{}) {
	if traceEnabled {
		log.GetDefaultLogger().Trace(args...)
	}
}

// Tracef logs to TRACE log. Arguments are handled in the manner of fmt.Printf.
func Tracef(format string, args ...interface{}) {
	if traceEnabled {
		log.GetDefaultLogger().Tracef(format, args...)
	}
}

// TraceContext logs to TRACE log. Arguments are handled in the manner of fmt.Println.
func TraceContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Trace(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Trace(args...)
	}
}

// TraceContextf logs to TRACE log. Arguments are handled in the manner of fmt.Printf.
func TraceContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Tracef(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Tracef(format, args...)
	}
}

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Println.
func Debug(args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Debug(args...)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Debugf(format, args...)
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Println.
func Info(args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(format string, args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Infof(format, args...)
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func Warn(args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Warn(args...)
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func Error(args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Errorf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// All Fatal logs will exit by calling os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func Fatal(args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Fatal(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, args ...interface{}) {
	loggerOnce.Do(func() { SetDefaultLogger() })
	log.GetDefaultLogger().Fatalf(format, args...)
}

// DebugContext logs to DEBUG log. Arguments are handled in the manner of fmt.Println.
func DebugContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.Debug(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Debug(args...)
	}
}

// DebugContextf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func DebugContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Debugf(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Debugf(format, args...)
	}
}

// InfoContext logs to INFO log. Arguments are handled in the manner of fmt.Println.
func InfoContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Info(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Info(args...)
	}
}

// InfoContextf logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func InfoContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Infof(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Infof(format, args...)
	}
}

// WarnContext logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func WarnContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Warn(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Warn(args...)
	}
}

// WarnContextf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func WarnContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Warnf(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Warnf(format, args...)
	}
}

// ErrorContext logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func ErrorContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Error(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Error(args...)
	}
}

// ErrorContextf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Errorf(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Errorf(format, args...)
	}
}

// AlarmContext logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// And send alarm msg.
func AlarmContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Error(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Error(args...)
	}

	_ = getAlarmer().Alarm(ctx, fmt.Sprint(args...))
}

// AlarmContextf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// And send alarm msg.
func AlarmContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Errorf(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Errorf(format, args...)
	}

	_ = getAlarmer().Alarm(ctx, fmt.Sprintf(format, args...))
}

// FatalContext logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// All Fatal logs will exit by calling os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func FatalContext(ctx context.Context, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Fatal(args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Fatal(args...)
	}
}

// FatalContextf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func FatalContextf(ctx context.Context, format string, args ...interface{}) {
	switch l := codec.Message(ctx).Logger().(type) {
	case log.Logger:
		l.With(getCommonFields()...).Fatalf(format, args...)
	default:
		loggerOnce.Do(func() { SetDefaultLogger() })
		log.GetDefaultLogger().Fatalf(format, args...)
	}
}

func getAlarmer() alarm.Alarmer {
	alarmerOnce.Do(func() {
		alarmer = alarm.GetAlarmClient(config.GetStringOrDefault("alarm.default", "default"))
	})
	return alarmer
}

// getCommonFields 获取自定义日志字段，用于对日志字段的扩展
func getCommonFields() []log.Field {
	return []log.Field{
		{"Namespace", trpc.GlobalConfig().Global.Namespace},
		{"Env", trpc.GlobalConfig().Global.EnvName},
		{"Container", trpc.GlobalConfig().Global.ContainerName},
		{"IP", trpc.GlobalConfig().Global.LocalIP},
	}
}
