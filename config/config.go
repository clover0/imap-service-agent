package config

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

var Conf Config

type Config struct {
	DB DataBaseConfig
	Host string
	Port string
	User string
	Password string
	Tlsn string
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
	Port string `toml:"port"`
}

// 環境変数として設定しておくこと
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


