package set

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// TestStringSet test
func TestStringSet(t *testing.T) {
	s := NewStringSet()
	k := ""
	rand.Seed(time.Now().Unix())
	n := 1000
	for i := 0; i < n; i++ {
		k = fmt.Sprintf("%v", rand.Int63())
		if !s.AddWithCheck(k) {
			t.Error()
		}
		if s.AddWithCheck(k) {
			t.Error()
		}
		if !s.Contain(k) {
			t.Error()
		}
		if !s.DeleteWithCheck(k) {
			t.Error()
		}
		s.Add(k)
	}

	r := s.Clone()
	fmt.Println(s.Get())
	fmt.Println(r.Get())
}
