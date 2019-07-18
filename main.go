// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/hex"
	"fmt"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/jroimartin/gocui"
	"github.com/liyue201/lazyredis/config"
	"github.com/liyue201/lazyredis/redis"
	"jlog"
	"log"
	"strings"
)

var appConfig *config.AppConfig
var redisCli redis.Client

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

func replyText(reply interface{}) string {
	if n, err := redigo.Int(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Int64(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Uint64(reply, nil); err == nil {
		return fmt.Sprintf("%d", n)
	}
	if n, err := redigo.Float64(reply, nil); err == nil {
		return fmt.Sprintf("%f", n)
	}
	if str, err := redigo.String(reply, nil); err == nil {
		return str
	}
	if bys, err := redigo.Bytes(reply, nil); err == nil {
		return hex.EncodeToString(bys)
	}
	if b, err := redigo.Bool(reply, nil); err == nil {
		if b {
			return "true"
		}
		return "false"
	}
	if values, err := redigo.Values(reply, nil); err == nil {
		str := ""
		for _, v := range values {
			if str != "" {
				str += "\n"
			}
			str += replyText(v)
		}
		return str
	}
	if floats, err := redigo.Float64s(reply, nil); err == nil {
		str := ""
		for _, v := range floats {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s", v)
		}
		return str
	}

	if strs, err := redigo.Strings(reply, nil); err == nil {
		str := ""
		for _, v := range strs {
			if str != "" {
				str += "\n"
			}
			str += v
		}
		return str
	}

	if byteSlids, err := redigo.ByteSlices(reply, nil); err == nil {
		str := ""
		for _, v := range byteSlids {
			if str != "" {
				str += "\n"
			}
			str += hex.EncodeToString(v)
		}
		return str
	}

	if ints, err := redigo.Int64s(reply, nil); err == nil {
		str := ""
		for _, v := range ints {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%d", v)
		}
		return str
	}
	if ints, err := redigo.Ints(reply, nil); err == nil {
		str := ""
		for _, v := range ints {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%d", v)
		}
		return str
	}

	if mps, err := redigo.StringMap(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %s", k, v)
		}
		return str
	}

	if mps, err := redigo.IntMap(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %d", k, v)
		}
		return str
	}

	if mps, err := redigo.Int64Map(reply, nil); err == nil {
		str := ""
		for k, v := range mps {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%s %d", k, v)
		}
		return str
	}

	if pos, err := redigo.Positions(reply, nil); err == nil {
		str := ""
		for _, v := range pos {
			if str != "" {
				str += "\n"
			}
			str += fmt.Sprintf("%f %f", v[0], v[1])
		}
		return str
	}
	return ""
}

func exec(g *gocui.Gui, v *gocui.View) error {
	var line string
	var err error

	defer func() {
		_, cy := v.Cursor()
		if err := v.SetCursor(0, cy+1); err != nil {
			_, oy := v.Origin()
			if err := v.SetOrigin(0, oy+1); err != nil {
				return
			}
			v.SetCursor(0, cy)
		}
		newLine(v)
	}()

	_, cy := v.Cursor()
	if line, err = v.Line(cy); err != nil {
		line = ""
		return nil
	}
	jlog.Logger.Infof("line: %v", line)

	inputs := ""
	splits := strings.Split(line, ">")
	if len(splits) > 1 {
		inputs = splits[1]
	}
	if inputs == "" {
		return nil
	}

	cmd := ""
	args := []interface{}{}

	inputSprits := strings.Split(inputs, " ")
	for _, item := range inputSprits {
		if item == " " {
			continue
		}
		if cmd == "" {
			cmd = item
		} else {
			args = append(args, item)
		}
	}
	relpy, err := redisCli.Do(cmd, args...)
	mainView, _ := g.View("main")
	mainView.Clear()
	mainView.SetCursor(0, 0)
	if err == nil {
		str := strings.ReplaceAll(replyText(relpy), "\r\n", "\n")
		if str != "" {
			fmt.Fprintln(mainView, str)
		}else {
			fmt.Fprintln(mainView, "<nil>")
		}
		jlog.Logger.Info(relpy)
	} else {
		fmt.Fprintln(mainView, err.Error())
		jlog.Logger.Error(err.Error())
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, exec); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	keys := []gocui.Key{gocui.KeyArrowLeft, gocui.KeyArrowDown, gocui.KeyArrowRight, gocui.KeyArrowUp, gocui.KeyDelete, gocui.KeyBackspace2}
	for _, key := range keys {
		g.SetKeybinding("side", key, gocui.ModNone, IgnoreKey);
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
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
		v.Title = "output"
	}
	return nil
}

func test()  {
	reply, err :=  redisCli.Do("SET", "aa", "bbb")
	if err != nil{
		fmt.Printf("set: %s", err.Error())
		return
	}
	fmt.Printf("set aa bb: \n  %v\n", replyText(reply))

	reply, err =  redisCli.Do("get", "aa")
	if err != nil{
		fmt.Printf("%s", err.Error())
		return
	}
	fmt.Printf("get aa:\n  %v\n", replyText(reply))

	reply, err =  redisCli.Do("info")
	if err != nil{
		fmt.Printf("%s", err.Error())
		return
	}
	fmt.Printf("info:\n  %v\n", replyText(reply))
}

func main() {
	conf, err := config.Load("conf.yaml")
	if err != nil {
		panic(err.Error())
		return
	}
	appConfig = conf

	redisCli, err = redis.NewRedisClient(appConfig.Redis.Addr, appConfig.Redis.Password, appConfig.Redis.Db)
	if err != nil {
		panic(err.Error())
	}
	defer redisCli.Close()

	//test()
	//return

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err.Error())
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
