package user

import (
    "fmt"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
)

func trafficUnit(t int64) (int64, string) {
    if t == 0 {
        return 0, "B"
    }
    if t < 1024 {
        return t, "B"
    }
    if t < 1024*1024 {
        return t / 1024, "KB"
    }
    if t < 1024*1024*1024 {
        return t / (1024 * 1024), "MB"
    }
    return t / (1024 * 1024 * 1024), "GB"
}

func list() (res []map[string]any, err error) {
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        err = fmt.Errorf("list() >> %w", err)
        return res, err
    }

    var rowCount int
    err = db.QueryRow("SELECT COUNT(*) from users;").Scan(&rowCount)
    config.Log.Debug("rowCount", "=", rowCount)
    if err != nil {
        return res, err
    }

    var rows *sql.Rows
    rows, err = db.Query(config.QUERY_USER_LIST)
    if err != nil {
        config.Log.Error("list", "db.Query", err)
        err = fmt.Errorf("list() / db.Query >> %w", err)
        return res, err
    }
    defer func(){
        errRows := rows.Close()
        if errRows != nil {
            err = fmt.Errorf("list() / rows.Close() %w", errRows)
        }
    }()

    fmt.Printf("%-15s %-20s %-10s %-10s %s\n", "username", "realname", "traffic", "session", "status")

    res = make([]map[string]any, rowCount, rowCount)

    for i := 0; rows.Next(); i++ {
        var username string
        var realname string
        var traffic int64
        var session int64
        var status string
        err = rows.Scan(&username, &realname, &traffic, &session, &status)
        if err != nil {
            err = fmt.Errorf("list() / rows.Next() %w", err)
            return res, err
        }
        tr, unit := trafficUnit(traffic)
        trafficStr := fmt.Sprintf("%d%s", tr, unit)
        fmt.Printf("%-15s %-20s %-10s %-10d %s\n", username, realname, trafficStr, session, status)

        res[i] = map[string]any{
            "user": username,
            "name": realname,
            "traffic": trafficStr,
            "session": session,
            "status": status,
        }
    }
    // fmt.Printf("res %+v\n", x)

    return res, nil
}
