.PHONY: build
build:
	go build -C cmd/gophermart -o gophermart

.PHONY: run
run: build
	./cmd/gophermart/gophermart