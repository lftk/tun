package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var Version = "0.1.0"

var (
	minClientVersion = "0.1.0"
	minServerVersion = "0.1.0"
)

func CompatClient(ver string) (err error) {
	b, err := compat(ver, minClientVersion)
	if err != nil {
		return
	}

	if !b {
		err = fmt.Errorf("Version is too low, at least %s", minClientVersion)
	}
	return
}

func CompatServer(ver string) (err error) {
	b, err := compat(ver, minServerVersion)
	if err != nil {
		return
	}

	if !b {
		err = fmt.Errorf("Version is too high, server is %s", ver)
	}
	return
}

func compat(ver, min string) (b bool, err error) {
	n, err := compare(ver, min)
	if err != nil {
		return
	}

	b = (n >= 0)
	return
}

func toInts(ver string) (vv []int, err error) {
	ss := strings.Split(ver, ".")
	if len(ss) != 3 {
		err = errors.New("Invalid version")
		return
	}

	vv = make([]int, len(ss))
	for i, s := range ss {
		vv[i], err = strconv.Atoi(s)
		if err != nil {
			return
		}
	}
	return
}

func compare(v1, v2 string) (n int, err error) {
	vv1, err := toInts(v1)
	if err != nil {
		return
	}

	vv2, err := toInts(v2)
	if err != nil {
		return
	}

	for i, n1 := range vv1 {
		if n2 := vv2[i]; n1 < n2 {
			return -1, nil
		} else if n1 > n2 {
			return 1, nil
		}
	}
	return
}
