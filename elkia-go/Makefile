.PHONY: gen
gen:
	@protoc \
		--go_out=. \
		--go_opt=paths=import\
		--go-grpc_out=. \
		--go-grpc_opt=paths=import\
		elkia/fleet/v1alpha1/fleet.proto
	@protoc \
		--go_out=. \
		--go_opt=paths=import\
		--go-grpc_out=. \
		elkia/eventing/v1alpha1/eventing.proto
	@protoc \
		--go_out=. \
		--go_opt=paths=import\
		--go-grpc_out=. \
		--go-grpc_opt=paths=import\
		elkia/world/v1alpha1/world.proto
	@cp LICENSE elkia-go