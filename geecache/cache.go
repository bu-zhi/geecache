package geecache
import (
	"geecache/geecache/lru"
	"sync"
)
//实例化lru,并发控制
type cache struct{
	mu sync.Mutex
	lru *lru.Cache
	cachebytes int64
}
//添加，无cache则新建
func (c *cache)Add(key string,value ByteView){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		c.lru=lru.NewCache(c.cachebytes,nil)
	}
	c.lru.AddorChange(key,value)
}
//获取
func (c *cache)Get(key string)(value ByteView,ok bool){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru==nil{
		return 
	}

	if value,ok:=c.lru.Find(key);ok{
		return value.(ByteView),true
	}
	return
}