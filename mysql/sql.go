package mysql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/didi/gendry/builder"
	"github.com/didi/gendry/scanner"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kaimixu/motor/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type OpMode = uint8

const (
	READ OpMode = iota
	WRITE
)

type DB struct {
	*sql.DB
	ctx *gin.Context

	IsMaster bool
	Dbname   string
	Table    string
}

func GetDB(ctx *gin.Context, dbname, table string, m OpMode) *DB {
	db, err := _mysqlPool.getDB(dbname, m)
	if err != nil {
		zap.L().Error("GetDB failed",
			zap.String("dbname", dbname),
			zap.Any("opmode", m),
			zap.Error(err))
		return nil
	}

	return &DB{
		DB:       db,
		ctx:      ctx,
		IsMaster: m == WRITE,
		Dbname:   dbname,
		Table:    table,
	}
}

func (db *DB) GetList(where map[string]interface{}, selectFields []string, result interface{}) error {
	if db.ctx != nil {
		ctx, exists := trace.GetTraceCtx(db.ctx)
		if exists {
			span, _ := opentracing.StartSpanFromContext(ctx, "GetList")
			defer span.Finish()
		}
	}

	cond, vals, err := builder.BuildSelect(db.Table, where, selectFields)
	if nil != err {
		return errors.Wrap(err,
			fmt.Sprintf("builder.BuildSelect failed, table:%s, where:%v, selectFields:%v", db.Table, where, selectFields))
	}
	now := time.Now()
	defer slowLog(fmt.Sprintf("cond(%s) args(%+v)", cond, vals), now)

	rows, err := db.Query(cond, vals...)
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("db.Query failed, table:%s, cond:%s, vals:%v", db.Table, cond, vals))

	}
	defer rows.Close()

	err = scanner.Scan(rows, result)
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("scanner.Scan failed, table:%s, cond:%s, vals:%v", db.Table, cond, vals))
	}

	return nil
}

func (db *DB) GetRow(where map[string]interface{}, selectFields []string, result interface{}) error {
	err := db.GetList(where, selectFields, result)
	if err != nil {
		return errors.WithMessage(err, "db.GetList failed")
	}

	return nil
}

func (db *DB) Insert(data []map[string]interface{}) (int64, error) {
	if db.ctx != nil {
		ctx, exists := trace.GetTraceCtx(db.ctx)
		if exists {
			span, _ := opentracing.StartSpanFromContext(ctx, "Insert")
			defer span.Finish()
		}
	}

	cond, vals, err := builder.BuildInsert(db.Table, data)
	if err != nil {
		return 0, errors.Wrap(err,
			fmt.Sprintf("builder.BuildInsert failed, table:%s, data:%v", db.Table, data))
	}
	now := time.Now()
	defer slowLog(fmt.Sprintf("cond(%s) args(%+v)", cond, vals), now)

	result, err := db.Exec(cond, vals...)
	if err != nil {
		return 0, errors.Wrap(err,
			fmt.Sprintf("db.Exec failed, cond:%s, vals:%v", cond, vals))
	}

	return result.LastInsertId()
}

func (db *DB) Update(where map[string]interface{}, update map[string]interface{}) (int64, error) {
	if db.ctx != nil {
		ctx, exists := trace.GetTraceCtx(db.ctx)
		if exists {
			span, _ := opentracing.StartSpanFromContext(ctx, "Update")
			defer span.Finish()
		}
	}

	cond, vals, err := builder.BuildUpdate(db.Table, where, update)
	if nil != err {
		return 0, errors.Wrap(err,
			fmt.Sprintf("builder.BuildUpdate failed, table:%s, where:%v, update:%v", db.Table, where, update))
	}
	now := time.Now()
	defer slowLog(fmt.Sprintf("cond(%s) args(%+v)", cond, vals), now)

	result, err := db.Exec(cond, vals...)
	if nil != err {
		return 0, errors.Wrap(err,
			fmt.Sprintf("db.Exec failed, table:%s, cond:%v, vals:%v", db.Table, cond, vals))
	}

	return result.RowsAffected()
}

func (db *DB) NamedQuery(query string, data map[string]interface{}, result interface{}) error {
	if db.ctx != nil {
		ctx, exists := trace.GetTraceCtx(db.ctx)
		if exists {
			span, _ := opentracing.StartSpanFromContext(ctx, "NamedQuery")
			defer span.Finish()
		}
	}

	cond, vals, err := builder.NamedQuery(query, data)
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("builder.NamedQuery failed, table:%s, query:%v, data:%v", db.Table, query, data))
	}
	now := time.Now()
	defer slowLog(fmt.Sprintf("cond(%s) args(%+v)", cond, vals), now)

	rows, err := db.Query(cond, vals...)
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("db.Query failed, table:%s, cond:%s, vals:%v", db.Table, cond, vals))

	}
	defer rows.Close()

	err = scanner.Scan(rows, result)
	if err != nil {
		return errors.Wrap(err,
			fmt.Sprintf("scanner.Scan failed, table:%s, cond:%s, vals:%v", db.Table, cond, vals))
	}

	return nil
}

func slowLog(statement string, now time.Time) {
	dur := time.Since(now)
	if dur > SlowLogDur {
		zap.L().Warn("slow log",
			zap.String("sqlinfo", statement),
			zap.Duration("dur", dur))
	}
}
