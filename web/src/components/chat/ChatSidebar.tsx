import { Plus } from 'lucide-react';
import type { Conversation } from '../../lib/types';

interface ChatSidebarProps {
  conversations: Conversation[];
  activeId: string | null;
  loading?: boolean;
  quotaLabel: string;
  onNew: () => void;
  onSelect: (id: string) => void;
}

export function ChatSidebar({
  conversations,
  activeId,
  loading,
  quotaLabel,
  onNew,
  onSelect,
}: ChatSidebarProps) {
  return (
    <aside className="flex flex-col h-full border-r border-ink/8 bg-white/80">
      <div className="p-4 border-b border-ink/8 space-y-3">
        <button
          type="button"
          onClick={onNew}
          className="w-full inline-flex items-center justify-center gap-2 rounded-full bg-banana text-ink font-semibold py-2.5 text-sm hover:bg-banana-deep transition-colors"
        >
          <Plus className="h-4 w-4" />
          Chat baru
        </button>
        <p className="text-center text-[11px] font-medium text-ink/45">{quotaLabel}</p>
      </div>

      <div className="flex-1 overflow-y-auto p-2 space-y-1">
        {loading && conversations.length === 0 && (
          <p className="text-xs text-ink/40 px-3 py-4">Memuat…</p>
        )}
        {!loading && conversations.length === 0 && (
          <p className="text-xs text-ink/40 px-3 py-4 leading-relaxed">
            Belum ada percakapan. Mulai chat baru di atas.
          </p>
        )}
        {conversations.map((c) => {
          const active = c.id === activeId;
          return (
            <button
              key={c.id}
              type="button"
              onClick={() => onSelect(c.id)}
              className={`w-full text-left rounded-2xl px-3 py-2.5 text-sm transition-colors truncate ${
                active ? 'bg-navy text-white' : 'hover:bg-mist text-ink/80'
              }`}
              title={c.title}
            >
              {c.title || 'Chat baru'}
            </button>
          );
        })}
      </div>
    </aside>
  );
}
