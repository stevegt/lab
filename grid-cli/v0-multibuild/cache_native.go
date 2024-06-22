//go:build !js
// +build !js

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type NativeCacheStorage struct {
	cacheDir string
}

func (ncs *NativeCacheStorage) Save(hash string, data []byte) error {
	cachePath := filepath.Join(os.Getenv("HOME"), ncs.cacheDir, hash)
	return ioutil.WriteFile(cachePath, data, 0755)
}

func (ncs *NativeCacheStorage) Load(hash string) ([]byte, error) {
	cachePath := filepath.Join(os.Getenv("HOME"), ncs.cacheDir, hash)
	return ioutil.ReadFile(cachePath)
}
