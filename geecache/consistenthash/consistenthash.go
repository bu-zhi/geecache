package consistenthash
import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct{
	hash Hash//哈希函数
	replicas int//一个真实节点对应几个虚拟节点
	rings []int//哈希环
	hashMap map[int]string//虚拟节点对应的真实节点
}
func New(replicas int,fn Hash)*Map{
	m :=&Map{
		replicas: replicas,
		hash: fn,
		hashMap: map[int]string{},
	}
	if m.hash==nil{
		m.hash=crc32.ChecksumIEEE
	}
	return m
}
//向哈希环添加节点
func (m *Map) Add(keys ...string){
	for _,key := range keys{
		for i:=0;i<m.replicas;i++{
			hash :=int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.rings =append(m.rings, hash)
			m.hashMap[hash]=key
		}
	}
	sort.Ints(m.rings)//切片由小到大排序
}
//得到报存当前key的节点
func (m *Map)Get(key string)string{
	if len(m.rings)==0{
		return ""
	}
	hash :=int(m.hash([]byte(key)))
	idx:=sort.Search(len(m.rings),func(i int) bool {//二分查找法
		return m.rings[i]>=hash
	})
	if idx==len(m.rings){
		idx=0
	}
	return m.hashMap[m.rings[idx]]
}
