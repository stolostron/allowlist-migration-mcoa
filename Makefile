# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
# Licensed under the Apache License 2.0

BIN_DIR := $(CURDIR)/bin

.PHONY: deps
deps: go.mod go.sum
	go mod tidy
	go mod download
	go mod verify

.PHONY: build
build: deps
	go build -o $(BIN_DIR)/allowlist-migration main.go

.PHONY: build-all
build-all: deps
	@for OS in linux darwin windows; do for ARCH in amd64 arm64 ppc64le s390x; do \
			if [[ $${OS} != "linux" ]] && [[ $${ARCH} != *"64" ]]; then continue; fi; \
			echo "# Building $${OS}-$${ARCH}-allowlist-migration"; \
			GOOS=$${OS} GOARCH=$${ARCH} CGO_ENABLED=0 \
				go build -mod=readonly -ldflags="$(GO_LDFLAGS)" -o build_output/$${OS}-$${ARCH}-allowlist-migration main.go \
				|| exit 1; \
		done; done
	# Adding .exe extension to Windows binaries
	@for FILE in $$(ls -1 build_output/windows-* | grep -v ".exe$$"); do \
		mv $${FILE} $${FILE}.exe \
		|| exit 1; \
	done

.PHONY: test
test: deps
	go test ./...

.PHONY: clean
clean:
	-rm build_output/*
	-rm $(BIN_DIR)/*

