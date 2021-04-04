package aghstrings

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneSlice_family(t *testing.T) {
	t.Run("cloneslice_simple", func(t *testing.T) {
		a := []string{"1", "2", "3"}
		assert.Equal(t, a, CloneSlice(a))
	})

	t.Run("cloneslice_nil", func(t *testing.T) {
		assert.Nil(t, CloneSlice(nil))
	})

	t.Run("clonesliceorempty_nil", func(t *testing.T) {
		assert.Equal(t, []string{}, CloneSliceOrEmpty(nil))
	})

	t.Run("clonesliceorempty_sameness", func(t *testing.T) {
		a := []string{"1", "2"}
		assert.Equal(t, CloneSlice(a), CloneSliceOrEmpty(a))
	})
}

func TestInSlice(t *testing.T) {
	simpleStrs := []string{"1", "2", "3"}

	testCases := []struct {
		name string
		str  string
		strs []string
		want bool
	}{{
		name: "yes",
		str:  "2",
		strs: simpleStrs,
		want: true,
	}, {
		name: "no",
		str:  "4",
		strs: simpleStrs,
		want: false,
	}, {
		name: "nil",
		str:  "any",
		strs: nil,
		want: false,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, InSlice(tc.strs, tc.str))
		})
	}
}

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
	t.Run("ordinary", func(t *testing.T) {
		s := " a,b , c "
		require.Equal(t, "a", SplitNext(&s, ','))
		require.Equal(t, "b", SplitNext(&s, ','))
		require.Equal(t, "c", SplitNext(&s, ','))

		assert.Empty(t, s)
	})

	t.Run("nil_source", func(t *testing.T) {
		assert.Equal(t, "", SplitNext(nil, 's'))
	})
}

func TestWriteToBuilder(t *testing.T) {
	b := &strings.Builder{}

	t.Run("single", func(t *testing.T) {
		assert.NotPanics(t, func() { WriteToBuilder(b, t.Name()) })
		assert.Equal(t, t.Name(), b.String())
	})

	b.Reset()
	t.Run("several", func(t *testing.T) {
		const (
			_1   = "one"
			_2   = "two"
			_123 = _1 + _2
		)
		assert.NotPanics(t, func() { WriteToBuilder(b, _1, _2) })
		assert.Equal(t, _123, b.String())
	})

	b.Reset()
	t.Run("nothing", func(t *testing.T) {
		assert.NotPanics(t, func() { WriteToBuilder(b) })
		assert.Equal(t, "", b.String())
	})

	t.Run("nil_builder", func(t *testing.T) {
		assert.Panics(t, func() { WriteToBuilder(nil, "a") })
	})
}
