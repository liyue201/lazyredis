package redis

import "github.com/garyburd/redigo/redis"

type SimpleClient struct {
	conn redis.Conn
}

func NewSimpleClient(addr string, opts ...redis.DialOption) (Client, error) {
	c, err := redis.Dial("tcp", addr, opts...)
	return &SimpleClient{conn: c}, err
}

func (c *SimpleClient) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	return c.conn.Do(cmd, args...)
}

func (c *SimpleClient) Close() error {
	return c.conn.Close()
}