package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
)

var (
	width, height           int
	year, day, hour, minute int
	month                   time.Month
	seconds                 bool
	done                    = make(chan struct{})
	logger                  *log.Logger
	color                   string
)

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("ctr", maxX/2-int(math.Ceil(float64(width)/2)), maxY/2-int(math.Ceil(float64(height)/2)), maxX/2+int(math.Ceil(float64(width)/2)), maxY/2+int(math.Ceil(float64(height)/2))); err != nil {
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
	t := time.Date(year, month, day, hour, minute, 0, 0, now.Location())

	var h, m, s int
	if time.Until(t).Seconds() > 0 {
		h = int(math.Floor(time.Until(t).Hours()))
		if seconds {
			m = int((time.Until(t).Hours() - math.Floor(time.Until(t).Hours())) * 60)
			s = int((time.Until(t).Minutes() - math.Floor(time.Until(t).Minutes())) * 60)
		} else {
			m = int(math.Ceil((time.Until(t).Hours() - math.Floor(time.Until(t).Hours())) * 60))
			if m == 60 {
				m = 0
			}
		}
	}

	if seconds {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
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

func init() {
	var endTime string
	flag.StringVar(&color, "c", "red", "color of the countdown timer")
	flag.BoolVar(&seconds, "s", false, "display countdown timer with seconds")
	flag.StringVar(&endTime, "t", "5:00pm", "time to count down to in format: HH:MM(am|pm) (eg. \"5:00pm\")")
	flag.Parse()

	logger = log.New(os.Stdout, "deathclock:", log.LstdFlags)

	if len(endTime) < 6 || len(endTime) > 7 {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		os.Exit(1)
	}
	period := endTime[len(endTime)-2:]
	period = strings.ToLower(period)
	if period != "pm" && period != "am" {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		os.Exit(1)
	}
	t := strings.Split(endTime[:len(endTime)-2], ":")
	if len(t) != 2 {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		os.Exit(1)
	}
	hs := t[0]
	ms := t[1]
	h, err := strconv.Atoi(hs)
	if err != nil {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		logger.Printf("%+v", errors.WithStack(err))
		os.Exit(1)
	}
	if h < 1 || h > 12 {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		os.Exit(1)
	}
	if period == "pm" {
		h += 12
	}
	m, err := strconv.Atoi(ms)
	if err != nil {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		logger.Printf("%+v", errors.WithStack(err))
		os.Exit(1)
	}
	if m < 0 || m > 59 {
		logger.Println("t string should be in format: HH:MM(am|pm) (eg. \"5:00pm\")")
		os.Exit(1)
	}

	now := time.Now()
	parsedTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location())

	if int64(time.Until(parsedTime)) <= int64(0) {
		parsedTime = parsedTime.AddDate(0, 0, 1)
	}

	logger.Println(time.Until(parsedTime))

	year = parsedTime.Year()
	month = parsedTime.Month()
	day = parsedTime.Day()
	hour = parsedTime.Hour()
	minute = parsedTime.Minute()

	logger.Println(now.Year(), now.Month(), now.Day(), h, m)
	logger.Println(year, month, day, hour, minute)
}

func main() {
	if seconds {
		width = 49
	} else {
		width = 31
	}
	height = 5

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logger.Panicln(err)
	}
	defer g.Close()

	switch color {
	case "black":
		g.FgColor = gocui.ColorBlack
	case "red":
		g.FgColor = gocui.ColorRed
	case "green":
		g.FgColor = gocui.ColorGreen
	case "yellow":
		g.FgColor = gocui.ColorYellow
	case "blue":
		g.FgColor = gocui.ColorBlue
	case "magenta":
		g.FgColor = gocui.ColorMagenta
	case "cyan":
		g.FgColor = gocui.ColorCyan
	case "white":
		g.FgColor = gocui.ColorWhite
	default:
		g.FgColor = gocui.ColorRed
	}

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		logger.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		logger.Panicln(err)
	}

	go counter(g)

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		logger.Panicln(err)
	}
}
