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
	"math/rand"
)

const twitch = "irc.chat.twitch.tv:443"
const clientID = "dlpf1993tub698zw0ic6jlddt9e893"

type ChatLine struct {
	Nick string
	NickColor termbox.Attribute
	Line string
}

func main() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}

	scopes := []string{"chat_login", "user_read"}

	token, err := twitchAuth.GetToken(clientID, scopes)

	if err != nil {
		panic(err)
	}

	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	messages := make([]ChatLine, 0)

	uInput := ""
	username := getUsername(token, clientID)
	cfg := irc.NewConfig(strings.ToLower(username))
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: "irc.chat.twitch.tv"}
	cfg.Server = twitch
	cfg.Pass = "oauth:" + token

	channel := "#" + strings.ToLower(os.Args[1])

	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED, func(conn *irc.Conn, line *irc.Line) {
		conn.Join(channel)
	})


	userColors := make(map[string]termbox.Attribute)


	c.HandleFunc(irc.PRIVMSG, func(conn *irc.Conn, line *irc.Line) {
		w, _ := termbox.Size()

		toPrint := line.Args[1]

		if userColors[line.Nick] == termbox.ColorDefault {
			userColors[line.Nick] = getRandomColor()
		}

		message := ChatLine{line.Nick, userColors[line.Nick], ""}

		for _, r := range toPrint {
			chr := string(r)

			if chr == "\n"{
				messages = newMessage(messages, message)
				message = ChatLine{line.Nick, userColors[line.Nick], ""}
			} else if len(message.Line) + 1 <= int(float64(w) * 0.8) || chr != " "{
				message.Line += chr
			} else {
				messages = newMessage(messages, message)
				message = ChatLine{"", userColors[line.Nick], ""}
			}
		}
		messages = newMessage(messages, message)
		redraw_all(messages, uInput, username)
	})

	if err := c.Connect(); err != nil {
		panic(err)
	}

	redraw_all(messages, uInput, username)

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
					messages = newMessage(messages, ChatLine{username, termbox.ColorGreen, uInput})
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
		redraw_all(messages, uInput, username)
	}
}

func getRandomColor()(termbox.Attribute){
	rng := rand.Intn(5)
	switch rng {
	case 0:
		return termbox.ColorBlue
	case 1:
		return termbox.ColorRed
	case 2:
		return termbox.ColorGreen
	case 3:
		return termbox.ColorCyan
	case 4:
		return termbox.ColorMagenta
	case 5:
		return termbox.ColorYellow
	default:
		return termbox.ColorCyan
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

	return name
}

func newMessage(messages []ChatLine, message ChatLine) ([]ChatLine){
	_, h := termbox.Size()
	if len(messages) == h - 2 {
		messages = messages[1:]
	}
	return append(messages, message)
}

func redraw_all(messages []ChatLine, uInput string, username string) {
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	w, h := termbox.Size()
	tbprint(w - 18, 0, termbox.ColorRed, coldef, "Press ESC to quit")

	pos := 0

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Nick != ""{
			tbprint(1, h - pos - 2, messages[i].NickColor, coldef, messages[i].Nick + ": ")

			if strings.Contains(messages[i].Line, "@" + username){
				tbprint(1 + len(messages[i].Nick + ": "), h - pos - 2, termbox.ColorBlack, termbox.ColorWhite, messages[i].Line)
			} else{
				tbprint(1 + len(messages[i].Nick + ": "), h - pos - 2, termbox.ColorWhite, termbox.ColorDefault, messages[i].Line)
			}

		} else {
			tbprint(1, h - pos - 2, termbox.ColorWhite, coldef, messages[i].Line)
		}
		pos++
	}

	tbprint(1, h - 1, termbox.ColorGreen, coldef, username + ": " + uInput)
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}