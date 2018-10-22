package lib

import "os"

type Config struct {
	host string
	port string
	user string
	password string
	tlsn string
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
	if c.port == "" {
		c.port = "993"
	}
	return c
}
