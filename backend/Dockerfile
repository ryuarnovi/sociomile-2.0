# Build stage
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go app (ganti main.go jika nama file utama berbeda)
RUN go build -o app cmd/server/main.go
# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .

# Expose port (ganti jika port berbeda)
EXPOSE 8000

# Jalankan aplikasi
CMD ["./app"]