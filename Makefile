# SPDX-FileCopyrightText: 2020 - 2025 SAP SE
#
# SPDX-License-Identifier: Apache-2.0

GO ?= go

BINS ?= $(patsubst cmd/%,%,$(wildcard cmd/*))

REUSE_ARGS = --skip-unrecognised --year='2020-$(shell date +%Y)' --copyright='SAP SE' --license='Apache-2.0' --merge-copyrights

build: $(BINS)
$(BINS):
	go build -o $@ ./cmd/$@/

generate:
	grep -r '^// Code generated by ".*"\; DO NOT EDIT.$\' ./ | awk -F: '{ print $$1 }' | xargs --no-run-if-empty rm
	$(GO) generate ./...
	reuse annotate $(REUSE_ARGS) $(shell find . -type f -not -path '*/.git/*')

lint:
	golangci-lint run ./...

integration:
	$(GO) test -race -cover ./... --tags=integration

.PHONY: report
report:
	$(GO) clean -testcache
	$(GO) test -race -cover -coverprofile=/tmp/cov.out ./... --tags=integration
	$(GO) tool cover -html=/tmp/cov.out -o ./report.html
	rm /tmp/cov.out

EXAMPLES := $(dir $(wildcard examples/*/main.go))

examples: $(EXAMPLES)

.PHONY: $(EXAMPLES)
$(EXAMPLES):
	@echo Running example: $@
	$(GO) run ./$@/main.go
