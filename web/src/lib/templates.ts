import { agentLabel } from './agents';

export type TaskTemplate = {
  id: string;
  title: string;
  hint: string;
  pipeline: string[];
  draft: string;
};

/** Keep in sync with router.go /template cases */
export const TASK_TEMPLATES: TaskTemplate[] = [
  {
    id: 'joki_makalah',
    title: 'Joki Makalah',
    hint: 'Paste referensi/topik makalah',
    pipeline: [
      'web_scraper',
      'data_mining',
      'literature_reviewer',
      'essay_writer',
      'citation_reference',
      'typo_checker',
      'pdf_formatter',
    ],
    draft: '/template joki_makalah\n[Paste referensi/topik makalah di sini]',
  },
  {
    id: 'joki_koding',
    title: 'Joki Koding',
    hint: 'Jelaskan tugas koding/error',
    pipeline: ['prompt_generator', 'programmer', 'qa_bug_hunter', 'supervisor'],
    draft: '/template joki_koding\n[Jelaskan tugas koding/error di sini]',
  },
  {
    id: 'joki_presentasi',
    title: 'PPT Generator',
    hint: 'Paste bahan materi PPT',
    pipeline: ['web_scraper', 'summarizer', 'outliner', 'ppt_generator'],
    draft: '/template joki_presentasi\n[Paste bahan materi PPT di sini]',
  },
  {
    id: 'review_tugas',
    title: 'Review Tugas',
    hint: 'Paste jawaban/teks tugas',
    pipeline: ['typo_checker', 'fact_checker', 'supervisor', 'kesimpulan_saran'],
    draft: '/template review_tugas\n[Paste jawaban/teks tugas kamu di sini]',
  },
  {
    id: 'analisis_data',
    title: 'Analisis Data',
    hint: 'Jelaskan data / pertanyaan analisis',
    pipeline: ['data_mining', 'database_querier', 'diagram_builder'],
    draft: '/template analisis_data\n[Jelaskan data atau pertanyaan analisis di sini]',
  },
];

export function parseTemplateFromInput(text: string): TaskTemplate | null {
  const trimmed = text.trim();
  const lower = trimmed.toLowerCase();
  if (!lower.startsWith('/template ')) return null;
  const rest = trimmed.slice('/template '.length);
  const id = rest.split(/\s|\n/)[0]?.trim().toLowerCase() ?? '';
  return TASK_TEMPLATES.find((t) => t.id === id) ?? null;
}

export function pipelinePreviewLabel(pipeline: string[]): string {
  return pipeline.map((k) => agentLabel(k)).join(' → ');
}
