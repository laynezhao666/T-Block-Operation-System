package osal

// Element is the element type of the set.
type Element interface{ comparable }

// Set is a set.
type Set[T Element] map[T]struct{}

// NewSet creates a new set.
func NewSet[T Element](elements ...T) Set[T] {
	set := make(Set[T])
	set.Add(elements...)
	return set
}

// Add adds elements to the set.
func (s Set[T]) Add(elements ...T) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// Remove removes elements from the set.
func (s Set[T]) Remove(elements ...T) {
	for _, element := range elements {
		delete(s, element)
	}
}

// Contains returns true if the set contains the element.
func (s Set[T]) Contains(element T) bool {
	_, ok := s[element]
	return ok
}

// ToSlice returns a slice of elements.
func (s Set[T]) ToSlice() []T {
	elements := []T{}
	for v := range s {
		elements = append(elements, v)
	}
	return elements
}
