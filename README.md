# Coffee Shop POS API

Backend API sistem Point of Sales (POS) untuk coffee shop skala menengah yang dibangun menggunakan **Go (Golang)**, basis data **MySQL**, dan terintegrasi dengan gateway pembayaran digital **Midtrans Snap**.

---

## ☕ Tentang Project

Project ini dirancang untuk menangani operasional harian kedai kopi secara cepat, akurat, dan transparan, meliputi:
- **Manajemen Operasional Kasir & Shift**: Buka/tutup shift kasir dengan rekonsiliasi kas fisik.
- **Transaksi & Pemesanan**: Layanan *Dine-in* (dengan pemilihan meja) dan *Takeaway*.
- **Pembayaran Digital Modern**: Integrasi langsung dengan Midtrans Snap (QRIS, E-Wallet, VA) dan pencatatan uang tunai (*Cash*).
- **Manajemen Inventaris & Stok**: Audit trail pencatatan mutasi stok (*Stock In, Out Sale, Adjustment*).
- **Laporan & Analitik Bisnis**: Dashboard revenue, menu terlaris, kinerja per kasir, serta ekspor laporan ke CSV.

---

## 🛠️ Tech Stack

- **Bahasa Pemrograman**: Go (Golang) v1.23+
- **Database**: MySQL 8.0+
- **Cache & Session**: Redis
- **Payment Gateway**: Midtrans Snap API
- **Autentikasi**: JWT (JSON Web Token) dengan otorisasi RBAC (Owner & Cashier)

---

## 📂 Struktur Folder Project

Mengikuti arsitektur standar **Clean Architecture / Standard Go Project Layout**:

```text
.
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── config/
│   └── config.go                # Konfigurasi aplikasi & mapping environment
├── internal/                    # Private application & business logic code
│   ├── dto/                     # Data Transfer Objects (Request / Response struct)
│   ├── entity/                  # Domain model / Database entity struct
│   ├── handler/                 # HTTP controller / endpoint handlers
│   ├── middleware/              # HTTP middleware (Auth JWT, CORS, Logger, Recover)
│   ├── repository/              # Interface repository abstraksi data access
│   │   └── mysql/               # Implementasi data access layer MySQL
│   └── service/                 # Business logic layer
├── migrations/                  # Script file migrasi database SQL
├── pkg/                         # Reusable library / shared packages
│   ├── database/                # Koneksi database MySQL & pool settings
│   ├── jwt/                     # Helper pembuatan & validasi JWT
│   ├── redis/                   # Koneksi & utilitas Redis client
│   ├── response/                # Standar serializer response JSON & error helper
│   └── validator/               # Custom input request validation helper
├── .env.example                 # Template environment variables
├── .gitignore                   # Aturan pengabaian file Git
├── API_CONTRACT.md              # Spesifikasi RESTful API endpoint lengkap
├── PRD.md                       # Product Requirements Document
├── SCHEMA.md                    # Desain skema basis data MySQL & ERD
├── postman_collection.json      # Postman Collection siap import
└── README.md                    # Dokumentasi utama project
```

---

## 🚀 Memulai (Getting Started)

### Prasyarat
- Go 1.23 atau lebih baru
- MySQL 8.0+
- Redis (opsional untuk caching)

### Langkah Menjalankan
1. **Clone repositori dan salin environment**:
   ```bash
   cp .env.example .env
   ```
2. **Sesuaikan konfigurasi** pada file `.env` (kredensial MySQL, secret JWT, Midtrans).
3. **Jalankan aplikasi**:
   ```bash
   go run cmd/api/main.go
   ```

---

## 📖 Dokumentasi Terkait

- [PRD.md](file:///c:/Users/Zeva/Documents/PROJECTTTT/FOKUS/POS-COFFE/PRD.md) — Product Requirements Document lengkap
- [SCHEMA.md](file:///c:/Users/Zeva/Documents/PROJECTTTT/FOKUS/POS-COFFE/SCHEMA.md) — Desain skema database MySQL, tipe data & relasi
- [API_CONTRACT.md](file:///c:/Users/Zeva/Documents/PROJECTTTT/FOKUS/POS-COFFE/API_CONTRACT.md) — Kontrak endpoint API lengkap
- [postman_collection.json](file:///c:/Users/Zeva/Documents/PROJECTTTT/FOKUS/POS-COFFE/postman_collection.json) — Postman collection untuk pengujian API
