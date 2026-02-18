package user

import (
    "os"
    "flag"
    "fmt"
    "io"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/httpx"
)

func _delete(args []string) (res []map[string]any, err error) {
    config.Log.Debug("args", "=", args)

    df := flag.NewFlagSet("delete", flag.ExitOnError)
    var username string
    df.StringVar(&username, "user", "", "username to be deleted")
    df.Parse(args)

    if username == "" {
        println("user delete args:")
        df.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("-user", "=", username)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    config.Log.Info("agentPrefix", "=", agentPrefix)
    ssUsername := agentPrefix + username

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        err = fmt.Errorf("delete >> %w", err)
        return res, err
    }

    var rowCount int
    err = db.QueryRow("SELECT COUNT(*) from users where username = ?", username).Scan(&rowCount)
    if err != nil {
        return res, err
    }
    config.Log.Debug("rowCount", "=", rowCount)
    if rowCount == 0 {
        err = fmt.Errorf("delete / username: %s not found", username)
        return res, err
    }

    var ssmApiAddr = ""
    for _, server := range yaml.Pools.Servers {
        ssmApiAddr = fmt.Sprintf("http://%s:%d/%s/%s", server.APIAddr, server.APIPort, config.SsmApiPathUsers,ssUsername)
        config.Log.Info("ssmApiAddr", "=", ssmApiAddr)

        fmt.Printf("%-30s", "user.delete." + server.Location)
        resp, err := httpx.Delete(ssmApiAddr)
        if err != nil {
            return res, err
        }
        defer func(){
            errClose := resp.Body.Close()
            if errClose != nil {
                err = errClose
            }
        }()

        if resp.StatusCode == 404 {
            println("unavailable")
            continue
        }

        if resp.StatusCode == 204 {
            println("deleted")
            config.Log.Info("response status code", "=", resp.StatusCode)
        } else {
            body, _ := io.ReadAll(resp.Body)
            err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
            return res, err
        }
    }

    // delete the user from fetched table
    result, err := db.Exec(`DELETE FROM fetched WHERE username = ?`, username)
    if err != nil {
        config.Log.Error("delete", "db.Exec", err)
        return res, err
    }

    rowsAffected, _ := result.RowsAffected()
    var status string
    if rowsAffected == 0 {
        config.Log.Info("no record found for username", "=", username)
        status = "unavailable"
    } else {
        config.Log.Info("successfully deleted username", "=", username)
        status = "deleted"
    }

    res = make([]map[string]any, 1, 1)
    res[0] = map[string]any{
        "user": username,
        "error": "",
        "status": status,
    }

    return res, nil
}
