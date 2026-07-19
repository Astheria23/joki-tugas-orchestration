# Joki Tugas AI: Multi-Agent System (MAS) Orchestrator

Joki Tugas AI adalah platform otomatisasi pengerjaan tugas akademik berbasis Multi-Agent System (MAS). Platform ini bertindak sebagai otak utama (API Gateway & Pipeline Manager) yang menerjemahkan bahasa natural dari pengguna, menyusun urutan rencana tugas (pipeline), dan mendistribusikannya ke 22 agen mikro secara sekuensial dan asinkron.

---

## 🚀 Fitur Utama & Arsitektur
1. **API Gateway & Router LLM:** Backend berbasis Go (Gin Gonic) yang memproses input user menggunakan model LLM (DeepSeek) untuk menyusun alur kerja agen.
2. **WebSocket Logs Stream:** Menyalurkan log progres real-time langsung ke antarmuka pengguna dari setiap agen yang berjalan di background.
3. **Smart Skip (Type-Safe Circuit Breaker):** Jika salah satu agen mengalami kegagalan, orkestrator akan mengecek tipe data input-output (Contract-Safe) dan melakukan bypass secara otomatis (Smart Skip) jika tipe data kompatibel, atau melakukan penghentian darurat (Hard Stop).
5. **JWT Authentication:** Pembatasan akses gateway operator menggunakan otentikasi token JWT.
6. **State Manager:** Penyimpanan data histori eksekusi, log obrolan, dan status tugas di MongoDB.

---

## 📊 Tabel Integrasi Agen Mandiri (22 Agents Tracking)

Tabel ini digunakan untuk melacak status kesiapan dan pengujian 22 agen mandiri yang didelegasikan oleh orkestrator:

| No | Agent | Author | Status |
| :--- | :--- | :--- | :--- |
| 1 | Agent Citation & Reference | Hisyam | 🟡 Pending Verification / Simulated Fallback |
| 2 | Agent Translator | Hekal | 🟢 **Integrated (Active Vercel URL)** |
| 3 | Agent Parafrase | Viola | 🟢 **Integrated (Active Vercel URL)** |
| 4 | Agent Database Querier | Adit | 🟢 **Integrated (Active Vercel URL)** |
| 5 | Agent Summarizer | Qila | 🟡 Pending Verification / Simulated Fallback |
| 6 | Agent Outliner | Eka | 🟡 Pending Verification / Simulated Fallback |
| 7 | Agent Simulasi Tanya-Jawab | Ferdy | 🟡 Pending Verification / Simulated Fallback |
| 8 | Agent PPT Generator | Indra | 🟡 Pending Verification / Simulated Fallback |
| 9 | Agent Joki Programmer | Aghni | 🟡 Pending Verification / Simulated Fallback |
| 10 | Agent Task Requirement Analyzer | Srh | 🟡 Pending Verification / Simulated Fallback |
| 11 | Agent Diagram Builder | Ahmad | 🟡 Pending Verification / Simulated Fallback |
| 12 | Agent Fact Checker | Razif | 🟡 Pending Verification / Simulated Fallback |
| 13 | Agent Typo Checker | Dina | 🟡 Pending Verification / Simulated Fallback |
| 14 | Agent Data Mining | Dwi | 🟡 Pending Verification / Simulated Fallback |
| 15 | Agent Dokumen Comparator (PR Reviewer) | Fadel | 🟡 Pending Verification / Simulated Fallback |
| 16 | Agent Format & Layout Formatter PDF | Restu | 🟡 Pending Verification / Simulated Fallback |
| 17 | Agent Web Scrapper | Octa | 🟢 **Integrated (Active Local Python Service)** |
| 18 | Agent Literature Reviewer | Iqbal | 🟡 Pending Verification / Simulated Fallback |
| 19 | Agent Pembuat Kesimpulan & Rekomendasi | Farid | 🟡 Pending Verification / Simulated Fallback |
| 20 | Agent Prompt Generator | Hilmi | 🟡 Pending Verification / Simulated Fallback |
| 21 | Agent QA & Bug Hunter | Rizki | 🟡 Pending Verification / Simulated Fallback |
| 22 | Agent Essay Writer | Reyhan | 🟡 Pending Verification / Simulated Fallback |

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
