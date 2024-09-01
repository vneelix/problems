package ordered_map

import (
	doubleLinkedList "container/list"
)

type OrderedMap[kT comparable, vT any] struct {
	kv      map[kT]vT
	queue   *doubleLinkedList.List
	keyNode map[kT]*doubleLinkedList.Element
}

func New[kT comparable, kV any]() *OrderedMap[kT, kV] {
	return &OrderedMap[kT, kV]{
		kv:      make(map[kT]kV),
		queue:   doubleLinkedList.New(),
		keyNode: make(map[kT]*doubleLinkedList.Element),
	}
}

func (m *OrderedMap[kT, kV]) Insert(key kT, value kV) {
	if _, exists := m.kv[key]; !exists {
		m.keyNode[key] = m.queue.PushBack(key)
	}
	m.kv[key] = value
}

func (m *OrderedMap[kT, kV]) Delete(key kT) {
	if _, exists := m.kv[key]; !exists {
		return
	}
	delete(m.kv, key)
	m.queue.Remove(m.keyNode[key])
	delete(m.keyNode, key)
}

func (m *OrderedMap[kT, kV]) Get(key kT) (value kV, exists bool) {
	value, exists = m.kv[key]
	return
}

func (m *OrderedMap[_, _]) Length() int {
	return len(m.kv)
}

func (m *OrderedMap[kT, _]) Keys() func() (key kT, ok bool) {
	node := m.queue.Front()
	return func() (key kT, ok bool) {
		if node != nil {
			key, ok = node.Value.(kT), true
			node = node.Next()
		}
		return
	}
}
