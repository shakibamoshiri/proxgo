package app

import (
	"log"
	"net/http"
	"os"
    "fmt"
	"path/filepath"
	"strings"
    "encoding/json"
    "flag"
    "time"

	"github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "github.com/shakibamoshiri/proxgo/user"
    "github.com/shakibamoshiri/proxgo/config"
)

func dash2(args []string) error {
    ad2 := flag.NewFlagSet("appDash2", flag.ExitOnError)
    var __help bool
    var __addr string
    var __port int

    ad2.BoolVar(&__help, "help", false, "show help")
    ad2.StringVar(&__addr, "addr", "127.0.0.1", "address default: 127.0.0.1")
    ad2.IntVar(&__port, "port", 8000, "port number default: 8000")
    ad2.Parse(args)

    if __help {
        ad2.PrintDefaults()
        os.Exit(0)
    }


    err := loadTableAdmins()
    if err != nil {
        return err
    }

	r := chi.NewRouter()

	// Optional: logging middleware
	r.Use(middleware.Logger)

	r.Get("/login", loginHandler)
	r.Post("/login", loginHandler)
	r.Get("/logout", logoutHandler)

	// API routes (take priority over static files)
    //r.Get("/*", rootHandler2)
	//r.Post("/user/create", userCreate2)
	//r.Get("/user/list", userList2)
	//r.Get("/user/page", userPage2)

r.Route("/", func(r chi.Router) {
    // Apply authentication middleware to ALL routes inside this group
    r.Use(authMiddleware)

    // API endpoints (now protected)
    r.Post("/user/create", userCreate2)
    r.Post("/user/delete", userDelete2)
    r.Post("/user/renew", userRenew2)
    r.Post("/user/lock", userLock2)
    r.Post("/user/unlock", userUnlock2)
    r.Get("/user/list", userList2)
    r.Get("/user/page", userPage2)

    // Frontend pages / SPA fallback (also protected)
    r.Get("/*", rootHandler2) // serves index.html for client-side routing
})
    addressPort := fmt.Sprintf("%s:%d", __addr, __port)
    addressPortHTTP := fmt.Sprintf("http://%s:%d", __addr, __port)
	log.Println("Server starting on ", addressPortHTTP)
	log.Fatal(http.ListenAndServe(addressPort, r))
    return nil
}

// Handles all static file serving with proper index.html support
func staticSiteHandler2(w http.ResponseWriter, r *http.Request) {
	// Prevent directory traversal
	path := filepath.Clean(r.URL.Path)
	if strings.HasPrefix(path, "/.") || strings.Contains(path, "\\") {
		http.NotFound(w, r)
		return
	}

	// Full path on disk
	fsPath := filepath.Join("site", path)

	// Check if it's a directory → serve index.html
	if info, err := os.Stat(fsPath); err == nil && info.IsDir() {
		indexPath := filepath.Join(fsPath, "index.html")
		if indexInfo, err := os.Stat(indexPath); err == nil && !indexInfo.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}
	}

	// Otherwise, try to serve the requested file
	if info, err := os.Stat(fsPath); err == nil && !info.IsDir() {
		// Optional: set cache headers for assets
		// if strings.Contains(path, "/assets/") {
		//     w.Header().Set("Cache-Control", "public, max-age=31536000")
		// }
		http.ServeFile(w, r, fsPath)
		return
	}

	// Optional: SPA fallback (uncomment if you want all 404s to serve root index.html)
	// http.ServeFile(w, r, "site/index.html")
	// return

	// 404 Not Found
	http.NotFound(w, r)
}

func userCreate2(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        Name     string `json:"name"`
        Device   string `json:"device"`
    }

    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    config.Log.Warn("request payload", "=", payload)

    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "create", "--name", payload.Name}
    res, err := user.Parse(args)
    if err != nil {
        if len(res) == 0 {
            res = make([]map[string]any, 1, 1)
        }
        res[0] = map[string]any {
            "error": fmt.Sprintf("%s", err),
        }
        config.Log.Debug("user create payload (response)", "=", res[0])
        json.NewEncoder(w).Encode(res[0])
        return
    }

////////////////////////////////////////////////////////////////////////////////
// request to user dash to reload user list
////////////////////////////////////////////////////////////////////////////////

    agents, err := yaml.Agents.Load()
    activePoolId := agents.Agent.PoolID
    pools, _ := yaml.Pools.Load(activePoolId)

    activeInfoIndex := yaml.ActiveInfoIndex()
    groupName := pools.DB.Info[activeInfoIndex].Name
    address := pools.DB.Info[activeInfoIndex].Profile.Link
    endpoint := fmt.Sprintf("%s/%s/db/reload", address, groupName)
    config.Log.Info("reload user list", "address", address)
    config.Log.Warn("REST endpoint", "endpoint", endpoint)

	client := &http.Client{
		Timeout: (time.Second * config.ClientTimeout),
	}

    resp, err := client.Get(endpoint)
    if err != nil {
        res[0]["error"] = fmt.Sprintf("%s", err)
    } else {
        errClose := resp.Body.Close()
        if errClose != nil {
            err = errClose
            config.Log.Error("resp.Body.Close", "error", err)
        }
    }

    // io.Copy(&body, resp.Body)
    // if resp.StatusCode != 200 {
    //     err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, body.String())
    //     return err
    // }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res[0])
}

func userDelete2(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        User     string `json:"username"`
    }

    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    config.Log.Warn("request payload", "=", payload)

    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "delete", "--user", payload.User}
    res, err := user.Parse(args)
    if err != nil {
        config.Log.Error("res", "=", res)
        config.Log.Error("err", "=", err)
        if len(res) == 0 {
            res = make([]map[string]any, 1, 1)
        }
        res[0] = map[string]any {
            "error": fmt.Sprintf("%s", err),
        }
        json.NewEncoder(w).Encode(res[0])
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res[0])
}

func userRenew2(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        User     string `json:"username"`
    }

    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    config.Log.Warn("request payload", "=", payload)

    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "renew", "--user", payload.User}
    res, err := user.Parse(args)
    if err != nil {
        config.Log.Error("res", "=", res)
        config.Log.Error("err", "=", err)
        if len(res) == 0 {
            res = make([]map[string]any, 1, 1)
        }
        res[0] = map[string]any {
            "error": fmt.Sprintf("%s", err),
        }
        json.NewEncoder(w).Encode(res[0])
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res[0])
}

func userLock2(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        User     string `json:"username"`
    }

    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    config.Log.Warn("request payload", "=", payload)

    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "lock", "--user", payload.User}
    res, err := user.Parse(args)
    if err != nil {
        config.Log.Error("res", "=", res)
        config.Log.Error("err", "=", err)
        if len(res) == 0 {
            res = make([]map[string]any, 1, 1)
        }
        res[0] = map[string]any {
            "error": fmt.Sprintf("%s", err),
        }
        json.NewEncoder(w).Encode(res[0])
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res[0])
}

func userUnlock2(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        User     string `json:"username"`
    }

    err := json.NewDecoder(r.Body).Decode(&payload)
    if err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    config.Log.Warn("request payload", "=", payload)

    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "lock", "--user", payload.User}
    res, err := user.Parse(args)
    if err != nil {
        config.Log.Error("res", "=", res)
        config.Log.Error("err", "=", err)
        if len(res) == 0 {
            res = make([]map[string]any, 1, 1)
        }
        res[0] = map[string]any {
            "error": fmt.Sprintf("%s", err),
        }
        json.NewEncoder(w).Encode(res[0])
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res[0])
}

func userList2(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "list"}
    res, err := user.Parse(args)
    if err != nil {
        json.NewEncoder(w).Encode(res)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res)
}

func userPage2(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    args := []string{"user", "page"}
    res, err := user.Parse(args)
    if err != nil {
        json.NewEncoder(w).Encode(res)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(res)
}

func loginPostHandler2(w http.ResponseWriter, r *http.Request) {
	// Your login logic here
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
