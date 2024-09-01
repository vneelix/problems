package lfu_cache

import (
	doubleLinkedList "container/list"
	"problems/containers/ordered_map"
)

type LFUCache struct {
	capacity int
	ranks    *doubleLinkedList.List
	keyRank  map[int]int
	rankNode map[int]*doubleLinkedList.Element
}

func New(capacity int) LFUCache {
	return LFUCache{
		capacity: capacity,
		ranks:    doubleLinkedList.New(),
		keyRank:  map[int]int{},
		rankNode: map[int]*doubleLinkedList.Element{},
	}
}

// updateRank move key from current rank to rank + 1 table
func (cache *LFUCache) updateRank(key, value int) {
	rank, _ := cache.keyRank[key]

	// increase rank for key/value pair
	currentRankNode, _ := cache.rankNode[rank]
	nextRankNode, exists := cache.rankNode[rank+1]
	if !exists {
		// create rank + 1 table
		nextRankNode = cache.ranks.InsertAfter(ordered_map.New[int, int](), currentRankNode)
		cache.rankNode[rank+1] = nextRankNode
	}

	// remove key/value pair from current rank
	currentRank := currentRankNode.Value.(*ordered_map.OrderedMap[int, int])
	currentRank.Delete(key)
	if currentRank.Length() == 0 {
		// if current rank table become empty - delete it
		cache.ranks.Remove(currentRankNode)
		delete(cache.rankNode, rank)
	}

	// insert key/value into rank + 1 table
	nextRankNode.Value.(*ordered_map.OrderedMap[int, int]).Insert(key, value)

	cache.keyRank[key] = rank + 1
}

func (cache *LFUCache) Get(key int) int {
	if rank, exists := cache.keyRank[key]; !exists {
		return -1
	} else {
		value, _ := cache.rankNode[rank].Value.(*ordered_map.OrderedMap[int, int]).Get(key)
		cache.updateRank(key, value)
		return value
	}
}

func (cache *LFUCache) Put(key int, value int) {
	if _, exists := cache.keyRank[key]; !exists {
		// if current key doesn't have rank - it's new key

		// if capacity limit reached - delete the least used key with the smallest rank
		if cache.capacity == len(cache.keyRank) {
			lowestRankNode := cache.ranks.Front()
			lowestRank := lowestRankNode.Value.(*ordered_map.OrderedMap[int, int])

			k, _ := lowestRank.Keys()()
			lowestRank.Delete(k)
			if lowestRank.Length() == 0 {
				cache.ranks.Remove(lowestRankNode)
				delete(cache.rankNode, cache.keyRank[k])
			}
			delete(cache.keyRank, k)
		}

		// insert key/value into rank-1
		node, exists := cache.rankNode[1]
		if !exists {
			node = cache.ranks.PushFront(ordered_map.New[int, int]())
			cache.rankNode[1] = node
		}
		node.Value.(*ordered_map.OrderedMap[int, int]).Insert(key, value)
		cache.keyRank[key] = 1
		return

	} else {
		// key exists - increase rank
		cache.updateRank(key, value)
	}
}
