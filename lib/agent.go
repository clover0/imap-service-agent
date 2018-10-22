package lib

import (
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"os"
)

func Do() {
	log.Println("Connecting to server...")

	// 環境変数として設定しておくこと
	host := os.Getenv("IMAP_AGENT_PJ_IMAP_HOST")
	port := os.Getenv("IMAP_AGENT_PJ_IMAP_PORT")
	user := os.Getenv("IMAP_AGENT_PJ_IMAP_USER")
	pass := os.Getenv("IMAP_AGENT_PJ_IMAP_PASS")
	tlsn := os.Getenv("IMAP_AGENT_PJ_IMAP_TLS_SERVERNAME")
	if port == "" {
		port = "993"
	}

	connStr := fmt.Sprintf("%s:%s", host, port)

	tlsc := &tls.Config{InsecureSkipVerify: true}
	if tlsn != "" {
		tlsc.ServerName = tlsn
	}

	// Connect to server
	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(user, pass); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func () {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

}