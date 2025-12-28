package user

import (
    "fmt"
    "io"
    "os"
    "time"
    "context"

    "github.com/shakibamoshiri/proxgo/config"
)

var yaml config.YamlFiles

const timeout time.Duration = 5

type userColumn struct {
    username    string
    realname    string
    sessions     int64

    ctime       int64
    atime       int64
    etime       int64

    bytesBase   int64
    bytesUsed   int64
    bytesPday   int64
    bytesLimit  bool

    secondBase  int64
    secondUsed  int64
    secondLimit bool
}

type userData struct {
    row     userColumn
    msg     string
    err     error
}

type LimitPipe struct {
    next    <-chan userData
}


func Parse(args []string) (err error) {
    config.Log.Debug("args", "=", args)
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
    argsLen := len(args)
    if argsLen == 1 {
        help.For("user").Exit(0)
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout * time.Second)
    defer cancel()

    nextCmd := args[1]
    nextArgs := args[2:]
    switch nextCmd {
        case "create":
            err = create(nextArgs)
        case "delete":
            err = _delete(nextArgs)
        case "list":
            err = list()
        case "setup":
            err = setup(nextArgs, os.Stdout)
        case "limit":
            err = limit(ctx, nextArgs, os.Stdout)
        case "limit_b":
            err = limit_b(ctx, nextArgs, os.Stdout)
        case "limit2":
            err = limit2(ctx, nextArgs, os.Stdout)
        case "limit3":
            err = limit3(ctx, nextArgs, os.Stdout)
        case "limit4":
            err = limit4(ctx, nextArgs, os.Stdout)
        case "archive":
            err = archive(nextArgs)
        case "config":
            err = confiG(nextArgs)
        case "renew":
            err = renew(nextArgs)
        case "lock":
            err = lock(nextArgs)
        case "unlock":
            err = unlock(nextArgs)
        default:
            help.
            For("user").
            Say("%s command not found", nextCmd).
            Exit(1)
            
    }

    if err != nil {
        return fmt.Errorf("user >> %w", err)
    }

    return nil
}

func Run(fn string, args []string, dev io.Writer) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout * time.Second)
    defer cancel()

    switch fn {
        case "setup":
            return setup(args, dev)
        case "limit":
            return limit(ctx, args, dev)
        default:
            return fmt.Errorf("server.Run(%s) not found", fn)
    }
}

