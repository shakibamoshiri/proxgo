package user

import (
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)

func page() (res []map[string]any, err error) {

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        err = fmt.Errorf("config / OpenDB(%s) >> %w", dbFile,  err)
        return res, err
    }

    var rowCount int
    err = db.QueryRow("SELECT COUNT(*) from users;").Scan(&rowCount)
    if err != nil {
        return res, err
    }
    config.Log.Debug("rowCount", "=", rowCount)

    stmt, errPer := db.Prepare(`SELECT * FROM users;`)
    if errPer != nil {
        config.Log.Error("db.Prepare", "error", errPer)
        err = fmt.Errorf("config / db.Prepare() >> %w", errPer)
        return res, err
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = fmt.Errorf("config / stmt.Close() >> %w", errClose)
        }
    }()

    rows, err := stmt.Query(`SELECT * FROM users;`)
    if err != nil {
        config.Log.Error("stmt.QueryRow failed", "error", err)
        err = fmt.Errorf("config / stmt.QueryRow() failed %w", err)
        return res, err
    }

    users := make([]User, rowCount, rowCount)
    res = make([]map[string]any, rowCount, rowCount)
    for i := 0; rows.Next(); i++ {
        rows.Scan(
            &users[i].Username,
            &users[i].Realname,
            &users[i].Ctime,
            &users[i].Period,
            &users[i].Traffic,
            &users[i].Password,
            &users[i].Page,
            &users[i].Profile,
            &users[i].Device,
        )
    }
    config.Log.Info("done", "number of users =", rowCount)

    agents, err := yaml.Agents.Load()
    activePoolId := agents.Agent.PoolID
    pools, _ := yaml.Pools.Load(activePoolId)
    activeInfoIndex := yaml.ActiveInfoIndex()
    groupName := pools.DB.Info[activeInfoIndex].Name
    address := pools.DB.Info[activeInfoIndex].Profile.Link

    var link string
    for i := 0; i < rowCount; i++ {
        link = fmt.Sprintf("%s/%s/%s", address, groupName, users[i].Page)
        fmt.Printf("%s - %s - %s\n", link, users[i].Username, users[i].Realname)
        res[i] = map[string]any{
            "link": link,
            "user": users[i].Username,
            "name": users[i].Realname,
        }
    }

    return res, nil
}
