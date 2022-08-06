// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package errors

import "fmt"

type CustomError interface {
	GetErrorString(int) error
}

const (
	// Error Code defines in xdp-proxy.c
	OK = iota
	ERROR_DETACH_PROG_FROM_INTERF
	ERROR_BPF_GET_LINK_XDP_ID
	ERROR_NOT_FOUND_BPF_PROG_ON_INTERF
	ERROR_NOT_EXPECTED_BPF_PROG_FOUND
	ERROR_ATTACH_PROG_TO_INTERF
	ERROR_NOT_FOUND_INTERFACE
	ERROR_OPENING_BPF_OBJECT
	ERROR_LOADING_BPF_OBJECT
	ERROR_SET_RLIMIT_MEMLOCK
	ERROR_BPF_FILE_OPEN
	ERROR_BPF_LOADING_TO_KERN
)

// private variable for mapping int error code to error string
var errorIntTOString map[int]error

func init() {
	errorIntTOString = map[int]error{
		OK:                                 fmt.Errorf("OK"),
		ERROR_DETACH_PROG_FROM_INTERF:      fmt.Errorf("error in detaching program from interface"),
		ERROR_BPF_GET_LINK_XDP_ID:          fmt.Errorf("error in getting link XDP index"),
		ERROR_NOT_FOUND_BPF_PROG_ON_INTERF: fmt.Errorf("error when trying to found BPF program on interface"),
		ERROR_NOT_EXPECTED_BPF_PROG_FOUND:  fmt.Errorf("error not find expected BPF program"),
		ERROR_ATTACH_PROG_TO_INTERF:        fmt.Errorf("failed to attach program to given interface"),
		ERROR_NOT_FOUND_INTERFACE:          fmt.Errorf("failed to find given interface"),
		ERROR_OPENING_BPF_OBJECT:           fmt.Errorf("failed to open and/or load BPF object"),
		ERROR_LOADING_BPF_OBJECT:           fmt.Errorf("failed to load BPF object"),
		ERROR_SET_RLIMIT_MEMLOCK:           fmt.Errorf("setrlimit(RLIMIT_MEMLOCK)"),
		ERROR_BPF_FILE_OPEN:                fmt.Errorf("error when reading eBPF binary program into memory"),
		ERROR_BPF_LOADING_TO_KERN:          fmt.Errorf("error when loading eBPF binary program into kernel"),
	}
}

func GetErrorString(errorCode int) error {
	return errorIntTOString[errorCode]
}
