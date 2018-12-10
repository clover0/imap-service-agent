package lib

import (
	"crypto/tls"
	"fmt"
	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/db"
	"github.com/develop/imap-agent/lib/services"
	"github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
	"log"
	"os"
)

// TODO: INBOXなどの指定を外部から指定できるようにする

func RunIdleDaemon(){
	dbis := db.NewDB()
	//defer dbis.Close()
	conf := config.NewIMAPConfig()
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

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(conf.User, conf.Password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")
	// Select a mailbox
	if _, err := c.Select("INBOX", false); err != nil {
		log.Fatal(err)
	}

	idleClient := idle.NewClient(c)

	// Create a channel to receive mailbox updates
	// IDLEを行ったコネクションで他のコマンドは送信できない
	// 切断するか別コネクションにする 現状別コネクション
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
			fmt.Println("data:", update)
			// 新たにコネクションを生成
			newClient := newConnection()
			services.Execute(newClient, conf, dbis)
			// サービス実行後は切断
			newClient.Close()
		case err := <-done:
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Not idling anymore")
			return
		}
	}
}

// TODO: 最初からコネクションを作っておく方式にする
func newConnection() *client.Client{
	conf := config.NewIMAPConfig()
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
	if _, err := c.Select("INBOX", false); err != nil {
		log.Fatal(err)
	}

	return c
}