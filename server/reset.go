package server

import (
    "encoding/json"
	"net/http"
    "bytes"
	"time"
	"io"
	"fmt"
    "flag"
    "strings"

    "database/sql"

	"github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/httpx"
)

func reset(args []string, pools *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)

    var __ds string
    var __list bool

    fsr := flag.NewFlagSet("server_reset", flag.ExitOnError)

    fsr.StringVar(&__ds, "dst", "", "destination server <NAME|all>")
    fsr.BoolVar(&__list, "list", false, "list of servers to choose")
    fsr.Parse(args)

    serverCount := len(pools.Servers) + 1
    dstList := make([]string, serverCount, serverCount)
    var counter int
    for i, server := range pools.Servers {
        dstList[i] = server.Location
        counter = i
    }
    counter += 1
    dstList[counter] = "all"

    if __list {
        for i := 0; i < len(dstList); i++ {
            fmt.Printf("dst: %s\n", dstList[i])
        }

        return nil
    }


    if len(args) == 0 || __ds == "" {
        fsr.PrintDefaults()
        return nil
    }
    config.Log.Debug("__ds", "=", __ds)

    if __ds == "all" {
        for _, server := range pools.Servers {
            err = resetAll(pools, server.Location)
            if err != nil {
                return fmt.Errorf("reset / resetAll() %w", err)
            }
        }
    }


    if __ds != "all" {
        err = resetAll(pools, __ds)
        if err != nil {
            return fmt.Errorf("reset / resetAll() %w", err)
        }
    }

    return err
}

type userCredentional struct {
    Username string `json:"username"`
    UPSK string `json:"uPSK"`
}

////////////////////////////////////////////////////////////////////////////////
// from a remote server to all remote servers
////////////////////////////////////////////////////////////////////////////////
func resetAll(pools *config.Pools, dst string) (err error) {
    var dstAddr string
    for _, server := range pools.Servers {
        if dst == server.Location {
            dstAddr = server.Addr("users")
        }
    }

    if dstAddr == "" {
        return fmt.Errorf("destination server [%s] not found!", dst)
    }

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

    resp, err := client.Get(dstAddr)
    if err != nil {
        return fmt.Errorf("cannot fetch users from %s server!", dst)
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
    config.Log.Info("agentPrefix", "=", agentPrefix)
    var ssmApiAddr string
    var userCount int
    var excluedCount int
    var endpoint string
    body.Reset()
    for i, server := range pools.Servers {
        if dst != server.Location {
            continue
        }
        ssmApiAddr = server.Addr("users")

        config.Log.Info("SSM API address", "ssmApiAddr", ssmApiAddr)

        fmt.Printf("Reset %s servers\n", dst)
        for _, user := range result.Users {
            if ! strings.HasPrefix(user.Username, agentPrefix) {
                excluedCount += 1
                continue
            }

            endpoint = fmt.Sprintf("%s/%s", ssmApiAddr, user.Username)
            username := strings.TrimPrefix(user.Username, agentPrefix)
            config.Log.Debug("ssmApiAddr", "=", endpoint)
            fmt.Printf("%-30s", "user.delete." + username)

            resp, err := httpx.Delete(endpoint)
            if err != nil {
                return err
            }
            defer func(){
                errClose := resp.Body.Close()
                if errClose != nil {
                    err = errClose
                }
            }()

            io.Copy(&body, resp.Body)
            switch resp.StatusCode {
                case 204:
                fmt.Printf("%s\n", "deleted")
                userCount += 1
                case 404:
                fmt.Printf("%s\n", "unavailable")
                default:
                    err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, body)
                    fmt.Println(err)
            }

            config.Log.Debug("http response", "status code", resp.StatusCode)
            config.Log.Debug("http response", "body", body.String())
            body.Reset()
            endpoint = ""
        }
        fmt.Printf("Users of %s server deleted %d/%d\n", server.Location, userCount, (len(result.Users) - excluedCount))
        if ((len(pools.Servers) - 1) > i) {
            fmt.Printf("\n")
        }
        userCount = 0
        excluedCount = 0
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
