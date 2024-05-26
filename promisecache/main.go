package main

import (
	"sync"
	// . "github.com/stevegt/goadapt"
)

/*

PromiseCache is an interface for a cache that supports promises. A
promise is also known as a future in some languages, a capability in
some security contexts, and an assertion here.  A generalized theory
of promises can be found in the works of Mark Burgess.  Some basic
principles of promises:

1. A promise describes an agent's intention to perform an action. 
2. An agent cannot make promises on behalf of another agent.
3. Promises are not contracts and can be broken.

In the context of a PromiseCache, a promise is a message.  A message
contains an author ID, a function address, zero or more arguments, a
value, and a signature.  The author ID is the author's public key. The
function address is a location in the author's content-addressable
storage.  The arguments are the parameters to the function.  The value
is the result of the function. The message is signed by the author. In
essence, the message is an assertion, promising that the value is
correct.  The value might be multipart, might be a final value, or it
might be the address and arguments of another function in the author's
content-addressable storage that when called will return another
promise; a callback.  

Here's an example flow:

1. Alice creates promise A.  Promise A describes callback Ac that can
   be used to retrieve value Av from Alice's content-addressable storage.
2. Alice provides the promise to the cache, the cache stores the
   promise, and create and returns promise B and revocation promise Br
   to Alice.
3. Alice provides promise B to Bob and Carol.
4. Bob calls the cache with promise B, the cache calls callback Ac,
   Ac returns a new promise Av containing the value, and the cache stores 
   Av and returns Acp to Bob.
5. Carol calls the cache with promise B, and the cache returns Av.
6. Alice calls the cache with promise Br, and the cache removes A and
   Av from the cache.

If callback Ac fails, the cache returns an error to Bob, and retries
the callback when Carol calls the cache.  

Alice might want only Bob to have access to Av.  In this case, she can
encrypt Av with Bob's public key.  If she wants to avoid callback Ac
from being called via the cache by Carol, she can also encrypt B.








//

, a promise is a value that is not
// yet available.  The cache will use the promise as a key.  When
// Alice calls the Store method, the cache returns a promise.  to invoke the
// promise for a given amount of time.  When Bob calls the Load method
// with the 
// 
// can 
// value associated with the key is another promise that contains the
// value.  When the value is available, the promise is resolved and
// the value is available.  The cache is thread-safe.  The cache




Applications"
// (2015).  A promise is a value that is not yet available. available.
// Keys are promises, and the value associated with the key is another
// promise that contains the value.  When the value is available, the
// promise is resolved and the value is stored in the cache.


The cache will store the promise and return it when
// the value is available.  The cache is thread-safe.  The cache
// supports a TTL (time-to-live) for keys, which is the maximum amount
// of time a key can be stored in the cache.

type Cache interface {


}


// Vset is a set of values to be stored in a hash map
type Vset struct {
	keys []any
	values []any
}

// Get returns the value associated with the key
func (v *Vset) Get(key any) (value any, ok bool) {
	for i, k := range v.keys {
		if k == key {
			return v.values[i]
		}
	}
	return nil, false
}

// Set sets the value associated with the key
func (v *Vset) Set(key, value any) {
	for i, k := range v.keys {
		if k == key {
			v.values[i] = value
			return
		}
	}
	v.keys = append(v.keys, key)
	v.values = append(v.values, value)
}

// Delete removes the key and its associated value
func (v *Vset) Delete(key any) {
	for i, k := range v.keys {
		if k == key {
			v.keys = append(v.keys[:i], v.keys[i+1:]...)
			v.values = append(v.values[:i], v.values[i+1:]...)
			return
		}
	}
}

// Get returns the value associated with the key
func (m *Map) Get(key any) (value any, ok bool) {
	vset, ok := m.vsets.Load(key)
	return vset.Get(key)
}

// Set sets the value associated with the key
func (m *Map) Set(key, value any) {
	vset, ok := m.vsets.Load(key)
	if !ok {
		vset = new(Vset)
		m.vsets.Store(key, vset)
	}
	vset.Set(key, value)
}


func main() {
	var m sync.Map
	k1 := []byte("key1")
	m.Store(k1, true)
	k2 := []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	m.Store(k2, true)

	fn := func(k, v interface{}) bool {
		println(string(k.([]byte)), v)
		return true
	}

	m.Range(fn)
}
