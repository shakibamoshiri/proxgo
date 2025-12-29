package user

import (
    "fmt"
    "time"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
)

func stats(pools *config.Pools) (err error) {
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

    serverCount := int64(len(pools.Servers))
    config.Log.Debug("serverCount", "=", serverCount)
    fmt.Printf("%-15s %-20s %-10s %-10s %-10s %-10s %-10s %-10s %s\n",
        "username", "realname", "elapsed", "remained",
        "expired", "traffic", "perDay", "sessions", "spend")

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

        now := time.Now().Unix()
        elapsed := formatDurationDH(now - row.ctime)
        remained := formatDurationDH(row.secondBase - row.secondUsed)
        expired := (now > row.etime)
        days := ((now - row.ctime) / 86400)
        if days == 0 {
            days = 1
        }
        spend := formatDurationHM(row.sessions / days / serverCount * 5 * 60)

        fmt.Printf("%-15s %-20s %-10s %-10s %-10t %-10s %-10s %-10d %s\n",
            row.username, row.realname, elapsed, remained,
            expired, traffic, pday, row.sessions, spend)
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
    secondsPerMinute = 60
)

func formatDurationDH(seconds int64) string {
    if seconds < 0 {
        return "0d0h"
    }

    days := seconds / secondsPerDay
    remainingSeconds := seconds % secondsPerDay
    hours := remainingSeconds / secondsPerHour

    return fmt.Sprintf("%dd:%dh", days, hours)
}

func formatDurationHM(seconds int64) string {
    if seconds <= 0 {
        return "0h0m"
    }

    hours := seconds / secondsPerHour
    minutes := (seconds % secondsPerHour) / secondsPerMinute

    return fmt.Sprintf("%dh%dm", hours, minutes)
}

func formatDurationDHM(seconds int64) string {
    if seconds <= 0 {
        return "0d0h0m"
    }

    days := seconds / secondsPerDay
    remaining := seconds % secondsPerDay

    hours := remaining / secondsPerHour
    remaining %= secondsPerHour

    minutes := remaining / secondsPerMinute

    return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
}
