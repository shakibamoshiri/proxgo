package app

import (
    "fmt"
    "io"
    "time"
    "flag"
    //"strings"
    //"strconv"
    //"regexp"
    // "context"
    "log"
    "os"
    "syscall"
    "os/signal"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/user"
)

func service(args []string, pc *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)
    config.Log.Debug("io.Writer", "dev", dev)


    flags := flag.NewFlagSet("app_run", flag.ExitOnError)
    var duration string
    var uDash bool
    var runCount int
    flags.StringVar(&duration, "tick", "", "custom time to tick <NUMBER>[s|m|h]")
    flags.BoolVar(&uDash, "udash", false, "enable/run user dash")
    flags.IntVar(&runCount, "runc", 0, "run counter <NUMBER> (stop after N count)")
    flags.Parse(args)

    if duration == "" {
        flags.PrintDefaults()
        return
    }
    
    customTick, err := time.ParseDuration(duration)
    if err != nil {
        return fmt.Errorf("service / time.ParseDuration(%v) %w", duration, err)
    }
    log.Printf("service is polling every %+v\n", customTick)

    if uDash {
        nextArgs := []string{"user", "dash"}
        go user.Parse(nextArgs)
    }

    if runCount < 0 {
        log.Printf("run count cannot be negative %d\n", runCount)
        os.Exit(0)
    }

    // ctx, cancel := context.WithCancel(context.Background())
    // defer cancel()

    // Handle OS osSignals
    // go func() {
    //     osSig := make(chan os.Signal, 1)
    //     <-osSig
    //     signal.Notify(osSig, os.Interrupt, os.Kill, syscall.SIGTERM)
    //     cancel()
    // }()

    osSig := make(chan os.Signal, 1)
    signal.Notify(osSig, os.Interrupt, os.Kill, syscall.SIGTERM)
    defer signal.Stop(osSig)

    ticker := time.NewTicker(customTick)
    defer ticker.Stop()

    for {
        err = run(args, pc, dev)
        if err != nil {
            return fmt.Errorf("service / for-run() failure %w", err)
        }
        if runCount >= 0 {
            runCount--
            if runCount == 0 {
                config.Log.Warn("run count is zero", "=", runCount)
                log.Printf("run count is zero %d\n", runCount)
                return
            }
        }

        select {
        case <-ticker.C:
            // continue
        case sig := <-osSig:
            config.Log.Warn("service / for-run() signal received", "=", sig)
            return fmt.Errorf("service / for-run() signal received  %v", sig)
        }
    }
    return
}

