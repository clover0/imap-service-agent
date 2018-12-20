package main

import (
	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/idle"
)

func main() {
	config.InitConfig()

	// improve: run every account.
	// e.g)
	// for 10 times { run ... }
	idle.Run()
}
