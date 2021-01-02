package trace

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var jLogger = &jaegerLogger{}

type jaegerLogger struct{}

func (l *jaegerLogger) Error(msg string) {
	zap.L().Error(msg)
}

func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	zap.L().Info(msg, zap.Field{Key: "jaeger_msg", Type: zapcore.ReflectType, Interface: args})
}
