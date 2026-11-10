# Product Requirements Document (PRD)
# Coffee Shop Point of Sales (POS) System

| Dokumen | Spesifikasi Kebutuhan Produk (PRD) |
|---|---|
| **Nama Proyek** | Coffee Shop POS System |
| **Tech Stack Utama** | Go (Golang) & MySQL |
| **Payment Gateway** | Midtrans Snap API |
| **Target Pengguna** | Bisnis Coffee Shop Skala Menengah |
| **Versi Dokumen** | 1.1.0 |
| **Status** | Final Draft |

---

## 1. Deskripsi Produk

Sistem **Point of Sales (POS)** ini adalah aplikasi berbasis API (Backend) yang dirancang khusus untuk memenuhi alur kerja dan kebutuhan spesifik kedai kopi (coffee shop) skala menengah. Sistem mencakup pengelolaan operasional transaksi harian kasir, manajemen meja (dine-in / takeaway), pencatatan pergerakan inventaris/stok bahan & produk, shift kerja kasir, integrasi pembayaran digital via Midtrans Snap, serta dashboard analitik dan pelaporan komprehensif bagi pemilik usaha.

---

## 2. User Roles

Sistem membagi pengguna ke dalam 2 (dua) role utama:

### 2.1 Owner
- Bertanggung jawab atas konfigurasi master data (kategori, produk, meja, promo, akun staf).
- Memantau kesehatan finansial kafe melalui grafik revenue, produk terlaris, dan rekap per kasir.
- Mengontrol rantai pasok stok produk dan mengunduh laporan periodik.
- Tidak bisa dihapus dari sistem untuk menjaga akses root/admin.

### 2.2 Cashier
- Bertanggung jawab atas pencatatan order pemesanan di meja kasir.
- Wajib membuka shift dengan deklarasi modal awal laci kas dan menutup shift dengan rekonsiliasi kas riil.
- Menghubungkan pesanan ke meja pelanggan, menerapkan promo valid, dan menghasilkan tagihan pembayaran via Midtrans Snap atau tunai.
- Jika akun dinonaktifkan, kasir tidak bisa login, namun data historisnya tetap ada.

---

## 3. Fitur Owner

### 3.1 Manajemen Produk
Mengelola seluruh katalog menu kafe (CRUD, soft delete, foto, status aktif/nonaktif).
**Acceptance Criteria:**
- [ ] Owner dapat membuat produk baru dengan validasi data wajib (Nama, Kategori, Harga Jual > 0).
- [ ] Owner dapat mengunggah file foto produk (format `.jpg`, `.jpeg`, `.png`, maksimal 2MB).
- [ ] Owner dapat mengubah status produk menjadi `Nonaktif`; produk nonaktif tidak bisa ditambahkan di pesanan baru.
- [ ] Saat produk dihapus, sistem menerapkan *soft delete*. Data pesanan terdahulu yang memuat produk tersebut tetap utuh.

### 3.2 Manajemen Kategori Produk
Pengelompokan menu untuk mempermudah navigasi staf dan filter pelaporan.
**Acceptance Criteria:**
- [ ] Owner dapat membuat kategori baru dengan nama unik.
- [ ] Owner dapat mengubah nama kategori tanpa mengubah relasi produk di dalamnya.
- [ ] Kategori tidak dapat dihapus jika masih ada produk aktif yang berelasi dengannya.

### 3.3 Manajemen Stok
Pengelolaan jumlah stok fisik produk dan pencatatan riwayat pergerakan stok.
**Acceptance Criteria:**
- [ ] Owner dapat melakukan *stock adjustment* (tambah/kurang) disertai kolom catatan alasan yang WAJIB diisi.
- [ ] Setiap perubahan stok wajib tercatat pada histori pergerakan stok (memuat: produk, tipe mutasi, kuantitas, catatan, waktu).
- [ ] Sistem menampilkan riwayat mutasi stok per produk yang dapat difilter.

### 3.4 Manajemen Meja
Pengaturan ketersediaan meja makan di area kafe.
**Acceptance Criteria:**
- [ ] Owner dapat menambah, mengedit, dan menonaktifkan meja dengan nomor/nama meja yang unik.
- [ ] Status meja berganti otomatis menjadi `Terisi` saat ada order aktif, dan kembali `Tersedia` setelah transaksi lunas.

### 3.5 Manajemen User Cashier
Pengelolaan kredensial akun staf operasional kasir.
**Acceptance Criteria:**
- [ ] Owner dapat mendaftarkan akun kasir baru dengan kredensial yang unik.
- [ ] Password kasir dapat di-reset oleh Owner.
- [ ] Owner dapat menonaktifkan akun kasir (memblokir akses login baru).

### 3.6 Dashboard Laporan
Pusat visualisasi performa penjualan kafe.
**Acceptance Criteria:**
- [ ] Halaman dashboard menampilkan Total Revenue, Total Transaksi, revenue harian/mingguan/bulanan.
- [ ] Menampilkan Top 5 / Top 10 produk terlaris berdasarkan kuantitas dan nilai revenue.
- [ ] Menampilkan breakdown jumlah transaksi dan revenue per kasir.

### 3.7 Manajemen Promo
Pengaturan skema promosi (persentase dan nominal).
**Acceptance Criteria:**
- [ ] Owner dapat membuat promo dengan tipe Persentase (dengan batas max diskon) atau Nominal tetap.
- [ ] Promo dapat diatur periode masa berlakunya (tanggal mulai dan berakhir).
- [ ] Promo dapat diatur kuota batas penggunaannya.

### 3.8 Export CSV
Fitur pembukuan eksternal dan rekonsiliasi akuntansi.
**Acceptance Criteria:**
- [ ] Owner dapat memfilter rentang tanggal sebelum melakukan export.
- [ ] Sistem dapat menghasilkan file `.csv` untuk laporan transaksi yang kompatibel dengan Excel/Google Sheets.

---

## 4. Fitur Cashier

### 4.1 Shift Management
Menjamin akuntabilitas uang fisik (buka shift dengan modal kas, tutup shift dengan rekap).
**Acceptance Criteria:**
- [ ] Kasir dapat melakukan "Buka Shift" dengan menginput modal kas awal.
- [ ] Kasir dapat melakukan "Tutup Shift" dengan menginput jumlah fisik kas akhir (menampilkan selisih kas).
- [ ] Kasir tidak bisa membuat order baru jika belum ada shift yang berstatus buka/aktif.

### 4.2 Buat Transaksi
Alur pencatatan pesanan pelanggan (pilih meja, produk, quantity).
**Acceptance Criteria:**
- [ ] Kasir dapat memilih nomor meja yang berstatus `Tersedia`.
- [ ] Kasir dapat menambah produk ke keranjang belanja dan mengubah quantity.
- [ ] Kasir TIDAK dapat menambah produk jika stok tidak mencukupi (Out-of-stock guard).
- [ ] Subtotal ter-update secara otomatis sesuai harga dan quantity.

### 4.3 Apply Promo
Pengaplikasian promo aktif ke transaksi aktif.
**Acceptance Criteria:**
- [ ] Kasir dapat memasukkan / memilih satu promo aktif ke keranjang.
- [ ] Sistem menolak promo jika kriteria (minimum transaksi, kuota, atau masa berlaku) tidak terpenuhi.
- [ ] Nilai potongan diskon langsung memotong total tagihan.

### 4.4 Checkout Midtrans
Integrasi pembayaran menggunakan Midtrans Snap.
**Acceptance Criteria:**
- [ ] Kasir dapat men-trigger checkout untuk menghasilkan URL/Token Midtrans Snap.
- [ ] Sistem merespons webhook dari Midtrans; saat berstatus settlement/capture, transaksi otomatis di-update menjadi `PAID`.
- [ ] Order yang sudah masuk tahap checkout (token generated) tidak bisa diubah lagi itemnya.

### 4.5 Riwayat Transaksi
Melihat histori transaksi shift hari ini.
**Acceptance Criteria:**
- [ ] Kasir dapat melihat daftar transaksi (beserta statusnya: PAID, PENDING, CANCEL) khusus untuk shift miliknya yang sedang aktif.
- [ ] Kasir dapat melihat detail item dari suatu riwayat transaksi.

---

## 5. Business Rules

Aturan bisnis yang HUKUMNYA WAJIB ditegakkan oleh sistem:

**Tentang Stok:**
- Stok **hanya berkurang** setelah pembayaran *confirmed* via webhook Midtrans (bukan saat order dibuat, bukan saat checkout di-trigger).
- Cashier **tidak bisa** menambah item ke order kalau stok tidak mencukupi.
- Owner bisa *adjustment* stok manual, namun **wajib mencantumkan alasan** yang dicatat.

**Tentang Shift:**
- Cashier **harus** buka shift sebelum bisa membuat transaksi apapun.
- Satu cashier **hanya bisa** punya satu shift aktif di satu waktu.
- Shift yang sudah ditutup **tidak bisa** dibuka kembali.

**Tentang Transaksi:**
- Order yang sudah checkout (token Midtrans terbentuk) **tidak bisa diubah** itemnya.
- Satu transaksi **hanya bisa** menggunakan maksimal satu promo (non-stackable).
- Promo **hanya bisa** diaplikasikan jika masih dalam periode aktif dan belum mencapai batas penggunaan (kuota).

**Tentang User:**
- User dengan role Owner **tidak bisa dihapus** dari sistem.
- Cashier yang dinonaktifkan **tidak bisa login**, tapi data transaksinya tetap ada secara permanen.

---

## 6. Out of Scope

Fitur-fitur berikut **TIDAK DIBANGUN** di versi ini untuk menjaga fokus pada MVP (Minimum Viable Product):

1. **Aplikasi mobile (iOS/Android)** — Sistem ini hanya berbasis web/API.
2. **Multi-outlet** — Sistem ini didesain khusus untuk satu outlet tunggal.
3. **Loyalty program dan membership pelanggan** — Poin reward atau keanggotaan pelanggan belum didukung.
4. **Integrasi dengan mesin kasir atau printer struk hardware** — Tidak ada komunikasi level driver ke hardware printer/laci kas otomatis.
5. **Manajemen karyawan selain user cashier** — HRIS, absensi barista, atau manajemen staf dapur tidak termasuk.
6. **Fitur reservasi meja** — Pelanggan tidak bisa mem-booking meja untuk masa depan; meja dikelola secara *go-show*.
7. **Frontend** — Scope utama proyek ini fokus murni pada pengembangan API/Backend.