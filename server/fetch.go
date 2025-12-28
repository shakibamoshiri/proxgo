package server

import (
    "fmt"
    "net/http"
    "encoding/json"
    "io"
    "regexp"
    "strings"
    "time"
    "os"

    "github.com/shakibamoshiri/proxgo/config"
)

var agentMatch = regexp.MustCompile(`^_\d+_`)
func trimAgentPrefix(s string) string {
    return agentMatch.ReplaceAllString(s, "")
}

func fetch(args []string, pc *config.Pools, dev io.Writer) (err error) {
    type ServerUser struct {
        Username         string `json:"username"`
        UPSK             string `json:"uPSK"`
        DownlinkBytes    int64  `json:"downlinkBytes"`
        UplinkBytes      int64  `json:"uplinkBytes"`
        DownlinkPackets  int64  `json:"downlinkPackets"`
        UplinkPackets    int64  `json:"uplinkPackets"`
        TCPSessions      int64  `json:"tcpSessions"`
        UDPSessions      int64  `json:"udpSessions"`
    }

    type ServerUsersResponse struct {
        Users []ServerUser `json:"users"`
    }

    var result ServerUsersResponse

    type Agg struct {
        Traffic int64
        Session int64
    }

    agg := make(map[string]*Agg)

    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    ob := config.NewOutputBuffer()

    var ssmApiAddr = ""
    for _, server := range pc.Servers {
        ssmApiAddr = fmt.Sprintf("http://%s:%d/%s", server.APIAddr, server.APIPort, config.SsmApiPathStats)
        config.Log.Info("ssmApiAddr", "=", ssmApiAddr)

        // fmt.Printf("%-30s", "server.fetch." + server.Location)
        ob.Fprintf(dev, "%-30s", "server.fetch." + server.Location)
        resp, err := client.Get(ssmApiAddr)
        if err != nil {
            ob.Fprintln(dev, err)
            ob.Fprintln(os.Stderr, err)
            err = nil
            continue
        }
        defer func(){
            errClose := resp.Body.Close()
            if errClose != nil {
                err = errClose
            }
        }()

        body, _ := io.ReadAll(resp.Body)
        config.Log.Debug("body", "string(body)", string(body))
        if resp.StatusCode != 200 {
            err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
            return err
        }


        err = json.Unmarshal(body, &result);
        if err != nil {
            // log.Fatal(err)
            return err
        }

        for i, user := range result.Users {
            config.Log.Debug("result.Users", "index", i, "user", user)

            a, ok := agg[user.Username]
            if !ok {
                a = &Agg{}
                agg[user.Username] = a
            }

            a.Traffic += user.DownlinkBytes + user.UplinkBytes
            a.Session += user.TCPSessions + user.UDPSessions
        }
        ob.Fprintln(dev, "fetched")
        config.Log.Info("response status code", "=", resp.StatusCode)
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("%s", "server fetch failed, no data written to disk!")
        ob.Stderr.Reset()
        ob.Fprintln(dev)
        ob.Flush()
        return err
    }

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    config.Log.Debug("agentPrefix", "=", agentPrefix)
    config.Log.Debug("dbFile", "=", dbFile)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        return fmt.Errorf("fetch() >> %w", err)
    }

    var recordCount = 0
    for username, a := range agg {
        if ! strings.HasPrefix(username, agentPrefix) {
            continue
        }
        recordCount++
        username = trimAgentPrefix(username)
        config.Log.Debug(username, "traffic", a.Traffic, "session", a.Session)
        _, errExec := db.Exec(`
            INSERT INTO fetched (username, traffic, session)
            VALUES (?, ?, ?)
            ON CONFLICT(username) DO UPDATE SET
                traffic = excluded.traffic,
                session = excluded.session;
            `, username, a.Traffic, a.Session)

        if errExec != nil {
            return errExec
        }
    }
    config.Log.Debug("row affected", "recordCount", recordCount)

    ob.Flush()
    return nil
}
