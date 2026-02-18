package user

import (
    "fmt"
    "encoding/hex"
    "encoding/json"
    "encoding/base64"
    "math/rand"
    "log/slog"
    "time"
    "net/http"
    "io"
    "bytes"
    "regexp"
    "strings"
    "strconv"
    "flag"
    "os"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/server"
)


type User struct {
    Username string
    Realname string
    Ctime    int64
    Period   int64
    Traffic  int64
    Password string
    Page     string
    Profile  string
    Device   string
}

func RandomHex(n int) string {
    bytes := make([]byte, n)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}

func RandomBase64(n int) string {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        slog.Error("something went wrong", "err", err)
    }
    return base64.StdEncoding.EncodeToString(b)
}

type userCredentional struct {
    Username string `json:"username"`
    UPSK string `json:"uPSK"`
}

var sizeMatch = regexp.MustCompile(`(\d+)([bkmg])`)
func parseTrafficCustom(format *string) int64 {
    *format = strings.ToLower(*format)
    matchList := sizeMatch.FindStringSubmatch(*format)
    config.Log.Info("matchList", "=", matchList)
    if len(matchList) == 3 {
        value, _ := strconv.ParseInt(matchList[1], 10, 64)
        unit := matchList[2]
        switch unit {
            case "b":
                return value
            case "k":
                return (value * 1024)
            case "m":
                return (value * 1024 * 1024)
            case "g":
                return (value * 1024 * 1024 * 1024)
            default:
                return value
        }
    }
    return 0
}

var timeMatch = regexp.MustCompile(`(\d+)([smhd])`)
func parsePeriodCustom(format *string) int64 {
    *format = strings.ToLower(*format)
    matchList := timeMatch.FindStringSubmatch(*format)
    config.Log.Info("matchList", "=", matchList)
    if len(matchList) == 3 {
        value, _ := strconv.ParseInt(matchList[1], 10, 64)
        unit := matchList[2]
        switch unit {
            case "s":
                return value
            case "m":
                return (value * 60)
            case "h":
                return (value * 60 * 60)
            case "d":
                return (value * 60 * 60 * 24)
            default:
                return value
        }
    }

    return 0
}

func create(args []string) (res []map[string]any, err error) {
    config.Log.Debug("args", "=", args)

    ucf := flag.NewFlagSet("userCreate", flag.ExitOnError)
    var __name string
    var __period string
    var __traffic string
    var __password string
    var __noserver bool
    var __user string
    var __help bool
    var __api bool

    ucf.BoolVar(&__api, "api", false, "api call (DO NOT USE THIS DIRECTLY)")
    ucf.BoolVar(&__help, "help", false, "show help")
    ucf.BoolVar(&__noserver, "nosrv", false, "add user but skip servers")
    ucf.StringVar(&__name, "name", "", "real/alias name for a user (REQUIRED)")
    ucf.StringVar(&__period, "period", "", "duration time <NUMBER|s|m|h>")
    ucf.StringVar(&__traffic, "traffic", "", "traffic volume <NUMBERb|k|m|g>")
    ucf.StringVar(&__password, "password", "", "base64 len(16) - openssl rand -base64 16")
    ucf.Parse(args)

    if __help {
        ucf.PrintDefaults()
        os.Exit(0)
    }

    if __name == "" {
        config.Log.Error("create / --name is required")
        err = fmt.Errorf("create / --name is required")
        return res, err
    }

    userTraffic := parseTrafficCustom(&__traffic)
    userPeriod := parsePeriodCustom(&__period)

    config.Log.Info("parseTrafficCustom(...)", "trafficCustom", userTraffic)
    config.Log.Info("parsePeriodCustom(...)", "periodCustom", userPeriod)


    if __user == "" {
        __user = RandomHex(3)
    }
    config.Log.Info("flag set check", "--username", __user)

    if __password == "" {
        __password = RandomBase64(16)
    }
    config.Log.Info("flag set check", "--password", __password)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    ssUsername := fmt.Sprintf("_%d_%s", config.AgentID, __user)
    config.Log.Info("agent-prefix", "=", agentPrefix)
    config.Log.Info("shadowsocks-username", "=", ssUsername)

////////////////////////////////////////////////////////////////////////////////
// safeguard, servers should be up
////////////////////////////////////////////////////////////////////////////////
    serverArgs := make([]string, 0, 0)
    err = server.Run("check", serverArgs, &yaml.Pools, io.Discard)
    if err != nil {
        err = fmt.Errorf("server.Run(check) %w", err)
        return res, err
    }

////////////////////////////////////////////////////////////////////////////////
// request body in JSON
////////////////////////////////////////////////////////////////////////////////
    var activePoolIndex = yaml.ActivePoolIndex()
    var activeInfoIndex = yaml.ActiveInfoIndex()
    config.Log.Debug("DB.Info", "=", yaml.Pools.DB.Info[activeInfoIndex])

    reqBody := userCredentional {
        Username: agentPrefix + __user,
        UPSK: __password,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return res, err
    }

    config.Log.Info("json.Marshal(reqBody)", "jsonData", string(jsonData))

////////////////////////////////////////////////////////////////////////////////
// make http POST request
////////////////////////////////////////////////////////////////////////////////
    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    var ssmApiAddr = ""
    for _, server := range yaml.Pools.Servers {
        if __noserver {
            continue
        }
        ssmApiAddr = server.Addr("users")
        config.Log.Info("SSM API address", "ssmApiAddr", ssmApiAddr)

        fmt.Printf("%-30s", "user.create." + __user)
        resp, err := client.Post(ssmApiAddr, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            return res, err
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
            return res, err
        }

        fmt.Printf("%s\n", "created")
        config.Log.Info("http response", "status code", resp.StatusCode)
    }

////////////////////////////////////////////////////////////////////////////////
// add new use to database
////////////////////////////////////////////////////////////////////////////////
    var traffic int64
    var period int64
    if userTraffic == 0 {
        traffic = yaml.Agents.Pools[activePoolIndex].Traffic
    } else {
        traffic = userTraffic
    }
    if userPeriod == 0 {
        period = yaml.Agents.Pools[activePoolIndex].Period
    } else {
        period = userPeriod
    }
    // page := yaml.Pools.DB.Info[activeInfoIndex].Profile.Link + "/" + yaml.Agents.Agent.GroupName
    page := RandomHex(4)
    now := time.Now().Unix()
    profile := RandomHex(4)
    newUser := User{
        Username: __user,
        Realname: __name,
        Password: __password,
        Ctime: now,
        Profile: profile,
        Page: page,
        Traffic: traffic,
        Period: period,
    }

    config.Log.Info("new user data", "newUser", newUser)
    config.Log.Info("prefix of the agent", "agentPrefix", agentPrefix)
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    config.Log.Info("database file for user create", "dbFile", dbFile)

    db, err := config.OpenDB(dbFile)
    if err != nil {
        err = fmt.Errorf("create() >> %w", err)
        return res, err
    }

    stmt, errPer := db.Prepare(`
        INSERT INTO users
        (username, realname, ctime, period, traffic, password, page, profile)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)

    if errPer != nil {
        config.Log.Error("db.Prepare", "=", errPer)
        return res, errPer
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = errClose
        }
    }()

    _, errExec := stmt.Exec(
        newUser.Username,
        newUser.Realname,
        newUser.Ctime,
        newUser.Period,
        newUser.Traffic,
        newUser.Password,
        newUser.Page,
        newUser.Profile,
    )
    if errExec != nil {
        config.Log.Error("db.Prepare", "=", errExec)
        return res, errExec
    }

    res = make([]map[string]any, 1, 1)
    res[0] = map[string]any{
        "user": __user,
        "name": __name,
        "error": "",
        "status": "ok",
    }


    return res, nil
}
