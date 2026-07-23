import { LogOut, Menu, X } from 'lucide-react';
import { useState } from 'react';
import { BrandLogo } from '../components/BrandLogo';
import { ChatInput } from '../components/chat/ChatInput';
import { ChatSidebar } from '../components/chat/ChatSidebar';
import { MessageList } from '../components/chat/MessageList';
import { useAuth } from '../hooks/useAuth';
import { useChat } from '../hooks/useChat';

export function ChatPage() {
  const { user, logout } = useAuth();
  const {
    me,
    conversations,
    activeId,
    messages,
    loadingList,
    sending,
    thinking,
    decidingTaskId,
    error,
    quotaLeft,
    quotaExhausted,
    openConversation,
    startNewChat,
    send,
    decideTask,
  } = useChat();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [inputValue, setInputValue] = useState('');

  const quotaLabel =
    me != null ? `Sisa chat: ${quotaLeft}/${me.chatLimit}` : 'Sisa chat: …';

  return (
    <div className="h-svh bg-canvas text-ink flex flex-col overflow-hidden">
      <header className="shrink-0 border-b border-ink/8 bg-white/90 backdrop-blur px-4 py-3 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="lg:hidden rounded-xl border border-ink/10 p-2"
            onClick={() => setSidebarOpen(true)}
            aria-label="Buka menu"
          >
            <Menu className="h-4 w-4" />
          </button>
          <BrandLogo size="sm" variant="wordmark" to="/app" />
        </div>
        <div className="flex items-center gap-3">
          <span className="hidden sm:inline text-xs font-medium text-ink/45">{quotaLabel}</span>
          <span className="text-sm text-ink/55 hidden md:inline truncate max-w-[120px]">
            {user?.username ?? me?.username}
          </span>
          <button
            type="button"
            onClick={logout}
            className="inline-flex items-center gap-1.5 rounded-full border border-ink/10 bg-white px-3 py-1.5 text-xs font-semibold hover:bg-mist"
          >
            <LogOut className="h-3.5 w-3.5" />
            Keluar
          </button>
        </div>
      </header>

      <div className="flex-1 flex min-h-0 relative">
        {/* Desktop sidebar */}
        <div className="hidden lg:block w-72 shrink-0 h-full">
          <ChatSidebar
            conversations={conversations}
            activeId={activeId}
            loading={loadingList}
            quotaLabel={quotaLabel}
            onNew={() => void startNewChat()}
            onSelect={(id) => void openConversation(id)}
          />
        </div>

        {/* Mobile drawer */}
        {sidebarOpen && (
          <div className="lg:hidden absolute inset-0 z-30 flex">
            <div className="w-72 h-full bg-white shadow-xl relative z-10">
              <button
                type="button"
                className="absolute top-3 right-3 p-2"
                onClick={() => setSidebarOpen(false)}
                aria-label="Tutup"
              >
                <X className="h-4 w-4" />
              </button>
              <ChatSidebar
                conversations={conversations}
                activeId={activeId}
                loading={loadingList}
                quotaLabel={quotaLabel}
                onNew={() => {
                  void startNewChat();
                  setSidebarOpen(false);
                }}
                onSelect={(id) => {
                  void openConversation(id);
                  setSidebarOpen(false);
                }}
              />
            </div>
            <button
              type="button"
              className="flex-1 bg-ink/30"
              aria-label="Tutup overlay"
              onClick={() => setSidebarOpen(false)}
            />
          </div>
        )}

        <section className="flex-1 flex flex-col min-w-0 min-h-0">
          {error && (
            <div className="mx-4 mt-3 rounded-2xl border border-rose-300 bg-rose-50 text-rose-800 px-4 py-2.5 text-sm">
              {error}
            </div>
          )}

          <MessageList
            messages={messages}
            thinking={thinking}
            decidingTaskId={decidingTaskId}
            onDecide={(taskId, action) => void decideTask(taskId, action)}
          />

          {messages.length === 0 && !thinking && (
            <div className="flex flex-wrap gap-2 px-4 pb-2 justify-center">
              <button
                type="button"
                onClick={() => setInputValue("/template joki_makalah\n[Paste referensi/topik makalah di sini]")}
                className="text-xs bg-ink/5 hover:bg-ink/10 rounded-full px-3 py-1.5 transition-colors font-medium border border-ink/10 text-ink/70"
              >
                🎓 Joki Makalah
              </button>
              <button
                type="button"
                onClick={() => setInputValue("/template joki_koding\n[Jelaskan tugas koding/error di sini]")}
                className="text-xs bg-ink/5 hover:bg-ink/10 rounded-full px-3 py-1.5 transition-colors font-medium border border-ink/10 text-ink/70"
              >
                💻 Joki Koding
              </button>
              <button
                type="button"
                onClick={() => setInputValue("/template joki_presentasi\n[Paste bahan materi PPT di sini]")}
                className="text-xs bg-ink/5 hover:bg-ink/10 rounded-full px-3 py-1.5 transition-colors font-medium border border-ink/10 text-ink/70"
              >
                📊 PPT Generator
              </button>
              <button
                type="button"
                onClick={() => setInputValue("/template review_tugas\n[Paste jawaban/teks tugas kamu di sini]")}
                className="text-xs bg-ink/5 hover:bg-ink/10 rounded-full px-3 py-1.5 transition-colors font-medium border border-ink/10 text-ink/70"
              >
                🕵️‍♂️ Review Tugas
              </button>
            </div>
          )}

          <ChatInput
            value={inputValue}
            onChange={setInputValue}
            disabled={quotaExhausted}
            sending={sending}
            placeholder={
              quotaExhausted
                ? `Kuota chat habis (${me?.chatLimit ?? 0}/${me?.chatLimit ?? 0})`
                : thinking
                  ? 'Sedang mikir…'
                  : 'Tulis pesan… (Enter kirim, Shift+Enter baris baru)'
            }
            onSend={() => {
              void send(inputValue);
              setInputValue('');
            }}
          />
        </section>
      </div>
    </div>
  );
}
