package tolerant

import (
	"testing"

	"github.com/kaimixu/motor/conf"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	assert := assert.New(t)
	assert.Nil(conf.Parse("../test/configs"))

	Init()
}
