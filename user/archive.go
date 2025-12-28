package user

import (
    "os"
    "flag"
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)

func archive(args[] string) (err error) {
    config.Log.Debug("args []string", "=", args)

    flags := flag.NewFlagSet("archive", flag.ExitOnError)

    // Define flags
    username := flags.String("user", "", "username to be archived")

    // Important: Parse the flags!
    flags.Parse(args)


    if *username == "" {
        println("user delete args:")
        flags.PrintDefaults()
        os.Exit(0)
    }
    config.Log.Info("-user", "=", *username)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    config.Log.Debug("agentPrefix", "=", agentPrefix)

    // delete the user from fetched table
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("archive() >> %w", err)
    }

    result, err := db.Exec(`
        BEGIN;
        INSERT INTO archive SELECT NULL, * FROM bytes WHERE username = ?;
        DELETE FROM bytes WHERE username = ?;
        COMMIT;
    `, *username, *username)
    if err != nil {
        config.Log.Error("archive()", "db.Exec", err)
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        fmt.Printf("username %s not found or already archived\n", *username)
        config.Log.Debug("no record found with username", "=", *username)
    } else {
        config.Log.Debug("successfully deleted record with username", "=", *username)
    }

    return nil
}
