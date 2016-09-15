package birc

import "fmt"

// Digester is a handler function for parsing IRC chat messages.
type Digester func(c <-chan Message)

// Logger is a digester that simply echoes out all messages to stdout.
func Logger(c <-chan Message) {
	for {
		select {
		case m := <-c:
			if m.Username != "" {
				fmt.Printf("\n%s: %s", m.Username, m.Content)
			}
		}
	}
}
