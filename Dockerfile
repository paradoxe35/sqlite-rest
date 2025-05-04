FROM golang:1.21-bullseye AS builder
WORKDIR /build

# Import the codebase
COPY . .

# Create binary with optimizations
RUN VERSION=$(git describe --tags || echo "dev") && \
    go build -mod vendor -trimpath -a -tags netgo -ldflags "-s -w -X main.VERSION=${VERSION} -extldflags \"-static\"" \
    -o ./bin/sqlite-rest ./cmd/sqlite-rest.go


FROM scratch AS runner
WORKDIR /app

# Server binary from builder
COPY --from=builder /build/bin/sqlite-rest ./bin/sqlite-rest

# Self-signed certificate from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Run the server
ENTRYPOINT ["/app/bin/sqlite-rest"]