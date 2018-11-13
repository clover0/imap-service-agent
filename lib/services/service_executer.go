package services

import "github.com/emersion/go-imap/client"

func Execute(c *client.Client) {
	ftss := FirstTimeSenderService{c}
	ftss.execute()
	
}