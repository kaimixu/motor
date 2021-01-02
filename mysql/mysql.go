package mysql

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/didi/gendry/manager"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kaimixu/motor/conf"
	"github.com/kaimixu/motor/naming"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	cacheKey = "conncache"
)

type MysqlConfLoadMode = uint8

const (
	ModeNaming MysqlConfLoadMode = iota
	ModeFile
)

var (
	_mysqlPool *mysqlPool

	SlowLogDur = 300 * time.Millisecond
)

type namingMysqlInstanceAttr = mysqlConf

type mysqlNodeConf struct {
	Username string `json:"username" toml:"username"`
	Password string `json:"password" toml:"password"`
	IP       string `json:"host" toml:"ip"`
	Port     int    `json:"port" toml:"port"`
}

type mysqlClusterConf struct {
	// 仅从文件中加载配置时生效
	Idc string `json:"-" toml:"idc"`

	Master []mysqlNodeConf `json:"master" toml:"master"`
	Slave  []mysqlNodeConf `json:"slave" toml:"slave"`

	// 单位：秒
	ConnMaxLifetime int `json:"conn_max_lifetime" toml:"connMaxLifetime"`
	MaxOpenConns    int `json:"conn_max_open" toml:"maxOpenConns"`
	MaxIdleConns    int `json:"conn_max_idle" toml:"maxIdleConns"`

	// 单位：秒
	ConnTimeout int `json:"conn_timeout" toml:"connTimeout"`
	// 单位：秒
	ReadTimeout int `json:"read_timeout" toml:"readTimeout"`
	// 单位：秒
	WriteTimeout int `json:"write_timeout" toml:"writeTimeout"`
}

type mysqlConf struct {
	Database map[string]mysqlClusterConf `json:"database" toml:"database"`
}

type mysqlPool struct {
	productLine string
	idc         string
	pubenv      string

	confLoadMode MysqlConfLoadMode

	// 内容格式：map[dbname][]*sql.DB
	mMap sync.Map
	sMap sync.Map
}

// confLoadMode: 配置加载方式
// productLine: 产品线，仅confLoadMode=ModeNaming有效
// idc: 机房,仅confLoadMode=ModeFile 有效，表示仅生成指定机房的连接
// pubenv: 部署环境，仅confLoadMode=ModeNaming有效
func InitMysql(confLoadMode MysqlConfLoadMode, productLine, idc, pubenv string) {
	if _mysqlPool != nil {
		zap.L().Info("cannot be re-iinitialized")
		return
	}

	db := &mysqlPool{
		productLine: productLine,
		idc:         idc,
		pubenv:      pubenv,

		confLoadMode: confLoadMode,
	}

	if confLoadMode == ModeFile {
		err := db.loadConfFromFile()
		if err != nil {
			panic(err.Error())
		}
	} else {
		err := db.loadConfFromNaming()
		if err != nil {
			panic(err.Error())
		}
	}

	_mysqlPool = db
}

// 连接池释放
func Close() {
	if _mysqlPool != nil {
		_mysqlPool.close()
	}
}

func (my *mysqlPool) loadConfFromFile() error {
	var cfg mysqlConf
	if err := conf.Get("mysql.toml").UnmarshalTOML(&cfg); err != nil {
		return errors.Wrap(err, "Get(mysql.toml).UnmarshalTOML failed")
	}
	my.parseFileConf(&cfg)

	go func() {
		for _ = range conf.WatchEvent("mysql.toml") {
			var cfg mysqlConf

			if err := conf.Get("mysql.toml").UnmarshalTOML(&cfg); err != nil {
				zap.L().Error("Get(mysql.toml).UnmarshalTOML failed",
					zap.Error(err))
				continue
			}

			my.parseFileConf(&cfg)
		}
	}()

	return nil
}

func (my *mysqlPool) parseFileConf(cfg *mysqlConf) {
	mMap := make(map[string][]*sql.DB)
	sMap := make(map[string][]*sql.DB)

	for dbname, cluster := range cfg.Database {
		if my.idc != "" && my.idc != cluster.Idc {
			continue
		}

		for _, dbconf := range cluster.Master {
			db, err := manager.New(
				dbname,
				dbconf.Username,
				dbconf.Password,
				dbconf.IP).Set(
				manager.SetCharset("utf8"),
				manager.SetAllowCleartextPasswords(true),
				manager.SetInterpolateParams(true),
				manager.SetParseTime(true),
				manager.SetTimeout(time.Duration(cluster.ConnTimeout)*time.Second),
				manager.SetReadTimeout(time.Duration(cluster.ReadTimeout)*time.Second),
				manager.SetWriteTimeout(time.Duration(cluster.WriteTimeout)*time.Second),
			).Port(dbconf.Port).Open(true)
			if err != nil {
				zap.L().Error("create mysql master db failed",
					zap.Error(err),
					zap.Any("dbconf", dbconf))
				continue
			}
			db.SetMaxIdleConns(cluster.MaxIdleConns)
			db.SetMaxOpenConns(cluster.MaxOpenConns)
			db.SetConnMaxLifetime(time.Duration(cluster.ConnMaxLifetime) * time.Second)
			mMap[dbname] = append(mMap[dbname], db)
		}

		for _, dbconf := range cluster.Slave {
			db, err := manager.New(
				dbname,
				dbconf.Username,
				dbconf.Password,
				dbconf.IP).Set(
				manager.SetCharset("utf8"),
				manager.SetAllowCleartextPasswords(true),
				manager.SetInterpolateParams(true),
				manager.SetParseTime(true),
				manager.SetTimeout(time.Duration(cluster.ConnTimeout)*time.Second),
				manager.SetReadTimeout(time.Duration(cluster.ReadTimeout)*time.Second),
				manager.SetWriteTimeout(time.Duration(cluster.WriteTimeout)*time.Second),
			).Port(dbconf.Port).Open(true)
			if err != nil {
				zap.L().Error("create mysql slave db failed",
					zap.Error(err),
					zap.Any("dbconf", dbconf))
				continue
			}
			db.SetMaxIdleConns(cluster.MaxIdleConns)
			db.SetMaxOpenConns(cluster.MaxOpenConns)
			db.SetConnMaxLifetime(time.Duration(cluster.ConnMaxLifetime) * time.Second)
			sMap[dbname] = append(sMap[dbname], db)
		}
	}

	my.mMap.Store(cacheKey, mMap)
	my.sMap.Store(cacheKey, sMap)
}

func (my *mysqlPool) loadConfFromNaming() error {
	builder := naming.Build()
	if builder == nil {
		return errors.New("naming.Build failed")
	}

	sn := "mysql/" + my.productLine
	resolver, err := builder.Discovery(sn)
	if err != nil {
		return errors.New(fmt.Sprintf("builder.Discovery failed, sn:%s", sn))
	}

	// 首次获取
	for {
		<-resolver.Watch()
		c, ok := resolver.Fetch()
		if !ok {
			continue
		}
		my.parseNamingInstance(c)
		break
	}

	// 监听配置改动
	go func() {
		for {
			<-resolver.Watch()
			c, ok := resolver.Fetch()
			if !ok {
				continue
			}

			my.parseNamingInstance(c)
		}
	}()

	return nil
}

func (my *mysqlPool) parseNamingInstance(ins []*naming.Instance) {
	// 配置被删除
	if len(ins) == 0 {
		my.mMap.Store(cacheKey, make(map[string][]*sql.DB))
		my.sMap.Store(cacheKey, make(map[string][]*sql.DB))
		return
	}

	mMap := make(map[string][]*sql.DB)
	sMap := make(map[string][]*sql.DB)
	for _, in := range ins {
		if my.idc != "" && in.Idc != my.idc {
			continue
		}
		if my.pubenv != "" && in.PubEnv != my.pubenv {
			continue
		}

		var attr mysqlConf
		err := in.StructuredAttr(&attr)
		if err != nil {
			zap.L().Error("invalid instance",
				zap.Error(err),
				zap.Any("in", in))
			continue
		}

		for dbname, cluster := range attr.Database {
			for _, dbconf := range cluster.Master {
				db, err := manager.New(
					dbname,
					dbconf.Username,
					dbconf.Password,
					dbconf.IP).Set(
					manager.SetCharset("utf8"),
					manager.SetAllowCleartextPasswords(true),
					manager.SetInterpolateParams(true),
					manager.SetParseTime(true),
					manager.SetTimeout(time.Duration(cluster.ConnTimeout)*time.Second),
					manager.SetReadTimeout(time.Duration(cluster.ReadTimeout)*time.Second),
					manager.SetWriteTimeout(time.Duration(cluster.WriteTimeout)*time.Second),
				).Port(dbconf.Port).Open(true)
				if err != nil {
					zap.L().Error("create mysql master db failed",
						zap.Error(err),
						zap.Any("dbconf", dbconf))
					continue
				}
				db.SetMaxIdleConns(cluster.MaxIdleConns)
				db.SetMaxOpenConns(cluster.MaxOpenConns)
				db.SetConnMaxLifetime(time.Duration(cluster.ConnMaxLifetime) * time.Second)
				mMap[dbname] = append(mMap[dbname], db)
			}

			for _, dbconf := range cluster.Slave {
				db, err := manager.New(
					dbname,
					dbconf.Username,
					dbconf.Password,
					dbconf.IP).Set(
					manager.SetCharset("utf8"),
					manager.SetAllowCleartextPasswords(true),
					manager.SetInterpolateParams(true),
					manager.SetParseTime(true),
					manager.SetTimeout(time.Duration(cluster.ConnTimeout)*time.Second),
					manager.SetReadTimeout(time.Duration(cluster.ReadTimeout)*time.Second),
					manager.SetWriteTimeout(time.Duration(cluster.WriteTimeout)*time.Second),
				).Port(dbconf.Port).Open(true)
				if err != nil {
					zap.L().Error("create mysql slave db failed",
						zap.Error(err),
						zap.Any("dbconf", dbconf))
					continue
				}
				db.SetMaxIdleConns(cluster.MaxIdleConns)
				db.SetMaxOpenConns(cluster.MaxOpenConns)
				db.SetConnMaxLifetime(time.Duration(cluster.ConnMaxLifetime) * time.Second)
				sMap[dbname] = append(sMap[dbname], db)
			}
		}
	}

	my.mMap.Store(cacheKey, mMap)
	my.sMap.Store(cacheKey, sMap)
}

// 释放连接
func (my *mysqlPool) close() {
	val, ok := my.mMap.Load(cacheKey)
	if ok {
		go func(val interface{}) {

			oldmMap, _ := val.(map[string][]*sql.DB)
			for dbname, dbs := range oldmMap {
				for _, db := range dbs {
					err := db.Close()
					if err != nil {
						zap.L().Warn("db.Close failed",
							zap.Error(err),
							zap.String("dbname", dbname))
					}
				}
			}
		}(val)
	}

	val, ok = my.sMap.Load(cacheKey)
	if ok {
		go func(val interface{}) {
			oldsMap, _ := val.(map[string][]*sql.DB)
			for dbname, dbs := range oldsMap {
				for _, db := range dbs {
					err := db.Close()
					if err != nil {
						zap.L().Warn("db.Close failed",
							zap.Error(err),
							zap.String("dbname", dbname))
					}
				}
			}
		}(val)
	}
}

func (my *mysqlPool) getDB(dbname string, m OpMode) (*sql.DB, error) {
	if m == READ {
		val, ok := my.sMap.Load(cacheKey)
		if !ok {
			return nil, errors.New("mysql slave config uninitialized")
		}

		sMap := val.(map[string][]*sql.DB)
		dbSlice, ok := sMap[dbname]
		if !ok || len(dbSlice) == 0 {
			return nil, errors.New(fmt.Sprintf("db(%s) slave config uninitialized", dbname))
		}

		return dbSlice[randInt(len(dbSlice))], nil
	} else {
		val, ok := my.mMap.Load(cacheKey)
		if !ok {
			return nil, errors.New("mysql master config uninitialized")
		}

		mMap := val.(map[string][]*sql.DB)
		dbSlice, ok := mMap[dbname]
		if !ok || len(dbSlice) == 0 {
			return nil, errors.New(fmt.Sprintf("db(%s) master config uninitialized", dbname))
		}

		return dbSlice[randInt(len(dbSlice))], nil
	}
}

func randInt(max int) int {
	rand.Seed(time.Now().Unix())

	return rand.Intn(max)
}
