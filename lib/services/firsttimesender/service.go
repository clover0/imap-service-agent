package firsttimesender

import (
	"bufio"
	"bytes"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap-move"
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	go_mail "net/mail"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/develop/imap-agent/config"
)

const (
	serviceName string = "first_time_sender"
	CRLF        string = "\r\n"
	TRASH       string = "[Gmail]/ゴミ箱" // gmail
	//TRASH string = "Trash" // yahoo
)

type FirstTimeSenderService struct {
	Cli      *client.Client
	ImapConf *config.IMAPConfig
	Db       *sqlx.DB
}

func (self *FirstTimeSenderService) BeforeService() {
	log.Println("before service")
}

// DoService executes check if first or already email sender, and does service.
func (self *FirstTimeSenderService) DoService() {
	log.Printf("%s service starting process...", serviceName)

	cri := imap.NewSearchCriteria()
	cri.WithoutFlags = []string{imap.SeenFlag}

	log.Println("search mails...")
	ids, searchErr := self.Cli.Search(cri)
	if searchErr != nil {
		log.Fatal(searchErr)
	}
	log.Println("IDs found:", ids)

	// improve: use UID
	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)

		messages := make(chan *imap.Message, 10)
		section := &imap.BodySectionName{Peek: true}
		done := make(chan error, 1)
		go func() {
			done <- self.Cli.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchFlags, imap.FetchInternalDate, section.FetchItem()}, messages)
		}()

		log.Println("Unseen messages:")
		for msg := range messages {
			r := msg.GetBody(section)

			// copy reader
			var copyBuf bytes.Buffer
			tee := io.TeeReader(r, &copyBuf)

			// read header for rewrite header
			goMail, readMsgErr := go_mail.ReadMessage(tee)
			if readMsgErr != nil {
				log.Fatal(readMsgErr)
			}
			tp := textproto.NewReader(bufio.NewReader(&copyBuf))
			mh, _ := tp.ReadMIMEHeader()

			s := mh.Get("X-imap-agent-serviced")
			if strings.Contains(s, serviceName) {
				continue
			}

			fromAddress := msg.Envelope.From[0].MailboxName + "@" + msg.Envelope.From[0].HostName
			isFoundSenderInfo := self.findOrInsert(fromAddress, self.ImapConf.User)
			if isFoundSenderInfo {
				log.Println("skip append mail.")
				continue
			}

			mh.Set("Subject", "[初回送信者]"+msg.Envelope.Subject)
			mh.Set("X-imap-agent-serviced", serviceName)
			mh.Set("Message-ID", newMessageId(mh.Get("Message-ID")))
			//mh.Del("Date")
			//mh.Del("X-Gm-Message-State")
			//mh.Del("X-Google-Smtp-Source")
			//mh.Del("Delivered-To")
			//mh.Del("From")
			//mh.Del("To")

			// build header from header map
			header := ""
			for k, v := range mh {
				header = header + k + ": " + v[0] + CRLF
			}
			byteHeader, readHeaderErr := ioutil.ReadAll(strings.NewReader(header))
			if readHeaderErr != nil {
				log.Fatal(readHeaderErr)
			}

			// for building body
			buf := new(bytes.Buffer)

			buf.Write(byteHeader)
			buf.Write([]byte(CRLF))

			byteBody, err4 := ioutil.ReadAll(goMail.Body)
			if err4 != nil {
				log.Fatal(err4)
			}
			buf.Write(byteBody)

			date := msg.Envelope.Date

			seqset := new(imap.SeqSet)
			seqset.AddNum(msg.SeqNum)
			//log.Println("copy original mail to trash: ", msg.SeqNum)
			//self.c.Copy(seqset, "[Gmail]/ゴミ箱")

			// TODO strategy. gmail or yahoo or ...
			log.Println("delete original mail: ", msg.SeqNum)
			deleteitem := imap.FormatFlagsOp(imap.AddFlags, true)
			deleteflags := []interface{}{imap.DeletedFlag}
			// 単純削除ではなくてゴミ箱へ移動してそこで削除
			//self.Cli.Store(seqset, deleteitem, deleteflags, nil)

			moveCli := move.NewClient(self.Cli)
			moveSeq := new(imap.SeqSet)
			moveSeq.AddNum(msg.SeqNum)
			log.Printf("uid: %d", msg.Uid)
			moveCli.Move(moveSeq, TRASH)
			self.Cli.Select(TRASH, false)
			se := new(imap.SeqSet)
			se.AddNum(msg.Uid)
			self.Cli.UidStore(se, deleteitem, deleteflags, nil)

			self.Cli.Expunge(nil)

			// 元のボックスにもどっておく
			self.Cli.Select(self.ImapConf.MailBox, false)

			log.Println("append mail.")
			self.Cli.Append(self.ImapConf.MailBox, []string{imap.RecentFlag}, date, buf)
		}
	}
}

func (self *FirstTimeSenderService) AfterService() {
	log.Println("after service")
}

// findOrInsert finds or insert new record about mail sender info.
func (self *FirstTimeSenderService) findOrInsert(fromAddress string, account string) (found bool) {
	log.Println("find or insert sender info.")
	var count int

	tx := self.Db.MustBegin()

	err := tx.Get(&count, "SELECT COUNT(*) FROM senders WHERE mail_address = $1 AND to_account = $2", fromAddress, account)
	if err != nil {
		log.Fatal("select count error. err: ", err)
	}
	found = false
	if count > 0 {
		log.Println("found sender info")
		found = true
	} else { // if not exist sender info, insert new record
		tx.MustExec("INSERT INTO senders (mail_address, to_account, send_datetime) VALUES ($1, $2, current_timestamp)", fromAddress, account)
		log.Println("insert new record to sender table.")
		found = false
	}

	tx.Commit()
	log.Println("find or insert complete.")

	return found
}

// newMessageId creates a new MessageId joining original MessageId
func newMessageId(original string) string {
	r := strings.Replace(original, "<", "", -1)
	r = strings.Replace(r, ">", "", -1)

	ra := strings.Split(r, "@")
	ra[0] = ra[0] + strconv.Itoa(rand.Intn(10000000))

	r = strings.Join(ra, "@")

	return "<" + r + ">"
}
