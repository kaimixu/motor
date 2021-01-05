package trace

import (
	"testing"

	"github.com/kaimixu/motor/conf"
	"github.com/stretchr/testify/assert"
)

func TestInitJaeger(t *testing.T) {
	assert := assert.New(t)

	assert.Nil(conf.Parse("../test/configs"))
	o := newJaeger("TestJaeger")
	defer o.Close()
	return
}
