VERSION=$(shell git rev-parse HEAD)
GO_LDFLAGS=-ldflags "-s -w -X `go list ./version`.Version=$(VERSION)"

.PHONY: all
all: build

.PHONY: build
build: generate
	CGO_ENABLED=0 go build -v ${GO_LDFLAGS}

.PHONY: generate
generate:
	go generate ./templates

# To install gometalinter on macOS:
# brew tap alecthomas/homebrew-tap
# brew install gometalinter
.PHONY: setup
setup:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

.PHONY: lint
lint:
	gometalinter \
		--vendor --tests --disable-all \
		--exclude templates/src/ \
		-E gofmt -E vet -E goimports -E golint ./...

.PHONY: test
test:
	go test -v ./...
