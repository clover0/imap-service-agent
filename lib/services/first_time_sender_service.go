package services

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
)

type FirstTimeSenderService struct {
	c *client.Client
}

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

	if len(ids) > 0 {
		seqset := new(imap.SeqSet)
		seqset.AddNum(ids...)

		messages := make(chan *imap.Message, 10)
		done := make(chan error, 1)
		go func() {
			done <- self.c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		log.Println("Unseen messages:")
		for msg := range messages {
			log.Println("subject: " + msg.Envelope.Subject)
			log.Println("uid: " + msg.Envelope.Subject)
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}
	}
}
