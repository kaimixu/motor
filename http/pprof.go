package http

import (
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	perfOnce sync.Once
)

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func OpenPerf(engine *gin.Engine) {
	perfOnce.Do(func() {
		prefixRouter := engine.Group("/debug/pprof")
		{
			prefixRouter.GET("/", pprofHandler(pprof.Index))
			prefixRouter.GET("/cmdline", pprofHandler(pprof.Cmdline))
			prefixRouter.GET("/profile", pprofHandler(pprof.Profile))
			prefixRouter.GET("/symbol", pprofHandler(pprof.Symbol))
			prefixRouter.GET("/trace", pprofHandler(pprof.Trace))
			prefixRouter.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
			prefixRouter.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
			prefixRouter.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
			prefixRouter.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
			prefixRouter.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
			prefixRouter.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
		}
		return
	})
}
