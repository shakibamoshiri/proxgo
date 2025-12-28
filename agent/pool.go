package agent

import (
    "fmt"
    "flag"

    "github.com/shakibamoshiri/proxgo/config"
)



func pool(args []string) (err error) {
    config.Log.Debug("passed arguments", "[]string", args)

    poolFlag := flag.NewFlagSet("pool", flag.ExitOnError)
    helpFlag := poolFlag.Bool("h", false, "show help")
    newPoolID := poolFlag.Int("id", 0, "set new pool id, OPTIONAL")
    poolFlag.Parse(args)

    if *helpFlag {
        poolFlag.PrintDefaults()
        return nil
    }

    const maxPoolID = 0xFF
    oldPoolId := yaml.Agents.Agent.PoolID
    if *newPoolID > 0 && *newPoolID <= maxPoolID {
        yaml.Agents.Agent.PoolID = *newPoolID
    }
    config.Log.Info("parse new pool id", "-id", *newPoolID)

    var newPoolIDFound bool = false
    fmt.Printf("%-5s %-10s %-15s %-10s %-10s %s\n", "ID", "Period", "Traffic", "Sessions", "Capacity", "servers")
    for _, pool := range yaml.Agents.Pools {
        fmt.Printf("%-5d %-10d %-15d %-10d %-10d %d\n",
            pool.ID, 
            pool.Period,
            pool.Traffic,
            pool.Sessions,
            pool.Capacity,
            pool.Servers)
        if pool.ID == *newPoolID {
            newPoolIDFound = true
        }
    }

    fmt.Printf("\nActive Pool ID %#v\n", oldPoolId)

    config.Log.Debug("check agent yaml file", "yaml.Agents", yaml.Agents)

    // if -id was not set, nothing to do
    if *newPoolID == 0 {
        return nil
    }

    if newPoolIDFound == false {
        fmt.Printf("New Pool ID %#v not found!\n", *newPoolID)
        return nil
    }

    if *newPoolID != oldPoolId {
        err = yaml.SaveAgent2()
        if err != nil {
            return err
        }
        fmt.Printf("New Pool ID %#v was set\n", *newPoolID)
    }

    return nil
}
