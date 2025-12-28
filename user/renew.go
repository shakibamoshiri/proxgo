package user

import (
    "os"
    "flag"
    "fmt"
    "io"
    "encoding/json"
    "net/http"
    "bytes"
    "time"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/server"
)

func renew(args[] string) (err error) {
    config.Log.Debug("args []string", "=", args)

    flags := flag.NewFlagSet("renew", flag.ExitOnError)
    username := flags.String("user", "", "username to renew")
    flags.Parse(args)


    if *username == "" {
        println("user renew args:")
        flags.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("-user", "=", *username)

////////////////////////////////////////////////////////////////////////////////
// get the user from database
////////////////////////////////////////////////////////////////////////////////
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    config.Log.Info("dbFile", "=", dbFile)

    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("renew() >> %w", err)
    }

    stmt, errPer := db.Prepare(` SELECT * FROM users WHERE username = ?`)
    if errPer != nil {
        config.Log.Error("renew", "db.Prepare", errPer)
        return fmt.Errorf("renew() >> %w", errPer)
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = fmt.Errorf("renew() >> %w", errClose)
        }
    }()

    newUser := User{}

    errExec := stmt.QueryRow(*username).Scan(
        &newUser.Username,
        &newUser.Realname,
        &newUser.Ctime,
        &newUser.Period,
        &newUser.Traffic,
        &newUser.Password,
        &newUser.Page,
        &newUser.Profile,
    )
    if errExec != nil {
        config.Log.Error("db.Prepare", "error", errExec)
        return fmt.Errorf("renew() / stmt.QueryRow(%v) >> %w", *username, errExec)
    }

    config.Log.Info("newUser", "=",  newUser)
    config.Log.Info("newUser.Username", "=", newUser.Username)

////////////////////////////////////////////////////////////////////////////////
// check if servers are up?
////////////////////////////////////////////////////////////////////////////////
    serverArgs := make([]string, 0, 0)
    err = server.Run("check", serverArgs, &yaml.Pools, io.Discard)
    if err != nil {
        return err
    }

////////////////////////////////////////////////////////////////////////////////
// prepare json data
////////////////////////////////////////////////////////////////////////////////
    activePoolIndex := yaml.ActivePoolIndex()
    activeInfoIndex := yaml.ActiveInfoIndex()
    config.Log.Info("yaml.ActivePoolIndex", "=", activePoolIndex)
    config.Log.Info("yaml.ActiveInfoIndex", "=", activeInfoIndex)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    reqBody := userCredentional {
        Username: agentPrefix + *username,
        UPSK: newUser.Password,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return err
    }

    config.Log.Info("json.Marshal(reqBody)", "jsonData", string(jsonData))

////////////////////////////////////////////////////////////////////////////////
// prepare http client to POST
////////////////////////////////////////////////////////////////////////////////
    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    ob := config.NewOutputBuffer()
    var ssmApiAddr = ""
    for _, server := range yaml.Pools.Servers {
        ssmApiAddr = server.Addr("users")

        config.Log.Info("ssmApiAddr", "=", ssmApiAddr)
        ob.Printf("%-30s", "user.renew." + server.Location)

        resp, err := client.Post(ssmApiAddr, "application/json", bytes.NewBuffer(jsonData))
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

        if resp.StatusCode != 201 {
            body, _ := io.ReadAll(resp.Body)
            err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
            return fmt.Errorf("renew() / client.Post(%s) %w", *username, err)
        }

        ob.Println("renewed")
        config.Log.Info("http response", "status code", resp.StatusCode)
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "user renew failure!")
        ob.Stderr.Reset()
        ob.Fprintln("")
        ob.Flush()
        return err
    }

    ob.Flush()

////////////////////////////////////////////////////////////////////////////////
// sync data to fetch table
////////////////////////////////////////////////////////////////////////////////
    var nextArgs = make([]string, 0, 0)
    err = server.Run("fetch", nextArgs, &yaml.Pools, io.Discard)
    if err != nil {
        return fmt.Errorf("renew() / server.Run(fetch) %w", err)
    }

    return nil
}
