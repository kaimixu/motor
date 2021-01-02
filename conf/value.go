package conf

import (
	"encoding"

	"github.com/BurntSushi/toml"
)

type Value struct {
	raw string
}

func (v *Value) Raw() string {
	if v.raw == "" {
		return ""
	}
	return v.raw
}

func (v *Value) Unmarshal(un encoding.TextUnmarshaler) error {
	return un.UnmarshalText([]byte(v.Raw()))
}

func (v *Value) UnmarshalTOML(dst interface{}) error {
	return toml.Unmarshal([]byte(v.Raw()), dst)
}
