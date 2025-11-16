package intmap

import (
	"maps"
	"math/rand"
	"slices"
	"testing"
	"time"
	"unsafe"
)

func TestSet(t *testing.T) {
	s := NewSet[int](0)
	if sz := s.Len(); sz != 0 {
		t.Fatalf("length of set must be 0: %d", sz)
	}

	if added := s.Add(1); !added {
		t.Fatalf("1 must be added")
	}
	if sz := s.Len(); sz != 1 {
		t.Fatalf("length of set must be 1: %d", sz)
	}

	if added := s.Add(1); added {
		t.Fatalf("1 must not be added")
	}
	if sz := s.Len(); sz != 1 {
		t.Fatalf("length of set must be 1: %d", sz)
	}

	if added := s.Add(2); !added {
		t.Fatalf("2 must be added")
	}
	if sz := s.Len(); sz != 2 {
		t.Fatalf("length of set must be 2: %d", sz)
	}

	if !s.Has(1) || !s.Has(2) {
		t.Fatalf("set must have both 1 and 2")
	}

	sum := 0
	s.ForEach(func(k int) bool {
		sum += k
		return true
	})
	if sum != 3 {
		t.Fatalf("total sum of elements must be 3")
	}
}

func TestSetClear(t *testing.T) {
	s := NewSet[int](0)

	s.Add(0)
	s.Add(1)
	s.Add(2)
	if sz := s.Len(); sz != 3 {
		t.Fatalf("unexpected set len %d", sz)
	}
	if !s.Has(0) || !s.Has(1) || !s.Has(2) {
		t.Fatalf("set must contain 1 and 2 before Clear()")
	}

	s.Clear()
	if sz := s.Len(); sz != 0 {
		t.Fatalf("unexpected set len %d", sz)
	}
	if s.Has(0) || s.Has(1) || s.Has(2) {
		t.Fatalf("set must not contain 1 or 2 after Clear()")
	}
}

// Reference Set implementation.
type RefSet[K IntKey] struct {
	m *Map[K, struct{}]
}

func (s *RefSet[K]) init() {
	if s.m == nil {
		s.m = New[K, struct{}](10)
	}
}

func (s *RefSet[K]) Len() int {
	return s.m.Len()
}

func (s *RefSet[K]) Has(k K) bool {
	s.init()
	return s.m.Has(k)
}

func (s *RefSet[K]) Del(k K) bool {
	return s.m.Del(k)
}

func (s *RefSet[K]) Add(k K) bool {
	s.init()
	_, ok := s.m.PutIfNotExists(k, struct{}{})
	return ok
}

func (s *RefSet[K]) ForEach(f func(K) bool) {
	s.m.ForEach(func(k K, _ struct{}) bool {
		return f(k)
	})
}

func TestSetExhaustive(t *testing.T) {
	s := NewSet[int](0)
	m := RefSet[int]{}

	test := func(t *testing.T, op string, key int, got, want any) {
		t.Helper()
		if got != want {
			t.Errorf("%s(%d) = %v; want: %v", op, key, got, want)
		}
	}

	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	for range 100_000 {
		k := rr.Intn(10_000)
		ff := rr.Float64()
		if ff >= 0.5 {
			test(t, "Has", k, s.Has(k), m.Has(k))
			test(t, "Add", k, s.Add(k), m.Add(k))
			test(t, "Has", k, s.Has(k), true)
			test(t, "Len", k, s.Len(), m.Len())
		} else {
			test(t, "Has", k, s.Has(k), m.Has(k))
			test(t, "Del", k, s.Del(k), m.Del(k))
			test(t, "Has", k, s.Has(k), false)
			test(t, "Len", k, s.Len(), m.Len())
		}
		// Test ForEach
		if ff <= 0.01 {
			keys := make(map[int]struct{}, m.Len())
			m.ForEach(func(k int) bool {
				keys[k] = struct{}{}
				return true
			})
			s.ForEach(func(k int) bool {
				delete(keys, k)
				return true
			})
			if len(keys) != 0 {
				t.Errorf("ForEach: failed to visit the following keys: %+v", keys)
			}
		}
	}
}

func TestSetDel(t *testing.T) {
	s := NewSet[int](0)

	s.Add(1)
	s.Add(2)
	if sz := s.Len(); sz != 2 {
		t.Fatalf("unexpected set len %d", sz)
	}
	if !s.Has(1) || !s.Has(2) {
		t.Fatalf("set must contain 1 and 2 before Clear()")
	}

	if found := s.Del(27); found {
		t.Fatalf("set must not contain 27")
	}

	// Delete 2
	if found := s.Del(2); !found {
		t.Fatalf("set must contain 2 on delete")
	}
	if sz := s.Len(); sz != 1 {
		t.Fatalf("unexpected set len %d", sz)
	}
	if s.Has(2) {
		t.Fatalf("set must not contain 2 after Del(2)")
	}
	if found := s.Del(2); found {
		t.Fatalf("set must not contain 2 on seconb deletion")
	}

	// Delete 1
	if found := s.Del(1); !found {
		t.Fatalf("set must contain 1 on delete")
	}
	if sz := s.Len(); sz != 0 {
		t.Fatalf("unexpected set len %d", sz)
	}
	if s.Has(1) {
		t.Fatalf("set must not contain 1 after Del(1)")
	}
	if found := s.Del(1); found {
		t.Fatalf("set must not contain 1 on seconb deletion")
	}
}

func TestNilSet(t *testing.T) {
	var s *Set[int]

	if sz := s.Len(); sz != 0 {
		t.Fatalf("length of nil set must be 0: %d", sz)
	}

	if s.Has(0) || s.Has(1) {
		t.Fatalf("nil set must not have 0 or 1")
	}

	count := 0
	s.ForEach(func(k int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Fatalf("total count of elements in nil set must be 0")
	}

}

func TestSetIter(t *testing.T) {
	s := NewSet[int](10)
	for i := 0; i < 100; i++ {
		s.Add(i)
	}

	sum := 0
	for k := range s.All() {
		sum += k
	}

	const sumTo99 = 99 * (99 + 1) / 2
	if sum != sumTo99 {
		t.Fatalf("unexpected sum: %d, want %d", sum, sumTo99)
	}
}

func testSetMemoryUsage[K IntKey]() func(t *testing.T) {
	return func(t *testing.T) {
		s := NewSet[K](16)
		m := New[K, struct{}](16)
		szs := int(unsafe.Sizeof(s.data[0]))
		msz := int(unsafe.Sizeof(m.data[0]))
		if szs*2 < msz {
			t.Fatalf("The per-element size of Set[%T] (%d) should be half that of Map[%T] (%d)",
				s.data[0], szs, m.data[0], msz)
		}
		t.Logf("Sizeof(Set[%T]) = %d; Sizeof(Map[%T]) = %d",
			s.data[0], szs, m.data[0], msz)
	}
}

func TestSetMemoryUsage(t *testing.T) {
	t.Run("int", testSetMemoryUsage[int]())
	t.Run("int8", testSetMemoryUsage[int8]())
	t.Run("int16", testSetMemoryUsage[int16]())
	t.Run("int32", testSetMemoryUsage[int32]())
	t.Run("int64", testSetMemoryUsage[int64]())

	t.Run("uint", testSetMemoryUsage[uint]())
	t.Run("uint8", testSetMemoryUsage[uint8]())
	t.Run("uint16", testSetMemoryUsage[uint16]())
	t.Run("uint32", testSetMemoryUsage[uint32]())
	t.Run("uint64", testSetMemoryUsage[uint64]())
}

func BenchmarkSetAdd10K(b *testing.B) {
	rr := rand.New(rand.NewSource(12345))
	seen := make(map[int]struct{}, 10_000)
	for len(seen) < 10_000 {
		seen[int(rr.Int31())] = struct{}{}
	}
	keys := slices.AppendSeq(make([]int, 0, 10_000), maps.Keys(seen))

	s := NewSet[int](10_000)
	b.ResetTimer()

	j := 0
	for i := 0; i < b.N; i++ {
		s.Add(keys[j])
		j++
		if j == 10_000 {
			j = 0
			s.Clear()
		}
	}
}

func BenchmarkSetDel10K(b *testing.B) {
	rr := rand.New(rand.NewSource(12345))

	s := NewSet[int](10_000)
	keys := make([]int, 0, 10_000)
	for s.Len() < 10_000 {
		v := rr.Int()
		if s.Add(v) {
			keys = append(keys, v)
		}
	}

	// Duplicate S to make resetting it back
	// to a populated state faster.
	dup := *s
	dup.data = slices.Clone(dup.data)

	b.ResetTimer()

	j := 0
	for i := 0; i < b.N; i++ {
		s.Del(keys[j])
		j++
		if j == 10_000 {
			j = 0
			b.StopTimer()
			copy(s.data, dup.data)
			s.size = dup.size
			s.hasZeroKey = dup.hasZeroKey
			b.StartTimer()
		}
	}
}

func BenchmarkSetHas(b *testing.B) {
	rr := rand.New(rand.NewSource(12345))
	s := NewSet[int32](256)
	var keys [32]int32
	for i := 0; s.Len() < 256; i++ {
		v := int32(rr.Int31())
		if s.Add(v) && i < len(keys) {
			keys[i] = v
			i++
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Has(keys[i%len(keys)])
	}
}
