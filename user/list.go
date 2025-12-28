package user

import (
    "fmt"
    "time"

    "database/sql"
    _ "modernc.org/sqlite"

    "github.com/shakibamoshiri/proxgo/config"
)

func trafficUnit(t int64) (int64, string) {
    if t == 0 {
        return 0, "b"
    }
    if t < 1024 {
        return t, "b"
    }
    if t < 1024*1024 {
        return t / 1024, "k"
    }
    if t < 1024*1024*1024 {
        return t / (1024 * 1024), "m"
    }
    return t / (1024 * 1024 * 1024), "g"
}

func list() (err error) {


    /// fmt.Printf("dbFile %s\n", dbFile)
    /// fmt.Printf("os.Args %s\n", os.Args)

    /// arg := config.NewArg().Setup()
    /// aid := arg.Find("aid").Int()
    /// log := arg.Find("log").Default("info").String()
    /// fmt.Printf("aid %v\n", aid)
    /// fmt.Printf("log %v\n", log)

    /// arg2 := config.NewArg()
    /// fmt.Printf("arg p  %p\n", arg)
    /// fmt.Printf("arg2 p  %p\n", arg2)
    /// return 

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("list() >> %w", err)
    }
    // defer func(){
    //     errClose := db.Close()
    //     if errClose != nil {
    //         err = fmt.Errorf("list() / db.Close() %w", errClose)
    //     }
    // }()

    var rows *sql.Rows
    rows, err = db.Query(config.QUERY_USER_LIST)
    if err != nil {
        config.Log.Error("list", "db.Query", err)
        return fmt.Errorf("list() / db.Query >> %w", err)
    }
    defer func(){
        errRows := rows.Close()
        if errRows != nil {
            err = fmt.Errorf("list() / rows.Close() %w", errRows)
        }
    }()

    fmt.Printf("%-15s %-20s %-10s %-10s %-10s %s\n", "username", "realname", "elapsed", "traffic", "session", "status")

    for rows.Next() {
        var username string
        var realname string
        var ctime int64
        var etime int64
        var traffic int64
        var session int64
        var status string
        //var started time.Time
        err = rows.Scan(&username, &realname, &ctime, &traffic, &session, &status)
        if err != nil {
            return fmt.Errorf("list() / rows.Next() %w", err)
        }
        tr, unit := trafficUnit(traffic)
        trafficStr := fmt.Sprintf("%d%s", tr, unit)
        past := time.Unix(ctime, 0)
        now := time.Now()
        elapsed := now.Sub(past)
        days := int(elapsed.Hours() / 24)
        hours := int(elapsed.Hours()) % 24
        // minutes := int(elapsed.Minutes()) % 60
        // seconds := int(elapsed.Seconds()) % 60
        elapsedFormated := fmt.Sprintf("%dd%dh", days, hours)
        fmt.Printf("%-15s %-20s %-10s %-10s %-10d %s\n", username, realname, elapsedFormated, trafficStr, session, status)
    }

    return nil
}
