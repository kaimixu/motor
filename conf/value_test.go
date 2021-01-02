package conf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testconf struct {
	Addr    string
	Timeout Duration
}

func TestValue(t *testing.T) {
	v := &Value{`
[Server]
addr="127.0.0.1:8080"
timeout="10s"
`}
	var cf Toml
	var tc testconf
	require.Nil(t, v.Unmarshal(&cf))
	require.Nil(t, cf.Get("server").UnmarshalTOML(&tc))
	require.Equal(t, tc.Addr, "127.0.0.1:8080")
	require.Equal(t, tc.Timeout, Duration(10*time.Second))
}
