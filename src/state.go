package main

import "database/sql"
import _ "github.com/mattn/go-sqlite3"
import "path/filepath"

var stateDb *sql.DB

func initStateDb() *sql.DB {
	dbfile := filepath.Join(CFG.StateDir, "imagestate.db")
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	stateDb = db
	initTables()
	return stateDb
}

func initTables() {
	stateDb.Exec("CREATE TABLE IF NOT EXISTS snapshot (run_id text, image_id text, snapshot_id integer);")
	stateDb.Exec("CREATE TABLE IF NOT EXISTS volume (image_id text, volume_id integer);")
	stateDb.Exec("CREATE VIEW IF NOT EXISTS descriptor AS select 'snapshot' as type, snapshot_id as id from snapshot UNION ALL select 'volume' as type, volume_id as id from volume;")
}

func generateVolumeId(imageId string) (int, error) {
	var volumeId int
	err := stateDb.QueryRow("INSERT INTO volume (image_id, volume_id) VALUES (?, (SELECT coalesce(max(id)+1, 0) FROM descriptor)) RETURNING volume_id", imageId).Scan(&volumeId)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return -1, err
	}
	return volumeId, nil
}

func getVolumeId(imageId string) (int, bool, error) {
	var volumeId int
	err := stateDb.QueryRow("SELECT volume_id FROM volume WHERE image_id = ?", imageId).Scan(&volumeId)
	if err == sql.ErrNoRows {
		return -1, false, nil
	} else if err != nil {
		return -1, false, err
	}
	return volumeId, true, nil
}

func deleteVolumeRecord(imageId string) error {
	_, err := stateDb.Exec("DELETE FROM volume WHERE image_id = ?", imageId)
	if err != nil {
		return err
	}
	return nil
}

func generateSnapshotId(runId string, imageId string) (int, error) {
	var snapshotId int
	err := stateDb.QueryRow("INSERT INTO snapshot (run_id, image_id, snapshot_id) VALUES (?, ?, (SELECT coalesce(max(id)+1, 0) FROM descriptor)) RETURNING snapshot_id", runId, imageId).Scan(&snapshotId)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return -1, err
	}
	return snapshotId, nil
}

func getSnapshotId(runId string) (int, bool, error) {
	var snapshotId int
	err := stateDb.QueryRow("SELECT snapshot_id FROM snapshot where run_id = ? LIMIT 1", runId).Scan(&snapshotId)
	if err == sql.ErrNoRows {
		return -1, false, nil
	} else if err != nil {
		return -1, false, err
	}
	return snapshotId, true, nil
}
