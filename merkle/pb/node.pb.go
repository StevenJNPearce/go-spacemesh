// Code generated by protoc-gen-go. DO NOT EDIT.
// source: node.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	node.proto

It has these top-level messages:
	Node
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type NodeType int32

const (
	NodeType_branch    NodeType = 0
	NodeType_extension NodeType = 1
	NodeType_leaf      NodeType = 2
)

var NodeType_name = map[int32]string{
	0: "branch",
	1: "extension",
	2: "leaf",
}
var NodeType_value = map[string]int32{
	"branch":    0,
	"extension": 1,
	"leaf":      2,
}

func (x NodeType) String() string {
	return proto.EnumName(NodeType_name, int32(x))
}
func (NodeType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Node struct {
	NodeType NodeType `protobuf:"varint,1,opt,name=nodeType,enum=pb.NodeType" json:"nodeType,omitempty"`
	Entries  [][]byte `protobuf:"bytes,2,rep,name=entries,proto3" json:"entries,omitempty"`
	Path     string   `protobuf:"bytes,3,opt,name=path" json:"path,omitempty"`
	Value    []byte   `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *Node) Reset()                    { *m = Node{} }
func (m *Node) String() string            { return proto.CompactTextString(m) }
func (*Node) ProtoMessage()               {}
func (*Node) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Node) GetNodeType() NodeType {
	if m != nil {
		return m.NodeType
	}
	return NodeType_branch
}

func (m *Node) GetEntries() [][]byte {
	if m != nil {
		return m.Entries
	}
	return nil
}

func (m *Node) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Node) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func init() {
	proto.RegisterType((*Node)(nil), "pb.Node")
	proto.RegisterEnum("pb.NodeType", NodeType_name, NodeType_value)
}

func init() { proto.RegisterFile("node.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 182 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8e, 0xb1, 0x6a, 0xc3, 0x30,
	0x14, 0x45, 0x2b, 0x59, 0x75, 0xed, 0x87, 0x5b, 0xcc, 0xa3, 0x83, 0x46, 0xd1, 0x49, 0x74, 0x70,
	0xa1, 0xf9, 0x83, 0x7c, 0x80, 0x07, 0x91, 0x29, 0x9b, 0x14, 0xbf, 0x60, 0x83, 0x91, 0x84, 0xad,
	0x84, 0xe4, 0xef, 0x43, 0x9c, 0x38, 0xdb, 0x3d, 0xf7, 0xc0, 0xe5, 0x02, 0xf8, 0xd0, 0x51, 0x13,
	0xa7, 0x90, 0x02, 0xf2, 0xe8, 0x7e, 0x12, 0x88, 0x36, 0x74, 0x84, 0x1a, 0x8a, 0xbb, 0xd9, 0x5d,
	0x23, 0x49, 0xa6, 0x98, 0xfe, 0xfa, 0xaf, 0x9a, 0xe8, 0x9a, 0xf6, 0xd9, 0x99, 0x97, 0x45, 0x09,
	0x1f, 0xe4, 0xd3, 0x34, 0xd0, 0x2c, 0xb9, 0xca, 0x74, 0x65, 0x56, 0x44, 0x04, 0x11, 0x6d, 0xea,
	0x65, 0xa6, 0x98, 0x2e, 0xcd, 0x92, 0xf1, 0x1b, 0xde, 0xcf, 0x76, 0x3c, 0x91, 0x14, 0x8a, 0xe9,
	0xca, 0x3c, 0xe0, 0xf7, 0x0f, 0x8a, 0x75, 0x19, 0x01, 0x72, 0x37, 0x59, 0x7f, 0xe8, 0xeb, 0x37,
	0xfc, 0x84, 0x92, 0x2e, 0x89, 0xfc, 0x3c, 0x04, 0x5f, 0x33, 0x2c, 0x40, 0x8c, 0x64, 0x8f, 0x35,
	0xdf, 0x8a, 0x3d, 0x8f, 0xce, 0xe5, 0xcb, 0xef, 0xcd, 0x2d, 0x00, 0x00, 0xff, 0xff, 0xdb, 0x2a,
	0xb6, 0xad, 0xc5, 0x00, 0x00, 0x00,
}
