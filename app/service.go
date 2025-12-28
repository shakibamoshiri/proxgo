package app

import (
    "fmt"
    "io"
    "time"
    "flag"
    //"strings"
    //"strconv"
    //"regexp"
    "context"
    "log"
    "os"
    "syscall"
    "os/signal"

    "github.com/shakibamoshiri/proxgo/config"
)


func service(args []string, pc *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)
    config.Log.Debug("io.Writer", "dev", dev)

    flags := flag.NewFlagSet("app_run", flag.ExitOnError)
    duration := flags.String("tick", "", "custom time to tick <NUMBER>[s|m|h]")
    flags.Parse(args)

    if *duration == "" {
        flags.PrintDefaults()
        return
    }
    
    customTick, err := time.ParseDuration(*duration)
    if err != nil {
        return fmt.Errorf("service / time.ParseDuration(%v) %w", *duration, err)
    }
    fmt.Printf("ticking every %+v\n", customTick)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle OS signals
    go func() {
        sig := make(chan os.Signal, 1)
        signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
        <-sig
        cancel()
    }()

    ticker := time.NewTicker(customTick)
    defer ticker.Stop()

    for {
        err = run(args, pc, dev)
        if err != nil {
            ticker.Stop()
            cancel()
            return err
        }

        select {
        case <-ticker.C:
            // continue
        case <-ctx.Done():
            log.Println("Shutting down gracefully")
            return nil
        }
    }
}

