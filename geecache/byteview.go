package geecache
//缓存值的抽象与封装
type ByteView struct{
	b []byte
}
func (v ByteView) Len() int{//Value接口中用Len方法
	return len(v.b)
}
func (v ByteView)ByteSlice()[]byte{//深拷贝一份，防止修改到缓存
	return copyslice(v.b)
}
func copyslice(b []byte)[]byte{//深拷贝
	aslice:=make([]byte,len(b))
	copy(aslice,b)
	return aslice
}
func (v ByteView)String()string{//转为字符串
	return string(v.b)
}
