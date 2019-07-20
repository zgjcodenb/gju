package cache

import (
	"container/list"
	"fmt"
	"sync"
)

type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, v interface{})
	Del(k string) bool
}

type lruNode struct {
	k    string
	v    interface{}
	addr *list.List
}

//引用计数和冷热队列的方式实现lru
type LruCache struct {
	table           sync.Map
	hot_queue       *list.List
	cold_queue      *list.List
	use_size        uint32
	hot_queue_size  uint32
	cold_queue_size uint32
	mu              sync.Mutex
}

func New(hot_size, cold_size uint32) *LruCache {
	return &LruCache{
		hot_queue:       list.New(),
		cold_queue:      list.New(),
		use_size:        0,
		hot_queue_size:  hot_size,
		cold_queue_size: cold_size,
	}
}

func (this *LruCache) Size() uint32 {
	return this.use_size
}

func (this *LruCache) MoveToCold(e *list.Element) {
	node := e.Value.(*lruNode)
	if node.addr != this.hot_queue {
		panic("shoud be in hot_queue when MoveToCold!")
	}
	this.hot_queue.Remove(e)
	node.addr = this.cold_queue
	cold_e := this.cold_queue.PushFront(node)
	//会判断所属的list很坑
	this.table.Store(node.k, cold_e)
	if uint32(this.cold_queue.Len()) > this.cold_queue_size {
		ele := this.cold_queue.Back()
		this.cold_queue.Remove(ele)
		del_node := ele.Value.(*lruNode)
		this.table.Delete(del_node.k)
		this.use_size--
	}
}

func (this *LruCache) ref(e *list.Element) *lruNode {
	if e == nil {
		return nil
	}
	node := e.Value.(*lruNode)
	if node.addr == this.cold_queue {
		this.cold_queue.Remove(e)
		node.addr = this.hot_queue
		this.hot_queue.PushBack(node)
		//this.table.Store(node.k, hot_e)
	} else if node.addr == this.hot_queue {
		this.hot_queue.MoveToFront(e)
	} else {
		panic("data not in hot or cold queue!")
	}
	return node
}

func (this *LruCache) add_node(k string, v interface{}) {
	new_node := &lruNode{
		k:    k,
		v:    v,
		addr: this.hot_queue,
	}
	ele := this.hot_queue.PushFront(new_node)
	this.table.Store(k, ele)
	this.use_size++
	return
}

func (this *LruCache) set(k string, v interface{}) {

	this.add_node(k, v)
	this.check()
	return
}

func (this *LruCache) Set(k string, v interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.set(k, v)
}

func (this *LruCache) get(k string) (interface{}, bool) {
	if v, ok := this.table.Load(k); ok {
		ele := v.(*list.Element)
		node_value := this.ref(ele)
		if node_value != nil {
			return node_value.v, true
		} else {
			return nil, false
		}
	}
	return nil, false

}

func (this *LruCache) Get(k string) (interface{}, bool) {
	this.mu.Lock()
	v, status := this.get(k)
	this.mu.Unlock()
	return v, status
}

func (this *LruCache) RemoveFromLru(ele *list.Element) {
	this.hot_queue.Remove(ele)
	this.cold_queue.Remove(ele)
	return
}

func (this *LruCache) Reset() {
	this.table.Range(func(key interface{}, value interface{}) bool {
		this.table.Delete(key)
		return true
	})
	this.cold_queue = list.New()
	this.hot_queue = list.New()
	this.use_size = 0
	return
}

func (this *LruCache) del(k string) bool {
	if v, ok := this.table.Load(k); ok {
		ele := v.(*list.Element)
		this.RemoveFromLru(ele)
		this.table.Delete(k)
		this.use_size--
		return true
	}
	return false
}

func (this *LruCache) Del(k string) bool {
	this.mu.Lock()
	ret := this.del(k)
	this.mu.Unlock()
	return ret
}

func (this *LruCache) check() {
	if this.use_size > this.hot_queue_size {
		var ele *list.Element = nil
		if this.hot_queue.Len() != 0 {
			ele = this.hot_queue.Back()
			this.MoveToCold(ele)
		}
	}
	return
}

func (this *LruCache) ToString() string {
	var str string

	str += fmt.Sprintf("hot queue: ")
	for e := this.hot_queue.Front(); e != nil; e = e.Next() {
		node := e.Value.(*lruNode)
		str += fmt.Sprintf("%v, ", node.k)
	}
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("cold queue: ")
	for e := this.cold_queue.Front(); e != nil; e = e.Next() {
		node := e.Value.(*lruNode)
		str += fmt.Sprintf("%v, ", node.k)
	}
	str += fmt.Sprintf("\n")
	str += fmt.Sprintf("usesize:%d, hotsize:%d, coldsize:%d\n", this.use_size, this.hot_queue.Len(),
		this.cold_queue.Len())
	return str
}
