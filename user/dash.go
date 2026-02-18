package user

import (
    "html/template"
    "log"
    "net/http"
    "net/url"
    "os"
    "path/filepath"
    "strings"
    "fmt"
    "time"
    "context"
    "syscall"
    "os/signal"
    "database/sql"
    "encoding/json"
    "flag"
    "strconv"

    "github.com/shakibamoshiri/proxgo/config"

    qrcode "github.com/skip2/go-qrcode"
)

const userProfileTemplate = "dash/user-profile.template"
var tmpl *template.Template
type PageData struct {
    InputLink string
    ImgSrc string
}

var db *sql.DB
var allUsers []User
var mapPages map[string]int

func dash(args []string) (err error) {

    agents, err := yaml.Agents.Load()
    activePoolId := agents.Agent.PoolID
    pools, _ := yaml.Pools.Load(activePoolId)

    activeInfoIndex := yaml.ActiveInfoIndex()
    groupName := pools.DB.Info[activeInfoIndex].Name
    profileLink := pools.DB.Info[activeInfoIndex].Profile.Link

    u, err := url.Parse(profileLink)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

    udf := flag.NewFlagSet("appDash2", flag.ExitOnError)
    var __help bool
    var __addr string
    var __port int

    udf.BoolVar(&__help, "help", false, "show help")
    udf.StringVar(&__addr, "addr", u.Hostname(), "host IP address")

    uPort, _ := strconv.Atoi(u.Port())
    udf.IntVar(&__port, "port", uPort, "port number")
    udf.Parse(args)

    if __help {
        udf.PrintDefaults()
        os.Exit(0)
    }


    mux := http.NewServeMux()

    fs := http.FileServer(http.Dir("dash"))
    mux.Handle("/", disableDirListing(fs))

    // mux.HandleFunc("GET /admin", rootHandler)
    // mux.HandleFunc("GET /admin/", rootHandler)
    // mux.HandleFunc("GET /admin/login", loginHandler)
    // mux.HandleFunc("POST /admin/login", loginHandler)
    // mux.HandleFunc("POST /admin/logout", logoutHandler)

    getUserPage := fmt.Sprintf("GET /%s/{page}", groupName)
    getUserPage2 := fmt.Sprintf("GET /%s/{page}/", groupName)
    mux.HandleFunc(getUserPage, userPage)
    mux.HandleFunc(getUserPage2, userPage)

    getUserBytes := fmt.Sprintf("GET /%s/{page}/bytes", groupName)
    mux.HandleFunc(getUserBytes, userBytes)

    getUserDevice := fmt.Sprintf("GET /%s/{page}/device", groupName)
    mux.HandleFunc(getUserDevice, userDevice)

    getUserProfile := fmt.Sprintf("GET /%s/{page}/{profile}", groupName)
    mux.HandleFunc(getUserProfile, userProfile)

    getUserProfilePng := fmt.Sprintf("GET /%s/{page}/{profile}/png", groupName)
    mux.HandleFunc(getUserProfilePng, qrProfile)

    getGetUsers := fmt.Sprintf("GET /%s/db/reload", groupName)
    mux.HandleFunc(getGetUsers, reGetUsersRest)
    getReloadTemp := fmt.Sprintf("GET /%s/temp/profile/reload", groupName)
    mux.HandleFunc(getReloadTemp, reloadTemplateRest)

    withLog := withLogMiddleware(mux)
    
    addressPort := fmt.Sprintf("%s:%d", __addr, __port)
    addressPortHTTP := fmt.Sprintf("http://%s:%d", __addr, __port)

    server := &http.Server{
        Addr:         addressPort,
        Handler:      withLog,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    osSig := make(chan os.Signal, 1)
    signal.Notify(osSig, os.Interrupt, os.Kill, syscall.SIGTERM)


    chanLT := make(chan error)
    go loadTemplate(chanLT)
    defer close(chanLT)

    chanRT := reloadTemplate()
    defer close(chanRT)

    chanGU := make(chan error)
    go getUsers(chanGU)
    defer close(chanGU)

    chanRU := reGetUsers()
    defer close(chanRU)

    go shutdown(server, osSig, chanLT, chanRT, chanGU)

    config.Log.Info("server.ListenAndServe()", "started", addressPortHTTP)
    err = server.ListenAndServe()
    if err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("dash / server.ListenAndServe() %w", err)
    }

    return nil
}

func disableDirListing(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasSuffix(r.URL.Path, "/") {
            // Check if index.html exists
            indexPath := filepath.Join("dash", r.URL.Path, "index.html")
            if _, err := os.Stat(indexPath); os.IsNotExist(err) {
                // http.Error(w, "403 Forbidden", http.StatusForbidden)
                http.Error(w, "404 Not Found", http.StatusNotFound)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}


////////////////////////////////////////////////////////////////////////////////
// show a static page to user request
////////////////////////////////////////////////////////////////////////////////
func userPage(w http.ResponseWriter, r *http.Request) {
    config.Log.Debug("userPage", "full Http.Request", r)
    config.Log.Debug("mapPages", "=", mapPages)

    page := r.PathValue("page")
    index, ok := mapPages[page]
    if !ok {
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }
    profile := allUsers[index].Profile
    username := allUsers[index].Username
    config.Log.Info("r.PathValue", "page =", page)
    config.Log.Info("[index].Profile", "profile =", profile)
    config.Log.Info("[index].Username", "username =", username)

    agents, err := yaml.Agents.Load()
    activePoolId := agents.Agent.PoolID
    pools, _ := yaml.Pools.Load(activePoolId)

    activeInfoIndex := yaml.ActiveInfoIndex()
    groupName := pools.DB.Info[activeInfoIndex].Name
    address := pools.DB.Info[activeInfoIndex].Profile.Link

    inputlink := fmt.Sprintf("%s/%s/%s/%s", address, groupName, page, profile) 
    imgsrc := fmt.Sprintf("%s/%s/%s/%s/png", address, groupName, page, profile) 
    config.Log.Info("variable", "inputlink =", inputlink)
    config.Log.Info("variable", "imgsrc =", imgsrc)
    data := PageData{
        InputLink: inputlink,
        ImgSrc: imgsrc,
    }

    // Execute template with data
    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        config.Log.Error("tmpl.Execute(...) failed", "cause",  err)
    }
    config.Log.Info("done", "page = ", page)
}

func userDevice(w http.ResponseWriter, r *http.Request) {
    config.Log.Debug("userPage", "full Http.Request", r)
    config.Log.Debug("mapPages", "=", mapPages)

    page := r.PathValue("page")
    index, ok := mapPages[page]
    if !ok {
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }
    profile := allUsers[index].Profile
    username := allUsers[index].Username
    config.Log.Info("r.PathValue", "page =", page)
    config.Log.Info("[index].Profile", "profile =", profile)
    config.Log.Info("[index].Username", "username =", username)

    query := r.URL.Query()
    device := query.Get("d")

    var nextArgs []string
    
    if device == "" {
        nextArgs = []string{"--user", username}
    } else {
        nextArgs = []string{"--user", username, "--device", device}
    }

    agents, err := yaml.Agents.Load()
    activePoolId := agents.Agent.PoolID
    pools, _ := yaml.Pools.Load(activePoolId)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8") 
    err = confiG(nextArgs, pools)
    if err != nil {
        fmt.Fprintf(w, "%s ignored", device)
        return
    }

    err = setUserDevice(device, username)
    if err != nil {
        fmt.Fprintf(w, "%s internal error", device)
        return
    }
    fmt.Fprintf(w, "%s selected", device)

    // dummy := map[string]any{
    //     "user":     "5ff634",
    //     "name":     device,
    //     "sessions": 2147,
    //     "ctime":    1764368190,
    //     "atime":    1767022725,
    //     "etime":    1769552190,
    //     "bytes": map[string]any{
    //         "base":  85899345920,
    //         "used":  98917722,
    //         "pday":  3297257.4,
    //         "limit": false,
    //     },
    //     "seconds": map[string]any{
    //         "base":  5184000,
    //         "used":  2654535,
    //         "limit": false,
    //     },
    // }
    // w.Header().Set("Content-Type", "application/json")
    // w.WriteHeader(http.StatusOK)

    // json.NewEncoder(w).Encode(dummy)
    

}

func userProfile(w http.ResponseWriter, r *http.Request) {
    config.Log.Debug("userProfile", "full Http.Request", r)

    page := r.PathValue("page")
    index, ok := mapPages[page]
    if !ok {
        config.Log.Warn("page not found", "page = ", page)
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }

    dbProfile := allUsers[index].Profile
    profile := r.PathValue("profile")
    if profile != dbProfile {
        config.Log.Warn("404 not found", "profile = ", profile)
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }

    // device := r.PathValue("device")

    // query := r.URL.Query()
    // device := query.Get("d")
    // config.Log.Info("device", "=", device)
    // r.URL.RawQuery = ""

    // username := allUsers[index].Username
    // var nextArgs []string
    // if device == "" {
    //     nextArgs = []string{"--user", username}
    // } else {
    //     nextArgs = []string{"--user", username, "--device", device}
    // }
    // agents, err := yaml.Agents.Load()
    // activePoolId := agents.Agent.PoolID
    // pools, _ := yaml.Pools.Load(activePoolId)
    // err = confiG(nextArgs, pools)
    // if err != nil {
    //     config.Log.Error("unable to update config", "error", err)
    // }

    jsonFile:= fmt.Sprintf("%s%s", "dash", r.URL)
    http.ServeFile(w, r, jsonFile)
    config.Log.Info("done", "jsonFile =", jsonFile)
}

func userBytes(w http.ResponseWriter, r *http.Request) {
    page := r.PathValue("page")
    config.Log.Info("r.PathValue", "page =", page)

    index, ok := mapPages[page]
    if !ok {
        config.Log.Warn("page not found", "page = ", page)
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }
    username := allUsers[index].Username

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        http.Error(w, "Failed to read", http.StatusInternalServerError)
        return
    }

    var user userColumn
    var rowCount int
    err = db.QueryRow("SELECT COUNT(*) from bytes WHERE username = ?;", username).Scan(&rowCount)
    if err != nil {
        config.Log.Error("db.QueryRow failed", "error", err)
        http.Error(w, "Failed to run", http.StatusInternalServerError)
        return
    }
    config.Log.Info("rowCount", "=", rowCount)
    config.Log.Debug("user", "=", user)

    var dummy map[string]any
    if rowCount == 0 {
        dummy = map[string]any{
            "user": username,
            "name":     "name",
            "sessions": 0,
            "ctime":    0,
            "atime":    0,
            "etime":    0,
            "bytes": map[string]any{
                "base":  0,
                "used":  0,
                "pday":  0,
                "limit": false,
            },
            "seconds": map[string]any{
                "base":  0,
                "used":  0,
                "limit": false,
            },
        }

    } else {
        err = db.QueryRow("SELECT * from bytes WHERE username = ?;", username).Scan(
                &user.username,
                &user.realname,
                &user.sessions,
                &user.ctime,
                &user.atime,
                &user.etime,
                &user.bytesBase,
                &user.bytesUsed,
                &user.bytesPday,
                &user.bytesLimit,
                &user.secondBase,
                &user.secondUsed,
                &user.secondLimit,
            )
        if err != nil {
            config.Log.Error("db.QueryRow failed", "error", err)
            http.Error(w, "Failed to run", http.StatusInternalServerError)
            return
        }
        config.Log.Info("user", "=", user)

        dummy = map[string]any{
            "user": user.username,
            "name": user.realname,
            "sessions": user.sessions,
            "ctime": user.ctime,
            "atime": user.atime,
            "etime": user.etime,
            "bytes": map[string]any{
                "base": user.bytesBase,
                "used": user.bytesUsed,
                "pday": user.bytesPday,
                "limit": user.bytesLimit,
            },
            "seconds": map[string]any{
                "base":  user.secondBase,
                "used":  user.secondUsed,
                "limit": user.secondLimit,
            },
        }
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(dummy)
}

////////////////////////////////////////////////////////////////////////////////
// logger
////////////////////////////////////////////////////////////////////////////////
func withLogMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Optional: wrap ResponseWriter to capture status code
        lw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Call the next handler
        next.ServeHTTP(lw, r)

        // Log after the request is handled
        duration := time.Since(start)
        log.Printf(
            "HTTP %s %s %s %d %v",
            r.RemoteAddr,
            r.Method,
            r.URL.Path,
            lw.statusCode,
            duration,
        )
    })
}

// Helper to capture status code
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

//log writer
func (lw *responseWriter) WriteHeader(code int) {
    lw.statusCode = code
    lw.ResponseWriter.WriteHeader(code)
}


////////////////////////////////////////////////////////////////////////////////
// create qr encoded PNG and response it (do not save it to disk)
////////////////////////////////////////////////////////////////////////////////
func qrProfile(w http.ResponseWriter, r *http.Request) {
    page := r.PathValue("page")
    index, ok := mapPages[page]
    if !ok {
        http.Error(w, "404 Not Found", http.StatusNotFound)
        return 
    }
    // profile := allUsers[index].Profile
    username := allUsers[index].Username

    url := filepath.Dir(r.URL.Path)
    addr := "https://" + r.Host
    content := fmt.Sprintf("sing-box://import-remote-profile?url=%s/%s#%s", addr, url, username)

    pngBytes, err := qrcode.Encode(content, qrcode.High, 20*25)
    if err != nil {
        http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
        return
    }

    // Set proper headers
    w.Header().Set("Content-Type", "image/png")
    w.Header().Set("Cache-Control", "no-cache") // optional

    // Optional: inline or attachment
    // w.Header().Set("Content-Disposition", `inline; filename="`+configName+`.png"`)

    // Write PNG bytes directly to response
    w.Write(pngBytes)
    config.Log.Info("done", "qr-link =", content)
}

func reGetUsersRest(w http.ResponseWriter, r *http.Request) {
    chanErr := make(chan error)
    go getUsers(chanErr)
    err := <-chanErr
    if err != nil {
        config.Log.Error("getUsers() error", "=", err)
        fmt.Fprintf(w, "db-reloaded", "error", err)
    } else {
        config.Log.Info("getUsers() done with no error")
        fmt.Fprintf(w, "db-reloaded")
    }
    close(chanErr)

    config.Log.Info("done", "db", "reloaded")
}

func reloadTemplateRest(w http.ResponseWriter, r *http.Request) {
    chanErr := make(chan error)
    go loadTemplate(chanErr)
    err := <-chanErr
    if err != nil {
        config.Log.Error("loadTemplate() error", "=", err)
        fmt.Fprintf(w, "template-reloaded", "error", err)
    } else {
        config.Log.Info("loadTemplate() done with no error")
        fmt.Fprintf(w, "template-reloaded")
    }
    close(chanErr)
}

////////////////////////////////////////////////////////////////////////////////
// retrieve all users from database
////////////////////////////////////////////////////////////////////////////////
func getUsers(chanErr chan error) {
    var err error
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        chanErr <- fmt.Errorf("config / OpenDB(%s) %w", dbFile,  err)
        return
    }

    var rowCount int64
    err = db.QueryRow("SELECT COUNT(*) from users;").Scan(&rowCount)
    if err != nil {
        config.Log.Error("db.QueryRow failed", "error", err)
        chanErr <- fmt.Errorf("db.QueryRow failed %w", err)
        return
    }
    config.Log.Debug("rowCount", "=", rowCount)

    stmt, err := db.Prepare(`SELECT * FROM users;`)
    if err != nil {
        config.Log.Error("db.Prepare", "error", err)
        chanErr <- fmt.Errorf("config / db.Prepare() %w", err)
        return
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            chanErr <- fmt.Errorf("config / stmt.Close() %w", errClose)
            return
        }
    }()

    allUsers = make([]User, rowCount, rowCount)
    mapPages = make(map[string]int, rowCount)

    rows, err := stmt.Query(`SELECT * FROM users;`)
    if err != nil {
        config.Log.Error("stmt.QueryRow failed", "error", err)
        chanErr <- fmt.Errorf("config / stmt.QueryRow() failed %w", err)
        return
    }

    for i := 0; rows.Next(); i++ {
        rows.Scan(
            &allUsers[i].Username,
            &allUsers[i].Realname,
            &allUsers[i].Ctime,
            &allUsers[i].Period,
            &allUsers[i].Traffic,
            &allUsers[i].Password,
            &allUsers[i].Page,
            &allUsers[i].Profile,
            &allUsers[i].Device,
        )
        mapPages[allUsers[i].Page] = i
    }
    config.Log.Info("done", "number of users =", rowCount)
    chanErr <- nil
    return
}

func setUserDevice(device string, username string) (err error) {
    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        err = fmt.Errorf("config / OpenDB(%s) %w", dbFile,  err)
        return err
    }

    _, err = db.Exec(`UPDATE users SET device = ? WHERE username = ?;`, device, username)
    if err != nil {
        config.Log.Error("db.Prepare", "error", err)
        err = fmt.Errorf("config / db.Prepare() %w", err)
        return err
    }

    return nil
}

////////////////////////////////////////////////////////////////////////////////
// load template files
////////////////////////////////////////////////////////////////////////////////
func loadTemplate(chanErr chan error){
    var err error
    tmpl, err = template.ParseFiles(userProfileTemplate)
    if err != nil {
        config.Log.Error("template.ParseFiles()", "error", err)
        chanErr <- err
        return
    }
    config.Log.Info("template.ParseFiles loaded", "file", userProfileTemplate)
    chanErr <- nil
    return
}

////////////////////////////////////////////////////////////////////////////////
// reload template if SIGUSR1 received
////////////////////////////////////////////////////////////////////////////////
func reloadTemplate() chan <-string {
    done := make(chan string)

    go func(){
        reload := make(chan os.Signal, 1)
        signal.Notify(reload, syscall.SIGUSR1)
        defer signal.Stop(reload)
        loop:
        for {
            select {
                case sig := <-reload:
                config.Log.Info("signal received", "sig =", sig)
                chanErr := make(chan error)
                go loadTemplate(chanErr)
                err := <-chanErr
                if err != nil {
                    config.Log.Error("loadTemplate() error", "=", err)
                } else {
                    config.Log.Info("loadTemplate() done with no error")
                }
                close(chanErr)
                case cause := <-done:
                config.Log.Warn("forced exit", "cause", cause)
                break loop 
            }
        }
        config.Log.Info("reload go-routine ended")
        return
    }()

    config.Log.Info("go-routine started ...")
    return done
}

////////////////////////////////////////////////////////////////////////////////
// reload getUsers if SIGUSR2 received
////////////////////////////////////////////////////////////////////////////////
func reGetUsers() chan <-string {
    done := make(chan string)

    go func(){
        reload := make(chan os.Signal, 1)
        signal.Notify(reload, syscall.SIGUSR2)
        defer signal.Stop(reload)
        loop:
        for {
            select {
                case sig := <-reload:
                config.Log.Info("signal received", "sig =", sig)
                chanErr := make(chan error)
                go getUsers(chanErr)
                err := <-chanErr
                if err != nil {
                    config.Log.Error("getUsers() error", "=", err)
                } else {
                    config.Log.Info("getUsers() done with no error")
                }
                close(chanErr)
                case cause := <-done:
                config.Log.Warn("forced exit", "cause", cause)
                break loop 
            }
        }
        config.Log.Info("reload go-routine ended")
        return
    }()

    config.Log.Info("go-routine started ...")
    return done
}

////////////////////////////////////////////////////////////////////////////////
// waiting for shutdown signals or errors
////////////////////////////////////////////////////////////////////////////////
func shutdown(server *http.Server, osSig chan os.Signal,
    chanLT chan error, chanRT chan <-string,
    chanGU chan error){
    config.Log.Warn("waiting for shutdown error or signals")
    loop:
    for {
        select {
            case lte := <- chanLT:
            if lte != nil {
                config.Log.Error("loadTemplate() error", "=", lte)
                break loop
            }
            config.Log.Info("loadTemplate() done with no error")

            case gue := <- chanGU:
            if gue != nil {
                config.Log.Error("getUsers() error", "=", gue)
                break loop
            }
            config.Log.Info("getUsers() done with no error")

            case sig := <-osSig:
            config.Log.Warn("go func / signal received from osSig", "=", sig)
            break loop
        }
    }
    chanRT <- "server is shutting down"
    // chanRU <- "server is shutting down"

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    err := server.Shutdown(ctx);
    if err != nil {
        config.Log.Error("go func / shutdown failed", "error", err)
    } else {
        config.Log.Warn("go func / graceful shutdown")
    }
}
