# Gunakan image resmi Golang sebagai base image untuk build aplikasi
FROM golang:1.24.2-alpine AS builder


# Set working directory di container
WORKDIR /app

# Copy semua file source code ke working directory
COPY . .

# Download dependencies dan build aplikasi jadi binary
RUN go mod download
RUN go build -o app .

# Tahap kedua: buat image minimal untuk running aplikasi (multi-stage build)
FROM alpine:latest

# Install library yang dibutuhkan (misal: untuk TLS, libc, dll)
RUN apk --no-cache add ca-certificates

# Copy binary hasil build dari tahap builder
COPY --from=builder /app/app /app/app

# Port yang akan diekspos
EXPOSE 8080

# Command untuk menjalankan aplikasi ketika container dijalankan
CMD ["/app/app"]
