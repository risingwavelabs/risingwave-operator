# Build the manager binary
FROM golang:1.16 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
## cache deps before building and copying source so that we don't need to re-download as much
## and so that source changes don't invalidate our downloaded layer
#RUN go mod download
ARG TARGETARCH
ARG USE_VENDOR

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY pkg/ pkg/
COPY vendor* vendor/

# Build
RUN if [ "$USE_VENDOR" = "true" ] ; then \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -mod=vendor -a -o manager main.go ;  \
    else  \
    CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -a -o manager main.go ;  \
    fi

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM ubuntu:20.04
WORKDIR /
RUN apt-get update && apt-get -y install ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /workspace/manager .
COPY template/ template/
USER 65532:65532

ENTRYPOINT ["/manager"]
