//go:build js && wasm
// +build js,wasm

package main

import "fmt"

type WasmCacheStorage struct {
	cache map[string][]byte
}

func (wcs *WasmCacheStorage) Save(hash string, data []byte) error {
	wcs.cache[hash] = data
	return nil
}

func (wcs *WasmCacheStorage) Load(hash string) ([]byte, error) {
	data, exists := wcs.cache[hash]
	if !exists {
		return nil, fmt.Errorf("data not found")
	}
	return data, nil
}
