package agent

import (
    "fmt"

    "github.com/shakibamoshiri/proxgo/config"
)



func setup(args []string) (err error) {
    config.Log.Info("passed arguments", "[]string", args)
    fmt.Println("agent setup")
    fmt.Println("yaml files should be created")


    return nil
}
