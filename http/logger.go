package http

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		end := time.Now()
		local, _ := time.LoadLocation("Local")
		referer := c.Request.Header.Get("Referer")
		if referer == "" {
			referer = "-"
		}
		accessLog := fmt.Sprintf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\"",
			c.ClientIP(),
			end.In(local).Format("02/Jan/2006:15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.Proto,
			c.Writer.Status(),
			c.Writer.Size(),
			referer,
			c.Request.Header.Get("User-Agent"),
		)

		zap.L().Info(accessLog,
			zap.Duration("latency", end.Sub(start)),
			zap.String("path", path),
			zap.String("errmsg", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
