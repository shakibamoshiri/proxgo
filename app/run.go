package app

import (
    "io"
    "log"

    "github.com/shakibamoshiri/proxgo/config"
    "github.com/shakibamoshiri/proxgo/server"
    "github.com/shakibamoshiri/proxgo/user"
)


func run(args []string, pc *config.Pools, dev io.Writer) (err error) {
    config.Log.Debug("args", "=", args)

    var nextArgs = make([]string, 0, 0)
    log.Print("server check ... ")
    err = server.Run("check", nextArgs, pc, io.Discard)
    //fmt.Println(err)
    if err != nil {
        return err
    }
    //fmt.Println("done")

    log.Printf("server fetch ... ")
    err = server.Run("fetch", nextArgs, pc, io.Discard)
    if err != nil {
        return err
    }
    //fmt.Println("done")

    log.Print("user setup ... ")
    err = user.Run("setup", nil, io.Discard)
    if err != nil {
        return err
    }
    //fmt.Println("done")

    log.Print("user limit ... ")
    err = user.Run("limit", nil, io.Discard)
    if err != nil {
        return err
    }
    //fmt.Println("done")

    return
}

// var timeMatch = regexp.MustCompile(`(\d+)([smhd])`)
// func parsePeriodCustom(format *string) time.Duration {
//     *format = strings.ToLower(*format)
//     matchList := timeMatch.FindStringSubmatch(*format)
//     config.Log.Info("matchList", "=", matchList)
//     if len(matchList) == 3 {
//         value, _ := strconv.ParseInt(matchList[1], 10, 64)
//         unit := matchList[2]
//         switch unit {
//             case "s":
//                 return value
//             case "m":
//                 return (value * 60)
//             case "h":
//                 return (value * 60 * 60)
//             case "d":
//                 return (value * 60 * 60 * 24)
//             default:
//                 return value
//         }
//     }
// 
//     return 0
// }
// 
// 
