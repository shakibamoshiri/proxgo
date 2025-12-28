package tell

import (
    //"fmt"

    "github.com/shakibamoshiri/proxgo/config"
)


func SendDoc(args []string) (err error) {
    config.Log.Debug("args", "=", args)
    config.Log.Debug("AgentID", "=", config.AgentID)

    return
}
