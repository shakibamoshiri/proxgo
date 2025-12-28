package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
    "bytes"

	"github.com/shakibamoshiri/proxgo/config"
)

func icheck() (err error) {

	type ServerUser struct {
		Username string   `json:"username"`
		Extra    struct{} `json:"-"`
	}

	type ServerUsersResponse struct {
		Users []ServerUser `json:"users"`
	}

	var result ServerUsersResponse
	var users [3]int

	client := &http.Client{
		Timeout: (time.Second * config.ClientTimeout),
	}

	ob := config.NewOutputBuffer()

	var ssmApiAddr string
    var body bytes.Buffer
    var userLen int
	for i, server := range yaml.Pools.Servers {
		ssmApiAddr = server.Addr("stats")

		config.Log.Info("ssmApiAddr", "=", ssmApiAddr)
		ob.Printf("%-30s", "data.icheck."+server.Location)
		resp, err := client.Get(ssmApiAddr)
		if err != nil {
			ob.Println(err)
			ob.Errorln(err)
			err = nil
			continue
		}
		defer func() {
			errClose := resp.Body.Close()
			if errClose != nil {
				err = errClose
			}
		}()

        io.Copy(&body, resp.Body)
		if resp.StatusCode != 200 {
			err := fmt.Errorf("bad status: %d %s\nResponse: %s", resp.StatusCode, resp.Status, body.String())
			return err
		}
		config.Log.Debug("body", "string(body)", body.String())

        // direct decoding
        // err = json.NewDecoder(resp.Body).Decode(&result)

        err = json.NewDecoder(&body).Decode(&result)
		if err != nil {
			return err
		}

		// err = json.Unmarshal(body, &result)
		// if err != nil {
		// 	return err
		// }

		userLen = len(result.Users)
		users[i] = userLen
        if userLen == 0 {
            ob.Println("users:empty")
        } else {
            ob.Printf("users:%d\n", userLen)
        }
	}

	if ob.Stderr.Len() > 0 {
		err := fmt.Errorf("%s server connection failed, no full data to check!", "icheck")
		ob.Stderr.Reset()
		ob.Printf("\n")
		ob.Flush()
		return err
	}

	min, max := MinMax(users[:])
	if min != max {
		err := fmt.Errorf("%s integrity check failed, servers out of sync!", "icheck")
		ob.Stderr.Reset()
		ob.Flush()
		return err
	}

	ob.Flush()
	return nil
}

func MinMax(nums []int) (min, max int) {
	if len(nums) == 0 {
		return
	}
	min, max = nums[0], nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return
}
