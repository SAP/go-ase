BINS ?= $(patsubst cmd/%,%,$(wildcard cmd/*))

build: $(BINS)
$(BINS):
	go build -o $@ ./cmd/$@/

test:
	go test -vet all -cover ./cgo/... ./go/... ./cmd/... ./libase/...

GO_EXAMPLES := $(wildcard examples/go/*)
CGO_EXAMPLES := $(wildcard examples/cgo/*)
EXAMPLES := $(GO_EXAMPLES) $(CGO_EXAMPLES)

examples: $(EXAMPLES)

.PHONY: $(EXAMPLES)
$(EXAMPLES):
	@echo Running example: $@
	go run ./$@/main.go
