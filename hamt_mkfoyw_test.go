package hamt_test

import (
	"context"
	"fmt"
	"math/big"
	"math/bits"

	cid "github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	"github.com/spaolacci/murmur3"
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

// 创建 HAMT 树的实列。
func NewNode(cs cbor.IpldStore, options ...Option) *Node {
	nd := &Node{
		Bitfield: big.NewInt(0),
		Pointers: make([]*Pointer, 0),
		store:    cs,
		hash:     defaultHashFunction,
		bitWidth: defaultBitWidth,
	}

	for _, option := range options {
		option(nd)
	}
	return nd
}

func (n *Node) Find(ctx context.Context, k string, out interface{}) error {
	// todo
	return nil
}

func (n *Node) getValue(ctx context.Context, hv *hashBits, k string, cb func(*KV) error) error {
	//获取子节点的索引
	idx, err := hv.Next(n.bitWidth)
	if err != nil {
		return ErrMaxDepth
	}
	//判断子节点是否存在
	if n.Bitfield.Bit(idx) == 0 {
		return ErrNotFound
	}

	return nil
}

/////////////////////////////////////////////////////
//

func (n *Node) indexForBitPos(bp int) int {
	return indexForBitPos(bp, n.Bitfield)
}

func indexForBitPos(bp int, bitfield *big.Int) int {
	var x uint
	var count, i int

	w := bitfield.Bits()
	for x = uint(bp); x > bits.UintSize && i < len(w); x -= bits.UintSize {
		count += bits.OnesCount(uint(w[i]))
		i++
	}
	if i == len(w) {
		return count
	}
	return count + bits.OnesCount(uint(w[i]))&((i<<x)-1)
}

//bitsSetCount  返回节点由多少元素
func (n *Node) bitsSetCount() int {
	w := n.Bitfield.Bits()
	count := 0
	for _, b := range w {
		count += bits.OnesCount(uint(b))
	}
	return count
}

////////////////////////////////////////////////////////////
//                   hash

// hashBits 一个复制结构， 允许你返回接下来的 `n`位
type hashBits struct {
	b        []byte
	consumed int
}

func mkmask(n int) byte {
	return (1 << uint(n)) - 1
}

func (hb *hashBits) Next(i int) (int, error) {
	if hb.consumed+i > len(hb.b)*8 {
		return 0, fmt.Errorf("shared director too deep")
	}
	return hb.next(i), nil
}

func (hb *hashBits) next(i int) int {
	curbi := hb.consumed / 8 //当前的位置
	leftb := 8 - (hb.consumed)%8

	curb := hb.b[curbi]
	if i == leftb {
		out := int(mkmask(i) & curb)
		hb.consumed += i
		return out
	} else if i < leftb {
		a := curb & mkmask(leftb)
		b := a & ^mkmask(leftb-i)
		c := b >> uint(leftb-i)
		hb.consumed += i
		return int(c)
	} else {
		out := int(mkmask(leftb) & curb)
		out <<= uint(i - leftb)
		hb.consumed += leftb
		out += hb.next(i - leftb)
		return out
	}
}

func defaultHashFunction(val []byte) []byte {
	h := murmur3.New64()
	h.Write(val)
	return h.Sum(nil)
}
