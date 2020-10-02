// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package namepool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type pool struct {
	format    string
	idCounter uint64
	idPool    *sync.Pool
}

// Pool is a wrapper around a sync.Pool utilizing sync/atomic to provide
// simple name pools that are safe to use by multiple goroutines.
//
// The argument format should include exactly one format verb for base
// 10 integers (%d).
// It is not an error if format doesn't include any verbs. The ID will
// still be stored in acquired Names.
func Pool(format string) *pool {
	pool := &pool{
		format:    format,
		idCounter: 0,
	}

	pool.idPool = &sync.Pool{
		New: func() interface{} {
			newId := atomic.AddUint64(&pool.idCounter, 1)
			return &newId
		},
	}

	return pool
}

// Acquire returns a Name from the name pool.
func (pool *pool) Acquire() *Name {
	id := pool.idPool.Get().(*uint64)

	return &Name{
		name: fmt.Sprintf(pool.format, *id),
		id:   id,
	}
}

// Release returns a name to the name pool. The value name points to
// will be reset to a default Name.
func (pool *pool) Release(name *Name) {
	pool.idPool.Put(name.id)
	*name = Name{}
}
