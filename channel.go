package birc

import (
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
}

// Channel represents a connected and active IRC channel.
type Channel struct {
	Config     *Config
	Digesters  []Digester
	connection net.Conn
	reader     Decoder
	writer     Encoder
	done       chan *ChannelError
}

// ChannelWriter represents a writer capable of sending messages to a channel.
type ChannelWriter interface {
	Send(message string) error
	SendCommand(command string, params []string, message string) error
}

// ChannelError is a struct consisting of a reference to a channel and an error that
// occurred on that channel.
type ChannelError struct {
	Channel *Channel
	error   error
}

// Error returns a description of the error. Satisfies the Error interface.
func (c *ChannelError) Error() string {
	return c.error.Error()
}

// NewTwitchChannel creates an IRC channel with Twitch's default server and port.
func NewTwitchChannel(channelName, username, token string, digesters ...Digester) *Channel {
	config := &Config{
		ChannelName: channelName,
		Username:    username,
		OAuthToken:  token,
		Server:      DefaultTwitchServer,
	}

	return &Channel{Config: config, Digesters: digesters[:]}
}

// Connect establishes a connection to an IRC server.
func (c *Channel) Connect() error {
	conn, err := net.Dial("tcp", c.Config.Server)
	if err != nil {
		return err
	}

	c.connection = conn
	c.reader = sirc.NewDecoder(conn)
	c.writer = sirc.NewEncoder(conn)
	c.done = make(chan *ChannelError)
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
func (c *Channel) Send(message string) error {
	return c.SendCommand("PRIVMSG", []string{fmt.Sprintf("#%s", c.Config.ChannelName)}, message)
}

// SendCommand sends a command to the channel.
func (c *Channel) SendCommand(command string, params []string, message string) error {
	m := &sirc.Message{
		Prefix: &sirc.Prefix{
			Name: c.Config.Username,
			User: c.Config.Username,
			Host: DefaultTwitchURI,
		},
		Command:  command,
		Params:   params,
		Trailing: message,
	}
	if err := c.writer.Encode(m); err != nil {
		return err
	}
	return nil
}

// Listen enters a loop and starts decoding IRC messages from the connected channel.
// Decoded messages are pushed to the digesters to be handled.
func (c *Channel) Listen() *ChannelError {
	// Close the connection when finished.
	defer c.connection.Close()

	err := c.startReceiving()

	return err
}

func (c *Channel) startReceiving() *ChannelError {
	for {
		select {
		case <-c.done:
			return nil
		default:
			c.connection.SetDeadline(time.Now().Add(10 * time.Minute))
			m, err := c.reader.Decode()
			if err != nil {
				return &ChannelError{Channel: c, error: err}
			}
			// If the message is a PING command from Twitch, respond with a PONG
			// without pushing the message through to the digesters
			if m.Command == "PING" {
				c.SendCommand("PONG", []string{}, "tmi.twitch.tv")
			} else {
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
				}
				c.handle(message)
			}
		}

	}
}

func (c *Channel) handle(m *Message) {
	for _, d := range c.Digesters {
		go d(*m, c)
	}
}
