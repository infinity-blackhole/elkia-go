%%%-------------------------------------------------------------------
%% @doc Behaviour to implement for grpc service io.elkia.eventing.v1alpha1.Gateway.
%% @end
%%%-------------------------------------------------------------------

%% this module was generated and should not be modified manually

-module(io_elkia_eventing_v_1alpha_1_gateway_bhvr).

%% 
-callback channel_interact(reference(), grpcbox_stream:t()) ->
    ok | grpcbox_stream:grpc_error_response().

%% 
-callback channel_watch(v1alpha1_eventing_pb:channel_watch_request(), grpcbox_stream:t()) ->
    ok | grpcbox_stream:grpc_error_response().

