package singleflight
//当有一个key请求发起后 fn处理函数处理过程中 如果有其他相同key的请求过来 那么就一直阻塞等待第一个fn处理完成 拿到完成后的值。
import "sync"
//call表示正在进行中或已经结束的请求
type call struct{
	wg sync.WaitGroup
	val interface{}
	err error
}
//主数据结构，管理不同key的请求——call
type Group struct{
	mu sync.Mutex
	m map[string]*call
}
//call与group存在时间很短，该函数作用只是防止并发过高
func (g *Group)Do(key string,fn func()(interface{},error))(interface{},error){
	g.mu.Lock()
	if g.m == nil{
		g.m = make(map[string]*call)
	}
	if c,ok:=g.m[key];ok{
		g.mu.Unlock()
		c.wg.Wait()//防止数据还没存进去
		return c.val,c.err
	}
	c:=new(call)
	c.wg.Add(1)
	g.m[key]=c
	g.mu.Unlock()

	c.val,c.err=fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.m,key)//更新g.m
	g.mu.Unlock()
	return c.val,c.err
}