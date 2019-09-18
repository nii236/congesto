package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS subscribers (
	first_name			VARCHAR(64) NOT NULL,
	last_name			VARCHAR(64) NOT NULL,
	user_name			VARCHAR(64) NOT NULL,
	chat_id	    		INTEGER NOT NULL,
	server_name 		VARCHAR(64) NOT NULL,
	creation_available	BOOL NOT NULL DEFAULT false,
	created_at			TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived			BOOL NOT NULL DEFAULT false,
	archived_at			TIMESTAMP,
	PRIMARY KEY (chat_id, server_name)
);
`

const drop = `
DROP TABLE IF EXISTS subscribers;
`

func initDB(dropDB bool) (*sqlx.DB, error) {
	conn, err := sqlx.Connect("sqlite3", "congesto.db")
	if err != nil {
		return nil, err
	}
	if dropDB {
		fmt.Println("Dropping database...")
		conn.MustExec(drop)
	}
	conn.MustExec(schema)
	return conn, nil
}
