// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/liyue201/lazyredis/config"
	"github.com/liyue201/lazyredis/gui"
	"github.com/liyue201/lazyredis/redis"
)
 
func main() {
	conf, err := config.Load("conf.yaml")
	if err != nil {
		panic(err.Error())
		return
	}
	redisCli, err := redis.NewRedisClient(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Db)
	if err != nil {
		panic(err.Error())
	}
	defer redisCli.Close()

	w := gui.NewWindow()
	w.SetRedisClient(redisCli)

	w.Run()
	w.Close()
}
