package birc

import "time"

// Message is a decoded IRC message.
type Message struct {
	Name     string
	Username string
	Content  string
	Command  string
	Params   []string
	Time     time.Time
}
