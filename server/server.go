package server

import (
    "os"
    "fmt"
    //"log"
    "io"

    "github.com/shakibamoshiri/proxgo/config"
)

var yaml config.YamlFiles

func Parse(args []string) (err error) {
    config.Log.Debug("args", "=", args)
    config.Log.Debug("AgentID", "=", config.AgentID)

    agents, err := yaml.Agents.Load()
    if err != nil {
        return fmt.Errorf("server >> %w", err)
    }
    config.Log.Debug("agents.Agent.PoolID", "=", agents.Agent.PoolID)

    activePoolId := agents.Agent.PoolID
    pools, err := yaml.Pools.Load(activePoolId)
    if err != nil {
        return fmt.Errorf("server >> %w", err)
    }
    config.Log.Debug("pools", "=", pools)
 
    help := config.NewHelp()
    if len(args) == 1 {
        help.
        For("server").
        Exit(0)
    }

    nextCmd := args[1]
    nextArgs := args[2:]
    var errParent error = nil
    switch nextCmd {
        case "check":
            errParent = check(nextArgs, &yaml.Pools, os.Stdout)
        case "fetch":
            errParent = fetch(nextArgs, &yaml.Pools, os.Stdout)
        case "stats":
            errParent = stats(nextArgs, &yaml.Pools)
        case "reset":
            errParent = reset(nextArgs, &yaml.Pools, os.Stdout)
        default:
            help.
            For("server").
            Say("%s command not found", nextCmd).
            Exit(1)
    }

    if errParent != nil {
        return fmt.Errorf("server >> %w", errParent)
    }

    return nil
}

func Run(fn string, args []string, pc *config.Pools, dev io.Writer) error {
    switch fn {
        case "check":
            return check(args, pc, dev)
        case "fetch":
            return fetch(args, pc, dev)
        default:
            return fmt.Errorf("server.Run(%s) not found", fn)
    }
}
