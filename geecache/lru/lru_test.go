package lru
import (
	"testing"
)
type String string

func (d String) Len() int {
	return len(d)
}
// func TestFind(t *testing.T){
// 	lru :=NewCache(int64(100),nil)
// 	lru.AddorChange("key1",String("123"))
// 	if v ,_:= lru.Find("key1");string(v.(String)) != "1234" {
// 		t.Fatalf("cache hit key1=1234 failed")
// 	}
// }

func TestDelete(t *testing.T){
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	//length := len(k1 + k2 + v1 + v2)
	lru := NewCache(int64(20), nil)
	lru.AddorChange(k1, String(v1))
	lru.AddorChange(k2, String(v2))
	lru.AddorChange(k3, String(v3))
	if _, ok := lru.Find("k1"); ok || len(lru.cache)!= 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}