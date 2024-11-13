package genericheap

import (
	"container/heap"
)

type GenericHeap[T any] struct {
	data     []T
	lessFunc func(a, b T) bool
}

func New[T any](lessFunc func(a, b T) bool) *GenericHeap[T] {
	return &GenericHeap[T]{
		data:     []T{},
		lessFunc: lessFunc,
	}
}

func (h *GenericHeap[T]) Len() int { return len(h.data) }

func (h *GenericHeap[T]) Less(i, j int) bool { return h.lessFunc(h.data[i], h.data[j]) }

func (h *GenericHeap[T]) Swap(i, j int) { h.data[i], h.data[j] = h.data[j], h.data[i] }

func (h *GenericHeap[T]) Push(x interface{}) {
	h.data = append(h.data, x.(T))
}

func (h *GenericHeap[T]) Pop() interface{} {
	old := h.data
	n := len(old)
	x := old[n-1]
	h.data = old[0 : n-1]
	return x
}

func (h *GenericHeap[T]) Add(item T) {
	heap.Push(h, item)
}

func (h *GenericHeap[T]) Remove() T {
	if h.Len() == 0 {
		var zero T
		return zero
	}
	return heap.Pop(h).(T)
}

func (h *GenericHeap[T]) First() T {
	if h.Len() == 0 {
		var zero T
		return zero
	}
	return h.data[0]
}
