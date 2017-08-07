package main

import (
	irc "github.com/fluffle/goirc/client"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	//"bufio"
	//"os"
	"fmt"
	//"strings"
)

const token = "your ouath token here"
const twitch = "irc.chat.twitch.tv:6667"

func main() {

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	queue := make([]string, 0)

	uInput := ""

	redraw_all(queue, uInput)

	cfg := irc.NewConfig("your username here")
	cfg.SSL = false
	cfg.Server = twitch
	cfg.Pass = token
	c := irc.Client(cfg)
	c.EnableStateTracking()

	channel := "#vinesauce"

	//channel := os.Args[1]
	//println(channel)

	c.HandleFunc(irc.CONNECTED , func(conn *irc.Conn, line *irc.Line) {
		conn.Join(channel)
	})

	quit := make(chan bool)

	c.HandleFunc(irc.DISCONNECTED, func(conn *irc.Conn, line *irc.Line) {
		quit <- true
	})

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		queue = addToQueue(queue, line.Nick + ": " + line.Args[1])
	})


	if err := c.ConnectTo(twitch); err != nil {
		queue = addToQueue(queue, fmt.Sprintf("Connection error: %s\n", err))

	}

	chat_loop:
		for {

			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key{
				case termbox.KeyEsc:
					break chat_loop
				case termbox.KeySpace:
					uInput += " "
				case termbox.KeyEnter:
					queue = addToQueue(queue, "You: " + uInput)
					c.Privmsg(channel, uInput)
					uInput = ""
				case termbox.KeyBackspace:
					if len(uInput) > 0{
						uInput = uInput[0:len(uInput)-1]
					}
				case termbox.KeyBackspace2:
					if len(uInput) > 0{
						uInput = uInput[0:len(uInput)-1]
					}
				default:
					if ev.Ch != 0 {
						uInput += string(ev.Ch)
					}
				}
			case termbox.EventError:
				panic(ev.Err)
			}

			redraw_all(queue, uInput)
		}
}



//it's actually a stack...
func addToQueue(queue []string, text string)([]string){
	_, h := termbox.Size()
	if len(queue) == h - 2{
		queue = queue[1:]
	}
	return append(queue, text)
}

func redraw_all(queue []string, uInput string){
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	tbprint(w - 18, 0, termbox.ColorWhite, coldef, "Press ESC to quit")

	pos := 0

	for i := len(queue) - 1; i >= 0; i--{
		tbprint(1, h - pos - 2, termbox.ColorWhite, coldef, queue[i])
		pos++
	}

	tbprint(1, h - 1, termbox.ColorWhite, coldef, "You: " + uInput)
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}