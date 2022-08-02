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
	&xdpProgram{&genericMeta{programName: "__test_trace_xdp", bpfLink: nil, hookPoint: "xdp"}, "ens33", 1},
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
