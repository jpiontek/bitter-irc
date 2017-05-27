# BitterBot IRC Client

Bitter IRC is a streamlined IRC library specifically designed for Twitch's IRC servers.
To see an example application go [here](https://github.com/jpiontek/bitter-irc-example).

## Example
```go
import "github.com/jpiontek/bitter-irc"

// Your OAuth Key
oauthKey := "my_oauth_key"
// Your account username
username := "fred_bot"
// The channel you'd like to listen to
channelName := "awesome_streamer"
// Use TLS
tls := true

// Create a new channel by supplying the necessary info and the Logger digester.
channel := birc.NewTwitchChannel(channelName, username, oauthKey, tls, birc.Logger)

// Connect estblishes the underlying TCP connection
err := channel.Connect()
if err != nil {
  // Handle error
}

// Authenticate will send the proper credentials and join the channel.
err := channel.Authenticate()
if err != nil {
  // Handle error
}

// Listen begins listening on the channel and handling messages. It is blocking,
// so you may want to wrap it in a go routine if you intend to continue executing
// on the current thread.
err := channel.Listen()
if err != nil {
 // Channel eventually had an error.
}
```

## Digesters
Digesters are simply functions used to handle incoming IRC messages. They have the signature:
```go
type Digester func(m Message, c ChannelWriter)
```

You can pass in any number of digesters to the NewTwitchChannel function. They **MUST** be threadsafe as
they will be called by multiple go routines. An example of the Logger digester:

```go
func Logger(m Message, w ChannelWriter) {
	if m.Username != "" && m.Content != "" {
		fmt.Printf("\n%s %s: %s", m.Time.Format(timeFormat), m.Username, m.Content)
	}
}
```

You can see the Logger digester just prints a formatted string to stdout if the message has a username and
some sort of content.

The ChannelWriter is an interface that represents a channel you can write to via the Send function.

```go
// Simply checks if the content of someone's message is !command. If so then
// the digester replies in the channel with "Executing command!".
if m.Content == "!command" {
  w.Send("Executing command!")
}
```

The ChannelWriter also supports SendMessage. You can send any message struct
via this function.

```go
message := &birc.Message{
  Command: "PONG",
  Content: "tmi.twitch.tv",
}

err := w.SendMessage(message)
```

The ChannelWriter also supports retrieving the Channel's configuration.

```go
config := w.GetConfig()
```

The Message struct passed into each digester:
```go
type Message struct {
  Name     string
  Username string
  Content  string
  Command  string
  Host     string
  Params   []string
  Time     time.Time
}
```

