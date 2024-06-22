//go:build !js
// +build !js

package main

import (
	"os"
	"os/exec"
)

type NativeCommandExecutor struct{}

func (nce *NativeCommandExecutor) ExecuteCommand(path string, args []string) error {
	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (nce *NativeCommandExecutor) ShowPromise(path string) (string, error) {
	cmd := exec.Command(path, "--show-promise")
	output, err := cmd.Output()
	return string(output), err
}
