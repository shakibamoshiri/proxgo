//go:build deprecated

package user

import (
    "fmt"
    "io"
    "sync"
    "time"
    "context"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/tell"
)


func limit(ctx context.Context, args []string, dev io.Writer) (err error) {
    config.Log.Debug("limit(args)", "=", args)

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("limit() >> %w", err)
    }

    var rowCount int64 = 0
    var rows *sql.Rows
    err = db.QueryRow("SELECT COUNT(*) from bytes;").Scan(&rowCount)
    if err != nil {
        config.Log.Error("limit()", "db.QueryRow()", err)
        return fmt.Errorf("limit() / db.QueryRow() %w", err)
    }

    rows, err = db.Query("SELECT * FROM bytes;")
    if err != nil {
        config.Log.Error("limit()", "db.Query()", err)
        return fmt.Errorf("limit() / db.Query() %w", err)
    }
    defer func(){
        errClose := rows.Close()
        if errClose != nil {
            err = errClose
        }
    }()

    userRow := make([]userColumn, rowCount, rowCount)
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
        if err != nil {
            return err
        }
    }

    userFetchedStr := fmt.Sprintf("%+v", userRow)
    config.Log.Info("active users (affected)", "=", rowCount)
    config.Log.Debug("userFetchedStr", "=", userFetchedStr)

    ob := config.NewOutputBuffer()

    const oneDay = 24*60*60
    const oneGig = 1<<30
    var waitForTell sync.WaitGroup
    const limitTimeout time.Duration = 5
    ctxRun, cancel := context.WithTimeout(context.Background(), limitTimeout * time.Second)
    defer cancel()
    done := make(chan struct{})

    for _, user := range userRow {
        ob.Fprintf(dev, "%-30s", "user.limit." + user.username)
        // if user.bytesLimit {
        if (user.bytesUsed >= user.bytesBase) {
            // delete([]string{"-user","xyz"})
            // println("delete: bytesLimit")
            nextArgs := []string{"-user", user.username}
            err = _delete(nextArgs)
            if err != nil {
                return fmt.Errorf("limit() / delete user %w", err)
            }
            err = archive(nextArgs)
            if err != nil {
                return fmt.Errorf("limit() / archive(%v) user %w", user.username, err)
            }
            ob.Fprintln(dev, "limited (bytesLimit)")
            config.Log.Warn("username", user.username, "delete (bytesLimit)")
            continue
        }

        // if user.secondLimit {
        if (user.secondUsed >= user.secondBase) {
            // println("delete: timeLimit")
            nextArgs := []string{"-user", user.username}
            err = _delete(nextArgs)
            if err != nil {
                return fmt.Errorf("limit() / delete(%v) user %w", user.username,  err)
            }
            err = archive(nextArgs)
            if err != nil {
                return fmt.Errorf("limit() / archive(%v) user %w", user.username, err)
            }
            ob.Fprintln(dev, "limited (timeLimit)")
            config.Log.Warn("username", user.username, "delete (timeLimit)")
            continue
        }

        if (user.secondUsed + oneDay >= user.secondBase) {
            if user.secondLimit == true {
                ob.Fprintln(dev, "notified")
                continue
            }
            ob.Fprintln(dev, "notify (expired in 1d)")
            msg := fmt.Sprintf("<b>alert</b> username <code>%s</code> will be expired in 1d", user.username)

            waitForTell.Add(1)

            go func(){
                defer waitForTell.Done()
                username := user.username
                errTell := tell.Fire("notify", ctxRun, msg)
                if errTell == nil {
                    _, errExec := db.ExecContext(ctxRun, `UPDATE bytes SET seconds_limit = true WHERE username = ? AND seconds_limit = false`, username)
                    if errExec != nil {
                        config.Log.Error("db.ExecContext to UPDATE failed", "username", username, "error", errExec)
                        err = fmt.Errorf("db.ExecContext to UPDATE username %s failed  %w", username, errExec)
                        return
                    }
                    config.Log.Info("db.Exec to UPDATE succeeded", "username", username)
                } else {
                    config.Log.Warn("tell.Fire failed, db.Exec UPDATE ignored", "username", username)
                }
            }()
            continue
        }

        if (user.bytesUsed + oneGig >= user.bytesBase) {
            ob.Fprintln(dev, "notify (limited in 1d)")
            config.Log.Info("notify (limited in 1d)", "username", user.username)
        }

        ob.Fprintln(dev, "ignored")
        config.Log.Info("username", user.username, "ignored")
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "user limit failure!")
        ob.Stderr.Reset()
        ob.Fprintln(dev)
        ob.Flush()
        return err
    }
    ob.Flush()


    go func(){
        waitForTell.Wait()
        close(done)
    }()

    started := time.Now()
    select {
    case <-ctx.Done():
        config.Log.Error("limit", "timeout", time.Since(started))
        return fmt.Errorf("limit timeout %w", ctx.Err())
    case <-done:
        config.Log.Info("limit / waitForTell() done", "error", err)
        return err
    }

    return err
}
