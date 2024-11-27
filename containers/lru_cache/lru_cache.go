package lru_cache

// Node double linked list
type Node[T any] struct {
	Prev  *Node[T]
	Next  *Node[T]
	Value T
}

// Remove unlink node from list
func (node *Node[T]) Remove() {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}
	node.Prev, node.Next = nil, nil
}

func (node *Node[T]) InsertBefore(newNode *Node[T]) {
	newNode.Next = node
	newNode.Prev = node.Prev
	if newNode.Prev != nil {
		newNode.Prev.Next = newNode
	}
	node.Prev = newNode
}

func (node *Node[T]) InsertAfter(newNode *Node[T]) {
	newNode.Prev = node
	newNode.Next = node.Next
	if newNode.Next != nil {
		newNode.Next.Prev = newNode
	}
	node.Next = newNode
}

// NodeAllocator static allocator with buffer
type NodeAllocator[T any] struct {
	pool          []Node[T]
	availableNode *Node[T]
}

func NewNodeAllocator[T any](capacity int) *NodeAllocator[T] {
	allocator := &NodeAllocator[T]{
		pool: make([]Node[T], 1, capacity),
	}
	allocator.availableNode = &(allocator.pool[0])
	return allocator
}

func (allocator *NodeAllocator[T]) Allocate() *Node[T] {
	if allocator.availableNode != nil {
		node := allocator.availableNode
		allocator.availableNode = allocator.availableNode.Next
		node.Next = nil
		return node
	}
	if len(allocator.pool) == cap(allocator.pool) {
		panic("overflow")
	}
	allocator.pool = allocator.pool[:len(allocator.pool)+1]
	return &allocator.pool[len(allocator.pool)-1]
}

func (allocator *NodeAllocator[T]) Release(node *Node[T]) {
	*node = Node[T]{Next: allocator.availableNode}
	allocator.availableNode = node
}

type KV[KT comparable, VT any] struct {
	K KT
	V VT
}

type LRUCache struct {
	allocator      *NodeAllocator[KV[int, int]]
	values         map[int]*Node[KV[int, int]]
	usagesListHead *Node[KV[int, int]]
	usagesListTail *Node[KV[int, int]]
	capacity       int
}

func Constructor(capacity int) LRUCache {
	return LRUCache{
		allocator: NewNodeAllocator[KV[int, int]](capacity),
		values:    make(map[int]*Node[KV[int, int]], capacity),
		capacity:  capacity,
	}
}

func (cache *LRUCache) refresh(node *Node[KV[int, int]]) {
	if node == cache.usagesListHead {
		return
	}
	if node == cache.usagesListTail {
		cache.usagesListTail = cache.usagesListTail.Prev
	}
	if node.Prev != nil || node.Next != nil {
		node.Remove()
	}
	if cache.usagesListHead != nil {
		cache.usagesListHead.InsertBefore(node)
	} else {
		cache.usagesListTail = node
	}
	cache.usagesListHead = node
}

func (cache *LRUCache) evict() {
	node := cache.usagesListTail
	cache.usagesListTail = cache.usagesListTail.Prev
	if cache.usagesListTail == nil {
		cache.usagesListHead = nil
	}
	node.Remove()
	delete(cache.values, node.Value.K)
	cache.allocator.Release(node)
}

func (cache *LRUCache) Get(key int) int {
	node, exists := cache.values[key]
	if !exists {
		return -1
	}
	cache.refresh(node)
	return node.Value.V
}

func (cache *LRUCache) Put(key int, value int) {
	node, exists := cache.values[key]
	if !exists {

		if len(cache.values) == cache.capacity {
			cache.evict()
		}

		node = cache.allocator.Allocate()
		cache.values[key] = node
	}
	node.Value = KV[int, int]{key, value}
	cache.refresh(node)
}
