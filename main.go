package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/jroimartin/gocui"
)

var (
	hour                    int
	hours, minutes, seconds bool
	done                    = make(chan struct{})
)

func main() {
	hour = 17

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.FgColor = gocui.ColorRed

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	go counter(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("ctr", maxX/2-15, maxY/2-3, maxX/2+16, maxY/2+3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		printCounter(g, counterString())
	}
	return nil
}

func counterString() string {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())

	var h, m int
	if time.Now().Hour() >= hour {
		h = 0
		m = 0
	} else {
		h = int(math.Floor(time.Until(t).Hours()))
		m = int(math.Ceil((time.Until(t).Hours() - math.Floor(time.Until(t).Hours())) * 60))
	}
	if m == 60 {
		m = 0
	}

	return fmt.Sprintf("%d:%02d", h, m)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}

func counter(g *gocui.Gui) {
	for {
		select {
		case <-done:
			return
		case <-time.After(500 * time.Millisecond):
			g.Update(func(g *gocui.Gui) error {
				return printCounter(g, counterString())
			})
		}
	}
}

func printCounter(g *gocui.Gui, toPrint string) error {
	v, err := g.View("ctr")
	if err != nil {
		return err
	}

	v.Clear()

	for i, c := range []rune(toPrint) {
		for row := range chars[c] {
			for col := range chars[c][row] {
				v.SetCursor(col+(i*(charWidth+1)), row)
				if chars[c][row][col] == 1 {
					v.EditWrite(block)
				}
			}
		}
	}

	return nil
}
