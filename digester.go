package birc

import (
	"fmt"
	"io"
)

const timeFormat = "2006-01-06 15:04:05"

// Digester is a handler function for parsing and reacting to IRC chat messages.
// All digesters MUST be thread safe, they will be called from multiple go routines.
type Digester func(m Message, c ChannelWriter)

// Logger is a digester that simply echoes out user's messages to stdout.
func Logger(m Message, c ChannelWriter) {
	if m.Username != "" && m.Content != "" {
		fmt.Printf("\n%s %s: %s", m.Time.Format(timeFormat), m.Username, m.Content)
	}
}

// CustomLogger will write the incoming messages to the supplied io.Writer.
func CustomLogger(w io.Writer) Digester {
	return func(m Message, c ChannelWriter) {
		if m.Username != "" && m.Content != "" {
			o := fmt.Sprintf("\n%s %s: %s", m.Time.Format(timeFormat), m.Username, m.Content)
			w.Write([]byte(o))
		}
	}
}
