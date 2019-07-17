// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jroimartin/gocui"
	"github.com/liyue201/lazyredis/config"
	"jlog"
	"log"
	"strings"
)

var appConfig *config.AppConfig
var redisCli redis.UniversalClient

func nextView(g *gocui.Gui, v *gocui.View) error {
	views := g.Views()

	curViewIdx := 0
	for i, v1 := range views {
		if v1.Name() == v.Name() {
			curViewIdx = i
			break
		}
	}
	curView := views[(curViewIdx+1)%len(views)]

	g.SetCurrentView(curView.Name())
	g.SetViewOnTop(curView.Name())
	if curView.Name() == "side" {
		g.Cursor = true
	} else {
		g.Cursor = false
	}
	jlog.Logger.Infof("cur view: %s", curView.Name())
	return nil
}

func newLine(v *gocui.View) {
	line := fmt.Sprintf("%s>", appConfig.Redis.Addr[0])
	writeText(v, line)
}

func writeText(v *gocui.View, str string) {
	runes := []rune(str)
	for _, c := range runes {
		v.EditWrite(c)
	}
}

func IgnoreKey(g *gocui.Gui, v *gocui.View) error {
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	var line string
	var err error

	defer func() {
		_, y := v.Cursor()
		v.SetCursor(0, y+1)
		newLine(v)
	}()

	_, cy := v.Cursor()
	if line, err = v.Line(cy); err != nil {
		line = ""
		return nil
	}
	jlog.Logger.Infof("line: %v", line)

	cmd := ""
	splits := strings.Split(line, ">")
	if len(splits) > 1 {
		cmd = splits[1]
	}

	retStr := ""
	if cmd == "ping" {
		ret := redisCli.Ping()
		retStr = ret.String()
		err = ret.Err()
	} else if cmd == "info" {
		ret := redisCli.Info()
		retStr = ret.String()
		err = ret.Err()
	} else {
		err = errors.New("unsupported command: " + cmd)
	}

	mainView, _ := g.View("main")
	mainView.Clear()
	mainView.SetCursor(0, 0)
	if err == nil {
		str := strings.ReplaceAll(retStr, "\r\n", "\n")
		fmt.Fprintln(mainView, str)
		jlog.Logger.Info(retStr)
	} else {
		fmt.Fprintln(mainView, err.Error())
		jlog.Logger.Error(err.Error())
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	keys := []gocui.Key{gocui.KeyArrowLeft, gocui.KeyArrowDown, gocui.KeyArrowRight, gocui.KeyArrowUp, gocui.KeyDelete, gocui.KeyBackspace2}
	for _, key := range keys {
		g.SetKeybinding("side", key, gocui.ModNone, IgnoreKey);
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", 1, 1, 30, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "cmd"
		v.Editable = true
		v.Autoscroll = true
		newLine(v)

		if _, err = g.SetCurrentView("side"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("main", 31, 1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Autoscroll = true
		v.Title = "output"
	}
	return nil
}

func main() {
	conf, err := config.Load("conf.yaml")
	if err != nil {
		log.Printf("load config failed: %v\n", err)
		return
	}
	appConfig = conf
	redisCli = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    conf.Redis.Addr,
		DB:       conf.Redis.Db,
		Password: conf.Redis.Password,
	})

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
