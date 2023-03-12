package lru
//对缓存排序使用lru，队首为最近使用的
import (
	"container/list"
)

//主要数据结构
type Cache struct{
	maxbytes int64 //缓存总大小
	nowbytes int64 //已用大小
	list *list.List //双向链表
	cache map[string]*list.Element //存储缓存的key与对应value的链表指针
	OnEvicted func(key string,value Value) //回调函数
}
//存储key和value
type entry struct{
	key string
	value Value
}
//用于返回value所占链表大小
type Value interface{
	Len() int
}
//实现Value接口
func (c *Cache) Len() int{
	return c.list.Len()
}

//NEW函数
func NewCache(max int64,OnE func(key string,value Value)) *Cache{
	return &Cache{
		maxbytes: max,
		nowbytes: 0,
		list: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: OnE,
	}
}

//增or改
func (c *Cache)AddorChange(key string,value Value) {
	if ele,ok := c.cache[key];ok{//当key存在时，则改
		c.list.MoveToFront(ele)
		c.nowbytes=c.nowbytes-int64(ele.Value.(*entry).value.Len())+int64(value.Len())
		ele.Value.(*entry).value=value
	}else {//key不存在时则增
		ele := c.list.PushFront(&entry{key,value})
		c.cache[key]=ele
		c.nowbytes+=int64(len(key))+int64(value.Len())
	}
	for c.maxbytes<c.nowbytes{//已用大小大于最大时，删
		c.Del()
	}
}
//删
func (c *Cache)Del(){
	ele := c.list.Back()//删除队尾
	if ele!=nil{
		c.list.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache,kv.key)
		c.nowbytes-=int64(len(kv.key))+int64(kv.value.Len())
	}
}
//查
func (c *Cache)Find(key string) (Value,bool){
	if ele,ok := c.cache[key];ok{//存在
		c.list.MoveToFront(ele)
		return ele.Value.(*entry).value,true
	}
	return nil,false
}