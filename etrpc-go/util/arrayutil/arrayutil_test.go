package arrayutil

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestRemove(t *testing.T) {
	assert.Equal(t, Remove([]int{1, 2, 3}, 4), []int{1, 2, 3})
	assert.Equal(t, Remove([]string{"a", "b", "c", "b", "c"}, "b"), []string{"a", "c", "c"})
}

func TestRemoveIf(t *testing.T) {
	assert.Equal(t, RemoveIf([]int{}, func(value int) bool { return value >= 3 }), []int{})
	assert.Equal(t, RemoveIf([]int{1, 2, 3, 4}, func(value int) bool { return value >= 3 }), []int{1, 2})
}

func TestFind(t *testing.T) {
	assert.Equal(t, Find([]int{}, 0), -1)
	assert.Equal(t, Find([]int{1, 2, 3, 4}, 0), -1)
	assert.Equal(t, Find([]int{1, 2, 3, 4, 4}, 4), 3)
}

func TestExist(t *testing.T) {
	assert.Equal(t, Exist([]string{"a", "b", "c", "b", "c"}, "a"), true)
	assert.Equal(t, Exist([]string{"a", "b", "c", "b", "c"}, "f"), false)
}

func TestFilter(t *testing.T) {
	assert.Equal(t, Filter([]int{1, 2, 3, 4, 5}, func(value int) bool { return value >= 3 }), []int{3, 4, 5})
}

func TestGroupBy(t *testing.T) {
	assert.Equal(t, GroupBy([]int{1, 2, 3, 4, 5, 6}, func(value int) int { return value % 2 }), map[int][]int{0: {2, 4, 6}, 1: {1, 3, 5}})
}

func TestToMap(t *testing.T) {
	assert.Equal(t, ToMap([]string{"a", "b", "c"}), map[string]interface{}{"a": 0, "b": 1, "c": 2})
}

func TestReverse(t *testing.T) {
	assert.True(t, reflect.DeepEqual(Reverse([]string{"a", "b", "c"}), []string{"c", "b", "a"}))
	assert.True(t, reflect.DeepEqual(Reverse([]int{1, 2, 3}), []int{3, 2, 1}))
}

func TestPartition(t *testing.T) {
	assert.Equal(t, Partition([]int{}, 1), [][]int{})
	assert.Equal(t, Partition([]int{1}, 1), [][]int{{1}})
	assert.Equal(t, Partition([]int{1, 2, 3}, 2), [][]int{{1, 2}, {3}})
}

func TestSort(t *testing.T) {
	type A struct {
		a int
		b string
	}
	assert.Equal(t, Sort([]A{{2, "a"}, {3, "b"}, {1, "c"}}, func(a, b A) bool {
		return a.a < b.a
	}), []A{{1, "c"}, {2, "a"}, {3, "b"}})
}

func TestSortAsc(t *testing.T) {
	assert.Equal(t, SortAsc([]int{2, 3, 1}), []int{1, 2, 3})
}

func TestSortDesc(t *testing.T) {
	assert.Equal(t, SortDesc([]int{2, 3, 1}), []int{3, 2, 1})
}

func TestDistinct(t *testing.T) {
	assert.True(t, reflect.DeepEqual(Distinct([]string{"a", "b", "c", "b", "c"}), []string{"a", "b", "c"}))
	assert.True(t, reflect.DeepEqual(Distinct([]int{1, 2, 3, 2, 3}), []int{1, 2, 3}))
}

func TestMap(t *testing.T) {
	assert.True(t, reflect.DeepEqual(Map([]string{"a", "b", "c"}, func(s string) string { return s + "x" }), []string{"ax", "bx", "cx"}))
}
