package gui
 
import (
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/liyue201/lazyredis/config"
	"github.com/liyue201/lazyredis/redis"
	"log"
	"strings"
)
 
const (
	topView    = "top"
	midleView  = "midle"
	BottomView = "bottom"
)

type Window struct {
	g        *gocui.Gui
	redisCli redis.Client
	history  History
}

func NewWindow() *Window {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err.Error())
	}
	w := &Window{g: g}

	g.Cursor = true
	g.Highlight = true
	g.SelFgColor = gocui.ColorGreen
	g.SetManager(w)
	w.keybindings()

	return w
}

func (w *Window) SetRedisClient(cli redis.Client) {
	w.redisCli = cli
}

func (w *Window) Close() {
	w.g.Close()
}

func (w *Window) Run() {
	if err := w.g.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Println(err)
	}
}

func (w *Window) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView(topView, 1, 1, maxX-1, 3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "db"
		fmt.Fprintln(v, fmt.Sprintf("%v", config.Conf.Redis.Addr))
	}

	if v, err := g.SetView(midleView, 1, 4, maxX-1, 12); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "cmd"
		v.Editable = true
		w.writeArrow(v)

		if _, err = g.SetCurrentView(midleView); err != nil {
			return err
		}
	}
	if v, err := g.SetView(BottomView, 1, 13, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Wrap = true
		v.Title = "output"
	}
	return nil
}

func (w *Window) keybindings() error {
	if err := w.g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, w.nextView); err != nil {
		log.Panicln(err)
	}
	if err := w.g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, w.quit); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(midleView, gocui.KeyEnter, gocui.ModNone, w.execCmd); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(midleView, gocui.KeyArrowUp, gocui.ModNone, w.preCmd); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(midleView, gocui.KeyArrowDown, gocui.ModNone, w.nextCmd); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(midleView, gocui.KeyBackspace, gocui.ModNone, w.onKeyBackspace); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(midleView, gocui.KeyBackspace2, gocui.ModNone, w.onKeyBackspace); err != nil {
		return err
	}

	if err := w.g.SetKeybinding(midleView, gocui.KeyArrowLeft, gocui.ModNone, w.onKeyArrowLeft); err != nil {
		return err
	}

	if err := w.g.SetKeybinding(midleView, gocui.KeyArrowRight, gocui.ModNone, w.onKeyArrowRight); err != nil {
		return err
	}

	keys := []gocui.Key{gocui.KeyDelete }
	for _, key := range keys {
		w.g.SetKeybinding(midleView, key, gocui.ModNone, w.ignoreKey);
	}
	if err := w.g.SetKeybinding(BottomView, gocui.KeyArrowDown, gocui.ModNone, w.cursorDown); err != nil {
		return err
	}
	if err := w.g.SetKeybinding(BottomView, gocui.KeyArrowUp, gocui.ModNone, w.cursorUp); err != nil {
		return err
	}
	return nil
}

func (w *Window) nextView(g *gocui.Gui, v *gocui.View) error {
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

	return nil
}

func (w *Window) execCmd(g *gocui.Gui, v *gocui.View) error {
	var line string
	var err error

	defer func() {
		v.EditNewLine()
		w.writeArrow(v)
	}()

	_, cy := v.Cursor()
	if line, err = v.Line(cy); err != nil {
		line = ""
		return nil
	}
	v.SetCursor(len(line), cy)

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
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if cmd == "" {
			cmd = item
		} else {
			args = append(args, item)
		}
	}

	mainView, _ := g.View(BottomView)
	mainView.Clear()
	mainView.SetCursor(0, 0)

	if w.redisCli == nil {
		fmt.Fprintln(mainView, "redis disconnected")
		return nil
	}

	relpy, err := w.redisCli.Do(cmd, args...)
	if err == nil {
		str := strings.ReplaceAll(relpy, "\r\n", "\n")
		if str != "" {
			fmt.Fprintln(mainView, str)
		} else {
			fmt.Fprintln(mainView, "<nil>")
		}
	} else {
		fmt.Fprintln(mainView, err.Error())
	}
	w.history.Add(inputs)
	return nil
}

func (w *Window) quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (w *Window) cursorDown(g *gocui.Gui, v *gocui.View) error {
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

func (w *Window) cursorUp(g *gocui.Gui, v *gocui.View) error {
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

func (w *Window) nextCmd(g *gocui.Gui, v *gocui.View) error {
	nextCmd := w.history.Next()
	if nextCmd == "" {
		return  nil
	}
	w.deleteCurrentLine(g, v)
	w.writeArrow(v)
	w.writeText(v, nextCmd)
	return nil
}

func (w *Window) preCmd(g *gocui.Gui, v *gocui.View) error {
	preCmd := w.history.Prev()
	if preCmd == "" {
		return  nil
	}
	w.deleteCurrentLine(g, v)
	w.writeArrow(v)
	w.writeText(v, preCmd)

	return nil
}

func (w *Window) onKeyBackspace(g *gocui.Gui, v *gocui.View) error {
	x, _ := v.Cursor()
	if x  > 1 {
		v.EditDelete(true)
	}
	return nil
}

func (w *Window) onKeyArrowLeft(g *gocui.Gui, v *gocui.View) error {
	x, _ := v.Cursor()
	if x  > 1 {
		v.MoveCursor(-1, 0, true)
	}
	return nil
}

func (w *Window) onKeyArrowRight(g *gocui.Gui, v *gocui.View) error {
	v.MoveCursor(1, 0, true)
	return nil
}

func (w *Window) writeArrow(v *gocui.View) {
	line := fmt.Sprintf(">")
	w.writeText(v, line)
}

func (w *Window) writeText(v *gocui.View, str string) {
	runes := []rune(str)
	for _, c := range runes {
		v.EditWrite(c)
	}
}

func (w *Window) ignoreKey(g *gocui.Gui, v *gocui.View) error {
	//do nothing
	return nil
}

func (w *Window) deleteCurrentLine(g *gocui.Gui, v *gocui.View) error {
	_, y := v.Cursor()
	line, _ := v.Line(y)
	n := len(line)
	v.SetCursor(n, y)

	for i := 0; i < n; i++ {
		v.EditDelete(true)
	}
	return nil
}