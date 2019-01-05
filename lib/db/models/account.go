package models

import "database/sql"

type Account struct {
	Id               int64          `db:"id"`
	Created_at       sql.NullString `db:"created_at"`
	Updated_at       sql.NullString `db:"update_at"`
	ImapHost         sql.NullString `db:"imap_host"`
	ConnectionMethod sql.NullString `db:"connection_method"`
	ImapPort         sql.NullString `db:"imap_port"`
	AccountName      sql.NullString `db:"account_name"`
	MailAddress      sql.NullString `db:"mail_address"`
	Password         sql.NullString `db:"password"`
}
