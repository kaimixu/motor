package conf

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type redisConf struct {
	Name           string
	Addr           string
	MaxIdle        int
	MaxActive      int
	MaxWaitTimeout Duration
	DialTimeout    Duration
	ReadTimeout    Duration
	WriteTimeout   Duration
	IdleTimeout    Duration
}

func TestTomlDecode(t *testing.T) {
	require := require.New(t)
	cpath = "/tmp/test_conf/"
	require.Nil(os.MkdirAll(cpath, 0666))

	data := `
[Server]
name = "test"
addr = "127.0.0.1:6379"
maxIdle = 50
maxActive = 100
maxWaitTimeout = "100ms"
dialTimeout = "2s"
readTimeout = "1s"
writeTimeout = "1s"
idleTimeout = "10s"
`
	filename := path.Join(cpath, "redis.toml")
	require.Nil(ioutil.WriteFile(filename, []byte(data), 0644), "write file failed")
	require.Nil(Parse(""))

	var rcfg redisConf
	var st Storage
	require.Nil(Get("redis.toml").Unmarshal(&st))
	require.Nil(st.Get("Server").UnmarshalTOML(&rcfg))

	require.Equal(rcfg.Name, "test")
	require.Equal(rcfg.MaxIdle, 50)
	require.Equal(rcfg.MaxWaitTimeout, Duration(100*time.Millisecond))
	require.Equal(rcfg.DialTimeout, Duration(2*time.Second))
}
