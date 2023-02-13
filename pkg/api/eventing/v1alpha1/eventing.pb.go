// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: pkg/api/eventing/v1alpha1/eventing.proto

package v1alpha1

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

type RequestHandoffMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Identifier     string `protobuf:"bytes,2,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Password       string `protobuf:"bytes,3,opt,name=password,proto3" json:"password,omitempty"`
	ClientVersion  string `protobuf:"bytes,5,opt,name=client_version,json=clientVersion,proto3" json:"client_version,omitempty"`
	ClientChecksum string `protobuf:"bytes,7,opt,name=client_checksum,json=clientChecksum,proto3" json:"client_checksum,omitempty"`
}

func (x *RequestHandoffMessage) Reset() {
	*x = RequestHandoffMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestHandoffMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestHandoffMessage) ProtoMessage() {}

func (x *RequestHandoffMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestHandoffMessage.ProtoReflect.Descriptor instead.
func (*RequestHandoffMessage) Descriptor() ([]byte, []int) {
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP(), []int{0}
}

func (x *RequestHandoffMessage) GetIdentifier() string {
	if x != nil {
		return x.Identifier
	}
	return ""
}

func (x *RequestHandoffMessage) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *RequestHandoffMessage) GetClientVersion() string {
	if x != nil {
		return x.ClientVersion
	}
	return ""
}

func (x *RequestHandoffMessage) GetClientChecksum() string {
	if x != nil {
		return x.ClientChecksum
	}
	return ""
}

type Gateway struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host       string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Port       string `protobuf:"bytes,2,opt,name=port,proto3" json:"port,omitempty"`
	Population uint32 `protobuf:"varint,3,opt,name=population,proto3" json:"population,omitempty"`
	Capacity   uint32 `protobuf:"varint,4,opt,name=capacity,proto3" json:"capacity,omitempty"`
	WorldId    uint32 `protobuf:"varint,5,opt,name=world_id,json=worldId,proto3" json:"world_id,omitempty"`
	ChannelId  uint32 `protobuf:"varint,6,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	WorldName  string `protobuf:"bytes,7,opt,name=world_name,json=worldName,proto3" json:"world_name,omitempty"`
}

func (x *Gateway) Reset() {
	*x = Gateway{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Gateway) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Gateway) ProtoMessage() {}

func (x *Gateway) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Gateway.ProtoReflect.Descriptor instead.
func (*Gateway) Descriptor() ([]byte, []int) {
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP(), []int{1}
}

func (x *Gateway) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Gateway) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *Gateway) GetPopulation() uint32 {
	if x != nil {
		return x.Population
	}
	return 0
}

func (x *Gateway) GetCapacity() uint32 {
	if x != nil {
		return x.Capacity
	}
	return 0
}

func (x *Gateway) GetWorldId() uint32 {
	if x != nil {
		return x.WorldId
	}
	return 0
}

func (x *Gateway) GetChannelId() uint32 {
	if x != nil {
		return x.ChannelId
	}
	return 0
}

func (x *Gateway) GetWorldName() string {
	if x != nil {
		return x.WorldName
	}
	return ""
}

type ProposeHandoffMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key      uint32     `protobuf:"varint,1,opt,name=key,proto3" json:"key,omitempty"`
	Gateways []*Gateway `protobuf:"bytes,2,rep,name=gateways,proto3" json:"gateways,omitempty"`
}

func (x *ProposeHandoffMessage) Reset() {
	*x = ProposeHandoffMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProposeHandoffMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProposeHandoffMessage) ProtoMessage() {}

func (x *ProposeHandoffMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProposeHandoffMessage.ProtoReflect.Descriptor instead.
func (*ProposeHandoffMessage) Descriptor() ([]byte, []int) {
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP(), []int{2}
}

func (x *ProposeHandoffMessage) GetKey() uint32 {
	if x != nil {
		return x.Key
	}
	return 0
}

func (x *ProposeHandoffMessage) GetGateways() []*Gateway {
	if x != nil {
		return x.Gateways
	}
	return nil
}

type SyncMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sequence uint32 `protobuf:"varint,1,opt,name=sequence,proto3" json:"sequence,omitempty"`
}

func (x *SyncMessage) Reset() {
	*x = SyncMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncMessage) ProtoMessage() {}

func (x *SyncMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncMessage.ProtoReflect.Descriptor instead.
func (*SyncMessage) Descriptor() ([]byte, []int) {
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP(), []int{3}
}

func (x *SyncMessage) GetSequence() uint32 {
	if x != nil {
		return x.Sequence
	}
	return 0
}

type PerformHandoffMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeySequence      uint32 `protobuf:"varint,1,opt,name=key_sequence,json=keySequence,proto3" json:"key_sequence,omitempty"`
	Key              uint32 `protobuf:"varint,2,opt,name=key,proto3" json:"key,omitempty"`
	PasswordSequence uint32 `protobuf:"varint,3,opt,name=password_sequence,json=passwordSequence,proto3" json:"password_sequence,omitempty"`
	Password         string `protobuf:"bytes,4,opt,name=password,proto3" json:"password,omitempty"`
}

func (x *PerformHandoffMessage) Reset() {
	*x = PerformHandoffMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PerformHandoffMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PerformHandoffMessage) ProtoMessage() {}

func (x *PerformHandoffMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PerformHandoffMessage.ProtoReflect.Descriptor instead.
func (*PerformHandoffMessage) Descriptor() ([]byte, []int) {
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP(), []int{4}
}

func (x *PerformHandoffMessage) GetKeySequence() uint32 {
	if x != nil {
		return x.KeySequence
	}
	return 0
}

func (x *PerformHandoffMessage) GetKey() uint32 {
	if x != nil {
		return x.Key
	}
	return 0
}

func (x *PerformHandoffMessage) GetPasswordSequence() uint32 {
	if x != nil {
		return x.PasswordSequence
	}
	return 0
}

func (x *PerformHandoffMessage) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

var File_pkg_api_eventing_v1alpha1_eventing_proto protoreflect.FileDescriptor

var file_pkg_api_eventing_v1alpha1_eventing_proto_rawDesc = []byte{
	0x0a, 0x28, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x69,
	0x6e, 0x67, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x69, 0x6f, 0x2e, 0x65,
	0x6c, 0x6b, 0x69, 0x61, 0x2e, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x22, 0xa3, 0x01, 0x0a, 0x15, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x61, 0x6e, 0x64, 0x6f, 0x66, 0x66, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x25, 0x0a, 0x0e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x68,
	0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x73, 0x75, 0x6d, 0x22, 0xc6, 0x01, 0x0a,
	0x07, 0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74,
	0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6f, 0x70, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x70, 0x6f, 0x70, 0x75, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x08, 0x63, 0x61, 0x70, 0x61, 0x63, 0x69, 0x74, 0x79, 0x12, 0x19, 0x0a, 0x08,
	0x77, 0x6f, 0x72, 0x6c, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07,
	0x77, 0x6f, 0x72, 0x6c, 0x64, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x63, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x77, 0x6f, 0x72, 0x6c,
	0x64, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x6a, 0x0a, 0x15, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65,
	0x48, 0x61, 0x6e, 0x64, 0x6f, 0x66, 0x66, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x3f, 0x0a, 0x08, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x69, 0x6f, 0x2e, 0x65, 0x6c, 0x6b, 0x69, 0x61, 0x2e, 0x65, 0x76,
	0x65, 0x6e, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x47, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x52, 0x08, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79,
	0x73, 0x22, 0x29, 0x0a, 0x0b, 0x53, 0x79, 0x6e, 0x63, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x1a, 0x0a, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x08, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x95, 0x01, 0x0a,
	0x15, 0x50, 0x65, 0x72, 0x66, 0x6f, 0x72, 0x6d, 0x48, 0x61, 0x6e, 0x64, 0x6f, 0x66, 0x66, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x6b, 0x65, 0x79, 0x5f, 0x73, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x6b, 0x65,
	0x79, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2b, 0x0a, 0x11, 0x70,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x5f, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x10, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x42, 0x1b, 0x5a, 0x19, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x69, 0x6e, 0x67, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescOnce sync.Once
	file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescData = file_pkg_api_eventing_v1alpha1_eventing_proto_rawDesc
)

func file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescGZIP() []byte {
	file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescOnce.Do(func() {
		file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescData)
	})
	return file_pkg_api_eventing_v1alpha1_eventing_proto_rawDescData
}

var file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_api_eventing_v1alpha1_eventing_proto_goTypes = []interface{}{
	(*RequestHandoffMessage)(nil), // 0: io.elkia.eventing.v1alpha1.RequestHandoffMessage
	(*Gateway)(nil),               // 1: io.elkia.eventing.v1alpha1.Gateway
	(*ProposeHandoffMessage)(nil), // 2: io.elkia.eventing.v1alpha1.ProposeHandoffMessage
	(*SyncMessage)(nil),           // 3: io.elkia.eventing.v1alpha1.SyncMessage
	(*PerformHandoffMessage)(nil), // 4: io.elkia.eventing.v1alpha1.PerformHandoffMessage
}
var file_pkg_api_eventing_v1alpha1_eventing_proto_depIdxs = []int32{
	1, // 0: io.elkia.eventing.v1alpha1.ProposeHandoffMessage.gateways:type_name -> io.elkia.eventing.v1alpha1.Gateway
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_api_eventing_v1alpha1_eventing_proto_init() }
func file_pkg_api_eventing_v1alpha1_eventing_proto_init() {
	if File_pkg_api_eventing_v1alpha1_eventing_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestHandoffMessage); i {
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
		file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Gateway); i {
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
		file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProposeHandoffMessage); i {
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
		file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncMessage); i {
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
		file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PerformHandoffMessage); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_api_eventing_v1alpha1_eventing_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pkg_api_eventing_v1alpha1_eventing_proto_goTypes,
		DependencyIndexes: file_pkg_api_eventing_v1alpha1_eventing_proto_depIdxs,
		MessageInfos:      file_pkg_api_eventing_v1alpha1_eventing_proto_msgTypes,
	}.Build()
	File_pkg_api_eventing_v1alpha1_eventing_proto = out.File
	file_pkg_api_eventing_v1alpha1_eventing_proto_rawDesc = nil
	file_pkg_api_eventing_v1alpha1_eventing_proto_goTypes = nil
	file_pkg_api_eventing_v1alpha1_eventing_proto_depIdxs = nil
}