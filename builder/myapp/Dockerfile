FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES curl git

# Set working directory for the build
WORKDIR /go/src/github.com/blocklayerhq/chainkit/builder/myapp

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    dep ensure -v  && \
    go install ./cmd/myappd ./cmd/myappcli

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/myappd /usr/bin/myappd
COPY --from=build-env /go/bin/myappcli /usr/bin/myappcli

# Run the daemon by default
CMD ["myappd"]
