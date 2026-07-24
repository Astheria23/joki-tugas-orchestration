# Joki Tugas AI: Multi-Agent System (MAS) Orchestrator

Joki Tugas AI (produk UI: **Bananacademic**) adalah platform otomatisasi pengerjaan tugas akademik berbasis Multi-Agent System (MAS). Platform ini bertindak sebagai otak utama (API Gateway & Pipeline Manager) yang menerjemahkan bahasa natural dari pengguna, menyusun urutan rencana tugas (pipeline), dan mendistribusikannya ke agen mikro secara sekuensial dan asinkron.

---

## Fitur utama

1. **API Gateway & Router LLM** — Backend Go (Gin) + OpenCode Go (DeepSeek) untuk klasifikasi chat vs task dan routing pipeline.
2. **Chat workspace** — Kuota chat lifetime, WebSocket progress, approval **Gas / Batal**, tombol **Stop** mid-run.
3. **Agent Registry (SSOT)** — [`shared/agents/registry.json`](shared/agents/registry.json): key, kontrak I/O, `envUrl`, `onError` (`stop` | `skip`). Dipakai Go + FE.
4. **Scraper + web search** — Orkestrator resolve topik → URL (`WEB_SEARCH_PROVIDER`) lalu panggil `web_scraper` dengan payload URL-only.
5. **Template chips** — `/template joki_makalah|joki_koding|joki_presentasi|review_tugas|analisis_data` + preview langkah di FE.
6. **JWT + MongoDB** — Auth user, task per user, histori chat & pipeline.

---

## Status integrasi agen

| No | Agent key | Author | Status |
| :--- | :--- | :--- | :--- |
| — | System Orchestrator | Okta | 🟢 Aktif |
| 1 | `web_scraper` | Fadhail | 🟢 Aktif |
| 2 | `data_mining` | Dwi | 🟢 Aktif |
| 3 | `summarizer` | Qila | 🟢 Aktif |
| 4 | `outliner` | Eka | 🟢 Aktif |
| 5 | `translator` | Haekal | 🟢 Aktif |
| 6 | `paraphrase` | Viola  | 🟢 Aktif |
| 7 | `typo_checker` | Dina | 🟢 Aktif |
| 8 | `fact_checker` | Razif  | 🟢 Aktif |
| 9 | `literature_reviewer` | Iqbal | 🟢 Aktif |
| 10 | `citation_reference` | Hisyam | 🟢 Aktif |
| 11 | `qna_simulator` | Ferdy | 🟢 Aktif |
| 12 | `requirement_analyzer` | Sarah | 🟢 Aktif |
| 13 | `diagram_builder` | Ahmad | 🟢 Aktif |
| 14 | `ppt_generator` | Indra | 🟢 Aktif |
| 15 | `pdf_formatter` | Restu | 🟢 Aktif |
| 16 | `programmer` | Aghni | 🟢 Aktif |
| 17 | `database_querier` | Adit | 🟢 Aktif |
| 18 | `supervisor` | Sarah | 🟢 Aktif |
| 19 | `essay_writer` | Reyhan | 🟢 Aktif |
| 20 | `prompt_generator` | Hilmi | 🟢 Aktif |
| 21 | `qa_bug_hunter` |  Iki | 🟢 Aktif |
| 22 | `kesimpulan_saran` | Farid | 🟢 Aktif |

Cek kontrak agen kapan saja:

```bash
task agents:check
task agents:check:md
```

---

## Cara menjalankan

### Prasyarat

- Docker
- [Go Task](https://taskfile.dev/) (opsional, tapi dipakai di bawah)
- Go 1.25+ & Node 20+ (untuk mode lokal tanpa compose penuh)

### 1) Setup env

```bash
cd app/
cp .env.example .env
```

Isi minimal di `.env`:

- `OPENCODE_GO_API_KEY` / `JWT_SECRET`
- URL agen (`AGENT_*_URL`) sesuai registry
- Opsional: `BRAVE_SEARCH_API_KEY`, `WEB_SEARCH_PROVIDER=wikipedia`

### 2A) Full stack via Docker Compose (disarankan)

```bash
task compose:up
# setara:
# docker compose -f deploy/compose/docker-compose.yml --env-file .env up -d --build
```

Stack yang naik:

| Service | Container | Port default |
| :--- | :--- | :--- |
| MongoDB | `joki-mongodb` | `27017` |
| Orchestrator | `joki-orchestrator` | `8080` |
| Web (Nginx) | `joki-web` | `3000` → UI |

Buka UI: [http://localhost:3000](http://localhost:3000)  
Health API: [http://localhost:8080/health](http://localhost:8080/health)

Stop:

```bash
task compose:down
```

Compose file: [`deploy/compose/docker-compose.yml`](deploy/compose/docker-compose.yml) (include infra + services + web). Nginx mem-proxy `/api`, `/ws`, `/health` ke orchestrator.

### 2B) Dev lokal (hot reload)

```bash
task init          # npm install + MongoDB container
task dev           # orchestrator :8080 + Vite :5173
```

UI: [http://localhost:5173](http://localhost:5173) (Vite proxy ke backend).

### Web search (sebelum scrape)

```bash
WEB_SEARCH_PROVIDER=wikipedia   # default
# WEB_SEARCH_PROVIDER=brave
# BRAVE_SEARCH_API_KEY=...
WEB_SEARCH_MAX_RESULTS=3
```

---

## Cara pakai aplikasi (step by step)

1. **Buka UI** — Compose: `http://localhost:3000` · Dev: `http://localhost:5173`.
2. **Daftar / masuk** — buat akun (username + password), lalu login.
3. **Chat baru** — klik **New chat** di sidebar (atau mulai ketik di chat kosong).
4. **Kirim permintaan** — dua cara:
   - **Bahasa natural**, contoh: *"Cari materi HTML dasar, ringkas, lalu buat PPT Bahasa Indonesia."*
   - **Template chip** / ketik `/template …`, contoh:
     ```text
     /template joki_presentasi
     Buat presentasi step-by-step belajar HTML, CSS, dan JS dasar.
     ```
     Preview pipeline muncul sebelum kirim.
5. **Tunggu rencana** — status jadi `awaiting_approval`; timeline menampilkan urutan agen.
6. **Gas atau Batal** — **Gas** menjalankan pipeline; **Batal** membatalkan task.
7. **Pantau progress** — WebSocket update tiap step (scrape, summarize, PPT, dll.). Bisa **Stop** di tengah jalan.
8. **Ambil hasil** — bila selesai, unduh file (PPTX/PDF/dll.) dari link di chat, atau baca teks hasil agen.
9. **Riwayat** — buka percakapan lama dari sidebar; kuota chat lifetime terlihat di header (`Sisa chat: x/y`).

### Tips prompt

- Sebutkan **output** yang diinginkan (PPT, PDF, kode, review).
- Untuk scrape, sebutkan **topik** jelas; orkestrator yang resolve URL.
- Hindari minta agen yang tidak relevan (mis. *"jangan pakai outliner"*) bila pipeline template sudah cukup.

### Alur singkat (diagram)

```text
User prompt → Router LLM → Pipeline plan → [Gas]
    → Agent 1 → Agent 2 → … → Result (file/teks) → Chat
```

---

## Struktur penting

```text
app/
├── service/orchestrator/   # API + pipeline runner
├── shared/agents/          # registry.json (SSOT)
├── web/                    # React UI (Bananacademic)
├── deploy/compose/         # docker-compose.*
├── deploy/docker/          # Dockerfile + nginx.conf
└── tools/agentcheck/       # cek kontrak agen
```

E2E multi-agen (rencana skenario): [`../docs/e2e_testing_plan.md`](../docs/e2e_testing_plan.md).
