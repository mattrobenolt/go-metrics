package metrics

import (
	"testing"

	"go.withmatt.com/metrics/internal/assert"
)

func TestHash(t *testing.T) {
	for _, tc := range []struct {
		family string
		tags   []string
	}{
		{"my_family", nil},
		{"my_family", []string{"foo", "bar"}},
		{"a", []string{"foo", "bar"}},
	} {
		assert.Equal(t,
			getHashTags(tc.family, MustTags(tc.tags...)),
			getHashTags(tc.family, MustTags(tc.tags...)),
		)
		assert.NotEqual(t,
			getHashTags(tc.family, MustTags(tc.tags...)),
			getHashStrings("___", nil),
		)

		// getHashTags and getHashStrings must yield the same hash!
		assert.Equal(t,
			getHashTags(tc.family, MustTags(tc.tags...)),
			getHashStrings(tc.family, tc.tags),
		)
	}
}

func TestHashPartial(t *testing.T) {
	family := "my_family"
	labels := []string{"a", "b"}
	// values := []string{"foo"}

	state := hashStart(family, labels)

	assert.Equal(t,
		hashFinish(state, []string{"1", "2"}),
		hashFinish(state, []string{"1", "2"}),
	)
	assert.NotEqual(t,
		hashFinish(state, []string{"1", "2"}),
		hashFinish(state, []string{"1", "3"}),
	)

	assert.Equal(t,
		hashFinish(state, []string{"1", "2"}),
		getHashTags(family, []Tag{
			MustTag("a", "1"),
			MustTag("b", "2"),
		}),
	)
}
