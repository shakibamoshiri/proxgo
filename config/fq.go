package config

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FlagQuery struct {
	flags map[string]string
	flag  string
	value string
	ok    bool
	err   error
	help  []string
}

func (fq *FlagQuery) Take(osArgs []string, flags map[string]string) *FlagQuery {
	if fq.flags == nil {
		fq.flags = make(map[string]string, len(flags))
	}

	line := strings.Join(osArgs, " ")

	space := regexp.MustCompile(`\s+`)
	line = space.ReplaceAllString(line, " ")

	re := regexp.MustCompile(`(\w+(?:=)[^ ]+)`)
	args := re.FindAllString(line, -1) // -1 = find ALL matches

	for flag, _ := range flags {
		fq.help = append(fq.help, flag)
		for _, arg := range args {
			//fmt.Println("arg =", arg)
			if strings.HasPrefix(arg, flag) {
				value := strings.TrimPrefix(arg, flag)
				// value = strings.TrimPrefix(value, " ")
				value = strings.TrimPrefix(value, "=")
				// value = strings.TrimSuffix(value, " ")
				fq.flags[flag] = value
			}
		}
	}
	return fq
}

func (fq *FlagQuery) Help(flags map[string]string) *FlagQuery {
	sort.Strings(fq.help)
	for _, key := range fq.help {
		desc, ok := flags[key]
		if ok {
			fmt.Printf("%-20s %s\n", key, desc)
		} else {
			desc = "NOT FOUND!"
			fmt.Printf("%-20s %s\n", key, desc)
		}
	}
	return fq
}
func (fq *FlagQuery) Exit(code int) {
	os.Exit(code)
}

func (fq *FlagQuery) Find(match ...string) *FlagQuery {
	for _, f := range match {
		fq.flag = f
		fq.value, fq.ok = fq.flags[f]
		if fq.ok {
			break
		}
	}
	return fq
}

// //////////////////////////////////////////////////////////////////////////////
// Default and Assert are exclusive
// //////////////////////////////////////////////////////////////////////////////
func (fq *FlagQuery) Default(v string) *FlagQuery {
	fq.value, fq.ok = fq.flags[fq.flag]
	if !fq.ok {
		fq.flags[fq.flag] = v
		fq.value = v
		fq.ok = true
	}
	return fq
}

func (fq *FlagQuery) Assert() *FlagQuery {
	if fq.value == "" {
		fq.err = fmt.Errorf("FlagQuery() error %s should not be empty!", fq.flag)
	}
	if fq.ok == false {
		fq.err = fmt.Errorf("FlagQuery() error %s is required!", fq.flag)
	}
	return fq
}

// //////////////////////////////////////////////////////////////////////////////
// get String, Int, Bool
// //////////////////////////////////////////////////////////////////////////////
func (fq *FlagQuery) String() (string, error) {
	// fmt.Printf("FlagQuery String() fq = %#v\n", fq)
	delete(fq.flags, fq.flag)
	return fq.value, fq.err
}

func (fq *FlagQuery) Int() (int, error) {
	if fq.err != nil {
		return 0, fq.err
	}
	valueInt, err := strconv.ParseInt(fq.value, 10, 32)
	// fmt.Printf("FlagQuery Int() fq = %#v\n", fq)
	delete(fq.flags, fq.flag)
	return int(valueInt), err
}

func (fq *FlagQuery) Bool() (bool, error) {
	// fmt.Printf("FlagQuery Bool() fq = %#v\n", fq)
	delete(fq.flags, fq.flag)
	return fq.ok, fq.err
}
