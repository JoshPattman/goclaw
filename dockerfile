# Stage 1: build the Go binary
FROM golang:1.25.6-bookworm AS builder
WORKDIR /workspace

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build in the current directory (./)
COPY . .
RUN CGO_ENABLED=0 go build -o /app/goclaw ./cmd

# Stage 2: minimal runtime
FROM gcr.io/distroless/static:nonroot
# Copy the built binary
COPY --from=builder /app/goclaw /usr/local/bin/goclaw
# Copy system CA certs so TLS works
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

USER nonroot
ENTRYPOINT ["/usr/local/bin/goclaw"]