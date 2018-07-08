package kiwibot

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/itsonlybinary/webhooks"
	"github.com/itsonlybinary/webhooks/github"
)

// StartWebook Configures Webhook modules and start its listeners
func StartWebook() error {
	conf := GetWebhookConf()
	hook := github.New(&github.Config{Secret: ""})
	hook.RegisterEvents(HandlePullRequest, github.PullRequestEvent)
	hook.RegisterEvents(HandlePush, github.PushEvent)

	err := webhooks.Run(hook, conf.Listen, conf.Path)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}

// HandlePullRequest handles GitHub pull_request events
func HandlePullRequest(payload interface{}, header webhooks.Header, raw []byte, rw http.ResponseWriter) {
	fmt.Println("Handling Pull Request")

	pl := payload.(github.PullRequestPayload)

	// Do whatever you want from here...
	fmt.Println(header)
	fmt.Printf("%+v", pl)
}

// HandlePush handles GitHub push events
func HandlePush(payload interface{}, header webhooks.Header, raw []byte, rw http.ResponseWriter) {
	fmt.Println("Handling Push")

	pl := payload.(github.PushPayload)

	fmt.Println("repo", pl.Repository.FullName)
	lcRepo := strings.ToLower(pl.Repository.FullName)
	reposConf := GetReposConf()
	if _, ok := reposConf[lcRepo]; !ok {
		log.Println("Error: got event for untracked repo")
		return
	}

	conf := GetRepoConf(pl.Repository.FullName)

	if !validateSecret(conf.Secret, header["X-Hub-Signature"][0], raw) {
		webhooks.DefaultLog.Error("HMAC verification failed")
		http.Error(rw, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
		return
	}

	url := ShortenURL(pl.Compare)
	repo := pl.Repository.FullName
	pushName := pl.Pusher.Name
	commits := pl.Commits
	branch := GetLast(strings.Split(pl.Ref, "/"))

	for _, dest := range conf.Channels {
		BotSend(dest, fmt.Sprintf("[%s] %s pushed %d commit(s) to %s. %s", repo, pushName, len(commits), branch, url))

		for i := 0; i < len(commits) && i < 5; i++ {
			commit := commits[i]
			BotSend(dest, fmt.Sprintf("[%s] %s %s: %s", repo, ShortenSHA(commit.ID), commit.Author.Username, commit.Message))
		}
	}
}

// HandleIssues handles GitHub issue events
func HandleIssues(payload interface{}, header webhooks.Header, raw []byte, resp http.ResponseWriter) {
	fmt.Println("Handling Issues")

	pl := payload.(github.IssuesPayload)

	// Do whatever you want from here...
	fmt.Println(header)
	fmt.Printf("%+v", pl)
}

func validateSecret(secret string, sig string, raw []byte) bool {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write(raw)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	if hmac.Equal([]byte(sig[5:]), []byte(expectedMAC)) {
		return true
	}
	return false
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
