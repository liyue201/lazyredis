package redis

import "github.com/garyburd/redigo/redis"

type SimpleClient struct {
	conn redis.Conn
} 

func NewSimpleClient(addr string, opts ...redis.DialOption) (Client, error) {
	c, err := redis.Dial("tcp", addr, opts...)
	return &SimpleClient{conn: c}, err
}

func (c *SimpleClient) Do(cmd string, args ...interface{}) (string, error) {
	reply, err := c.conn.Do(cmd, args...)
	if err != nil {
		return "", err
	}
	return replyText(reply), nil
}

func (c *SimpleClient) Close() error {
	return c.conn.Close()
}
