package conf

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewConf(t *testing.T) {
	require := require.New(t)
	cpath = "../test/tmp/"
	require.Nil(os.MkdirAll(cpath, 0666))

	data := `
[server]
ip=127.0.0.1
`
	filename := path.Join(cpath, "http.toml")
	require.Nil(ioutil.WriteFile(filename, []byte(data), 0644), "write file failed")

	conf, err := newConf(cpath)
	require.NoError(err, "create conf object failed")
	require.Equal(conf.Get("http.toml").Raw(), data, "should equal")
	ch := conf.WatchEvent("http.toml")
	time.Sleep(time.Second)
	timeout := time.NewTimer(2 * time.Second)

	data2 := `
[server]
ip=127.0.0.2
`
	require.Nil(ioutil.WriteFile(filename, []byte(data2), 0644), "rewrite file failed")
	select {
	case <-timeout.C:
		t.Fatal("receive event timeout")
	case event := <-ch:
		require.Equal(event.Op, EventUpdate, "error event type")
		require.Equal(event.Key, "http.toml", "error key")
		require.Equal(event.Val, data2, "error value")
	}

	conf.Stop()
	require.Equal(conf.Get("http.toml").Raw(), data2, "should equal")
}
