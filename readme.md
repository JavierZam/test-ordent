## Asumsi-Asumsi

1. **Autentikasi:**

   - Menggunakan JWT (JSON Web Token) untuk autentikasi
   - Token berlaku selama 24 jam
   - Memiliki dua peran pengguna: customer dan admin

2. **Database:**

   - Menggunakan PostgreSQL sebagai database
   - Menggunakan library pq untuk koneksi database
   - Mengimplementasikan connection pooling untuk efisiensi

3. **Keamanan:**

   - Password dihash menggunakan bcrypt
   - Semua endpoint sensitif dilindungi dengan middleware autentikasi
   - Implementasi CORS untuk keamanan browser

4. **Fitur E-Commerce:**

   - Pengguna dapat melihat produk tanpa login
   - Untuk menambahkan item ke keranjang, pengguna harus login
   - Hanya admin yang dapat menambah, mengupdate, atau menghapus produk
   - Order dapat dibuat dari item yang ada di keranjang

5. **Lingkungan:**
   - Aplikasi dapat dikonfigurasi melalui file config.yaml
   - Aplikasi dapat di-deploy menggunakan Docker
   - Lingkungan pengembangan menggunakan Go versi 1.20

## Setup dan Instalasi

### Prasyarat

- Go 1.20 atau lebih baru
- PostgreSQL
- Docker (opsional)

### Menjalankan Aplikasi Secara Lokal

1. Clone repositori
2. Salin `.env.example` ke `.env` dan sesuaikan dengan lingkungan Anda
3. Buat database PostgreSQL
4. Jalankan migrasi database (file SQL tersedia di direktori `migrations`)
5. Jalankan aplikasi:
   ```bash
   go run cmd/server/main.go
   ```

### Menjalankan dengan Docker

1. Build image Docker:

   ```bash
   docker build -t test-ordent .
   ```

2. Jalankan container:
   ```bash
   docker run -p 8080:8080 --env-file .env test-ordent
   ```

## API Endpoints

### Autentikasi

- `POST /api/auth/login` - Login pengguna
- `POST /api/auth/register` - Registrasi pengguna baru

### Produk

- `GET /api/products` - Mendapatkan daftar produk (publik)
- `GET /api/products/{id}` - Mendapatkan detail produk (publik)
- `POST /api/products` - Menambahkan produk baru (admin)
- `PUT /api/products/{id}` - Mengupdate produk (admin)
- `DELETE /api/products/{id}` - Menghapus produk (admin)

### Keranjang

- `GET /api/cart` - Mendapatkan keranjang belanja (login)
- `POST /api/cart/items` - Menambahkan item ke keranjang (login)
- `DELETE /api/cart/items/{id}` - Menghapus item dari keranjang (login)

### Order

- `POST /api/orders` - Membuat order baru dari keranjang (login)
- `GET /api/orders` - Mendapatkan daftar order (login)

## Dokumentasi API

Dokumentasi API tersedia menggunakan Swagger di `/api/swagger/index.html`

````

## 9. Dockerfile

```dockerfile
FROM golang:1.20-alpine AS build

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /test-ordent ./cmd/server

# Create a smaller image for just running the application
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy binary from build stage
COPY --from=build /test-ordent .
COPY config/config.yaml ./config/

# Set user to non-root
RUN adduser -D -g '' appuser
USER appuser

# Expose port
EXPOSE 8080

# Run the application
CMD ["./test-ordent"]
````
