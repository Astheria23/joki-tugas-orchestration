import { useEffect, useMemo, useRef, useState } from 'react';
import { Check, Copy, Download, Loader2 } from 'lucide-react';
import type { ChatMessage } from '../../lib/types';
import { AgentTimeline } from './AgentTimeline';

function ResultBlock({ content }: { content: string }) {
  const [copied, setCopied] = useState(false);
  const url = content.trim();
  const isUrl = /^https?:\/\//i.test(url);
  const lower = url.toLowerCase();
  const isPdf = isUrl && (lower.includes('.pdf') || lower.includes('/pdf') || lower.endsWith('pdf'));
  const isPpt =
    isUrl &&
    (lower.includes('.ppt') ||
      lower.includes('.pptx') ||
      lower.includes('pptx') ||
      lower.includes('presentation'));
  const isImage = isUrl && /\.(png|jpe?g|gif|webp|svg)(\?|$)/i.test(url);

  if (isUrl) {
    const canEmbedPdf = isPdf;
    // Office Online embed often blank for some storage hosts (e.g. Supabase) — prefer download.
    const canEmbedPpt = false;

    return (
      <div className="mt-2 space-y-2">
        {canEmbedPdf && (
          <iframe
            title="Preview PDF"
            src={url}
            className="w-full h-72 rounded-xl border border-ink/10 bg-white"
          />
        )}
        {canEmbedPpt && isPpt && (
          <iframe
            title="Preview PPT"
            src={`https://view.officeapps.live.com/op/embed.aspx?src=${encodeURIComponent(url)}`}
            className="w-full h-72 rounded-xl border border-ink/10 bg-white"
          />
        )}
        {isImage && (
          <a href={url} target="_blank" rel="noreferrer" className="block">
            <img
              src={url}
              alt="Hasil"
              className="max-h-72 rounded-xl border border-ink/10 object-contain bg-white"
            />
          </a>
        )}
        {(isPpt || isPdf) && (
          <p className="text-xs text-ink/50">
            {isPpt ? 'File PPTX siap diunduh.' : 'File PDF siap diunduh / dipreview.'}
          </p>
        )}
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-2 rounded-full bg-ink text-white px-3 py-1.5 text-xs font-semibold"
        >
          <Download className="h-3.5 w-3.5" />
          Unduh / buka hasil
        </a>
        {!isPdf && !isPpt && !isImage && (
          <p className="text-[11px] text-ink/40 break-all">{url}</p>
        )}
      </div>
    );
  }

  return (
    <div className="relative mt-2">
      <pre className="whitespace-pre-wrap text-sm leading-relaxed font-mono bg-white/60 rounded-xl p-3 max-h-64 overflow-y-auto border border-ink/8">
        {content}
      </pre>
      <button
        type="button"
        className="absolute top-2 right-2 p-1.5 rounded-lg bg-mist hover:bg-banana/40 border border-ink/8"
        onClick={async () => {
          await navigator.clipboard.writeText(content);
          setCopied(true);
          setTimeout(() => setCopied(false), 1500);
        }}
      >
        {copied ? <Check className="h-3.5 w-3.5 text-leaf" /> : <Copy className="h-3.5 w-3.5" />}
      </button>
    </div>
  );
}

function ThinkingBubble() {
  return (
    <div className="flex justify-start">
      <div className="rounded-3xl rounded-bl-md bg-white border border-ink/8 px-4 py-3 text-sm text-ink/60 inline-flex items-center gap-2">
        <Loader2 className="h-4 w-4 animate-spin text-banana-deep" />
        <span className="thinking-dots">Sedang mikir</span>
      </div>
    </div>
  );
}

interface MessageListProps {
  messages: ChatMessage[];
  thinking?: boolean;
  decidingTaskId?: string | null;
  onDecide?: (taskId: string, action: 'approve' | 'cancel') => void;
}

export function MessageList({ messages, thinking, decidingTaskId, onDecide }: MessageListProps) {
  const endRef = useRef<HTMLDivElement | null>(null);

  const liveByTask = useMemo(() => {
    const map = new Map<
      string,
      {
        currentStep: number;
        status: string;
        stopped: boolean;
        latestProgressIdx: number;
        skippedSteps: number[];
      }
    >();

    messages.forEach((m, idx) => {
      if (!m.taskId) return;
      const prev = map.get(m.taskId) ?? {
        currentStep: 0,
        status: 'pending',
        stopped: false,
        latestProgressIdx: -1,
        skippedSteps: [],
      };

      if (m.kind === 'task_pipeline') {
        if (m.approvalStatus === 'awaiting') prev.status = 'awaiting';
        else if (m.approvalStatus === 'cancelled') {
          prev.status = 'cancelled';
          prev.stopped = true;
        } else if (m.approvalStatus === 'approved') prev.status = 'running';
      } else if (m.kind === 'task_result') {
        prev.status = 'completed';
        prev.stopped = true;
        prev.currentStep = Number.MAX_SAFE_INTEGER;
      } else if (m.kind === 'task_error' || m.kind === 'task_cancelled') {
        prev.status = m.kind === 'task_cancelled' ? 'cancelled' : 'failed';
        prev.stopped = true;
      } else if (m.kind === 'task_progress') {
        const match = m.content.match(/\((\d+)\s*\/\s*(\d+)\)/);
        if (match) prev.currentStep = Number(match[1]);
        if (m.content.includes('dilewati') && match) {
          const step = Number(match[1]);
          if (!prev.skippedSteps.includes(step)) prev.skippedSteps.push(step);
        }
        if (!prev.stopped) prev.status = 'running';
        prev.latestProgressIdx = idx;
      }

      map.set(m.taskId, prev);
    });

    return map;
  }, [messages]);

  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, thinking]);

  if (messages.length === 0 && !thinking) {
    return (
      <div className="flex-1 flex items-center justify-center px-6 text-center">
        <div>
          <p className="font-display text-2xl font-bold text-ink/80">Mau dibantu apa hari ini?</p>
          <p className="text-sm text-ink/45 mt-2 max-w-sm mx-auto leading-relaxed">
            Chat biasa boleh. Kalau mau dikerjakan (ringkas, slide, terjemah), tulis aja — aku prosesin.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto px-4 sm:px-6 py-6 space-y-4">
      {messages.map((m, idx) => {
        const isUser = m.role === 'user';
        const isPipeline = m.kind === 'task_pipeline';
        const isProgress = m.kind === 'task_progress' || (m.role === 'system' && !isPipeline);
        const isResult = m.kind === 'task_result';
        const isError = m.kind === 'task_error';
        const live = m.taskId ? liveByTask.get(m.taskId) : undefined;

        if (isUser) {
          return (
            <div key={m.id} className="flex justify-end">
              <div className="max-w-[85%] sm:max-w-[70%] rounded-3xl rounded-br-md bg-navy text-white px-4 py-3 text-sm leading-relaxed whitespace-pre-wrap">
                {m.content}
              </div>
            </div>
          );
        }

        if (isPipeline) {
          const awaiting = m.approvalStatus === 'awaiting';
          const deciding = decidingTaskId === m.taskId;
          return (
            <div key={m.id} className="flex justify-start">
              <div className="max-w-[85%] sm:max-w-[75%] rounded-3xl rounded-bl-md bg-white border border-ink/8 px-4 py-3 text-sm text-ink">
                <p className="whitespace-pre-wrap">{m.content}</p>
                <AgentTimeline
                  agents={m.pipeline ?? []}
                  currentStep={
                    live?.status === 'completed'
                      ? (m.pipeline?.length ?? 0)
                      : live?.currentStep ?? 0
                  }
                  status={live?.status ?? (awaiting ? 'awaiting' : undefined)}
                  skippedSteps={live?.skippedSteps ?? []}
                />
                {awaiting && m.taskId && onDecide && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    <button
                      type="button"
                      disabled={deciding}
                      onClick={() => onDecide(m.taskId!, 'approve')}
                      className="rounded-full bg-navy text-white px-4 py-2 text-xs font-semibold hover:bg-navy/90 disabled:opacity-50"
                    >
                      {deciding ? 'Sebentar…' : 'Gas'}
                    </button>
                    <button
                      type="button"
                      disabled={deciding}
                      onClick={() => onDecide(m.taskId!, 'cancel')}
                      className="rounded-full border border-ink/15 bg-white px-4 py-2 text-xs font-semibold text-ink/70 hover:bg-mist disabled:opacity-50"
                    >
                      Batal
                    </button>
                  </div>
                )}
                {live?.status === 'running' && m.taskId && onDecide && (
                  <div className="mt-3 flex flex-wrap gap-2">
                    <button
                      type="button"
                      disabled={deciding}
                      onClick={() => onDecide(m.taskId!, 'cancel')}
                      className="rounded-full border border-rose-300 bg-rose-50 text-rose-700 px-4 py-2 text-xs font-semibold hover:bg-rose-100 disabled:opacity-50 flex items-center gap-1"
                    >
                      🛑 Stop (Batal)
                    </button>
                  </div>
                )}
                {m.approvalStatus === 'cancelled' && (
                  <p className="mt-2 text-xs text-ink/45">Rencana ini dibatalin.</p>
                )}
              </div>
            </div>
          );
        }

        if (isProgress) {
          const stopped = Boolean(live?.stopped);
          const isLatest =
            m.taskId && live ? live.latestProgressIdx === idx && !stopped : !stopped;
          return (
            <div key={m.id} className="flex justify-center">
              <p className="text-xs text-ink/45 bg-mist/80 rounded-full px-3 py-1.5 inline-flex items-center gap-1.5">
                {isLatest && <Loader2 className="h-3 w-3 animate-spin opacity-60" />}
                {m.content}
              </p>
            </div>
          );
        }

        return (
          <div key={m.id} className="flex justify-start">
            <div
              className={`max-w-[85%] sm:max-w-[75%] rounded-3xl rounded-bl-md border px-4 py-3 text-sm leading-relaxed ${
                isError
                  ? 'bg-rose-50 border-rose-200 text-rose-900'
                  : 'bg-white border-ink/8 text-ink'
              }`}
            >
              {!isResult && <p className="whitespace-pre-wrap">{m.content}</p>}
              {isResult && (
                <>
                  <p className="font-semibold text-ink mb-1">Hasil siap</p>
                  <ResultBlock content={m.content} />
                </>
              )}
            </div>
          </div>
        );
      })}
      {thinking && <ThinkingBubble />}
      <div ref={endRef} />
    </div>
  );
}
