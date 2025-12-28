package server

import (
    "fmt"
    "os"
    "net/http"
    "io"
    "time"
    "sync"
    "bytes"

    "github.com/shakibamoshiri/proxgo/config"
)

func check(args []string, pc *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)
    config.Log.Debug("io.Writer", "dev", dev)

    ob := config.NewOutputBuffer()

    client := &http.Client{
        Timeout: (time.Second * config.ClientTimeout),
    }

    type Result struct {
        Resp *http.Response
        Err  error
        Loc string
        Line int
    }

    // One channel that carries both resp and err together
    results := make(chan Result, len(pc.Servers)) // buffered = safe

    var wg sync.WaitGroup
    wg.Add(len(pc.Servers))

    var ssmApiAddr string
    for i, server := range pc.Servers {
        ssmApiAddr = server.Addr("")
        config.Log.Info("ssmApiAddr", "=", ssmApiAddr)

        go func(addr, loc string, line int) {
            defer wg.Done()
            resp, err := client.Get(addr)
            results <- Result{
                Resp: resp,
                Err: err,
                Loc: loc,
                Line: line,
            }
            // Always close resp.Body in real code!
            if resp != nil {
                defer resp.Body.Close()
            }
        }(ssmApiAddr, server.Location, i)
    }


    wg.Wait()
    close(results)

    /// for i, server := range pc.Servers {
    ///      fmt.Printf("check.%d.%-30s\n", i, server.Location)
    /// }
    /// fmt.Printf("\033[%dA", len(pc.Servers))

    /// var checked string
    /// for result := range results {
    ///     checked = "checked"
    ///     if result.Err != nil {
    ///         checked = fmt.Sprintf("%s\n", result.Err)
    ///     }
    ///     time.Sleep(time.Millisecond * 1000)
    ///     fmt.Printf("\033[s")
    ///     fmt.Printf("\033[%dB", result.Line)
    ///     fmt.Printf("\033[%dC %d.%s\n", 30, result.Line, checked)
    ///     fmt.Printf("\033[u")

    /// }
    /// fmt.Printf("\033[%dB", len(pc.Servers))


    // read all results
    for result := range results {
        ob.Fprintf(dev, "%-30s", "server.check." + result.Loc)
        if result.Err != nil {
            ob.Fprintln(dev, result.Err)
            ob.Fprintln(os.Stderr, result.Err)
            continue
        }
        if result.Resp.StatusCode != 200 {
            var body bytes.Buffer
            io.Copy(&body, result.Resp.Body)
            err := fmt.Errorf("bad status: %s, Response: %s", result.Resp.Status, body.String())
            ob.Fprintln(dev, err)
            ob.Fprintln(os.Stderr, err)
            continue
        }
        config.Log.Info("response status code", result.Loc, result.Resp.StatusCode)
        ob.Fprintln(dev, "checked")
    }

    if ob.Stderr.Len() > 0 {
        err := fmt.Errorf("[%d] %s", ob.ErrCount, "server check failed, all servers should be up!")
        ob.Stderr.Reset()
        ob.Fprintln(dev)
        ob.Flush()
        return err
    }

    ob.Flush()

    return nil
}
