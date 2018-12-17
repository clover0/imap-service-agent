package services

import (
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"

	"github.com/develop/imap-agent/config"
)

func Execute(c *client.Client, ic config.IMAPConfig, db *sqlx.DB) {
	ftss := FirstTimeSenderService{
		c,
		ic,
		db,
	}
	ftss.execute()
}
