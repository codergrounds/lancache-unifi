.PHONY: build test clean run docker

BINARY_NAME=lancache-unifi
DOCKER_IMAGE=lancache-unifi:latest

build:
	go build -ldflags="-s -w" -o dist/$(BINARY_NAME) .

test:
	go test -v ./...

clean:
	rm -rf dist/
	rm -f $(BINARY_NAME)

run:
	go run .

docker:
	docker build -t $(DOCKER_IMAGE) .
