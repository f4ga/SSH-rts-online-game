// Package logger provides a structured, leveled logger for the SSH Arena game.
// It wraps zerolog with a simple interface and configurable output.
package logger

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger is the interface for logging methods.
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// zeroLogger wraps zerolog.Logger to implement Logger.
type zeroLogger struct {
	log zerolog.Logger
}

var (
	globalLogger Logger
	once         sync.Once
	mu           sync.RWMutex
)

// Init initializes the global logger with the given configuration.
// It must be called before using Get.
func Init(level, format, output string, withCaller bool) error {
	var writers []io.Writer

	switch output {
	case "stdout":
		writers = append(writers, os.Stdout)
	case "stderr":
		writers = append(writers, os.Stderr)
	case "file":
		f, err := os.OpenFile("game.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, f)
	default:
		writers = append(writers, os.Stdout)
	}

	multi := zerolog.MultiLevelWriter(writers...)

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var zlog zerolog.Logger
	if format == "console" {
		zlog = zerolog.New(zerolog.ConsoleWriter{Out: multi, TimeFormat: "15:04:05"}).With().Timestamp().Logger()
	} else {
		zlog = zerolog.New(multi).With().Timestamp().Logger()
	}

	// Set log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zlog = zlog.Level(lvl)

	if withCaller {
		zlog = zlog.With().Caller().Logger()
	}

	globalLogger = &zeroLogger{log: zlog}
	return nil
}

// Get returns the global logger instance.
// If Init hasn't been called, a default logger (info level, stdout) is created.
func Get() Logger {
	once.Do(func() {
		if globalLogger == nil {
			// fallback default
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			zlog := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(zerolog.InfoLevel)
			globalLogger = &zeroLogger{log: zlog}
		}
	})
	mu.RLock()
	defer mu.RUnlock()
	return globalLogger
}

// Debug logs a debug message.
func (z *zeroLogger) Debug(msg string, fields ...interface{}) {
	z.log.Debug().Fields(fieldsToMap(fields)).Msg(msg)
}

// Info logs an info message.
func (z *zeroLogger) Info(msg string, fields ...interface{}) {
	z.log.Info().Fields(fieldsToMap(fields)).Msg(msg)
}

// Warn logs a warning message.
func (z *zeroLogger) Warn(msg string, fields ...interface{}) {
	z.log.Warn().Fields(fieldsToMap(fields)).Msg(msg)
}

// Error logs an error message.
func (z *zeroLogger) Error(msg string, fields ...interface{}) {
	z.log.Error().Fields(fieldsToMap(fields)).Msg(msg)
}

// Fatal logs a fatal message and exits.
func (z *zeroLogger) Fatal(msg string, fields ...interface{}) {
	z.log.Fatal().Fields(fieldsToMap(fields)).Msg(msg)
}

// With returns a new logger with the given fields attached.
func (z *zeroLogger) With(fields ...interface{}) Logger {
	return &zeroLogger{log: z.log.With().Fields(fieldsToMap(fields)).Logger()}
}

// fieldsToMap converts a variadic list of key‑value pairs to a map.
// It expects an even number of arguments, alternating between string keys and values.
func fieldsToMap(fields []interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	m := make(map[string]interface{}, len(fields)/2)
	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		m[key] = fields[i+1]
	}
	return m
}