package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS subscribers (
	id	    			INTEGER UNIQUE NOT NULL PRIMARY KEY,
	email				VARCHAR(64) NOT NULL,
	server				VARCHAR(64) NOT NULL,
	created_at			TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	archived			BOOL NOT NULL DEFAULT false,
	archived_at			TIMESTAMP
);
`

const drop = `
DROP TABLE IF EXISTS events;
`

func initDB(dropDB bool) (*sqlx.DB, error) {
	conn, err := sqlx.Connect("sqlite3", "congesto.db")
	if err != nil {
		return nil, err
	}
	if dropDB {
		conn.MustExec(drop)
	}
	conn.MustExec(schema)
	return conn, nil
}
