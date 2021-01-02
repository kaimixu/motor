package tolerant

import (
	"fmt"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/kaimixu/motor/conf"
)

var (
	Svc SentinelConf
)

type serverConf struct {
	CollectIntervalMs uint32
	AppName           string
}

type logConf struct {
	UsePid            bool
	MaxFileCount      uint32
	FlushInterval     conf.Duration
	SingleFileMaxSize conf.ByteSize
	LogDir            string
}

type flowRuleConf struct {
	Enabled         bool
	Count           uint64
	MetricType      MetricType
	ControlBehavior ControlBehavior
	Resource        string
}

type breakerRuleConf struct {
	Enabled          bool
	MinRequestAmount uint64
	Threshold        float64
	StatInterval     conf.Duration
	MaxAllowedRt     conf.Duration
	RetryTimeout     conf.Duration
	Resource         string
	Strategy         Strategy
}

type SentinelConf struct {
	Server      serverConf
	Log         logConf
	FlowRule    flowRuleConf
	BreakerRule breakerRuleConf
}

func getConf() SentinelConf {
	var cfg SentinelConf
	if err := conf.Get("sentinel.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

func Init() {
	Svc = getConf()

	entity := config.NewDefaultConfig()
	entity.Sentinel.App.Name = Svc.Server.AppName
	entity.Sentinel.Log.Dir = Svc.Log.LogDir
	entity.Sentinel.Log.UsePid = Svc.Log.UsePid
	entity.Sentinel.Log.Metric.SingleFileMaxSize = uint64(Svc.Log.SingleFileMaxSize)
	entity.Sentinel.Log.Metric.MaxFileCount = Svc.Log.MaxFileCount
	entity.Sentinel.Log.Metric.FlushIntervalSec = uint32(time.Duration(Svc.Log.FlushInterval).Seconds())
	entity.Sentinel.Stat.System.CollectIntervalMs = Svc.Server.CollectIntervalMs

	if err := sentinel.InitWithConfig(entity); err != nil {
		panic(err)
	}

	if Svc.FlowRule.Enabled {
		_, err := flow.LoadRules([]*flow.Rule{
			{
				Resource:               Svc.FlowRule.Resource,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.ControlBehavior(Svc.FlowRule.ControlBehavior),
				MetricType:             flow.MetricType(Svc.FlowRule.MetricType),
				Count:                  float64(Svc.FlowRule.Count),
			},
		})
		if err != nil {
			panic(fmt.Sprintf("Unexpected error: %+v", err))
		}
	}

	if Svc.BreakerRule.Enabled {
		_, err := circuitbreaker.LoadRules([]*circuitbreaker.Rule{
			{
				Resource:         Svc.BreakerRule.Resource,
				Strategy:         circuitbreaker.Strategy(Svc.BreakerRule.Strategy),
				RetryTimeoutMs:   uint32(time.Duration(Svc.BreakerRule.RetryTimeout).Milliseconds()),
				MinRequestAmount: Svc.BreakerRule.MinRequestAmount,
				StatIntervalMs:   uint32(time.Duration(Svc.BreakerRule.StatInterval).Milliseconds()),
				Threshold:        Svc.BreakerRule.Threshold,
				MaxAllowedRtMs:   uint64(time.Duration(Svc.BreakerRule.MaxAllowedRt).Milliseconds()),
			},
		})
		if err != nil {
			panic(fmt.Sprintf("Unexpected error: %+v", err))
		}
	}
}
