%%%-------------------------------------------------------------------
%% @doc Client module for grpc service io.elkia.eventing.v1alpha1.Auth.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_eventing_v_1alpha_1_auth_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'io.elkia.eventing.v1alpha1.Auth').
-define(PROTO_MODULE, 'elkia_v1alpha1_eventing_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc 
-spec auth_interact() ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_interact() ->
    auth_interact(ctx:new(), #{}).

-spec auth_interact(ctx:t() | grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_interact(Ctx) when ?is_ctx(Ctx) ->
    auth_interact(Ctx, #{});
auth_interact(Options) ->
    auth_interact(ctx:new(), Options).

-spec auth_interact(ctx:t(), grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_interact(Ctx, Options) ->
    grpcbox_client:stream(Ctx, <<"/io.elkia.eventing.v1alpha1.Auth/AuthInteract">>, ?DEF(auth_interact_request, auth_interact_response, <<"io.elkia.eventing.v1alpha1.AuthInteractRequest">>), Options).

%% @doc 
-spec auth_watch(elkia_v1alpha1_eventing_pb:auth_watch_request()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_watch(Input) ->
    auth_watch(ctx:new(), Input, #{}).

-spec auth_watch(ctx:t() | elkia_v1alpha1_eventing_pb:auth_watch_request(), elkia_v1alpha1_eventing_pb:auth_watch_request() | grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_watch(Ctx, Input) when ?is_ctx(Ctx) ->
    auth_watch(Ctx, Input, #{});
auth_watch(Input, Options) ->
    auth_watch(ctx:new(), Input, Options).

-spec auth_watch(ctx:t(), elkia_v1alpha1_eventing_pb:auth_watch_request(), grpcbox_client:options()) ->
    {ok, grpcbox_client:stream()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_watch(Ctx, Input, Options) ->
    grpcbox_client:stream(Ctx, <<"/io.elkia.eventing.v1alpha1.Auth/AuthWatch">>, Input, ?DEF(auth_watch_request, auth_interact_response, <<"io.elkia.eventing.v1alpha1.AuthWatchRequest">>), Options).

