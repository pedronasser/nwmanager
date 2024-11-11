package database

import (
	"database/sql"
)

func NewLocalSqlite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./local.db")
	if err != nil {
		return nil, err
	}

	sqlStmt := `
		create table foo (id integer not null primary key, name text);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}
