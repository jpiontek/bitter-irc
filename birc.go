package birc

import sirc "github.com/sorcix/irc"

const (
	// DefaultTwitchPort is Twitch's default IRC port
	DefaultTwitchPort = "6667"
	// DefaultTwitchTlsPort is Twitch's default IRC port for TLS connections
	DefaultTwitchTlsPort = "443"
	// DefaultTwitchURI is Twitch's default IRC server
	DefaultTwitchURI = "irc.chat.twitch.tv"
	// DefaultTwitchServer is a helper with the DefaultTwitchURI and
	// DefaultTwitchPort combined.
	DefaultTwitchServer = DefaultTwitchURI + ":" + DefaultTwitchPort
	// DefaultTwitchTlsServer is the default TLS server and port
	DefaultTwitchTlsServer = DefaultTwitchURI + ":" + DefaultTwitchTlsPort
)

// Encoder represents a struct capable of encoding an IRC message.
type Encoder interface {
	Encode(m *sirc.Message) error
}

// Decoder represents a struct capable of decoding incoming IRC messages.
type Decoder interface {
	Decode() (*sirc.Message, error)
}
