%%%-------------------------------------------------------------------
%% @doc Behaviour to implement for grpc service io.elkia.fleet.v1alpha1.Presence.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_fleet_v_1alpha_1_presence_bhvr).

%% Unary RPC
-callback auth_login(ctx:t(), v1alpha1_fleet_pb:auth_login_request()) ->
    {ok, v1alpha1_fleet_pb:auth_login_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback auth_refresh_login(ctx:t(), v1alpha1_fleet_pb:auth_refresh_login_request()) ->
    {ok, v1alpha1_fleet_pb:auth_refresh_login_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback auth_handoff(ctx:t(), v1alpha1_fleet_pb:auth_handoff_request()) ->
    {ok, v1alpha1_fleet_pb:auth_handoff_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback auth_logout(ctx:t(), v1alpha1_fleet_pb:auth_logout_request()) ->
    {ok, v1alpha1_fleet_pb:auth_logout_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback session_get(ctx:t(), v1alpha1_fleet_pb:session_get_request()) ->
    {ok, v1alpha1_fleet_pb:session_get_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback session_put(ctx:t(), v1alpha1_fleet_pb:session_put_request()) ->
    {ok, v1alpha1_fleet_pb:session_put_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback session_delete(ctx:t(), v1alpha1_fleet_pb:session_delete_request()) ->
    {ok, v1alpha1_fleet_pb:session_delete_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

