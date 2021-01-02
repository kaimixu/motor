package conf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorage(t *testing.T) {
	m := map[string]*Value{
		"key1": &Value{"value1"},
		"key2": &Value{"value2"},
		"key3": &Value{"value3"},
		"key4": &Value{"value4"},
		"key5": &Value{"value5"},
	}

	var s Storage
	s.Store(m)
	v := s.Get("key1")
	require := require.New(t)
	require.Equal(v.Raw(), "value1", "key1.value != value1")

	v = s.Get("KEY1")
	require.Equal(v.Raw(), "value1", "KEY1.value != value1")

	m2 := s.Load()
	require.Equal(len(m2), len(m), "storage.Load() return error map")
	for k, v := range m2 {
		require.Equal(m[k], v, "storage.Load() return map include error k/v")
	}
}
