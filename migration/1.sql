-- +migrate Up
CREATE TABLE senders (
mail_address varchar(500) NULL ,
to_account varchar(500) NULL ,
send_datetime timestamp
);

