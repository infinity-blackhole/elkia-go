%%%-------------------------------------------------------------------
%% @doc Client module for grpc service io.elkia.fleet.v1alpha1.Presence.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_fleet_v_1alpha_1_presence_client).

-compile(export_all).
-compile(nowarn_export_all).

-include_lib("grpcbox/include/grpcbox.hrl").

-define(is_ctx(Ctx), is_tuple(Ctx) andalso element(1, Ctx) =:= ctx).

-define(SERVICE, 'io.elkia.fleet.v1alpha1.Presence').
-define(PROTO_MODULE, 'v1alpha1_fleet_pb').
-define(MARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:encode_msg(I, T) end).
-define(UNMARSHAL_FUN(T), fun(I) -> ?PROTO_MODULE:decode_msg(I, T) end).
-define(DEF(Input, Output, MessageType), #grpcbox_def{service=?SERVICE,
                                                      message_type=MessageType,
                                                      marshal_fun=?MARSHAL_FUN(Input),
                                                      unmarshal_fun=?UNMARSHAL_FUN(Output)}).

%% @doc Unary RPC
-spec auth_login(v1alpha1_fleet_pb:auth_login_request()) ->
    {ok, v1alpha1_fleet_pb:auth_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_login(Input) ->
    auth_login(ctx:new(), Input, #{}).

-spec auth_login(ctx:t() | v1alpha1_fleet_pb:auth_login_request(), v1alpha1_fleet_pb:auth_login_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_login(Ctx, Input) when ?is_ctx(Ctx) ->
    auth_login(Ctx, Input, #{});
auth_login(Input, Options) ->
    auth_login(ctx:new(), Input, Options).

-spec auth_login(ctx:t(), v1alpha1_fleet_pb:auth_login_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_login(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/AuthLogin">>, Input, ?DEF(auth_login_request, auth_login_response, <<"io.elkia.fleet.v1alpha1.AuthLoginRequest">>), Options).

%% @doc Unary RPC
-spec auth_refresh_login(v1alpha1_fleet_pb:auth_refresh_login_request()) ->
    {ok, v1alpha1_fleet_pb:auth_refresh_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_refresh_login(Input) ->
    auth_refresh_login(ctx:new(), Input, #{}).

-spec auth_refresh_login(ctx:t() | v1alpha1_fleet_pb:auth_refresh_login_request(), v1alpha1_fleet_pb:auth_refresh_login_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_refresh_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_refresh_login(Ctx, Input) when ?is_ctx(Ctx) ->
    auth_refresh_login(Ctx, Input, #{});
auth_refresh_login(Input, Options) ->
    auth_refresh_login(ctx:new(), Input, Options).

-spec auth_refresh_login(ctx:t(), v1alpha1_fleet_pb:auth_refresh_login_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_refresh_login_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_refresh_login(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/AuthRefreshLogin">>, Input, ?DEF(auth_refresh_login_request, auth_refresh_login_response, <<"io.elkia.fleet.v1alpha1.AuthRefreshLoginRequest">>), Options).

%% @doc Unary RPC
-spec auth_handoff(v1alpha1_fleet_pb:auth_handoff_request()) ->
    {ok, v1alpha1_fleet_pb:auth_handoff_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_handoff(Input) ->
    auth_handoff(ctx:new(), Input, #{}).

-spec auth_handoff(ctx:t() | v1alpha1_fleet_pb:auth_handoff_request(), v1alpha1_fleet_pb:auth_handoff_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_handoff_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_handoff(Ctx, Input) when ?is_ctx(Ctx) ->
    auth_handoff(Ctx, Input, #{});
auth_handoff(Input, Options) ->
    auth_handoff(ctx:new(), Input, Options).

-spec auth_handoff(ctx:t(), v1alpha1_fleet_pb:auth_handoff_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_handoff_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_handoff(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/AuthHandoff">>, Input, ?DEF(auth_handoff_request, auth_handoff_response, <<"io.elkia.fleet.v1alpha1.AuthHandoffRequest">>), Options).

%% @doc Unary RPC
-spec auth_logout(v1alpha1_fleet_pb:auth_logout_request()) ->
    {ok, v1alpha1_fleet_pb:auth_logout_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_logout(Input) ->
    auth_logout(ctx:new(), Input, #{}).

-spec auth_logout(ctx:t() | v1alpha1_fleet_pb:auth_logout_request(), v1alpha1_fleet_pb:auth_logout_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_logout_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_logout(Ctx, Input) when ?is_ctx(Ctx) ->
    auth_logout(Ctx, Input, #{});
auth_logout(Input, Options) ->
    auth_logout(ctx:new(), Input, Options).

-spec auth_logout(ctx:t(), v1alpha1_fleet_pb:auth_logout_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:auth_logout_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
auth_logout(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/AuthLogout">>, Input, ?DEF(auth_logout_request, auth_logout_response, <<"io.elkia.fleet.v1alpha1.AuthLogoutRequest">>), Options).

%% @doc Unary RPC
-spec session_get(v1alpha1_fleet_pb:session_get_request()) ->
    {ok, v1alpha1_fleet_pb:session_get_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_get(Input) ->
    session_get(ctx:new(), Input, #{}).

-spec session_get(ctx:t() | v1alpha1_fleet_pb:session_get_request(), v1alpha1_fleet_pb:session_get_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_get_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_get(Ctx, Input) when ?is_ctx(Ctx) ->
    session_get(Ctx, Input, #{});
session_get(Input, Options) ->
    session_get(ctx:new(), Input, Options).

-spec session_get(ctx:t(), v1alpha1_fleet_pb:session_get_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_get_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_get(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/SessionGet">>, Input, ?DEF(session_get_request, session_get_response, <<"io.elkia.fleet.v1alpha1.SessionGetRequest">>), Options).

%% @doc Unary RPC
-spec session_put(v1alpha1_fleet_pb:session_put_request()) ->
    {ok, v1alpha1_fleet_pb:session_put_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_put(Input) ->
    session_put(ctx:new(), Input, #{}).

-spec session_put(ctx:t() | v1alpha1_fleet_pb:session_put_request(), v1alpha1_fleet_pb:session_put_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_put_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_put(Ctx, Input) when ?is_ctx(Ctx) ->
    session_put(Ctx, Input, #{});
session_put(Input, Options) ->
    session_put(ctx:new(), Input, Options).

-spec session_put(ctx:t(), v1alpha1_fleet_pb:session_put_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_put_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_put(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/SessionPut">>, Input, ?DEF(session_put_request, session_put_response, <<"io.elkia.fleet.v1alpha1.SessionPutRequest">>), Options).

%% @doc Unary RPC
-spec session_delete(v1alpha1_fleet_pb:session_delete_request()) ->
    {ok, v1alpha1_fleet_pb:session_delete_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_delete(Input) ->
    session_delete(ctx:new(), Input, #{}).

-spec session_delete(ctx:t() | v1alpha1_fleet_pb:session_delete_request(), v1alpha1_fleet_pb:session_delete_request() | grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_delete_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_delete(Ctx, Input) when ?is_ctx(Ctx) ->
    session_delete(Ctx, Input, #{});
session_delete(Input, Options) ->
    session_delete(ctx:new(), Input, Options).

-spec session_delete(ctx:t(), v1alpha1_fleet_pb:session_delete_request(), grpcbox_client:options()) ->
    {ok, v1alpha1_fleet_pb:session_delete_response(), grpcbox:metadata()} | grpcbox_stream:grpc_error_response() | {error, any()}.
session_delete(Ctx, Input, Options) ->
    grpcbox_client:unary(Ctx, <<"/io.elkia.fleet.v1alpha1.Presence/SessionDelete">>, Input, ?DEF(session_delete_request, session_delete_response, <<"io.elkia.fleet.v1alpha1.SessionDeleteRequest">>), Options).

