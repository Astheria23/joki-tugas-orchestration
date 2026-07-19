interface StatusBadgeProps {
  status: string;
}

const styles: Record<string, string> = {
  completed: 'bg-leaf/15 text-leaf border-leaf/30',
  failed: 'bg-rose-500/10 text-rose-700 border-rose-500/25',
  running: 'bg-banana/25 text-ink border-banana/50 animate-pulse',
  pending: 'bg-mist text-ink/60 border-ink/10',
  skipped: 'bg-amber-100 text-amber-800 border-amber-300/50',
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const key = status.toLowerCase();
  return (
    <span
      className={`inline-flex items-center px-2.5 py-1 text-[11px] font-semibold uppercase tracking-wider rounded-full border ${
        styles[key] ?? 'bg-mist text-ink/70 border-ink/10'
      }`}
    >
      {status}
    </span>
  );
}
