package agent

import (
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)



func server(args []string) (err error) {
    config.Log.Debug("passed arguments", "[]string", args)

    fmt.Printf("%-5s %-10s %-15s %-5s %s\n", "ID", "Location", "Address", "Port", "Active")
    for _, server := range yaml.Pools.Servers {
        fmt.Printf("%-5d %-10s %-15s %-5d %t\n", server.ID, server.Location, server.APIAddr, server.APIPort, server.Active)
    }

    return nil
}
