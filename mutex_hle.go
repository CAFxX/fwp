//go:build amd64 && !race

package fwp

import "github.com/linux4life798/safetyfast"

type mutex struct{ safetyfast.SpinHLEMutex }
