package birc

import (
	"crypto/tls"
	"fmt"
	"net"
	"time"

	sirc "github.com/sorcix/irc"
)

// Config contains fields required to connect to the IRC server.
type Config struct {
	ChannelName string
	Server      string
	Username    string
	OAuthToken  string
	tls         bool
}

// Channel represents a connected and active IRC channel.
type Channel struct {
	Config     *Config
	Digesters  []Digester
	connection net.Conn
	reader     Decoder
	writer     Encoder
	done       chan error
}

// ChannelWriter represents a writer capable of sending messages to a channel.
type ChannelWriter interface {
	Send(content string) error
	SendMessage(message *Message) error
	GetConfig() Config
}

func (c *Channel) GetConfig() Config {
	return *c.Config
}

// NewTwitchChannel creates an IRC channel with Twitch's default server and port.
func NewTwitchChannel(channelName, username, token string, tls bool, digesters ...Digester) *Channel {
	config := &Config{
		ChannelName: channelName,
		Username:    username,
		OAuthToken:  token,
		Server:      DefaultTwitchServer,
		tls:         tls,
	}

	if tls {
		config.Server = DefaultTwitchTlsServer
	}

	return &Channel{Config: config, Digesters: digesters[:]}
}

// Connect establishes a connection to an IRC server.
func (c *Channel) Connect() error {
	var err error
	var conn net.Conn
	if c.Config.tls {
		conn, err = tls.Dial("tcp", c.Config.Server, nil)
	} else {
		conn, err = net.Dial("tcp", c.Config.Server)
	}

	if err != nil {
		return err
	}

	c.connection = conn
	c.reader = sirc.NewDecoder(conn)
	c.writer = sirc.NewEncoder(conn)
	if c.done == nil {
		c.done = make(chan error)
	}
	return nil
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
		// Twitch specific capability registration
		sirc.Message{
			Command: "CAP REQ",
			Params:  []string{":twitch.tv/commands"},
		},
		//sirc.Message{
		//Command: "CAP REQ",
		//Params:  []string{":twitch.tv/tags"},
		//},
	} {
		if err := c.writer.Encode(&m); err != nil {
			return err
		}
	}
	return nil
}

// Disconnect ends the current listener and closes the TCP connection.
func (c *Channel) Disconnect() {
	c.done <- nil
}

// Send writes a message to the channel.
func (c *Channel) Send(content string) error {
	return c.SendMessage(&Message{
		Name:     c.Config.Username,
		Username: c.Config.Username,
		Content:  content,
		Command:  sirc.PRIVMSG,
		Params:   []string{fmt.Sprintf("#%s", c.Config.ChannelName)},
	})
}

func (c *Channel) SendMessage(message *Message) error {
	if err := c.writer.Encode(message.prepare()); err != nil {
		return err
	}
	return nil
}

// Listen enters a loop and starts decoding IRC messages from the connected channel.
// Decoded messages are pushed to the digesters to be handled.
func (c *Channel) Listen() error {
	// Close the connection when finished.
	defer c.connection.Close()

	return c.startReceiving()
}

func (c *Channel) startReceiving() error {
	for {
		select {
		case <-c.done:
			return nil
		default:
			c.connection.SetDeadline(time.Now().Add(10 * time.Minute))
			m, err := c.reader.Decode()
			if err != nil {
				return err
			}
			// If the message is a PING command from Twitch, respond with a PONG
			// without pushing the message through to the digesters
			if m.Command == "PING" {
				c.SendMessage(PongMessage())
				break
			}

			// Handle Twitch restarting their IRC servers.
			if m.Command == "RECONNECT" {
				err := c.Reconnect()
				if err != nil {
					return err
				}
				break
			}

			message := &Message{
				Content: m.Trailing,
				Command: m.Command,
				Params:  m.Params,
				Time:    time.Now(),
			}
			if m.Prefix != nil {
				message.Name = m.Name
				message.Username = m.User
				message.Content = m.Trailing
				message.Host = m.Host
			}
			c.handle(message)
		}
	}
}

func (c *Channel) Reconnect() error {
	err := c.Connect()
	if err != nil {
		return err
	}

	err = c.Authenticate()
	if err != nil {
		return err
	}

	return nil
}

func (c *Channel) handle(m *Message) {
	for _, d := range c.Digesters {
		go d(*m, c)
	}
}
