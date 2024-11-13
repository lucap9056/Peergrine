package genericheap_test

import (
	GenericHeap "peergrine/utils/generic-heap"
	"testing"
)

func TestgenericHeap(t *testing.T) {
	lessFunc := func(a, b int) bool {
		return a < b
	}
	h := GenericHeap.New(lessFunc)

	// 添加元素
	h.Add(3)
	h.Add(1)
	h.Add(4)
	h.Add(1)
	h.Add(5)
	h.Add(9)

	// 測試堆頂元素
	if h.First() != 1 {
		t.Fatalf("Expected top element to be 1, but got %d", h.First())
	}

	// 測試移除元素
	if removed := h.Remove(); removed != 1 {
		t.Fatalf("Expected removed element to be 1, but got %d", removed)
	}
	if h.First() != 1 {
		t.Fatalf("Expected top element to be 1, but got %d", h.First())
	}

	// 測試空堆處理
	h.Remove() // Remove another 1
	h.Remove() // Remove 3
	h.Remove() // Remove 4
	h.Remove() // Remove 5
	if h.First() != 9 {
		t.Fatalf("Expected top element to be 9, but got %d", h.First())
	}

	h.Remove() // Remove last element
	if h.First() != 0 {
		t.Fatalf("Expected top element to be zero value, but got %d", h.First())
	}
}
