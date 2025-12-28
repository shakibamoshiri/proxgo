package config

import (
	"fmt"
	"os"
)

type help struct {
	cmd string
	err error
}

func NewHelp() *help {
	return &help{}
}

func (h *help) For(cmd string) *help {
	switch cmd {
	case "main":
		mainHelp(cmd)
	case "app":
		appHelp(cmd)
	case "server":
		serverHelp(cmd)
	case "agent":
		agentHelp(cmd)
	case "user":
		userHelp(cmd)
	case "data":
		dataHelp(cmd)
	case "tell":
		tellHelp(cmd)
	default:
		fmt.Printf("match for %s command not found", h.cmd)
	}
	return h
}

func (h *help) Say(format string, msg ...any) *help {
	h.err = fmt.Errorf(format, msg...)
	return h
}

func (h *help) Exit(code int) {
	if h.err != nil {
		fmt.Println("error,", h.err)
	}
	os.Exit(code)
}

func mainHelp(cmd string) {
	cmdOrder := [...]string{"app", "server", "agent", "user", "data", "tell"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["app"] = "manage the app"
	subCmds["server"] = "manage servers"
	subCmds["agent"] = "manage agents"
	subCmds["user"] = "manage users"
	subCmds["data"] = "manage database"
	subCmds["tell"] = "manage notification (telegram)"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func appHelp(cmd string) {
	cmdOrder := [...]string{"run", "service", "setup", "test"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["run"] = "run the app once"
	subCmds["service"] = "run the app as service"
	subCmds["setup"] = "setup the app"
	subCmds["test"] = "test the app"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func serverHelp(cmd string) {
	cmdOrder := [...]string{"check", "fetch", "list", "stats", "reset"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["check"] = "check if servers are up?"
	subCmds["fetch"] = "fetch from servers"
	subCmds["list"] = "list of servers"
	subCmds["stats"] = "stats of servers"
	subCmds["reset"] = "delete all users of all servers"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func agentHelp(cmd string) {
	cmdOrder := [...]string{"setup", "pool", "server"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["setup"] = "setup agents"
	subCmds["pool"] = "list of pools or select a pool"
	subCmds["server"] = "list of servers"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func userHelp(cmd string) {
	cmdOrder := [...]string{"create", "delete", "renew", "setup", "limit", "archive", "config", "lock", "unlock"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["create"] = "create a new user"
	subCmds["delete"] = "delete a user (online user)"
	subCmds["renew"] = "renew a user (if deleted)"
	subCmds["setup"] = "setup a user (if connected)"
	subCmds["limit"] = "disable a user (if passed limits)"
	subCmds["archive"] = "archive a user (if connected)"
	subCmds["config"] = "create configuration for a user"
	subCmds["lock"] = "lock a user (if connected)"
	subCmds["unlock"] = "unlock a user (if connected)"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func dataHelp(cmd string) {
	cmdOrder := [...]string{"setup", "wipe", "ich", "sync"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["setup"] = "setup a new database"
	subCmds["wipe"] = "wipe all local data (disk)"
	subCmds["ich"] = "check user integrity"
	subCmds["sync"] = "sync users from/to db/servers"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}

func tellHelp(cmd string) {
	cmdOrder := [...]string{"test", "msg", "doc"}
	subCmds := make(map[string]string, len(cmdOrder))
	subCmds["test"] = "send a test message"
	subCmds["msg"] = "send a message"
	subCmds["doc"] = "send a document"

	fmt.Printf("%s commands\n", cmd)
	for _, v := range cmdOrder {
		fmt.Printf("%-20s %s\n", v, subCmds[v])
	}
}
