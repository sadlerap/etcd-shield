# Build the manager binary
FROM registry.access.redhat.com/ubi9/go-toolset@sha256:0fb743b47cb9c6181b3a4cda9d845a3a218d843542e82f120925be8d4043fe96 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Copy the Go Modules manifests
COPY go.mod go.sum /build
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# copy source
COPY . .

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -trimpath -a -o /tmp/server ./cmd/etcd-shield/main.go

FROM registry.access.redhat.com/ubi9/ubi-micro@sha256:2d6db1e3434b10f338a5545de6b09c565a75e4d3f744e1b0604ba57ab2c53c2d
WORKDIR /
COPY --from=builder /tmp/server .
USER 65532:65532

ENTRYPOINT ["/server"]

