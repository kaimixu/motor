package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/jwt"
)

func Jwt(secret string) gin.HandlerFunc {
	j := jwt.NewJWT(secret)
	return func(c *gin.Context) {
		tokenStr := c.Request.Header.Get("JWT-TOKEN")
		if tokenStr == "" {
			c.JSON(http.StatusOK, gin.H{
				"status": -1,
				"errmsg": "请求未携带jwt-token",
				"data":   nil,
			})

			c.Abort()
			return
		}

		claims, err := j.ParseToken(tokenStr)
		if err != nil {
			if j.IsExpires(err) {
				c.JSON(http.StatusOK, gin.H{
					"status": 1,
					"errmsg": "token已过期",
					"data":   nil,
				})
				c.Abort()
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status": 1,
				"errmsg": "token无效",
				"data":   nil,
			})
			c.Abort()
			return
		}

		c.Set("jwtClaims", claims)
	}
}
