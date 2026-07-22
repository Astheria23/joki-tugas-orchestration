# Joki Tugas AI: Multi-Agent System (MAS) Orchestrator

Joki Tugas AI adalah platform otomatisasi pengerjaan tugas akademik berbasis Multi-Agent System (MAS). Platform ini bertindak sebagai otak utama (API Gateway & Pipeline Manager) yang menerjemahkan bahasa natural dari pengguna, menyusun urutan rencana tugas (pipeline), dan mendistribusikannya ke 22 agen mikro secara sekuensial dan asinkron.

---

## 🚀 Fitur Utama & Arsitektur
1. **API Gateway & Router LLM:** Backend berbasis Go (Gin Gonic) yang memproses input user menggunakan model LLM (DeepSeek) untuk menyusun alur kerja agen.
2. **WebSocket Logs Stream:** Menyalurkan log progres real-time langsung ke antarmuka pengguna dari setiap agen yang berjalan di background.
3. **Smart Skip (Type-Safe Circuit Breaker):** Jika salah satu agen mengalami kegagalan, orkestrator akan mengecek tipe data input-output (Contract-Safe) dan melakukan bypass secara otomatis (Smart Skip) jika tipe data kompatibel, atau melakukan penghentian darurat (Hard Stop).
4. **Brutalist Editorial UI:** Desain dashboard modern minimalis kaku (Artoria Style) berlatar krem (`#ECE7DC`) dengan sudut 90 derajat tajam (`rounded-none`), tipografi editorial Serif (Cormorant Garamond), dan terminal konsol Monospace (JetBrains Mono).
5. **JWT Authentication:** Pembatasan akses gateway operator menggunakan otentikasi token JWT.
6. **State Manager:** Penyimpanan data histori eksekusi, log obrolan, dan status tugas di MongoDB.

---

## 📊 Tabel Integrasi Agen Mandiri (22 Agents Tracking)

Tabel ini digunakan untuk melacak status kesiapan dan pengujian 22 agen mandiri yang didelegasikan oleh orkestrator:

| No | Agent | Author | Status |
| :--- | :--- | :--- | :--- |
| 1 | System Orchestrator | Okta | 🟢 **Aktif** |
| 2 | Agent Citation & Reference | Hisyam / Syam | 🟢 **Aktif** |
| 3 | Agent Translator | Haekal | 🟢 **Aktif** |
| 4 | Agent Parafrase | Viola / Piol | 🟢 **Aktif** |
| 5 | Agent Database Querier | Adit | 🟢 **Aktif** |
| 6 | Agent Summarizer | Qila | 🟡 Verifikasi |
| 7 | Agent Outliner | Eka | 🟡 Verifikasi |
| 8 | Agent Simulasi Tanya-Jawab | Ferdy | 🔴 **Error (Contract Mismatch)** |
| 9 | Agent PPT Generator | Indra | 🟢 **Aktif** |
| 10 | Agent Joki Programmer | Aghni | 🟢 **Aktif** |
| 11 | Agent Task Requirement Analyzer (Supervisor) | Sarah | 🔴 **Error (Endpoint Timeout/Down)** |
| 12 | Agent Diagram Builder | Ahmad | 🔴 **Error (LLM Token Expired)** |
| 13 | Agent Fact Checker | Razif / Zif | 🟢 **Aktif** |
| 14 | Agent Typo Checker | Dina | 🟢 **Aktif** |
| 15 | Agent Data Mining | Dwi | 🔴 **Error (Contract Mismatch)** |
| 16 | Agent Dokumen Comparator (PR Reviewer) | Fadel | 🟡 Verifikasi |
| 17 | Agent Format & Layout Formatter PDF | Restu | 🟢 **Aktif** |
| 18 | Agent Web Scraper | Fadhail | 🔴 **Error (Contract Mismatch)** |
| 19 | Agent Literature Reviewer | Iqbal | 🔴 **Error (LLM Token Expired)** |
| 20 | Agent Pembuat Kesimpulan & Rekomendasi | Farid | 🔴 **Error (Contract Mismatch)** |
| 21 | Agent Prompt Generator | Hilmi | 🟢 **Aktif** |
| 22 | Agent QA & Bug Hunter | Iqmal / Iki | 🟢 **Aktif** |
| 23 | Agent Essay Writer | Reyhan | 🔴 **Error (Contract Mismatch)** |

---

## 📝 Log Revisi Integrasi (23 Juli 2026)

Berikut adalah catatan revisi untuk tim pembuat agen berdasarkan hasil uji coba API `(Payload Contract: task_id, agent_type, payload.raw_text)`:

| No | Agent / Author | Bukti Payload Response (Error) | Action Required (Tindakan) |
| :--- | :--- | :--- | :--- |
| 1 | **Kesimpulan Saran** (Farid) | `<pre>Status: 400<br/>{<br/>  "success": false,<br/>  "error": "Properti \"text\" atau \"workspaceFile\" wajib dikirim..."<br/>}</pre>` | Agent mengharapkan parameter bernama `"text"` di dalam request body. Tolong ubah logic kode agar mengekstrak string teks dari `payload.raw_text`. |
| 2 | **Task Requirement Analyzer** (Sarah) | `<pre>Status: Timeout 15 Detik<br/>Error: HTTPSConnectionPool(host='agent-supervisor-production...', port=443): Read timed out.</pre>` | API Server di Railway nyangkut/mati total tidak memberi respon. Tolong cek status deploy Railway apakah sedang *cold start* lama atau error *crash*. |
| 3 | **Tanya Jawab** (Ferdy) | `<pre>Status: 400<br/>{<br/>  "status": "error",<br/>  "message": "Input \"pertanyaan\" tidak boleh kosong."<br/>}</pre>` | Agent mengharapkan adanya parameter bernama `"pertanyaan"`. Tolong ubah logic kode agar mengekstrak string teks dari `payload.raw_text`. |
| 4 | **Diagram Builder** (Ahmad) | `<pre>Status: 200<br/>{<br/>  "status": "error",<br/>  "message": "LLM translation error: 404 models/gemini-1.5-flash is not found..."<br/>}</pre>` | Server merespon sukses, tetapi proses di internal error karena *Limit/Token API* Gemini habis, atau penamaan *model* salah. Tolong ganti API Key atau update library Gemini. |
| 5 | **Data Mining** (Dwi) | `<pre>Status: 422<br/>{<br/>  "detail": [<br/>    {<br/>      "type": "missing",<br/>      "loc": ["body", "metadata"],<br/>      "msg": "Field required"<br/>    }<br/>  ]<br/>}</pre>` | Agent mewajibkan adanya objek `"metadata"`. Orchestrator tidak mengirim ini. Tolong hapus mandatory validasi untuk `metadata`. |
| 6 | **Essay Writer** (Reyhan) | `<pre>Status: 400<br/>{<br/>  "status": "error",<br/>  "message": "Parameter 'topic' wajib diisi"<br/>}</pre>` | Agent mengharapkan parameter `"topic"`. Tolong ubah logic kode agar membaca *topic* langsung dari string di `payload.raw_text`. |
| 7 | **Web Scraper** (Fadhail) | `<pre>Status: 422<br/>{<br/>  "detail": [<br/>    {<br/>      "loc": ["body", "payload", "url"],<br/>      "msg": "Field required"<br/>    },<br/>    {<br/>      "loc": ["body", "metadata"],<br/>      "msg": "Field required"<br/>    }<br/>  ]<br/>}</pre>` | Agent meminta adanya `"url"` dan `"metadata"`. Tolong ambil target URL langsung dari `payload.raw_text` saja, dan abaikan `metadata`. |
| 8 | **Literature Review** (Iqbal) | `<pre>Status: 500<br/>{<br/>  "message": "Gagal memproses review: 429 You exceeded your current quota..."<br/>}</pre>` | Path API sudah benar di `/api/review`, tetapi *Limit/Token API* Gemini (`gemini-2.0-flash`) habis/melebihi limit *Free Tier*. Tolong ganti API Key Google AI Studionya dengan akun yang lain. |

---

## 🛠️ Cara Menjalankan Project (Lokal)
Repositori ini dirancang sebagai monorepo dengan task runner untuk mempermudah pengerjaan:
1. Pastikan Docker dan Go Task sudah terinstal.
2. Jalankan perintah inisialisasi awal (instalasi library frontend + database infra MongoDB):
   ```bash
   cd app/
   task init
   ```
3. Jalankan aplikasi local development (paralel Go Backend + React Frontend):
   ```bash
   task dev
   ```
4. Buka dashboard di browser: [http://localhost:5173](http://localhost:5173).
