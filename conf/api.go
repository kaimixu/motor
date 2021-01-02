package conf

import "log"

var g Client

func Parse(path string) (err error) {
	if g, err = newConf(path); err != nil {
		return err
	}

	return nil
}

func Get(key string) *Value {
	return g.Get(key)
}

func WatchEvent(keys ...string) <-chan Event {
	return g.WatchEvent(keys...)
}

func Dump() string {
	return g.Dump()
}

func Stop() {
	g.Stop()
}

func Watch(key string, s Setter) error {
	v := g.Get(key)
	str := v.Raw()
	if err := s.Set(str); err != nil {
		return err
	}
	go func() {
		for event := range WatchEvent(key) {
			err := s.Set(event.Val)
			if err != nil {
				log.Printf("Watch: set value failed, event:%v", event)
			}
		}
	}()
	return nil
}
