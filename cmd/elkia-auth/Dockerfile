FROM golang:1.19 AS base

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

FROM base AS build

# Enable all tracing
ENV GOTRACEBACK=all

# Build binary
ARG SKAFFOLD_GO_GCFLAGS
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go build \
    -gcflags="${SKAFFOLD_GO_GCFLAGS}" \
    -ldflags "-linkmode=external" \
    -o /usr/local/bin/elkia-auth \
    ./cmd/elkia-auth

CMD ["/usr/local/bin/elkia-auth"]

FROM gcr.io/distroless/base-debian11:nonroot

# Connect container image to source code
LABEL org.opencontainers.image.source https://github.com/infinity-blackhole/elkia-go

# Copy binary
COPY --from=build /usr/local/bin/elkia-auth /usr/local/bin/elkia-auth

# gRPC API
EXPOSE 8080

ENTRYPOINT ["elkia-auth"]
