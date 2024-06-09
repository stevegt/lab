package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type KernelNative struct {
	fs      afero.Fs
	baseDir string
	util    *afero.Afero
}

// NewKernelNative creates a new Kernel instance that uses the native
// filesystem, CPU, and network stack.
func NewKernelNative(fs afero.Fs, baseDir string) *KernelNative {
	sys := &KernelNative{
		fs:      fs,
		baseDir: baseDir,
		util:    &afero.Afero{Fs: fs},
	}
	sys.ensureDirectories()
	return sys
}

func (sys *KernelNative) ensureDirectories() {
	directories := []string{gridDir, cacheDir}
	for _, dir := range directories {
		if _, err := sys.fs.Stat(filepath.Join(sys.baseDir, dir)); os.IsNotExist(err) {
			sys.fs.MkdirAll(filepath.Join(sys.baseDir, dir), os.ModePerm)
		}
	}
}

func (sys *KernelNative) getSymbolTableHash() (hash string, err error) {
	configPath := filepath.Join(sys.baseDir, configFile)
	data, err := sys.util.ReadFile(configPath)
	if err != nil {
		err = fmt.Errorf("Failed to read configuration: %v", err)
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "symbol_table_hash=") {
			return strings.TrimPrefix(line, "symbol_table_hash="), nil
		}
	}
	err = fmt.Errorf("Symbol table hash not found in configuration.")
	return "", err
}

func (sys *KernelNative) loadPeers() {
	peersPath := filepath.Join(sys.baseDir, peerList)
	file, err := sys.fs.Open(peersPath)
	if err != nil {
		fmt.Println("No peers available.")
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerAddress := scanner.Text()
		Peers[peerAddress] = &Peer{Address: peerAddress}
	}
}
