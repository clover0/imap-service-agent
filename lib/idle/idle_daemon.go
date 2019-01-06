package idle

import (
	"crypto/tls"
	"fmt"
	"github.com/develop/imap-agent/lib/db/models"
	"github.com/jmoiron/sqlx"
	"log"
	"os"

	"github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"

	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/db"
	"github.com/develop/imap-agent/lib/services"
)

// Run runs imap idle.
// when updating mailbox, this does services.
func Run() {
	dbis := db.NewDB()
	defer dbis.Close()

	//conf := config.NewIMAPConfig()
	conf := NewIMAPConfigFrom(dbis)
	connStr := fmt.Sprintf("%s:%s", conf.Host, conf.Port)

	// confirm this skipping verify on production env
	tlsc := &tls.Config{InsecureSkipVerify: true}
	if conf.Tlsn != "" {
		tlsc.ServerName = conf.Tlsn
	}

	// Connect to server
	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	c.SetDebug(os.Stdout)

	defer c.Logout()

	// Login
	if err := c.Login(conf.User, conf.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")
	// Select a mailbox
	if _, err := c.Select(conf.MailBox, false); err != nil {
		log.Fatal(err)
	}

	idleClient := idle.NewClient(c)

	// Create a channel to receive mailbox updates
	updates := make(chan client.Update)
	c.Updates = updates

	// Start idling
	done := make(chan error, 1)
	go func() {
		done <- idleClient.IdleWithFallback(nil, 0)
	}()

	// Listen for updates
	for {
		log.Println("Idling...")
		select {
		case update := <-updates:
			log.Println("New update:", update)
			con4Service := newConnection(conf)
			services.Execute(con4Service, &conf, dbis)
			con4Service.Logout()
		case err := <-done:
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Not idling anymore")
			return
		}
	}
}

func newConnection(conf config.IMAPConfig) *client.Client {
	connStr := fmt.Sprintf("%s:%s", conf.Host, conf.Port)

	// 本番運用の際はスキップしてよいのか確認すること
	tlsc := &tls.Config{InsecureSkipVerify: true}
	if conf.Tlsn != "" {
		tlsc.ServerName = conf.Tlsn
	}

	// Connect to server
	c, err := client.DialTLS(connStr, tlsc)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	c.SetDebug(os.Stdout)

	// Login
	if err := c.Login(conf.User, conf.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")
	// Select a mailbox
	if _, err := c.Select(conf.MailBox, false); err != nil {
		log.Fatal(err)
	}

	return c
}

func NewIMAPConfigFrom(db *sqlx.DB) config.IMAPConfig {
	//mailAddress := "imap_agent_test_y@yahoo.co.jp"
	mailAddress := "imap.agent.test@gmail.com"
	account := models.Account{}
	err := db.Get(&account, "SELECT * FROM accounts WHERE mail_address = $1", mailAddress)
	if err != nil {
		log.Fatal(err)
	}
	c := config.IMAPConfig{
		account.ImapHost.String,
		account.ImapPort.String,
		account.MailAddress.String,
		account.Password.String,
		account.ImapHost.String,
		"INBOX",
	}
	if c.Port == "" {
		c.Port = "993"
	}
	return c
}
