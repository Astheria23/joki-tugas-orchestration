import { StatusBadge } from './StatusBadge';
import type { Task } from '../lib/types';

interface TaskHistoryProps {
  tasks: Task[];
  activeId?: string;
  loading?: boolean;
  onSelect: (id: string) => void;
}

function truncate(text: string, max = 64) {
  if (text.length <= max) return text;
  return `${text.slice(0, max)}…`;
}

function formatTime(iso: string) {
  try {
    return new Intl.DateTimeFormat('id-ID', {
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    }).format(new Date(iso));
  } catch {
    return '';
  }
}

export function TaskHistoryList({ tasks, activeId, loading, onSelect }: TaskHistoryProps) {
  if (loading && tasks.length === 0) {
    return <p className="text-sm text-ink/45 py-4">Memuat riwayat…</p>;
  }

  if (tasks.length === 0) {
    return (
      <p className="text-sm text-ink/45 py-4 leading-relaxed">
        Belum ada tugas. Tulis prompt di sebelah kiri untuk mulai.
      </p>
    );
  }

  return (
    <ul className="space-y-2 max-h-72 overflow-y-auto pr-1">
      {tasks.map((task) => {
        const active = task.id === activeId;
        return (
          <li key={task.id}>
            <button
              type="button"
              onClick={() => onSelect(task.id)}
              className={`w-full text-left rounded-2xl border px-3.5 py-3 transition-all ${
                active
                  ? 'bg-ink text-white border-ink shadow-lg'
                  : 'bg-white border-ink/8 hover:border-banana/50 hover:bg-banana/10 text-ink'
              }`}
            >
              <div className="flex items-start justify-between gap-2 mb-1.5">
                <p className={`text-sm font-medium leading-snug ${active ? 'text-white' : 'text-ink'}`}>
                  {truncate(task.prompt)}
                </p>
                <StatusBadge status={task.status} />
              </div>
              <p className={`text-[11px] ${active ? 'text-white/55' : 'text-ink/40'}`}>
                {formatTime(task.createdAt)}
                {task.pipeline?.length ? ` · ${task.pipeline.length} langkah` : ''}
              </p>
            </button>
          </li>
        );
      })}
    </ul>
  );
}
