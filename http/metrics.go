package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/metrics"
)

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == metrics.DefaultPath {
			c.Next()
			return
		}
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		sstatus := strconv.Itoa(status)
		path := c.Request.URL.Path

		metrics.ReqCnt.WithLabelValues(path, sstatus).Inc()
		metrics.ReqDur.WithLabelValues(path).
			Observe(float64(time.Since(start).Milliseconds()))
		if status >= http.StatusInternalServerError {
			metrics.ReqErr.WithLabelValues(path, sstatus).Inc()
		}
	}
}
