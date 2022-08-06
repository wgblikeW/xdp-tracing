// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

/*
Pacakge ebpf aims to help developers to create, load, and attach different eBPF program to
specific target quickly and elegantly. What's more, it's important to manage the
ebpf program life cycle in a proper approach.

The package will use libbpfgo to implement high level abstraction that required in
building a eBPF program.
*/
package ebpf

import (
	"context"
	"fmt"
	"sync"

	"github.com/aquasecurity/libbpfgo"
	"github.com/p1nant0m/xdp-tracing/pkg/ebpf/probe"
	"github.com/sirupsen/logrus"
)

const ()

var once sync.Once

// Option defines optional parameters for initializing the BPFManager struct,
// and it will return an error when something goes wrong in initializing.
type Option func(*BPFManager) error

// The BPFManager manages a single BPFModule, and BPFModule is the program collection of
//  a specific eBPF type. Or in some not very strict scenerio, we can put different types of
// eBPF programs in a single *.bpf.c file
type BPFManager struct {
	programMaps map[string]probe.BPFProgram
	bpfModule   *libbpfgo.Module
	bpfPrograms []probe.BPFProgram
	ctx         context.Context
}

// WithContext allows passing context parameter to BPFManager in order
// to keep on the whole program state.
func WithContext(ctx context.Context) Option {
	return func(b *BPFManager) error {
		b.ctx = ctx
		return nil
	}
}

// WithBPFModuleFromFile is used to instantiate the BPFModule from
// given bpfObjPath. It will return an error when get not get BPFModule from BPFObjFile.
func WithBPFModuleFromFile(bpfObjPath string) Option {
	return func(b *BPFManager) error {
		bpfModule, err := libbpfgo.NewModuleFromFile(bpfObjPath)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"bpfObjPath": bpfObjPath,
			}).Warning("error occurs when create bpfModule from file")
			return err
		}
		b.bpfModule = bpfModule

		return nil
	}
}

// WithBPFModuleFromBuffer is used to instantiate the BPFModule from bytes in
// buffer and set the bpfObjName.
func WithBPFModuleFromBuffer(bpfObjBuff []byte, bpfObjName string) Option {
	return func(b *BPFManager) error {
		bpfModule, err := libbpfgo.NewModuleFromBuffer(bpfObjBuff, bpfObjName)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":        err,
				"bpfObjBuff": bpfObjBuff,
				"bpfObjName": bpfObjName,
			}).Warning("error occurs when create bpfModule from Buffer")
			return err
		}
		b.bpfModule = bpfModule

		return nil
	}
}

// WithBPFProgramList is used to set the parameter of BPFPrograms. It regists all
// BPFPrograms that will be attached to specific point as a probe and maintain some
// unique metadata for its bpf program type.
func WithBPFProgramList(programList []probe.BPFProgram) Option {
	return func(b *BPFManager) error {
		b.bpfPrograms = append(b.bpfPrograms, programList...)
		for _, prog := range programList {
			b.programMaps[prog.GetName()] = prog
		}
		return nil
	}
}

// NewBPFManager instantiates the BPFManager with give Options.
// It will return an error whenever an error occurs in initializing
// the parameters with give options.
func NewBPFManager(opts ...Option) (*BPFManager, error) {
	ins := &BPFManager{
		programMaps: map[string]probe.BPFProgram{},
	}

	for _, opt := range opts {
		if err := opt(ins); err != nil {
			return nil, err
		}
	}

	return ins, nil
}

func checkBPFObjLoadOr(bpfModule *libbpfgo.Module) error {
	var err error

	if bpfModule == nil {
		return fmt.Errorf("bpfModule has not been initialized properly yet")
	}

	once.Do(func() {
		err = bpfModule.BPFLoadObject()
	})
	if err != nil {
		return err
	}

	return nil
}

// AttachAll attach all registed BPFProgram in the manager.
func (manager *BPFManager) AttachAll() error {
	var err error

	err = checkBPFObjLoadOr(manager.bpfModule)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":      err,
			"location": "(*BPFManager) AttachAll",
		}).Warningf("validation of BPFObj fails")
		return err
	}

	for _, prog := range manager.bpfPrograms {
		// try to attach all registed BPFProgram to their Point
		err = prog.Attach(manager.bpfModule)
	}

	// maybe not immediately return when error occurs, that will be helpful
	// to find all possible mistakes in one shot. We just return an error while
	// there maybe more than one error occurs.
	if err != nil {
		return fmt.Errorf("error occurs when doing BPFProgram Attach")
	}

	return nil
}

// AttachGiven attach BPFProgram which was registed with given name.
// It will return the first error where it fails to attach eBPF program.
// and the rest of them will not try to attach.
func (manager *BPFManager) AttachGiven(progNames ...string) error {
	var err error

	err = checkBPFObjLoadOr(manager.bpfModule)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":      err,
			"location": "(*BPFManager) AttachGiven",
		}).Warningf("validation of BPFObj fails")
		return err
	}

	for _, progName := range progNames {
		if prog, exists := manager.programMaps[progName]; exists {
			err := prog.Attach(manager.bpfModule)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":       err,
					"location":  "(*BPFManager) AttachGiven",
					"progName":  progName,
					"hookpoint": prog.GetHookPoint(),
				}).Warningf("error occurs when attach bpfProgram to its Hook point")

				return err
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"location": "(*BPFManager) AttachGiven",
				"progName": progName,
			}).Warningf("there is no bpfProgram was registed with given progName")

			return fmt.Errorf("bpfProgram with name %v was not registed", progName)
		}
	}

	return nil
}

func (manager *BPFManager) DetachAll() error {
	var err error

	err = checkBPFObjLoadOr(manager.bpfModule)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":      err,
			"location": "(*BPFManager) DetachGiven",
		}).Warningf("validation of BPFObj fails")
		return err
	}

	for _, prog := range manager.bpfPrograms {
		err = prog.Detach(manager.bpfModule)
	}

	if err != nil {
		return fmt.Errorf("error occurs when doing BPFProgram Attach")
	}

	return nil
}

func (manager *BPFManager) DetachGiven(progName string) error {
	var err error

	err = checkBPFObjLoadOr(manager.bpfModule)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":      err,
			"location": "(*BPFManager) DetachGiven",
		}).Warningf("validation of BPFObj fails")
		return err
	}

	if prog, exists := manager.programMaps[progName]; exists {
		err := prog.Detach(manager.bpfModule)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":       err,
				"location":  "(*BPFManager) DetachGiven",
				"progName":  progName,
				"hookpoint": prog.GetHookPoint(),
			}).Warningf("error occurs when detach bpfProgram from its Hook point")

			return err
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"location": "(*BPFManager) DetachGiven",
			"progName": progName,
		}).Warningf("there is no bpfProgram was registed with given progName")

		return fmt.Errorf("bpfProgram with name %v was not registed", progName)
	}

	return nil
}
