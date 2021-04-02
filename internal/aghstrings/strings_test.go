package aghstrings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetsSubtract(t *testing.T) {
	testCases := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{{
		name: "nil_a",
		a:    nil,
		b:    []string{"1", "2", "3"},
		want: nil,
	}, {
		name: "nil_b",
		a:    []string{"1", "2", "3"},
		b:    nil,
		want: []string{"1", "2", "3"},
	}, {
		name: "actual_intersection",
		a:    []string{"1", "2", "3"},
		b:    []string{"1", "2", "4"},
		want: []string{"3"},
	}, {
		name: "full_intersection",
		a:    []string{"1", "2", "3"},
		b:    []string{"1", "2", "3"},
		want: nil,
	}, {
		name: "no_intersection",
		a:    []string{"1", "2", "3"},
		b:    []string{"4", "5", "6"},
		want: []string{"1", "2", "3"},
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := SetSubtract(tc.a, tc.b)
			require.Len(t, c, len(tc.want))
			for _, str := range tc.want {
				assert.Contains(t, c, str)
			}
		})
	}
}

func TestSplitNext(t *testing.T) {
	s := " a,b , c "

	assert.Equal(t, "a", SplitNext(&s, ','))
	assert.Equal(t, "b", SplitNext(&s, ','))
	assert.Equal(t, "c", SplitNext(&s, ','))
	require.Empty(t, s)
}
