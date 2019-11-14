// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tendermint.proto

package tendermint

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Tx struct {
	Payload              *Payload `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
	Signature            []byte   `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Tx) Reset()         { *m = Tx{} }
func (m *Tx) String() string { return proto.CompactTextString(m) }
func (*Tx) ProtoMessage()    {}
func (*Tx) Descriptor() ([]byte, []int) {
	return fileDescriptor_04f926c8da23c367, []int{0}
}

func (m *Tx) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Tx.Unmarshal(m, b)
}
func (m *Tx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Tx.Marshal(b, m, deterministic)
}
func (m *Tx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Tx.Merge(m, src)
}
func (m *Tx) XXX_Size() int {
	return xxx_messageInfo_Tx.Size(m)
}
func (m *Tx) XXX_DiscardUnknown() {
	xxx_messageInfo_Tx.DiscardUnknown(m)
}

var xxx_messageInfo_Tx proto.InternalMessageInfo

func (m *Tx) GetPayload() *Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *Tx) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

type Payload struct {
	Method               string   `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Params               []byte   `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Payload) Reset()         { *m = Payload{} }
func (m *Payload) String() string { return proto.CompactTextString(m) }
func (*Payload) ProtoMessage()    {}
func (*Payload) Descriptor() ([]byte, []int) {
	return fileDescriptor_04f926c8da23c367, []int{1}
}

func (m *Payload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Payload.Unmarshal(m, b)
}
func (m *Payload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Payload.Marshal(b, m, deterministic)
}
func (m *Payload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Payload.Merge(m, src)
}
func (m *Payload) XXX_Size() int {
	return xxx_messageInfo_Payload.Size(m)
}
func (m *Payload) XXX_DiscardUnknown() {
	xxx_messageInfo_Payload.DiscardUnknown(m)
}

var xxx_messageInfo_Payload proto.InternalMessageInfo

func (m *Payload) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *Payload) GetParams() []byte {
	if m != nil {
		return m.Params
	}
	return nil
}

type Query struct {
	Method               string   `protobuf:"bytes,1,opt,name=method,proto3" json:"method,omitempty"`
	Params               string   `protobuf:"bytes,2,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}
func (*Query) Descriptor() ([]byte, []int) {
	return fileDescriptor_04f926c8da23c367, []int{2}
}

func (m *Query) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Query.Unmarshal(m, b)
}
func (m *Query) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Query.Marshal(b, m, deterministic)
}
func (m *Query) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Query.Merge(m, src)
}
func (m *Query) XXX_Size() int {
	return xxx_messageInfo_Query.Size(m)
}
func (m *Query) XXX_DiscardUnknown() {
	xxx_messageInfo_Query.DiscardUnknown(m)
}

var xxx_messageInfo_Query proto.InternalMessageInfo

func (m *Query) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *Query) GetParams() string {
	if m != nil {
		return m.Params
	}
	return ""
}

func init() {
	proto.RegisterType((*Tx)(nil), "Tx")
	proto.RegisterType((*Payload)(nil), "Payload")
	proto.RegisterType((*Query)(nil), "Query")
}

func init() { proto.RegisterFile("tendermint.proto", fileDescriptor_04f926c8da23c367) }

var fileDescriptor_04f926c8da23c367 = []byte{
	// 150 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x28, 0x49, 0xcd, 0x4b,
	0x49, 0x2d, 0xca, 0xcd, 0xcc, 0x2b, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0x72, 0xe3, 0x62,
	0x0a, 0xa9, 0x10, 0x52, 0xe2, 0x62, 0x2f, 0x48, 0xac, 0xcc, 0xc9, 0x4f, 0x4c, 0x91, 0x60, 0x54,
	0x60, 0xd4, 0xe0, 0x36, 0xe2, 0xd0, 0x0b, 0x80, 0xf0, 0x83, 0x60, 0x12, 0x42, 0x32, 0x5c, 0x9c,
	0xc5, 0x99, 0xe9, 0x79, 0x89, 0x25, 0xa5, 0x45, 0xa9, 0x12, 0xcc, 0x0a, 0x8c, 0x1a, 0x3c, 0x41,
	0x08, 0x01, 0x25, 0x4b, 0x2e, 0x76, 0xa8, 0x0e, 0x21, 0x31, 0x2e, 0xb6, 0xdc, 0xd4, 0x92, 0x8c,
	0x7c, 0x88, 0x59, 0x9c, 0x41, 0x50, 0x1e, 0x48, 0xbc, 0x20, 0xb1, 0x28, 0x31, 0xb7, 0x58, 0x82,
	0x09, 0xac, 0x1b, 0xca, 0x53, 0x32, 0xe7, 0x62, 0x0d, 0x2c, 0x4d, 0x2d, 0xaa, 0x24, 0x52, 0x23,
	0x27, 0x4c, 0x63, 0x12, 0x1b, 0xd8, 0x0b, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5b, 0x67,
	0x15, 0x34, 0xd6, 0x00, 0x00, 0x00,
}
