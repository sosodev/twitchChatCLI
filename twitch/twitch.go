package twitch

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	irc "github.com/fluffle/goirc/client"
	twitchoauth "github.com/sosodev/twitchOAuth"
)

const (
	ircEndpoint  = "irc.chat.twitch.tv"
	ircPort      = "443"
	userEndpoint = "https://api.twitch.tv/kraken/user"
	clientID     = "dlpf1993tub698zw0ic6jlddt9e893"
)

// Twitch interface lets you interact with stuff
type Twitch struct {
	token string
}

// New get you a new *Twitch or error
func New() (*Twitch, error) {
	requiredScopes := []string{"chat_login", "user_read"}

	token, err := twitchoauth.GetToken(clientID, requiredScopes)

	return &Twitch{
		token,
	}, err
}

// Username returns the user's username or error
func (t *Twitch) Username() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", userEndpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Authorization", "OAuth "+t.token)

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return jsonparser.GetString(body, "display_name")
}

// IrcServer returns the irc server in the form "endpoint:port"
func IrcServer() string {
	return fmt.Sprintf("%s:%s", ircEndpoint, ircPort)
}

// IrcConfig returns the irc config
func (t *Twitch) IrcConfig() (*irc.Config, error) {
	username, err := t.Username()
	if err != nil {
		return nil, err
	}

	cfg := irc.NewConfig(strings.ToLower(username))
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{ServerName: ircEndpoint}
	cfg.Server = IrcServer()
	cfg.Pass = "oauth:" + t.token

	return cfg, nil
}
