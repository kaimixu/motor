package conf

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

type Toml = Storage

func (m *Toml) Set(data string) error {
	if err := m.UnmarshalText([]byte(data)); err != nil {
		return err
	}
	return nil
}

func (m *Toml) UnmarshalText(data []byte) error {
	raws := map[string]interface{}{}
	if err := toml.Unmarshal(data, &raws); err != nil {
		return err
	}
	values := map[string]*Value{}
	for k, v := range raws {
		k = KeyName(k)
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Map:
			buf := bytes.NewBuffer(nil)
			err := toml.NewEncoder(buf).Encode(v)
			// b, err := toml.Marshal(v)
			if err != nil {
				return errors.Wrap(err, "toml.NewEncoder.Encode")
			}
			// NOTE: value is map[string]interface{}
			values[k] = &Value{raw: buf.String()}
		case reflect.Slice:
			raw := map[string]interface{}{
				k: v,
			}
			buf := bytes.NewBuffer(nil)
			err := toml.NewEncoder(buf).Encode(raw)
			// b, err := toml.Marshal(raw)
			if err != nil {
				return errors.Wrap(err, "toml.NewEncoder.Encode")
			}
			// NOTE: value is []interface{}
			values[k] = &Value{raw: buf.String()}
		case reflect.Bool:
			b := v.(bool)
			values[k] = &Value{raw: strconv.FormatBool(b)}
		case reflect.Int64:
			i := v.(int64)
			values[k] = &Value{raw: strconv.FormatInt(i, 10)}
		case reflect.Float64:
			f := v.(float64)
			values[k] = &Value{raw: strconv.FormatFloat(f, 'f', -1, 64)}
		case reflect.String:
			s := v.(string)
			values[k] = &Value{raw: s}
		default:
			return errors.Wrapf(fmt.Errorf("unknown kind(%v)", rv.Kind()), "UnmarshalText")
		}
	}
	m.Store(values)
	return nil
}
