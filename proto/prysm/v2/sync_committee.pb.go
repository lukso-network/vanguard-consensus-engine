// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.15.8
// source: proto/prysm/v2/sync_committee.proto

package v2

import (
	proto "github.com/golang/protobuf/proto"
	github_com_prysmaticlabs_eth2_types "github.com/prysmaticlabs/eth2-types"
	github_com_prysmaticlabs_go_bitfield "github.com/prysmaticlabs/go-bitfield"
	_ "github.com/prysmaticlabs/prysm/proto/eth/ext"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type SyncCommitteeMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Slot           github_com_prysmaticlabs_eth2_types.Slot           `protobuf:"varint,1,opt,name=slot,proto3" json:"slot,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Slot"`
	BlockRoot      []byte                                             `protobuf:"bytes,2,opt,name=block_root,json=blockRoot,proto3" json:"block_root,omitempty" ssz-size:"32"`
	ValidatorIndex github_com_prysmaticlabs_eth2_types.ValidatorIndex `protobuf:"varint,3,opt,name=validator_index,json=validatorIndex,proto3" json:"validator_index,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.ValidatorIndex"`
	Signature      []byte                                             `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *SyncCommitteeMessage) Reset() {
	*x = SyncCommitteeMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncCommitteeMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncCommitteeMessage) ProtoMessage() {}

func (x *SyncCommitteeMessage) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncCommitteeMessage.ProtoReflect.Descriptor instead.
func (*SyncCommitteeMessage) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v2_sync_committee_proto_rawDescGZIP(), []int{0}
}

func (x *SyncCommitteeMessage) GetSlot() github_com_prysmaticlabs_eth2_types.Slot {
	if x != nil {
		return x.Slot
	}
	return github_com_prysmaticlabs_eth2_types.Slot(0)
}

func (x *SyncCommitteeMessage) GetBlockRoot() []byte {
	if x != nil {
		return x.BlockRoot
	}
	return nil
}

func (x *SyncCommitteeMessage) GetValidatorIndex() github_com_prysmaticlabs_eth2_types.ValidatorIndex {
	if x != nil {
		return x.ValidatorIndex
	}
	return github_com_prysmaticlabs_eth2_types.ValidatorIndex(0)
}

func (x *SyncCommitteeMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type SyncCommitteeContribution struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Slot              github_com_prysmaticlabs_eth2_types.Slot          `protobuf:"varint,1,opt,name=slot,proto3" json:"slot,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.Slot"`
	BlockRoot         []byte                                            `protobuf:"bytes,2,opt,name=block_root,json=blockRoot,proto3" json:"block_root,omitempty" ssz-size:"32"`
	SubcommitteeIndex uint64                                            `protobuf:"varint,3,opt,name=subcommittee_index,json=subcommitteeIndex,proto3" json:"subcommittee_index,omitempty"`
	AggregationBits   github_com_prysmaticlabs_go_bitfield.Bitvector128 `protobuf:"bytes,4,opt,name=aggregation_bits,json=aggregationBits,proto3" json:"aggregation_bits,omitempty" cast-type:"github.com/prysmaticlabs/go-bitfield.Bitvector128" ssz-size:"16"`
	Signature         []byte                                            `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *SyncCommitteeContribution) Reset() {
	*x = SyncCommitteeContribution{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncCommitteeContribution) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncCommitteeContribution) ProtoMessage() {}

func (x *SyncCommitteeContribution) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncCommitteeContribution.ProtoReflect.Descriptor instead.
func (*SyncCommitteeContribution) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v2_sync_committee_proto_rawDescGZIP(), []int{1}
}

func (x *SyncCommitteeContribution) GetSlot() github_com_prysmaticlabs_eth2_types.Slot {
	if x != nil {
		return x.Slot
	}
	return github_com_prysmaticlabs_eth2_types.Slot(0)
}

func (x *SyncCommitteeContribution) GetBlockRoot() []byte {
	if x != nil {
		return x.BlockRoot
	}
	return nil
}

func (x *SyncCommitteeContribution) GetSubcommitteeIndex() uint64 {
	if x != nil {
		return x.SubcommitteeIndex
	}
	return 0
}

func (x *SyncCommitteeContribution) GetAggregationBits() github_com_prysmaticlabs_go_bitfield.Bitvector128 {
	if x != nil {
		return x.AggregationBits
	}
	return github_com_prysmaticlabs_go_bitfield.Bitvector128(nil)
}

func (x *SyncCommitteeContribution) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type ContributionAndProof struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AggregatorIndex github_com_prysmaticlabs_eth2_types.ValidatorIndex `protobuf:"varint,1,opt,name=aggregator_index,json=aggregatorIndex,proto3" json:"aggregator_index,omitempty" cast-type:"github.com/prysmaticlabs/eth2-types.ValidatorIndex"`
	Contribution    *SyncCommitteeContribution                         `protobuf:"bytes,2,opt,name=contribution,proto3" json:"contribution,omitempty"`
	SelectionProof  []byte                                             `protobuf:"bytes,3,opt,name=selection_proof,json=selectionProof,proto3" json:"selection_proof,omitempty" ssz-size:"96"`
}

func (x *ContributionAndProof) Reset() {
	*x = ContributionAndProof{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ContributionAndProof) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ContributionAndProof) ProtoMessage() {}

func (x *ContributionAndProof) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ContributionAndProof.ProtoReflect.Descriptor instead.
func (*ContributionAndProof) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v2_sync_committee_proto_rawDescGZIP(), []int{2}
}

func (x *ContributionAndProof) GetAggregatorIndex() github_com_prysmaticlabs_eth2_types.ValidatorIndex {
	if x != nil {
		return x.AggregatorIndex
	}
	return github_com_prysmaticlabs_eth2_types.ValidatorIndex(0)
}

func (x *ContributionAndProof) GetContribution() *SyncCommitteeContribution {
	if x != nil {
		return x.Contribution
	}
	return nil
}

func (x *ContributionAndProof) GetSelectionProof() []byte {
	if x != nil {
		return x.SelectionProof
	}
	return nil
}

type SignedContributionAndProof struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message   *ContributionAndProof `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Signature []byte                `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty" ssz-size:"96"`
}

func (x *SignedContributionAndProof) Reset() {
	*x = SignedContributionAndProof{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignedContributionAndProof) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignedContributionAndProof) ProtoMessage() {}

func (x *SignedContributionAndProof) ProtoReflect() protoreflect.Message {
	mi := &file_proto_prysm_v2_sync_committee_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignedContributionAndProof.ProtoReflect.Descriptor instead.
func (*SignedContributionAndProof) Descriptor() ([]byte, []int) {
	return file_proto_prysm_v2_sync_committee_proto_rawDescGZIP(), []int{3}
}

func (x *SignedContributionAndProof) GetMessage() *ContributionAndProof {
	if x != nil {
		return x.Message
	}
	return nil
}

func (x *SignedContributionAndProof) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

var File_proto_prysm_v2_sync_committee_proto protoreflect.FileDescriptor

var file_proto_prysm_v2_sync_committee_proto_rawDesc = []byte{
	0x0a, 0x23, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x76, 0x32,
	0x2f, 0x73, 0x79, 0x6e, 0x63, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e,
	0x70, 0x72, 0x79, 0x73, 0x6d, 0x2e, 0x76, 0x32, 0x1a, 0x1b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x65, 0x74, 0x68, 0x2f, 0x65, 0x78, 0x74, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x86, 0x02, 0x0a, 0x14, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x40,
	0x0a, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2c, 0x82, 0xb5,
	0x18, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79,
	0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x53, 0x6c, 0x6f, 0x74, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x74,
	0x12, 0x25, 0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x33, 0x32, 0x52, 0x09, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x5f, 0x0a, 0x0f, 0x76, 0x61, 0x6c, 0x69, 0x64,
	0x61, 0x74, 0x6f, 0x72, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65,
	0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0e, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x24, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18,
	0x02, 0x39, 0x36, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x22, 0xc1,
	0x02, 0x0a, 0x19, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65,
	0x43, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x40, 0x0a, 0x04,
	0x73, 0x6c, 0x6f, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x2c, 0x82, 0xb5, 0x18, 0x28,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d,
	0x61, 0x74, 0x69, 0x63, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x53, 0x6c, 0x6f, 0x74, 0x52, 0x04, 0x73, 0x6c, 0x6f, 0x74, 0x12, 0x25,
	0x0a, 0x0a, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x33, 0x32, 0x52, 0x09, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x52, 0x6f, 0x6f, 0x74, 0x12, 0x2d, 0x0a, 0x12, 0x73, 0x75, 0x62, 0x63, 0x6f, 0x6d, 0x6d,
	0x69, 0x74, 0x74, 0x65, 0x65, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x11, 0x73, 0x75, 0x62, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x49,
	0x6e, 0x64, 0x65, 0x78, 0x12, 0x66, 0x0a, 0x10, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x62, 0x69, 0x74, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x3b,
	0x8a, 0xb5, 0x18, 0x02, 0x31, 0x36, 0x82, 0xb5, 0x18, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c, 0x61,
	0x62, 0x73, 0x2f, 0x67, 0x6f, 0x2d, 0x62, 0x69, 0x74, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x2e, 0x42,
	0x69, 0x74, 0x76, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x31, 0x32, 0x38, 0x52, 0x0f, 0x61, 0x67, 0x67,
	0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x69, 0x74, 0x73, 0x12, 0x24, 0x0a, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x42,
	0x06, 0x8a, 0xb5, 0x18, 0x02, 0x39, 0x36, 0x52, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x22, 0xfc, 0x01, 0x0a, 0x14, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74,
	0x69, 0x6f, 0x6e, 0x41, 0x6e, 0x64, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x12, 0x61, 0x0a, 0x10, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x36, 0x82, 0xb5, 0x18, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c,
	0x61, 0x62, 0x73, 0x2f, 0x65, 0x74, 0x68, 0x32, 0x2d, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x56,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x0f, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x6f, 0x72, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x50,
	0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e,
	0x70, 0x72, 0x79, 0x73, 0x6d, 0x2e, 0x76, 0x32, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d,
	0x6d, 0x69, 0x74, 0x74, 0x65, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x2f, 0x0a, 0x0f, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x72,
	0x6f, 0x6f, 0x66, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x39,
	0x36, 0x52, 0x0e, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x6f,
	0x66, 0x22, 0x85, 0x01, 0x0a, 0x1a, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x74,
	0x72, 0x69, 0x62, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x41, 0x6e, 0x64, 0x50, 0x72, 0x6f, 0x6f, 0x66,
	0x12, 0x41, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x27, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x79,
	0x73, 0x6d, 0x2e, 0x76, 0x32, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x69, 0x62, 0x75, 0x74, 0x69,
	0x6f, 0x6e, 0x41, 0x6e, 0x64, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x24, 0x0a, 0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x06, 0x8a, 0xb5, 0x18, 0x02, 0x39, 0x36, 0x52, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0x87, 0x01, 0x0a, 0x15, 0x6f, 0x72,
	0x67, 0x2e, 0x65, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x79, 0x73, 0x6d,
	0x2e, 0x76, 0x32, 0x42, 0x12, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74,
	0x65, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x61, 0x74, 0x69, 0x63, 0x6c,
	0x61, 0x62, 0x73, 0x2f, 0x70, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x70, 0x72, 0x79, 0x73, 0x6d, 0x2f, 0x76, 0x32, 0x3b, 0x76, 0x32, 0xaa, 0x02, 0x11, 0x45, 0x74,
	0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x2e, 0x50, 0x72, 0x79, 0x73, 0x6d, 0x2e, 0x56, 0x32, 0xca,
	0x02, 0x11, 0x45, 0x74, 0x68, 0x65, 0x72, 0x65, 0x75, 0x6d, 0x5c, 0x50, 0x72, 0x79, 0x73, 0x6d,
	0x5c, 0x76, 0x32, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_prysm_v2_sync_committee_proto_rawDescOnce sync.Once
	file_proto_prysm_v2_sync_committee_proto_rawDescData = file_proto_prysm_v2_sync_committee_proto_rawDesc
)

func file_proto_prysm_v2_sync_committee_proto_rawDescGZIP() []byte {
	file_proto_prysm_v2_sync_committee_proto_rawDescOnce.Do(func() {
		file_proto_prysm_v2_sync_committee_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_prysm_v2_sync_committee_proto_rawDescData)
	})
	return file_proto_prysm_v2_sync_committee_proto_rawDescData
}

var file_proto_prysm_v2_sync_committee_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_prysm_v2_sync_committee_proto_goTypes = []interface{}{
	(*SyncCommitteeMessage)(nil),       // 0: ethereum.prysm.v2.SyncCommitteeMessage
	(*SyncCommitteeContribution)(nil),  // 1: ethereum.prysm.v2.SyncCommitteeContribution
	(*ContributionAndProof)(nil),       // 2: ethereum.prysm.v2.ContributionAndProof
	(*SignedContributionAndProof)(nil), // 3: ethereum.prysm.v2.SignedContributionAndProof
}
var file_proto_prysm_v2_sync_committee_proto_depIdxs = []int32{
	1, // 0: ethereum.prysm.v2.ContributionAndProof.contribution:type_name -> ethereum.prysm.v2.SyncCommitteeContribution
	2, // 1: ethereum.prysm.v2.SignedContributionAndProof.message:type_name -> ethereum.prysm.v2.ContributionAndProof
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_prysm_v2_sync_committee_proto_init() }
func file_proto_prysm_v2_sync_committee_proto_init() {
	if File_proto_prysm_v2_sync_committee_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_prysm_v2_sync_committee_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncCommitteeMessage); i {
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
		file_proto_prysm_v2_sync_committee_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncCommitteeContribution); i {
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
		file_proto_prysm_v2_sync_committee_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ContributionAndProof); i {
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
		file_proto_prysm_v2_sync_committee_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignedContributionAndProof); i {
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
			RawDescriptor: file_proto_prysm_v2_sync_committee_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_prysm_v2_sync_committee_proto_goTypes,
		DependencyIndexes: file_proto_prysm_v2_sync_committee_proto_depIdxs,
		MessageInfos:      file_proto_prysm_v2_sync_committee_proto_msgTypes,
	}.Build()
	File_proto_prysm_v2_sync_committee_proto = out.File
	file_proto_prysm_v2_sync_committee_proto_rawDesc = nil
	file_proto_prysm_v2_sync_committee_proto_goTypes = nil
	file_proto_prysm_v2_sync_committee_proto_depIdxs = nil
}
