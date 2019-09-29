// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"strings"

	"github.com/liyue201/lazyredis/config"
	"github.com/liyue201/lazyredis/gui"
	"github.com/liyue201/lazyredis/redis"
)

var (
	addr    = flag.String("addr", "localhost:6379", "redis address")
	pass    = flag.String("pass", "", "redis password")
	db      = flag.Int("db", 0, "redis database")
	cfgFile = flag.String("conf", "", "yaml config file")
)

func main() {
	flag.Parse()

	redisAddrs := strings.Split(*addr, ",")

	if *cfgFile != "" {
		conf, err := config.Load(*cfgFile)
		if err != nil {
			panic(err.Error())
		}

		redisAddrs = conf.Redis.Addr
		*pass = conf.Redis.Password
		*db = conf.Redis.Db
	}

	redisCli, err := redis.NewRedisClient(redisAddrs, *pass, *db)
	if err != nil {
		panic(err.Error())
	}
	defer redisCli.Close()

	w := gui.NewWindow()
	w.SetRedisClient(redisCli)

	w.Run()
	w.Close()
}
