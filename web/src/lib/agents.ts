/** Canonical agent registry — aligned with orchestrator AgentContracts */
export const AGENTS = [
  { key: 'web_scraper', label: 'Web Scraper', in: 'url', out: 'text', desc: 'Ambil konten dari URL target' },
  { key: 'data_mining', label: 'Data Mining', in: 'text | url', out: 'text', desc: 'Ekstrak pola dan insight dari teks' },
  { key: 'summarizer', label: 'Summarizer', in: 'text', out: 'text', desc: 'Ringkas laporan panjang jadi poin padat' },
  { key: 'outliner', label: 'Outliner', in: 'text', out: 'outline', desc: 'Susun kerangka struktur materi' },
  { key: 'translator', label: 'Translator', in: 'text', out: 'text', desc: 'Terjemahkan ke bahasa target' },
  { key: 'paraphrase', label: 'Paraphrase', in: 'text', out: 'text', desc: 'Parafrase agar lebih natural' },
  { key: 'typo_checker', label: 'Typo Checker', in: 'text', out: 'text', desc: 'Perbaiki ejaan dan tata bahasa' },
  { key: 'fact_checker', label: 'Fact Checker', in: 'text', out: 'text', desc: 'Verifikasi klaim terhadap sumber' },
  { key: 'literature_reviewer', label: 'Literature Reviewer', in: 'text', out: 'text', desc: 'Ringkas literatur akademik relevan' },
  { key: 'citation_reference', label: 'Citation', in: 'text', out: 'text', desc: 'Format sitasi APA / MLA / IEEE' },
  { key: 'qna_simulator', label: 'QnA Simulator', in: 'text', out: 'text', desc: 'Simulasi tanya-jawab dari materi' },
  { key: 'math_calculator', label: 'Math Solver', in: 'text', out: 'text', desc: 'Selesaikan soal matematika step-by-step' },
  { key: 'spatial_gis', label: 'Spatial GIS', in: 'text', out: 'text', desc: 'Proses data spasial / GIS' },
  { key: 'requirement_analyzer', label: 'Requirement Analyzer', in: 'text', out: 'text', desc: 'Analisis kebutuhan perangkat lunak' },
  { key: 'diagram_builder', label: 'Diagram Builder', in: 'text', out: 'file', desc: 'Bangun diagram dari deskripsi' },
  { key: 'ppt_generator', label: 'PPT Generator', in: 'text | outline', out: 'file', desc: 'Susun slide presentasi PPTX' },
  { key: 'pdf_formatter', label: 'PDF Formatter', in: 'text', out: 'file', desc: 'Format dokumen PDF siap kirim' },
  { key: 'programmer', label: 'Programmer', in: 'text', out: 'code', desc: 'Tulis kode dari spesifikasi' },
  { key: 'pr_reviewer', label: 'PR Reviewer', in: 'code', out: 'text', desc: 'Review kode dan saran perbaikan' },
  { key: 'database_querier', label: 'DB Querier', in: 'text', out: 'text', desc: 'Query dan jelaskan data' },
  { key: 'context_memory', label: 'Context Memory', in: 'text', out: 'text', desc: 'Simpan dan lanjutkan konteks' },
  { key: 'supervisor', label: 'Supervisor', in: 'text | code', out: 'text', desc: 'Validasi hasil akhir vs permintaan' },
] as const;

export function agentLabel(key: string): string {
  const found = AGENTS.find((a) => a.key === key.toLowerCase());
  return found?.label ?? key.replace(/_/g, ' ');
}
