package main

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
	"log"
	"math/rand"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/nsf/termbox-go"
	"github.com/sosodev/twitchChatCLI/state"
	"github.com/sosodev/twitchChatCLI/twitch"

	"os"
	"strings"
)

var (
	userInput string
)

func main() {
	// setup the debugging log if requested
	if os.Getenv("debug") == "true" {
		logFile, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file: %s", err))
		}
		defer logFile.Close()

		log.SetOutput(logFile)
	}

	// seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// grab the channel name from the input args
	if len(os.Args) < 2 {
		fmt.Println("usage: twitchChatCli [channel-name]")
		os.Exit(1)
	}
	channel := "#" + strings.ToLower(os.Args[1])

	// initalize termbox (the thing that renders stuff in the terminal)
	if err := termbox.Init(); err != nil {
		panic(fmt.Sprintf("failed to initialize termbox: %s", err))
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	// creating the twitch object actually handles the Twitch authentication stuff
	twitch, err := twitch.New()
	if err != nil {
		panic(fmt.Sprintf("failed to get authorization from Twitch: %s", err))
	}

	username, err := twitch.Username()
	if err != nil {
		panic(fmt.Sprintf("failed to get username from Twitch: %s", err))
	}
	username = strings.ToLower(username)

	// get the IRC config
	cfg, err := twitch.IrcConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize IRC configuration: %s", err))
	}

	// create the IRC client
	client := irc.Client(cfg)
	client.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		conn.Join(channel)
	})

	// tbPrint prints a string with termbox
	tbPrint := func(x int, y int, foreground termbox.Attribute, background termbox.Attribute, message string) {
		condition := runewidth.NewCondition()
		condition.StrictEmojiNeutral = false

		graphemes := uniseg.NewGraphemes(message)
		for graphemes.Next() {
			for _, graphemeRune := range graphemes.Runes() {
				termbox.SetCell(x, y, graphemeRune, foreground, background)
				w := runewidth.RuneWidth(graphemeRune)
				if w == 0 || (w == 2 && runewidth.IsAmbiguousWidth(graphemeRune)) {
					w = 1
				}
				x += w
			}
		}
	}

	// drawLine draws a state.ChatLine on the given y line
	drawLine := func(line state.ChatLine, y int) {
		// the nick portion of the line can either be "nick: " or indented "     "
		nick := func() string {
			if line.ShowNick {
				return fmt.Sprintf("%s: ", line.Nick)
			}

			emptySpace := "  "
			for range line.Nick {
				emptySpace += " "
			}

			return emptySpace
		}()

		// draw the nick portion of the line
		tbPrint(1, y, line.NickColor, termbox.ColorDefault, nick)

		var foreground termbox.Attribute
		background := termbox.ColorDefault
		if strings.Contains(line.Line, "@"+username) {
			foreground = termbox.ColorBlack
			background = termbox.ColorWhite
		} else {
			foreground = termbox.ColorWhite
		}

		// draw the body of the line
		tbPrint(len(nick)+1, y, foreground, background, line.Line)
	}

	// drawClient is a function that... draws the client
	drawClient := func() {
		// clear the terminal
		err = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		if err != nil {
			panic(fmt.Sprintf("failed to clear termbox buffer: %s\n", err))
		}
		err = termbox.Flush()
		if err != nil {
			panic(fmt.Sprintf("failed to flush termbox buffer: %s\n", err))
		}

		width, height := termbox.Size()

		// draw the escape message
		if os.Getenv("CLEAN") != "true" {
			tbPrint(width-18, 0, termbox.ColorLightBlue, termbox.ColorDefault, "Press ESC to quit")
		}

		// iterate through all of the lines in reverse (to draw them newest to oldest)
		state.ReverseEachLine(func(position int, line state.ChatLine) {
			subtract := 2
			if os.Getenv("CLEAN") == "true" {
				subtract = 1
			}

			// no pointless iteration here
			if position > height-subtract {
				return
			}

			drawLine(line, height-subtract-position)
		})

		// draw the user's input at the bottom of the terminal
		if os.Getenv("CLEAN") != "true" {
			tbPrint(1, height-1, state.NickColor(username), termbox.ColorDefault, fmt.Sprintf("%s: %s", username, userInput))
		}

		// flush the termbox buffer to the terminal
		err = termbox.Flush()
		if err != nil {
			panic(fmt.Sprintf("failed to flush termbox: %s\n", err))
		}
	}

	// handle incoming messages by adding them to the state and redrawing the client
	client.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		state.NewMessage(line.Nick, true, line.Args[1])
		drawClient()
	})

	// startup ye olde IRC client
	if err := client.Connect(); err != nil {
		panic(fmt.Sprintf("failed to connect to irc: %s", err))
	}

	// do an inital drawing
	drawClient()

	// big event handler loop
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				os.Exit(0)
			case termbox.KeySpace:
				userInput += " "
			case termbox.KeyEnter:
				if len(userInput) > 2 {
					state.NewMessage(username, true, userInput)
					client.Privmsg(channel, userInput)
					userInput = ""
				}
			case termbox.KeyBackspace2:
				if len(userInput) > 0 {
					userInput = userInput[0 : len(userInput)-1]
				}
			case termbox.KeyBackspace:
				if len(userInput) > 0 {
					userInput = userInput[0 : len(userInput)-1]
				}
			default:
				if ev.Ch != 0 {
					userInput += string(ev.Ch)
				}
			}
		case termbox.EventError:
			panic(fmt.Sprintf("termbox event error: %s", ev.Err.Error()))
		}

		// draw the client after every handled event
		drawClient()
	}
}
