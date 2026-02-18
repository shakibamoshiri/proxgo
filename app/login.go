package app

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"sync"
	"time"
    "html/template"
    "strings"

	"crypto/rand"
	"encoding/base64"

    "github.com/shakibamoshiri/proxgo/config"
)

const userLoginPage = "login.html"
const maxLoginTry = 2
var loginTry int

var tmpl *template.Template
type PageData struct {
    Title string
}

type Admin struct {
    ID int
    Phone int64
    Name string
    Username string
    Password string
    Access bool
}

var admins []Admin

// In-memory session store (for single-instance apps only)
// In production with multiple servers → use Redis or database
type Session struct {
	Username  string
	ExpiresAt time.Time
}

var (
	sessions = make(map[string]Session) // sessionID → Session
	mu       sync.RWMutex               // protects the sessions map
)

// Generate a cryptographically secure random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32) // 256 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Set session cookie + store data in memory
func setSession(w http.ResponseWriter, username string) error {
	sessionID, err := generateSessionID()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(1 * time.Hour) // adjust duration as needed

	mu.Lock()
	sessions[sessionID] = Session{
		Username:  username,
		ExpiresAt: expiresAt,
	}
	mu.Unlock()

	// Set secure cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,              // set true when using HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// Retrieve session from cookie (and check expiration)
func getSession(r *http.Request) (Session, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return Session{}, false
	}

	mu.RLock()
	sess, exists := sessions[cookie.Value]
	mu.RUnlock()

	if !exists || time.Now().After(sess.ExpiresAt) {
		// Clean up expired session
		if exists {
			mu.Lock()
			delete(sessions, cookie.Value)
			mu.Unlock()
		}
		return Session{}, false
	}
	return sess, true
}

// Clear session (logout)
func clearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		mu.Lock()
		delete(sessions, cookie.Value)
		mu.Unlock()
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}


// Login page (GET = show form, POST = process login)
func loginHandler(w http.ResponseWriter, r *http.Request) {
    // globally deny if reached maxLoginTry (all admins)
    if loginTry >= maxLoginTry {
        http.Error(w, "No access", http.StatusForbidden)
        return
    }

	// If already logged in → redirect to dashboard
	if _, ok := getSession(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

    // POST /login
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

        var dbUser string
        var dbPass string
        var adminAccess bool
        for _, admin := range admins {
            if admin.Username == username && admin.Password == password {
                dbUser = admin.Username
                dbPass = admin.Password
                adminAccess = admin.Access
            }
        }

        config.Log.Debug("admin credentials match", "dbUser = ", dbUser, "dbPass = ", dbPass)
		if username == dbUser && password == dbPass {
            if adminAccess == false {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
            }
			if err := setSession(w, username); err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

            loginTry = 0
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

        // GET /login
		// Invalid credentials
		// renderLoginForm(w, "Invalid username or password")
        data := PageData{ Title: "Login Failed!" }
        loginTry += 1

        err := tmpl.Execute(w, data)
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            log.Println("tmpl.Execute(...) failed", "cause",  err)
        }
		return
	}

	// GET request → show form
	// renderLoginForm(w, "")

    // http.ServeFile(w, r, "login.html")

    data := PageData{
        Title: "Login",
    }

    err := tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        log.Println("tmpl.Execute(...) failed", "cause",  err)
    }

}


func loadTemplate(){
    var err error
    tmpl, err = template.ParseFiles(userLoginPage)
    if err != nil {
        log.Printf("loadTemplate(%s) %s\n", userLoginPage,  err)
        return
    }
}


func loadTableAdmins() (err error){

    dbFile := fmt.Sprintf("./%s/%d.sqlite3", config.DbPath, config.AgentID)
    db, err := config.OpenDB(dbFile)
    if err != nil {
        config.Log.Error("config.OpenDB failed", "error", err)
        return err
    }

    var rowCount int64
    err = db.QueryRow("SELECT COUNT(*) from admins;").Scan(&rowCount)
    if err != nil {
        config.Log.Error("db.QueryRow failed", "error", err)
        err = fmt.Errorf("db.QueryRow failed %w", err)
        return err
    }
    config.Log.Debug("rowCount", "=", rowCount)

    stmt, err := db.Prepare(`SELECT * FROM admins;`)
    if err != nil {
        config.Log.Error("db.Prepare", "error", err)
        err = fmt.Errorf("config / db.Prepare() %w", err)
        return err
    }
    defer func(){
        errClose := stmt.Close()
        if errClose != nil {
            err = fmt.Errorf("config / stmt.Close() %w", errClose)
        }
    }()

    admins = make([]Admin, rowCount, rowCount)

    rows, err := stmt.Query(`SELECT * FROM admins;`)
    if err != nil {
        config.Log.Error("stmt.QueryRow failed", "error", err)
        err = fmt.Errorf("config / stmt.QueryRow() failed %w", err)
        return err
    }

    for i := 0; rows.Next(); i++ {
        err = rows.Scan(
            &admins[i].ID,
            &admins[i].Name,
            &admins[i].Phone,
            &admins[i].Username,
            &admins[i].Password,
            &admins[i].Access,
        )
        if err != nil {
            config.Log.Error("loadTableAdmins / rows.Next() failed", "error", err)
            return err
        }
    }
    config.Log.Info("done", "number of admins =", rowCount)
    return err
}





func renderLoginForm(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	msg := ""
	if errorMsg != "" {
		msg = fmt.Sprintf("<p style='color:red;'>%s</p>", html.EscapeString(errorMsg))
	}

	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Login</title></head>
<body>
    <h2>Login!!!!</h2>
    %s
    <form method="POST" action="/login">
        <label>Username: <input type="text" name="username" required></label><br><br>
        <label>Password: <input type="password" name="password" required></label><br><br>
        <button type="submit">Log In</button>
    </form>
    <p><small>ask for credentials if you want to book class</small></p>
</body>
</html>`, msg)
}

// Logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func rootHandler2(w http.ResponseWriter, r *http.Request) {
	_, ok := getSession(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

    // http.ServeFile(w, r, "index.html")
    staticSiteHandler2(w, r)

}


// --------------------- Main ---------------------
func init(){
    loadTemplate()
    // err := loadTableAdmins()
    // if err != nil {
    //     panic("app/login.go loadTableAdmins() failed")
    // }
}
// Public routes
// mux.HandleFunc("GET /", rootHandler2)
// mux.HandleFunc("GET /login", loginHandler)       // handles both GET and POST
// mux.HandleFunc("POST /login", loginHandler)       // handles both GET and POST
// mux.HandleFunc("POST /logout", logoutHandler)


func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if _, ok := getSession(r); !ok {
            // If it's an API request (expects JSON), return 401
            if strings.HasPrefix(r.URL.Path, "/user/") ||
               r.Header.Get("Accept") == "application/json" ||
               r.Header.Get("Content-Type") == "application/json" {
                // http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            // Otherwise (HTML page request), redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // User is authenticated → continue
        next.ServeHTTP(w, r)
    })
}
