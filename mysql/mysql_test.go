package mysql

import (
	"sync"
	"testing"

	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/log"
	"github.com/kaimixu/motor/naming"
	"github.com/stretchr/testify/require"
)

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
			Name:   "mysql/motor_test",
			Idc:    "default",
			PubEnv: "test",
			Attr: naming.InstanceAttr{
				Data: namingMysqlInstanceAttr{
					Database: map[string]mysqlClusterConf{
						"test": mysqlClusterConf{
							Master: []mysqlNodeConf{
								{
									Username: "",
									Password: "",
									IP:       "127.0.0.1",
									Port:     3306,
								},
							},
							Slave: []mysqlNodeConf{
								{
									Username: "",
									Password: "",
									IP:       "127.0.0.1",
									Port:     3306,
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

	InitMysql(ModeNaming, "motor_test", "", "test")
}

func TestFileConf(t *testing.T) {
	require.Nil(t, conf.Parse("../test/configs"))
	log.Init()

	InitMysql(ModeFile, "motor_test", "default", "test")
}
