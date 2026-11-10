# Database Schema Documentation
# Coffee Shop Point of Sales (POS) System

| Dokumen | Spesifikasi Skema Database MySQL |
|---|---|
| **Sistem** | Coffee Shop POS System |
| **Database Engine** | MySQL 8.0+ (InnoDB Storage Engine) |
| **Collation** | `utf8mb4_unicode_ci` |
| **Versi Dokumen** | 1.2.0 |

---

## 1. Daftar Tabel dan Deskripsi Singkat

Sistem POS ini terdiri dari 10 (sepuluh) tabel utama yang saling berelasi:

| No | Nama Tabel | Deskripsi Singkat |
|---|---|---|
| 1 | **`users`** | Menyimpan data akun pengguna (Owner dan Cashier), kredensial login, dan status aktif. |
| 2 | **`categories`** | Menyimpan kategori produk (misal: *Coffee*, *Non-Coffee*, *Pastry*). |
| 3 | **`products`** | Menyimpan data katalog menu, harga jual, foto, dan status ketersediaan. |
| 4 | **`stock_movements`** | Buku besar riwayat pergerakan stok mutlak (purchase, sale, adjustment, return). |
| 5 | **`tables`** | Menyimpan data master meja makan, kapasitas, dan status okupansi. |
| 6 | **`promos`** | Menyimpan data program promosi, diskon, syarat belanja, masa berlaku, dan kuota. |
| 7 | **`shifts`** | Mencatat sesi kerja kasir, modal kas awal, rekonsiliasi akhir, dan total penjualan. |
| 8 | **`orders`** | Menyimpan *header* transaksi penjualan, kasir, shift, subtotal, diskon, dan status. |
| 9 | **`order_items`** | Menyimpan rincian produk yang dipesan dalam transaksi berserta *snapshot* harga. |
| 10 | **`payments`** | Mencatat riwayat pembayaran, integrasi Midtrans, dan payload webhook mentah. |

---

## 2. Struktur Kolom Setiap Tabel

### 2.1 Tabel `users`
Menyimpan akun pengguna dengan dukungan soft delete dan toggle aktif/nonaktif.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `name` | `VARCHAR(100)` | `NOT NULL` | Nama lengkap pengguna |
| `email` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | Alamat email unik untuk login |
| `password` | `VARCHAR(255)` | `NOT NULL` | Hash kata sandi terenkripsi |
| `role` | `ENUM('owner', 'cashier')` | `NOT NULL` | Hak akses otorisasi sistem |
| `is_active` | `BOOLEAN` | `DEFAULT TRUE` | Toggle izin login (tanpa hapus data riwayat) |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu akun dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu akun diperbarui |
| `deleted_at` | `TIMESTAMP` | `NULL` | Penanda *soft delete* |

### 2.2 Tabel `categories`
Menyimpan grup kategori menu produk.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `name` | `VARCHAR(100)` | `NOT NULL, UNIQUE` | Nama kategori (misal: "Espresso Based") |
| `description` | `TEXT` | `NULL` | Deskripsi singkat kategori |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu data dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu data diperbarui |
| `deleted_at` | `TIMESTAMP` | `NULL` | Penanda *soft delete* |

### 2.3 Tabel `products`
Menyimpan katalog item produk atau menu yang dijual.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `category_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke tabel `categories` |
| `name` | `VARCHAR(100)` | `NOT NULL` | Nama menu/produk |
| `description` | `TEXT` | `NULL` | Deskripsi detail menu |
| `price` | `BIGINT` | `NOT NULL` | Harga jual dalam satuan **sen** |
| `image_url` | `VARCHAR(255)` | `NULL` | Path URL atau nama file foto produk |
| `is_available` | `BOOLEAN` | `DEFAULT TRUE` | Toggle ketersediaan menu di kasir |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu produk ditambahkan |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu produk diperbarui |
| `deleted_at` | `TIMESTAMP` | `NULL` | Penanda *soft delete* |

### 2.4 Tabel `stock_movements`
Buku besar (*ledger*) pergerakan stok untuk menjaga riwayat persediaan secara absolut.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `product_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke tabel `products` |
| `type` | `ENUM('purchase', 'sale', 'adjustment', 'return')` | `NOT NULL` | Jenis mutasi pergerakan stok |
| `quantity` | `INT` | `NOT NULL` | Jumlah delta unit (positif = masuk, negatif = keluar) |
| `note` | `VARCHAR(255)` | `NULL` | Catatan alasan mutasi (*wajib diisi di level aplikasi untuk adjustment*) |
| `created_by` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `users` (pihak yang melakukan mutasi) |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu mutasi dicatat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu data diperbarui |

### 2.5 Tabel `tables`
Menyimpan data meja makan kafe.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `table_number` | `VARCHAR(20)` | `NOT NULL, UNIQUE` | Nomor/nama meja unik |
| `capacity` | `INT` | `NOT NULL, DEFAULT 2` | Kapasitas kursi |
| `status` | `ENUM('available', 'occupied', 'reserved')` | `DEFAULT 'available'` | Status okupansi meja |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu meja dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu status meja diperbarui |
| `deleted_at` | `TIMESTAMP` | `NULL` | Penanda *soft delete* |

### 2.6 Tabel `promos`
Menyimpan aturan promosi diskon transaksi.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `code` | `VARCHAR(50)` | `NOT NULL, UNIQUE` | Kode kupon promo unik |
| `type` | `ENUM('percentage', 'fixed')` | `NOT NULL` | Tipe potongan harga |
| `discount_value` | `BIGINT` | `NOT NULL` | Nilai diskon (dalam persentase basis point atau nominal sen) |
| `start_date` | `TIMESTAMP` | `NOT NULL` | Waktu mulai promo berlaku |
| `end_date` | `TIMESTAMP` | `NOT NULL` | Waktu promo berakhir/kedaluwarsa |
| `is_active` | `BOOLEAN` | `DEFAULT TRUE` | Toggle aktif/nonaktif promo |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu promo dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu promo diperbarui |
| `deleted_at` | `TIMESTAMP` | `NULL` | Penanda *soft delete* |

### 2.7 Tabel `shifts`
Menyimpan sesi kerja kasir dan denormalisasi total penjualan.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `cashier_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `users` |
| `opening_cash` | `BIGINT` | `NOT NULL` | Saldo modal awal laci dalam **sen** |
| `closing_cash` | `BIGINT` | `NULL` | Jumlah fisik kas akhir dalam **sen** |
| `total_sales` | `BIGINT` | `DEFAULT 0` | Denormalisasi total penjualan order di shift ini dalam **sen** |
| `status` | `ENUM('open', 'closed')` | `DEFAULT 'open'` | Status operasional shift |
| `opened_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu buka shift |
| `closed_at` | `TIMESTAMP` | `NULL` | Waktu tutup shift |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu data dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu data diperbarui |

### 2.8 Tabel `orders`
Menyimpan data induk (*header*) pesanan transaksi.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `shift_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `shifts` |
| `table_id` | `VARCHAR(36)` | `NULL, FOREIGN KEY` | Relasi ke `tables` (Boleh NULL jika *takeaway*) |
| `cashier_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `users` (Kasir yang melayani) |
| `promo_id` | `VARCHAR(36)` | `NULL, FOREIGN KEY` | Relasi ke `promos` |
| `status` | `ENUM('draft', 'pending_payment', 'paid', 'cancelled')` | `NOT NULL, DEFAULT 'draft'` | Status pembayaran order |
| `subtotal` | `BIGINT` | `NOT NULL` | Subtotal harga dalam **sen** |
| `discount_amount` | `BIGINT` | `DEFAULT 0` | Potongan diskon dalam **sen** |
| `total` | `BIGINT` | `NOT NULL` | Total akhir belanja setelah diskon dalam **sen** |
| `payment_method` | `VARCHAR(50)` | `NULL` | Metode bayar (misal: "cash", "midtrans") |
| `midtrans_order_id` | `VARCHAR(100)` | `NULL, UNIQUE` | Referensi ID order eksternal Midtrans |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu order dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu order diperbarui |

### 2.9 Tabel `order_items`
Menyimpan baris produk yang dipesan dengan *snapshot* harga saat itu.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `order_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `orders` |
| `product_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `products` |
| `quantity` | `INT` | `NOT NULL` | Jumlah unit yang dipesan |
| `price` | `BIGINT` | `NOT NULL` | **Snapshot** harga saat transaksi terjadi (dalam **sen**) |
| `subtotal` | `BIGINT` | `NOT NULL` | Harga $\times$ Kuantitas (dalam **sen**) |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu baris dibuat |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu baris diperbarui |

### 2.10 Tabel `payments`
Menyimpan rekaman riwayat pembayaran dan payload integrasi Midtrans.

| nama_kolom | tipe_data | constraint | keterangan |
|---|---|---|---|
| `id` | `VARCHAR(36)` | `PRIMARY KEY` | UUID v4 unik |
| `order_id` | `VARCHAR(36)` | `NOT NULL, FOREIGN KEY` | Relasi ke `orders` |
| `midtrans_order_id` | `VARCHAR(100)` | `NOT NULL` | ID Order Midtrans |
| `midtrans_transaction_id` | `VARCHAR(100)` | `NULL` | Transaction ID dari respons Midtrans |
| `status` | `ENUM('pending', 'paid', 'failed', 'expired')` | `NOT NULL, DEFAULT 'pending'` | Status proses pembayaran |
| `amount` | `BIGINT` | `NOT NULL` | Jumlah tagihan (dalam **sen**) |
| `payment_type` | `VARCHAR(50)` | `NULL` | Tipe pembayaran (qris, gopay, dll) |
| `raw_notification` | `JSON` | `NULL` | Payload webhook mentah dari Midtrans untuk audit |
| `paid_at` | `TIMESTAMP` | `NULL` | Waktu pembayaran terkonfirmasi lunas |
| `created_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP` | Waktu proses bayar diinisiasi |
| `updated_at` | `TIMESTAMP` | `DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP` | Waktu data diperbarui |

---

## 3. Relasi Antar Tabel (Foreign Keys)

*   `users` (id) &rarr; `orders` (`cashier_id`)
*   `users` (id) &rarr; `shifts` (`cashier_id`)
*   `users` (id) &rarr; `stock_movements` (`created_by`)
*   `categories` (id) &rarr; `products` (`category_id`)
*   `products` (id) &rarr; `order_items` (`product_id`)
*   `products` (id) &rarr; `stock_movements` (`product_id`)
*   `orders` (id) &rarr; `order_items` (`order_id`)
*   `orders` (id) &rarr; `payments` (`order_id`)
*   `shifts` (id) &rarr; `orders` (`shift_id`)
*   `tables` (id) &rarr; `orders` (`table_id`)
*   `promos` (id) &rarr; `orders` (`promo_id`)

---

## 4. Keputusan Desain yang Penting (Design Decisions)

1.  **UUID (`VARCHAR 36`) vs `AUTO_INCREMENT`**:
    *   Penggunaan UUID mencegah celah keamanan IDOR (*Insecure Direct Object Reference*). Pihak luar tidak bisa menebak jumlah transaksi bisnis atau memanipulasi ID order karena UUID sangat acak.
    *   Sangat bersahabat untuk sistem terdistribusi; Backend Go dapat men-*generate* UUID secara asinkron sebelum memasukkan (*insert*) data ke MySQL tanpa takut terjadi konflik ID.
2.  **Uang Menggunakan `BIGINT` dalam Satuan Sen**:
    *   Disimpan tanpa desimal. Harga Rp 25.000 disimpan sebagai `2500000`. Ini menghindari *rounding error* yang sering terjadi pada tipe data `FLOAT` atau `DOUBLE` (*floating-point imprecision*).
    *   API Midtrans bekerja paling baik dan akurat menggunakan tipe integer/sen.
3.  **Penggunaan `ENUM` untuk Kolom Status**:
    *   Memastikan hanya *value* tertentu yang valid di level mesin database (seperti `'draft'`, `'paid'`, `'pending_payment'`). Ini menjaga integritas data tanpa perlu validasi berulang di layer aplikasi yang berisiko terlewat.
4.  **Snapshot Harga (`price`) pada Tabel `order_items`**:
    *   Harga di- *copy* ke dalam rincian pesanan. Jika harga produk naik di masa depan di tabel `products`, laporan total transaksi masa lalu di `orders` tidak akan ikut berubah karena menggunakan data historis yang sudah di-*snapshot*.
5.  **Soft Delete (`deleted_at`) dan `is_active` Toggle**:
    *   `deleted_at` digunakan pada entitas master (produk, kategori, meja) agar pesanan historis yang memuat entitas tersebut tidak kehilangan referensi datanya.
    *   `is_active` di tabel `users` digunakan khusus sebagai saklar otorisasi (mencegah kasir login), di mana akun tetap berwujud entitas valid untuk laporan *shift* tanpa perlu menganggap datanya "dihapus".
6.  **Tabel `stock_movements` sebagai Buku Besar Log**:
    *   Daripada hanya bergantung pada kolom angka stok di tabel `products`, sistem ini mencatat setiap histori pergerakan barang. Perhitungan sisa barang bisa diaudit mutlak 100% dan terhindar dari *race conditions*.
7.  **Penyimpanan Payload Mentah (`raw_notification`) di `payments`**:
    *   Sangat krusial saat bekerja dengan Payment Gateway pihak ketiga (Webhook). Menyimpan payload lengkap (*raw JSON*) memudahkan tim backend dalam melakukan *debugging* jika terjadi perbedaan status (*dispute*).