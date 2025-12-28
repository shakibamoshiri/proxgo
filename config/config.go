package config

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"database/sql"
	_ "modernc.org/sqlite"
)

const (
	OneKB = 1 << 10
	OneMB = 1 << 20
	OneGB = 1 << 30

	AgentPath  = ".prox/agent"
	PoolPath   = ".prox/pool"
	LogPath    = ".prox/log"
	ClientPath = ".prox/cc"
	TmpPath    = ".prox/tmp"
	DbPath     = "proxdb"

	WebPath = "web"

	SsmApiPath      = "server/v1"
	SsmApiPathTest  = "delay/11000"
	SsmApiPathUsers = SsmApiPath + "/users"
	SsmApiPathStats = SsmApiPath + "/stats"
	ClientTimeout   = 10
)

var (
	HelpFlag  bool
	LogFlag   string
	ColorFlag bool
	Version   bool
	AgentID   int

	AgentPrefix = 0
	Log         *slog.Logger
)

type AnyArg map[string]string

func init() {
	flag.BoolVar(&Version, "version", false, "show version")
	flag.BoolVar(&HelpFlag, "help", false, "show help menu")
	flag.BoolVar(&ColorFlag, "color", true, "enable color for -log")
	flag.StringVar(&LogFlag, "log", "", "enable log [error|warn|info|debug]")
	flag.IntVar(&AgentID, "aid", 0, "agent id")
}

func Parse() {
	Log = NewLogger(LogFlag)

	if ColorFlag {
		Log.Debug("os flag set", "--color", ColorFlag)
	}

	if Version == true {
		Log.Debug("os flag set", "--version", Version)
		println("version: 3.2.1")
		os.Exit(0)
	}

	if AgentID == 0 {
		fmt.Println("Agent ID must be set, --aid ?")
		os.Exit(1)
	}
}

func FuncName() string {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		panic("FuncName runtime error")
	}
	fmt.Fprint(io.Discard, file, line)

	fn := runtime.FuncForPC(pc).Name()
	fn = filepath.Base(fn)
	fmt.Println(fn)

	return fn
}

// var db *sql.DB
// var once sync.Once
//
// // Database Connections
// func OpenDbOnce(dbFile string) (_ *sql.DB, err error) {
//     Log.Info("dbFile", dbFile)
//
//     once.Do(func(){
//         Log.Info("OpenDbOnce() / once.Do(...)", "dbFile", dbFile)
//         db, err = sql.Open("sqlite", dbFile)
//     })
//     Log.Warn("*sql.DB", "value", db)
//     if err != nil {
//         return nil, err
//     }
//     /* we should not close DB here */
//     // defer func(){
//     //     errClose := db.Close()
//     //     if errClose != nil {
//     //         err = errClose
//     //     }
//     // }()
//     err = db.Ping()
//     if err != nil {
//         return nil, err
//     }
//
//     // db is a pointer
//     // var db *sql.DB
//     // sql.Open returns *DB => *sql.DB
//     return db, nil
// }

// Database Connections
var db *sql.DB

func OpenDB(dbFile string) (_ *sql.DB, err error) {
	const title string = "Open database connection pool"
	if db != nil {
		Log.Warn(title + " ignored (already opened)")
		Log.Debug(title+" ignored (already opened)", "*sql.DB", db)
		return db, nil
	}

	Log.Info(title, "dbFile", dbFile)
	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		Log.Error("sql.Open(%s) failed %w", dbFile, err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		Log.Error("db.Ping() failed %w", err)
		return nil, err
	}

	Log.Info(title + " done")
	Log.Debug(title+" done", "*sql.DB", db)
	return db, nil
}

func CloseDB() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	if err != nil {
		Log.Error("CloseDB", "db.Close()", err)
		return fmt.Errorf("CloseDB() / db.Close() %w", err)
	}
	Log.Info("database connection pool closed")
	Log.Debug("Database connection pool closed", "*sql.DB", db)
	return nil
}

func ParseArgs(osArgs []string, flags []string) map[string]string {
	line := strings.Join(osArgs, " ")

	space := regexp.MustCompile(`\s+`)
	line = space.ReplaceAllString(line, " ")

	//fmt.Println(line)

	re := regexp.MustCompile(`((?:-?-)?\w+(?:=|\s+)?)([^- ]+)?`)
	args := re.FindAllString(line, -1) // -1 = find ALL matches
	result := make(map[string]string, len(flags))
	for _, arg := range args {
		for _, flag := range flags {
			if strings.HasPrefix(arg, flag) {
				fmt.Printf("[%s]\n", arg)
				value := strings.TrimPrefix(arg, flag)
				value = strings.TrimPrefix(value, " ")
				value = strings.TrimPrefix(value, "=")
				value = strings.TrimSuffix(value, " ")
				result[flag] = value
			}
		}
	}
	return result
}
