package metrics

import "hash/maphash"

// Metrics must to be uniqued by their family name and optional
// tags. The hash we generate is done so that we don't trigger
// allocations and can consistently hash the same metrics names
// no matter if they are pre-validated or not.
//
// Our consistent hash format is generated by:
//
//	hash( family *( 0x00 + label) *( 0xff + value) )
//
// * Write the family name as a string
// * if there are no tags, we're done
// * Write all of the tag labels, separated by 0x00
// * Write all of the values, separated by 0xff
//
// The hash is written this way so it's possible to generate
// a partial hash out of just the labels without known values.

const (
	labelDelimeter = 0x00
	valueDelimeter = 0xff
)

type metricHash uint64

const emptyHash metricHash = 0

// use a consistent seed for all hashing
var globalSeed = maphash.MakeSeed()

// getHashTags creates a hash out of metric parts
func getHashTags(family string, tags []Tag) metricHash {
	// Optimize when no tags, this variant is internally optimized.
	if len(tags) == 0 {
		return hashString(family)
	}

	var h maphash.Hash
	h.SetSeed(globalSeed)
	h.WriteString(family)

	for _, tag := range tags {
		h.WriteByte(labelDelimeter)
		h.WriteString(tag.label.String())
	}
	for _, tag := range tags {
		h.WriteByte(valueDelimeter)
		h.WriteString(tag.value.String())
	}

	return metricHash(h.Sum64())
}

// getHashStrings generates an identical hash to getHash,
// but operates on interleaved level value pairs.
func getHashStrings(family string, bits []string) metricHash {
	// Optimize when no tags, this variant is internally optimized.
	if len(bits) == 0 {
		return hashString(family)
	}

	var h maphash.Hash
	h.SetSeed(globalSeed)
	h.WriteString(family)

	// bits are interleaved, [label, value, label, value]
	for i := 0; i < len(bits); i += 2 {
		h.WriteByte(labelDelimeter)
		h.WriteString(bits[i])
	}
	for i := 1; i < len(bits); i += 2 {
		h.WriteByte(valueDelimeter)
		h.WriteString(bits[i])
	}

	return metricHash(h.Sum64())
}

// hashStart writes out the family + labels, and returns
// the hash to be finished with hashFinish.
func hashStart(family string, labels ...string) *maphash.Hash {
	var h maphash.Hash
	h.SetSeed(globalSeed)
	h.WriteString(family)

	for _, label := range labels {
		h.WriteByte(labelDelimeter)
		h.WriteString(label)
	}
	return &h
}

// hashFinish takes the input hash state, and finishes writing
// values, returning the resulting value. hashFinish does not mutate
// the starting state, so it can reused with different series of
// values.
func hashFinish(h *maphash.Hash, values ...string) metricHash {
	// Create a copy of our maphash so we can write to
	// this copy without mutating the starting partial hash.
	h2 := *h
	for _, value := range values {
		h2.WriteByte(valueDelimeter)
		h2.WriteString(value)
	}
	return metricHash(h2.Sum64())
}

// hashString generates a hash for a string.
func hashString(s string) metricHash {
	return metricHash(maphash.String(globalSeed, s))
}
