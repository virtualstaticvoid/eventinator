// Code generated by protoc-gen-go. DO NOT EDIT.
// source: eventinator.proto

package protobuf

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
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

var E_Topic = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         74000,
	Name:          "eventinator.protobuf.topic",
	Tag:           "bytes,74000,opt,name=topic",
	Filename:      "eventinator.proto",
}

var E_Version = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MessageOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         74001,
	Name:          "eventinator.protobuf.version",
	Tag:           "bytes,74001,opt,name=version",
	Filename:      "eventinator.proto",
}

func init() {
	proto.RegisterExtension(E_Topic)
	proto.RegisterExtension(E_Version)
}

func init() { proto.RegisterFile("eventinator.proto", fileDescriptor_01d632fc99275ef7) }

var fileDescriptor_01d632fc99275ef7 = []byte{
	// 142 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0x2d, 0x4b, 0xcd,
	0x2b, 0xc9, 0xcc, 0x4b, 0x2c, 0xc9, 0x2f, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0xc1,
	0x10, 0x4a, 0x2a, 0x4d, 0x93, 0x52, 0x48, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x87, 0x09, 0xe8,
	0xa7, 0xa4, 0x16, 0x27, 0x17, 0x65, 0x16, 0xc0, 0x15, 0x59, 0x99, 0x73, 0xb1, 0x96, 0xe4, 0x17,
	0x64, 0x26, 0x0b, 0xc9, 0xeb, 0x41, 0xd4, 0xc2, 0x35, 0xeb, 0xf9, 0xa6, 0x16, 0x17, 0x27, 0xa6,
	0xa7, 0xfa, 0x17, 0x94, 0x64, 0xe6, 0xe7, 0x15, 0x4b, 0x4c, 0x38, 0xc4, 0xa2, 0xc0, 0xa8, 0xc1,
	0x19, 0x04, 0x51, 0x6f, 0x65, 0xcd, 0xc5, 0x5e, 0x96, 0x5a, 0x54, 0x9c, 0x99, 0x9f, 0x47, 0x58,
	0xeb, 0x44, 0xa8, 0x56, 0x98, 0x0e, 0x27, 0xae, 0x28, 0x0e, 0x98, 0xda, 0x24, 0x36, 0x30, 0xcb,
	0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xaf, 0xe3, 0x00, 0xe7, 0xd5, 0x00, 0x00, 0x00,
}
