package user

import (
    "os"
    "flag"
    "fmt"
    "io"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/httpx"
)

func _delete(args []string) (err error) {
    config.Log.Debug("args", "=", args)

    flags := flag.NewFlagSet("delete", flag.ExitOnError)

    // Define flags
    username := flags.String("user", "", "username to be deleted")

    // Important: Parse the flags!
    flags.Parse(args)


    if *username == "" {
        println("user delete args:")
        flags.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("-user", "=", *username)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    config.Log.Info("agentPrefix", "=", agentPrefix)
    ssUsername := agentPrefix + *username

    var ssmApiAddr = ""
    for _, server := range yaml.Pools.Servers {
        ssmApiAddr = fmt.Sprintf("http://%s:%d/%s/%s", server.APIAddr, server.APIPort, config.SsmApiPathUsers,ssUsername)
        config.Log.Info("ssmApiAddr", "=", ssmApiAddr)

        fmt.Printf("%-30s", "user.delete." + server.Location)
        resp, err := httpx.Delete(ssmApiAddr)
        if err != nil {
            return err
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
            return err
        }
    }

    // delete the user from fetched table
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)

    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("delete() >> %w", err)
    }
    // db, openError := sql.Open("sqlite", dbFile)
    // if openError != nil {
    //     return openError
    // }
    // defer func(){
    //     errClose := db.Close()
    //     if errClose != nil {
    //         err = errClose
    //     }
    // }()

    result, err := db.Exec(`DELETE FROM fetched WHERE username = ?`, *username)
    if err != nil {
        config.Log.Error("delete", "db.Exec", err)
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        config.Log.Info("no record found for username", "=", *username)
    } else {
        config.Log.Info("successfully deleted username", "=", *username)
    }

    return nil
}
