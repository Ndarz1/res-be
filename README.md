# Backend Wisata API

Ini adalah backend service untuk aplikasi pariwisata yang dibangun menggunakan Go (Golang) dan PostgreSQL.

## 📋 Fitur

- **Autentikasi & Otorisasi**: Login, Register, Logout, Profil User (`/api/login`, `/api/register`, dll).
- **Manajemen Wisata**: CRUD untuk data destinasi wisata (`/api/wisata`).
- **Kategori**: Manajemen kategori wisata (`/api/categories`).
- **Booking**: Sistem pemesanan, riwayat, pembayaran, dan pembatalan (`/api/booking`).
- **Dashboard**: Statistik untuk admin, booking terbaru, dan wisata populer (`/api/dashboard`).
- **User Management**: Pengelolaan pengguna (`/api/users`).
- **Review**: Sistem ulasan untuk destinasi wisata (`/api/reviews`).
- **Blog**: Artikel dan postingan blog (`/api/blog`).
- **Upload File**: Penanganan upload file statis.
- **CORS**: Sudah dikonfigurasi untuk mengizinkan akses dari frontend.

## 🛠️ Teknologi yang Digunakan

- **Bahasa**: Go 1.25.3
- **Database**: PostgreSQL (Driver: `pgx/v5`)
- **Session**: Gorilla Sessions (`github.com/gorilla/sessions`)
- **Routing**: Standard `net/http` ServeMux

## 🚀 Cara Menjalankan

### Prasyarat

Pastikan Anda telah menginstal:

- [Go](https://go.dev/dl/) (versi 1.25 atau lebih baru)
- [PostgreSQL](https://www.postgresql.org/download/)

### Langkah Instalasi

1. **Clone repository ini:**

   ```bash
   git clone <url-repository-anda>
   cd res-be
   ```

2. **Download dependencies:**

   ```bash
   go mod tidy
   ```

3. **Konfigurasi Database:**

   - Secara default, aplikasi akan mencoba terhubung ke database PostgreSQL dengan konfigurasi berikut (terdapat di `config/database.go`):
     - **User**: `postgres`
     - **Host**: `localhost:5432`
   - Jika konfigurasi database Anda berbeda, silakan edit file `config/database.go`.

4. **Jalankan Aplikasi:**
   ```bash
   go run main.go
   ```
   Server akan berjalan di `http://localhost:8080`.

## 📂 Struktur API

Berikut adalah beberapa endpoint utama yang tersedia:

### Auth

- `POST /api/login` - Masuk ke aplikasi
- `POST /api/register` - Pendaftaran pengguna baru
- `GET /api/me` - Cek user yang sedang login

### Wisata

- `GET /api/wisata` - Ambil semua data wisata
- `GET /api/wisata/detail?id=...` - Detail wisata
- `POST /api/wisata/create` - Tambah wisata baru
- `PUT /api/wisata/update` - Update data wisata
- `DELETE /api/wisata/delete` - Hapus wisata

### Booking

- `POST /api/booking/create` - Buat pesanan baru
- `GET /api/booking/history` - Lihat riwayat pesanan
- `POST /api/booking/pay` - Proses pembayaran

_(Silakan cek `main.go` untuk daftar endpoint lengkap)_

## ⚠️ Catatan Penting

- **Session**: Konfigurasi session key terdapat di `config/session.go`. Jangan lupa untuk menggantinya jika akan di-deploy ke production.
- **Uploads**: File yang diupload akan tersimpan di direktori `uploads/` dan dapat diakses melalui endpoint `/uploads/`.
