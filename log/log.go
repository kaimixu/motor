package log

import (
	"fmt"
	"strings"

	"github.com/kaimixu/motor/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Initialized = false

type LogConf struct {
	Level          string                 `toml:level`
	Encoding       string                 `toml:encoding`
	OutputPaths    []string               `toml:outputPaths`
	ErrOutputPaths []string               `toml:errOutputPaths`
	InitialFields  map[string]interface{} `toml:initialFields`
}

// create zap log object
func Init() {
	cfg := getConf()

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(getLevel(cfg.Level)),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         cfg.Encoding,
		EncoderConfig:    encoderCfg,
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrOutputPaths,
		InitialFields:    cfg.InitialFields,
	}

	logger, err := zapCfg.Build(zap.AddCaller())
	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(logger)
	Initialized = true
}

func getConf() *LogConf {
	var st conf.Storage
	var cfg LogConf
	if err := conf.Get("application.toml").Unmarshal(&st); err != nil {
		panic(err)
	}
	if err := st.Get("Log").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func getLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info", "": // make the zero value useful
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		panic(fmt.Sprintf("invalid log level, level:%s", level))
	}
}
