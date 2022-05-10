package bpf

/*
#include "headers/load-bpf.h"
#include <net/if.h>
*/
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/p1nant0m/xdp-tracing/config"
)

func Warp_if_nametoindex(devName string) int {
	return (int)(C.if_nametoindex(convertToCType(devName)[0].(*C.char)))
}

func Attach_bpf_prog_to_if(cfg *config.XDPConfig) int {
	C_cfg := newC_XDPConfig(cfg)
	return (int)(C.attach_bpf_prog_to_if(*C_cfg))
}

func convertToCType(goResp interface{}) []interface{} {
	vG := reflect.ValueOf(goResp)
	ret := make([]interface{}, 0)

	if vG.Type().Kind() == reflect.Struct {
		for i := 0; i < vG.NumField(); i++ {
			if vG.Field(i).CanInterface() { // Filed that can be exported
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
	case reflect.Uint32:
		ret = append(ret, (C.uint)(vG.Interface().(uint32)))
	}
	return ret
}

func newC_XDPConfig(cfg *config.XDPConfig) *C.struct_input_args {
	ret := convertToCType(*cfg)
	return &C.struct_input_args{ret[0].(C.uint), ret[1].(*C.char), ret[2].(*C.char)}
}
