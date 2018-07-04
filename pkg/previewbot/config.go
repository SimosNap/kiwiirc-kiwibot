package previewbot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type IrcBot struct {
	Nick     string   `json:"nick"`
	Name     string   `json:"name"`
	Server   string   `json:"server"`
	Tls      bool     `json:"tls"`
	Channels []string `json:"channels"`
	Debug    bool     `json:"debug"`
	Verbose  bool     `json:"verbose"`
}

type Redis struct {
	Server   string `json:"server"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

type Config struct {
	Ircbot IrcBot `json:"ircbot"`
	Redis  Redis  `json:"redis"`
}

var config Config

func LoadConfig() error {
	raw, err := ioutil.ReadFile("config.conf")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	config = Config{}
	fmt.Println("config loaded")

	err = json.Unmarshal(raw, &config)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func GetBotConf() IrcBot {
	return config.Ircbot
}

func GetRedisConf() Redis {
	return config.Redis
}
