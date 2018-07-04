package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-redis/redis"
	"github.com/itsonlybinary/kiwiirc-previewbot/pkg/previewbot"
	"github.com/thoj/go-ircevent"
)

var botConf previewbot.IrcBot
var redisConf previewbot.Redis

func main() {
	previewbot.LoadConfig()
	botConf = previewbot.GetBotConf()
	redisConf = previewbot.GetRedisConf()

	irccon := irc.IRC(botConf.Nick, botConf.Name)
	irccon.VerboseCallbackHandler = botConf.Verbose
	irccon.Debug = botConf.Debug
	irccon.UseTLS = botConf.Tls

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range botConf.Channels {
			irccon.Join(channel)
		}
	})

	err := irccon.Connect(botConf.Server)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	rclient := redis.NewClient(&redis.Options{
		Addr:     redisConf.Server,
		Password: redisConf.Password,
		DB:       redisConf.Database,
	})

	go redisstart(rclient, irccon)

	irccon.Loop()
}

func builderMessage(irccon *irc.Connection, jsonMessage map[string]interface{}) {
	if jsonMessage["cmd"] == "job_finished" {
		jsonData := jsonMessage["data"].(map[string]interface{})

		if jsonData["repo_user"] == "kiwiirc" {
			reg, _ := regexp.Compile("^pull/(\\d+)$")
			repoRef := fmt.Sprint(jsonData["repo_ref"])

			if reg.MatchString(repoRef) {
				jsonData["tidy_ref"] = "#" + reg.FindStringSubmatch(repoRef)[1]
			} else {
				jsonData["tidy_ref"] = repoRef
			}

			if jsonData["github_sha"] == jsonData["hosted_sha"] {
				var out string

				if jsonData["build_success"] == "1" {
					out = fmt.Sprintf(
						"Preview of %v is now available here: %v",
						jsonData["tidy_ref"],
						buildURL("preview", jsonData),
					)
				} else {
					out = fmt.Sprintf(
						"Preivew of %v failed to build successfully. log: %v",
						jsonData["tidy_ref"],
						buildURL("log", jsonData),
					)
				}

				// append the title if it exists
				if jsonData["title"] != nil && jsonData["title"] != "" {
					out += fmt.Sprintf(" - %v", jsonData["title"])
				}

				for _, channel := range botConf.Channels {
					irccon.Privmsg(channel, out)
				}
			}

		}
	}
}

func buildURL(urlType string, jsonData map[string]interface{}) string {
	var tidyRef, url string
	reg, _ := regexp.Compile("^pull/(\\d+)$")
	repoRef := fmt.Sprint(jsonData["repo_ref"])

	if reg.MatchString(repoRef) {
		tidyRef = "pull-" + reg.FindStringSubmatch(repoRef)[1]
	} else {
		tidyRef = repoRef
	}

	if urlType == "preview" {
		url = fmt.Sprintf(
			"http://builds.kiwiirc.com/%v/%v/",
			jsonData["repo_user"],
			tidyRef,
		)
	} else {
		url = fmt.Sprintf(
			"http://builds.kiwiirc.com/logs/%v_%v.txt",
			jsonData["repo_user"],
			tidyRef,
		)
	}

	return url
}

func redisstart(rclient *redis.Client, irccon *irc.Connection) {
	pubsub := rclient.Subscribe("kiwibuilder")
	defer pubsub.Close()

	_, err := pubsub.ReceiveTimeout(time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		var data map[string]interface{}
		json.Unmarshal([]byte(msg.Payload), &data)
		builderMessage(irccon, data)
	}
}
