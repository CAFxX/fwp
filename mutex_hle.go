//go:build amd64 && !race

package fwp

import (
	"sync"

	"github.com/intel-go/cpuid"
	"github.com/linux4life798/safetyfast"
)

var hle = cpuid.HasExtendedFeature(cpuid.HLE)

type mutex struct {
	m  sync.Mutex
	ms safetyfast.SpinHLEMutex
}

func (m *mutex) Lock() {
	if hle {
		m.ms.Lock()
	} else {
		m.m.Lock()
	}
}

func (m *mutex) Unlock() {
	if hle {
		m.ms.Unlock()
	} else {
		m.m.Unlock()
	}
}
