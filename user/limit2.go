package user

import (
    "fmt"
    "io"
    // "sync"
    //"time"
    "context"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
    // "github.com/shakibamoshiri/proxgo/tell"
)

func limit2(ctx context.Context, args []string, dev io.Writer) (err error) {
    config.Log.Debug("limit(args)", "=", args)

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("limit() >> %w", err)
    }

    // gel all rows
    chanRow := make(chan userData)
    go getRows2(db, chanRow)

    // delete if byte exceeded
    chanByte := make(chan userData)
    go byteCheck2(chanRow, chanByte)

    // delete if expired
    chanTime := make(chan userData)
    go timeCheck2(chanByte, chanTime)

    // notify if time will be expired
    chanTimeNotif := make(chan userData)
    go timeNotif2(chanTime, chanTimeNotif)

    // notify if byte will be used
    chanByteNotif := make(chan userData)
    go byteNotif2(chanTimeNotif, chanByteNotif)

    // print what remained
    chanDone := make(chan struct{})
    go chanPrint2(chanByteNotif, chanDone)
    <- chanDone

    return err
}


func getRows2(db *sql.DB, output chan userData) <- chan userData  {
    defer close(output)
    var rows *sql.Rows
    var err error

    var rowCount int64 = 0
    err = db.QueryRow("SELECT COUNT(*) from bytes;").Scan(&rowCount)
    if err != nil {
        config.Log.Error("limit()", "db.QueryRow()", err)
        output <- userData{
            err: err,
        }
        return output 
    }

    userRow := make([]userColumn, rowCount, rowCount)

    rows, err = db.Query("SELECT * FROM bytes;")
    if err != nil {
        config.Log.Error("limit()", "db.Query()", err)
        output <- userData{
            err: err,
        }
        return output 
    }
    defer func(){
        errClose := rows.Close()
        if errClose != nil {
            output <- userData{
                err: err,
            }
        }
    }()

    for i := 0; rows.Next(); i++ {
        err = rows.Scan(
            &userRow[i].username,
            &userRow[i].realname,
            &userRow[i].sessions,
            &userRow[i].ctime,
            &userRow[i].atime,
            &userRow[i].etime,
            &userRow[i].bytesBase,
            &userRow[i].bytesUsed,
            &userRow[i].bytesPday,
            &userRow[i].bytesLimit,
            &userRow[i].secondBase,
            &userRow[i].secondUsed,
            &userRow[i].secondLimit,
        )
            
        output <- userData{
            row: userRow[i],
            err: err,
        }
        config.Log.Debug("userRow", "=", userRow[i])
    }

    return output 
}

func byteCheck2(input <- chan userData, output chan userData) {
    defer close(output)
    for ch := range input {
        if (ch.row.bytesUsed >= ch.row.bytesBase) {
            output <- userData{
                row: ch.row,
                msg: "deleted (byte-limit)",
            }
            continue
        }
        output <- userData{
            row: ch.row,
            msg: "ignored",
        }
    }
}

func timeCheck2(input <- chan userData, output chan userData) {
    defer close(output)
    for ch := range input {
        if (ch.row.secondUsed >= ch.row.secondBase) {
            output <- userData{
                row: ch.row,
                msg: "deleted (time-limit)",
            }
            continue
        }
        output <- userData{
            row: ch.row,
            msg: "ignored",
        }
    }
}

func timeNotif2(input <- chan userData, output chan userData) {
    defer close(output)
    const oneDay = 24*60*60
    for ch := range input {
        if (ch.row.secondUsed + oneDay >= ch.row.secondBase) {
            output <- userData{
                row: ch.row,
                msg: "notified (time limit in 1d)",
            }
            continue
        }
        output <- userData{
            row: ch.row,
            msg: "ignored",
        }
    }
}

func byteNotif2(input <- chan userData, output chan userData) {
    defer close(output)
    const oneGig = 1 << 30
    for ch := range input {
        if (ch.row.bytesUsed + oneGig >= ch.row.bytesBase) {
            output <- userData{
                row: ch.row,
                msg: "notified (byte limit in 1d)",
            }
            continue
        }
        output <- userData{
            row: ch.row,
            msg: "ignored",
        }
    }
}

func chanPrint2(input <- chan userData, done chan struct{}) {
    for ch := range input {
        fmt.Printf("user.limit.%-20v %s\n", ch.row.username, ch.msg)
    }
    close(done)
}
