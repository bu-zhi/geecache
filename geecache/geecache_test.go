package geecache
import (
	"testing"
	"fmt"
)
var db = map[string]string{
	"Tom": "630",
	"Jack": "589",
	"Sam": "567",
}

func TestGet(t *testing.T){
	loadCounts:=make(map[string]int,len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k,v :=range db{
		if view,err:=gee.Get(k);err!=nil || view.String()!=v{
			t.Fatalf("failed to get value of %s",k)
		}
		if _,err:=gee.Get(k);err!=nil || loadCounts[k]>1{
			t.Fatalf("cache miss")
		}
	}
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}