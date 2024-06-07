package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/multiformats/go-multihash"
	"github.com/spf13/afero"
	. "github.com/stevegt/goadapt"
)

func setupTestEnv() (sys *System) {
	sys = NewSys(afero.NewMemMapFs(), "/tmp/foo")
	return
}

func TestEnsureDirectories(t *testing.T) {
	sys := setupTestEnv()

	expectedDirs := []string{gridDir, cacheDir}
	for _, dir := range expectedDirs {
		_, err := sys.Fs.Stat(filepath.Join(sys.BaseDir, dir))
		if os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}

func TestFetchLocalData(t *testing.T) {
	sys := setupTestEnv()

	expectedData := []byte("test data")
	// generate a sha256 multihash for the test data

	algo := multihash.SHA2_256
	mBuf, err := GenerateHash(algo, expectedData)
	Tassert(t, err == nil, "Failed to generate hash: %v", err)

	// convert the hash to a string for use as a filename
	fn := fmt.Sprintf("%x", mBuf)
	fn2 := fmt.Sprintf("%s", hex.EncodeToString(mBuf))
	Tassert(t, fn == fn2, "Mismatched hash strings: %s != %s", fn, fn2)

	cachePath := filepath.Join(sys.BaseDir, cacheDir, fn)

	err = sys.util.WriteFile(cachePath, []byte(expectedData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to %s: %v", cachePath, err)
	}

	data, err := sys.fetchLocalData(mBuf)
	if err != nil {
		t.Errorf("fetchLocalData returned an error: %v", err)
	}
	if bytes.Compare(data, expectedData) != 0 {
		t.Errorf("fetchLocalData returned unexpected data: got %v want %v", string(data), expectedData)
	}
}

// Further tests would follow the established pattern of setting up necessary test data
// and then calling the function under test. For example:

// Test the scenario where file is not found
func TestFetchLocalData_NotFound(t *testing.T) {
	sys := setupTestEnv()

	_, err := sys.fetchLocalData([]byte("non-existent"))
	if err == nil {
		t.Error("Expected an error for non-existent data, but got nil")
	}
}

// Ensure test setup includes expected environment
func TestGetSymbolTableHash_NonExistentFile(t *testing.T) {
	sys := setupTestEnv()

	// Intentionally not creating the file to trigger the file not found path
	_, err := sys.getSymbolTableHash()
	if err == nil {
		t.Fatal("Expected error when configuration file does not exist, got nil")
	}
}

// test loadPeers
func TestLoadPeers(t *testing.T) {
	sys := setupTestEnv()

	// create a test file with some peers
	peerData := []byte("peer1\npeer2\npeer3")
	err := sys.util.WriteFile(filepath.Join(sys.BaseDir, peerList), peerData, 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to peers.txt: %v", err)
	}

	sys.loadPeers()
	if err != nil {
		t.Fatalf("loadPeers returned an error: %v", err)
	}
	if len(Peers) != 3 {
		t.Errorf("loadPeers returned unexpected number of peers: got %d want 3", len(Peers))
	}
}

// test getSymbolTableHash
func TestGetSymbolTableHash(t *testing.T) {
	sys := setupTestEnv()

	// create a test configuration file with symbol_table_hash set
	expectedHash := "testhash"
	line := []byte(fmt.Sprintf("symbol_table_hash=%s", expectedHash))
	err := sys.util.WriteFile(filepath.Join(sys.BaseDir, configFile), line, 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to symbol_table_hash: %v", err)
	}

	hash, err := sys.getSymbolTableHash()
	if err != nil {
		t.Fatalf("getSymbolTableHash returned an error: %v", err)
	}
	if hash != expectedHash {
		t.Errorf("getSymbolTableHash returned unexpected hash: got %v want %v", hash, expectedHash)
	}
}

// test getSubcommandHash
func TestGetSubcommandHash(t *testing.T) {

	// create a symbol table with several entries
	// each entry is a line with the format: <subcommand> <hash>
	symbolTable := "subcommand1 testhash1\nsubcommand2 testhash2\nsubcommand3 testhash3"

	hash := getSubcommandHash(symbolTable, "subcommand2")
	if hash != "testhash2" {
		t.Errorf("getSubcommandHash returned unexpected hash: got %v want testhash1", hash)
	}
}
