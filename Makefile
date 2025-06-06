# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.1.1

# Get go version number. e.g. 1.19
GO_VERSION := $(shell echo `go version | sed 's|.*\(1\.[0-9][0-9]\).*$$|\1|'`)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')

# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# risingwavelabs.com/risingwave-operator-bundle:$VERSION and risingwavelabs.com/risingwave-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= risingwavelabs.com/risingwave-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

REGISTRY ?= ghcr.io/risingwavelabs
TAG ?= latest
# Image URL to use all building/pushing image targets
IMG ?= $(REGISTRY)/risingwave-operator:$(TAG)

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= crd:crdVersions=v1 
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.21

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

generate-all: generate manifests generate-docs generate-test-yaml generate-manager fmt

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	@$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./apis/..." output:crd:artifacts:config=config/crd/bases
	@$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./pkg/..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	@$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./apis/..."

generate-manager: ctrlkit-gen goimports-reviser ## Generate codes of controller managers.
	@$(CTRLKIT-GEN) -o pkg/manager/ -p "github.com/risingwavelabs/ctrlkit" -b hack/boilerplate.go.txt pkg/manager/risingwave_controller_manager.cm
	@mv pkg/manager/risingwave_controller_manager.go pkg/manager/risingwave_controller_manager_generated.go
	@$(GOIMPORTS-REVISER) -apply-to-generated-files -format -rm-unused -set-alias -company-prefixes "github.com/risingwavelabs/risingwave-operator" pkg/manager/risingwave_controller_manager_generated.go

	@$(CTRLKIT-GEN) -o pkg/manager/ -p "github.com/risingwavelabs/ctrlkit" -b hack/boilerplate.go.txt pkg/manager/risingwave_scale_view_controller_manager.cm
	@mv pkg/manager/risingwave_scale_view_controller_manager.go pkg/manager/risingwave_scale_view_controller_manager_generated.go
	@$(GOIMPORTS-REVISER) -apply-to-generated-files -format -rm-unused -set-alias -company-prefixes "github.com/risingwavelabs/risingwave-operator" pkg/manager/risingwave_scale_view_controller_manager_generated.go

go-work: ## create a new go.work file for this project. Will fix error 'gopls was not able to find modules in your workspace'
	rm -f go.work
	go work init
	go work use -r .

fmt: ## Run go fmt against code.
	@go fmt ./...

vendor: ## Ensure go vendor files
	@go mod vendor

vet: ## Run go vet against code.
	@go vet ./...

lint: golangci-lint
	$(GOLANGCI-LINT) run --config .golangci.yaml --fix

test: manifests generate fmt vet lint envtest ## Run tests.
	@KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out.tmp
	@cat cover.out.tmp | grep -v "_generated.go" > cover.out && rm -f cover.out.tmp

spellcheck:
	@if command -v cspell > /dev/null 2>&1 ; then \
	    cspell lint -c hack/cspell/cspell.json --relative --no-progress --no-summary --show-suggestions -e 'vendor/*' --gitignore **/*.go **/*.md; \
	else \
		echo "ERROR: cspell not found, install it manually! Link: https://cspell.org/docs/getting-started"; \
		exit 1; \
	fi

shellcheck:
	@bash -c 'shopt -s globstar; shellcheck -x -e SC1091 -s bash test/**/*.sh'

buildx:
	docker buildx install

##@ Build

proto: 
	cd pkg/controller/proto ; \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--experimental_allow_proto3_optional \
		meta.proto common.proto

build: build-manager

build-manager: generate fmt vet lint vendor ## Build manager binary.
	go build -ldflags "-X main.operatorVersion=$(shell git describe --tags)" -o bin/$(OS)/manager cmd/manager/manager.go

# Helper target for generating new local certs used in development. Use install-local instead
# if you also use Docker for Desktop as your development environment.
build-local-certs:
	mkdir -p ${TMPDIR}/k8s-webhook-server/serving-certs
	openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
		-keyout ${TMPDIR}/k8s-webhook-server/serving-certs/tls.key \
		-out ${TMPDIR}/k8s-webhook-server/serving-certs/tls.crt -subj "/CN=localhost" \
		-extensions san -config <(echo '[req]'; echo 'distinguished_name=req'; echo '[san]'; echo 'subjectAltName=DNS:host.docker.internal')

K8S_LOCAL_CONTEXT ?= docker-desktop

use-local-context:
	kubectl config use $(K8S_LOCAL_CONTEXT)

install-local: use-local-context kustomize manifests
	$(KUSTOMIZE) build config/local | kubectl apply --server-side --force-conflicts -f - >/dev/null

uninstall-local: use-local-context kustomize manifests
	$(KUSTOMIZE) build config/local | kubectl delete -f - >/dev/null

copy-local-certs:
	mkdir -p ${TMPDIR}/k8s-webhook-server/serving-certs
	cp -R config/local/certs/* ${TMPDIR}/k8s-webhook-server/serving-certs

run-local: use-local-context manifests generate fmt vet lint install-local
	go run -ldflags "-X main.operatorVersion=$(shell git describe --tags)" cmd/manager/manager.go -zap-time-encoding rfc3339

build-e2e-image:
	docker buildx build -f build/Dockerfile --build-arg USE_VENDOR=true --build-arg VERSION="$(shell git describe --tags)" -t docker.io/risingwavelabs/risingwave-operator:dev . --load

e2e-test: generate-test-yaml vendor build-e2e-image
	E2E_KUBERNETES_RUNTIME=kind ./test/e2e/e2e.sh

manifest-test:
	./test/e2e/e2e.sh -m

docker-cross-build: test buildx## Build docker image with the manager.
	docker buildx build -f build/Dockerfile --build-arg USE_VENDOR=false --build-arg VERSION="$(shell git describe --tags)" --platform=linux/amd64,linux/arm64 -t ${IMG} . --push

docker-cross-build-vendor: test buildx vendor
	docker buildx build -f build/Dockerfile --build-arg USE_VENDOR=true --build-arg VERSION="$(shell git describe --tags)" --platform=linux/amd64,linux/arm64 -t ${IMG} . --push

docker-build: test ## Build docker image with the manager.
	docker buildx build -f build/Dockerfile --build-arg USE_VENDOR=false --build-arg VERSION="$(shell git describe --tags)" -t ${IMG} . --load

docker-build-vendor: vendor test
	docker buildx build -f build/Dockerfile --build-arg USE_VENDOR=true --build-arg VERSION="$(shell git describe --tags)" -t ${IMG} . --load

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply --server-side -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

generate-yaml: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	@cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	@$(KUSTOMIZE) build config/default --output config/risingwave-operator.yaml

generate-test-yaml: generate-yaml
	@$(KUSTOMIZE) build config --output config/risingwave-operator-test.yaml

deploy: generate-yaml
	kubectl apply --server-side -f config/risingwave-operator.yaml

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/risingwave-operator.yaml

CONTROLLER_GEN = $(shell pwd)/bin/$(OS)/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.1)

KUSTOMIZE = $(shell pwd)/bin/$(OS)/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.5)

ENVTEST = $(shell pwd)/bin/$(OS)/setup-envtest
envtest: ## Download envtest-setup locally if necessary.
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

GOLANGCI-LINT = $(shell pwd)/bin/$(OS)/golangci-lint
golangci-lint: ## Download envtest-setup locally if necessary.
# $(call get-golangci-lint)
	$(call go-get-tool,$(GOLANGCI-LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2)

CTRLKIT-GEN = $(shell pwd)/bin/$(OS)/ctrlkit-gen
ctrlkit-gen: ## Download ctrlkit locally if necessary.
	$(call go-get-tool,$(CTRLKIT-GEN),github.com/arkbriar/ctrlkit/ctrlkit/cmd/ctrlkit-gen@latest)

GOIMPORTS-REVISER = $(shell pwd)/bin/$(OS)/goimports-reviser
goimports-reviser: ## Download goimports-reviser locally if necessary.
	$(call go-get-tool,$(GOIMPORTS-REVISER),github.com/incu6us/goimports-reviser/v3@v3.6.0)

GEN_CRD_API_REFERENCE_DOCS = $(shell pwd)/bin/$(OS)/gen-crd-api-reference-docs
gen-crd-api-reference-docs: ## Download gen-crd-api-reference-docs locally if necessary
	$(call go-get-tool,$(GEN_CRD_API_REFERENCE_DOCS),github.com/ahmetb/gen-crd-api-reference-docs@6cf1ede4da6128d8d489215698525e8289e707c4)

TYPES_V1ALPHA1_TARGET := $(shell find apis/risingwave/v1alpha1/ -name "*_types.go")

docs/general/api.md: gen-crd-api-reference-docs $(TYPES_V1ALPHA1_TARGET)
	@$(GEN_CRD_API_REFERENCE_DOCS) -api-dir "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1" -config "$(PWD)/hack/docs/config.json" -template-dir "$(PWD)/hack/docs/templates" -out-file "$(PWD)/docs/general/api.md"

.PHONY: generate-docs
generate-docs: docs/general/api.md

# get-golangci-lint will download golangci-lint binary into ./bin
define get-golangci-lint
@[ -f $(GOLANGCI-LINT) ] || { \
set -e ;\
echo "Downloading golangci-lint" ;\
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PROJECT_DIR)/bin v1.46.2; \
}
endef

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin/$(OS) go install $(2); \
}
endef

.PHONY: bundle
bundle: manifests kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundlehttps://github.com/risingwavelabs/risingwave-operator.git

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.15.1/$${OS}-$${ARCH}-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:v$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

CI_RUNNER_IMAGE := public.ecr.aws/q7q2v8s0/ci/rw-operator-ci-runner:$(shell date +'%Y%m%d')

build-ci-runner-image:
	docker buildx build --platform linux/amd64,linux/arm64 -f ci/runner/Dockerfile -t $(CI_RUNNER_IMAGE) . --push

SHFMT = $(shell pwd)/bin/$(OS)/shfmt
ensure_shfmt: ## Download shfmt locally if necessary.
# $(call get-golangci-lint)
	$(call go-get-tool,$(SHFMT),mvdan.cc/sh/v3/cmd/shfmt@latest)

shfmt: shellcheck ensure_shfmt
	shfmt -s -w .
