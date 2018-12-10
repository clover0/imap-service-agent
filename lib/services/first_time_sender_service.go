package services

import (
	"bufio"
	"bytes"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"io"
	"io/ioutil"
	"log"
	go_mail "net/mail"
	"net/textproto"
	"strings"
)

type FirstTimeSenderService struct {
	c *client.Client
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
			log.Println("subject: " + msg.Envelope.Subject)
			r := msg.GetBody(section)

			log.Println("go target message:", r)

			// copy reader
			var copyBuf bytes.Buffer
			tee := io.TeeReader(r, &copyBuf)

			log.Println("go read message:...")
			log.Println("before reader:", tee)

			goMail, err2 := go_mail.ReadMessage(tee)

			log.Println("after reader:", &copyBuf)

			if err2 != nil {
				log.Fatal(err2)
			}
			log.Println("goMail is:", goMail)
			log.Println("Appending mail...")
			log.Println("write buff header")
			tp := textproto.NewReader(bufio.NewReader(&copyBuf))
			mh, _ := tp.ReadMIMEHeader()
			log.Println("before header is :", mh)

			s := mh.Get("X-imap-agent-serviced");
			if s != "" {
				log.Println("X-imap-agent-serviced exists")
				continue
			}

			mh.Set("Subject", "[初回送信者]"+msg.Envelope.Subject)
			mh.Add("X-imap-agent-serviced", "first_time_sender")
			mh.Del("Message-ID")
			log.Println("header is :", mh)
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
			log.Println("write buff body")
			buf.Write(byteHeader)
			buf.Write([]byte("\r\n"))
			byteBody, err4 := ioutil.ReadAll(goMail.Body)
			if err4 != nil {
				log.Fatal(err4)
			}
			buf.Write(byteBody)

			self.c.Append("INBOX", nil, msg.Envelope.Date, buf)

		}
	}
}

