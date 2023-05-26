%%%-------------------------------------------------------------------
%% @doc Client module for grpc service io.elkia.world.v1alpha1.Bridge.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_world_v_1alpha_1_bridge_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'io.elkia.world.v1alpha1.Bridge').
-define(PROTO_MODULE, 'elkia_v1alpha1_world_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc Unary RPC
-spec character_add(elkia_v1alpha1_world_pb:character_add_request()) ->
    {ok, elkia_v1alpha1_world_pb:character_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_add(Input) ->
    character_add(ctx:new(), Input, #{}).

-spec character_add(ctx:t() | elkia_v1alpha1_world_pb:character_add_request(), elkia_v1alpha1_world_pb:character_add_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_add(Ctx, Input) when ?is_ctx(Ctx) ->
    character_add(Ctx, Input, #{});
character_add(Input, Options) ->
    character_add(ctx:new(), Input, Options).

-spec character_add(ctx:t(), elkia_v1alpha1_world_pb:character_add_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_add_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_add(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.world.v1alpha1.Bridge/CharacterAdd">>, Input, ?DEF(character_add_request, character_add_response, <<"io.elkia.world.v1alpha1.CharacterAddRequest">>), Options).

%% @doc Unary RPC
-spec character_remove(elkia_v1alpha1_world_pb:character_remove_request()) ->
    {ok, elkia_v1alpha1_world_pb:character_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_remove(Input) ->
    character_remove(ctx:new(), Input, #{}).

-spec character_remove(ctx:t() | elkia_v1alpha1_world_pb:character_remove_request(), elkia_v1alpha1_world_pb:character_remove_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_remove(Ctx, Input) when ?is_ctx(Ctx) ->
    character_remove(Ctx, Input, #{});
character_remove(Input, Options) ->
    character_remove(ctx:new(), Input, Options).

-spec character_remove(ctx:t(), elkia_v1alpha1_world_pb:character_remove_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_remove_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_remove(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.world.v1alpha1.Bridge/CharacterRemove">>, Input, ?DEF(character_remove_request, character_remove_response, <<"io.elkia.world.v1alpha1.CharacterRemoveRequest">>), Options).

%% @doc Unary RPC
-spec character_update(elkia_v1alpha1_world_pb:character_update_request()) ->
    {ok, elkia_v1alpha1_world_pb:character_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_update(Input) ->
    character_update(ctx:new(), Input, #{}).

-spec character_update(ctx:t() | elkia_v1alpha1_world_pb:character_update_request(), elkia_v1alpha1_world_pb:character_update_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_update(Ctx, Input) when ?is_ctx(Ctx) ->
    character_update(Ctx, Input, #{});
character_update(Input, Options) ->
    character_update(ctx:new(), Input, Options).

-spec character_update(ctx:t(), elkia_v1alpha1_world_pb:character_update_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_update_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_update(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.world.v1alpha1.Bridge/CharacterUpdate">>, Input, ?DEF(character_update_request, character_update_response, <<"io.elkia.world.v1alpha1.CharacterUpdateRequest">>), Options).

%% @doc Unary RPC
-spec character_list(elkia_v1alpha1_world_pb:character_list_request()) ->
    {ok, elkia_v1alpha1_world_pb:character_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_list(Input) ->
    character_list(ctx:new(), Input, #{}).

-spec character_list(ctx:t() | elkia_v1alpha1_world_pb:character_list_request(), elkia_v1alpha1_world_pb:character_list_request() | grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_list(Ctx, Input) when ?is_ctx(Ctx) ->
    character_list(Ctx, Input, #{});
character_list(Input, Options) ->
    character_list(ctx:new(), Input, Options).

-spec character_list(ctx:t(), elkia_v1alpha1_world_pb:character_list_request(), grpcbox_client:options()) ->
    {ok, elkia_v1alpha1_world_pb:character_list_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
character_list(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.world.v1alpha1.Bridge/CharacterList">>, Input, ?DEF(character_list_request, character_list_response, <<"io.elkia.world.v1alpha1.CharacterListRequest">>), Options).

