package trace

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

const (
	Tracer_Ctx_Key = "tracer_ctx"
)

const (
	TYPE_JAEGER = iota
	TYPE_ZIPKIN
)

var (
	_Trace ITrace
)

type ITrace interface {
	GetTraceCtx(*gin.Context) (context.Context, bool)
	GetTraceID(opentracing.Span) (string, bool)

	Close() error
}

func Init(typ int, serverName string) ITrace {
	if _Trace != nil {
		return _Trace
	}
	switch typ {
	case TYPE_JAEGER:
		_Trace = newJaeger(serverName)
		return _Trace
	case TYPE_ZIPKIN:
		// todo
		return nil
	default:
		panic("invalid type")
	}
}

func GetTraceCtx(c *gin.Context) (context.Context, bool) {
	return _Trace.GetTraceCtx(c)
}

func GetTraceID(span opentracing.Span) (string, bool) {
	return _Trace.GetTraceID(span)
}

func Close() error {
	return _Trace.Close()
}
