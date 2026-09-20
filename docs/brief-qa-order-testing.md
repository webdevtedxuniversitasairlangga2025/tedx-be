# Brief QA: Testing Order — War Ticket & Hold 15 Menit

**Untuk:** Anggota QA yang ditugaskan — **Fokus:** Pembelian tiket manual QRIS static, hold race
**Build:** `go vet/build` 0, Bruno `API_Test/Order/` siap, DB `AutoMigrate` aman

## 1. Tujuan
Pastikan 1 buyer 1 email beli 1..5 tiket tidak oversell, hold 15m lepas otomatis, approve 1 email berisi 3 QR, dan 2 user war sisa 1 tiket hanya 1 sukses.

## 2. Prasyarat (siapkan 1x, admin)
- Login admin: `admin@tedxunair.com` (role admin) → copy `access_token` ke `environments/Dev.yml`.
- Buat tier war: `POST /tickets` + `POST /tickets/:id/tiers` body `{"tier":"war-test","price":"100000.00","quota":3}` → catat `tier_id`. Atau set existing tier `PATCH /tickets/:id/tiers/:tierId {"quota":3}`.
- Buat 2 user buyer: `POST /auth/register` userA/userB → login → catat `TOKEN_A`, `TOKEN_B`.
- Base URL: `http://localhost:8888/api/v1` (dev) atau `https://api.tedxuniversitasairlangga.com/api/v1` (prod via `docker exec`).

## 3. Flow yang diuji (end-to-end)
`GET /tickets` → `POST /orders` (hold) → QRIS + `total` + countdown → `PATCH /orders/:id/proof {payment_proof_url}` (upload screenshot) → `GET /orders/:id` cek `payment_proof_url` → `GET /orders/admin/all` lihat bukti → `PATCH approve/reject` → email 1 berisi 3 kode → `GET /orders/:id` → check-in scan.

## 4. Test Cases — checklist wajib centang
| ID | Skenario | Langkah | Expected |
|----|----------|---------|----------|
| TC01 | Happy hold | `POST /orders {tier, qty:2}` TOKEN_A | 201 `status awaiting_approval`, `expired_at ~+15m`, `total 200000.00`, `GET /tickets` `quota_left -2` |
| TC02 | Avail 2 beli 3 | Setelah TC01 (left 1), `POST qty:3` | 400 `quota exceeded`, `held` tetap |
| TC03 | War 2 buyer qty2 sisa 3 | `TIER quota 3` → 2 terminal bareng `POST qty:2` (TOKEN_A & B `&`) | Tepat 1x 201, 1x 400 `quota exceeded`, DB `held=2 left=1` |
| TC04 | War 3 buyer qty1 sisa 3 | 3 terminal `qty:1` bareng | 3x 201 `held 3`, extra ke-4 `qty:1` → 400 |
| TC05 | Upload proof happy | `POST qty:2` → `PATCH /orders/:id/proof {"payment_proof_url":"https://example.com/bukti.jpg"}` | 200 `payment_proof_url` terisi, `GET /orders/:id` ada url |
| TC06 | Upload proof expired | `POST qty:1` → `UPDATE expired_at` lalu `PATCH proof` | 400 `order expired` |
| TC07 | Approve race | 1 order `qty:2` (sudah upload proof) → 2 admin `PATCH approve` bareng | 1x 200 `paid`, 1x 400 `order not awaiting approval`, `filled 2 held 0` |
| TC08 | Approve happy 1 email 3 QR | `qty:3` upload proof → admin approve | 200 `paid` + 3 `ticket_code` unik, 1 email ke buyer berisi 3 kode, `GET /orders/:id` 3 kode |
| TC09 | Reject lepas hold | `qty:2` → admin `PATCH reject {"reason":"nominal kurang"}` | 200 `rejected`, `held 0`, `left` balik |
| TC10 | Expiry auto 15m | `qty:2` → `UPDATE orders SET expired_at=now()-1m` → tunggu 60s | `status expired`, `held 0`, next buyer bisa beli lagi |
| TC11 | Validasi qty | `qty:0` / `qty:6` | 400 `quantity must be between` |
| TC12 | Validasi proof url | `PATCH proof {"payment_proof_url":"not-a-url"}` | 400 `failed get data from body` |
| TC13 | Isolasi order | buyerA buat order → buyerB `GET /orders/:id` order A | 404 `order not found`, `GET /orders` buyerB hanya lihat miliknya |
| TC14 | Isolasi proof | buyerA order → buyerB `PATCH proof` order A | 404 `order not found` |
| TC15 | Sale window | tier `sale_end` lewat → `POST qty:1` | 400 `sale ended` |
| TC16 | Tier inactive | `PATCH tier is_active:false` → `POST` | 400 `tier inactive` |

## 5. Langkah War 2 Terminal (copy-paste)
```bash
TIER="PASTE_TIER_UUID"
curl -s -X POST http://localhost:8888/api/v1/orders -H "Authorization: Bearer $TOKEN_A" -H "Content-Type: application/json" -d "{\"ticket_tier_id\":\"$TIER\",\"quantity\":2}" | jq . &
curl -s -X POST http://localhost:8888/api/v1/orders -H "Authorization: Bearer $TOKEN_B" -H "Content-Type: application/json" -d "{\"ticket_tier_id\":\"$TIER\",\"quantity\":2}" | jq . &
wait
# cek: GET /tickets/:id → quota_left harus -2 bukan -4
```
Bruno: `Order/Create Order` ganti `tier_id`/`quantity` → Run 2 tab bareng.

## 6. Data yang dicatat per TC
- `order_number`, `status`, `expired_at`, `payment_proof_url`, `held/filled/left` via `GET /tickets`, response code, `ticket_code` uniqueness.

## 7. Kriteria lulus
- Tidak pernah `quota_left` minus / oversell, war selalu 1 sukses 1 gagal, approve selalu 1 email berisi N kode, expired selalu lepas dalam 1 menit, isolasi 404 benar.

## 8. Lapor
Buat sheet: TC ID | Hasil | Bukti (screenshot Bruno / curl output) | `quota_left` sebelum/sesudah. Tag aku jika TC03/TC05 gagal — itu race.

**Catatan:** SMTP `failed open .env` di log test = email tidak terkirim tapi `paid` tetap — cek fallback `GET /orders/:id` 3 kode tetap ada.
