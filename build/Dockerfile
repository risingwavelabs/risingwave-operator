ARG USE_VENDOR=false

# Build the manager binary
FROM golang:1.20 as base

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
## cache deps before building and copying source so that we don't need to re-download as much
## and so that source changes don't invalidate our downloaded layer
#RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY apis/ apis/
COPY pkg/ pkg/
COPY vendor* vendor/

FROM base as vendor-true
ARG TARGETARCH
ARG VERSION=""
RUN echo "version is ${VERSION}"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags "-X main.version=${VERSION}" -mod=vendor -a -o manager cmd/manager/manager.go

FROM base as vendor-false
ARG TARGETARCH
ARG VERSION=""
RUN echo "version is ${VERSION}"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -ldflags "-X main.version=${VERSION}" -a -o manager cmd/manager/manager.go

FROM vendor-${USE_VENDOR} as vendor
RUN echo "vendor is ${USE_VENDOR}"

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static-debian11
WORKDIR /
COPY --from=vendor /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
