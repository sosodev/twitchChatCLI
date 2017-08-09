package main

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	irc "github.com/fluffle/goirc/client"
	//"os"
	"crypto/tls"
	"github.com/simplyserenity/twitchOAuth"
	"net/http"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"strings"
	"os"
)

const twitch = "irc.chat.twitch.tv:443"
const clientID = "dlpf1993tub698zw0ic6jlddt9e893"

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	token := twitchAuth.GetToken(clientID)

	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	messages := make([]string, 0)

	uInput := ""

	cfg := irc.NewConfig(getUsername(token, clientID))
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.chat.twitch.tv"}
	cfg.Server = twitch
	cfg.Pass = "oauth:" + token

	channel := "#" + strings.ToLower(os.Args[1])

	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		conn.Join(channel)
	})

	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		messages = newMessage(messages, line.Nick + ": " + line.Args[1])
		redraw_all(messages, uInput)
	})

	if err := c.Connect(); err != nil {
		messages = newMessage(messages, err.Error())
		redraw_all(messages, uInput)
	}

chat_loop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				break chat_loop
			case termbox.KeySpace:
				uInput += " "
			case termbox.KeyEnter:
				if len(uInput) > 2 {
					messages = newMessage(messages, "You: " + uInput)
					c.Privmsg(channel, uInput)
					uInput = ""
				}
			case termbox.KeyBackspace2:
				if len(uInput) > 0{
					uInput = uInput[0:len(uInput)-1]
				}
			case termbox.KeyBackspace:
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
		redraw_all(messages, uInput)
	}
}

func getUsername(token string, cId string) (username string){
	rclient := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/kraken/user", nil)
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", cId)
	req.Header.Set("Authorization", "OAuth " + token)

	res, err := rclient.Do(req)

	defer res.Body.Close()

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err)
	}

	name, err :=jsonparser.GetString(body, "display_name")

	if err != nil {
		panic(err)
	}

	return strings.ToLower(name)
}

func newMessage(messages []string, text string) ([]string){
	_, h := termbox.Size()
	if len(messages) == h - 2 {
		messages = messages[1:]
	}
	return append(messages, text)
}

func redraw_all(messages []string, uInput string) {
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	tbprint(w - 18, 0, termbox.ColorWhite, coldef, "Press ESC to quit")

	pos := 0

	for i := len(messages) - 1; i >= 0; i-- {
		tbprint(1, h - pos - 2, termbox.ColorWhite, coldef, messages[i])
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