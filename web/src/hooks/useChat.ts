import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import * as api from '../lib/api';
import type { ChatMessage, Conversation, Me } from '../lib/types';

function isTerminalKind(kind?: string) {
  return kind === 'task_result' || kind === 'task_error' || kind === 'task_cancelled';
}

export function useChat() {
  const [me, setMe] = useState<Me | null>(null);
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loadingList, setLoadingList] = useState(false);
  const [sending, setSending] = useState(false);
  const [thinking, setThinking] = useState(false);
  const [decidingTaskId, setDecidingTaskId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const applyApproval = useCallback((taskId: string, approvalStatus: string) => {
    setMessages((prev) =>
      prev.map((m) =>
        m.kind === 'task_pipeline' && m.taskId === taskId ? { ...m, approvalStatus } : m,
      ),
    );
  }, []);

  const refreshMe = useCallback(async () => {
    const profile = await api.getMe();
    setMe(profile);
    return profile;
  }, []);

  const refreshList = useCallback(async () => {
    setLoadingList(true);
    try {
      const list = await api.listConversations();
      setConversations(list);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal memuat chat');
    } finally {
      setLoadingList(false);
    }
  }, []);

  useEffect(() => {
    void refreshMe();
    void refreshList();
  }, [refreshMe, refreshList]);

  const connectWs = useCallback(
    (conversationId: string) => {
      wsRef.current?.close();
      const ws = new WebSocket(api.conversationWsUrl(conversationId));
      wsRef.current = ws;

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data) as { type: string; message: ChatMessage };
          if (payload.type === 'approval' && payload.message?.taskId && payload.message.approvalStatus) {
            applyApproval(payload.message.taskId, payload.message.approvalStatus);
            if (payload.message.approvalStatus === 'cancelled') {
              setThinking(false);
            }
            return;
          }
          if (payload?.message?.id) {
            setMessages((prev) => {
              if (prev.some((m) => m.id === payload.message.id)) return prev;
              return [...prev, payload.message];
            });
            const kind = payload.message.kind;
            if (isTerminalKind(kind) || (kind === 'task_pipeline' && payload.message.approvalStatus === 'awaiting')) {
              setThinking(false);
            }
          }
        } catch {
          // ignore
        }
      };
    },
    [applyApproval],
  );

  useEffect(() => {
    return () => wsRef.current?.close();
  }, []);

  const openConversation = useCallback(
    async (id: string) => {
      setError(null);
      setThinking(false);
      setActiveId(id);
      try {
        const detail = await api.getConversation(id);
        setMessages(detail.messages ?? []);
        connectWs(id);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Gagal membuka chat');
      }
    },
    [connectWs],
  );

  const startNewChat = useCallback(async () => {
    setError(null);
    setThinking(false);
    try {
      const conv = await api.createConversation();
      setConversations((prev) => [conv, ...prev.filter((c) => c.id !== conv.id)]);
      setActiveId(conv.id);
      setMessages([]);
      connectWs(conv.id);
      return conv;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal buat chat');
      return null;
    }
  }, [connectWs]);

  const send = useCallback(
    async (text: string) => {
      setError(null);
      let convId = activeId;
      if (!convId) {
        const created = await startNewChat();
        if (!created) return;
        convId = created.id;
      }

      const tempId = `temp-${Date.now()}`;
      const optimistic: ChatMessage = {
        id: tempId,
        conversationId: convId,
        userId: me?.id ?? '',
        role: 'user',
        content: text,
        kind: 'chat',
        createdAt: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, optimistic]);
      setSending(true);
      setThinking(true);

      try {
        const result = await api.sendMessage(convId, text);
        setMessages((prev) => {
          const withoutTemp = prev.filter((m) => m.id !== tempId);
          const next = [...withoutTemp];
          if (!next.some((m) => m.id === result.userMessage.id)) next.push(result.userMessage);
          if (!next.some((m) => m.id === result.assistantMessage.id)) next.push(result.assistantMessage);
          return next;
        });
        setMe((prev) =>
          prev
            ? { ...prev, chatUsed: result.chatUsed, chatLimit: result.chatLimit }
            : { id: '', username: '', chatUsed: result.chatUsed, chatLimit: result.chatLimit },
        );
        setConversations((prev) => {
          const updated = prev.map((c) =>
            c.id === convId
              ? {
                  ...c,
                  title: c.title === 'Chat baru' ? text.slice(0, 60) : c.title,
                  updatedAt: new Date().toISOString(),
                }
              : c,
          );
          return updated.sort((a, b) => +new Date(b.updatedAt) - +new Date(a.updatedAt));
        });
        // Task ack → keep light thinking until pipeline / terminal via WS
        if (result.assistantMessage.kind === 'task_ack' || result.taskId) {
          setThinking(true);
        } else {
          setThinking(false);
        }
      } catch (err) {
        setMessages((prev) => prev.filter((m) => m.id !== tempId));
        setError(err instanceof Error ? err.message : 'Gagal mengirim');
        setThinking(false);
      } finally {
        setSending(false);
      }
    },
    [activeId, startNewChat, me?.id],
  );

  const decideTask = useCallback(
    async (taskId: string, action: 'approve' | 'cancel') => {
      setError(null);
      setDecidingTaskId(taskId);
      try {
        await api.decideTask(taskId, action);
        applyApproval(taskId, action === 'approve' ? 'approved' : 'cancelled');
        setThinking(false);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Gagal memproses keputusan');
      } finally {
        setDecidingTaskId(null);
      }
    },
    [applyApproval],
  );

  const taskBusy = useMemo(() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const m = messages[i];
      const k = m.kind;
      if (isTerminalKind(k)) return false;
      if (k === 'task_pipeline' && m.approvalStatus === 'awaiting') return false;
      if (k === 'task_pipeline' && m.approvalStatus === 'cancelled') return false;
      if (k === 'task_ack' || k === 'task_progress') return true;
      if (k === 'task_pipeline' && m.approvalStatus === 'approved') return true;
    }
    return false;
  }, [messages]);

  const quotaLeft = me ? Math.max(0, me.chatLimit - me.chatUsed) : null;
  const quotaExhausted = me ? me.chatUsed >= me.chatLimit : false;

  return {
    me,
    conversations,
    activeId,
    messages,
    loadingList,
    sending,
    thinking: thinking || sending,
    decidingTaskId,
    taskBusy,
    error,
    quotaLeft,
    quotaExhausted,
    openConversation,
    startNewChat,
    send,
    decideTask,
    clearError: () => setError(null),
  };
}
