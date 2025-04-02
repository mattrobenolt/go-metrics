package metrics

import (
	"fmt"
	"hash/maphash"
	"runtime"
	"strings"
	"weak"

	"go.withmatt.com/metrics/internal/syncx"
)

var identCache syncx.Map[uint64, weak.Pointer[string]]

// makeIdentHandle is a specialized version of unique.Make[string]
// that also caches the fact that a given Ident is a valid identifier.
// Caching uniqueness at this point makes it significantly faster to
// repeatedly look up the same identifier.
func makeIdentHandle(value string) Ident {
	key := maphash.String(globalSeed, value)

	// Keep around any values we allocate for insertion. There
	// are a few different ways we can race with other threads
	// and create values that we might discard. By keeping
	// the first one we make around, we can avoid generating
	// more than one per racing thread.
	var (
		toInsert     *string // Keep this around to keep it alive.
		toInsertWeak weak.Pointer[string]
	)
	var ptr *string
	for {
		// Check the map.
		wp, ok := identCache.Load(key)
		if !ok {
			// Try to insert a new value into the map.
			if toInsert == nil {
				if !validateIdent(value) {
					panic(fmt.Sprintf("metrics: invalid identifier: %q", value))
				}
				toInsert = new(string)
				*toInsert = strings.Clone(value)
				toInsertWeak = weak.Make(toInsert)
			}
			wp, _ = identCache.LoadOrStore(key, toInsertWeak)
		}
		// Now that we're sure there's a value in the map, let's
		// try to get the pointer we need out of it.
		ptr = wp.Value()
		if ptr != nil {
			break
		}
		// The weak pointer is nil, so the old value is truly dead.
		// Try to remove it and start over.
		identCache.CompareAndDelete(key, wp)
	}
	runtime.KeepAlive(toInsert)
	return Ident{ptr}
}

func rangeIdentCache(f func(uint64, weak.Pointer[string]) bool) {
	identCache.Range(func(key uint64, wp weak.Pointer[string]) bool {
		// while we are iterating, we might as well clean up any dead pointers
		if ptr := wp.Value(); ptr == nil {
			if identCache.CompareAndDelete(key, wp) {
				return true
			}
		}
		return f(key, wp)
	})
}
