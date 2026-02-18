package app

import (
    "os"
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)

var yaml = config.NewYamlFile()

func Parse(args []string) (err error) {
    config.Log.Debug("args []string", "=", args)
    config.Log.Debug("AgentID", "=", config.AgentID)


    agents, err := yaml.Agents.Load()
    if err != nil {
        return fmt.Errorf("app >> %w", err)
    }
    config.Log.Debug("agents.Agent.PoolID", "=", agents.Agent.PoolID)

    activePoolId := agents.Agent.PoolID
    pools, err := yaml.Pools.Load(activePoolId)
    if err != nil {
        return fmt.Errorf("app >> %w", err)
    }
    config.Log.Debug("pools", "=", pools)

    argsLen := len(args)
    help := config.NewHelp()
    if argsLen == 1 {
        help.
        For("app").
        Exit(0)
    }

    nextCmd := args[1]
    nextArgs := args[2:]
    switch nextCmd {
        case "run":
            err = run(nextArgs, &yaml.Pools, os.Stdout)
        case "service":
            err = service(nextArgs, &yaml.Pools, os.Stdout)
        case "dash":
            err = dash()
        case "dash2":
            err = dash2(nextArgs)
        default:
            help.
            For("app").
            Say("%s command not found", nextCmd).
            Exit(1)
    }

    if err != nil {
        return fmt.Errorf("app >> %w", err)
    }

    return nil
}
