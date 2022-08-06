// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

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
	"github.com/sirupsen/logrus"
)

func MapRevoke(srcIP uint32, id uint32) {
	errcode, err := C.bpf_revoke_map(convertToCType(srcIP)[0].(C.uint), convertToCType(id)[0].(C.uint))
	if err != nil {
		logrus.Warnf("[bpf] errors occurs when doing mapRevoke err=%v errcode=%v", err, errcode)
	} else {
		logrus.Warnf("[bpf] successfully revoke map elem %v", srcIP)
	}
}

func MapUpdate(srcIP uint32, id uint32) {
	errcode, err := C.bpf_update_map(convertToCType(srcIP)[0].(C.uint), convertToCType(id)[0].(C.uint))
	if err != nil {
		logrus.Warnf("[bpf] errors occurs when doing mapUpdate err=%v errcode=%v", err, errcode)
	} else {
		logrus.Warnf("[bpf] successfully update map elem %v", srcIP)
	}
}

func Warp_do_detach(ifname string, prog_id int) int {
	ifIdx := Warp_if_nametoindex(ifname)
	C_type_ifIdx := convertToCType(ifIdx)[0].(C.int)
	C_type_prog_id := convertToCType(prog_id)[0].(C.int)

	return (int)(C.do_detach(C_type_ifIdx, C_type_prog_id))
}

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
	case reflect.Int:
		ret = append(ret, (C.int)(vG.Interface().(int)))
	}
	return ret
}

func newC_XDPConfig(cfg *config.XDPConfig) *C.struct_input_args {
	ret := convertToCType(*cfg)
	return &C.struct_input_args{ret[0].(C.uint), ret[1].(*C.char), ret[2].(*C.char)}
}
