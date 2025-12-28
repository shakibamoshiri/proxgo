package data

import (
	"fmt"
	"io"
    "os"

	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/shakibamoshiri/proxgo/config"
)

func setup(args []string, pc *config.Pools, dev io.Writer) (err error) {
	config.Log.Debug("args", "=", args)

	dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
	db, err := config.OpenDB(dbFile)
	if err != nil {
		return fmt.Errorf("setup >> %w", err)
	}

    ob := config.NewOutputBuffer()

	tables := [tableMax]*table{
		{name: "users", cmd: config.TABLE_CREATE_USERS},
		{name: "bytes", cmd: config.TABLE_CREATE_BYTES},
		{name: "archive", cmd: config.TABLE_CREATE_ARCHIVE},
		{name: "fetched", cmd: config.TABLE_CREATE_FETCHED},
	}

	for i := 0; i < tableMax; i++ {
		ob.Fprintf(dev, "%-30s", "setup.table."+tables[i].name)

		err := tables[i].create(db)
		if err != nil {
            ob.Fprintln(dev, err)
            ob.Fprintln(os.Stderr, err)
            err = nil
            continue
		}

        ob.Fprintln(dev, "created")
	}

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "data setup failed!")
        ob.Stderr.Reset()
        ob.Flush()
        return err
    }

    ob.Flush()
	return nil
}

func (t *table) create(db *sql.DB) (err error) {
	stmt, errPer := db.Prepare(t.cmd)
	if errPer != nil {
		config.Log.Error("setup()", "db.Prepare error", errPer)
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
		config.Log.Error("setup()", "stmt.Exec error", errExec)
		return errExec
	}

	return nil
}
