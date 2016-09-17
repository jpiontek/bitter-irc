package birc

import "fmt"

// Digester is a handler function for parsing and reacting to  IRC chat messages.
// All digesters MUST be thread safe, they will be called from multiple threads.
type Digester func(m Message, channelWriter ChannelWriter)

// Logger is a digester that simply echoes out all messages to stdout.
func Logger(m Message, channelWriter ChannelWriter) {
	if m.Username != "" {
		fmt.Printf("\n%s: %s", m.Username, m.Content)
	}
}
