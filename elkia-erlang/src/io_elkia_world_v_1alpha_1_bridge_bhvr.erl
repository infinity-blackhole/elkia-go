%%%-------------------------------------------------------------------
%% @doc Behaviour to implement for grpc service io.elkia.world.v1alpha1.Bridge.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_world_v_1alpha_1_bridge_bhvr).

%% Unary RPC
-callback character_add(ctx:t(), v1alpha1_world_pb:character_add_request()) ->
    {ok, v1alpha1_world_pb:character_add_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback character_remove(ctx:t(), v1alpha1_world_pb:character_remove_request()) ->
    {ok, v1alpha1_world_pb:character_remove_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback character_update(ctx:t(), v1alpha1_world_pb:character_update_request()) ->
    {ok, v1alpha1_world_pb:character_update_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback character_list(ctx:t(), v1alpha1_world_pb:character_list_request()) ->
    {ok, v1alpha1_world_pb:character_list_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

