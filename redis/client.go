package redis

import (
	"errors"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/wuxibin89/redis-go-cluster"
)

type Client interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
	Close() error
}

func NewRedisClient(addr [] string, password string, db int) (Client, error) {
	if len(addr) == 0 {
		return nil, errors.New("addr is empty")
	}
	if len(addr) > 1 {
		return NewClusterClient(&redis.Options{StartNodes: addr})
	}
	return NewSimpleClient(addr[0], redigo.DialPassword(password), redigo.DialDatabase(db))
}
