package redis

import "github.com/go-redis/redis"

func NewClient(opts *redis.UniversalOptions) redis.UniversalClient {
	return redis.NewUniversalClient(opts)
}

func Ping(cli redis.UniversalClient)  {

}