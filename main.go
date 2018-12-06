package main

import (
	"flag"

	"github.com/SimosNap/kiwiirc-kiwibot/pkg/kiwibot"
)

func main() {
	configFile := flag.String("config", "config.conf", "Config file location")
	flag.Parse()

	kiwibot.LoadConfig(*configFile)
	kiwibot.CreateBot()
	kiwibot.StartUDP()
	kiwibot.BotLoop()
}
