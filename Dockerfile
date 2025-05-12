# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

# Copy go mod/sum and download dependencies (enables docker cache for deps)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
# Specify GOOS and GOARCH for cross-compilation
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app ./cmd/main.go
RUN chmod +x /app/app

# Final stage
FROM scratch
# Copy only the compiled binary from the builder stage
# Ensure binary is copied to /app/app
COPY --from=builder /app/app /app/app
# Set entrypoint
# Be explicit with the binary path
ENTRYPOINT ["/app/app"]