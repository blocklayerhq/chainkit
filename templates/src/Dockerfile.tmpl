FROM golang:alpine AS build-env

# Set working directory for the build
WORKDIR /go/src/{{ .GoPkg }}

# Setup build environment
RUN apk add --no-cache curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Copy only the dependency manifests
COPY Gopkg.toml Gopkg.lock ./

# Fetch dependencies
RUN dep ensure -v --vendor-only

# Build dependencies
RUN find vendor -maxdepth 2 -mindepth 2 -type d -exec \
    sh -c 'CGO_ENABLED=0 go build -v -ldflags "-s -w" {{ .GoPkg }}/{}/... || true' \;

# Add source files
COPY . ./

# Build and install
RUN \
    CGO_ENABLED=0 go build -v -ldflags "-s -w" -o build/{{ .Name }}d ./cmd/{{ .Name}}d && \
    CGO_ENABLED=0 go build -v -ldflags "-s -w" -o build/{{ .Name }}cli ./cmd/{{ .Name}}cli

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/{{ .GoPkg }}/build/{{ .Name }}d /usr/bin/{{ .Name }}d
COPY --from=build-env /go/src/{{ .GoPkg }}/build/{{ .Name }}cli /usr/bin/{{ .Name }}cli

# Run the daemon by default
CMD ["{{ .Name }}d"]
