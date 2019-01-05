package services

import (
	"github.com/emersion/go-imap/client"
	"github.com/jmoiron/sqlx"

	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/services/firsttimesender"
)

type Executor struct {}

func (self Executor) Execute (service Service) {
	service.BeforeService()
	service.DoService()
	service.AfterService()
}

// Execute executes all services
func Execute(c *client.Client, ic *config.IMAPConfig, db *sqlx.DB) {
	e := Executor{}
	ftss := &firsttimesender.FirstTimeSenderService{c,ic,db}
	e.Execute(ftss)
}
