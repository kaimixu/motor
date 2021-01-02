package conf

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

const (
	defChSize = 10
)

var cpath string

type conf struct {
	raw     map[string]*Value
	content *Storage
	sync.Mutex
	watchChs map[string][]chan Event
	wg       sync.WaitGroup
	done     chan struct{}
}

func init() {
	flag.StringVar(&cpath, "c", "./configs", "config file path")
}

// create a conf object
func newConf(path string) (*conf, error) {
	if path != "" {
		cpath = path
	}
	cpath = filepath.FromSlash(cpath)
	cpath, _ = filepath.Abs(cpath)

	cmap, err := readDir(cpath)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("readDir(%v)", cpath))
	} else if len(cmap) == 0 {
		return nil, fmt.Errorf("empty conf path")
	}

	s := &Storage{}
	s.Store(cmap)
	c := &conf{
		raw:      cmap,
		content:  s,
		watchChs: make(map[string][]chan Event),
		done:     make(chan struct{}),
	}

	c.wg.Add(1)
	go c.monitor()
	return c, nil
}

// read all configuration files from rpath directory
func readDir(rpath string) (map[string]*Value, error) {
	fi, err := os.Stat(rpath)
	if err != nil {
		return nil, errors.Wrap(err, "os.Stat")
	} else if !fi.IsDir() {
		return nil, errors.Wrap(err, "rpath not a directory")
	}

	files, err := ioutil.ReadDir(rpath)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil.ReadDir")
	}

	cmap := make(map[string]*Value)
	for _, f := range files {
		if !f.IsDir() {
			content, err := readFile(path.Join(rpath, f.Name()))
			if err != nil {
				return nil, err
			}
			cmap[f.Name()] = &Value{raw: content}
		}
	}

	return cmap, nil
}

func readFile(fpath string) (string, error) {
	content, err := ioutil.ReadFile(fpath)
	return string(content), errors.Wrap(err, "ioutil.ReadFile")
}

// get value by key
func (c *conf) Get(key string) *Value {
	return c.content.Get(key)
}

// receive notification of configuration change
func (c *conf) WatchEvent(keys ...string) <-chan Event {
	ch := make(chan Event, defChSize)
	c.Lock()
	defer c.Unlock()
	for _, key := range keys {
		c.watchChs[key] = append(c.watchChs[key], ch)
	}

	return ch
}

// stop monitor goroutine
func (c *conf) Stop() {
	c.done <- struct{}{}
	c.wg.Wait()
}

func (c *conf) Dump() string {
	m := c.content.Load()
	var buf bytes.Buffer

	for k, v := range m {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v.Raw())
	}

	return buf.String()
}

// monitor file change
func (c *conf) monitor() {
	defer c.wg.Done()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("fsnotify.NewWatcher failed, err: %+v", err)
		return
	}
	defer watcher.Close()
	err = watcher.Add(cpath)
	if err != nil {
		log.Printf("watcher.Add(%s) failed, err: %+v", cpath, err)
		return
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				break
			}
			// vim 修改文件时会先触发Create事件，再触发Write事件
			if event.Op&fsnotify.Write != 0 {
				c.reloadFile(event.Name)
			} else {
				log.Printf("monitor: unsupport event %+v", event)
			}
		case err := <-watcher.Errors:
			log.Printf("error: %+v", err)
		case <-c.done:
			log.Printf("receive stop notify")
			return
		}
	}
}

// read the file again, and send event notification
func (c *conf) reloadFile(name string) {
	if filterHideFile(name) || filterBackupFile(name) {
		return
	}

	// 这里休眠一段时间，避免文件内容还未更新
	time.Sleep(500 * time.Millisecond)
	key := filepath.Base(name)
	val, err := readFile(name)
	if err != nil {
		log.Printf("readFile(%s) failed, error: %+v", name, err)
		return
	}
	c.raw[key] = &Value{raw: val}
	c.content.Store(c.raw)

	c.Lock()
	chs := c.watchChs[key]
	c.Unlock()

	for _, ch := range chs {
		select {
		case ch <- Event{Op: EventUpdate, Key: key, Val: val}:
		default:
			log.Printf("event channel full discard file %s update event, content:%+v", name, val)
		}
	}
}

func filterHideFile(name string) bool {
	if runtime.GOOS == "linux" {
		return strings.HasPrefix(filepath.Base(name), ".")
	}

	return false
}

func filterBackupFile(name string) bool {
	if runtime.GOOS == "linux" {
		return strings.HasSuffix(name, "~")
	}

	return false
}
