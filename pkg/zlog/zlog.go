package zlog

import (
	"io"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Level  = zapcore.Level
	Option = zap.Option
	Field  = zap.Field
)

type ZLogger struct {
	zap   *zap.Logger
	level zap.AtomicLevel
}

var (
	Named = std.zap.Named

	Info   = std.zap.Info
	Warn   = std.zap.Warn
	Error  = std.zap.Error
	DPanic = std.zap.DPanic
	Panic  = std.zap.Panic
	Fatal  = std.zap.Fatal
	Debug  = std.zap.Debug
	Sugar  = std.zap.Sugar
	Debugf = std.zap.Sugar().Debugf
	Infof  = std.zap.Sugar().Infof
	Warnf  = std.zap.Sugar().Warnf
	Errorf = std.zap.Sugar().Errorf
	Fatalf = std.zap.Sugar().Fatalf
)

var std = New(os.Stdout, InfoLevel, WithCaller(true), AddCallerSkip(0))

func New(writer io.Writer, level Level, opts ...Option) *ZLogger {
	if writer == nil {
		panic("the writer is nil")
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	cfg.EncoderConfig.EncodeCaller = func(ec zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(ec.TrimmedPath())
	}
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(level)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg.EncoderConfig),
		zapcore.AddSync(writer),
		atomicLevel,
	)
	logger := &ZLogger{
		zap:   zap.New(core, opts...),
		level: atomicLevel,
	}
	return logger
}

type RotateOptions struct {
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

type LevelEnablerFunc func(lvl Level) bool

type TeeOption struct {
	Filename string
	Ropt     RotateOptions
	Lef      LevelEnablerFunc
}

func NewTeeWithRotate(tops []TeeOption, opts ...Option) *ZLogger {
	cores := make([]zapcore.Core, 0, len(tops))
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	cfg.EncoderConfig.EncodeCaller = func(ec zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(ec.TrimmedPath())
	}

	for _, top := range tops {
		top := top

		lv := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return top.Lef(lvl)
		})

		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   top.Filename,
			MaxSize:    top.Ropt.MaxSize,
			MaxBackups: top.Ropt.MaxBackups,
			MaxAge:     top.Ropt.MaxAge,
			Compress:   top.Ropt.Compress,
		})

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zapcore.AddSync(w),
			lv,
		)
		cores = append(cores, core)
	}

	logger := &ZLogger{
		zap: zap.New(zapcore.NewTee(cores...), opts...),
	}
	return logger
}

// not safe for concurrent use
func ResetDefault(l *ZLogger) {
	std = l

	Info = std.zap.Info
	Warn = std.zap.Warn
	Error = std.zap.Error
	DPanic = std.zap.DPanic
	Panic = std.zap.Panic
	Fatal = std.zap.Fatal
	Debug = std.zap.Debug
	Sugar = std.zap.Sugar
	Debugf = std.zap.Sugar().Debugf
	Infof = std.zap.Sugar().Infof
	Warnf = std.zap.Sugar().Warnf
	Errorf = std.zap.Sugar().Errorf
	Fatalf = std.zap.Sugar().Fatalf
}

func (z *ZLogger) Sync() error {
	return z.zap.Sync()
}

func Sync() error {
	if std != nil {
		return std.Sync()
	}
	return nil
}

func GetZapLogger() *zap.Logger {
	return std.zap
}

func ChangeLogLevel(l string) {
	if std == nil {
		return
	}
	switch l {
	case "debug":
		std.level.SetLevel(zap.DebugLevel)
	case "info":
		std.level.SetLevel(zap.InfoLevel)
	case "warn":
		std.level.SetLevel(zap.WarnLevel)
	case "error":
		std.level.SetLevel(zap.ErrorLevel)
	default:
		std.level.SetLevel(zap.InfoLevel)
	}
}
