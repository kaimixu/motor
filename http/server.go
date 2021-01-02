package http

import (
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/conf"
)

type ServerConf struct {
	Addr         string
	ReadTimeout  conf.Duration
	WriteTimeout conf.Duration
	CertFile     string
	KeyFile      string
}

type Server struct {
	*gin.Engine
	//srv http.Server
	conf *ServerConf
}

func DefaultServer(svc *ServerConf) *Server {
	server := &Server{
		conf:   svc,
		Engine: gin.New(),
	}

	server.Engine.NoRoute(func(c *gin.Context) {
		c.Data(404, "text/plain", []byte("404 page not found"))
		c.Abort()
	})
	server.Engine.NoMethod(func(c *gin.Context) {
		c.Data(405, "text/plain", []byte("Method Not Allowed"))
		c.Abort()
	})

	server.Engine.Use(Logger(), Recovery(), Trace())
	return server
}

// add middleware
func (s *Server) Use(middleware ...gin.HandlerFunc) *Server {
	s.Engine.Use(middleware...)
	return s
}

// starts listening and serving HTTP requests
func (s *Server) Run() (err error) {
	/*s.srv = http.Server{
		Addr:         s.conf.Addr,
		Handler:      s.Engine,
		ReadTimeout:  time.Duration(s.conf.ReadTimeout),
		WriteTimeout: time.Duration(s.conf.WriteTimeout),
	}
	err = s.srv.ListenAndServe()
	*/
	endless.DefaultReadTimeOut = time.Duration(s.conf.ReadTimeout)
	endless.DefaultWriteTimeOut = time.Duration(s.conf.WriteTimeout)

	err = endless.ListenAndServe(s.conf.Addr, s.Engine)
	return
}

// starts listening and serving HTTPS requests
func (s *Server) RunTLS() (err error) {
	endless.DefaultReadTimeOut = time.Duration(s.conf.ReadTimeout)
	endless.DefaultWriteTimeOut = time.Duration(s.conf.WriteTimeout)

	err = endless.ListenAndServeTLS(s.conf.Addr, s.conf.CertFile, s.conf.KeyFile, s.Engine)
	return
}

// open profiles
func (s *Server) OpenPerf() {
	OpenPerf(s.Engine)
}
