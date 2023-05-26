%%%-------------------------------------------------------------------
%% @doc Behaviour to implement for grpc service io.elkia.fleet.v1alpha1.Cluster.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_fleet_v_1alpha_1_cluster_bhvr).

%% Unary RPC
-callback member_add(ctx:t(), elkia_v1alpha1_fleet_pb:member_add_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_add_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback member_remove(ctx:t(), elkia_v1alpha1_fleet_pb:member_remove_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_remove_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback member_update(ctx:t(), elkia_v1alpha1_fleet_pb:member_update_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_update_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

%% Unary RPC
-callback member_list(ctx:t(), elkia_v1alpha1_fleet_pb:member_list_request()) ->
    {ok, elkia_v1alpha1_fleet_pb:member_list_response(), ctx:t()} | grpcbox_stream:grpc_error_response().

