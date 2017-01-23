package birc

import (
	"time"

	sirc "github.com/sorcix/irc"
)

// Message is a decoded IRC message.
type Message struct {
	Name     string
	Username string
	Content  string
	Command  string
	Host     string
	Params   []string
	Time     time.Time
}

// Encode converts a Message struct into an IRC messsage
func (m *Message) prepare() *sirc.Message {
	message := &sirc.Message{
		Prefix: &sirc.Prefix{
			Name: m.Name,
			User: m.Username,
			Host: m.Host,
		},
		Command:  m.Command,
		Params:   m.Params,
		Trailing: m.Content,
	}

	if m.Name != "" || m.Username != "" || m.Host != "" {
		message.Prefix = &sirc.Prefix{}
	}
	if m.Name != "" {
		message.Name = m.Name
	}
	if m.Username != "" {
		message.User = m.Username
	}
	if m.Host != "" {
		message.Host = m.Host
	}

	return message
}

// PongMessage returns a Message struct containing a PONG message,
// which should be used as a response to Twitch's PING message
func PongMessage() *Message {
	return &Message{
		Command: "PONG",
		Content: "tmi.twitch.tv",
	}
}
