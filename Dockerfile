FROM golang:1.18.0-bullseye AS builder
WORKDIR /build

# Import the codebase
COPY . .

# Create binary
RUN go build -mod vendor -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./bin/sqlite-rest ./cmd/sqlite-rest.go & wait


FROM scratch AS runner
WORKDIR /app

# Server binary from builder
COPY --from=builder /build/bin/sqlite-rest ./bin/sqlite-rest

# Self-signed certificate from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Run the server
ENTRYPOINT ["/app/bin/sqlite-rest"]