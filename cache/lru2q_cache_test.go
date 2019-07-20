package cache

import (
	"fmt"
	"testing"
)

//测试点:
//set get del key
//热队列满了移到冷队列
//冷队列满了删除
//冷队列又被命中移到热队列, 然后热队列又满了进冷
//为了简单队列间的移动用string输出
func getkey(lrucache *LruCache, key string) bool {
	if val, ok := lrucache.Get(key); ok {
		if _, status := val.(int); status {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}
func Test_QueueMove(t *testing.T) {
	var hot_qlen uint32 = 8
	var cold_qlen uint32 = 2
	lrucache := New(hot_qlen, cold_qlen)
	vals := make([]int, 10, 10)
	for i := 0; i < 10; i++ {
		lrucache.Set(fmt.Sprintf("%d", i), i)
		vals = append(vals, i)
	}
	t.Log(lrucache.ToString())
	lrucache.get("8")
	t.Log(lrucache.ToString())
	addlen := 10 + int(cold_qlen)
	for i := 10; i < addlen; i++ {
		lrucache.Set(fmt.Sprintf("%d", i), i)
	}
	t.Log(lrucache.ToString())
	for i := 3; i < 5; i++ {
		if getkey(lrucache, fmt.Sprintf("%d", i)) {
			t.Log("lru get key successful")
		} else {
			t.Error("lru get key failed")
		}
	}
	t.Log(lrucache.ToString())
	lrucache.Del("8")
	_, ok := lrucache.Get("8")
	if ok {
		t.Error("lru del key failed")
	}
	for i := addlen + 1; i < addlen+3; i++ {
		lrucache.Set(fmt.Sprintf("%d", i), i)
	}
	t.Log(lrucache.ToString())
	lrucache.Reset()
	if lrucache.cold_queue.Len() != 0 || lrucache.hot_queue.Len() != 0 {
		t.Error("lru reset failed")
	}
}
