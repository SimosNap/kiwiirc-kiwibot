package kiwibot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/go-redis/redis"
)

var rclient *redis.Client

// RedisCreate creates a redis client instance
func RedisCreate() {
	conf := GetRedisConf()
	rclient = redis.NewClient(&redis.Options{
		Addr:     conf.Server,
		Password: conf.Password,
		DB:       conf.Database,
	})
}

// RedisStart starts the redis Subscription
func RedisStart() {
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
			fmt.Println(err)
			return
		}
		var data map[string]interface{}
		json.Unmarshal([]byte(msg.Payload), &data)
		builderMessage(data)
	}
}

func builderMessage(jsonMessage map[string]interface{}) {
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

				conf := GetBotConf()
				irccon := GetIrcCon()
				for _, channel := range conf.Channels {
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
