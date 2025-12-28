package user

import (
    "fmt"
    "time"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
)

func stats() (err error) {
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("stats() >> %w", err)
    }

    var rows *sql.Rows
    rows, err = db.Query(config.QUERY_USER_STATS)
    if err != nil {
        config.Log.Error("stats", "db.Query", err)
        return fmt.Errorf("stats() / db.Query >> %w", err)
    }
    defer func(){
        errRows := rows.Close()
        if errRows != nil {
            err = fmt.Errorf("stats() / rows.Close() %w", errRows)
        }
    }()

    fmt.Printf("%-15s %-20s %-10s %-10s %-10s %-10s %-10s %s\n",
        "username", "realname", "elapsed", "remained", "expired", "traffic", "perDay", "sessions")

    row := userColumn{}
    for rows.Next() {
        //var started time.Time
        // var status string
        err = rows.Scan(
            &row.username,
            &row.realname,
            &row.sessions,
            &row.ctime,
            &row.etime,
            &row.bytesUsed,
            &row.bytesPday,
            &row.secondBase,
            &row.secondUsed,
        )
        if err != nil {
            return fmt.Errorf("stats() / rows.Next() %w", err)
        }
        tr, unit := trafficUnit(row.bytesUsed)
        traffic := fmt.Sprintf("%d%s", tr, unit)

        pd, unit := trafficUnit(row.bytesPday)
        pday := fmt.Sprintf("%d%s", pd, unit)

        // elapsed := elapsedFmt(row.ctime)
        now := time.Now().Unix()
        elapsed := fmtDuration(now - row.ctime)
        remained := fmtDuration(row.secondBase - row.secondUsed)
        expired := (now > row.etime)

        fmt.Printf("%-15s %-20s %-10s %-10s %-10t %-10s %-10s %d\n",
            row.username, row.realname, elapsed, remained, expired, traffic, pday, row.sessions)
    }

    return nil
}

func elapsedFmt(v int64) string {
    past := time.Unix(v, 0)
    now := time.Now()
    elapsed := now.Sub(past)
    days := int(elapsed.Hours() / 24)
    hours := int(elapsed.Hours()) % 24
    // minutes := int(elapsed.Minutes()) % 60
    // seconds := int(elapsed.Seconds()) % 60
    return fmt.Sprintf("%dd%dh", days, hours)
}

const (
    secondsPerDay  = 24 * 60 * 60
    secondsPerHour = 60 * 60
)

func fmtDuration(seconds int64) string {
    if seconds < 0 {
        seconds = 0
    }

    days := seconds / secondsPerDay
    remainingSeconds := seconds % secondsPerDay
    hours := remainingSeconds / secondsPerHour

    return fmt.Sprintf("%dd:%dh", days, hours)
}
