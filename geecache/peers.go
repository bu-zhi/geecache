package geecache
import (
	pb "geecache/geecache/geecachepb"
)
//根据传入的key来选择相应节点
type PeerPicker interface{
	PickPeer(key string) (peer PeerGetter,ok bool)
}

//利用Group+key来组成规定url获取value
type PeerGetter interface{
	Get(in *pb.Request,out *pb.Response) error
}