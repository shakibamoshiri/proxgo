package data

import (
	"fmt"
	"io"
	"os"

	"github.com/shakibamoshiri/proxgo/config"
)

const tableMax int = 4

type table struct {
	name string
	cmd  string
}

var yaml config.YamlFiles

func Parse(args []string) (err error) {
	config.Log.Debug("args []string", "=", args)
	config.Log.Debug("AgentID", "=", config.AgentID)

	agents, err := yaml.Agents.Load()
	if err != nil {
		return fmt.Errorf("user >> %w", err)
	}
	config.Log.Debug("agents.Agent.PoolID", "=", agents.Agent.PoolID)

	activePoolId := agents.Agent.PoolID
	pools, err := yaml.Pools.Load(activePoolId)
	if err != nil {
		return fmt.Errorf("user >> %w", err)
	}
	config.Log.Debug("pools", "=", pools)

	help := config.NewHelp()
	if len(args) == 1 {
		help.
			For("data").
			Exit(0)
	}

	nextCmd := args[1]
	nextArgs := args[2:]
	switch nextCmd {
	case "setup":
		err = setup(nextArgs, &yaml.Pools, os.Stdout)
	case "wipe":
		err = wipe(nextArgs)
	case "ich":
		err = icheck()
	case "sync":
		err = sync(nextArgs, pools, os.Stdout)
	default:
		help.
			For("data").
			Say("%s command not found", nextCmd).
			Exit(1)
	}

	if err != nil {
		config.Log.Error("data", "subcommand error", err)
		return fmt.Errorf("data >> %w", err)
	}

	return nil
}

func Run(fn string, args []string, pc *config.Pools, out io.Writer) error {
	switch fn {
	case "setup":
		return setup(args, pc, out)
	default:
		return nil
	}
}
