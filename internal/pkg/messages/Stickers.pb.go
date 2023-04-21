//
// Copyright 2023 Signal Messenger, LLC
// SPDX-License-Identifier: AGPL-3.0-only
//

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: Stickers.proto

package messages

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StickerPack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Title    *string                `protobuf:"bytes,1,opt,name=title,proto3,oneof" json:"title,omitempty"`
	Author   *string                `protobuf:"bytes,2,opt,name=author,proto3,oneof" json:"author,omitempty"`
	Cover    *StickerPack_Sticker   `protobuf:"bytes,3,opt,name=cover,proto3,oneof" json:"cover,omitempty"`
	Stickers []*StickerPack_Sticker `protobuf:"bytes,4,rep,name=stickers,proto3" json:"stickers,omitempty"`
}

func (x *StickerPack) Reset() {
	*x = StickerPack{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Stickers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StickerPack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StickerPack) ProtoMessage() {}

func (x *StickerPack) ProtoReflect() protoreflect.Message {
	mi := &file_Stickers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StickerPack.ProtoReflect.Descriptor instead.
func (*StickerPack) Descriptor() ([]byte, []int) {
	return file_Stickers_proto_rawDescGZIP(), []int{0}
}

func (x *StickerPack) GetTitle() string {
	if x != nil && x.Title != nil {
		return *x.Title
	}
	return ""
}

func (x *StickerPack) GetAuthor() string {
	if x != nil && x.Author != nil {
		return *x.Author
	}
	return ""
}

func (x *StickerPack) GetCover() *StickerPack_Sticker {
	if x != nil {
		return x.Cover
	}
	return nil
}

func (x *StickerPack) GetStickers() []*StickerPack_Sticker {
	if x != nil {
		return x.Stickers
	}
	return nil
}

type StickerPack_Sticker struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    *uint32 `protobuf:"varint,1,opt,name=id,proto3,oneof" json:"id,omitempty"`
	Emoji *string `protobuf:"bytes,2,opt,name=emoji,proto3,oneof" json:"emoji,omitempty"`
}

func (x *StickerPack_Sticker) Reset() {
	*x = StickerPack_Sticker{}
	if protoimpl.UnsafeEnabled {
		mi := &file_Stickers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StickerPack_Sticker) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StickerPack_Sticker) ProtoMessage() {}

func (x *StickerPack_Sticker) ProtoReflect() protoreflect.Message {
	mi := &file_Stickers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StickerPack_Sticker.ProtoReflect.Descriptor instead.
func (*StickerPack_Sticker) Descriptor() ([]byte, []int) {
	return file_Stickers_proto_rawDescGZIP(), []int{0, 0}
}

func (x *StickerPack_Sticker) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *StickerPack_Sticker) GetEmoji() string {
	if x != nil && x.Emoji != nil {
		return *x.Emoji
	}
	return ""
}

var File_Stickers_proto protoreflect.FileDescriptor

var file_Stickers_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x53, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x93, 0x02, 0x0a, 0x0b, 0x53, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x50, 0x61, 0x63, 0x6b,
	0x12, 0x19, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48,
	0x00, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x88, 0x01, 0x01, 0x12, 0x1b, 0x0a, 0x06, 0x61,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x06, 0x61,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x88, 0x01, 0x01, 0x12, 0x2f, 0x0a, 0x05, 0x63, 0x6f, 0x76, 0x65,
	0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x53, 0x74, 0x69, 0x63, 0x6b, 0x65,
	0x72, 0x50, 0x61, 0x63, 0x6b, 0x2e, 0x53, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x48, 0x02, 0x52,
	0x05, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x88, 0x01, 0x01, 0x12, 0x30, 0x0a, 0x08, 0x73, 0x74, 0x69,
	0x63, 0x6b, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x53, 0x74,
	0x69, 0x63, 0x6b, 0x65, 0x72, 0x50, 0x61, 0x63, 0x6b, 0x2e, 0x53, 0x74, 0x69, 0x63, 0x6b, 0x65,
	0x72, 0x52, 0x08, 0x73, 0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x73, 0x1a, 0x4a, 0x0a, 0x07, 0x53,
	0x74, 0x69, 0x63, 0x6b, 0x65, 0x72, 0x12, 0x13, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x48, 0x00, 0x52, 0x02, 0x69, 0x64, 0x88, 0x01, 0x01, 0x12, 0x19, 0x0a, 0x05, 0x65,
	0x6d, 0x6f, 0x6a, 0x69, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52, 0x05, 0x65, 0x6d,
	0x6f, 0x6a, 0x69, 0x88, 0x01, 0x01, 0x42, 0x05, 0x0a, 0x03, 0x5f, 0x69, 0x64, 0x42, 0x08, 0x0a,
	0x06, 0x5f, 0x65, 0x6d, 0x6f, 0x6a, 0x69, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x42, 0x09, 0x0a, 0x07, 0x5f, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x42, 0x08, 0x0a, 0x06,
	0x5f, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x42, 0x17, 0x5a, 0x15, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e,
	0x61, 0x6c, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_Stickers_proto_rawDescOnce sync.Once
	file_Stickers_proto_rawDescData = file_Stickers_proto_rawDesc
)

func file_Stickers_proto_rawDescGZIP() []byte {
	file_Stickers_proto_rawDescOnce.Do(func() {
		file_Stickers_proto_rawDescData = protoimpl.X.CompressGZIP(file_Stickers_proto_rawDescData)
	})
	return file_Stickers_proto_rawDescData
}

var file_Stickers_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_Stickers_proto_goTypes = []interface{}{
	(*StickerPack)(nil),         // 0: StickerPack
	(*StickerPack_Sticker)(nil), // 1: StickerPack.Sticker
}
var file_Stickers_proto_depIdxs = []int32{
	1, // 0: StickerPack.cover:type_name -> StickerPack.Sticker
	1, // 1: StickerPack.stickers:type_name -> StickerPack.Sticker
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_Stickers_proto_init() }
func file_Stickers_proto_init() {
	if File_Stickers_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_Stickers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StickerPack); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_Stickers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StickerPack_Sticker); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_Stickers_proto_msgTypes[0].OneofWrappers = []interface{}{}
	file_Stickers_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_Stickers_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_Stickers_proto_goTypes,
		DependencyIndexes: file_Stickers_proto_depIdxs,
		MessageInfos:      file_Stickers_proto_msgTypes,
	}.Build()
	File_Stickers_proto = out.File
	file_Stickers_proto_rawDesc = nil
	file_Stickers_proto_goTypes = nil
	file_Stickers_proto_depIdxs = nil
}
