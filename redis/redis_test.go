package redis

import (
	"sync"
	"testing"

	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/log"
	"github.com/kaimixu/motor/naming"
	"github.com/stretchr/testify/require"
)

func TestFileConf(t *testing.T) {
	require.Nil(t, conf.Parse("../test/configs"))
	log.Init()

	InitRedis(ModeFile, "motor_test", "default", "test")
}

func TestNamingConf(t *testing.T) {
	require.Nil(t, conf.Parse("../test/configs"))
	log.Init()
	builder := naming.Build()
	require.NotNil(t, builder)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		ins := &naming.Instance{
			Name:   "redis/motor_test",
			Idc:    "default",
			PubEnv: "test",
			Attr: naming.InstanceAttr{
				Data: namingRedisInstanceAttr{
					Server: map[string]redisClusterConf{
						"cluster1": redisClusterConf{
							Master: []redisNodeConf{
								{
									Addr:     "127.0.0.1:6379",
									Password: "",
								},
							},
							Slave: []redisNodeConf{
								{
									Addr:     "127.0.0.1:6379",
									Password: "",
								},
							},
						},
					},
				},
			},
		}
		_, err := builder.Register(ins)
		require.Nil(t, err)
	}()
	wg.Wait()

	InitRedis(ModeNaming, "motor_test", "", "test")
}
