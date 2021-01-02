package naming

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kaimixu/motor/conf"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/pkg/transport"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_once    sync.Once
	_builder Builder
)

type etcdConf struct {
	LeaseTTL    int64
	DialTimeout uint64
	SnPrefix    string
	Addrs       []string

	BasicAuth bool
	Username  string
	Password  string

	TlsEnable  bool
	Cacert     string
	ServerCert string
	ServerKey  string
}

type EtcdBuilder struct {
	conf *etcdConf

	client     *clientv3.Client
	ctx        context.Context
	cancelFunc context.CancelFunc

	mutex   sync.RWMutex
	servers map[string]*serverInfo

	rmutex   sync.RWMutex
	registry map[string]struct{}
}

type serverInfo struct {
	sn       string
	resolver map[*Resolve]struct{}
	ins      atomic.Value
	e        *EtcdBuilder
	once     sync.Once
}

type Resolve struct {
	sn    string
	e     *EtcdBuilder
	event chan struct{}
}

func getConf() (*etcdConf, error) {
	var st conf.Storage
	var cfg etcdConf
	if err := conf.Get("naming.toml").Unmarshal(&st); err != nil {
		return nil, errors.Wrap(err, "Get(naming.toml).Unmarshal failed")
	}
	if err := st.Get("Etcd").UnmarshalTOML(&cfg); err != nil {
		return nil, errors.Wrap(err, "Get(etcd).UnmarshalTOML failed")
	}

	if len(cfg.Addrs) == 0 {
		return nil, errors.New(fmt.Sprintf("invalid etcd config addr:%+v", cfg.Addrs))
	}

	return &cfg, nil
}

func singleton() Builder {
	_once.Do(func() {
		b := create()
		if b != nil {
			_builder = b
		}
	})
	return _builder
}

func Build() Builder {
	return singleton()
}

func create() *EtcdBuilder {
	econf, err := getConf()
	if err != nil {
		zap.L().Error(fmt.Sprintf("%+v", err))
		return nil
	}

	c := clientv3.Config{
		Endpoints:   econf.Addrs,
		DialTimeout: time.Second * time.Duration(econf.DialTimeout),
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}
	if econf.BasicAuth {
		c.Username = econf.Username
		c.Password = econf.Password
	} else if econf.TlsEnable {
		tlsInfo := transport.TLSInfo{
			CertFile:      econf.ServerCert,
			KeyFile:       econf.ServerKey,
			TrustedCAFile: econf.Cacert,
		}
		tlsconf, err := tlsInfo.ClientConfig()
		if err != nil {
			zap.L().Error(fmt.Sprintf("%+v", err))
			return nil
		}
		c.TLS = tlsconf
	}

	client, err := clientv3.New(c)
	if err != nil {
		zap.L().Error(fmt.Sprintf("clientv3.New failed, err:%+v", err))
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &EtcdBuilder{
		conf:       econf,
		client:     client,
		ctx:        ctx,
		cancelFunc: cancel,

		servers:  map[string]*serverInfo{},
		registry: map[string]struct{}{},
	}
}

// 服务发现
func (e *EtcdBuilder) Discovery(sn string) (Resolver, error) {
	if len(sn) == 0 {
		return nil, errors.New("sn argument cannot be null")
	}

	r := &Resolve{
		sn:    sn,
		e:     e,
		event: make(chan struct{}, 1),
	}
	e.mutex.Lock()
	srv, ok := e.servers[sn]
	if !ok {
		srv = &serverInfo{
			sn:       sn,
			resolver: make(map[*Resolve]struct{}),
			e:        e,
		}
		e.servers[sn] = srv
	}
	srv.resolver[r] = struct{}{}
	e.mutex.Unlock()

	srv.once.Do(func() {
		go srv.watch()
	})

	return r, nil
}

// 服务注册
func (e *EtcdBuilder) Register(in *Instance) (cancelFunc context.CancelFunc, err error) {
	if len(in.Name) == 0 || len(in.Idc) == 0 || len(in.PubEnv) == 0 || in.Attr.Data == nil {
		return nil, errors.New(fmt.Sprintf("argument cannot be empty, ins:%v", in))
	}

	e.rmutex.Lock()
	if _, ok := e.registry[in.Name]; ok {
		err = errors.New(fmt.Sprintf("cannot duplicate register, ins:%+v", in.Name))
	} else {
		e.registry[in.Name] = struct{}{}
	}
	e.rmutex.Unlock()
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(e.ctx)
	if err = e.registerLease(ctx, in); err != nil {
		e.rmutex.Lock()
		delete(e.registry, in.Name)
		e.rmutex.Unlock()
		cancel()
		zap.L().Error(fmt.Sprintf("%+v", err))
		return
	}
	ch := make(chan struct{}, 1)
	cancelFunc = context.CancelFunc(func() {
		cancel()
		<-ch
	})

	go func() {
		ticker := time.NewTicker(time.Duration(e.conf.LeaseTTL/3) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := e.registerLease(ctx, in)
				if err != nil {
					zap.L().Error(fmt.Sprintf("%+v", err))
				}
			case <-ctx.Done():
				_ = e.unregister(in)
				ch <- struct{}{}
				return
			}
		}
	}()

	return
}

// 服务注册与续约
func (e *EtcdBuilder) registerLease(ctx context.Context, in *Instance) (err error) {
	key := e.key(in.Name, in.Idc, in.PubEnv)
	val, _ := json.Marshal(in)

	ttlResp, err := e.client.Grant(context.TODO(), e.conf.LeaseTTL)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("client.Grant failed, ins:%+v", in))
	}

	_, err = e.client.Put(ctx, key, string(val), clientv3.WithLease(ttlResp.ID))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("client.Put failed, key:%s, in:%+v", key, in))
	}

	return nil
}

// 服务注销
func (e *EtcdBuilder) unregister(ins *Instance) (err error) {
	key := e.key(ins.Name, ins.Idc, ins.PubEnv)
	if _, err = e.client.Delete(context.TODO(), key); err != nil {
		zap.L().Error(fmt.Sprintf("client.Delete failed, err:%+v", err),
			zap.String("key", key),
			zap.Any("ins", ins))
		return
	}

	zap.L().Info("client.Delete success", zap.String("key", key), zap.Any("ins", ins))
	return
}

func (e *EtcdBuilder) key(args ...string) string {
	key := e.conf.SnPrefix
	if !strings.HasSuffix(key, "/") {
		key = key + "/"
	}

	var buf bytes.Buffer
	buf.WriteString(key)
	for _, arg := range args {
		buf.WriteString(arg)
		buf.WriteString("/")
	}
	return strings.TrimRight(buf.String(), "/")
}

func (e *EtcdBuilder) Close() {
	e.cancelFunc()
	e.client.Close()
}

func (srv *serverInfo) watch() {
	_ = srv.getstore("get")
	watchChan := srv.e.client.Watch(srv.e.ctx, srv.e.key(srv.sn), clientv3.WithPrefix())
	for wresp := range watchChan {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT || ev.Type == mvccpb.DELETE {
				_ = srv.getstore("watch")
			}
		}
	}
}

func (srv *serverInfo) getstore(typ string) error {
	resp, err := srv.e.client.Get(srv.e.ctx, srv.e.key(srv.sn), clientv3.WithPrefix())
	if err != nil {
		zap.L().Error(fmt.Sprintf("client.Get failed, err:%+v", err), zap.String("sn", srv.sn))
		return err
	}
	// 首次get时服务可能还会注册，此情况下不唤醒resolver
	if typ == "get" && len(resp.Kvs) == 0 {
		zap.L().Info("naming.getstore: client.get return null",
			zap.String("typ", typ),
			zap.String("sn", srv.sn))
		return nil
	}

	ins, err := srv.parseIns(resp)
	if err != nil {
		zap.L().Error(fmt.Sprintf("parseIns failed, err:%+v", err), zap.String("sn", srv.sn))
		return err
	}

	srv.store(ins)
	return nil
}

func (srv *serverInfo) parseIns(resp *clientv3.GetResponse) (insI *InstancesInfo, err error) {
	insI = &InstancesInfo{
		Instances: make(map[string][]*Instance),
	}

	for _, kv := range resp.Kvs {
		in := new(Instance)

		err := json.Unmarshal(kv.Value, in)
		if err != nil {
			return nil, err
		}

		insI.Instances[in.Name] = append(insI.Instances[in.Name], in)
	}

	return insI, nil
}

func (srv *serverInfo) store(ins *InstancesInfo) {
	srv.ins.Store(ins)
	srv.e.mutex.RLock()
	for r := range srv.resolver {
		select {
		case r.event <- struct{}{}:
		default:
		}
	}
	srv.e.mutex.RUnlock()
}

func (r *Resolve) Watch() <-chan struct{} {
	return r.event
}

// 服务被delete后返回的ins=nil，但ok=true
func (r *Resolve) Fetch() (ins []*Instance, ok bool) {
	r.e.mutex.RLock()
	srv, ok := r.e.servers[r.sn]
	r.e.mutex.RUnlock()
	if ok {
		insI, ok := srv.ins.Load().(*InstancesInfo)
		if !ok {
			return nil, ok
		}
		return insI.Instances[r.sn], ok
	}
	return
}

func (r *Resolve) Close() {
	r.e.mutex.Lock()
	if srv, ok := r.e.servers[r.sn]; ok && len(srv.resolver) != 0 {
		delete(srv.resolver, r)
	}
	r.e.mutex.Unlock()
}
