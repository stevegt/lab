package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// XXX Consider always compiling to WASM.  Move to promisegrid/x.

// System represents the underlying system, presenting a syscall-like
// interface to callers.  All operations go through this interface,
// including access to data storage, network and execution of subcommands.
type System struct {
	Fs      afero.Fs
	BaseDir string
	util    *afero.Afero
}

// NewSys creates a new Sys.
func NewSys(fs afero.Fs, baseDir string) *System {
	sys := &System{
		Fs:      fs,
		BaseDir: baseDir,
		util:    &afero.Afero{Fs: fs},
	}
	sys.ensureDirectories()
	return sys
}

func (sys *System) ensureDirectories() {
	directories := []string{gridDir, cacheDir}
	for _, dir := range directories {
		if _, err := sys.Fs.Stat(filepath.Join(sys.BaseDir, dir)); os.IsNotExist(err) {
			sys.Fs.MkdirAll(filepath.Join(sys.BaseDir, dir), os.ModePerm)
		}
	}
}

func (sys *System) getSymbolTableHash() (hash string, err error) {
	configPath := filepath.Join(sys.BaseDir, configFile)
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

func (sys *System) loadPeers() {
	peersPath := filepath.Join(sys.BaseDir, peerList)
	file, err := sys.Fs.Open(peersPath)
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
