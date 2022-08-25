TARGET_DIR := target
TARGET_FILE := $(TARGET_DIR)/risingwave-overview.json

JSONNET_SOURCES := $(shell find . -name '*.json' -not -path './vendor/*' -not -path './target/*')
JSONNET_SOURCES += $(shell find . -name '*.jsonnet' -not -path './vendor/*' -not -path './target/*')
JSONNET_SOURCES += $(shell find . -name '*.libsonnet' -not -path './vendor/*' -not -path './target/*')

build: ensure-target-dir $(TARGET_FILE)

$(TARGET_FILE): $(JSONNET_SOURCES)
	jsonnet -J vendor -o $(TARGET_FILE) risingwave-overview.jsonnet

ensure-target-dir:
	mkdir -p $(TARGET_DIR)

.PHONY: clean
clean:
	rm -r $(TARGET_DIR)

.PHONY: vendor
vendor:
	jb update