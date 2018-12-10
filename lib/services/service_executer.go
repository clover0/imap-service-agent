package services

import (
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
)

func Execute(c *client.Client, db *sqlx.DB) {
	ftss := FirstTimeSenderService{c, db}
	ftss.execute()
}