package sqlite

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var maxOldEl = time.Hour

const (
	insertSQL = `
INSERT INTO previews (
	url, created_at, file
) VALUES (
	?, ?, ?
)
`

	schemaSQL = `
CREATE TABLE IF NOT EXISTS previews (

url TEXT PRIMARY KEY NOT NULL,

created_at TIMESTAMP,

file BLOB NOT NULL

);
CREATE INDEX IF NOT EXISTS trades_time ON previews(created_at);`

	selectUrlSQL = `SELECT * FROM previews
    WHERE url == ?`

	deleteElSQL = `
	 DELETE FROM previews
 WHERE created_at <= ?`
)

type PreviewTableModel struct {
	URL  string
	Time time.Time
	File []byte
}

type DB struct {
	sql  *sql.DB
	stmt *sql.Stmt
}

func NewDB(dbFile string) (*DB, error) {
	fmt.Println(os.Args[0])
	sqlDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("in NewDB can`t open file %s: %w", dbFile, err)
	}

	if _, err = sqlDB.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("in NewDB can`t sqlDB.Exec(schemaSQL): %w", err)
	}

	stmt, err := sqlDB.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("in NewDB can`t sqlDB.Prepare(insertSQL): %w", err)
	}

	db := DB{
		sql:  sqlDB,
		stmt: stmt,
	}
	return &db, nil
}

func (db *DB) Add(elDB *PreviewTableModel) error {
	_, err := db.sql.Exec(insertSQL, elDB.URL, elDB.Time, elDB.File)
	if err != nil {
		return fmt.Errorf("in Add can`t Exec(): %w", err)
	}

	return nil
}

func (db *DB) Close() error {
	err := db.stmt.Close()
	if err != nil {
		return fmt.Errorf("in Close() can`t db.stmt.Close(): %w", err)
	}

	err = db.sql.Close()
	if err != nil {
		return fmt.Errorf("in Close() can`t db.sql.Close(): %w", err)
	}

	return nil
}

func (db *DB) Select(url string) (*PreviewTableModel, error) {
	resElDb := &PreviewTableModel{}
	row := db.sql.QueryRow(selectUrlSQL, url)

	err := row.Scan(&resElDb.URL, &resElDb.Time, &resElDb.File)
	if err != nil {
		return nil, fmt.Errorf("in Select can`t row.Scan(): %w", err)
	}

	return resElDb, nil
}

func (db *DB) Clearing(period time.Duration) {
	tic := time.NewTicker(period)
	for {
		select {
		case <-tic.C:
			_, err := db.sql.Exec(deleteElSQL, time.Now().Add(-1*maxOldEl))
			if err != nil {
				log.Printf("in Add can`t Exec(): %v", err)
			}
		}
	}
}
