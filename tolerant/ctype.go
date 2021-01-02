package tolerant

import (
	"errors"
	"strings"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
)

type MetricType flow.MetricType

// satisfies the encoding.TextUnmarshaler interface
func (d *MetricType) UnmarshalText(text []byte) error {
	str := strings.ToLower(string(text))
	if str == "qps" {
		*d = MetricType(flow.QPS)
	} else if str == "concurrency" {
		*d = MetricType(flow.Concurrency)
	} else {
		return errors.New("unknown metricType")
	}

	return nil
}

type ControlBehavior flow.ControlBehavior

// satisfies the encoding.TextUnmarshaler interface
func (d *ControlBehavior) UnmarshalText(text []byte) error {
	str := strings.ToLower(string(text))
	if str == "reject" {
		*d = ControlBehavior(flow.Reject)
	} else if str == "throttling" {
		*d = ControlBehavior(flow.Throttling)
	} else {
		return errors.New("unknown controlBehavior")
	}

	return nil
}

type Strategy circuitbreaker.Strategy

// satisfies the encoding.TextUnmarshaler interface
func (d *Strategy) UnmarshalText(text []byte) error {
	str := strings.ToLower(string(text))
	if str == "slowrequestratio" {
		*d = Strategy(circuitbreaker.SlowRequestRatio)
	} else if str == "errorratio" {
		*d = Strategy(circuitbreaker.ErrorRatio)
	} else if str == "errorcount" {
		*d = Strategy(circuitbreaker.ErrorCount)
	} else {
		return errors.New("unknown Strategy")
	}

	return nil
}
