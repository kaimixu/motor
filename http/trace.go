package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/trace"
	"github.com/micro/go-micro/metadata"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"
)

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		if opentracing.IsGlobalTracerRegistered() {
			tracer := opentracing.GlobalTracer()
			spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
			var span opentracing.Span
			if err != nil {
				span = tracer.StartSpan(c.Request.URL.Path)
			} else {
				span = tracer.StartSpan(c.Request.URL.Path, opentracing.ChildOf(spanCtx))
			}
			defer span.Finish()

			carrier := opentracing.TextMapCarrier{}
			// 将请求的Baggage数据inject到carrier中
			err = tracer.Inject(span.Context(), opentracing.TextMap, carrier)
			if err != nil {
				zap.L().Error("tracer.Inject failed",
					zap.String("clientIp", c.ClientIP()),
					zap.String("url", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("errmsg", fmt.Sprintf("%v", err)),
				)
			}

			ctx := opentracing.ContextWithSpan(context.Background(), span)
			ctx = metadata.NewContext(ctx, metadata.Metadata(carrier))
			// 存放到gin框架的context中，方便后面的中间件及路由handler获取
			c.Set(trace.Tracer_Ctx_Key, ctx)

			c.Next()

			statusCode := c.Writer.Status()
			ext.HTTPStatusCode.Set(span, uint16(statusCode))
			ext.HTTPMethod.Set(span, c.Request.Method)
			ext.HTTPUrl.Set(span, c.Request.URL.EscapedPath())
			if statusCode >= http.StatusInternalServerError {
				ext.Error.Set(span, true)
			}
		} else {
			c.Next()
		}
	}
}
