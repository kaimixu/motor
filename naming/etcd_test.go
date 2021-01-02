package naming

import (
	"sync"
	"testing"
	"time"

	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/log"
	"github.com/stretchr/testify/require"
)

func TestDiscoveryAndRegistry(t *testing.T) {
	require.Nil(t, conf.Parse("../test/configs"))
	log.Init()
	builder := Build()
	require.NotNil(t, builder)

	go func() {
		resolver, _ := builder.Discovery("motor/naming")
		for {
			<-resolver.Watch()
			_, ok := resolver.Fetch()
			require.Equal(t, ok, true)
		}
	}()

	go func() {
		resolver, _ := builder.Discovery("motor/naming")
		<-resolver.Watch()
		_, ok := resolver.Fetch()
		require.Equal(t, ok, true)
		resolver.Close()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		ins := &Instance{
			Name:   "motor",
			Idc:    "default",
			PubEnv: "test",
			Attr: InstanceAttr{
				Data: &DefaultInstanceAttr{
					Addrs:        []string{"127.0.0.1:6998", "127.0.0.1:6999"},
					ReadTimeout:  2,
					WriteTimeout: 2,
				},
			},
		}
		cancelFunc, err := builder.Register(ins)
		require.Nil(t, err)

		time.Sleep(time.Second * 5)
		cancelFunc()
		time.Sleep(time.Second * 2) // 等待处理delete 事件
	}()

	wg.Wait()
	builder.Close()
}
