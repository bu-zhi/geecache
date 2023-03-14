package geecache

import (
	"fmt"
	"geecache/geecache/singleflight"
	"log"
	"sync"
	pb "geecache/geecache/geecachepb"
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
	getter Getter//回调函数
	maincache cache//保存缓存值
	peers PeerPicker //保存了一个httpPool/grpcpool结构体
	loader *singleflight.Group//防止并发过大
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
		loader: &singleflight.Group{},
	}
	groups[name]=g
	return g
}
//根据name来获取对应group
func GetGroup(name string)*Group{
	mu.RLock()
	defer mu.RUnlock()
	g :=groups[name]
	return g
}
//获取缓存，不存在就从其它地方加载
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
//根据key来获取对应节点，再从对应节点获取value，失败则从本地加载
func (g *Group)load(key string)(value ByteView,err error){
	viewi,err := g.loader.Do(key,func()(interface{}, error){
		if g.peers!=nil{
			if peer,ok:=g.peers.PickPeer(key);ok{
				if value,err = g.getFrompeer(peer,key);err ==nil{
					return value,nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getlocally(key)
	})

	if err==nil{
		return viewi.(ByteView),nil
	}
	return
}
//用节点加group加key来获取对应value
func (g *Group) getFrompeer(peer PeerGetter,key string)(ByteView,error){
	req := &pb.Request{
		Group: g.name,
		Key: key,
	}
	res :=&pb.Response{}
	err := peer.Get(req,res)

	if err !=nil{
		return ByteView{},err
	}
	return ByteView{b:res.Value},nil
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
func (g *Group)RegisterPeers(peers PeerPicker){
	if g.peers!=nil{
		panic("RegisterPeerPicker called more than once")
	}
	g.peers=peers
}