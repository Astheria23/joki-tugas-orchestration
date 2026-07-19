import { useState } from 'react';
import { Check, Copy, Download, FileText } from 'lucide-react';

export function TaskResult({ result }: { result: string }) {
  const [copied, setCopied] = useState(false);
  const isUrl = /^https?:\/\//i.test(result);

  const copy = async () => {
    await navigator.clipboard.writeText(result);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  if (isUrl) {
    return (
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-2xl bg-white border border-ink/8 p-4">
        <div className="flex items-start gap-3 min-w-0">
          <div className="rounded-xl bg-banana/20 p-2.5 text-ink shrink-0">
            <FileText className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-ink">File siap diunduh</p>
            <p className="text-xs text-ink/45 font-mono truncate mt-0.5">{result}</p>
          </div>
        </div>
        <a
          href={result}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center justify-center gap-2 rounded-full bg-ink text-white px-4 py-2.5 text-sm font-semibold hover:bg-ink/90 transition-colors"
        >
          <Download className="h-4 w-4" />
          Unduh
        </a>
      </div>
    );
  }

  return (
    <div className="relative">
      <pre className="rounded-2xl bg-white border border-ink/8 p-4 text-sm text-ink whitespace-pre-wrap max-h-56 overflow-y-auto leading-relaxed font-mono">
        {result}
      </pre>
      <button
        type="button"
        onClick={() => void copy()}
        className="absolute top-3 right-3 rounded-lg bg-mist hover:bg-banana/30 border border-ink/8 p-2 text-ink transition-colors"
        title="Salin"
      >
        {copied ? <Check className="h-4 w-4 text-leaf" /> : <Copy className="h-4 w-4" />}
      </button>
    </div>
  );
}
