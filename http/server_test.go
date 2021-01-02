package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/gin-gonic/gin"
	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/metrics"
	"github.com/kaimixu/motor/tolerant"
	"github.com/kaimixu/motor/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HttpTestSuite struct {
	suite.Suite
	addr string
}

func (suite *HttpTestSuite) SetupSuite() {
	to, _ := time.ParseDuration("3s")
	suite.addr = "127.0.0.1:18082"
	svc := &ServerConf{
		Addr:         suite.addr,
		ReadTimeout:  conf.Duration(to),
		WriteTimeout: conf.Duration(to),
	}

	srv := DefaultServer(svc)
	srv.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// TestTrace
	srv.GET("/trace", func(c *gin.Context) {
		ctx, exists := trace.GetTraceCtx(c)
		require.True(suite.T(), exists)
		firstF(ctx)
		c.String(200, "trace done")
	})

	// TestMetrics
	srv.GET(metrics.DefaultPath, metrics.MetricsHandler())
	metricsR := srv.Group("/metrics/", Metrics())
	metricsR.GET("*action", func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics/1" {
			c.String(500, "%s", "server error")
		} else {
			c.String(200, "%s", "done")
		}
	})

	// TestRatelimit
	ratelimitR := srv.Group("/ratelimit/", Ratelimit())
	ratelimitR.GET("*action", func(c *gin.Context) {
		c.String(200, "%s", "done")
	})

	go srv.Run()
	// 等待服务启动完成
	time.Sleep(time.Second)
}

func (suite *HttpTestSuite) TestHttp() {
	require := require.New(suite.T())

	resp, err := http.Get(fmt.Sprintf("http://%s%s", suite.addr, "/ping"))
	require.NoError(err)
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(err)
	require.Equal(string(bytes), "pong")
}

// test trace middleware
func (suite *HttpTestSuite) TestTrace() {
	require.Nil(suite.T(), conf.Parse("../test/configs"))
	trace.Init(trace.TYPE_JAEGER, "testJaeger")
	defer require.Nil(suite.T(), trace.Close())

	resp, err := http.Get(fmt.Sprintf("http://%s/trace", suite.addr))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
}

// test metrics middleware
func (suite *HttpTestSuite) TestMetrics() {
	metrics.Init("testMetrics")

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func(urls []string) {
		defer wg2.Done()

		for _, url := range urls {
			resp, err := http.Get(fmt.Sprintf("http://%s%s", suite.addr, url))
			require.NoError(suite.T(), err)
			resp.Body.Close()
		}
	}([]string{"/metrics/1", "/metrics/2", "/metrics/3", "/metrics/4"})
	wg2.Wait()

	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", suite.addr))
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(suite.T(), err)
	suite.T().Logf("metrics response:\n%s", string(bytes))
}

// test ratelimit middleware
func (suite *HttpTestSuite) TestRatelimit() {
	require.Nil(suite.T(), conf.Parse("../test/configs"))
	tolerant.Init()

	if tolerant.Svc.FlowRule.Enabled &&
		tolerant.Svc.FlowRule.MetricType == tolerant.MetricType(flow.QPS) &&
		tolerant.Svc.FlowRule.ControlBehavior == tolerant.ControlBehavior(flow.Reject) {
		var wg2 sync.WaitGroup
		count := int(tolerant.Svc.FlowRule.Count)
		wg2.Add(count)
		// success request
		for i := 1; i <= count; i++ {
			go func(urls []string) {
				defer wg2.Done()

				for _, url := range urls {
					resp, err := http.Get(fmt.Sprintf("http://%s%s", suite.addr, url))
					require.NoError(suite.T(), err, fmt.Sprintf("%v", err))
					require.Equal(suite.T(), resp.StatusCode, http.StatusOK)
					resp.Body.Close()
				}
			}([]string{"/ratelimit/1"})
		}
		wg2.Wait()

		// limited request
		resp, err := http.Get(fmt.Sprintf("http://%s/ratelimit/%d", suite.addr, count+1))
		require.NoError(suite.T(), err, fmt.Sprintf("%v", err))
		require.Equal(suite.T(), resp.StatusCode, http.StatusTooManyRequests)
		resp.Body.Close()

		// wait log flush
		dur := time.Duration(tolerant.Svc.Log.FlushInterval).Seconds()
		if dur > 10 {
			dur = 10
		}
		time.Sleep(time.Duration(dur) * time.Second)
	}
}

func (suite *HttpTestSuite) TearDownSuite() {
	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	require.NoError(suite.T(), err)
}

func TestHttp(t *testing.T) {
	suite.Run(t, new(HttpTestSuite))
}

func firstF(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "firstF")
	defer span.Finish()

	secondF(ctx)
}

func secondF(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "secondF")
	defer span.Finish()
}
