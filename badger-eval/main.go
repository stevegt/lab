package main

import (
	"log"
	"math/rand"
	"time"

	. "github.com/stevegt/goadapt"

	badger "github.com/dgraph-io/badger/v4"
)

func main() {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	// opt := badger.DefaultOptions("/tmp/badger-testdb")
	// in-memory database
	opt := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opt)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	totalOps := 100000
	start := time.Now()
	for ops := 0; ops < totalOps; ops++ {
		// create a long byte slice to be used as a key
		var key []byte
		for i := 0; i < 32; i++ {
			k := rand.Intn(256)
			key = append(key, byte(k))
		}

		// create a long byte slice to be used as a value
		var value []byte
		for i := 0; i < 80; i++ {
			// random value
			v := rand.Intn(256)
			value = append(value, byte(v))
		}

		// Pf("The key is: %x\n", key)
		// Pf("The value is: %x\n", value)

		// set the key and value in the database
		err = db.Update(func(txn *badger.Txn) error {
			err := txn.Set(key, value)
			return err
		})
		Ck(err)

		// get the value from the database
		err = db.View(func(txn *badger.Txn) error {
			item, err := txn.Get(key)
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				// Pf("The retrieved value is: %x\n", val)
				return nil
			})
			return err
		})
		Ck(err)
	}
	stop := time.Now()
	Pf("elapsed: %v\n", stop.Sub(start))
	Pf("total ops: %v\n", totalOps)
	Pf("ops/sec: %v\n", float64(totalOps)/stop.Sub(start).Seconds())

}
