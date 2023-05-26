%%%-------------------------------------------------------------------
%% @doc Client module for grpc service io.elkia.fleet.v1alpha1.Cluster.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_fleet_v_1alpha_1_cluster_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'io.elkia.fleet.v1alpha1.Cluster').
-define(PROTO_MODULE, 'elkia_v1alpha1_fleet_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc Unary RPC
-spec member_add(elkia_v1alpha1_fleet_pb:member_add_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_add(Input) ->
    member_add(ctx:new(), Input, #{}).

-spec member_add(ctx:t() | elkia_v1alpha1_fleet_pb:member_add_request(), elkia_v1alpha1_fleet_pb:member_add_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_add(Ctx, Input) when ?is_ctx(Ctx) ->
    member_add(Ctx, Input, #{});
member_add(Input, Options) ->
    member_add(ctx:new(), Input, Options).

-spec member_add(ctx:t(), elkia_v1alpha1_fleet_pb:member_add_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_add(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Cluster/MemberAdd">>, Input, ?DEF(member_add_request, member_add_response, <<"io.elkia.fleet.v1alpha1.MemberAddRequest">>), Options).

%% @doc Unary RPC
-spec member_remove(elkia_v1alpha1_fleet_pb:member_remove_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_remove(Input) ->
    member_remove(ctx:new(), Input, #{}).

-spec member_remove(ctx:t() | elkia_v1alpha1_fleet_pb:member_remove_request(), elkia_v1alpha1_fleet_pb:member_remove_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_remove(Ctx, Input) when ?is_ctx(Ctx) ->
    member_remove(Ctx, Input, #{});
member_remove(Input, Options) ->
    member_remove(ctx:new(), Input, Options).

-spec member_remove(ctx:t(), elkia_v1alpha1_fleet_pb:member_remove_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_remove(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Cluster/MemberRemove">>, Input, ?DEF(member_remove_request, member_remove_response, <<"io.elkia.fleet.v1alpha1.MemberRemoveRequest">>), Options).

%% @doc Unary RPC
-spec member_update(elkia_v1alpha1_fleet_pb:member_update_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_update(Input) ->
    member_update(ctx:new(), Input, #{}).

-spec member_update(ctx:t() | elkia_v1alpha1_fleet_pb:member_update_request(), elkia_v1alpha1_fleet_pb:member_update_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_update(Ctx, Input) when ?is_ctx(Ctx) ->
    member_update(Ctx, Input, #{});
member_update(Input, Options) ->
    member_update(ctx:new(), Input, Options).

-spec member_update(ctx:t(), elkia_v1alpha1_fleet_pb:member_update_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_update(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Cluster/MemberUpdate">>, Input, ?DEF(member_update_request, member_update_response, <<"io.elkia.fleet.v1alpha1.MemberUpdateRequest">>), Options).

%% @doc Unary RPC
-spec member_list(elkia_v1alpha1_fleet_pb:member_list_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_list(Input) ->
    member_list(ctx:new(), Input, #{}).

-spec member_list(ctx:t() | elkia_v1alpha1_fleet_pb:member_list_request(), elkia_v1alpha1_fleet_pb:member_list_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_list(Ctx, Input) when ?is_ctx(Ctx) ->
    member_list(Ctx, Input, #{});
member_list(Input, Options) ->
    member_list(ctx:new(), Input, Options).

-spec member_list(ctx:t(), elkia_v1alpha1_fleet_pb:member_list_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
member_list(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Cluster/MemberList">>, Input, ?DEF(member_list_request, member_list_response, <<"io.elkia.fleet.v1alpha1.MemberListRequest">>), Options).

