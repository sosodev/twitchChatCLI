package state

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/nsf/termbox-go"
)

// ChatLine is the data for a line of chat
// Implemented as a doubly-linked list
type ChatLine struct {
	Nick      string
	NickColor termbox.Attribute
	ShowNick  bool
	Line      string
}

const (
	maxLines = 500
)

var (
	lines      []ChatLine                   = make([]ChatLine, 0)
	userColors map[string]termbox.Attribute = make(map[string]termbox.Attribute)
	colorLock  sync.Mutex
	linesLock  sync.Mutex
)

func getRandomColor() termbox.Attribute {
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

// NickColor returns the randomly assigned color for a nick
func NickColor(nick string) termbox.Attribute {
	colorLock.Lock()
	defer colorLock.Unlock()

	if userColors[nick] == termbox.ColorDefault {
		userColors[nick] = getRandomColor()
	}

	return userColors[nick]
}

func newChatLine(nick string, showNick bool, line string) ChatLine {
	return ChatLine{
		Nick:      nick,
		NickColor: NickColor(nick),
		Line:      line,
		ShowNick:  showNick,
	}
}

// NewMessage stores an incoming message
func NewMessage(nick string, showNick bool, message string) {
	width, _ := termbox.Size()
	lineLimit := width - 2 - len(fmt.Sprintf("%s: ", nick))

	if len(message) > lineLimit {
		nuMessage := ""

		first := true
		for _, char := range message {
			nuMessage += string(char)

			if len(nuMessage) == lineLimit {
				NewMessage(nick, first, nuMessage)
				nuMessage = ""
				first = false
			}
		}

		if nuMessage != "" {
			NewMessage(nick, first, nuMessage)
		}
	} else {
		linesLock.Lock()
		defer linesLock.Unlock()

		if len(lines) == maxLines {
			lines = lines[1:]
		}

		lines = append(lines, newChatLine(nick, showNick, message))
	}
}

// ReverseEachLine iterates each of the stored chat lines in reverse order with a counter
func ReverseEachLine(callback func(int, ChatLine)) {
	linesLock.Lock()
	defer linesLock.Unlock()

	// i'm too lazy to do math here so I'm just going to keep a counter
	counter := 0
	for i := len(lines) - 1; i >= 0; i-- {
		callback(counter, lines[i])
		counter++
	}
}
