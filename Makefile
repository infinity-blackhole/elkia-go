SYNC_ELKIA_PACKAGES = \
	sync-elkia-erlang \
	sync-elkia-go

.PHONY: sync
sync: $(SYNC_ELKIA_PACKAGES)

.PHONY: sync-elkia-erlang
sync-elkia-erlang:
	@cp elkia/fleet/v1alpha1/fleet.proto elkia-erlang/protos/elkia_v1alpha1_fleet.proto
	@cp elkia/eventing/v1alpha1/eventing.proto elkia-erlang/protos/elkia_v1alpha1_eventing.proto
	@cp elkia/world/v1alpha1/world.proto elkia-erlang/protos/elkia_v1alpha1_world.proto
	@cp LICENSE elkia-erlang

.PHONY: sync-elkia-go
sync-elkia-go:
	@cp LICENSE elkia-go
