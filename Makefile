BINS ?= $(patsubst cmd/%,%,$(wildcard cmd/*))

build: $(BINS)
$(BINS):
	go build -o $@ ./cmd/$@/

test:
	go test -vet all -cover ./cgo/... ./go/... ./cmd/... ./libase/...
