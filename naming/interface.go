package naming

import (
	"context"
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type InstancesInfo struct {
	Instances map[string][]*Instance
}

type Instance struct {
	// 服务名
	Name string `json:"name"`
	// 机房
	Idc string `json:"idc"`
	// 沙盒、小流量、全流量
	PubEnv string `json:"pubenv"`
	// 服务属性
	Attr InstanceAttr `json:"attr"`
}

type InstanceAttr struct {
	Data interface{} `json:"data"`
}

func (ia *InstanceAttr) UnmarshalJSON(b []byte) error {
	ia.Data = string(b)
	return nil
}

// 默认服务属性
type DefaultInstanceAttr struct {
	Addrs          []string `json:"addrs"`
	ConnectRetry   uint32   `json:"connect_retry"`
	ConnectTimeout uint32   `json:"connect_timeout"`
	ReadTimeout    uint32   `json:"read_timeout"`
	WriteTimeout   uint32   `json:"write_timeout"`
	AuthEnabled    bool     `json:"auth_enabled"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
}

func (in *Instance) StructuredAttr(s interface{}) error {
	attrStr, ok := in.Attr.Data.(string)
	if !ok {
		return errors.New("attr.data field not a string")
	}
	attrStr = jsoniter.Get([]byte(attrStr), "data").ToString()
	return json.Unmarshal([]byte(attrStr), s)
}

type Builder interface {
	Discovery(sn string) (Resolver, error)
	Register(ins *Instance) (cancelFunc context.CancelFunc, err error)
	// 关闭所有服务注册和发现
	Close()
}

// 服务发现
type Resolver interface {
	// 监听服务改变
	Watch() <-chan struct{}
	// 获取服务配置
	Fetch() (ins []*Instance, ok bool)
	// 关闭服务监听
	Close()
}

// 服务注册
type Registry interface {
	Register()
	Close()
}
