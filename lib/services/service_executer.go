package services

import (
	"github.com/develop/imap-agent/config"
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"
)

func Execute(c *client.Client, ic config.IMAPConfig, db *sqlx.DB) {
	ftss := FirstTimeSenderService{
		c,
		ic,
		db,
	}
	ftss.execute()
}
