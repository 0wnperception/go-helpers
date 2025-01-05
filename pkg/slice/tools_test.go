package slice

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContains(t *testing.T) {
	testData := []struct {
		result bool
		value  string
		data   []string
	}{
		{false, "a", nil},
		{false, "a", []string{}},
		{false, "a", []string{"b"}},
		{false, "a", []string{"b", "c"}},
		{true, "a", []string{"b", "c", "a"}},
	}

	for i, d := range testData {
		require.Equalf(t, d.result, Contains(d.value, d.data), "test index: %d", i)
	}
}

func TestDeduplicate(t *testing.T) {
	i := []string{"a", "b", "c", "a", "c"}

	result := Deduplicate(i)

	sort.Strings(result)

	require.Equal(t, 3, len(result))
	require.Equal(t, []string{"a", "b", "c"}, result)

	i = []string{"b", "a", "c"}
	result = Deduplicate(i)

	sort.Strings(result)

	require.Equal(t, 3, len(result))
	require.Equal(t, []string{"a", "b", "c"}, result)
}

func TestHasSubSlice(t *testing.T) {
	testData := []struct {
		orig      []string
		requested []string
		result    bool
	}{
		{[]string{"e1", "e5", "e3"}, []string{"e3", "e1"}, true},
		{[]string{"e1", "e5", "e3"}, []string{"e6", "e1"}, false},
		{[]string{"e1", "e5", "e3"}, []string{}, true},
		{[]string{"e1", "e5", "e3"}, nil, true},
		{[]string{}, []string{"e3", "e1"}, false},
		{nil, []string{"e3", "e1"}, false},
		{nil, []string{}, true},
		{nil, nil, true},
	}

	for i, d := range testData {
		require.Equal(t, d.result, HasSubSlice(d.orig, d.requested), "test index: %d", i)
	}
}

func TestAppend(t *testing.T) {
	r := Append("1", "2", "3")

	require.Equal(t, 3, len(r))
	require.Equal(t, 3, cap(r))
	require.Equal(t, []string{"2", "3", "1"}, r)

	r = Append("1")
	require.Equal(t, 1, len(r))
	require.Equal(t, 1, cap(r))
	require.Equal(t, []string{"1"}, r)
}
