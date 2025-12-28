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

func limit4(ctx context.Context, args []string, dev io.Writer) (err error) {
    config.Log.Debug("limit(args)", "=", args)

    done := make(chan struct{})
    getRows4().
    byteCheck4().
    timeCheck4().
    byteNotif4().
    timeNotif4().
    chanPrint4(done)

    return err
}



func getRows4() LimitPipe  {
    output := make(chan userData)
    go func(){
        defer close(output)
        var err error
        dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
        db, err := config.OpenDB(dbFile)
        if err != nil {
            config.Log.Error("limit()", "config.OpenDB()", err)
            output <- userData{
                err: err,
            }
        }

        var rows *sql.Rows
        var rowCount int64 = 0
        err = db.QueryRow("SELECT COUNT(*) from bytes;").Scan(&rowCount)
        if err != nil {
            config.Log.Error("limit()", "db.QueryRow()", err)
            output <- userData{
                err: err,
            }
        }

        userRow := make([]userColumn, rowCount, rowCount)

        rows, err = db.Query("SELECT * FROM bytes;")
        if err != nil {
            config.Log.Error("limit()", "db.Query()", err)
            output <- userData{
                err: err,
            }
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

    }()
    return LimitPipe{next: output}
}

func (input LimitPipe) byteCheck4() LimitPipe {
    output := make(chan userData)
    go func(){
        defer close(output)
        for ch := range input.next {
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
    }()
    return LimitPipe{next: output}
}

func (input LimitPipe) timeCheck4() LimitPipe { 
    output := make(chan userData)
    go func(){
        defer close(output)
        for ch := range input.next {
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
    }()
    return LimitPipe{next: output}
}

func (input LimitPipe) timeNotif4() LimitPipe {
    output := make(chan userData)
    go func(){
        defer close(output)
        const oneDay = 24*60*60
        for ch := range input.next {
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
    }()
    return LimitPipe{next: output}
}

func (input LimitPipe) byteNotif4() LimitPipe {
    output := make(chan userData)
    go func(){
        defer close(output)
        const oneGig = 1 << 30
        for ch := range input.next {
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
    }()
    return LimitPipe{next: output}
}

func (input LimitPipe)chanPrint4(done chan struct{}) {
    for ch := range input.next {
        fmt.Printf("user.limit.%-20v %s\n", ch.row.username, ch.msg)
    }
    close(done)
}
