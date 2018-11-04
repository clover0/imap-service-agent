package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

var Conf Config

type Config struct {
	DB DataBaseConfig `toml:"database"`
}

type IMAPConfig struct {
	Host string
	Port string
	User string
	Password string
	Tlsn string
}

type DataBaseConfig struct {
	Host string `toml:"host"`
	Port int `toml:"port"`
}

/**
New configuration for IMAP conncet.

Configure variables on environmental variables.
 */
func NewIMAPConfig() IMAPConfig {
	c := IMAPConfig{
		os.Getenv("IMAP_AGENT_PJ_IMAP_HOST"),
		os.Getenv("IMAP_AGENT_PJ_IMAP_PORT"),
		os.Getenv("IMAP_AGENT_PJ_IMAP_USER"),
		os.Getenv("IMAP_AGENT_PJ_IMAP_PASS"),
		os.Getenv("IMAP_AGENT_PJ_IMAP_TLS_SERVERNAME"),
	}
	if c.Port == "" {
		c.Port = "993"
	}
	return c
}

/**
Initialization configuration for imap-agent project 
 every environment.
 */
func InitConfig() {
	var config Config
	var configFile string
	configDir := "config"

	switch os.Getenv("ENV") {
	case "local":
		configFile = "config.local.toml"
	case "development":
		configFile = "config.development.toml"
	case "staging":
		configFile = "config.staging.toml"
	case "production":
		configFile = "config.production.toml"
	default:
		configFile = "config.local.toml"
	}

	_, err := toml.DecodeFile(configDir + "/" + configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	Conf = config
}


