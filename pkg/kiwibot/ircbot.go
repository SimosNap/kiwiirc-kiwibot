package kiwibot

import (
	"log"
	"time"

	irc "github.com/thoj/go-ircevent"
)

var spamSafeEnabled = false
var irccon *irc.Connection
var spamLimiter = make(map[string]*[8]time.Time)

// CreateBot creates and instance of the ircbot
func CreateBot() {
	conf := GetConfig()
	irccon = irc.IRC(conf.Nick, conf.Name)
	irccon.VerboseCallbackHandler = conf.Verbose
	irccon.Debug = conf.Debug
	irccon.UseTLS = conf.TLS

	for _, channel := range conf.Channels {
		spamLimiter[channel] = &[8]time.Time{}
	}

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range conf.Channels {
			irccon.Join(channel)
		}
	})

	err := irccon.Connect(conf.Server)
	if err != nil {
		log.Println("IRC connect error:", err)
		return
	}
}

// BotSend send a message to irc
func BotSend(channel string, message string) {
	conf := GetConfig()
	if !Contains(conf.Channels, channel) {
		log.Println("Attempted to send a message to an unconfigured channel:", channel)
		return
	}
	spam := spamLimiter[channel]
	if spamSafe(spam) {
		if spamSafeEnabled {
			spamSafeEnabled = false
		}
		irccon.Privmsg(channel, message)
	} else {
		if !spamSafeEnabled {
			irccon.Action(channel, "starts throwing spam towards /dev/null")
			spamSafeEnabled = true
		}
	}
	spamAdd(spam, time.Now())
}

// BotLoop starts the ircbots loop
func BotLoop() {
	irccon.Loop()
}

// GetIrcCon returns the irc connection
func GetIrcCon() *irc.Connection {
	return irccon
}

func spamSafe(spam *[8]time.Time) bool {
	now := time.Now()
	if (spam[0] != time.Time{}) && now.Sub(spam[0]).Seconds() <= 8 {
		return false
	}
	if !spamSafeEnabled {
		spamSafeEnabled = true
	}
	return true
}

func spamAdd(spam *[8]time.Time, new time.Time) {
	for i := 1; i < len(spam); i++ {
		spam[i-1] = spam[i]
	}
	spam[7] = new
}
