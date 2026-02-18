package main

import (
	"fmt"
	"os"
	//"strings"
	"flag"
	"log"
	"path/filepath"
	// "time"

	"github.com/shakibamoshiri/proxgo/agent"
	"github.com/shakibamoshiri/proxgo/app"
	"github.com/shakibamoshiri/proxgo/config"
	"github.com/shakibamoshiri/proxgo/data"
	"github.com/shakibamoshiri/proxgo/server"
	"github.com/shakibamoshiri/proxgo/tell"
	"github.com/shakibamoshiri/proxgo/user"
)

func main() {

	appName := filepath.Base(os.Args[0])

	flag.Parse()
	if len(os.Args) == 1 {
		flag.PrintDefaults()
	}
	config.Parse()

	config.Log.Debug("os first arguments", "appName", appName)
	config.Log.Debug("all arguments", "os.Args", os.Args)
	config.Log.Debug("arguments after flag.Parse", "flag.Args()", flag.Args())

	help := config.NewHelp()
	if config.HelpFlag == true {
		config.Log.Debug("os flag set", "--help", config.HelpFlag)
		flag.PrintDefaults()
		help.For("main").Exit(0)
	}

	args := flag.Args()
	argsLen := len(args)
	if argsLen == 0 {
		help.For("main").Exit(0)
	}

	activeCommand := args[0]
	config.Log.Debug("next subcommand after main", "activeCommand", activeCommand)
	var errChild error = nil
	var errMain error = nil
	switch activeCommand {
	case "app":
		errChild = app.Parse(args)
	case "server":
		errChild = server.Parse(args)
	case "agent":
		errChild = agent.Parse(args)
	case "user":
		_, errChild = user.Parse(args)
	case "data":
		errChild = data.Parse(args)
	case "tell":
		errChild = tell.Parse(args)
	default:
		help.
			For("main").
			Say("%s command not found", activeCommand).
			Exit(1)
	}

	errMain = config.CloseDB()
	if errMain != nil {
		config.Log.Error("main()", "errMain", errMain)
	}

	if errChild != nil {
		finalError := fmt.Errorf("main >> %w", errChild)
		log.Fatal(finalError)
	}

	config.Log.Debug("end of main()")
}
