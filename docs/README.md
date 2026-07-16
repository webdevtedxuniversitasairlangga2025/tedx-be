# Dokumentasi Backend TEDx Universitas Airlangga

Selamat datang. Folder ini berisi panduan resmi untuk mengembangkan dan memelihara
backend TEDx UNAIR. Tujuannya sederhana: **siapa pun di tim bisa menambah fitur baru
dengan pola yang sama**, sehingga codebase tetap konsisten, mudah direview, dan mudah
dirawat seiring tim bertumbuh.

## Untuk siapa dokumen ini

- **Anggota tim baru** — baca berurutan dari atas untuk memahami cara kerja project.
- **Developer fitur** — saat membuat endpoint baru, ikuti [SOP Membuat Modul CRUD](./sop-create-crud-module.md) langkah demi langkah.
- **Reviewer / lead** — gunakan checklist di tiap dokumen sebagai acuan saat menyetujui Pull Request.

## Daftar dokumen

| Dokumen | Isi |
|---------|-----|
| [Development Workflow](./development-workflow.md) | Cara kerja harian: setup lokal, arsitektur request, alur Git & Pull Request, testing, dan penggunaan graph pengetahuan. |
| [SOP Membuat Modul CRUD](./sop-create-crud-module.md) | Langkah baku membuat fitur baru dengan endpoint Create, GetAll, GetByID, Update, Delete — mengikuti pola modul `todo`. |
| [Skema Database](./database-schema.md) | Arsitektur database dan relasi antarentitas sebagai fondasi project |

## Prinsip utama

Ada satu aturan yang mengikat semua dokumen di sini:

> **Modul `todo` adalah contoh acuan (reference implementation).**
> Ketika ragu bagaimana menulis sesuatu, buka `modules/todo/` dan tiru polanya.
> Jangan menciptakan pola baru tanpa kesepakatan tim.

Konsistensi lebih penting daripada preferensi pribadi. Kode yang "kreatif" tapi berbeda
dari pola tim justru memperlambat review dan mempersulit pemeliharaan.

## Menjaga dokumen tetap akurat

Dokumen ini hidup bersama kode. Jika sebuah pola berubah (misalnya format response
envelope atau cara dependency injection), **perbarui dokumen terkait di Pull Request yang sama**.
Dokumentasi yang usang lebih berbahaya daripada tidak ada dokumentasi, karena
menyesatkan tim.
