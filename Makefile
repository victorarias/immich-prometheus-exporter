.PHONY: build test clean run docker

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o immich-prometheus-exporter ./cmd/immich-prometheus-exporter

test:
	go test -v -race ./...

clean:
	rm -f immich-prometheus-exporter

run: build
	./immich-prometheus-exporter

docker:
	docker build -t immich-prometheus-exporter:local .
