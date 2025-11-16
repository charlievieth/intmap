package intmap

import "iter"

// Set is a specialization of Map modelling a set of integers.
// Like Map, methods that read from the set are valid on the nil Set.
// This include Has, Len, and ForEach.
type Set[K IntKey] struct {
	data       []K
	size       int
	hasZeroKey bool
}

// New creates a new map with keys being any integer subtype.
// The map can store up to the given capacity before reallocation and rehashing occurs.
func NewSet[K IntKey](capacity int) *Set[K] {
	return &Set[K]{
		data: make([]K, arraySize(capacity, fillFactor64)),
	}
}

// Add an element to the set. Returns true if the element was not already present.
func (s *Set[K]) Add(key K) bool {
	if key == K(0) {
		exists := s.hasZeroKey
		if !exists {
			s.size++
			s.hasZeroKey = true
		}
		return !exists
	}

	idx := s.startIndex(key)
	p := s.data[idx]
	if p == key {
		return false
	}

	sizeThreshold := s.sizeThreshold()
	if p == K(0) { // end of chain already
		s.data[idx] = key
		if s.size >= sizeThreshold {
			s.rehash()
		} else {
			s.size++
		}
		return true
	}

	// hash collision, seek next empty or key match
	for {
		idx = s.nextIndex(idx)
		p = s.data[idx]
		if p == K(0) {
			s.data[idx] = key
			if s.size >= sizeThreshold {
				s.rehash()
			} else {
				s.size++
			}
			return true
		} else if p == key {
			// Value already present
			return false
		}
	}
}

// Del deletes a key, returning true if the key was found.
func (s *Set[K]) Del(key K) bool {
	if key == K(0) {
		if s.hasZeroKey {
			s.hasZeroKey = false
			s.size--
			return true
		}
		return false
	}

	idx := s.startIndex(key)
	p := s.data[idx]
	if p == key {
		s.shiftKeys(idx)
		s.size--
		return true
	}
	if p == K(0) {
		return false
	}

	// hash collision, seek next empty or key match
	for {
		idx = s.nextIndex(idx)
		p = s.data[idx]
		if p == key {
			s.shiftKeys(idx)
			s.size--
			return true
		} else if p == K(0) {
			// Value already present
			return false
		}
	}
}

// Clear removes all items from the Set, but keeps the internal buffers for reuse.
func (s *Set[K]) Clear() {
	s.hasZeroKey = false
	s.size = 0
	clear(s.data)
}

// Has checks if the given key exists in the map.
// Calling this method on a nil map will return false.
func (s *Set[K]) Has(key K) bool {
	if s == nil {
		return false
	}

	if key == K(0) {
		return s.hasZeroKey
	}

	idx := s.startIndex(key)
	p := s.data[idx]

	if p == K(0) { // end of chain already
		return false
	}
	if p == key { // we check zero prior to this call
		return true
	}

	// hash collision, seek next hash match, bailing on first empty
	for {
		idx = s.nextIndex(idx)
		p = s.data[idx]
		if p == K(0) {
			return false
		}
		if p == key {
			return true
		}
	}
}

// Len returns the number of elements in the set.
// If the set is nil this method return 0.
func (s *Set[K]) Len() int {
	if s == nil {
		return 0
	}
	return s.size
}

// ForEach iterates over the elements in the set while the visit function returns true.
// This method returns immediately if the set is nil.
//
// The iteration order of a Set is not defined, so please avoid relying on it.
func (s *Set[K]) ForEach(visit func(k K) bool) {
	if s == nil {
		return
	}
	if s.hasZeroKey && !visit(K(0)) {
		return
	}
	for _, k := range s.data {
		if k != K(0) && !visit(k) {
			return
		}
	}
}

// All returns an iterator over keys from the set.
// The iterator returns immediately if the set is nil.
//
// The iteration order of a Set is not defined, so please avoid relying on it.
func (s *Set[K]) All() iter.Seq[K] {
	return s.ForEach
}

func (s *Set[K]) shiftKeys(idx int) int {
	// Shift entries with the same hash.
	// We need to do this on deletion to ensure we don't have zeroes in the hash chain
	for {
		var p K
		lastIdx := idx
		idx = s.nextIndex(idx)
		for {
			p = s.data[idx]
			if p == K(0) {
				s.data[lastIdx] = 0
				return lastIdx
			}

			slot := s.startIndex(p)
			if lastIdx <= idx {
				if lastIdx >= slot || slot > idx {
					break
				}
			} else {
				if lastIdx >= slot && slot > idx {
					break
				}
			}
			idx = s.nextIndex(idx)
		}
		s.data[lastIdx] = p
	}
}

func (s *Set[K]) rehash() {
	oldData := s.data
	s.data = make([]K, 2*len(s.data))

	// reset size
	if s.hasZeroKey {
		s.size = 1
	} else {
		s.size = 0
	}

	if s.hasZeroKey {
		s.Add(K(0))
	}
	for _, k := range oldData {
		if k != 0 {
			s.Add(k)
		}
	}
}

func (s *Set[K]) sizeThreshold() int {
	return int(uint64(len(s.data)) * fillFactorBase64 / 10)
}

func (s *Set[K]) startIndex(key K) int {
	return startIndex(int(key), len(s.data))
}

func (s *Set[K]) nextIndex(idx int) int {
	return nextIndex(idx, len(s.data))
}
