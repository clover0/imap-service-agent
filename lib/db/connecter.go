package db

import (
	"fmt"
	"github.com/develop/imap-agent/config"
	"github.com/jmoiron/sqlx"
	"log"
	_ "github.com/lib/pq"

)

func NewDB() *sqlx.DB{
	user := config.Conf.DB.User
	host := config.Conf.DB.Host
	port := config.Conf.DB.Port
	password := config.Conf.DB.Password
	dbName := config.Conf.DB.DBName
	
	ds :=fmt.Sprintf(
		"user=%s host=%s port=%d password=%s dbname=%s sslmode=disable",
		user,
		host,
		port,
		password,
		dbName,
	)
	log.Println("db config: ", ds)
	db, err := sqlx.Connect("postgres", ds)
	if err != nil {
		log.Fatalln(err)
	}
	return db
}