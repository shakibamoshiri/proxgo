package user

import (
    "os"
    "flag"
    "fmt"
    "encoding/json"
    "path/filepath"

    "github.com/shakibamoshiri/proxgo/config"
)

func confiG(args[] string) (err error) {
    config.Log.Debug("args []string", "=", args)

    flags := flag.NewFlagSet("user_conf", flag.ExitOnError)
    username := flags.String("user", "", "username to be deleted")
    flags.Parse(args)


    if *username == "" {
        flags.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("username", "-user", *username)

    // delete the user from fetched table
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)

    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("list() >> %w", err)
    }

    stmt, errPer := db.Prepare(` SELECT * FROM users WHERE username = ?`)
    if errPer != nil {
        config.Log.Debug("db.Prepare", "error", errPer)
        return errPer
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = errClose
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
        config.Log.Debug("db.Prepare", "error", errExec)
        return errExec
    }

    config.Log.Debug("newUser", "=",  newUser)
    config.Log.Info("newUser.Username", "=", newUser.Username)

    activePoolIndex := yaml.ActivePoolIndex()
    activeInfoIndex := yaml.ActiveInfoIndex()
    config.Log.Info("yaml.ActivePoolIndex", "=", activePoolIndex)
    config.Log.Info("yaml.ActiveInfoIndex", "=", activeInfoIndex)
    config.Log.Debug("yaml.Pools.DB.Info", "=", yaml.Pools.DB.Info[activeInfoIndex])

    ssPassword := yaml.Pools.DB.Info[activeInfoIndex].Pass.SS + ":" + newUser.Password
    tlsPassword := yaml.Pools.DB.Info[activeInfoIndex].Pass.TLS
    config.Log.Info("ssPassword", "=", ssPassword)
    config.Log.Info("tlsPassword", "=", tlsPassword)

    var ccJson map[string]any
    ccPath := fmt.Sprintf("./%s/%d.json", config.ClientPath, config.AgentID)
    err = loadClientConfig(ccPath, &ccJson)
    if err != nil {
        return err
    }
    //config.Log.Debug(ccPath, "=", ccJson)
    outbounds, ok := ccJson["outbounds"].([]any)
    if ok {
        for _, outs := range outbounds {
            outbound, ok := outs.(map[string]any)
            if ok {
                _type, _ := outbound["type"].(string)
                if _type == "shadowsocks" {
                    outbound["password"] = ssPassword
                    config.Log.Info("_type", "=", _type)
                }
                if _type == "shadowtls" {
                    outbound["password"] = tlsPassword
                    config.Log.Info("_type", "=", _type)
                }
            }
        }
    }
    config.Log.Debug(ccPath, "=", ccJson)

    ncPath := fmt.Sprintf("./%s/%s.json", "dash/web", newUser.Username)
    err = saveClientConfigCompact(ncPath, &ccJson)
    if err != nil {
        return err
    }


    return nil
}

func loadClientConfig(path string, holder any) (err error) {
    file, err := os.Open(path)
    if err != nil {
        fmt.Println("os.Open", "err", err)
        return err
    }
    defer func(){
        errClose := file.Close()
        if errClose != nil {
            err = errClose
            return
        }
    }()

    err = json.NewDecoder(file).Decode(holder)
    return err
}


func saveClientConfigCompact(path string, content any) (err error) {
    dir := filepath.Dir(path)

    tmp, err := os.CreateTemp(dir, ".tmp-*.json")
    if err != nil {
            return err
    }
    tmpName := tmp.Name()
    config.Log.Info("os.CreateTemp", "tmpName", tmpName)

    defer func() {
        errClose := tmp.Close()
        if err != nil {
            err = errClose
            return
        }
    }()

    // do we need this ?
    // tmp.Sync()

    compacted, err := json.Marshal(content)
    if err != nil {
            return err
    }

    err = os.WriteFile(tmpName, compacted, 0o644)
    if err != nil {
            return err
    }

    err = os.Rename(tmpName, path)
    return
}

