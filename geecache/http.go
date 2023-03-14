package geecache

import (
	"fmt"
	"geecache/geecache/consistenthash"
	pb "geecache/geecache/geecachepb"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"google.golang.org/protobuf/proto"
)

//定义一个地址作为节点间通讯前缀，以免其它服务干扰
const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)
type HTTPPool struct{
	self string//用来记录自己的相关信息，ip+端口
	basePath string//节点后默认后缀
	mu sync.Mutex
	peers *consistenthash.Map//一致性哈希
	httpGetters map[string]*httpGetter//不同的远程节点用不同的httpGetter表示
}

func NewHTTPPool(self string)*HTTPPool{
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}
//日志
func (p *HTTPPool)Log(format string,v ...interface{}){
	log.Printf("[server %s] %s",p.self,fmt.Sprintf(format,v...))
}
//服务端http处理请求，请求链接默认/defaultBasepath/groupname/key
func (p *HTTPPool)ServeHTTP(w http.ResponseWriter,r *http.Request){
	if !strings.HasPrefix(r.URL.Path,p.basePath){//判断初始路径是否正确
		panic("httppol serving unexpected path:"+r.URL.Path)
	}
	p.Log("%s %s",r.Method,r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(p.basePath):],"/",2)
	if len(parts)!=2{
		http.Error(w,"bad request",http.StatusBadRequest)//400，请求无效，客户端无法处理
		return
	}
	groupname :=parts[0]
	key := parts[1]
	group:=GetGroup(groupname)
	if group==nil{//不存在该groupname
		http.Error(w,"no such group:"+groupname,http.StatusNotFound)//404请求的资源不存在
		return
	}
	view,err:=group.Get(key)
	if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)//500，服务器内部错误
		return
	}
	body,err:=proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err!=nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(body)

}

//客户端
type httpGetter struct{
	baseURL string//baseurl存储要访问的IP地址，例：http://x.x.x.x/_geecache
}
//根据group和key来从服务端获取缓存值
func (h *httpGetter)Get(in *pb.Request,out *pb.Response)error{
	u:=fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)//拼接完整的url
	res,err:=http.Get(u)
	if err!=nil{
		return err
	}
	defer res.Body.Close()
	if res.StatusCode!=http.StatusOK{
		return fmt.Errorf("server returned: %v",res.Status)

	}
	bytes,err:=ioutil.ReadAll(res.Body)
	if err!=nil{
		return fmt.Errorf("reading response body: %v",err)
	}
	if err = proto.Unmarshal(bytes,out);err!=nil{
		return fmt.Errorf("decoding response body: %v",err)
	}
	return nil
}
var _ PeerGetter = (*httpGetter)(nil)//这是 Go 语言的一个编译时机制，用于确保 httpGetter 类型确实实现了 PeerGetter 接口，如果没有实现，程序将在编译时报错。

//向节点池中设置节点，初始化map，向哈希环添加节点，并保存节点
func (p*HTTPPool) Set (peers ...string){
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers=consistenthash.New(defaultReplicas,nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter,len(peers))
	for _,peer:=range peers{
		p.httpGetters[peer]=&httpGetter{baseURL: peer+p.basePath}
	}
}
//根据key选择节点
func (p *HTTPPool)PickPeer(key string)(PeerGetter,bool){
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer:=p.peers.Get(key);peer!="" && peer != p.self{
		p.Log("Pick peer %s",peer)
		return p.httpGetters[peer],true
	}
	return nil,false
}
var _ PeerPicker = (*HTTPPool)(nil)
