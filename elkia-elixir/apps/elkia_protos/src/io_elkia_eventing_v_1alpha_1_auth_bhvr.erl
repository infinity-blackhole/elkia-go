%%%-------------------------------------------------------------------
%% @doc Behaviour to implement for grpc service io.elkia.eventing.v1alpha1.Auth.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_eventing_v_1alpha_1_auth_bhvr).

%% 
-callback auth_interact(reference(), grpcbox_stream:t()) ->
    ok | grpcbox_stream:grpc_error_response().

%% 
-callback auth_watch(elkia_v1alpha1_eventing_pb:auth_watch_request(), grpcbox_stream:t()) ->
    ok | grpcbox_stream:grpc_error_response().

