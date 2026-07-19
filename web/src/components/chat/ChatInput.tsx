import { useState } from 'react';
import { Send } from 'lucide-react';

interface ChatInputProps {
  disabled?: boolean;
  sending?: boolean;
  placeholder?: string;
  onSend: (text: string) => Promise<void> | void;
}

export function ChatInput({ disabled, sending, placeholder, onSend }: ChatInputProps) {
  const [text, setText] = useState('');

  const submit = async () => {
    const value = text.trim();
    if (!value || disabled || sending) return;
    setText('');
    await onSend(value);
  };

  return (
    <div className="border-t border-ink/8 bg-white/90 backdrop-blur p-4">
      <div className="max-w-3xl mx-auto flex items-end gap-2">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
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
          disabled={disabled || sending || !text.trim()}
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
