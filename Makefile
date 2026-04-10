.PHONY: build test clean run docker

BINARY_NAME=lancache-unifi
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "latest")
DOCKER_IMAGE=lancache-unifi:$(GIT_BRANCH)

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
