default: build

%.o: %.c
	gcc -I includes -o $@ -c $^

libbridge.a: bridge.o
	ar cru $@ $^

.PHONY: build
build: libbridge.a
	go build -a

.PHONY: test
test: build
	go test
