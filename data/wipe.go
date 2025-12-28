package data

import (
	"fmt"
    "os"

	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/shakibamoshiri/proxgo/config"
)

func wipe(args []string) (err error) {
	config.Log.Debug("args", "=", args)

	var confirm string
	fmt.Print("YOU ARE 100% SURE WIPING ALL YOUR DATA? type YES: ")
	fmt.Scan(&confirm)
	if confirm == "YES" {
		config.Log.Warn("wiping all data confirmed")
	} else {
		fmt.Println("wiping cancelled")
        os.Exit(0)
	}

	dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
	db, err := config.OpenDB(dbFile)
	if err != nil {
		return err
	}

	tables := [tableMax]*table{
		{name: "users", cmd: config.TABLE_DELETE_USERS},
		{name: "bytes", cmd: config.TABLE_DELETE_BYTES},
		{name: "archive", cmd: config.TABLE_DELETE_ARCHIVE},
		{name: "fetched", cmd: config.TABLE_DELETE_FETCHED},
	}

	for i := 0; i < tableMax; i++ {
		fmt.Printf("%-30s", "wipe.table."+tables[i].name)

		err := tables[i].wipeTable(db)
		if err != nil {
			return err
		}

		fmt.Printf("%s\n", "wiped")
	}

	return nil
}

func (t *table) wipeTable(db *sql.DB) (err error) {
	stmt, errPer := db.Prepare(t.cmd)
	if errPer != nil {
		config.Log.Error("wipe()", "db.Prepare error", errPer)
		return errPer
	}
	defer func() {
		errClose := stmt.Close()
		if errClose != nil {
			err = errClose
		}
	}()

	_, errExec := stmt.Exec()
	if errExec != nil {
		config.Log.Error("wipe()", "stmt.Exec error", errExec)
		return errExec
	}

	return nil
}
