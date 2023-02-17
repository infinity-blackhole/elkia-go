FROM golang:1.19 as base

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

FROM base as build

# Build binary
ARG SKAFFOLD_GO_GCFLAGS
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build/ \
    go build \
    -gcflags="${SKAFFOLD_GO_GCFLAGS}" \
    -ldflags "-linkmode=external" \
    -o /usr/bin/elkia-authserver \
    ./cmd/elkia-authserver

# NosTale auth server
EXPOSE 4123

ENTRYPOINT ["elkia-authserver"]

FROM gcr.io/distroless/base-debian11:nonroot

# Copy binary
COPY --from=build /usr/bin/elkia-authserver /usr/bin

# NosTale auth server
EXPOSE 4123

ENTRYPOINT ["elkia-authserver"]
