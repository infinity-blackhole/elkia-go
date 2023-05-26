%%%-------------------------------------------------------------------
%% @doc Client module for grpc service io.elkia.eventing.v1alpha1.Gateway.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_eventing_v_1alpha_1_gateway_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'io.elkia.eventing.v1alpha1.Gateway').
-define(PROTO_MODULE, 'elkia_v1alpha1_eventing_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc 
-spec channel_interact() ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_interact() ->
    channel_interact(ctx:new(), #{}).

-spec channel_interact(ctx:t() | grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_interact(Ctx) when ?is_ctx(Ctx) ->
    channel_interact(Ctx, #{});
channel_interact(Options) ->
    channel_interact(ctx:new(), Options).

-spec channel_interact(ctx:t(), grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_interact(Ctx, Options) ->
    grpcbox_client:stream(Ctx, <<"/io.elkia.eventing.v1alpha1.Gateway/ChannelInteract">>, ?DEF(channel_interact_request, channel_interact_response, <<"io.elkia.eventing.v1alpha1.ChannelInteractRequest">>), Options).

%% @doc 
-spec channel_watch(elkia_v1alpha1_eventing_pb:channel_watch_request()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_watch(Input) ->
    channel_watch(ctx:new(), Input, #{}).

-spec channel_watch(ctx:t() | elkia_v1alpha1_eventing_pb:channel_watch_request(), elkia_v1alpha1_eventing_pb:channel_watch_request() | grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_watch(Ctx, Input) when ?is_ctx(Ctx) ->
    channel_watch(Ctx, Input, #{});
channel_watch(Input, Options) ->
    channel_watch(ctx:new(), Input, Options).

-spec channel_watch(ctx:t(), elkia_v1alpha1_eventing_pb:channel_watch_request(), grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
channel_watch(Ctx, Input, Options) ->
    grpcbox_client:stream(Ctx, <<"/io.elkia.eventing.v1alpha1.Gateway/ChannelWatch">>, Input, ?DEF(channel_watch_request, channel_interact_response, <<"io.elkia.eventing.v1alpha1.ChannelWatchRequest">>), Options).

