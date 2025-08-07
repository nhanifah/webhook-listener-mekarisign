# Gunakan image resmi Go
FROM golang:1.23

# Set working directory di dalam container
WORKDIR /app

# Copy go.mod ke dalam container
COPY go.mod go.sum ./

# Jalankan `go mod tidy` untuk memastikan semua dependensi tersedia
RUN go mod download

# Copy seluruh kode ke dalam container
COPY . .

# Build aplikasi Go
RUN go build -o app ./cmd

# Jalankan aplikasi
CMD ["./app"]
