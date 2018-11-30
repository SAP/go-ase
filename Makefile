build:
	make -C ./cgo build
	go build -o cgo-ase ./cmd/cgo-ase

test:
	go test -vet all ./cgo ./cmd/cgo-ase
