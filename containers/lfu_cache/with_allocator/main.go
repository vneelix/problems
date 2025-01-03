package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Node double linked list with generic value
type Node[T any] struct {
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}

// Allocator fixed size nodes allocator
type Allocator[T any] struct {
	items     []Node[T]
	available *Node[T]
}

// NewAllocator Pre-allocate slice for nodes with capacity
func NewAllocator[T any](capacity int) *Allocator[T] {
	allocator := &Allocator[T]{
		items: make([]Node[T], 1, capacity),
	}
	allocator.available = &allocator.items[0]
	return allocator
}

// Get returns available node from pre-allocated slice
func (allocator *Allocator[T]) Get() *Node[T] {
	if allocator.available != nil {
		node := allocator.available
		allocator.available = allocator.available.Next
		node.Next = nil
		return node
	}
	if len(allocator.items) == cap(allocator.items) {
		panic("buffer overflow")
	}
	allocator.items = allocator.items[:len(allocator.items)+1]
	return &allocator.items[len(allocator.items)-1]
}

// Put returns node into slice with nodes
func (allocator *Allocator[T]) Put(node *Node[T]) {
	*node = Node[T]{Next: allocator.available}
	allocator.available = node
}

// List linked container with fixed size
type List[T any] struct {
	head      *Node[T]
	tail      *Node[T]
	allocator *Allocator[T]
}

// NewList create container with fixed size allocator
func NewList[T any](capacity int) *List[T] {
	return &List[T]{
		allocator: NewAllocator[T](capacity),
	}
}

// Get returns node from allocator
func (list *List[T]) Get(value T) *Node[T] {
	node := list.allocator.Get()
	node.Value = value
	return node
}

// Put returns node into allocator
func (list *List[T]) Put(node *Node[T]) {
	list.allocator.Put(node)
}

func (list *List[T]) Head() *Node[T] {
	return list.head
}

func (list *List[T]) Tail() *Node[T] {
	return list.tail
}

func (list *List[T]) PushFront(node *Node[T]) {
	if list.head == nil {
		list.head, list.tail = node, node
		return
	}
	list.head = list.InsertBefore(list.head, node)
}

func (list *List[T]) PushBack(node *Node[T]) {
	if list.tail == nil {
		list.head, list.tail = node, node
		return
	}
	list.tail = list.InsertAfter(list.tail, node)
}

func (list *List[T]) InsertBefore(target *Node[T], node *Node[T]) *Node[T] {
	node.Next, node.Prev = target, target.Prev
	if target.Prev != nil {
		target.Prev.Next = node
	}
	target.Prev = node
	return node
}

func (list *List[T]) InsertAfter(target *Node[T], node *Node[T]) *Node[T] {
	node.Next, node.Prev = target.Next, target
	if target.Next != nil {
		target.Next.Prev = node
	}
	target.Next = node
	return node
}

func (list *List[T]) Remove(node *Node[T]) *Node[T] {
	if node.Prev != nil {
		node.Prev.Next = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}
	if node == list.head {
		list.head = node.Next
	}
	if node == list.tail {
		list.tail = node.Prev
	}
	node.Next, node.Prev = nil, nil
	return node
}

type Rank struct {
	rank  int
	items List[Item]
}

type Item struct {
	Rank  *Node[Rank]
	Key   int
	Value int
}

type LFUCache struct {
	cap             int
	ranks           *List[Rank]
	items           *List[Item]
	itemsDictionary map[int]*Node[Item]
}

func Constructor(capacity int) LFUCache {
	return LFUCache{
		cap:             capacity,
		ranks:           NewList[Rank](capacity),
		items:           NewList[Item](capacity),
		itemsDictionary: make(map[int]*Node[Item], capacity),
	}
}

func (cache *LFUCache) update(key, value int) {
	itemNode := cache.itemsDictionary[key]
	rankNode := itemNode.Value.Rank

	var newRankNode *Node[Rank]
	rankNode.Value.items.Remove(itemNode)
	if rankNode.Next == nil || rankNode.Value.rank+1 != rankNode.Next.Value.rank {

		if rankNode.Value.items.Head() == nil {
			rankNode.Value.rank++
			newRankNode = rankNode
		} else {
			newRankNode = cache.ranks.InsertAfter(rankNode, cache.ranks.Get(Rank{rank: rankNode.Value.rank + 1}))
		}

	} else {
		newRankNode = rankNode.Next
		if rankNode.Value.items.Head() == nil {
			cache.ranks.Put(cache.ranks.Remove(rankNode))
		}
	}

	itemNode.Value.Rank, itemNode.Value.Value = newRankNode, value
	newRankNode.Value.items.PushFront(itemNode)
}

func (cache *LFUCache) Get(key int) int {
	if itemNode, exists := cache.itemsDictionary[key]; exists {
		cache.update(itemNode.Value.Key, itemNode.Value.Value)
		return itemNode.Value.Value
	}
	return -1
}

func (cache *LFUCache) Put(key int, value int) {
	if _, exists := cache.itemsDictionary[key]; !exists {

		// evict
		if len(cache.itemsDictionary) == cache.cap {
			itemNodeToEvict := cache.ranks.Head().Value.items.Tail()
			keyToEvict := itemNodeToEvict.Value.Key
			cache.items.Put(cache.ranks.Head().Value.items.Remove(itemNodeToEvict))
			if cache.ranks.Head().Value.items.Head() == nil {
				cache.ranks.Put(cache.ranks.Remove(cache.ranks.Head()))
			}
			delete(cache.itemsDictionary, keyToEvict)
		}

		lowestRankNode := cache.ranks.Head()
		if lowestRankNode == nil || lowestRankNode.Value.rank != 1 {
			cache.ranks.PushFront(
				cache.ranks.Get(Rank{rank: 1}),
			)
		}

		cache.ranks.Head().Value.items.PushFront(
			cache.items.Get(Item{
				Rank:  cache.ranks.Head(),
				Key:   key,
				Value: value,
			}),
		)
		cache.itemsDictionary[key] = cache.ranks.Head().Value.items.Head()

	} else {
		cache.update(key, value)
	}
}

func DebugCache(cache *LFUCache) {
	fmt.Println(strings.Repeat("-", 16))
	for rank := cache.ranks.Head(); rank != nil; rank = rank.Next {
		items := make([]string, 0, 16)
		items = append(items, strconv.Itoa(rank.Value.rank))
		for item := rank.Value.items.Head(); item != nil; item = item.Next {
			items = append(items, fmt.Sprint(item.Value))
		}
		fmt.Println(strings.Join(items, "\t"), " <- ", rank.Value.items.Tail().Value)
	}
	fmt.Println(strings.Repeat("-", 16))
}

func RunTests(commands []string, arguments [][]int) {
	var cache LFUCache
	for idx, command := range commands {
		switch command {
		case "LFUCache":
			cache = Constructor(arguments[idx][0])
		case "get":
			fmt.Println("get", arguments[idx][0])
			cache.Get(arguments[idx][0])
			DebugCache(&cache)
		case "put":
			fmt.Println("put", arguments[idx][0], arguments[idx][1])
			cache.Put(arguments[idx][0], arguments[idx][1])
			DebugCache(&cache)
		}
	}
}

func main() {

	RunTests(
		[]string{"LFUCache", "put", "put", "get", "put", "get", "get", "put", "get", "get", "get"},
		[][]int{{2}, {1, 1}, {2, 2}, {1}, {3, 3}, {2}, {3}, {4, 4}, {1}, {3}, {4}},
	)
}
