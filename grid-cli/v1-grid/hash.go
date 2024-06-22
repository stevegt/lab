package main

import (
	"crypto/sha256"

	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// GenerateHash generates a hash of the given data using the specified algorithm.
func GenerateHash(algo int, inBuf []byte) (mBuf []byte, err error) {
	sumBuf := sha256.Sum256(inBuf)
	// Create a new multihash with it.
	mBuf, err = multihash.Encode(sumBuf[:], uint64(algo))
	Ck(err)
	// Print the multihash as hex string
	// fmt.Printf("hex: %s\n", hex.EncodeToString(mHashBuf))
	return
}
