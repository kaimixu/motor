package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToBytes(t *testing.T) {
	assert := assert.New(t)

	unitm := map[string]float64{
		"1b": 1,
		"1B": 1,

		"1k":   KB,
		"1K":   KB,
		"1KB":  KB,
		"1kb":  KB,
		"10KB": 10 * KB,

		"1m":     MB,
		"1M":     MB,
		"1MB":    MB,
		"1mb":    MB,
		"1024mb": GB,

		"1g":     GB,
		"1G":     GB,
		"1gb":    GB,
		"1GB":    GB,
		"1024GB": TB,

		"1t":     TB,
		"1T":     TB,
		"1TB":    TB,
		"1tb":    TB,
		"1024TB": PB,

		"1p":     PB,
		"1P":     PB,
		"1PB":    PB,
		"1pb":    PB,
		"1024PB": 1024 * PB,
	}

	for s, v := range unitm {
		val, err := StringToBytes(s)
		assert.Nil(err)
		assert.Equal(val, v)
	}

	t.Log(unitm["1024PB"])
	_, err := StringToBytes("1kbb")
	assert.NotEmpty(err)
}

func TestBytesToString(t *testing.T) {
	assert := assert.New(t)

	v1 := float64(1)
	assert.Equal(BytesToString(v1), "1B")

	v2 := float64(1024)
	assert.Equal(BytesToString(v2), "1KB")
	v3 := float64(1025)
	assert.Equal(BytesToString(v3), "1KB")

	v4 := float64(5*MB + 10*KB)
	assert.Equal(BytesToString(v4), "5.01MB")
	v5 := float64(5*MB + 10*KB + 10*B)
	assert.Equal(BytesToString(v5), "5.01MB")

	v6 := float64(10*GB + 5*KB)
	assert.Equal(BytesToString(v6), "10GB")

	v7 := float64(1024*GB + 1*TB)
	assert.Equal(BytesToString(v7), "2TB")

	v8 := float64(1025 * PB)
	assert.Equal(BytesToString(v8), "1025PB")

	assert.Equal(BytesToString(0), "0")
}
