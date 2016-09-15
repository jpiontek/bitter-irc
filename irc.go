// Package birc is designed to interact with a Twitch IRC channel.
package birc

import (
	"fmt"
	"net"
	"time"

	sirc "github.com/sorcix/irc"
)

// DefaultTwitchServer is the primary Twitch.tv IRC server.
var DefaultTwitchServer = "irc.chat.twitch.tv:6667"

// Encoder represents a struct capable of encoding an IRC message.
type Encoder interface {
	Encode(m *sirc.Message) error
}

// Decoder represents a struct capable of decoding incoming IRC messages.
type Decoder interface {
	Decode() (*sirc.Message, error)
}

// Config contains fields required to connect to the IRC server.
type Config struct {
	ChannelName string
	Server      string
	Username    string
	OAuthToken  string
}

// Channel represents a connected and active IRC channel.
type Channel struct {
	Config     *Config
	Connection net.Conn
	Digesters  []Digester
	reader     Decoder
	writer     Encoder
	data       chan Message
}

// Message is a decoded IRC message.
type Message struct {
	Username string
	Content  string
}

// NewTwitchChannel creates an IRC channel with Twitch's default server and port.
func NewTwitchChannel(channelName, username string, token string, digesters ...Digester) (*Channel, error) {
	config := &Config{
		ChannelName: channelName,
		Username:    username,
		OAuthToken:  token,
		Server:      DefaultTwitchServer,
	}

	return Connect(config, digesters[:]...)
}

// Connect establishes a connection to an IRC server.
func Connect(c *Config, digesters ...Digester) (*Channel, error) {
	conn, err := net.Dial("tcp", c.Server)
	if err != nil {
		return nil, err
	}

	channel := &Channel{Config: c, Connection: conn, Digesters: digesters}
	channel.reader = sirc.NewDecoder(conn)
	channel.writer = sirc.NewEncoder(conn)

	channel.data = make(chan Message)
	for _, digester := range channel.Digesters {
		go digester(channel.data)
	}

	return channel, nil
}

// SetWriter sets the channel's underlying writer. This is not threadsafe.
func (c *Channel) SetWriter(e Encoder) {
	c.writer = e
}

// Authenticate send the PASS and NICK to authenticate against the server. It also sends
// the JOIN message in order to join the specified channel in the configuration.
func (c *Channel) Authenticate() error {
	for _, m := range []*sirc.Message{
		&sirc.Message{
			Command: sirc.PASS,
			Params:  []string{fmt.Sprintf("oauth:%s", c.Config.OAuthToken)},
		},
		&sirc.Message{
			Command: sirc.NICK,
			Params:  []string{c.Config.Username},
		},
		&sirc.Message{
			Command: sirc.JOIN,
			Params:  []string{fmt.Sprintf("#%s", c.Config.ChannelName)},
		},
	} {
		if err := c.writer.Encode(m); err != nil {
			return err
		}
	}
	return nil
}

// Listen enters a loop and starts decoding IRC messages from the connected channel.
// Decoded messages are pushed to the data channel.
func (c *Channel) Listen(done <-chan int) error {
	for {
		c.Connection.SetDeadline(time.Now().Add(120 * time.Second))
		select {
		case <-done:
			c.Connection.Close()
			return nil
		default:
			m, err := c.reader.Decode()
			if err != nil {
				return err
			}
			if m.Prefix != nil {
				message := Message{Username: m.User, Content: m.Trailing}
				c.data <- message
			}
		}
	}
}
