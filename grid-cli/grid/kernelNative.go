package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	. "github.com/stevegt/goadapt"
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

func (sys *KernelNative) Exec(subcommand string, args []string) (err error) {
	symbolTableHash, err := sys.getSymbolTableHash()
	Ck(err)
	symbolTable := fetchSymbolTable(symbolTableHash)
	subcommandHash := getSubcommandHash(symbolTable, subcommand)
	module := sys.fetchModule(subcommandHash)
	cmd := exec.Command(module, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error executing %v %v: %v", subcommand, args, err)
	}
	return nil
}

func (sys *KernelNative) fetchLocalData(mBuf []byte) ([]byte, error) {
	fn := fmt.Sprintf("%x", mBuf)
	cachePath := filepath.Join(sys.baseDir, cacheDir, fn)
	data, err := sys.util.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	// XXX If data not found in cache, check if it's a known handler
	// handlerPath := filepath.Join(os.Getenv("HOME"), gridDir, "handlers", hash)
	// return ioutil.ReadFile(handlerPath)

	return nil, fmt.Errorf("Data not found.")
}
