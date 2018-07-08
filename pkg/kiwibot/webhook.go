package kiwibot

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// WebhookStart Configures Webhook modules and start its listeners
func WebhookStart() {
	conf := GetWebhookConf()
	http.HandleFunc(conf.Path, hookHandler)
	http.ListenAndServe(conf.Listen, nil)
}

func hookHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	log.Println("Webhook received")
	if r.Method != "POST" {
		hookError(w, "Disallowed method used", http.StatusMethodNotAllowed)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	log.Println("X-GitHub-Event:", event)
	if !Contains([]string{"push", "pull_request", "issues"}, event) {
		hookError(w, "Unexpected event received", http.StatusNotImplemented)
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		hookError(w, "Could not read payload from body", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("X-Hub-Signature")[5:]
	log.Println("X-Hub-Signature:", signature)
	if len(signature) == 0 {
		hookError(w, "Missing X-Hub-Signature", http.StatusForbidden)
		return
	}

	var data map[string]interface{}
	json.Unmarshal([]byte(payload), &data)
	repo := data["repository"].(map[string]interface{})["full_name"].(string)
	lcRepo := strings.ToLower(repo)

	if _, ok := GetReposConf()[lcRepo]; !ok {
		hookError(w, "Hook fired for unexpected repo", http.StatusForbidden)
		return
	}

	conf := GetRepoConf(lcRepo)

	mac := hmac.New(sha1.New, []byte(conf.Secret))
	mac.Write(payload)
	expSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expSignature)) {
		hookError(w, "HMAC verification failed", http.StatusForbidden)
		return
	}

	if !eventWanted(conf, event) {
		log.Println("Unwanted event:", event)
		return
	}

	switch event {
	case "push":
		pusher := data["pusher"].(map[string]interface{})["name"].(string)
		commits := data["commits"].([]interface{})
		branch := GetLast(strings.Split(data["ref"].(string), "/"))
		url := ShortenURL(data["compare"].(string))

		for _, dest := range conf.Channels {
			BotSend(dest, fmt.Sprintf("[%s] %s pushed %d commit(s) to %s. %s", repo, pusher, len(commits), branch, url))

			for i := 0; i < len(commits) && i < 5; i++ {
				commit := commits[i].(map[string]interface{})
				sha := ShortenSHA(commit["id"].(string))
				var author string
				if commit["author"].(map[string]interface{})["username"] != nil {
					author = commit["author"].(map[string]interface{})["username"].(string)
				} else {
					author = commit["author"].(map[string]interface{})["name"].(string)
				}
				message := commit["message"].(string)
				BotSend(dest, fmt.Sprintf("[%s] %s %s: %s", repo, sha, author, message))
			}
		}
		break
	case "pull_request":
	case "issues":
		// '[%s] Issue opened. #%d %s', event.payload.repository.full_name, event.payload.issue.number, event.payload.issue.title)
		action := data["action"].(string)
		var number int
		if event == "issues" {
			number = data["issue"].(map[string]interface{})["number"].(int)
		} else {
			number = 1
		}
		title := data["issue"].(map[string]interface{})["title"].(string)
		for _, dest := range conf.Channels {
			BotSend(dest, fmt.Sprintf("[%s] %s %s #%d %2", repo, event, action, number, title))
		}
		break
	}
}

func hookError(w http.ResponseWriter, m string, i int) {
	log.Println("Erorr:", m)
	http.Error(w, fmt.Sprintf("%v %s - %s", i, http.StatusText(i), m), i)
}

func eventWanted(repo Repo, event string) bool {
	if Contains(repo.Events, "*") {
		return true
	}
	if Contains(repo.Events, event) {
		return true
	}
	return false
}
