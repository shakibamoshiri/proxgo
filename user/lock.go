package user

import (
    "os"
    "flag"
    "fmt"
    "io"
    "encoding/json"
    // "net/http"
    "bytes"
    // "time"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/server"
    "github.com/shakibamoshiri/proxgo/httpx"
)

func lock(args[] string) (res []map[string]any, err error) {
    config.Log.Debug("args", "=", args)

    flags := flag.NewFlagSet("lock", flag.ExitOnError)
    var username string
    flags.StringVar(&username, "user", "", "username to lock")
    flags.Parse(args)


    if username == "" {
        println("user lock args:")
        flags.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("-user", "=", username)

////////////////////////////////////////////////////////////////////////////////
// get the user from database
////////////////////////////////////////////////////////////////////////////////
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        err = fmt.Errorf("lock >> %w", err)
        return res, err
    }

    stmt, errPer := db.Prepare(` SELECT * FROM users WHERE username = ?`)
    if errPer != nil {
        config.Log.Error("lock", "db.Prepare", errPer)
        err = fmt.Errorf("lock >> %w", errPer)
        return res, err
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = fmt.Errorf("lock >> %w", errClose)
        }
    }()

    newUser := User{}

    errExec := stmt.QueryRow(username).Scan(
        &newUser.Username,
        &newUser.Realname,
        &newUser.Ctime,
        &newUser.Period,
        &newUser.Traffic,
        &newUser.Password,
        &newUser.Page,
        &newUser.Profile,
        &newUser.Device,
    )
    if errExec != nil {
        config.Log.Error("lock / stmt.QueryRow for", username, errExec)
        err = fmt.Errorf("lock / stmt.QueryRow(%v) >> %w", username, errExec)
        return res, err
    }

    config.Log.Info("newUser", "=",  newUser)
    config.Log.Info("newUser.Username", "=", newUser.Username)

////////////////////////////////////////////////////////////////////////////////
// check if servers are up?
////////////////////////////////////////////////////////////////////////////////
    serverArgs := make([]string, 0, 0)
    err = server.Run("check", serverArgs, &yaml.Pools, io.Discard)
    if err != nil {
        return res, err
    }

////////////////////////////////////////////////////////////////////////////////
// prepare json data
////////////////////////////////////////////////////////////////////////////////
    activePoolIndex := yaml.ActivePoolIndex()
    activeInfoIndex := yaml.ActiveInfoIndex()
    config.Log.Info("yaml.ActivePoolIndex", "=", activePoolIndex)
    config.Log.Info("yaml.ActiveInfoIndex", "=", activeInfoIndex)

    // agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    ssUsername := fmt.Sprintf("_%d_%s", config.AgentID, username)
    reqBody := userCredentional {
        Username: ssUsername,
        UPSK: RandomBase64(16),
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return res, err
    }

    config.Log.Info("json.Marshal(reqBody)", "jsonData", string(jsonData))

////////////////////////////////////////////////////////////////////////////////
// prepare http client to POST
////////////////////////////////////////////////////////////////////////////////
    // client := &http.Client{
    //     Timeout: (time.Second * config.ClientTimeout),
    // }

    ob := config.NewOutputBuffer()
    var ssmApiAddr = ""
    for _, server := range yaml.Pools.Servers {
        ssmApiAddr = fmt.Sprintf("%s/%s", server.Addr("users"), ssUsername)

        config.Log.Debug("ssmApiAddr", "=", ssmApiAddr)
        ob.Printf("%-30s", "user.lcok." + server.Location)

        resp, err := httpx.Put(ssmApiAddr, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            ob.Println(err)
            ob.Errorln(err)
            err = nil
            continue
        }
        defer func(){
            errClose := resp.Body.Close()
            if errClose != nil {
                err = errClose
            }
        }()

        if resp.StatusCode == 204 {
            ob.Println("locked")
        }

        if resp.StatusCode != 204 {
            ob.Println("not found")
            body, _ := io.ReadAll(resp.Body)
            err := fmt.Errorf("bad status: %d %s Response: %s", resp.StatusCode, resp.Status, string(body))
            config.Log.Warn("lock / client.Put failure", "username", username, "error", err)
        }

        config.Log.Debug("http response", "status code", resp.StatusCode)
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "user lock failure!")
        ob.Stderr.Reset()
        ob.Fprintln("")
        ob.Flush()
        return res, err
    }

    res = make([]map[string]any, 1, 1)
    res[0] = map[string]any{
        "user": newUser.Username,
        "name": newUser.Realname,
        "error": "",
        "status": "ok",
    }

    ob.Flush()
    return res, nil
}
