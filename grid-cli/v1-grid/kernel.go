package main

import (
	"os"
)

// XXX Consider always compiling to WASM.
// XXX Move to promisegrid/x.

// Kernel represents the external world, presenting a syscall-like
// interface to callers.  Operations that go through this interface
// include anything that can have external side effects, including
// access to data storage, network, and execution of subcommands.
// This interface is intended to be implemented by a variety of
// backends, including a native backend that uses the host OS, and a
// WASM backend that runs in a sandboxed environment.
type Kernel interface {
	Stat(path string) (os.FileInfo, error)
	Open(path string) (os.File, error)
	Close() error
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Remove(path string) error
	RemoveAll(path string) error
	MkdirAll(path string, perm os.FileMode) error
	Exec(name string, args []string) error
}
