import { useEffect, useRef } from 'react';

export function LiveConsole({ logs, open }: { logs: string[]; open: boolean }) {
  const endRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (open) endRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs, open]);

  if (!open) return null;

  return (
    <div className="rounded-2xl bg-ink text-white/90 font-mono text-[11px] p-4 h-52 overflow-y-auto">
      {logs.length === 0 ? (
        <p className="text-white/40">Belum ada log.</p>
      ) : (
        <div className="space-y-1.5">
          {logs.map((log, i) => (
            <p key={`${i}-${log.slice(0, 24)}`} className="leading-relaxed break-all">
              <span className="text-banana/80 mr-2">›</span>
              {log}
            </p>
          ))}
          <div ref={endRef} />
        </div>
      )}
    </div>
  );
}
