import { useState } from 'react';
import { Loader2, Sparkles } from 'lucide-react';

const PRESETS = [
  'Scrape artikel wiki AI, ringkas isinya, terjemahkan ke Bahasa Indonesia, lalu buat file PPT',
  'Analisis requirement software, buat diagram UML, lalu format jadi PDF dokumentasi',
  'Parafrase teks akademik ini, cek typo, lalu buat sitasi APA',
];

interface PromptFormProps {
  loading: boolean;
  onSubmit: (prompt: string) => Promise<void>;
}

export function PromptForm({ loading, onSubmit }: PromptFormProps) {
  const [prompt, setPrompt] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!prompt.trim() || loading) return;
    await onSubmit(prompt.trim());
    setPrompt('');
  };

  return (
    <div className="space-y-4">
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-3">
        <label htmlFor="prompt" className="block text-sm font-semibold text-ink">
          Apa yang mau dikerjakan?
        </label>
        <textarea
          id="prompt"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          rows={4}
          disabled={loading}
          placeholder="Contoh: Ringkas materi kuliah ini dan buat slide presentasi…"
          className="w-full rounded-2xl border border-ink/10 bg-white px-4 py-3 text-sm text-ink placeholder:text-ink/35 focus:outline-none focus:ring-2 focus:ring-banana focus:border-transparent resize-none transition"
        />
        <button
          type="submit"
          disabled={loading || !prompt.trim()}
          className="w-full inline-flex items-center justify-center gap-2 rounded-full bg-banana text-ink font-semibold py-3 text-sm hover:bg-banana-deep disabled:opacity-40 disabled:cursor-not-allowed transition-colors shadow-[0_8px_24px_rgba(245,197,24,0.35)]"
        >
          {loading ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Sedang dikerjakan…
            </>
          ) : (
            <>
              <Sparkles className="h-4 w-4" />
              Kerjakan sekarang
            </>
          )}
        </button>
      </form>

      <div>
        <p className="text-xs font-medium text-ink/45 mb-2 uppercase tracking-wider">Coba template</p>
        <div className="flex flex-col gap-2">
          {PRESETS.map((preset) => (
            <button
              key={preset}
              type="button"
              disabled={loading}
              onClick={() => setPrompt(preset)}
              className="text-left text-xs leading-relaxed rounded-xl border border-ink/8 bg-white/70 hover:bg-banana/15 hover:border-banana/40 px-3 py-2.5 text-ink/70 transition-colors disabled:opacity-50"
            >
              {preset}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
