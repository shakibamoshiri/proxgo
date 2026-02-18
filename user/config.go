package user

import (
    "os"
    "flag"
    "fmt"
    "encoding/json"
    "path/filepath"
    "time"

    "github.com/shakibamoshiri/proxgo/config"
)

func confiG(args[] string, pools *config.Pools) (err error) {
    config.Log.Debug("args []string", "=", args)

    ucf := flag.NewFlagSet("user_conf", flag.ExitOnError)
    var username string
    var device string
    ucf.StringVar(&username, "user", "", "username")
    ucf.StringVar(&device, "device", "", "[android|iphone|linux|windows]")
    ucf.Parse(args)

    if username == "" {
        ucf.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("username", "user", username)

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        return fmt.Errorf("config / OpenDB(%s) >> %w", dbFile,  err)
    }

    stmt, errPer := db.Prepare(`SELECT * FROM users WHERE username = ?`)
    if errPer != nil {
        config.Log.Error("db.Prepare", "error", errPer)
        return fmt.Errorf("config / db.Prepare() >> %w", errPer)
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = fmt.Errorf("config / stmt.Close() >> %w", errClose)
        }
    }()

    thisUser := User{}

    errExec := stmt.QueryRow(username).Scan(
        &thisUser.Username,
        &thisUser.Realname,
        &thisUser.Ctime,
        &thisUser.Period,
        &thisUser.Traffic,
        &thisUser.Password,
        &thisUser.Page,
        &thisUser.Profile,
        &thisUser.Device,
    )
    if errExec != nil {
        config.Log.Error("stmt.QueryRow failed", "error", errExec)
        return fmt.Errorf("config / stmt.QueryRow() failed %w", errExec)
    }

    config.Log.Debug("thisUser", "=",  thisUser)
    config.Log.Info("thisUser.Username", "=", thisUser.Username)

    activePoolIndex := yaml.ActivePoolIndex()
    activeInfoIndex := yaml.ActiveInfoIndex()
    config.Log.Debug("yaml.ActivePoolIndex", "=", activePoolIndex)
    config.Log.Debug("yaml.ActiveInfoIndex", "=", activeInfoIndex)
    config.Log.Debug("yaml.Pools.DB.Info", "=", yaml.Pools.DB.Info[activeInfoIndex])

    groupName := pools.DB.Info[activeInfoIndex].Name
    dashPath := pools.DB.Info[activeInfoIndex].Profile.WebRoot
    dir := fmt.Sprintf("%s/%s/%s", dashPath, groupName, thisUser.Page)
    configFile := fmt.Sprintf("%s/%s", dir, thisUser.Profile)
    config.Log.Info("config file", "=", configFile)


    // aged24, err := fileStat(configFile)
    // if err != nil {
    //     return fmt.Errorf("config / fileStat() failed %w", err)
    // }

    // if !aged24 {
    //     config.Log.Info("config file has been modified in less than 24h, updating ignored")
    //     return nil
    // }
    // config.Log.Info("config file not found or aged more than 24h, should be created/updated")

    config.Log.Info("config file directory", "=", dir)
    err = os.MkdirAll(dir, 0o755)
    if err != nil {
        return fmt.Errorf("config / os.MkdirAll(%s) failed %w", dir, err)
    }

    ssPass := yaml.Pools.DB.Info[activeInfoIndex].Pass.SS
    hy2Password := fmt.Sprintf("_%d_%s:%s", config.AgentID, username, thisUser.Password)
    ssPassword := fmt.Sprintf("%s:%s", ssPass, thisUser.Password)
    tlsPassword := yaml.Pools.DB.Info[activeInfoIndex].Pass.TLS
    config.Log.Debug("ssPassword", "=", ssPassword)
    config.Log.Debug("tlsPassword", "=", tlsPassword)

    var ccJson map[string]any
    var ccPath string
    if device == "" {
        if thisUser.Device == "undefined" {
            ccPath = fmt.Sprintf("./%s/%d.json", config.ClientPath, config.AgentID)
        } else {
            ccPath = fmt.Sprintf("./%s/%d.%s.json", config.ClientPath, config.AgentID, thisUser.Device)
        }
    } else {
        ccPath = fmt.Sprintf("./%s/%d.%s.json", config.ClientPath, config.AgentID, device)
    }
    config.Log.Info("ccPath", "=", ccPath)
    err = loadClientConfig(ccPath, &ccJson)
    if err != nil {
        if os.IsNotExist(err) {
            config.Log.Error("ccPath file not found, load ignored", "ccPath", ccPath )
            return
        }
        config.Log.Error("config / loadClientConfig() failed", "ccPath:", ccPath, "error", err)
        return fmt.Errorf("config / loadClientConfig(%s) failed %w", ccPath, err)
    }
    outbounds, ok := ccJson["outbounds"].([]any)
    if ok {
        for _, outs := range outbounds {
            outbound, ok := outs.(map[string]any)
            if ok {
                _type, _ := outbound["type"].(string)
                if _type == "shadowsocks" {
                    outbound["password"] = ssPassword
                    // config.Log.Debug("_type", "=", _type)
                }

                if _type == "shadowtls" {
                    outbound["password"] = tlsPassword
                    // config.Log.Debug("_type", "=", _type)
                }

                if _type == "hysteria2" {
                    outbound["password"] = hy2Password
                }
            }
        }
    }
    config.Log.Debug(ccPath, "=", ccJson)

    err = saveClientConfigCompact(configFile, &ccJson)
    if err != nil {
        return fmt.Errorf("config / saveClientConfigCompact(%s) failed %w", configFile, err)
    }

    link := pools.DB.Info[activeInfoIndex].Profile.Link
    fmt.Printf("page: %s/%s/%s/\n", link, groupName, thisUser.Page)
    fmt.Printf("link: %s/%s/%s/%s\n", link, groupName, thisUser.Page, thisUser.Profile)

    return err
}

func loadClientConfig(path string, holder any) (err error) {
    file, err := os.Open(path)
    if err != nil {
        config.Log.Error("config / loadClientConfig() / os.Open", "err", err)
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
    config.Log.Debug("os.CreateTemp", "tmpName", tmpName)

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

    config.Log.Info("os.Rename", "tmpName", tmpName, "path", path)
    err = os.Rename(tmpName, path)
    return
}

func fileStat(path string) (bool, error) {
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return true, nil // file does not exist
        }
        return false, err // other error (permission, etc.)
    }

    // Check if it's a regular file (optional, but recommended)
    if !info.Mode().IsRegular() {
        return false, nil // not a regular file (e.g., directory)
    }

    // Check modification time
    cutoff := time.Now().Add(-24 * time.Hour)
    if info.ModTime().Before(cutoff) {
        return true, nil // modified more than 24h ago
    }

    return false, nil // modified within last 24h
}

