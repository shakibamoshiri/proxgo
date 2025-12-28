package user

import (
    "fmt"
    "time"
    "io"

    "github.com/shakibamoshiri/proxgo/config"
)

func setup(args []string, dev io.Writer) (err error) {

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("setup() >> %w", err)
    }

    var rowCount int64 = 0
    err = db.QueryRow(`SELECT COUNT(*) from fetched where session > 0;`).Scan(&rowCount)
    if err != nil {
        config.Log.Error("setup()", "db.QueryRow", err)
        return fmt.Errorf("setup() / db.QueryRow %w", err)
    }

    // rows is new, err is reused
    rows, err := db.Query(config.QUERY_USER_SETUP)
    if err != nil {
        config.Log.Error("setup()", "db.Query", err)
        return fmt.Errorf("setup() / db.Query >> %w", err)
    }
    defer func(){
        errClose := rows.Close()
        if errClose != nil {
            err = errClose
        }
    }()

    type userFetch struct {
        username    string
        realname    string
        sessions     int64

        ctime       int64
        atime       int64
        etime       int64

        bytesBase   int64
        bytesUsed   int64
        bytesPday   int64
        bytesLimit  bool

        secondBase  int64
        secondUsed  int64
        secondLimit bool

        init        bool
    }

    userFetched := make([]userFetch, 0, rowCount)
    for rows.Next() {
        var u userFetch
        if err = rows.Scan(&u.username, &u.realname, &u.sessions, &u.bytesBase, &u.bytesUsed, &u.secondBase, &u.init); err != nil {
            return err
        }
        userFetched = append(userFetched, u)
    }

    userFetchedStr := fmt.Sprintf("%+v", userFetched)
    config.Log.Debug("rowCount", "=", rowCount)
    config.Log.Debug("userFetchedStr", "=",  userFetchedStr)

    ob := config.NewOutputBuffer()

    now := time.Now().Unix()
    for i, user := range userFetched {
        ob.Fprintf(dev, "%-30s", "user.setup." + user.username)
        if user.init == false {
            user.ctime = now
            user.atime = now
            user.etime = now + user.secondBase
            affectedUser := fmt.Sprintf("%d %+v\n",i, user)
            config.Log.Debug("affectedUser", "=", affectedUser)

            _, errExec := db.Exec(`
                INSERT INTO bytes
                (username, realname, sessions, ctime, atime, etime, bytes_base, bytes_used, bytes_pday, bytes_limit, seconds_base, seconds_used, seconds_limit)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
                user.username,
                user.realname,
                user.sessions,

                user.ctime,
                user.atime,
                user.etime,

                user.bytesBase,
                user.bytesUsed,
                user.bytesPday,
                user.bytesLimit,

                user.secondBase,
                user.secondUsed,
                user.secondLimit,
            )
            if errExec != nil {
                return errExec
            }
            ob.Fprintln(dev, "initialized")
        } else {
            var currentSessions int64
            err = db.QueryRow("SELECT sessions FROM bytes WHERE username = ?", user.username).Scan(&currentSessions)
            if err != nil {
                return err
            }

            if user.sessions > currentSessions {
                affectedUser := fmt.Sprintf("%d %+v\n",i, user)
                config.Log.Debug("affectedUser", "=", affectedUser)

                _, errExec := db.Exec(`
                    UPDATE bytes SET
                        sessions       = ?,
                        atime          = unixepoch(),
                        bytes_used     = ?,
                        bytes_pday     = (? / (unixepoch() - ctime)),
                        bytes_limit    = (? > bytes_base),
                        seconds_used   = (unixepoch() - ctime),
                        seconds_limit  = ((unixepoch() - ctime) > seconds_base)
                    WHERE username = ?`,
                    user.sessions,
                    user.bytesUsed,
                    user.bytesUsed,
                    user.bytesUsed,
                    user.username,
                )
                if errExec != nil {
                    return errExec
                }
                ob.Fprintln(dev, "updated")
            } else {
                ob.Fprintln(dev, "ignored")
            }

        }
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "user setup error!")
        ob.Stderr.Reset()
        ob.Fprintln(dev)
        ob.Flush()
        return err
    }

    ob.Flush()
    return nil
}
