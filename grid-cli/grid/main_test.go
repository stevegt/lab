package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

func setupTestEnvironment() (cleanup func()) {
	// Create a temporary HOME directory for testing
	homeDir, err := ioutil.TempDir("", "grid-home")
	if err != nil {
		fmt.Println("Failed to create temporary HOME directory:", err)
		os.Exit(1)
	}

	// Override HOME environment variable
	os.Setenv("HOME", homeDir)

	return func() {
		os.RemoveAll(homeDir) // clean up
	}
}

func TestEnsureDirectories(t *testing.T) {
	cleanup := setupTestEnvironment()
	defer cleanup()
	ensureDirectories()

	expectedDirs := []string{gridDir, cacheDir, peersDir}
	for _, dir := range expectedDirs {
		_, err := os.Stat(filepath.Join(os.Getenv("HOME"), dir))
		if os.IsNotExist(err) {
			t.Errorf("Directory %s was not created", dir)
		}
	}
}

func TestFetchLocalData(t *testing.T) {
	cleanup := setupTestEnvironment()
	defer cleanup()
	ensureDirectories()

	expectedData := []byte("test data")
	// generate a sha256 multihash for the test data

	algo := multihash.SHA2_256
	mBuf, err := GenerateHash(algo, expectedData)
	Tassert(t, err == nil, "Failed to generate hash: %v", err)

	// convert the hash to a string for use as a filename
	fn := fmt.Sprintf("%x", mBuf)
	fn2 := fmt.Sprintf("%s", hex.EncodeToString(mBuf))
	Tassert(t, fn == fn2, "Mismatched hash strings: %s != %s", fn, fn2)

	cachePath := filepath.Join(os.Getenv("HOME"), cacheDir, fn)

	err = ioutil.WriteFile(cachePath, []byte(expectedData), 0644)
	if err != nil {
		t.Fatalf("Failed to write test data to %s: %v", cachePath, err)
	}

	data, err := fetchLocalData(mBuf)
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
	cleanup := setupTestEnvironment()
	defer cleanup()

	_, err := fetchLocalData([]byte("non-existent"))
	if err == nil {
		t.Error("Expected an error for non-existent data, but got nil")
	}
}

// Ensure test setup includes expected environment
func TestGetSymbolTableHash_NonExistentFile(t *testing.T) {
	cleanup := setupTestEnvironment()
	defer cleanup()

	// Intentionally not creating the file to trigger the file not found path
	_, err := getSymbolTableHash()
	if err == nil {
		t.Fatal("Expected error when configuration file does not exist, got nil")
	}
}
