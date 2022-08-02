package ebpf

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/p1nant0m/xdp-tracing/pkg/ebpf/probe"
)

var bpfObjPath string = "../../bpf/output/test.bpf.o"

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

func TestAttachAll(t *testing.T) {
	m, err := NewBPFManager(WithBPFModuleFromFile(bpfObjPath), WithBPFProgramList(probe.GetAllBPFProgs()))
	if err != nil || m == nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = m.AttachAll()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	eventCh := make(chan []byte)
	rb, err := m.bpfModule.InitRingBuf("events", eventCh)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	rb.Start()
	defer rb.Stop()

	go func() {
		_, err := exec.Command("ping", "192.168.176.128", "-c", "20").Output()
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
