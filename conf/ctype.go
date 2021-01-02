package conf

import (
	"time"

	"github.com/kaimixu/motor/util"
	"github.com/pkg/errors"
)

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	value, err := time.ParseDuration(string(text))
	if err == nil {
		*d = Duration(value)
		return nil
	}

	return errors.Wrap(err, "time.ParseDuration")
}

type ByteSize uint64

func (d *ByteSize) UnmarshalText(text []byte) error {
	value, err := util.StringToBytes(string(text))
	if err == nil {
		*d = ByteSize(value)
		return nil
	}

	return errors.Wrap(err, "util.ToBytes")
}
