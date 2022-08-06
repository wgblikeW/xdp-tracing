// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ebpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/p1nant0m/xdp-tracing/pkg/ebpf/probe"
	"github.com/sirupsen/logrus"
)

var bpfObjPath string = "../../bpf/output/core.bpf.o"
var bpfManager *BPFManager
var once1 sync.Once

func getSameManager() (*BPFManager, error) {
	var err error
	once1.Do(func() {
		bpfManager, err = NewBPFManager(WithBPFModuleFromFile(bpfObjPath), WithBPFProgramList(probe.GetAllBPFProgs()))
	})

	if err != nil {
		return nil, err
	}

	return bpfManager, nil
}

type eventS struct {
	Comm       string `size:"16"`
	Pid        int
	Tgid       int
	Ppid       int
	Uid        uint
	Retval     int
	Args_count int
	Args_size  uint
	Args       string `size:"7680"`
}

func findTruncate(buf []byte) int {
	bufLen := len(buf)
	for i := bufLen - 1; i > 0; i-- {
		if buf[i] != 0 {
			return i + 1
		}
	}

	return bufLen - 1
}

func MapToStruct(buf []byte, data interface{}) error {
	off := 0
	bufSize := len(buf)

	for i := 0; i < reflect.ValueOf(data).Elem().NumField(); i++ {
		v := reflect.ValueOf(data).Elem().Field(i)
		t := reflect.TypeOf(data).Elem().Field(i)
		switch v.Kind() {
		case reflect.String:
			sizeS, ok := t.Tag.Lookup("size")
			if !ok {
				return errors.New("error when parse tag")
			}

			sizeI, _ := strconv.Atoi(sizeS)
			trunidx := off
			if off+sizeI > bufSize {
				trunidx += findTruncate(buf[off:])
			} else {
				trunidx += findTruncate(buf[off : off+sizeI])
			}

			v.SetString(string(buf[off:trunidx]))
			off += sizeI
		case reflect.Uint:
			v.SetUint(uint64(binary.LittleEndian.Uint32(buf[off : off+4])))
			off += 4
		case reflect.Int:
			v.SetInt(int64(binary.LittleEndian.Uint32(buf[off : off+4])))
			off += 4
		}
	}
	return nil
}

func TestNewBPFManagerWithInvaildBPFObjPath(t *testing.T) {
	invaildPath := "foo.bpf.o"
	_, err := NewBPFManager(WithBPFModuleFromFile(invaildPath))
	if err == nil {
		t.Fatalf("Expected an error when passing invalidPath, got %v", err)
	}
}

func TestLoadValidbpfObj(t *testing.T) {
	m, err := NewBPFManager(WithBPFModuleFromFile(bpfObjPath))
	if err != nil || m == nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestTracingProgram(t *testing.T) {
	m, err := getSameManager()
	if err != nil || m == nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = m.AttachGiven("tracepoint__syscalls__sys_enter_execve", "tracepoint__syscalls__sys_exit_execve")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventsCh := make(chan []byte)
	lostCh := make(chan uint64)
	pb, err := m.bpfModule.InitPerfBuf("events", eventsCh, lostCh, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	pb.Start()
	defer pb.Stop()

	eventT := &eventS{}

	for event := range eventsCh {
		err = MapToStruct(event, eventT)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		logrus.Infof(
			"Comm:%v pid:%v tgid:%v ppid:%v uid:%v retval:%v args:%v \n\n",
			eventT.Comm,
			eventT.Pid,
			eventT.Tgid,
			eventT.Ppid,
			eventT.Uid,
			eventT.Retval,
			eventT.Args,
		)
		break
	}
}

func TestAttachXDP(t *testing.T) {
	m, err := getSameManager()
	if err != nil || m == nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = m.AttachGiven("__test_trace_xdp")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventCh := make(chan []byte)
	rb, err := m.bpfModule.InitRingBuf("perfs", eventCh)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	rb.Start()
	defer rb.Stop()

	go func() {
		_, err := exec.Command("ping", "192.168.176.128", "-c", "30").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error when execute ping command")
			os.Exit(1)
		}
	}()

	count := 0
	timeout := time.Tick(time.Second * 20)
	for {
		select {
		case event := <-eventCh:
			if got := binary.LittleEndian.Uint32(event); got != 7070 {
				t.Errorf("Expeted 7070 from eventCh, got %v", got)
			}
			if count = count + 1; count > 15 {
				goto out
			}
		case <-timeout:
			t.Errorf("Timeout happends before getting enouth counts 15, just got %v", count)
		}
	}
out:
	return
}
