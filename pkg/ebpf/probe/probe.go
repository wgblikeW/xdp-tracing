// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package probe

import (
	"fmt"

	"github.com/aquasecurity/libbpfgo"
	"github.com/sirupsen/logrus"
)

// BPFProgram is an interface that used in BPFManager for maintaing different BPFProgram type.
// It should provide these method for BPFManager, which can help to manage their lifecycle.
// attach will attach specific BPFProgram to its hook point
// detach will detach specific BPFProgram from its hook point
type BPFProgram interface {
	Attach(*libbpfgo.Module) error
	Detach(*libbpfgo.Module) error
	GetName() string
	GetHookPoint() string
}

var allBPFProgs []BPFProgram = []BPFProgram{
	// &xdpProgram{&genericMeta{programName: "__test_trace_xdp", bpfLink: nil, hookPoint: "xdp"}, "ens33", 1},
	&tracepointProgram{&genericMeta{programName: "tracepoint__syscalls__sys_enter_execve", bpfLink: nil, hookPoint: "syscalls"}, "sys_enter_execve"},
	&tracepointProgram{&genericMeta{programName: "tracepoint__syscalls__sys_exit_execve", bpfLink: nil, hookPoint: "syscalls"}, "sys_exit_execve"},
	&xdpProgram{&genericMeta{programName: "xdp_proxy", bpfLink: nil, hookPoint: "xdp"}, "ens33", 0},
	&xdpProgram{&genericMeta{programName: "__test_trace_xdp", bpfLink: nil, hookPoint: "xdp"}, "ens33", 0},
}

func GetAllBPFProgs() []BPFProgram {
	return allBPFProgs
}

type genericMeta struct {
	programName string
	bpfLink     *libbpfgo.BPFLink
	hookPoint   string
}

type xdpProgram struct {
	*genericMeta
	targetDevice string
	attachMode   uint32
}

func (prog *xdpProgram) GetHookPoint() string {
	return "xdp"
}

func (prog *xdpProgram) GetName() string {
	return prog.programName
}

func (prog *xdpProgram) Attach(bpfModule *libbpfgo.Module) error {
	if prog.bpfLink != nil {
		logrus.WithFields(logrus.Fields{
			"location":  "(*xdpProgram) attach",
			"interface": prog.targetDevice,
		}).Warning("the xdpProgram has already been attached")
		return nil
	}

	if bpfModule == nil {
		logrus.WithFields(logrus.Fields{
			"location":  "(*xdpProgram) attach",
			"interface": prog.targetDevice,
		}).Warning("bpfModule has not been initialized properly yet")
		return fmt.Errorf("bpfModule has not been initialized properly yet")
	}

	bpfProg, err := bpfModule.GetProgram(string(prog.programName))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":       err,
			"location":  "(*xdpProgram) attach",
			"interface": prog.targetDevice,
		}).Warning("cannot get specific eBPF program from name=%v", prog.programName)
		return err
	}

	prog.bpfLink, err = bpfProg.AttachXDP(prog.targetDevice)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":       err,
			"location":  "(*xdpProgram) attach",
			"interface": prog.targetDevice,
		}).Warning("error occurs when attach target BPF Program to given interface")
		return err
	}

	return nil
}

func (prog *xdpProgram) Detach(bpfModule *libbpfgo.Module) error {
	if prog.bpfLink == nil {
		// BPFProgram has already been detached, it's okay that method was called more than once.
		logrus.WithFields(logrus.Fields{
			"location":  "(*xdpProgram) detach",
			"interface": prog.targetDevice,
		}).Warningf("bpfProgram %v has already been detached, but method get called", prog.programName)
		return nil
	}

	if bpfModule == nil {
		logrus.WithFields(logrus.Fields{
			"location":  "(*xdpProgram) detach",
			"interface": prog.targetDevice,
		}).Warning("bpfModule has not been initialized properly yet")
		return fmt.Errorf("bpfModule has not been initialized properly yet")
	}

	err := prog.bpfLink.Destroy()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"location":  "(*xdpProgram) detach",
			"err":       err,
			"interface": prog.targetDevice,
		}).Warningf("error occurs when try to detach xdpProgram %v from %v interface", prog.programName, prog.targetDevice)
		return err
	}

	prog.bpfLink = nil

	return nil
}

type tracepointProgram struct {
	*genericMeta
	tracePoint string
}

func (prog *tracepointProgram) GetHookPoint() string {
	return prog.tracePoint
}

func (prog *tracepointProgram) GetName() string {
	return prog.programName
}

func (prog *tracepointProgram) Attach(bpfModule *libbpfgo.Module) error {
	if prog.bpfLink != nil {
		logrus.WithFields(logrus.Fields{
			"location":   "(*tracepointProgram) attach",
			"tracepoint": prog.tracePoint,
		}).Warning("the xdpProgram has already been attached")
		return nil
	}

	if bpfModule == nil {
		logrus.WithFields(logrus.Fields{
			"location":   "(*tracepointProgram) attach",
			"tracepoint": prog.tracePoint,
		}).Warning("bpfModule has not been initialized properly yet")
		return fmt.Errorf("bpfModule has not been initialized properly yet")
	}

	bpfProg, err := bpfModule.GetProgram(prog.programName)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"location":   "(*tracepointProgram) attach",
			"tracepoint": prog.tracePoint,
		}).Warning("cannot get specific eBPF program from name=%v", prog.programName)
		return err
	}

	bpfLink, err := bpfProg.AttachTracepoint(prog.hookPoint, prog.tracePoint)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":        err,
			"location":   "(*tracepointProgram) attach",
			"tracepoint": prog.tracePoint,
		}).Warning("error occurs when attach target BPF Program to given interface")
		return err
	}

	prog.bpfLink = bpfLink

	logrus.Infof("Attach %v to %v", prog.programName, prog.hookPoint+"/"+prog.tracePoint)
	return nil
}

func (prog *tracepointProgram) Detach(bpfModule *libbpfgo.Module) error {
	if prog.bpfLink == nil {
		// BPFProgram has already been detached, it's okay that method was called more than once.
		logrus.WithFields(logrus.Fields{
			"location":   "(*tracepointProgram) detach",
			"tracepoint": prog.tracePoint,
		}).Warningf("bpfProgram %v has already been detached, but method get called", prog.programName)
		return nil
	}

	if bpfModule == nil {
		logrus.WithFields(logrus.Fields{
			"location":   "(*tracepointProgram) detach",
			"tracepoint": prog.tracePoint,
		}).Warning("bpfModule has not been initialized properly yet")
		return fmt.Errorf("bpfModule has not been initialized properly yet")
	}

	err := prog.bpfLink.Destroy()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"location":   "(*tracepointProgram) detach",
			"err":        err,
			"tracepoint": prog.tracePoint,
		}).Warningf("error occurs when try to detach tracepointProgram %v from %v", prog.programName, prog.tracePoint)
		return err
	}

	prog.bpfLink = nil

	return nil
}
