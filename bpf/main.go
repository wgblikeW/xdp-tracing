package main

/*
#include "headers/load-bpf.h"
// #include <net/if.h>
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/p1nant0m/xdp-tracing/bpf/errors"
	"github.com/p1nant0m/xdp-tracing/config"
)

// func detach() {
// 	DEV_NAME := []byte("ens33")
// 	ifidx := int(C.if_nametoindex((*C.char)(unsafe.Pointer(&DEV_NAME[0]))))
// 	fmt.Print(errors.GetErrorString(int(C.do_detach((C.int)(ifidx), 53))), "\n")
// }

func convertToC_CharPtr(str string) *C.char {
	return (*C.char)(unsafe.Pointer(&[]byte(str)[0]))
}

func convertToCType(goResp interface{}) []interface{} {
	vG := reflect.ValueOf(goResp)
	ret := make([]interface{}, 0)

	if vG.Type().Kind() == reflect.Struct {
		for i := 0; i < vG.NumField(); i++ {
			if vG.Field(i).CanInterface() {
				switch vG.Field(i).Type().Kind() {
				case reflect.String:
					ret = append(ret, (*C.char)(unsafe.Pointer(&[]byte(vG.Field(i).Interface().(string))[0])))
				case reflect.Uint32:
					ret = append(ret, (C.uint)(vG.Field(i).Interface().(uint32)))
				}
			}
		}
	}

	switch vG.Type().Kind() {
	case reflect.String:
		ret = append(ret, (*C.char)(unsafe.Pointer(&[]byte(vG.Interface().(string))[0])))
	}
	return ret
}

type C__XDP_Config struct {
	Xdp_flags C.uint
	Filename  *C.char
	if_name   *C.char
}

func main() {
	// ret := convertToCType(if_name)
	cfg := config.NewXDPConfig()
	// reconfigure the XDP program
	cfg.Ifname = "ens33"
	ret := convertToCType(*cfg)
	C_cfg := C.struct_input_args{ret[0].(C.uint), ret[1].(*C.char), ret[2].(*C.char)}
	fmt.Print(errors.GetErrorString(int(
		C.attach_bpf_prog_to_if(C_cfg))))
}
