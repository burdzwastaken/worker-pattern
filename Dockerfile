# hadolint ignore=DL3026
FROM golang:1.16.5@sha256:be0e3a0f3ffa448b0bcbb9019edca692b8278407a44dc138c60e6f12f0218f87 AS go
WORKDIR /build/worker-pattern

# Copy the go.mod over so docker can cache the module downloads if possible.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM go AS test
RUN make test

# Build stage
FROM go AS builder
ARG VERSION
ENV VERSION $VERSION
RUN make install

FROM go AS executor

COPY --from=builder /go/bin/worker-pattern /usr/local/bin/worker-pattern
ENTRYPOINT ["worker-pattern"]
