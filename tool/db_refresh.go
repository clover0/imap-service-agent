package main

import (
	"flag"
	"fmt"
	"github.com/develop/imap-agent/config"
	"github.com/develop/imap-agent/lib/db"
	"github.com/jmoiron/sqlx"
)

func main() {
	config.InitConfig()
	flag.Parse()
	args := flag.Args()
	
	_db := db.NewDB()
	tx := _db.MustBegin()

	fmt.Println("refresh db...")

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "trunc_senders":
			truncSenders(tx)
		default:
			fmt.Println("skip due to not configured ", args[i])
		}
	}
	tx.Commit()

}

func truncSenders(tx *sqlx.Tx) {
	fmt.Println("truncate senders table")

	sql := "TRUNCATE senders"
	tx.MustExec(sql)
}
