import { Send } from 'lucide-react';

interface ChatInputProps {
  value: string;
  onChange: (val: string) => void;
  disabled?: boolean;
  sending?: boolean;
  placeholder?: string;
  onSend: () => Promise<void> | void;
}

export function ChatInput({ value, onChange, disabled, sending, placeholder, onSend }: ChatInputProps) {
  const submit = async () => {
    const trimmed = value.trim();
    if (!trimmed || disabled || sending) return;
    await onSend();
  };

  return (
    <div className="border-t border-ink/8 bg-white/90 backdrop-blur p-4">
      <div className="max-w-3xl mx-auto flex items-end gap-2">
        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={1}
          disabled={disabled || sending}
          placeholder={placeholder ?? 'Tulis pesan…'}
          className="flex-1 resize-none rounded-2xl border border-ink/10 bg-canvas px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-banana disabled:opacity-50 max-h-32"
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
              e.preventDefault();
              void submit();
            }
          }}
        />
        <button
          type="button"
          disabled={disabled || sending || !value.trim()}
          onClick={() => void submit()}
          className="shrink-0 rounded-full bg-banana text-ink p-3 hover:bg-banana-deep disabled:opacity-40 transition-colors"
          aria-label="Kirim"
        >
          <Send className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
}
