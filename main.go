package main

import (
	"flag"

	"github.com/itsonlybinary/kiwiirc-kiwibot/pkg/kiwibot"
)

func main() {
	configFile := flag.String("config", "config.conf", "Config file location")
	flag.Parse()
	kiwibot.LoadConfig(*configFile)

	// Create the ircbot instance
	kiwibot.CreateBot()

	// Create and start redis client
	kiwibot.RedisCreate()
	go kiwibot.RedisStart()

	// Start the webhook listener
	kiwibot.WebhookStart()

	// Loop the ircbot
	kiwibot.BotLoop()
}
