package main

import "database/sql"
import _ "github.com/mattn/go-sqlite3"
import "path/filepath"

var stateDb *sql.DB

func initStateDb() *sql.DB {
	dbfile := filepath.Join(CFG.stateDir, "imagestate.db")
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	stateDb = db
	initTables()
	return stateDb
}

func initTables() {
	stateDb.QueryRow("CREATE TABLE IF NOT EXISTS state (uuid text, image_id text, snapshot_id integer);")
}

func imageSnapshotId(image_id string) (int, bool) {
	var snapshotId int
	err := stateDb.QueryRow("SELECT TOP 1 snapshot_id FROM state where image_id = ?", image_id).Scan(&snapshotId)
	if err == sql.ErrNoRows {
		return -1, false
	} else if err != nil {
		panic("Couldn't query state database.")
	}
	return snapshotId, true
}
