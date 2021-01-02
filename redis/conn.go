package redis

import (
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type OpMode = uint8

const (
	READ OpMode = iota
	WRITE
)

type RedisConn struct {
	redis.Conn
	IsMaster    bool
	clusterName string
}

func GetConn(clusterName string, m OpMode) *RedisConn {
	conn, err := _redisPool.getConn(clusterName, m)
	if err != nil {
		zap.L().Error("getConn failed",
			zap.String("clusterName", clusterName),
			zap.Any("opmode", m),
			zap.Error(err))
		return nil
	}

	return &RedisConn{
		Conn:        conn,
		IsMaster:    m == WRITE,
		clusterName: clusterName,
	}
}
