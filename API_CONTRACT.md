# API Contract Documentation
# Coffee Shop Point of Sales (POS) System

| Dokumen | Spesifikasi Kontrak API RESTful |
|---|---|
| **Sistem** | Coffee Shop POS System |
| **Arsitektur** | RESTful API |
| **Format Data** | JSON (`application/json`) / Multipart Form |
| **Base URL** | `https://api.poscoffee.local/api/v1` |
| **Versi Dokumen** | 1.2.0 |

---

## 1. Base URL dan Versioning
Seluruh endpoint API berada di bawah prefix versioning `/api/v1`. Prefix khusus `/owner/` dan `/cashier/` digunakan untuk memisahkan *middleware* otorisasi berbasis *Role* (RBAC) pada *framework router* Go.

---

## 2. Standard Response Format
Format respon dijamin konsisten untuk semua endpoint, baik pada saat sukses maupun gagal.

### 2.1 Respon Sukses (GET, PUT, POST, DELETE)
```json
{
  "success": true,
  "message": "Operasi berhasil",
  "data": {
    "id": "uuid-here",
    "field": "value"
  }
}
```

### 2.2 Respon Sukses dengan Daftar (List/Pagination)
```json
{
  "success": true,
  "message": "Berhasil mengambil data",
  "data": [
    {
      "id": "uuid-here"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

### 2.3 Respon Error Validasi (422)
```json
{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "name": "Nama produk wajib diisi",
    "price": "Harga harus lebih dari 0"
  }
}
```

### 2.4 Respon Error Umum (400, 401, 403, 404, 500)
```json
{
  "success": false,
  "message": "Deskripsi error (misal: Data tidak ditemukan)"
}
```

---

## 3. HTTP Status Codes
Daftar kode status HTTP baku yang digunakan sistem ini:
- `200 OK`: Sukses GET dan PUT (serta DELETE atau POST yang tidak membuat entitas).
- `201 Created`: Sukses POST (untuk pembuatan data baru).
- `400 Bad Request`: Request tidak valid / pelanggaran *business rules*.
- `401 Unauthorized`: Tidak ada token atau token expired.
- `403 Forbidden`: Token valid tapi tidak punya akses (salah *Role*).
- `404 Not Found`: Data tidak ditemukan.
- `422 Unprocessable Entity`: Validasi input (JSON/Form) gagal.
- `429 Too Many Requests`: Rate limit API tercapai.
- `500 Internal Server Error`: Server error tak terduga.

---

## 4. Daftar Lengkap Endpoint API

### 4.1 Auth (Public)
Endpoint *Public* tanpa proteksi *middleware* RBAC.

- **`POST /api/v1/auth/login`**
  - **Fungsi**: Autentikasi email & password.
- **`POST /api/v1/auth/refresh`**
  - **Fungsi**: Memperbarui *access token* yang akan expired.
- **`POST /api/v1/auth/logout`**
  - **Fungsi**: Menghapus sesi / invalidate token.

---

### 4.2 Categories (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/categories`**
  - **Fungsi**: Mengambil daftar semua kategori.
- **`POST /api/v1/owner/categories`**
  - **Fungsi**: Membuat kategori baru (Status 201).
- **`PUT /api/v1/owner/categories/:id`**
  - **Fungsi**: Memperbarui nama/deskripsi kategori.
- **`DELETE /api/v1/owner/categories/:id`**
  - **Fungsi**: Hapus kategori (*soft delete*).

---

### 4.3 Products (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/products`**
  - **Fungsi**: Mengambil katalog produk (dengan *pagination*).
- **`GET /api/v1/owner/products/:id`**
  - **Fungsi**: Mendapatkan detail spesifik produk.
- **`POST /api/v1/owner/products`**
  - **Fungsi**: Tambah produk baru (Status 201, mendukung multipart image upload).
- **`PUT /api/v1/owner/products/:id`**
  - **Fungsi**: Ubah data produk / nonaktifkan.
- **`DELETE /api/v1/owner/products/:id`**
  - **Fungsi**: Hapus produk (*soft delete*).

---

### 4.4 Stock (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/products/:id/stock`**
  - **Fungsi**: Lihat sisa stok saat ini untuk 1 produk.
- **`POST /api/v1/owner/products/:id/stock/adjustment`**
  - **Fungsi**: Update stok manual, wajib menyertakan field `notes` (Status 201).
- **`GET /api/v1/owner/products/:id/stock/movements`**
  - **Fungsi**: Lihat buku besar riwayat pergerakan stok 1 produk (dengan *pagination*).

---

### 4.5 Tables (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/tables`**
  - **Fungsi**: Daftar denah meja dan status okupansi.
- **`POST /api/v1/owner/tables`**
  - **Fungsi**: Tambah meja baru.
- **`PUT /api/v1/owner/tables/:id`**
  - **Fungsi**: Edit nomor atau kapasitas meja.
- **`DELETE /api/v1/owner/tables/:id`**
  - **Fungsi**: Hapus meja (selama tidak *OCCUPIED*).

---

### 4.6 Cashier Management (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/cashiers`**
  - **Fungsi**: Daftar semua akun kasir.
- **`POST /api/v1/owner/cashiers`**
  - **Fungsi**: Registrasi akun staf kasir baru (Status 201).
- **`PUT /api/v1/owner/cashiers/:id`**
  - **Fungsi**: Ubah nama/profil/reset password kasir.
- **`PATCH /api/v1/owner/cashiers/:id/toggle-status`**
  - **Fungsi**: Saklar aktif/nonaktif akun kasir (blokir akses login).

---

### 4.7 Promos (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/promos`**
  - **Fungsi**: Daftar semua skema promosi aktif/tidak aktif.
- **`POST /api/v1/owner/promos`**
  - **Fungsi**: Buat promo diskon (persentase/nominal) baru.
- **`PUT /api/v1/owner/promos/:id`**
  - **Fungsi**: Ubah data promo.
- **`DELETE /api/v1/owner/promos/:id`**
  - **Fungsi**: Hapus promo (*soft delete*).

---

### 4.8 Reports (Owner)
Dilindungi middleware role Owner (`/owner/`).

- **`GET /api/v1/owner/reports/revenue`**
  - **Fungsi**: Total omset harian/mingguan/bulanan.
- **`GET /api/v1/owner/reports/top-products`**
  - **Fungsi**: Daftar produk paling laris berdasarkan kuantitas.
- **`GET /api/v1/owner/reports/cashier-summary`**
  - **Fungsi**: Kinerja jumlah transaksi per staf kasir.
- **`GET /api/v1/owner/reports/export`**
  - **Fungsi**: Men-generate & mengunduh CSV berdasarkan rentang tanggal.

---

### 4.9 Shifts (Cashier)
Dilindungi middleware role Cashier (`/cashier/`).

- **`POST /api/v1/cashier/shifts/open`**
  - **Fungsi**: Deklarasi laci kas, buka shift baru agar bisa mulai transaksi (Status 201).
- **`POST /api/v1/cashier/shifts/close`**
  - **Fungsi**: Input kas fisik akhir dan hitung selisih rekonsiliasi.
- **`GET /api/v1/cashier/shifts/current`**
  - **Fungsi**: Informasi shift kasir yang sedang aktif.

---

### 4.10 Orders (Cashier)
Dilindungi middleware role Cashier (`/cashier/`).

- **`POST /api/v1/cashier/orders`**
  - **Fungsi**: Inisiasi pembuatan draft order (pilih meja/tipe pesanan).
- **`GET /api/v1/cashier/orders/:id`**
  - **Fungsi**: Melihat rincian tagihan pesanan spesifik.
- **`POST /api/v1/cashier/orders/:id/items`**
  - **Fungsi**: Tambah item produk ke dalam order aktif.
- **`PUT /api/v1/cashier/orders/:id/items/:item_id`**
  - **Fungsi**: Ubah qty produk/catatan spesifik dalam order.
- **`DELETE /api/v1/cashier/orders/:id/items/:item_id`**
  - **Fungsi**: Hapus satu baris produk dari order.
- **`POST /api/v1/cashier/orders/:id/promo`**
  - **Fungsi**: Terapkan kode promo pada keranjang pesanan.
- **`DELETE /api/v1/cashier/orders/:id/promo`**
  - **Fungsi**: Copot/buang promo dari pesanan.
- **`GET /api/v1/cashier/orders/history`**
  - **Fungsi**: Daftar riwayat transaksi pada shift kasir saat ini (dengan *pagination*).

---

### 4.11 Payments
Gabungan endpoint *Cashier* dan *Public Gateway* Webhook.

- **`POST /api/v1/cashier/orders/:id/checkout`**
  - **Fungsi**: *Lock* order, generate Midtrans Snap Token, atau eksekusi pelunasan tunai (Cashier Access).
- **`GET /api/v1/cashier/orders/:id/payment-status`**
  - **Fungsi**: Cek real-time status bayar dari Midtrans.
- **`POST /api/v1/webhooks/midtrans`**
  - **Fungsi**: *Endpoint Public* khusus menerima *HTTP Notification* mutasi status dari server Midtrans (Verifikasi *Signature Key* lalu Update stok/status lunas).