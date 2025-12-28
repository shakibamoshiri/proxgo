package user

import (
    // "os"
    "fmt"
    "log"
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

func create(args []string) (err error) {
    config.Log.Debug("args", "=", args)

    flags := map[string]string{
        "name":       "name (alias) for a user [string]",
        "period" :    "period (duration) <number>[s|m|h]",
        "traffic" :   "traffic (volume) <number>[b|k|m|g]]",
        "password" :  "password (base64)",
        "username" :  "custom username for a user",
        "help" :  "show help",
    }


    fq := &config.FlagQuery{}
    fq.Take(args, flags)

    __help, _ := fq.Find("help").Bool()
    if __help {
        fq.Help(flags).Exit(0)
    }

    __realname, err := fq.Find( "name").Assert().String()
    if err != nil {
        log.Fatal(err)
    }

    __username, err := fq.Find("username").String()
    if err != nil {
        log.Fatal(err)
    }

    __traffic, err := fq.Find("traffic").String()
    if err != nil {
        log.Fatal(err)
    }

    __period, err := fq.Find("period").String()
    if err != nil {
        log.Fatal(err)
    }

    __password, err := fq.Find("password").String()
    if err != nil {
        log.Fatal(err)
    }




    userTraffic := parseTrafficCustom(&__traffic)
    userPeriod := parsePeriodCustom(&__period)

    config.Log.Info("parseTrafficCustom(...)", "trafficCustom", userTraffic)
    config.Log.Info("parsePeriodCustom(...)", "periodCustom", userPeriod)


    if __username == "" {
        __username = RandomHex(3)
    }
    config.Log.Info("flag set check", "--username", __username)

    if __password == "" {
        __password = RandomBase64(16)
    }
    config.Log.Info("flag set check", "--password", __password)

    agentPrefix := fmt.Sprintf("_%d_", config.AgentID)
    ssUsername := fmt.Sprintf("_%d_%s", config.AgentID, __username)
    config.Log.Info("agent-prefix", "=", agentPrefix)
    config.Log.Info("shadowsocks-username", "=", ssUsername)

////////////////////////////////////////////////////////////////////////////////
// safeguard, servers should be up
////////////////////////////////////////////////////////////////////////////////
    serverArgs := make([]string, 0, 0)
    err = server.Run("check", serverArgs, &yaml.Pools, io.Discard)
    if err != nil {
        return fmt.Errorf("server.Run(check) %w", err)
    }

////////////////////////////////////////////////////////////////////////////////
// request body in JSON
////////////////////////////////////////////////////////////////////////////////
    var activePoolIndex = yaml.ActivePoolIndex()
    var activeInfoIndex = yaml.ActiveInfoIndex()
    config.Log.Debug("DB.Info", "=", yaml.Pools.DB.Info[activeInfoIndex])

    reqBody := userCredentional {
        Username: agentPrefix + __username,
        UPSK: __password,
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return err
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
        ssmApiAddr = server.Addr("users")
        config.Log.Info("SSM API address", "ssmApiAddr", ssmApiAddr)

        fmt.Printf("%-30s", "user.create." + __username)
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

        if resp.StatusCode != 201 {
            body, _ := io.ReadAll(resp.Body)
            err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, string(body))
            return err
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
        Username: __username,
        Realname: __realname,
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
        return fmt.Errorf("create() >> %w", err)
    }

    stmt, errPer := db.Prepare(`
        INSERT INTO users
        (username, realname, ctime, period, traffic, password, page, profile)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)

    if errPer != nil {
        config.Log.Error("db.Prepare", "=", errPer)
        return errPer
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
        return errExec
    }

    return nil
}
