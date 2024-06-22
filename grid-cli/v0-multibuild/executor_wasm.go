//go:build js && wasm
// +build js,wasm

package main

import "fmt"

type WasmCommandExecutor struct{}

func (wce *WasmCommandExecutor) ExecuteCommand(path string, args []string) error {
	fmt.Println("Executing command in WASM environment is not supported yet")
	return nil
}

func (wce *WasmCommandExecutor) ShowPromise(path string) (string, error) {
	fmt.Println("Showing promise in WASM environment is not supported yet")
	return "", nil
}
