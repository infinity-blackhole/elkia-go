// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: pkg/api/world/v1alpha1/world.proto

package v1alpha1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Lobby_LobbyInteract_FullMethodName   = "/shikanime.elkia.world.v1alpha1.Lobby/LobbyInteract"
	Lobby_CharacterAdd_FullMethodName    = "/shikanime.elkia.world.v1alpha1.Lobby/CharacterAdd"
	Lobby_CharacterRemove_FullMethodName = "/shikanime.elkia.world.v1alpha1.Lobby/CharacterRemove"
	Lobby_CharacterUpdate_FullMethodName = "/shikanime.elkia.world.v1alpha1.Lobby/CharacterUpdate"
	Lobby_CharacterList_FullMethodName   = "/shikanime.elkia.world.v1alpha1.Lobby/CharacterList"
)

// LobbyClient is the client API for Lobby service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LobbyClient interface {
	// LobbyInteract is used to interact with the lobby.
	LobbyInteract(ctx context.Context, opts ...grpc.CallOption) (Lobby_LobbyInteractClient, error)
	// CharacterAdd adds a new character to the world.
	CharacterAdd(ctx context.Context, in *CharacterAddRequest, opts ...grpc.CallOption) (*CharacterAddResponse, error)
	// CharacterRemove removes an existing character from the world.
	CharacterRemove(ctx context.Context, in *CharacterRemoveRequest, opts ...grpc.CallOption) (*CharacterRemoveResponse, error)
	// CharacterUpdate updates an existing character in the world.
	CharacterUpdate(ctx context.Context, in *CharacterUpdateRequest, opts ...grpc.CallOption) (*CharacterUpdateResponse, error)
	// CharacterList lists all characters in the world.
	CharacterList(ctx context.Context, in *CharacterListRequest, opts ...grpc.CallOption) (*CharacterListResponse, error)
}

type lobbyClient struct {
	cc grpc.ClientConnInterface
}

func NewLobbyClient(cc grpc.ClientConnInterface) LobbyClient {
	return &lobbyClient{cc}
}

func (c *lobbyClient) LobbyInteract(ctx context.Context, opts ...grpc.CallOption) (Lobby_LobbyInteractClient, error) {
	stream, err := c.cc.NewStream(ctx, &Lobby_ServiceDesc.Streams[0], Lobby_LobbyInteract_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &lobbyLobbyInteractClient{stream}
	return x, nil
}

type Lobby_LobbyInteractClient interface {
	Send(*LobbyInteractRequest) error
	Recv() (*LobbyInteractResponse, error)
	grpc.ClientStream
}

type lobbyLobbyInteractClient struct {
	grpc.ClientStream
}

func (x *lobbyLobbyInteractClient) Send(m *LobbyInteractRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *lobbyLobbyInteractClient) Recv() (*LobbyInteractResponse, error) {
	m := new(LobbyInteractResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *lobbyClient) CharacterAdd(ctx context.Context, in *CharacterAddRequest, opts ...grpc.CallOption) (*CharacterAddResponse, error) {
	out := new(CharacterAddResponse)
	err := c.cc.Invoke(ctx, Lobby_CharacterAdd_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lobbyClient) CharacterRemove(ctx context.Context, in *CharacterRemoveRequest, opts ...grpc.CallOption) (*CharacterRemoveResponse, error) {
	out := new(CharacterRemoveResponse)
	err := c.cc.Invoke(ctx, Lobby_CharacterRemove_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lobbyClient) CharacterUpdate(ctx context.Context, in *CharacterUpdateRequest, opts ...grpc.CallOption) (*CharacterUpdateResponse, error) {
	out := new(CharacterUpdateResponse)
	err := c.cc.Invoke(ctx, Lobby_CharacterUpdate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *lobbyClient) CharacterList(ctx context.Context, in *CharacterListRequest, opts ...grpc.CallOption) (*CharacterListResponse, error) {
	out := new(CharacterListResponse)
	err := c.cc.Invoke(ctx, Lobby_CharacterList_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LobbyServer is the server API for Lobby service.
// All implementations must embed UnimplementedLobbyServer
// for forward compatibility
type LobbyServer interface {
	// LobbyInteract is used to interact with the lobby.
	LobbyInteract(Lobby_LobbyInteractServer) error
	// CharacterAdd adds a new character to the world.
	CharacterAdd(context.Context, *CharacterAddRequest) (*CharacterAddResponse, error)
	// CharacterRemove removes an existing character from the world.
	CharacterRemove(context.Context, *CharacterRemoveRequest) (*CharacterRemoveResponse, error)
	// CharacterUpdate updates an existing character in the world.
	CharacterUpdate(context.Context, *CharacterUpdateRequest) (*CharacterUpdateResponse, error)
	// CharacterList lists all characters in the world.
	CharacterList(context.Context, *CharacterListRequest) (*CharacterListResponse, error)
	mustEmbedUnimplementedLobbyServer()
}

// UnimplementedLobbyServer must be embedded to have forward compatible implementations.
type UnimplementedLobbyServer struct {
}

func (UnimplementedLobbyServer) LobbyInteract(Lobby_LobbyInteractServer) error {
	return status.Errorf(codes.Unimplemented, "method LobbyInteract not implemented")
}
func (UnimplementedLobbyServer) CharacterAdd(context.Context, *CharacterAddRequest) (*CharacterAddResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CharacterAdd not implemented")
}
func (UnimplementedLobbyServer) CharacterRemove(context.Context, *CharacterRemoveRequest) (*CharacterRemoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CharacterRemove not implemented")
}
func (UnimplementedLobbyServer) CharacterUpdate(context.Context, *CharacterUpdateRequest) (*CharacterUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CharacterUpdate not implemented")
}
func (UnimplementedLobbyServer) CharacterList(context.Context, *CharacterListRequest) (*CharacterListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CharacterList not implemented")
}
func (UnimplementedLobbyServer) mustEmbedUnimplementedLobbyServer() {}

// UnsafeLobbyServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LobbyServer will
// result in compilation errors.
type UnsafeLobbyServer interface {
	mustEmbedUnimplementedLobbyServer()
}

func RegisterLobbyServer(s grpc.ServiceRegistrar, srv LobbyServer) {
	s.RegisterService(&Lobby_ServiceDesc, srv)
}

func _Lobby_LobbyInteract_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LobbyServer).LobbyInteract(&lobbyLobbyInteractServer{stream})
}

type Lobby_LobbyInteractServer interface {
	Send(*LobbyInteractResponse) error
	Recv() (*LobbyInteractRequest, error)
	grpc.ServerStream
}

type lobbyLobbyInteractServer struct {
	grpc.ServerStream
}

func (x *lobbyLobbyInteractServer) Send(m *LobbyInteractResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *lobbyLobbyInteractServer) Recv() (*LobbyInteractRequest, error) {
	m := new(LobbyInteractRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Lobby_CharacterAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CharacterAddRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LobbyServer).CharacterAdd(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Lobby_CharacterAdd_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LobbyServer).CharacterAdd(ctx, req.(*CharacterAddRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Lobby_CharacterRemove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CharacterRemoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LobbyServer).CharacterRemove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Lobby_CharacterRemove_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LobbyServer).CharacterRemove(ctx, req.(*CharacterRemoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Lobby_CharacterUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CharacterUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LobbyServer).CharacterUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Lobby_CharacterUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LobbyServer).CharacterUpdate(ctx, req.(*CharacterUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Lobby_CharacterList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CharacterListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LobbyServer).CharacterList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Lobby_CharacterList_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LobbyServer).CharacterList(ctx, req.(*CharacterListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Lobby_ServiceDesc is the grpc.ServiceDesc for Lobby service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Lobby_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "shikanime.elkia.world.v1alpha1.Lobby",
	HandlerType: (*LobbyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CharacterAdd",
			Handler:    _Lobby_CharacterAdd_Handler,
		},
		{
			MethodName: "CharacterRemove",
			Handler:    _Lobby_CharacterRemove_Handler,
		},
		{
			MethodName: "CharacterUpdate",
			Handler:    _Lobby_CharacterUpdate_Handler,
		},
		{
			MethodName: "CharacterList",
			Handler:    _Lobby_CharacterList_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LobbyInteract",
			Handler:       _Lobby_LobbyInteract_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "pkg/api/world/v1alpha1/world.proto",
}
