package trace

import (
	"context"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/conf"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

var _ ITrace = &TJaeger{}

type TJaeger struct {
	serviceName string
	closer      io.Closer
}

func newJaeger(sn string) *TJaeger {
	j := &TJaeger{serviceName: sn}
	j.init()

	return j
}

func (t *TJaeger) init() {
	var st conf.Storage
	var sampler jaegercfg.SamplerConfig
	var reporter jaegercfg.ReporterConfig
	if err := conf.Get("jaeger.toml").Unmarshal(&st); err != nil {
		panic(err)
	}
	if err := st.Get("Sampler").UnmarshalTOML(&sampler); err != nil {
		panic(err)
	}
	if err := st.Get("Reporter").UnmarshalTOML(&reporter); err != nil {
		panic(err)
	}

	cfg := jaegercfg.Configuration{
		Sampler:  &sampler,
		Reporter: &reporter,
	}
	closer, err := cfg.InitGlobalTracer(t.serviceName, jaegercfg.Logger(jLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	t.closer = closer
}

func (t *TJaeger) GetTraceCtx(c *gin.Context) (context.Context, bool) {
	ctx, exists := c.Get(Tracer_Ctx_Key)
	if !exists {
		return nil, false
	}
	ctx2, ok := ctx.(context.Context)
	if !ok {
		return nil, false
	}

	return ctx2, true
}

func (t *TJaeger) GetTraceID(span opentracing.Span) (id string, ok bool) {
	if sc, ok := span.Context().(jaeger.SpanContext); ok {
		return sc.TraceID().String(), true
	}

	return "", false
}

func (t *TJaeger) Close() error {
	return t.closer.Close()
}
