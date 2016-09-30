package birc_test

import (
	"fmt"
	"testing"

	"github.com/jpiontek/bitter-irc"
)

func TestDefaultTwitchServer(t *testing.T) {
	if birc.DefaultTwitchServer != "irc.chat.twitch.tv:6667" {
		t.Error(fmt.Errorf("invalid DefaultTwitchServer: %s", birc.DefaultTwitchServer))
	}
}
