package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/thoj/go-ircevent"
	"regexp"
	"time"
)

// Config
var ircNick = "PreviewBot"
var ircName = "PreviewBot"
var ircServer = "chat.freenode.net:6697"
var ircTLS = true
var ircChannels = []string{"#kiwiirc"}
var ircDebug = false
var ircVerbose = false

var redisServer = "127.0.0.1:6379"
var redisPassword = ""
var redisDB = 0

func main() {
	irccon := irc.IRC(ircNick, ircName)
	irccon.VerboseCallbackHandler = ircVerbose
	irccon.Debug = ircDebug
	irccon.UseTLS = ircTLS

	irccon.AddCallback("001", func(e *irc.Event) {
		for _, channel := range ircChannels {
			irccon.Join(channel)
		}
	})

	err := irccon.Connect(ircServer)
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	rclient := redis.NewClient(&redis.Options{
		Addr:     redisServer,
		Password: redisPassword,
		DB:       redisDB,
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
						"Preivew of %v is now availiable here: %v",
						jsonData["tidy_ref"],
						buildUrl("preview", jsonData),
					)
				} else {
					out = fmt.Sprintf(
						"Preivew of %v failed to build successfully. log: %v",
						jsonData["tidy_ref"],
						buildUrl("log", jsonData),
					)
				}

				// append the title if it exists
				if jsonData["title"] != nil && jsonData["title"] != "" {
					out += fmt.Sprintf(" - %v", jsonData["title"])
				}
                
                for _, channel := range ircChannels {
                    irccon.Privmsg(channel, out)
                }
			}

		}
	}
}

func buildUrl(urlType string, jsonData map[string]interface{}) string {
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
		panic(err)
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
