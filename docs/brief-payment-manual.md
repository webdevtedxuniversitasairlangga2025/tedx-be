# Brief: Perubahan Pembayaran Midtrans → Manual QRIS Static + Hold 15 Menit

**Tanggal:** 20 Sep 2026 — **Status:** Implemented, `go vet/build` 0, siap testing
**Penulis:** Senior Engineer — untuk tim BE/FE/QA TEDx UNAIR

## 1. Kenapa berubah
Midtrans registrasi bermasalah → pindah ke **QRIS Static manual** (1 QR untuk semua transaksi, nominal di-input buyer). Trade-off: tidak ada webhook, verifikasi 100% manual via mutasi bank oleh admin. Time-to-market lebih cepat, tanpa MDR 0.7%, tanpa dependency payment gateway.

## 2. Flow baru (1 buyer 1 email, beli 3 = 1 email berisi 3 QR, wajib upload bukti)
1. User `GET /tickets` lihat `quota_left = quota - filled - held`.
2. `POST /orders {ticket_tier_id, quantity:1..5}` → hold `held+=qty`, `expired_at=now+15m`, status `awaiting_approval`, insert 3 `attendee_tickets` (kode unik beda, data buyer).
3. FE tampil QRIS static + total `price*qty` besar + countdown 15:00 + `order_number`.
4. Buyer scan QRIS via e-wallet, input nominal manual.
5. Buyer screenshot bukti TF → `PATCH /orders/:id/proof {payment_proof_url:"https://..."}` (FE upload image ke storage dapat URL, BE simpan `payment_proof_url`).
6. Admin `GET /orders/admin/all?status=awaiting_approval` → lihat `payment_proof_url` thumbnail + `total` cocokkan mutasi → `PATCH /orders/:id/approve` → `held-=qty, filled+=qty, status=paid` → kirim **1 email** ke buyer berisi list 3 kode + QR. `Reject` → `held-=qty, status=rejected`. Lewat 15m tanpa approve → `status=expired` auto via cron 1 menit → `held` lepas.
7. Buyer `GET /orders/:id` lihat 3 kode (`payment_proof_url` ikut terexpose), check-in scan `ticket_code` per orang.

## 3. Apa yang berubah di kode
- **DB:** `ticket_tiers.quota_held INT NOT NULL DEFAULT 0` (hold), `orders.approved_by/approved_at/rejected_reason/payment_proof_url` (audit + bukti). Kolom Midtrans lama dibiarkan nullable, tidak hapus. `AutoMigrate` aman.
- **Constants:** `awaiting_approval`, `rejected` baru.
- **Module baru `order`:** `POST /orders`, `GET /orders`, `GET /orders/:id`, `PATCH /orders/:id/proof` (user upload `payment_proof_url` url), `GET /orders/admin/all`, `PATCH /orders/:id/approve`, `PATCH /orders/:id/reject`. Semua ikut `routes→handler→service→repository→entities`, pakai `utils.BuildResponse*`, `decimal` string untuk uang.
- **Ticket:** `GET /tickets` sekarang `quota_left = quota - filled - held` dan `quota_held` terexpose.
- **Expiry:** ticker 1 menit `ReleaseExpiredHolds()` di `providers/core.go`.
- **Email:** 1 email gabungan per approve, fallback `GET /orders/:id` jika SMTP fail.
- **Bruno:** `API_Test/Order/` 8 request (Create, Get My, Get By ID, Upload Proof, Get All Admin, Approve, Reject).

## 4. Old vs New
| Aspek | Midtrans (lama) | Manual baru |
|-------|---------------|-------------|
| QR | Dynamic per transaksi, nominal terkunci | Static 1 QR, nominal input manual |
| Status | `pending → paid` via webhook | `awaiting_approval → paid/rejected/expired` via admin/cron |
| Hold | `filled` naik saat webhook | `held` naik saat create, pindah ke `filled` saat approve |
| Validasi nominal | Otomatis gateway | Manual cek `total_amount` di dashboard admin |

## 5. Race & hold
Dua user rebut sisa 1 tiket → `SELECT ... FOR UPDATE` di baris tier → yang kedua `quota exceeded` 400. `available` hitung real, tidak oversell. Test war sudah PASS (lihat QA).

## 6. Cara testing (untuk anggota yang ditugaskan)
- **Bruno:** isi `tier_id`, `quantity`, `{{access_token}}` di `Order/Create Order`, hit 2 terminal bareng `curl ... &` → 1 sukses 1 gagal.
- **Avail 2 beli 3:** harus 400 `quota exceeded`.
- **Upload bukti:** `PATCH /orders/:id/proof {"payment_proof_url":"https://..."}` sebelum approve → `GET /orders/:id` harus ada `payment_proof_url`. Cek admin `GET /admin/all` lihat thumbnail.
- **Expiry:** buat order lalu `UPDATE orders SET expired_at=now()-1m`, tunggu 1 menit → `held` lepas.
- **Admin:** login `admin@tedxunair.com`, `PATCH approve` cek email 1 berisi 3 kode.

## 7. Yang tidak berubah
- `auth` OTP, `bundle/merchandise/categories` display, `user` admin guard, `todo` reference. Manual admin via `UPDATE users SET role='admin'` tetap.

## 8. Tanggung jawab
- FE: tampilkan `total_amount` besar + countdown `expired_at`, QRIS image static + form upload screenshot → `PATCH proof` (upload ke storage dapat URL).
- BE: jangan ubah `quotaHeld` manual via psql; `payment_proof_url` wajib url valid max 500.
- QA: war 2 terminal, periksa `quota_left` tidak minus, upload bukti → `payment_proof_url` muncul, approve kirim 1 email berisi 3 kode.
- Ops: pastikan `SMTP_*` di `.env.production`, pantau `held` tidak gantung, storage URL bukti dapat diakses admin.

Kontak: tanya di thread ini jika `quota exceeded` tidak muncul atau email tidak 1 gabungan.
