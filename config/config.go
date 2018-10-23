package config

import "os"

type Config struct {
	Host string
	Port string
	User string
	Password string
	Tlsn string
}

// 環境変数として設定しておくこと
func NewConfig() Config {
	c := Config{
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
