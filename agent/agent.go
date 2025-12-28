package agent

import (
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)


var yaml config.YamlFiles
func Parse(args []string) (err error) {
    config.Log.Info("parsing agent args", "args", args)

    agents, err := yaml.Agents.Load()
    if err != nil {
        return fmt.Errorf("tell >> %w", err)
    }
    activePoolID := agents.Agent.PoolID
    config.Log.Debug("activePoolID", "=", activePoolID)

    pools, err := yaml.Pools.Load(activePoolID)
    if err != nil {
        return fmt.Errorf("tell >> %w", err)
    }
    config.Log.Debug("pools", "=", pools)

    argsLen := len(args)
    help := config.NewHelp()
    if argsLen == 1 {
        help.
        For("agent").
        Exit(0)
    }

    nextCmd := args[1]
    nextArgs := args[2:]
    config.Log.Info("nextArgs", "=", nextArgs)
    switch nextCmd {
        case "setup":
            err = setup(nextArgs)
        case "pool":
            err = pool(nextArgs)
        case "server":
            err = server(nextArgs)
        default:
            help.
            For("agent").
            Say("%s command not found", nextCmd).
            Exit(1)
    }

    if err != nil {
        return err
    }

    return nil

    return nil
}
