package main

import (
	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/idle"
)

func main(){
	config.InitConfig()
	idle.Run()
}
