package services

import (
	"bufio"
	"bytes"
	"github.com/develop/imap-agent/config"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	go_mail "net/mail"
	"net/textproto"
	"strings"
)

type FirstTimeSenderService struct {
	c  *client.Client
	ic config.IMAPConfig
	db *sqlx.DB
}

/**
メモ
未読にするには フラグをとる
 */
func (self *FirstTimeSenderService) execute() {
	log.Println("FirstTimeSenderService starting process...")

	cri := imap.NewSearchCriteria()
	cri.WithoutFlags = []string{imap.SeenFlag}

	log.Println("search mails...")
	ids, err := self.c.Search(cri)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("IDs found:", ids)

	// UIDで引いて処理するように変更する 
	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)

		messages := make(chan *imap.Message, 10)
		section := &imap.BodySectionName{}
		done := make(chan error, 1)
		go func() {
			done <- self.c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}, messages)
		}()

		log.Println("Unseen messages:")
		for msg := range messages {
			fromAddress := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
			log.Println("subject: " + msg.Envelope.Subject)
			log.Println("from: ", fromAddress)
			r := msg.GetBody(section)

			// copy reader
			var copyBuf bytes.Buffer
			tee := io.TeeReader(r, &copyBuf)

			goMail, err2 := go_mail.ReadMessage(tee)

			if err2 != nil {
				log.Fatal(err2)
			}
			log.Println("Appending mail...")
			tp := textproto.NewReader(bufio.NewReader(&copyBuf))
			mh, _ := tp.ReadMIMEHeader()

			s := mh.Get("X-imap-agent-serviced");
			if s != "" {
				continue
			}

			isFoundSenderInfo := self.findOrInsert(fromAddress, self.ic.User)
			if isFoundSenderInfo {
				continue
			}

			mh.Set("Subject", "[初回送信者]"+msg.Envelope.Subject)
			mh.Add("X-imap-agent-serviced", "first_time_sender")
			mh.Del("Message-ID") // deleteしなくてもいいかも
			header := ""
			for k, v := range mh {
				header = header + k + ": " + v[0] + "\r\n"
			}
			byteHeader, err3 := ioutil.ReadAll(strings.NewReader(header))

			// for build body
			buf := new(bytes.Buffer)

			if err3 != nil {
				log.Fatal(err3)
			}
			buf.Write(byteHeader)
			buf.Write([]byte("\r\n"))
			byteBody, err4 := ioutil.ReadAll(goMail.Body)
			if err4 != nil {
				log.Fatal(err4)
			}
			buf.Write(byteBody)

			log.Println("append mail.")
			self.c.Append("INBOX", nil, msg.Envelope.Date, buf)

		}
	}
}

func (self *FirstTimeSenderService) findOrInsert(fromAddress string, account string) (found bool) {
	log.Println("find or insert sender info.")
	var count int
	
	tx := self.db.MustBegin()

	err := tx.Get(&count, "SELECT COUNT(*) FROM senders WHERE mail_address = $1 AND to_account = $2", fromAddress, account)
	if err != nil {
		log.Fatal("select count error. err: ",err)
	}
	flag := false
	if count > 0 {
		flag = true
	} else { // if not exist sender info, insert new record
		tx.MustExec("INSERT INTO senders (mail_address, to_account, send_datetime) VALUES ($1, $2, current_timestamp)", fromAddress, account)
		flag = false
	}

	tx.Commit()
	log.Println("find or insert complete.")

	return flag
}
