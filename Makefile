SHELL := /bin/bash

BINS := $(notdir $(wildcard cmd/*))
DIST_DIR ?= dist
TOOL ?=
VERSION ?=
BUILD_VERSION := $(if $(VERSION),$(VERSION),dev)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
RELEASE_OSES ?= linux darwin
RELEASE_ARCHES ?= amd64 arm64
TAG_NAME = $(TOOL)-$(VERSION)
ARCHIVE = $(DIST_DIR)/$(TOOL)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz
SEMVER_PATTERN = ^v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$$

.PHONY: help list test build build-tool package release-archives print-tag tag push-tag clean check-tool check-version

help:
	@printf 'Targets:\n'
	@printf '  make list\n'
	@printf '  make test\n'
	@printf '  make build\n'
	@printf '  make build-tool TOOL=slack-post VERSION=v0.1.0\n'
	@printf '  make package TOOL=slack-post VERSION=v0.1.0 GOOS=linux GOARCH=amd64\n'
	@printf '  make release-archives TOOL=slack-post VERSION=v0.1.0\n'
	@printf '  make print-tag TOOL=slack-post VERSION=v0.1.0\n'
	@printf '  make tag TOOL=slack-post VERSION=v0.1.0\n'
	@printf '  make push-tag TOOL=slack-post VERSION=v0.1.0\n'

list:
	@printf '%s\n' $(BINS)

test:
	go test ./...

build:
	@mkdir -p "$(DIST_DIR)"
	@for bin in $(BINS); do \
		echo "building $$bin ($(BUILD_VERSION))"; \
		go build -trimpath -ldflags "-s -w -X main.version=$(BUILD_VERSION)" -o "$(DIST_DIR)/$$bin" "./cmd/$$bin"; \
	done

build-tool: check-tool
	@mkdir -p "$(DIST_DIR)"
	go build -trimpath -ldflags "-s -w -X main.version=$(BUILD_VERSION)" -o "$(DIST_DIR)/$(TOOL)" "./cmd/$(TOOL)"

package: check-tool check-version
	@mkdir -p "$(DIST_DIR)"
	@tmp_dir="$$(mktemp -d)"; \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	echo "packaging $(TOOL) $(VERSION) for $(GOOS)/$(GOARCH)"; \
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" CGO_ENABLED=0 \
		go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o "$$tmp_dir/$(TOOL)" "./cmd/$(TOOL)"; \
	tar -C "$$tmp_dir" -czf "$(ARCHIVE)" "$(TOOL)"; \
	echo "$(ARCHIVE)"

release-archives: check-tool check-version
	@for os in $(RELEASE_OSES); do \
		for arch in $(RELEASE_ARCHES); do \
			$(MAKE) --no-print-directory package TOOL="$(TOOL)" VERSION="$(VERSION)" GOOS="$$os" GOARCH="$$arch"; \
		done; \
	done

print-tag: check-tool check-version
	@echo "$(TAG_NAME)"

tag: check-tool check-version
	@if git rev-parse -q --verify "refs/tags/$(TAG_NAME)" >/dev/null; then \
		echo "tag $(TAG_NAME) already exists" >&2; \
		exit 1; \
	fi
	git tag -a "$(TAG_NAME)" -m "Release $(TOOL) $(VERSION)"
	@echo "created tag $(TAG_NAME)"

push-tag: check-tool check-version
	git push origin "$(TAG_NAME)"

clean:
	rm -rf "$(DIST_DIR)"

check-tool:
	@if [ -z "$(TOOL)" ]; then \
		echo "TOOL is required, e.g. TOOL=slack-post" >&2; \
		exit 2; \
	fi
	@if [ ! -d "cmd/$(TOOL)" ]; then \
		echo "unknown TOOL=$(TOOL); available: $(BINS)" >&2; \
		exit 2; \
	fi

check-version:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required, e.g. VERSION=v0.1.0" >&2; \
		exit 2; \
	fi
	@if ! printf '%s\n' "$(VERSION)" | grep -Eq '$(SEMVER_PATTERN)'; then \
		echo "VERSION must be SemVer with a leading v, e.g. v0.1.0" >&2; \
		exit 2; \
	fi
