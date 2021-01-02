package conf

const (
	EventUpdate EventType = iota
)

type EventType int

type Event struct {
	Op  EventType
	Key string
	Val string
}

type Setter interface {
	Set(string) error
}

type Client interface {
	// get configuration according to filename
	Get(string) *Value
	// receive notification of configuration change
	WatchEvent(...string) <-chan Event
	// stop monitoring configuration change
	Stop()
	// return all configuration as a string
	Dump() string
}
