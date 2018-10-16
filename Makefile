VERSION=$(shell git rev-parse HEAD)
GO_LDFLAGS=-ldflags "-s -w -X `go list ./pkg/version`.Version=$(VERSION)"

.PHONY: all
all: build

.PHONY: build
build: generate
	CGO_ENABLED=0 go build -v ${GO_LDFLAGS}

.PHONY: generate
generate:
	go generate ./...

# To install gometalinter on macOS:
# brew tap alecthomas/homebrew-tap
# brew install gometalinter
.PHONY: lint
lint:
	gometalinter \
		--vendor --tests --disable-all \
		-E gofmt -E vet -E goimports -E golint ./...
