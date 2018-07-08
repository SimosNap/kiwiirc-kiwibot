package kiwibot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// IrcBot config structure
type IrcBot struct {
	Nick     string   `json:"nick"`
	Name     string   `json:"name"`
	Server   string   `json:"server"`
	TLS      bool     `json:"tls"`
	Channels []string `json:"channels"`
	Debug    bool     `json:"debug"`
	Verbose  bool     `json:"verbose"`
}

// Redis config structure
type Redis struct {
	Server   string `json:"server"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// Webhook config structure
type Webhook struct {
	Path   string `json:"path"`
	Listen string `json:"listen"`
}

// Repo config stucture
type Repo struct {
	Secret   string   `json:"secret"`
	Channels []string `json:"channels"`
	Events   []string `json:"events"`
}

// Config stucture
type Config struct {
	Ircbot  IrcBot          `json:"ircbot"`
	Redis   Redis           `json:"redis"`
	Webhook Webhook         `json:"webhook"`
	Repos   map[string]Repo `json:"repos"`
}

var config Config

// LoadConfig Loads config data from file
func LoadConfig(configFile string) {
	raw, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Println("config error:", err.Error())
		os.Exit(1)
	}
	config = Config{}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		log.Println("config error:", err.Error())
		os.Exit(1)
	}
	log.Println("config loaded")
}

// GetBotConf gets the ircbots config
func GetBotConf() IrcBot {
	return config.Ircbot
}

// GetRedisConf gets the redis config
func GetRedisConf() Redis {
	return config.Redis
}

// GetWebhookConf gets the webhook config
func GetWebhookConf() Webhook {
	return config.Webhook
}

// GetReposConf return the Repos map
func GetReposConf() map[string]Repo {
	return config.Repos
}

// GetRepoConf returns a Repo config from the Repos map
func GetRepoConf(repo string) Repo {
	return config.Repos[repo]
}
