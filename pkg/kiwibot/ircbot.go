package kiwibot

import (
	"fmt"

	irc "github.com/thoj/go-ircevent"
)

var irccon *irc.Connection

// CreateBot creates and instance of the ircbot
func CreateBot() {
	conf := GetBotConf()
	irccon = irc.IRC(conf.Nick, conf.Name)
	irccon.VerboseCallbackHandler = conf.Verbose
	irccon.Debug = conf.Debug
	irccon.UseTLS = conf.TLS

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range conf.Channels {
			irccon.Join(channel)
		}
	})

	err := irccon.Connect(conf.Server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}
}

// BotSend send a message to irc
func BotSend(channel string, message string) {
	irccon.Privmsg(channel, message)
}

// BotLoop starts the ircbots loop
func BotLoop() {
	irccon.Loop()
}

// GetIrcCon returns the irc connection
func GetIrcCon() *irc.Connection {
	return irccon
}
