package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// Version describes the current version.
	Version = "0.1.1"

	// The server can be compatible with the minimum version of the client.
	minClientVersion = "0.1.0"
	// The client can be compatible with the minimum version of the server.
	minServerVersion = "0.1.0"
)

// CompatClient is used by the server to determine whether the client is compatible.
func CompatClient(ver string) (err error) {
	b, err := compat(ver, minClientVersion)
	if err != nil {
		return
	}

	if !b {
		err = fmt.Errorf("version is too low, at least %s", minClientVersion)
	}
	return
}

// CompatServer is used by the client to determine whether the server is compatible.
func CompatServer(ver string) (err error) {
	b, err := compat(ver, minServerVersion)
	if err != nil {
		return
	}

	if !b {
		err = fmt.Errorf("version is too high, server is %s", ver)
	}
	return
}

// compat to determine whether two versions are compatible or not.
func compat(ver, min string) (b bool, err error) {
	n, err := compare(ver, min)
	if err != nil {
		return
	}

	b = (n >= 0)
	return
}

// toInts splits the version number of the string format into an int slice.
func toInts(ver string) (vv []int, err error) {
	ss := strings.Split(ver, ".")
	if len(ss) != 3 {
		err = errors.New("invalid version")
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

// compare two versions.
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
