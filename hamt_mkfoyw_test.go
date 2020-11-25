package hamt_test

import (
	"fmt"
	"math/big"

	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Defaults

const buketSize = 3
const defaultBitWidth = 8

// 指定的 key 没有找到
var ErrNotFound = fmt.Errorf("not foudn")

// 超出最大的访问深度
var ErrMaxDepth = fmt.Errorf("attempted to traverse HAMT boyond max_depth")

// 序列化节点错误
var ErrMalfomedHamt = fmt.Errorf("Hamt node was malfored")

type Node struct {
	Bitfield *big.Int
	Pointers []*Pointer

	bitWidth int
	hash     func([]byte) []byte

	store cbor.IpldStore
}

type Pointer struct {
	KVS  []*KV   `refmt:"v,omitempty"`
	Link cid.Cid `refmt:"l, omitempty"`

	cache *Node
}

//该数据结构在 IPLD 数据库中表示为如下结构：
//
//		type KV struct {
//			key Bytes
//			value Any
//		} representation tuple
type KV struct {
	Key   []byte
	Value *cbg.Deferred
}

//函数用于配置Node
type Option func(*Node)

func UseTreeBitWidth(bitWidth int) Option {
	return func(nd *Node) {
		if bitWidth > 0 && bitWidth <= 8 {
			nd.bitWidth = bitWidth
		}
	}
}

func UseHanshFunc(hash func([]byte) []byte) Option {
	return func(nd *Node) {
		nd.hash = hash
	}
}
