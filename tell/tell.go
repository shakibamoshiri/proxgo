package tell

import (
    "fmt"
    "sync"
    "context"
    "time"

    "github.com/shakibamoshiri/proxgo/config"
)

var yaml config.YamlFiles

var waitForTell sync.WaitGroup
const timeout time.Duration = 5

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
        help.For("tell").Exit(0)
    }

    nextCmd := args[1]
    nextArgs := args[2:]
    switch nextCmd {
        case "test":
            // err = sendTest()
            waitForTell.Add(1)
            go sendTest(1)
        case "msg":
            err = sendMsg("")
        case "doc":
            err = SendDoc(nextArgs)
        default:
            help.
            For("tell").
            Say("%s command not found", nextCmd).
            Exit(1)

    }

    waitForTell.Wait()
    if err != nil {
        return fmt.Errorf("tell >> %w", err)
    }
    return nil
}

func Run(fn string, msg string) (err error) {
    ctx, cancel := context.WithTimeout(context.Background(), timeout * time.Second)
    defer cancel()

    switch fn {
        case "sendMsg":
            return sendMsg(msg)
        case "sendMsgHasContext":
            return sendMsgContext(ctx, msg)
        default:
            return fmt.Errorf("tell.Run(%s) not found", fn)
    }
}

func Fire(fn string, ctx context.Context, msg string) (err error) {
    // defer wg.Done()
    switch fn {
        case "sendMsg":
            err = sendMsg(msg)
        case "sendMsgContext":
            err = sendMsgContext(ctx, msg)
        case "notify":
            err = sendMsgContext(ctx, msg)
        default:
            return fmt.Errorf("tell.Run(%s) not found", fn)
    }

    if err != nil {
        config.Log.Error("tell / Run(", fn, ") failed" )
        config.Log.Debug("tell / Run(", fn, ") failed", "error", err )
    }

    return err
}

func Fire_b(fn string, ctx context.Context, chanErr chan error,  msg string) (err error) {
    // defer wg.Done()
    switch fn {
        case "sendMsg":
            err = sendMsg(msg)
        case "sendMsgContext":
            err = sendMsgContext(ctx, msg)
        case "notify":
            err = sendMsgContext(ctx, msg)
        default:
            return fmt.Errorf("tell.Run(%s) not found", fn)
    }

    chanErr <- err
    close(chanErr)
    
    if err != nil {
        config.Log.Error("tell / Run(", fn, ") failed" )
        config.Log.Debug("tell / Run(", fn, ") failed", "error", err )
    }

    return err
}
