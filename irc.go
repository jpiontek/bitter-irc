// Package birc is designed to interact with a Twitch IRC channel.
package birc

import (
	"fmt"
	"net"
	"time"

	sirc "github.com/sorcix/irc"
)

// DefaultTwitchURI is the primary Twitch.tv IRC server.
var DefaultTwitchURI = "irc.chat.twitch.tv"

// DefaultTwitchPort is the primary Twitch.tv IRC server's port.
var DefaultTwitchPort = "6667"

// DefaultTwitchServer is the primary Twitch.tv IRC server including the PORT.
var DefaultTwitchServer = DefaultTwitchURI + ":" + DefaultTwitchPort

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
	done       chan int
}

// ChannelWriter represents a writer capable of sending messages to a channel.
type ChannelWriter interface {
	SimpleMessage(message string) error
}

// Message is a decoded IRC message.
type Message struct {
	Username string
	Content  string
}

// NewTwitchChannel creates an IRC channel with Twitch's default server and port.
func NewTwitchChannel(channelName, username, token string, digesters ...Digester) (*Channel, error) {
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
	channel.done = make(chan int)

	return channel, nil
}

// SetWriter sets the channel's underlying writer. This is not threadsafe.
func (c *Channel) SetWriter(e Encoder) {
	c.writer = e
}

// Authenticate sends the PASS and NICK to authenticate against the server. It also sends
// the JOIN message in order to join the specified channel in the configuration.
func (c *Channel) Authenticate() error {
	for _, m := range []sirc.Message{
		sirc.Message{
			Command: sirc.PASS,
			Params:  []string{fmt.Sprintf("oauth:%s", c.Config.OAuthToken)},
		},
		sirc.Message{
			Command: sirc.NICK,
			Params:  []string{c.Config.Username},
		},
		sirc.Message{
			Command: sirc.JOIN,
			Params:  []string{fmt.Sprintf("#%s", c.Config.ChannelName)},
		},
	} {
		if err := c.writer.Encode(&m); err != nil {
			return err
		}
	}
	return nil
}

// Listen enters a loop and starts decoding IRC messages from the connected channel.
// Decoded messages are pushed to the data channel.
func (c *Channel) Listen() error {
	// Close the connection when finished.
	defer c.Connection.Close()
	for {
		c.Connection.SetDeadline(time.Now().Add(120 * time.Second))
		select {
		case <-c.done:
			return nil
		default:
			m, err := c.reader.Decode()
			if err != nil {
				return err
			}
			if m.Prefix != nil {
				message := Message{Username: m.User, Content: m.Trailing}
				go c.handle(&message)
			}
		}
	}
}

func (c *Channel) handle(m *Message) {
	for _, d := range c.Digesters {
		go d(*m, c)
	}
}

// Close ends the current listener and closes the TCP connection.
func (c *Channel) Close() {
	c.done <- 1
}

// SimpleMessage writes a message to the channel.
func (c *Channel) SimpleMessage(message string) error {
	m := &sirc.Message{
		Prefix: &sirc.Prefix{
			Name: c.Config.Username,
			User: c.Config.Username,
			Host: DefaultTwitchURI,
		},
		Command:  sirc.PRIVMSG,
		Params:   []string{fmt.Sprintf("#%s", c.Config.Username)},
		Trailing: message,
	}
	if err := c.writer.Encode(m); err != nil {
		return err
	}
	return nil
}
