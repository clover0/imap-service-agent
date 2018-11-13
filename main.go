package main

import (
	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib"
)

func main(){
	config.InitConfig()

	//lib.Do()
	//lib.Sample()
	//lib.SearchSample()
	lib.RunIdleDaemon()
}
