-- +migrate Up
CREATE TABLE accounts (
id serial PRIMARY KEY,
created_at timestamp,
update_at timestamp,
imap_host varchar(500) NULL,
connection_method varchar(16) NULL,
imap_port smallint,
account_name varchar(256) NULL ,
mail_address varchar(500) NULL ,
password varchar(256) NULL 
);
