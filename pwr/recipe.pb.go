// Code generated by protoc-gen-go.
// source: pwr/recipe.proto
// DO NOT EDIT!

/*
Package pwr is a generated protocol buffer package.

It is generated from these files:
	pwr/recipe.proto

It has these top-level messages:
	RecipeHeader
	RepoInfo
	RsyncSignatureHeader
	RsyncBlockHash
	RsyncOp
*/
package pwr

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type HashType int32

const (
	// librsync default
	HashType_MD5 HashType = 0
	// https://godoc.org/golang.org/x/crypto/sha3#ShakeSum256
	HashType_SHAKESUM256 HashType = 1
)

var HashType_name = map[int32]string{
	0: "MD5",
	1: "SHAKESUM256",
}
var HashType_value = map[string]int32{
	"MD5":         0,
	"SHAKESUM256": 1,
}

func (x HashType) String() string {
	return proto.EnumName(HashType_name, int32(x))
}
func (HashType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type RecipeHeader_Version int32

const (
	RecipeHeader_V1 RecipeHeader_Version = 0
)

var RecipeHeader_Version_name = map[int32]string{
	0: "V1",
}
var RecipeHeader_Version_value = map[string]int32{
	"V1": 0,
}

func (x RecipeHeader_Version) String() string {
	return proto.EnumName(RecipeHeader_Version_name, int32(x))
}
func (RecipeHeader_Version) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type RecipeHeader_Compression int32

const (
	RecipeHeader_UNCOMPRESSED RecipeHeader_Compression = 0
	RecipeHeader_BROTLI       RecipeHeader_Compression = 1
)

var RecipeHeader_Compression_name = map[int32]string{
	0: "UNCOMPRESSED",
	1: "BROTLI",
}
var RecipeHeader_Compression_value = map[string]int32{
	"UNCOMPRESSED": 0,
	"BROTLI":       1,
}

func (x RecipeHeader_Compression) String() string {
	return proto.EnumName(RecipeHeader_Compression_name, int32(x))
}
func (RecipeHeader_Compression) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 1} }

type RsyncOp_Type int32

const (
	RsyncOp_BLOCK          RsyncOp_Type = 0
	RsyncOp_BLOCK_RANGE    RsyncOp_Type = 1
	RsyncOp_DATA           RsyncOp_Type = 2
	RsyncOp_HEY_YOU_DID_IT RsyncOp_Type = 2049
)

var RsyncOp_Type_name = map[int32]string{
	0:    "BLOCK",
	1:    "BLOCK_RANGE",
	2:    "DATA",
	2049: "HEY_YOU_DID_IT",
}
var RsyncOp_Type_value = map[string]int32{
	"BLOCK":          0,
	"BLOCK_RANGE":    1,
	"DATA":           2,
	"HEY_YOU_DID_IT": 2049,
}

func (x RsyncOp_Type) String() string {
	return proto.EnumName(RsyncOp_Type_name, int32(x))
}
func (RsyncOp_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{4, 0} }

type RecipeHeader struct {
	Version          RecipeHeader_Version     `protobuf:"varint,1,opt,name=version,enum=io.itch.wharf.pwr.RecipeHeader_Version" json:"version,omitempty"`
	Compression      RecipeHeader_Compression `protobuf:"varint,2,opt,name=compression,enum=io.itch.wharf.pwr.RecipeHeader_Compression" json:"compression,omitempty"`
	CompressionLevel int32                    `protobuf:"varint,3,opt,name=compressionLevel" json:"compressionLevel,omitempty"`
}

func (m *RecipeHeader) Reset()                    { *m = RecipeHeader{} }
func (m *RecipeHeader) String() string            { return proto.CompactTextString(m) }
func (*RecipeHeader) ProtoMessage()               {}
func (*RecipeHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type RepoInfo struct {
	NumBlocks int64               `protobuf:"varint,16,opt,name=numBlocks" json:"numBlocks,omitempty"`
	Dirs      []*RepoInfo_Dir     `protobuf:"bytes,1,rep,name=dirs" json:"dirs,omitempty"`
	Files     []*RepoInfo_File    `protobuf:"bytes,2,rep,name=files" json:"files,omitempty"`
	Symlinks  []*RepoInfo_Symlink `protobuf:"bytes,3,rep,name=symlinks" json:"symlinks,omitempty"`
}

func (m *RepoInfo) Reset()                    { *m = RepoInfo{} }
func (m *RepoInfo) String() string            { return proto.CompactTextString(m) }
func (*RepoInfo) ProtoMessage()               {}
func (*RepoInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RepoInfo) GetDirs() []*RepoInfo_Dir {
	if m != nil {
		return m.Dirs
	}
	return nil
}

func (m *RepoInfo) GetFiles() []*RepoInfo_File {
	if m != nil {
		return m.Files
	}
	return nil
}

func (m *RepoInfo) GetSymlinks() []*RepoInfo_Symlink {
	if m != nil {
		return m.Symlinks
	}
	return nil
}

type RepoInfo_Dir struct {
	Path string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Mode uint32 `protobuf:"varint,2,opt,name=mode" json:"mode,omitempty"`
}

func (m *RepoInfo_Dir) Reset()                    { *m = RepoInfo_Dir{} }
func (m *RepoInfo_Dir) String() string            { return proto.CompactTextString(m) }
func (*RepoInfo_Dir) ProtoMessage()               {}
func (*RepoInfo_Dir) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

type RepoInfo_File struct {
	Path          string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Mode          uint32 `protobuf:"varint,2,opt,name=mode" json:"mode,omitempty"`
	Size          int64  `protobuf:"varint,3,opt,name=size" json:"size,omitempty"`
	BlockIndex    int64  `protobuf:"varint,4,opt,name=blockIndex" json:"blockIndex,omitempty"`
	BlockIndexEnd int64  `protobuf:"varint,5,opt,name=blockIndexEnd" json:"blockIndexEnd,omitempty"`
}

func (m *RepoInfo_File) Reset()                    { *m = RepoInfo_File{} }
func (m *RepoInfo_File) String() string            { return proto.CompactTextString(m) }
func (*RepoInfo_File) ProtoMessage()               {}
func (*RepoInfo_File) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 1} }

type RepoInfo_Symlink struct {
	Path string `protobuf:"bytes,1,opt,name=path" json:"path,omitempty"`
	Mode uint32 `protobuf:"varint,2,opt,name=mode" json:"mode,omitempty"`
	Dest string `protobuf:"bytes,3,opt,name=dest" json:"dest,omitempty"`
}

func (m *RepoInfo_Symlink) Reset()                    { *m = RepoInfo_Symlink{} }
func (m *RepoInfo_Symlink) String() string            { return proto.CompactTextString(m) }
func (*RepoInfo_Symlink) ProtoMessage()               {}
func (*RepoInfo_Symlink) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 2} }

type RsyncSignatureHeader struct {
	BlockCount uint64   `protobuf:"varint,1,opt,name=blockCount" json:"blockCount,omitempty"`
	HashType   HashType `protobuf:"varint,2,opt,name=hashType,enum=io.itch.wharf.pwr.HashType" json:"hashType,omitempty"`
}

func (m *RsyncSignatureHeader) Reset()                    { *m = RsyncSignatureHeader{} }
func (m *RsyncSignatureHeader) String() string            { return proto.CompactTextString(m) }
func (*RsyncSignatureHeader) ProtoMessage()               {}
func (*RsyncSignatureHeader) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type RsyncBlockHash struct {
	Index      uint64 `protobuf:"varint,1,opt,name=index" json:"index,omitempty"`
	StrongHash []byte `protobuf:"bytes,2,opt,name=strongHash,proto3" json:"strongHash,omitempty"`
	WeakHash   uint32 `protobuf:"varint,3,opt,name=weakHash" json:"weakHash,omitempty"`
}

func (m *RsyncBlockHash) Reset()                    { *m = RsyncBlockHash{} }
func (m *RsyncBlockHash) String() string            { return proto.CompactTextString(m) }
func (*RsyncBlockHash) ProtoMessage()               {}
func (*RsyncBlockHash) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type RsyncOp struct {
	Type          RsyncOp_Type `protobuf:"varint,1,opt,name=type,enum=io.itch.wharf.pwr.RsyncOp_Type" json:"type,omitempty"`
	BlockIndex    uint64       `protobuf:"varint,2,opt,name=blockIndex" json:"blockIndex,omitempty"`
	BlockIndexEnd uint64       `protobuf:"varint,3,opt,name=blockIndexEnd" json:"blockIndexEnd,omitempty"`
	Data          []byte       `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *RsyncOp) Reset()                    { *m = RsyncOp{} }
func (m *RsyncOp) String() string            { return proto.CompactTextString(m) }
func (*RsyncOp) ProtoMessage()               {}
func (*RsyncOp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func init() {
	proto.RegisterType((*RecipeHeader)(nil), "io.itch.wharf.pwr.RecipeHeader")
	proto.RegisterType((*RepoInfo)(nil), "io.itch.wharf.pwr.RepoInfo")
	proto.RegisterType((*RepoInfo_Dir)(nil), "io.itch.wharf.pwr.RepoInfo.Dir")
	proto.RegisterType((*RepoInfo_File)(nil), "io.itch.wharf.pwr.RepoInfo.File")
	proto.RegisterType((*RepoInfo_Symlink)(nil), "io.itch.wharf.pwr.RepoInfo.Symlink")
	proto.RegisterType((*RsyncSignatureHeader)(nil), "io.itch.wharf.pwr.RsyncSignatureHeader")
	proto.RegisterType((*RsyncBlockHash)(nil), "io.itch.wharf.pwr.RsyncBlockHash")
	proto.RegisterType((*RsyncOp)(nil), "io.itch.wharf.pwr.RsyncOp")
	proto.RegisterEnum("io.itch.wharf.pwr.HashType", HashType_name, HashType_value)
	proto.RegisterEnum("io.itch.wharf.pwr.RecipeHeader_Version", RecipeHeader_Version_name, RecipeHeader_Version_value)
	proto.RegisterEnum("io.itch.wharf.pwr.RecipeHeader_Compression", RecipeHeader_Compression_name, RecipeHeader_Compression_value)
	proto.RegisterEnum("io.itch.wharf.pwr.RsyncOp_Type", RsyncOp_Type_name, RsyncOp_Type_value)
}

var fileDescriptor0 = []byte{
	// 629 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x54, 0x41, 0x4f, 0xdb, 0x4c,
	0x10, 0xc5, 0xb1, 0x43, 0xcc, 0x24, 0xf0, 0x2d, 0xfb, 0x71, 0x88, 0xd2, 0xaa, 0x45, 0x2e, 0x52,
	0x11, 0xa8, 0xae, 0x0a, 0x82, 0x1e, 0x2b, 0x27, 0x76, 0x9b, 0x08, 0x42, 0xaa, 0x75, 0x40, 0xa2,
	0x97, 0xc8, 0xc4, 0x0b, 0xb1, 0x9a, 0x78, 0x2d, 0xdb, 0x40, 0xe9, 0xad, 0xb7, 0x1e, 0xfa, 0xff,
	0x7a, 0xe8, 0x9f, 0xa9, 0x77, 0x9c, 0x80, 0x29, 0x94, 0xb6, 0xb7, 0xd9, 0xd9, 0xf7, 0xde, 0xbc,
	0x9d, 0x19, 0x1b, 0x48, 0x74, 0x19, 0xbf, 0x8c, 0xf9, 0x30, 0x88, 0xb8, 0x19, 0xc5, 0x22, 0x15,
	0x74, 0x39, 0x10, 0x66, 0x90, 0x0e, 0x47, 0xe6, 0xe5, 0xc8, 0x8b, 0x4f, 0xcd, 0xec, 0xde, 0xf8,
	0x56, 0x82, 0x1a, 0x43, 0x4c, 0x9b, 0x7b, 0x3e, 0x8f, 0xa9, 0x05, 0x95, 0x0b, 0x1e, 0x27, 0x81,
	0x08, 0xeb, 0xca, 0xaa, 0xb2, 0xbe, 0xb4, 0xf5, 0xdc, 0xbc, 0xc3, 0x32, 0x8b, 0x0c, 0xf3, 0x28,
	0x87, 0xb3, 0x19, 0x8f, 0x76, 0xa1, 0x3a, 0x14, 0x93, 0x28, 0xe6, 0x09, 0xca, 0x94, 0x50, 0x66,
	0xf3, 0x4f, 0x32, 0xad, 0x1b, 0x0a, 0x2b, 0xf2, 0xe9, 0x06, 0x90, 0xc2, 0x71, 0x9f, 0x5f, 0xf0,
	0x71, 0x5d, 0xcd, 0x34, 0xcb, 0xec, 0x4e, 0xde, 0x58, 0x86, 0xca, 0xd4, 0x0e, 0x9d, 0x87, 0xd2,
	0xd1, 0x2b, 0x32, 0x67, 0x6c, 0x42, 0xb5, 0x20, 0x4d, 0x09, 0xd4, 0x0e, 0x0f, 0x5a, 0xbd, 0xee,
	0x7b, 0xe6, 0xb8, 0xae, 0x63, 0x93, 0x39, 0x0a, 0x30, 0xdf, 0x64, 0xbd, 0xfe, 0x7e, 0x87, 0x28,
	0xc6, 0x77, 0x15, 0x74, 0xc6, 0x23, 0xd1, 0x09, 0x4f, 0x05, 0x7d, 0x0c, 0x0b, 0xe1, 0xf9, 0xa4,
	0x39, 0x16, 0xc3, 0x8f, 0x49, 0x9d, 0x64, 0x15, 0x55, 0x76, 0x93, 0xa0, 0xdb, 0xa0, 0xf9, 0x41,
	0x9c, 0x64, 0x5d, 0x52, 0xd7, 0xab, 0x5b, 0x4f, 0xef, 0x7d, 0x5e, 0x2e, 0x64, 0xda, 0x41, 0xcc,
	0x10, 0x4c, 0x77, 0xa1, 0x7c, 0x1a, 0x8c, 0x79, 0x92, 0x35, 0x45, 0xb2, 0x56, 0x1f, 0x62, 0xbd,
	0xcd, 0x80, 0x2c, 0x87, 0xd3, 0x37, 0xa0, 0x27, 0x57, 0x93, 0x71, 0x10, 0x66, 0x4e, 0x54, 0xa4,
	0x3e, 0x7b, 0x88, 0xea, 0xe6, 0x58, 0x76, 0x4d, 0x6a, 0xbc, 0x00, 0x35, 0x73, 0x41, 0x29, 0x68,
	0x91, 0x97, 0x8e, 0x70, 0xb4, 0x0b, 0x0c, 0x63, 0x99, 0x9b, 0x08, 0x9f, 0xe3, 0x9c, 0x16, 0x19,
	0xc6, 0x8d, 0xaf, 0x0a, 0x68, 0xb2, 0xfe, 0xdf, 0x12, 0x64, 0x2e, 0x09, 0x3e, 0x73, 0x1c, 0x8c,
	0xca, 0x30, 0xa6, 0x4f, 0x00, 0x4e, 0x64, 0xaf, 0x3a, 0xa1, 0xcf, 0x3f, 0xd5, 0x35, 0xbc, 0x29,
	0x64, 0xe8, 0x1a, 0x2c, 0xde, 0x9c, 0x9c, 0xd0, 0xaf, 0x97, 0x11, 0x72, 0x3b, 0xd9, 0x70, 0xa0,
	0x32, 0x7d, 0xce, 0xbf, 0x98, 0xf1, 0x79, 0x92, 0xa2, 0x99, 0x0c, 0x27, 0x63, 0x43, 0xc0, 0x0a,
	0x4b, 0xae, 0xc2, 0xa1, 0x1b, 0x9c, 0x85, 0x5e, 0x7a, 0x1e, 0xcf, 0xf6, 0x7d, 0x66, 0xb2, 0x25,
	0xce, 0xc3, 0x14, 0x95, 0x35, 0x56, 0xc8, 0xd0, 0xd7, 0xa0, 0x8f, 0xbc, 0x64, 0xd4, 0xbf, 0x8a,
	0xf8, 0x74, 0x93, 0x1f, 0xdd, 0xd3, 0xf9, 0xf6, 0x14, 0xc2, 0xae, 0xc1, 0xc6, 0x09, 0x2c, 0x61,
	0x41, 0x5c, 0x17, 0x79, 0x4f, 0x57, 0xa0, 0x1c, 0x60, 0x2b, 0xf2, 0x2a, 0xf9, 0x41, 0x1a, 0x48,
	0xd2, 0x58, 0x84, 0x67, 0x12, 0x83, 0x25, 0x6a, 0xac, 0x90, 0xa1, 0x0d, 0xd0, 0x2f, 0xb9, 0x87,
	0x0a, 0xf8, 0xa0, 0x45, 0x76, 0x7d, 0x36, 0x7e, 0x28, 0x50, 0xc1, 0x22, 0xbd, 0x48, 0xee, 0x63,
	0x2a, 0x4d, 0xe6, 0x5f, 0xed, 0xbd, 0xfb, 0x98, 0x23, 0x4d, 0x34, 0x8a, 0xe0, 0x5f, 0x46, 0x54,
	0x2a, 0xbc, 0xfe, 0x37, 0x23, 0x52, 0x11, 0x72, 0x3b, 0x89, 0xfd, 0xf6, 0x52, 0x0f, 0x47, 0x5c,
	0x63, 0x18, 0x1b, 0x16, 0x68, 0xb2, 0x0e, 0x5d, 0x80, 0x72, 0x73, 0xbf, 0xd7, 0xda, 0xcb, 0x3e,
	0xb4, 0xff, 0xa0, 0x8a, 0xe1, 0x80, 0x59, 0x07, 0xef, 0x1c, 0xa2, 0x50, 0x1d, 0x34, 0xdb, 0xea,
	0x5b, 0xa4, 0x44, 0xff, 0x87, 0xa5, 0xb6, 0x73, 0x3c, 0x38, 0xee, 0x1d, 0x0e, 0xec, 0x8e, 0x3d,
	0xe8, 0xf4, 0xc9, 0x17, 0xb2, 0xb1, 0x06, 0xfa, 0xac, 0xaf, 0xb4, 0x02, 0x6a, 0xd7, 0xde, 0xc9,
	0x45, 0xdc, 0xb6, 0xb5, 0xe7, 0xb8, 0x87, 0xdd, 0xad, 0x9d, 0x5d, 0xa2, 0x34, 0xcb, 0x1f, 0xd4,
	0xec, 0x71, 0x27, 0xf3, 0xf8, 0x8b, 0xdb, 0xfe, 0x19, 0x00, 0x00, 0xff, 0xff, 0x04, 0x7c, 0xa5,
	0x87, 0xf6, 0x04, 0x00, 0x00,
}
