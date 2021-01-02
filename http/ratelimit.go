package http

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/tolerant"
)

func Ratelimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if tolerant.Svc.FlowRule.Enabled {
			e, err := sentinel.Entry(tolerant.Svc.FlowRule.Resource, sentinel.WithTrafficType(base.Inbound))
			if err != nil {
				c.AbortWithStatus(http.StatusTooManyRequests)
				return
			}
			defer e.Exit()
			c.Next()
		} else {
			c.Next()
		}
	}
}
