package app

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	// "strings"
)

func dash() error {
	mux := http.NewServeMux()

	// API endpoints first (they take priority)
	mux.HandleFunc("POST /user/create", createUserHandler)
	mux.HandleFunc("POST /login", loginPostHandler)
	// Add more API routes here...

	// Static site handler - serves from "site" directory
	mux.HandleFunc("/", staticSiteHandler)

	log.Println("Server starting on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", mux))
    return nil
}

// Main handler for all static content
func staticSiteHandler(w http.ResponseWriter, r *http.Request) {
	// Prevent directory traversal attacks
	path := filepath.Join("site", filepath.Clean(r.URL.Path))

	// Check if requested path is a directory → try to serve index.html
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		indexPath := filepath.Join(path, "index.html")
		if indexInfo, err := os.Stat(indexPath); err == nil && !indexInfo.IsDir() {
			http.ServeFile(w, r, indexPath)
			return
		}
	}

	// Otherwise, try to serve the file directly (assets, etc.)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		http.ServeFile(w, r, path)
		return
	}

	// Optional: For SPA-style fallback (if you want unknown paths to load root index.html)
	// http.ServeFile(w, r, "site/index.html")
	// return

	// 404 if nothing found
	http.NotFound(w, r)
}

// Example API handler
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"username":"rand123","password":"secret456"}`))
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	// your login logic
}

