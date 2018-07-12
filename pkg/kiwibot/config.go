package kiwibot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config stucture
type Config struct {
	Nick     string   `json:"nick"`
	Name     string   `json:"name"`
	Server   string   `json:"server"`
	TLS      bool     `json:"tls"`
	Channels []string `json:"channels"`
	Debug    bool     `json:"debug"`
	Verbose  bool     `json:"verbose"`
	UDPaddr  string   `json:"udpaddr"`
}

var config Config

// LoadConfig Loads config data from file
func LoadConfig(configFile string) {
	raw, err := ioutil.ReadFile(configFile)

	if err != nil {
		log.Println("Could not read config:", err.Error())
		os.Exit(1)
	}
	config = Config{}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		log.Println("Error in config JSON:", err.Error())
		os.Exit(1)
	}
	log.Println("Config loaded:", configFile)
}

// GetConfig gets the ircbots config
func GetConfig() Config {
	return config
}
