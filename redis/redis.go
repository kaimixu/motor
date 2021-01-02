package redis

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/naming"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	cacheKey = "conncache"
)

type RedisConfLoadMode = uint8

const (
	ModeNaming RedisConfLoadMode = iota
	ModeFile
)

var (
	_redisPool *redisPool
)

type namingRedisInstanceAttr = redisConf

type redisNodeConf struct {
	Addr     string `json:"addr" toml:"addr"`
	Password string `json:"password" toml:"password"`
}

type redisClusterConf struct {
	// 仅从文件中加载配置时生效
	Idc string `json:"-" toml:"idc"`

	Master []redisNodeConf `json:"master" toml:"master"`
	Slave  []redisNodeConf `json:"slave" toml:"slave"`

	MaxIdle   int `json:"conn_max_lifetime" toml:"maxIdle"`
	MaxActive int `json:"conn_max_idletime" toml:"maxActive"`
	// 单位：分钟
	IdleTimeout int `json:"idle_timeout" toml:"idleTimeout"`

	// 单位：秒
	ReadTimeout int `json:"read_timeout" toml:"readTimeout"`
	// 单位：秒
	WriteTimeout int `json:"write_timeout" toml:"writeTimeout"`
	// 单位：秒
	ConnTimeout int `json:"conn_timeout" toml:"connTimeout"`
	// 单位：分钟，0表示disable
	KeepAlive int `json:"keepalive" toml:"keepalive"`
}

type redisConf struct {
	Server map[string]redisClusterConf `json:"server" toml:"server"`
}

type redisPool struct {
	productLine string
	idc         string
	pubenv      string

	confLoadMode RedisConfLoadMode

	// 内容格式：map[cluster][]*redis.Pool
	mMap sync.Map
	sMap sync.Map
}

// confLoadMode: 配置加载方式
// productLine: 产品线，仅confLoadMode=ModeNaming有效
// idc: 机房,仅confLoadMode=ModeFile 有效，表示仅生成指定机房的连接
// pubenv: 部署环境，仅confLoadMode=ModeNaming有效
func InitRedis(confLoadMode RedisConfLoadMode, productLine, idc, pubenv string) {
	if _redisPool != nil {
		return
	}

	redi := &redisPool{
		productLine: productLine,
		idc:         idc,
		pubenv:      pubenv,

		confLoadMode: confLoadMode,
	}

	if confLoadMode == ModeFile {
		err := redi.loadConfFromFile()
		if err != nil {
			panic(err.Error())
		}
	} else {
		err := redi.loadConfFromNaming()
		if err != nil {
			panic(err.Error())
		}
	}

	_redisPool = redi
}

func Close() {

}

func (redi *redisPool) loadConfFromFile() error {
	var cfg redisConf
	if err := conf.Get("redis.toml").UnmarshalTOML(&cfg); err != nil {
		return errors.Wrap(err, "Get(redis.toml).UnmarshalTOML failed")
	}
	redi.parseFileConf(&cfg)

	go func() {
		for _ = range conf.WatchEvent("redis.toml") {
			var cfg redisConf

			if err := conf.Get("redis.toml").UnmarshalTOML(&cfg); err != nil {
				zap.L().Error("Get(redis.toml).UnmarshalTOML failed",
					zap.Error(err))
				continue
			}

			redi.parseFileConf(&cfg)
		}
	}()

	return nil
}

func (redi *redisPool) parseFileConf(cfg *redisConf) {
	mMap := make(map[string][]*redis.Pool)
	sMap := make(map[string][]*redis.Pool)

	for clusterName, cluster := range cfg.Server {
		if redi.idc != "" && redi.idc != cluster.Idc {
			continue
		}

		for _, nodeConf := range cluster.Master {
			pool := &redis.Pool{
				MaxIdle:     cluster.MaxIdle,
				MaxActive:   cluster.MaxActive,
				IdleTimeout: time.Minute * time.Duration(cluster.IdleTimeout),
				Wait:        true,
				Dial: func() (redis.Conn, error) {
					conn, err := redis.Dial("tcp", nodeConf.Addr,
						redis.DialReadTimeout(time.Second*time.Duration(cluster.ReadTimeout)),
						redis.DialWriteTimeout(time.Second*time.Duration(cluster.WriteTimeout)),
						redis.DialConnectTimeout(time.Second*time.Duration(cluster.ConnTimeout)),
						redis.DialKeepAlive(time.Minute*time.Duration(cluster.KeepAlive)),
					)
					if err != nil {
						return nil, err
					}
					if nodeConf.Password != "" {
						if _, err := conn.Do("AUTH", nodeConf.Password); err != nil {
							conn.Close()
							return nil, err
						}
					}
					return conn, err
				},
			}

			mMap[clusterName] = append(mMap[clusterName], pool)
		}

		for _, nodeConf := range cluster.Slave {
			pool := &redis.Pool{
				MaxIdle:     cluster.MaxIdle,
				MaxActive:   cluster.MaxActive,
				IdleTimeout: time.Minute * time.Duration(cluster.IdleTimeout),
				Wait:        true,
				Dial: func() (redis.Conn, error) {
					conn, err := redis.Dial("tcp", nodeConf.Addr,
						redis.DialReadTimeout(time.Second*time.Duration(cluster.ReadTimeout)),
						redis.DialWriteTimeout(time.Second*time.Duration(cluster.WriteTimeout)),
						redis.DialConnectTimeout(time.Second*time.Duration(cluster.ConnTimeout)),
						redis.DialKeepAlive(time.Minute*time.Duration(cluster.KeepAlive)),
					)
					if err != nil {
						return nil, err
					}
					if nodeConf.Password != "" {
						if _, err := conn.Do("AUTH", nodeConf.Password); err != nil {
							conn.Close()
							return nil, err
						}
					}
					return conn, err
				},
			}

			sMap[clusterName] = append(sMap[clusterName], pool)
		}
	}

	redi.mMap.Store(cacheKey, mMap)
	redi.sMap.Store(cacheKey, sMap)
}

func (redi *redisPool) loadConfFromNaming() error {
	builder := naming.Build()
	if builder == nil {
		return errors.New("naming.Build failed")
	}

	sn := "redis/" + redi.productLine
	resolver, err := builder.Discovery(sn)
	if err != nil {
		return errors.New(fmt.Sprintf("builder.Discovery failed, sn:%s", sn))
	}

	// 首次获取
	for {
		<-resolver.Watch()
		c, ok := resolver.Fetch()
		if !ok {
			continue
		}
		redi.parseNamingInstance(c)
		break
	}

	// 监听配置改动
	go func() {
		for {
			<-resolver.Watch()
			c, ok := resolver.Fetch()
			if !ok {
				continue
			}

			redi.parseNamingInstance(c)
		}
	}()

	return nil
}

func (redi *redisPool) parseNamingInstance(ins []*naming.Instance) {
	// 配置被删除
	if len(ins) == 0 {
		redi.mMap.Store(cacheKey, make(map[string][]*redis.Pool))
		redi.sMap.Store(cacheKey, make(map[string][]*redis.Pool))
		return
	}

	mMap := make(map[string][]*redis.Pool)
	sMap := make(map[string][]*redis.Pool)
	for _, in := range ins {
		if redi.idc != "" && in.Idc != redi.idc {
			continue
		}
		if redi.pubenv != "" && in.PubEnv != redi.pubenv {
			continue
		}

		var attr redisConf
		err := in.StructuredAttr(&attr)
		if err != nil {
			zap.L().Error("invalid instance",
				zap.Error(err),
				zap.Any("in", in))
			continue
		}

		for clusterName, cluster := range attr.Server {
			for _, nodeConf := range cluster.Master {
				pool := &redis.Pool{
					MaxIdle:     cluster.MaxIdle,
					MaxActive:   cluster.MaxActive,
					IdleTimeout: time.Minute * time.Duration(cluster.IdleTimeout),
					Wait:        true,
					Dial: func() (redis.Conn, error) {
						conn, err := redis.Dial("tcp", nodeConf.Addr,
							redis.DialReadTimeout(time.Second*time.Duration(cluster.ReadTimeout)),
							redis.DialWriteTimeout(time.Second*time.Duration(cluster.WriteTimeout)),
							redis.DialConnectTimeout(time.Second*time.Duration(cluster.ConnTimeout)),
							redis.DialKeepAlive(time.Minute*time.Duration(cluster.KeepAlive)),
						)
						if err != nil {
							return nil, err
						}
						if nodeConf.Password != "" {
							if _, err := conn.Do("AUTH", nodeConf.Password); err != nil {
								conn.Close()
								return nil, err
							}
						}
						return conn, err
					},
				}

				mMap[clusterName] = append(mMap[clusterName], pool)
			}

			for _, nodeConf := range cluster.Slave {
				pool := &redis.Pool{
					MaxIdle:     cluster.MaxIdle,
					MaxActive:   cluster.MaxActive,
					IdleTimeout: time.Minute * time.Duration(cluster.IdleTimeout),
					Wait:        true,
					Dial: func() (redis.Conn, error) {
						conn, err := redis.Dial("tcp", nodeConf.Addr,
							redis.DialReadTimeout(time.Second*time.Duration(cluster.ReadTimeout)),
							redis.DialWriteTimeout(time.Second*time.Duration(cluster.WriteTimeout)),
							redis.DialConnectTimeout(time.Second*time.Duration(cluster.ConnTimeout)),
							redis.DialKeepAlive(time.Minute*time.Duration(cluster.KeepAlive)),
						)
						if err != nil {
							return nil, err
						}
						if nodeConf.Password != "" {
							if _, err := conn.Do("AUTH", nodeConf.Password); err != nil {
								conn.Close()
								return nil, err
							}
						}
						return conn, err
					},
				}

				sMap[clusterName] = append(sMap[clusterName], pool)
			}
		}
	}

	redi.mMap.Store(cacheKey, mMap)
	redi.sMap.Store(cacheKey, sMap)
}

func (redi *redisPool) close() {
	val, ok := redi.mMap.Load(cacheKey)
	if ok {
		go func(val interface{}) {

			oldmMap, _ := val.(map[string][]*redis.Pool)
			for clusterName, pools := range oldmMap {
				for _, pool := range pools {
					err := pool.Close()
					if err != nil {
						zap.L().Warn("pool.Close failed",
							zap.Error(err),
							zap.String("clusterName", clusterName))
					}
				}
			}
		}(val)
	}

	val, ok = redi.sMap.Load(cacheKey)
	if ok {
		go func(val interface{}) {
			oldsMap, _ := val.(map[string][]*redis.Pool)
			for clusterName, pools := range oldsMap {
				for _, pool := range pools {
					err := pool.Close()
					if err != nil {
						zap.L().Warn("pool.Close failed",
							zap.Error(err),
							zap.String("clusterName", clusterName))
					}
				}
			}
		}(val)
	}
}

func (redi *redisPool) getConn(clusterName string, m OpMode) (redis.Conn, error) {
	if m == READ {
		val, ok := redi.sMap.Load(cacheKey)
		if !ok {
			return nil, errors.New("redis slave config uninitialized")
		}

		sMap := val.(map[string][]*redis.Pool)
		poolSlice, ok := sMap[clusterName]
		if !ok || len(poolSlice) == 0 {
			return nil, errors.New(fmt.Sprintf("redis(%s) slave config uninitialized", clusterName))
		}

		pool := poolSlice[randInt(len(poolSlice))]
		return pool.Get(), nil
	} else {
		val, ok := redi.mMap.Load(cacheKey)
		if !ok {
			return nil, errors.New("redis master config uninitialized")
		}

		mMap := val.(map[string][]*redis.Pool)
		poolSlice, ok := mMap[clusterName]
		if !ok || len(poolSlice) == 0 {
			return nil, errors.New(fmt.Sprintf("redis(%s) master config uninitialized", clusterName))
		}

		pool := poolSlice[randInt(len(poolSlice))]
		return pool.Get(), nil
	}
}

func randInt(max int) int {
	rand.Seed(time.Now().Unix())

	return rand.Intn(max)
}
