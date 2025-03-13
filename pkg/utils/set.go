package utils

import "github.com/open3fs/m3fs/pkg/common"

// Set is a set
type Set[T comparable] map[T]struct{}

// Add adds an item to the set
func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

// Remove removes an item from the set
func (s Set[T]) Remove(item T) {
	delete(s, item)
}

// Contains returns true if the set contains the item
func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

// Len returns the number of items in the set
func (s Set[T]) Len() int {
	return len(s)
}

// AddIfNotExists adds an item to the set if it does not already exist
func (s Set[T]) AddIfNotExists(item T) bool {
	add := false
	if !s.Contains(item) {
		s.Add(item)
		add = true
	}

	return add
}

// ToSlice converts the set to a slice
func (s Set[T]) ToSlice() []T {
	ret := make([]T, 0, len(s))
	for item := range s {
		ret = append(ret, item)
	}
	return ret
}

// Equal check two set euqal
func (s Set[T]) Equal(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}
	for item := range s {
		if _, ok := other[item]; !ok {
			return false
		}
	}
	return true
}

// NewSet creates a new Set
func NewSet[T comparable](elems ...T) *Set[T] {
	s := common.Pointer(make(Set[T]))
	for _, elem := range elems {
		s.Add(elem)
	}
	return s
}
