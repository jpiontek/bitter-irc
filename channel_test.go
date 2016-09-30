package birc_test

import (
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

	l.Close()
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
