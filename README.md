# token-share

Aplikasi kecil berbasis Go untuk menampilkan sebuah token di layar penuh, langsung di browser. Cocok untuk menampilkan kode OTP, PIN, atau token sementara agar mudah dibaca dari jarak jauh di layar yang sama.

Token diketik dari terminal, lalu halaman web menampilkannya dengan ukuran huruf yang otomatis menyesuaikan jendela. Saat token diganti, semua tab yang terbuka ikut ter-update secara real-time tanpa perlu refresh.

## Fitur

- **Tampilan layar penuh** — token ditampilkan besar, di tengah layar, dengan latar gelap.
- **Responsif** — ukuran teks dihitung otomatis agar muat di jendela, baik token pendek maupun panjang.
- **Update real-time** — perubahan token dikirim ke browser lewat Server-Sent Events (SSE).
- **Latar ambient dinamis** — gumpalan cahaya beranimasi lembut di belakang teks (otomatis nonaktif jika pengguna memilih *reduced motion*).
- **Penghitung umur token** — footer menampilkan berapa menit token sudah aktif.
- **Buka browser otomatis** — di Windows, halaman langsung dibuka setelah server berjalan.
- **Tanpa dependensi eksternal** — hanya memakai standard library Go.

## Prasyarat

- [Go](https://go.dev/dl/) 1.27 atau lebih baru (lihat `go.mod`)
- Browser modern yang mendukung `EventSource` (SSE)
- Windows, macOS, atau Linux. Fitur buka-browser-otomatis memakai `rundll32` sehingga hanya berjalan di Windows; di sistem lain buka URL-nya secara manual.

## Menjalankan

```bash
go run .
```

Atau build lalu jalankan binary-nya:

```bash
go build -o tokenshare.exe .
./tokenshare.exe
```

Server berjalan di `http://127.0.0.1:8080` dan akan mencoba membuka browser secara otomatis.

## Cara pakai

1. Jalankan aplikasi, lalu masukkan token pertama saat diminta:
   ```
   Masukkan token: 123456
   ```
2. Browser terbuka dan menampilkan token.
3. Untuk mengganti token, ketik token baru di terminal lalu tekan Enter:
   ```
   Masukkan token baru: 654321
   ```
4. Semua tab yang sedang membuka halaman akan langsung berubah.

Menekan Enter dengan input kosong akan ditolak; token tidak boleh kosong.

## Endpoint

| Metode | Path      | Keterangan                                            |
| ------ | --------- | ----------------------------------------------------- |
| GET    | `/`       | Halaman utama yang menampilkan token                  |
| GET    | `/events` | Stream SSE berisi token terbaru untuk update real-time |

Metode selain `GET` akan dibalas `405 Method Not Allowed`.

## Struktur proyek

```
.
├── main.go      # Seluruh aplikasi: server, template halaman, dan logika token
├── go.mod       # Definisi module Go
└── .gitignore   # Abaikan binary, secret, file OS/editor
```

## Catatan

- Token disimpan hanya di memori; tidak ada persistensi ke disk, dan token hilang saat aplikasi ditutup.
- Server hanya mendengarkan di `127.0.0.1`, jadi halaman tidak dapat diakses dari perangkat lain di jaringan.
- Port `8080` masih hard-code. Jika port sedang dipakai, aplikasi akan berhenti dengan pesan kesalahan.
- Halaman ini tidak memiliki autentikasi karena diperuntukkan bagi pemakaian lokal.

## Lisensi

Belum ditentukan.
