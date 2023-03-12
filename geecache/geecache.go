package geecache

import (
	"fmt"
	"log"
	"sync"
)
//回调函数
type Getter interface{
	Get(key string)([]byte,error)
}
type GetterFunc func(key string)([]byte,error)
func (f GetterFunc)Get(key string)([]byte,error){
	return f(key)
}
//每个缓存空间有其对应的name
type Group struct{
	name string
	getter Getter
	maincache cache
}
var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)
//新建一个group
func NewGroup(name string,cachebytes int64,getter Getter) *Group{
	mu.Lock()
	defer mu.Unlock()
	g :=&Group{
		name: name,
		getter: getter,
		maincache: cache{cachebytes:cachebytes},
	}
	groups[name]=g
	return g
}
//根据name来获取对应group
func GetGroup(name string)*Group{
	mu.RLock()
	defer mu.Unlock()
	g :=groups[name]
	return g
}
//获取函数，不存在就从其它地方加载
func (g *Group)Get(key string)(ByteView,error){
	if key==""{
		return ByteView{},fmt.Errorf("need a key")
	}
	if v,ok:=g.maincache.Get(key);ok{
		log.Println("geecache hit")
		return v,nil
	}
	return g.load(key)
}

func (g *Group)load(key string)(ByteView,error){
	return g.getlocally(key)
}
//利用回调函数来加载，并将其放入缓存
func (g *Group)getlocally(key string)(ByteView,error){
	bytes,err:=g.getter.Get(key)
	if err!=nil{
		return ByteView{},err
	}
	value := ByteView{b:copyslice(bytes)}
	g.populatecache(key,value)
	return value,nil
}
//放入缓存
func (g *Group)populatecache(key string,value ByteView){
	g.maincache.Add(key,value)
}