FROM golang:1.19 as debug

# Change path to Go source
WORKDIR /go/src/github.com/infinity-blackhole/elkia

# Pull dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download

# Build sources
COPY cmd cmd
COPY internal internal
COPY pkg pkg

FROM debug as build

# Build optimized binary
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go build \
    -ldflags "-linkmode=external" \
    -o /usr/bin/elkia-gateway \
    ./cmd/elkia-gateway

FROM gcr.io/distroless/base-debian11:nonroot

# Copy binary
COPY --from=build /usr/bin/elkia-gateway /usr/bin

# NosTale gateway
EXPOSE 4124

ENTRYPOINT ["elkia-gateway"]
