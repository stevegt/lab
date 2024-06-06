package main

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// generate expected multihash
func genExpectedHash(algo int, inBuf []byte) []byte {
	sumBuf := sha256.Sum256(inBuf)
	// Create a new multihash with it.
	mBuf, err := multihash.Encode(sumBuf[:], uint64(algo))
	Ck(err)
	// Print the multihash as hex string
	// fmt.Printf("hex: %s\n", hex.EncodeToString(mHashBuf))
	return mBuf
}

// test generating a sha256 multihash
func TestGenerateHash(t *testing.T) {
	// generate a sha256 multihash for the test data
	expectedData := []byte("hello world")
	algo := multihash.SHA2_256
	expectedHash := genExpectedHash(algo, []byte(expectedData))
	hash, err := GenerateHash(algo, expectedData)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}
	// compare the generated hash with the expected hash
	if bytes.Compare(hash, expectedHash) != 0 {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}
