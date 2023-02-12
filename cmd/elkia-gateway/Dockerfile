FROM golang:1.18 as debug

# Change path to Go source
WORKDIR /go/src/github.com/infinity-blackhole/elkia

# Pull dependencies
COPY go.mod go.sum ./
RUN --mount=type=secret,id=netrc,dst=/root/.netrc go mod download

# Build sources
COPY cmd cmd
COPY pkg pkg

FROM debug as build

# Build optimized binary
RUN go build -ldflags "-linkmode=external" -o /usr/bin/elkia-gateway ./cmd/api

FROM gcr.io/distroless/base-debian11

# Copy binary
COPY --from=build /usr/bin/elkia-gateway /usr/bin

# HTTP Function
EXPOSE 8080

ENTRYPOINT ["elkia-gateway"]
