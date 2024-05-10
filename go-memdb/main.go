package main

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/stevegt/goadapt"

	"github.com/hashicorp/go-memdb"
)

// Create a sample struct
type Person struct {
	Email  string
	Name   string
	Age    int
	Height float64
}

func main() {

	// Create the DB schema
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"person": &memdb.TableSchema{
				Name: "person",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Email"},
					},
					"age": &memdb.IndexSchema{
						Name:    "age",
						Unique:  false,
						Indexer: &memdb.IntFieldIndex{Field: "Age"},
					},
					"height": &memdb.IndexSchema{
						Name:    "height",
						Unique:  false,
						Indexer: &FloatFieldIndex{Field: "Height"},
					},
				},
			},
		},
	}

	// Create a new data base
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}

	// Create a write transaction
	txn := db.Txn(true)

	// Insert some people
	people := []*Person{
		&Person{"joe@aol.com", "Joe", 30, 110.3},
		&Person{"lucy@aol.com", "Lucy", 35, 120.3},
		&Person{"tariq@aol.com", "Tariq", 21, 80.5},
		&Person{"dorothy@aol.com", "Dorothy", 17, 98.5},
	}
	for _, p := range people {
		if err := txn.Insert("person", p); err != nil {
			panic(err)
		}
	}

	// Commit the transaction
	txn.Commit()

	// Create read-only transaction
	txn = db.Txn(false)
	// defer txn.Abort()

	// Lookup by email
	raw, err := txn.First("person", "id", "joe@aol.com")
	if err != nil {
		panic(err)
	}

	// Say hi!
	fmt.Printf("Hello %s!\n", raw.(*Person).Name)

	// List all the people
	it, err := txn.Get("person", "id")
	if err != nil {
		panic(err)
	}

	fmt.Println("All the people:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		fmt.Printf("  %s\n", p.Name)
	}

	// Range scan over people with ages between 25 and 35 inclusive
	it, err = txn.LowerBound("person", "age", 25)
	if err != nil {
		panic(err)
	}

	fmt.Println("People aged 25 - 35:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		if p.Age > 35 {
			break
		}
		fmt.Printf("  %s is aged %d and is %f tall\n", p.Name, p.Age, p.Height)
	}
	// Output:
	// Hello Joe!
	// All the people:
	//   Dorothy
	//   Joe
	//   Lucy
	//   Tariq
	// People aged 25 - 35:
	//   Joe is aged 30
	//   Lucy is aged 35

	// try a height range scan
	it, err = txn.LowerBound("person", "height", 100.0)
	if err != nil {
		panic(err)
	}

	fmt.Println("People taller than 100:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		fmt.Printf("  %s is %f tall\n", p.Name, p.Height)
	}

	// close the read-only transaction
	txn.Abort()

	// insert a bazillion people with random ages in a wide range
	count := 100000

	// Create a write transaction
	txn = db.Txn(true)
	start := time.Now()
	for i := 0; i < count; i++ {
		email := Spf("foo%d@example.com", i)
		name := Spf("Foo%d", i)
		// age := i
		age := rand.Intn(9999999)
		height := rand.Float64() * 100
		p := &Person{email, name, age, height}
		if err := txn.Insert("person", p); err != nil {
			panic(err)
		}
	}
	stop := time.Now()
	// Commit the transaction
	txn.Commit()
	elapsed := stop.Sub(start)
	Pf("write ops/sec: %f\n", float64(count)/elapsed.Seconds())

	// read and count all the people with ages in the range 999-999999
	// Create read-only transaction
	txn = db.Txn(false)
	it, err = txn.LowerBound("person", "age", 999)
	if err != nil {
		panic(err)
	}
	start = time.Now()
	found := 0
	for obj := it.Next(); obj != nil; obj = it.Next() {
		p := obj.(*Person)
		if p.Age > 9999 {
			break
		}
		found++
	}
	stop = time.Now()
	// commit the transaction
	txn.Abort()
	elapsed = stop.Sub(start)
	Pf("found %d people\n", found)
	Pf("read ops/sec: %f\n", float64(found)/elapsed.Seconds())

}
