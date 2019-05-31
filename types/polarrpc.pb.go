// Code generated by protoc-gen-go. DO NOT EDIT.
// source: polarrpc.proto

package types

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Paginations struct {
	Ref                  []byte   `protobuf:"bytes,1,opt,name=ref,proto3" json:"ref,omitempty"`
	Size                 uint32   `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Paginations) Reset()         { *m = Paginations{} }
func (m *Paginations) String() string { return proto.CompactTextString(m) }
func (*Paginations) ProtoMessage()    {}
func (*Paginations) Descriptor() ([]byte, []int) {
	return fileDescriptor_9eae49c68867e2c2, []int{0}
}

func (m *Paginations) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Paginations.Unmarshal(m, b)
}
func (m *Paginations) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Paginations.Marshal(b, m, deterministic)
}
func (m *Paginations) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Paginations.Merge(m, src)
}
func (m *Paginations) XXX_Size() int {
	return xxx_messageInfo_Paginations.Size(m)
}
func (m *Paginations) XXX_DiscardUnknown() {
	xxx_messageInfo_Paginations.DiscardUnknown(m)
}

var xxx_messageInfo_Paginations proto.InternalMessageInfo

func (m *Paginations) GetRef() []byte {
	if m != nil {
		return m.Ref
	}
	return nil
}

func (m *Paginations) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

type PolarisPeerList struct {
	Total                uint32         `protobuf:"varint,1,opt,name=total,proto3" json:"total,omitempty"`
	HasNext              bool           `protobuf:"varint,2,opt,name=hasNext,proto3" json:"hasNext,omitempty"`
	Peers                []*PolarisPeer `protobuf:"bytes,3,rep,name=peers,proto3" json:"peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *PolarisPeerList) Reset()         { *m = PolarisPeerList{} }
func (m *PolarisPeerList) String() string { return proto.CompactTextString(m) }
func (*PolarisPeerList) ProtoMessage()    {}
func (*PolarisPeerList) Descriptor() ([]byte, []int) {
	return fileDescriptor_9eae49c68867e2c2, []int{1}
}

func (m *PolarisPeerList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PolarisPeerList.Unmarshal(m, b)
}
func (m *PolarisPeerList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PolarisPeerList.Marshal(b, m, deterministic)
}
func (m *PolarisPeerList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PolarisPeerList.Merge(m, src)
}
func (m *PolarisPeerList) XXX_Size() int {
	return xxx_messageInfo_PolarisPeerList.Size(m)
}
func (m *PolarisPeerList) XXX_DiscardUnknown() {
	xxx_messageInfo_PolarisPeerList.DiscardUnknown(m)
}

var xxx_messageInfo_PolarisPeerList proto.InternalMessageInfo

func (m *PolarisPeerList) GetTotal() uint32 {
	if m != nil {
		return m.Total
	}
	return 0
}

func (m *PolarisPeerList) GetHasNext() bool {
	if m != nil {
		return m.HasNext
	}
	return false
}

func (m *PolarisPeerList) GetPeers() []*PolarisPeer {
	if m != nil {
		return m.Peers
	}
	return nil
}

type PolarisPeer struct {
	Address   *PeerAddress `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Connected int64        `protobuf:"varint,2,opt,name=connected,proto3" json:"connected,omitempty"`
	// lastCheck contains unixtimestamp with nanoseconds precision
	LastCheck            int64    `protobuf:"varint,3,opt,name=lastCheck,proto3" json:"lastCheck,omitempty"`
	Verion               string   `protobuf:"bytes,4,opt,name=verion,proto3" json:"verion,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PolarisPeer) Reset()         { *m = PolarisPeer{} }
func (m *PolarisPeer) String() string { return proto.CompactTextString(m) }
func (*PolarisPeer) ProtoMessage()    {}
func (*PolarisPeer) Descriptor() ([]byte, []int) {
	return fileDescriptor_9eae49c68867e2c2, []int{2}
}

func (m *PolarisPeer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PolarisPeer.Unmarshal(m, b)
}
func (m *PolarisPeer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PolarisPeer.Marshal(b, m, deterministic)
}
func (m *PolarisPeer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PolarisPeer.Merge(m, src)
}
func (m *PolarisPeer) XXX_Size() int {
	return xxx_messageInfo_PolarisPeer.Size(m)
}
func (m *PolarisPeer) XXX_DiscardUnknown() {
	xxx_messageInfo_PolarisPeer.DiscardUnknown(m)
}

var xxx_messageInfo_PolarisPeer proto.InternalMessageInfo

func (m *PolarisPeer) GetAddress() *PeerAddress {
	if m != nil {
		return m.Address
	}
	return nil
}

func (m *PolarisPeer) GetConnected() int64 {
	if m != nil {
		return m.Connected
	}
	return 0
}

func (m *PolarisPeer) GetLastCheck() int64 {
	if m != nil {
		return m.LastCheck
	}
	return 0
}

func (m *PolarisPeer) GetVerion() string {
	if m != nil {
		return m.Verion
	}
	return ""
}

func init() {
	proto.RegisterType((*Paginations)(nil), "types.Paginations")
	proto.RegisterType((*PolarisPeerList)(nil), "types.PolarisPeerList")
	proto.RegisterType((*PolarisPeer)(nil), "types.PolarisPeer")
}

func init() { proto.RegisterFile("polarrpc.proto", fileDescriptor_9eae49c68867e2c2) }

var fileDescriptor_9eae49c68867e2c2 = []byte{
	// 401 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0xeb, 0x26, 0xc5, 0x93, 0x36, 0xc0, 0x08, 0x2a, 0xcb, 0x42, 0xc8, 0xf2, 0xc9, 0x07,
	0x94, 0xaa, 0xed, 0x09, 0x71, 0x22, 0xb9, 0x42, 0x15, 0x6d, 0x0e, 0x48, 0xdc, 0xb6, 0xf6, 0x60,
	0xaf, 0xe2, 0xee, 0xba, 0xbb, 0x93, 0x8a, 0xf2, 0x13, 0xfc, 0x08, 0x1f, 0x89, 0xbc, 0x76, 0x9a,
	0x20, 0x4e, 0x70, 0xf2, 0xbe, 0xf7, 0xe6, 0x79, 0xdf, 0xcc, 0x2c, 0xcc, 0x5a, 0xd3, 0x48, 0x6b,
	0xdb, 0x62, 0xde, 0x5a, 0xc3, 0x06, 0xc7, 0xfc, 0xd8, 0x92, 0x4b, 0x40, 0x9b, 0x92, 0x7a, 0x2a,
	0x89, 0x9e, 0xd4, 0xe4, 0xf4, 0x8e, 0xd8, 0xaa, 0x01, 0x65, 0xd7, 0x30, 0x5d, 0xc9, 0x4a, 0x69,
	0xc9, 0xca, 0x68, 0x87, 0x2f, 0x20, 0xb4, 0xf4, 0x2d, 0x0e, 0xd2, 0x20, 0x3f, 0x15, 0xdd, 0x11,
	0x11, 0x8e, 0x9d, 0xfa, 0x41, 0x71, 0x98, 0x06, 0xf9, 0x99, 0xf0, 0xe7, 0x6c, 0x03, 0xcf, 0x57,
	0xdd, 0x95, 0xca, 0xad, 0x88, 0xec, 0x27, 0xe5, 0x18, 0x5f, 0xc1, 0x98, 0x0d, 0xcb, 0xc6, 0x5b,
	0xcf, 0x44, 0x0f, 0x30, 0x86, 0x93, 0x5a, 0xba, 0x1b, 0xfa, 0xce, 0xf1, 0x51, 0x1a, 0xe4, 0xcf,
	0xc4, 0x0e, 0x62, 0x0e, 0xe3, 0x96, 0xc8, 0xba, 0x38, 0x4c, 0xc3, 0x7c, 0x7a, 0x85, 0x73, 0x9f,
	0x79, 0x7e, 0xf0, 0x5b, 0xd1, 0x17, 0x64, 0x3f, 0x03, 0x98, 0x1e, 0xd0, 0xf8, 0x0e, 0x4e, 0x64,
	0x59, 0x5a, 0x72, 0xce, 0xdf, 0x75, 0xe0, 0x25, 0xb2, 0x1f, 0x7b, 0x45, 0xec, 0x4a, 0xf0, 0x0d,
	0x44, 0x85, 0xd1, 0x9a, 0x0a, 0xa6, 0xd2, 0x67, 0x08, 0xc5, 0x9e, 0xe8, 0xd4, 0x46, 0x3a, 0x5e,
	0xd6, 0x54, 0x6c, 0x7c, 0x87, 0xa1, 0xd8, 0x13, 0x78, 0x0e, 0x93, 0x07, 0xb2, 0xca, 0xe8, 0xf8,
	0x38, 0x0d, 0xf2, 0x48, 0x0c, 0xe8, 0xea, 0xd7, 0x11, 0xbc, 0x1c, 0x12, 0x89, 0xd5, 0x72, 0x4d,
	0xf6, 0x41, 0x15, 0x84, 0x97, 0x10, 0xdd, 0x98, 0x92, 0xd6, 0x2c, 0x99, 0x70, 0x36, 0x64, 0xea,
	0x18, 0x41, 0xf7, 0xc9, 0x2e, 0xe3, 0x5a, 0xe9, 0xaa, 0xa1, 0xc5, 0x23, 0x93, 0xcb, 0x46, 0x78,
	0x09, 0x93, 0xcf, 0x7e, 0x19, 0xf8, 0x7a, 0xd0, 0x7b, 0xe8, 0x04, 0xdd, 0x6f, 0xc9, 0x71, 0x32,
	0xfb, 0x93, 0xce, 0x46, 0xf8, 0x01, 0xa6, 0xcb, 0xad, 0xb5, 0xa4, 0xd9, 0x8f, 0xfd, 0xa9, 0xf7,
	0xfd, 0x0e, 0x93, 0xf3, 0xbf, 0x67, 0xd9, 0xd5, 0x66, 0x23, 0x7c, 0x0f, 0xd1, 0x97, 0x5a, 0x31,
	0xfd, 0x9f, 0x75, 0xd1, 0xc8, 0x62, 0xf3, 0xef, 0xd6, 0x45, 0xfa, 0xf5, 0x6d, 0xa5, 0xb8, 0xde,
	0xde, 0xce, 0x0b, 0x73, 0x77, 0x21, 0xc9, 0x56, 0x46, 0x99, 0xfe, 0x7b, 0xe1, 0x3d, 0xb7, 0x13,
	0xff, 0x16, 0xaf, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0x02, 0xd2, 0x28, 0x55, 0xc9, 0x02, 0x00,
	0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PolarisRPCServiceClient is the client API for PolarisRPCService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PolarisRPCServiceClient interface {
	// Returns the current state of this node
	NodeState(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*SingleBytes, error)
	// Returns node metrics according to request
	Metric(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*Metrics, error)
	CurrentList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error)
	WhiteList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error)
	BlackList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error)
}

type polarisRPCServiceClient struct {
	cc *grpc.ClientConn
}

func NewPolarisRPCServiceClient(cc *grpc.ClientConn) PolarisRPCServiceClient {
	return &polarisRPCServiceClient{cc}
}

func (c *polarisRPCServiceClient) NodeState(ctx context.Context, in *NodeReq, opts ...grpc.CallOption) (*SingleBytes, error) {
	out := new(SingleBytes)
	err := c.cc.Invoke(ctx, "/types.PolarisRPCService/NodeState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *polarisRPCServiceClient) Metric(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*Metrics, error) {
	out := new(Metrics)
	err := c.cc.Invoke(ctx, "/types.PolarisRPCService/Metric", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *polarisRPCServiceClient) CurrentList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error) {
	out := new(PolarisPeerList)
	err := c.cc.Invoke(ctx, "/types.PolarisRPCService/CurrentList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *polarisRPCServiceClient) WhiteList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error) {
	out := new(PolarisPeerList)
	err := c.cc.Invoke(ctx, "/types.PolarisRPCService/WhiteList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *polarisRPCServiceClient) BlackList(ctx context.Context, in *Paginations, opts ...grpc.CallOption) (*PolarisPeerList, error) {
	out := new(PolarisPeerList)
	err := c.cc.Invoke(ctx, "/types.PolarisRPCService/BlackList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PolarisRPCServiceServer is the server API for PolarisRPCService service.
type PolarisRPCServiceServer interface {
	// Returns the current state of this node
	NodeState(context.Context, *NodeReq) (*SingleBytes, error)
	// Returns node metrics according to request
	Metric(context.Context, *MetricsRequest) (*Metrics, error)
	CurrentList(context.Context, *Paginations) (*PolarisPeerList, error)
	WhiteList(context.Context, *Paginations) (*PolarisPeerList, error)
	BlackList(context.Context, *Paginations) (*PolarisPeerList, error)
}

func RegisterPolarisRPCServiceServer(s *grpc.Server, srv PolarisRPCServiceServer) {
	s.RegisterService(&_PolarisRPCService_serviceDesc, srv)
}

func _PolarisRPCService_NodeState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PolarisRPCServiceServer).NodeState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/types.PolarisRPCService/NodeState",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PolarisRPCServiceServer).NodeState(ctx, req.(*NodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _PolarisRPCService_Metric_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PolarisRPCServiceServer).Metric(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/types.PolarisRPCService/Metric",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PolarisRPCServiceServer).Metric(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PolarisRPCService_CurrentList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Paginations)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PolarisRPCServiceServer).CurrentList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/types.PolarisRPCService/CurrentList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PolarisRPCServiceServer).CurrentList(ctx, req.(*Paginations))
	}
	return interceptor(ctx, in, info, handler)
}

func _PolarisRPCService_WhiteList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Paginations)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PolarisRPCServiceServer).WhiteList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/types.PolarisRPCService/WhiteList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PolarisRPCServiceServer).WhiteList(ctx, req.(*Paginations))
	}
	return interceptor(ctx, in, info, handler)
}

func _PolarisRPCService_BlackList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Paginations)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PolarisRPCServiceServer).BlackList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/types.PolarisRPCService/BlackList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PolarisRPCServiceServer).BlackList(ctx, req.(*Paginations))
	}
	return interceptor(ctx, in, info, handler)
}

var _PolarisRPCService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "types.PolarisRPCService",
	HandlerType: (*PolarisRPCServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NodeState",
			Handler:    _PolarisRPCService_NodeState_Handler,
		},
		{
			MethodName: "Metric",
			Handler:    _PolarisRPCService_Metric_Handler,
		},
		{
			MethodName: "CurrentList",
			Handler:    _PolarisRPCService_CurrentList_Handler,
		},
		{
			MethodName: "WhiteList",
			Handler:    _PolarisRPCService_WhiteList_Handler,
		},
		{
			MethodName: "BlackList",
			Handler:    _PolarisRPCService_BlackList_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "polarrpc.proto",
}