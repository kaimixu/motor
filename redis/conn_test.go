package redis

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
)

func TestConn(t *testing.T) {
	TestFileConf(t)
	defer Close()

	wconn := GetConn("cluster1", WRITE)
	_, err := wconn.Do("SET", "key1", "value1")
	require.Nil(t, err)
	wconn.Close()

	rconn := GetConn("cluster1", READ)
	reply, err := redis.String(rconn.Do("GET", "key1"))
	require.Nil(t, err)
	defer rconn.Close()
	require.Equal(t, reply, "value1")
}
