package server

import (
    "fmt"
    "net/http"
    "encoding/json"
    "io"
    "bytes"
    "strings"
    "time"

    "github.com/shakibamoshiri/proxgo/config"
)

//var agentMatch = regexp.MustCompile(`^_\d+_`)
//func trimAgentPrefix(s string) string {
//    return agentMatch.ReplaceAllString(s, "")
//}

func stats(args []string, pc *config.Pools) (err error) {
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

    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    ob := config.NewOutputBuffer()

    var ssmApiAddr string
    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    for i, server := range pc.Servers {
        ssmApiAddr = server.Addr("stats")
        config.Log.Debug("ssmApiAddr", "=", ssmApiAddr)

        resp, err := client.Get(ssmApiAddr)
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

        if resp.StatusCode != 200 {
            var body bytes.Buffer
            io.Copy(&body, resp.Body)
            err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, body.String())
            return err
        }

        body, _ := io.ReadAll(resp.Body)
        config.Log.Debug("body", "string(body)", string(body))

        err = json.Unmarshal(body, &result);
        if err != nil {
            return err
        }

        ob.Printf("%-10s %-10s %-10s %s\n", "Username", "Traffic", "Session", "Location")
        for _, user := range result.Users {
            if ! strings.HasPrefix(user.Username, agentPrefix) {
                continue
            }
            //user.Username = strings.TrimPrefix(user.Username, agentPrefix)

            ob.Printf("%-10s %-10d %-10d %-10s\n",
                user.Username, 
                (user.DownlinkBytes + user.UplinkBytes), 
                (user.TCPSessions + user.UDPSessions), 
                server.Location,
            )
        }
        config.Log.Debug("response status code", "=", resp.StatusCode)
        ob.Printf( "%s server has %d users\n", server.Location, len(result.Users))

        if ((len(pc.Servers) - 1) > i) {
            ob.Println()
        }
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("%s", "server fetch failed, no data written to disk!")
        ob.Stderr.Reset()
        ob.Printf("\n")
        ob.Flush()
        return err
    }

    ob.Flush()
    return nil
}
