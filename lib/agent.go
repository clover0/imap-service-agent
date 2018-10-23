package lib

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/develop/imap-agent/config"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func Do() {
	log.Println("Connecting to server...")

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

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(conf.User, conf.Password); err != nil {
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

	// Select INBOX
	mbox, err := c.Select("[Gmail]/すべてのメール", true)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		//We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done2 := make(chan error, 1)
	go func() {
		done2 <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
		fmt.Println("-- ", msg.Body)

		seqSet4Body := new(imap.SeqSet)
		//fmt.Println("seqnum: ", msg.SeqNum)
		//fmt.Println("body: ", msg.Body)
		//fmt.Println("items: ", msg.Items)
		//fmt.Println("body: ", msg.Envelope)
		seqSet4Body.AddNum(msg.SeqNum)
		
		section := &imap.BodySectionName{}
		items := []imap.FetchItem{section.FetchItem()}

		messageBody := make(chan *imap.Message)
		done3 := make(chan error, 1)
		go func() {
			done3 <- c.Fetch(seqSet4Body, items, messageBody)
			//if err := c.Fetch(seqSet4Body, items, messageBody); err != nil {
			//	log.Fatal(err)
			//}
		}()
		body := <- messageBody

		if body == nil {
			log.Fatal("Server didn't returned message")
		}
		
		fmt.Println("-- ", body)

		r := body.GetBody(section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		// Print some info about the message
		header := mr.Header
		if date, err := header.Date(); err == nil {
			log.Println("Date:", date)
		}
		if from, err := header.AddressList("From"); err == nil {
			log.Println("From:", from)
		}
		if to, err := header.AddressList("To"); err == nil {
			log.Println("To:", to)
		}
		if subject, err := header.Subject(); err == nil {
			log.Println("Subject:", subject)
		}

		// Process each message's part
		//for {
		//	p, err := mr.NextPart()
		//	if err == io.EOF {
		//		break
		//	} else if err != nil {
		//		log.Fatal(err)
		//	}
		//
		//	switch h := p.Header.(type) {
		//	case mail.TextHeader:
		//		// This is the message's text (can be plain-text or HTML)
		//		b, _ := ioutil.ReadAll(p.Body)
		//		log.Println("Got text: %v", string(b))
		//	case mail.AttachmentHeader:
		//		// This is an attachment
		//		filename, _ := h.Filename()
		//		log.Println("Got attachment: %v", filename)
		//	}
		//}

	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

}