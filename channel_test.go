package birc_test

import (
	"bufio"
	"fmt"
	"net"
	"testing"

	"github.com/jpiontek/bitter-irc"
	sirc "github.com/sorcix/irc"
)

type Writer struct {
	Proxy func(m *sirc.Message)
}

func (w *Writer) Encode(m *sirc.Message) error {
	w.Proxy(m)
	return nil
}

func TestNewTwitchChannel(t *testing.T) {
	c := birc.NewTwitchChannel("test", "foobar", "abc123")

	if c == nil {
		t.Error(fmt.Errorf("channel was nil"))
	}

	if c.Config.ChannelName != "test" {
		t.Error(fmt.Errorf("Expected ChannelName test, got: %s", c.Config.ChannelName))
	}

	if c.Config.Username != "foobar" {
		t.Error(fmt.Errorf("Expected Username foobar, got: %s", c.Config.Username))
	}

	if c.Config.OAuthToken != "abc123" {
		t.Error(fmt.Errorf("Expected OAuthToken abc123, got: %s", c.Config.OAuthToken))
	}

	if c.Config.Server != birc.DefaultTwitchServer {
		t.Error(fmt.Errorf("Expected Server %s, got: %s", birc.DefaultTwitchServer, c.Config.Server))
	}
}

func TestConnect(t *testing.T) {
	config := &birc.Config{
		ChannelName: "test",
		Username:    "foobar",
		OAuthToken:  "abc123",
		Server:      "127.0.0.1:4444",
	}

	// Create a listener on the same port as the test configuration
	l, err := net.Listen("tcp", config.Server)
	if err != nil {
		t.Error(err)
	}

	var digesters = []birc.Digester{birc.Logger}

	c := &birc.Channel{Config: config, Digesters: digesters}

	if c == nil {
		t.Error("Expected a channel")
	}

	if len(c.Digesters) != 1 {
		t.Error("Expected Logger digester")
	}

	err = c.Connect()
	if err != nil {
		t.Error(err)
	}

	l.Close()
}

func TestConnectionError(t *testing.T) {
	config := &birc.Config{
		ChannelName: "test",
		Username:    "foobar",
		OAuthToken:  "abc123",
		Server:      "127.0.0.1:4444",
	}

	// Create a listener on the same port as the test configuration
	l, err := net.Listen("tcp", config.Server)
	if err != nil {
		t.Error(err)
	}

	c := &birc.Channel{Config: config}

	if c == nil {
		t.Error("Expected a channel")
	}

	err = c.Connect()
	if err != nil {
		t.Error(err)
	}

	err = c.Authenticate()
	if err != nil {
		t.Error(err)
	}

	ch := make(chan error, 1)
	go func() {
		err := c.Listen()
		ch <- err
	}()

	// Close the listener to simulate losing connection to the server
	go func() {
		l.Close()
	}()

	select {
	case err := <-ch:
		// should get an error
		if err == nil {
			t.Error("expected error")
		}
	}

	l.Close()
}

func TestPing(t *testing.T) {
	config := &birc.Config{
		ChannelName: "test",
		Username:    "foobar",
		OAuthToken:  "abc123",
		Server:      "127.0.0.1:4444",
	}

	// Create a listener on the same port as the test configuration
	l, err := net.Listen("tcp", config.Server)
	if err != nil {
		t.Error(err)
	}

	c := &birc.Channel{Config: config}

	if c == nil {
		t.Error("Expected a channel")
	}

	err = c.Connect()
	if err != nil {
		t.Error(err)
	}

	ch := make(chan string, 1)
	connection, _ := l.Accept()
	// start a go routine for the tcp server to listen to messages,
	// when it receives one send it out to the ch channel
	go func(conn net.Conn) {
		for {
			message, _, err := bufio.NewReader(conn).ReadLine()
			if err != nil {
				ch <- err.Error()
			}
			ch <- string(message)
			break
		}
	}(connection)

	// start a go routine to start the twich channel listening to the server
	go func() {
		err := c.Listen()
		ch <- err.Error()
	}()

	// simulate the twitch server's occasional ping command
	connection.Write([]byte("PING :tmi.twitch.tv\n"))

	// wait to get the pong response from the channel sent to the server
	for {
		select {
		case result := <-ch:
			// if the result is not the expectd properly formed PONG response then fail
			if result != ":foobar!foobar@irc.chat.twitch.tv PONG :tmi.twitch.tv" {
				t.Errorf("expected pong command got %s", result)
			}
			break
		}
		break
	}

}

func TestAuthenticate(t *testing.T) {
	c := birc.NewTwitchChannel("test", "foobar", "abc123")

	var passCalled, nickCalled, joinCalled bool
	handler := func(m *sirc.Message) {
		switch m.Command {
		case sirc.PASS:
			passCalled = true
			if m.Params[0] != fmt.Sprintf("oauth:%s", c.Config.OAuthToken) {
				t.Error(fmt.Errorf("Expected correct oauth string, got %s", m.Params[0]))
			}
		case sirc.NICK:
			nickCalled = true
			if m.Params[0] != c.Config.Username {
				t.Error(fmt.Errorf("Expected Username to be %s, got %s", c.Config.Username, m.Params[0]))
			}
		case sirc.JOIN:
			joinCalled = true
			if m.Params[0] != fmt.Sprintf("#%s", c.Config.ChannelName) {
				t.Error(fmt.Errorf("Expected channel #%s, got %s", c.Config.ChannelName, m.Params[0]))
			}
		}
	}

	stubWriter := &Writer{handler}
	c.SetWriter(stubWriter)
	c.Authenticate()

	if !passCalled {
		t.Error("Expected PASS to be sent")
	}
	if !nickCalled {
		t.Error("Expected NICK to be sent")
	}
	if !joinCalled {
		t.Error("Expected JOIN to be sent")
	}
}
