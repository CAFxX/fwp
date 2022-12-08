//go:build !amd64 || race

package fwp

import "sync"

type mutex struct{ sync.Mutex }
