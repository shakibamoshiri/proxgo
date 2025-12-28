package data

import (
    "encoding/json"
	"net/http"
    "bytes"
	"time"
	"io"
	"fmt"
    "flag"

    "database/sql"

	"github.com/shakibamoshiri/proxgo/config"
)

func sync(args []string, pools *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)

    var __ss string
    var __ds string
    var __list bool

    flagSync := flag.NewFlagSet("data_sync", flag.ExitOnError)

    flagSync.StringVar(&__ss, "src", "", "source server <NAME|db>")
    flagSync.StringVar(&__ds, "dst", "", "destination server <NAME|all>")
    flagSync.BoolVar(&__list, "list", false, "list of servers to choose")
    flagSync.Parse(args)

    serverCount := len(pools.Servers) + 1
    srcList := make([]string, serverCount, serverCount)
    dstList := make([]string, serverCount, serverCount)
    var counter int
    for i, server := range pools.Servers {
        srcList[i] = server.Location
        dstList[i] = server.Location
        counter = i
    }
    counter += 1
    srcList[counter] = "db"
    dstList[counter] = "all"

    if __list {
        for i := 0; i < len(srcList); i++ {
            fmt.Printf("src: %s\n", srcList[i])
        }

        fmt.Println()
        for i := 0; i < len(dstList); i++ {
            fmt.Printf("dst: %s\n", dstList[i])
        }

        return nil
    }


    if len(args) == 0 || __ss == "" || __ds == "" {
        flagSync.PrintDefaults()
        return nil
    }
    config.Log.Debug("__ss", "=", __ss)
    config.Log.Debug("__ds", "=", __ds)

    if __ss == "db" && __ds == "all" {
        err = dbToAll(pools, "all")
        if err != nil {
            return fmt.Errorf("sync / serverToServer() %w", err)
        }
    }

    if __ss != "db" && __ds == "all" {
        err = serverToAll(pools, __ss, "all")
        if err != nil {
            return fmt.Errorf("sync / serverToServer() %w", err)
        }
    }

    if __ss == "db" && __ds != "all" {
        err = dbToAll(pools, __ds)
        if err != nil {
            return fmt.Errorf("sync / serverToServer() %w", err)
        }
    }

    if __ss != "db" && __ds != "all" {
        err = serverToAll(pools, __ss, __ds)
        if err != nil {
            return fmt.Errorf("sync / serverToServer() %w", err)
        }

    }

    return err
}

type userCredentional struct {
    Username string `json:"username"`
    UPSK string `json:"uPSK"`
}

////////////////////////////////////////////////////////////////////////////////
// from local database to all remote servers
////////////////////////////////////////////////////////////////////////////////
func dbToAll(pools *config.Pools, dst string) (err error) {
	dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
	db, err := config.OpenDB(dbFile)
	if err != nil {
		return fmt.Errorf("sync / dbToAll() >> %w", err)
	}

    const query string = "select count(*) from users;";
    rowCount, err := dbRowCount(db, query)
    if err != nil {
		return fmt.Errorf("sync / dbRowCount() >> %w", err)
    }
    config.Log.Info("rowCount", "=", rowCount)
    
    users := make([]userCredentional, rowCount, rowCount)

    rows, err := db.Query("SELECT username,password FROM users;")
    if err != nil {
        config.Log.Error("sync", "db.Query()", err)
        return fmt.Errorf("sync / db.Query() %w", err)
    }
    defer func(){
        errClose := rows.Close()
        if errClose != nil {
            err = errClose
        }
    }()

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    config.Log.Info("agent-prefix", "=", agentPrefix)

    for i := 0; rows.Next(); i++ {
        err = rows.Scan(
            &users[i].Username,
            &users[i].UPSK,
        )
        if err != nil {
            return err
        }
    }
    config.Log.Debug("users", "=", users)

    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    // iterate over all servers
    var ssmApiAddr string
    var syncedCount int
    for i, server := range pools.Servers {
        if dst == "all" {
            ssmApiAddr = server.Addr("users")
        } else {
            if dst != server.Location {
                continue
            }
            ssmApiAddr = server.Addr("users")
        }

        config.Log.Info("SSM API address", "ssmApiAddr", ssmApiAddr)

        fmt.Printf("From DB to %s\n", server.Location)
        for _, user := range users {
            username := user.Username
            user.Username = fmt.Sprintf("_%d_%s", config.AgentID, user.Username)
            jsonData, err := json.Marshal(user)
            if err != nil {
                return fmt.Errorf("sync / json.Marshal() %w", err)
            }

            config.Log.Debug("json.Marshal(reqBody)", "jsonData", string(jsonData))

            fmt.Printf("%-30s", "user.sync." + username)
            resp, err := client.Post(ssmApiAddr, "application/json", bytes.NewBuffer(jsonData))
            if err != nil {
                return err
            }
            defer func(){
                errClose := resp.Body.Close()
                if errClose != nil {
                    err = errClose
                }
            }()

            body, _ := io.ReadAll(resp.Body)
            switch resp.StatusCode {
                case 201:
                fmt.Printf("%s\n", "synced")
                syncedCount += 1
                case 400:
                fmt.Printf("%s\n", "existed")
                default:
                    err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
                    fmt.Println(err)
            }

            config.Log.Debug("http response", "status code", resp.StatusCode)
        }
        fmt.Printf("%s server synced %d/%d\n", server.Location, syncedCount, len(users))
        if ((len(pools.Servers) - 1) > i) {
            fmt.Printf("\n")
        }
        syncedCount = 0
    }

    return nil
}

////////////////////////////////////////////////////////////////////////////////
// from a remote server to all remote servers
////////////////////////////////////////////////////////////////////////////////
func serverToAll(pools *config.Pools, src string, dst string) (err error) {
    fmt.Println("server to server")

    var srcAddr string
    var dstAddr string
    for _, server := range pools.Servers {
        if src == server.Location {
            srcAddr = server.Addr("users")
        }
        if dst == server.Location {
            dstAddr = server.Addr("users")
        }
    }

    if srcAddr == dstAddr {
        return fmt.Errorf("source server [%s] and destination server [%s] are the same", src, dst)
    }

    if srcAddr == "" {
        return fmt.Errorf("source server [%s] not found!", src)
    }

    if dstAddr == "" {
        if dst != "all" {
            return fmt.Errorf("destination server [%s] not found!", dst)
        }
    }

    config.Log.Info("src address", "=", srcAddr)
    config.Log.Info("dst address", "=", dstAddr)

    // json data holder
	type ServerUser struct {
		Username    string   `json:"username"`
        UPSK        string `json:"uPSK"`
		Extra       struct{} `json:"-"`
	}

	type ServerUsersResponse struct {
		Users []ServerUser `json:"users"`
	}

	var result ServerUsersResponse

    // get users from source server
    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    resp, err := client.Get(srcAddr)
    if err != nil {
        return fmt.Errorf("cannot fetch users from %s server!", src)
    }
    defer func(){
        errClose := resp.Body.Close()
        if errClose != nil {
            err = errClose
        }
    }()

    var body bytes.Buffer
    io.Copy(&body, resp.Body)
    config.Log.Debug("body", "encoded", body.String())
    if resp.StatusCode != 200 {
        err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, body.String())
        return err
    }

    json.NewDecoder(&body).Decode(&result)
    config.Log.Debug("body", "decoded", len(result.Users))

    // iterate over all servers
    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    config.Log.Info("agent-prefix", "=", agentPrefix)
    var ssmApiAddr string
    var syncedCount int
    for i, server := range pools.Servers {
        if dst == "all" {
            ssmApiAddr = server.Addr("users")
        } else {
            if dst != server.Location {
                continue
            }
            ssmApiAddr = server.Addr("users")
        }

        config.Log.Info("SSM API address", "ssmApiAddr", ssmApiAddr)

        fmt.Printf("From %s to %s\n", src, server.Location)
        for _, user := range result.Users {
            jsonData, err := json.Marshal(user)
            if err != nil {
                return fmt.Errorf("sync / json.Marshal() %w", err)
            }

            config.Log.Debug("json.Marshal(reqBody)", "jsonData", string(jsonData))

            fmt.Printf("%-30s", "user.sync." + user.Username)
            resp, err := client.Post(ssmApiAddr, "application/json", bytes.NewBuffer(jsonData))
            if err != nil {
                return err
            }
            defer func(){
                errClose := resp.Body.Close()
                if errClose != nil {
                    err = errClose
                }
            }()

            body, _ := io.ReadAll(resp.Body)
            switch resp.StatusCode {
                case 201:
                fmt.Printf("%s\n", "synced")
                syncedCount += 1
                case 400:
                fmt.Printf("%s\n", "existed")
                default:
                    err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
                    fmt.Println(err)
            }

            config.Log.Debug("http response", "status code", resp.StatusCode)
        }
        fmt.Printf("%s server synced %d/%d\n", server.Location, syncedCount, len(result.Users))
        if ((len(pools.Servers) - 1) > i) {
            fmt.Printf("\n")
        }
        syncedCount = 0
    }

    return err
}

////////////////////////////////////////////////////////////////////////////////
// from a single remote server to a single remote server
////////////////////////////////////////////////////////////////////////////////
func serverToServer(){
    fmt.Println("from a remote server to all")
}

////////////////////////////////////////////////////////////////////////////////
// from local database to a single remote server
////////////////////////////////////////////////////////////////////////////////
func dbToServer(){
    fmt.Println("db to a remote server")
}

func dbRowCount(db *sql.DB, query string) (int64, error) {
    var rowCount int64 = 0
    err := db.QueryRow(query).Scan(&rowCount)
    if err != nil {
        config.Log.Error("dbRowCount()", "db.QueryRow", err)
        return 0, fmt.Errorf("dbRowCount() / db.QueryRow %w", err)
    }

    return rowCount, nil
}
