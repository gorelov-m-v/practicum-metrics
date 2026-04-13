package pool

import (
	"testing"
)

type testObj struct {
	Name  string
	Value int
	Items []int
}

func (t *testObj) Reset() {
	if t == nil {
		return
	}
	t.Name = ""
	t.Value = 0
	t.Items = t.Items[:0]
}

func TestPool_GetPut(t *testing.T) {
	p := New(func() *testObj {
		return &testObj{Items: make([]int, 0, 8)}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}

	obj.Name = "test"
	obj.Value = 42
	obj.Items = append(obj.Items, 1, 2, 3)

	p.Put(obj)

	obj2 := p.Get()
	if obj2.Name != "" || obj2.Value != 0 || len(obj2.Items) != 0 {
		t.Fatalf("object not reset: Name=%q Value=%d Items=%v", obj2.Name, obj2.Value, obj2.Items)
	}
}

func TestPool_SliceCapPreserved(t *testing.T) {
	p := New(func() *testObj {
		return &testObj{Items: make([]int, 0, 16)}
	})

	obj := p.Get()
	obj.Items = append(obj.Items, 1, 2, 3)
	p.Put(obj)

	obj2 := p.Get()
	if cap(obj2.Items) < 3 {
		t.Fatal("slice capacity not preserved after reset")
	}
}

func BenchmarkPool(b *testing.B) {
	p := New(func() *testObj {
		return &testObj{Items: make([]int, 0, 8)}
	})

	for b.Loop() {
		obj := p.Get()
		obj.Name = "bench"
		obj.Value = 99
		p.Put(obj)
	}
}
