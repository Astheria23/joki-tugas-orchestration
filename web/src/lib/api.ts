import type {
  ApiResponse,
  AuthResponse,
  Conversation,
  ConversationDetail,
  Me,
  SendMessageResult,
  User,
} from './types';

const TOKEN_KEY = 'bananacademic_token';
const USER_KEY = 'bananacademic_user';

export function getApiBase(): string {
  const envBase = import.meta.env.VITE_API_BASE as string | undefined;
  if (envBase) return envBase.replace(/\/$/, '');

  if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
    return 'http://localhost:8080';
  }

  return `${window.location.protocol}//${window.location.host}`;
}

export function getWsBase(): string {
  const envBase = import.meta.env.VITE_WS_BASE as string | undefined;
  if (envBase) return envBase.replace(/\/$/, '');

  const api = getApiBase();
  if (api.startsWith('https://')) return api.replace(/^https/, 'wss');
  if (api.startsWith('http://')) return api.replace(/^http/, 'ws');

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host =
    window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
      ? 'localhost:8080'
      : window.location.host;
  return `${protocol}//${host}`;
}

export function getStoredToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function getStoredUser(): User | null {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}

export function persistAuth(token: string, user: User) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getStoredToken();
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
    ...(options.headers ?? {}),
  };

  if (token) {
    (headers as Record<string, string>)['Authorization'] = `Bearer ${token}`;
  }

  const resp = await fetch(`${getApiBase()}${path}`, {
    ...options,
    headers,
  });

  let body: ApiResponse<T> | null = null;
  try {
    body = (await resp.json()) as ApiResponse<T>;
  } catch {
    // ignore empty body
  }

  if (!resp.ok || body?.success === false) {
    const message = body?.error?.message || `Request failed (${resp.status})`;
    throw new Error(message);
  }

  return body!.data as T;
}

export async function login(username: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function register(username: string, password: string): Promise<User> {
  return request<User>('/api/auth/register', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export async function getMe(): Promise<Me> {
  return request<Me>('/api/me');
}

export async function listConversations(): Promise<Conversation[]> {
  return (await request<Conversation[]>('/api/conversations')) ?? [];
}

export async function createConversation(title?: string): Promise<Conversation> {
  return request<Conversation>('/api/conversations', {
    method: 'POST',
    body: JSON.stringify({ title: title ?? '' }),
  });
}

export async function getConversation(id: string): Promise<ConversationDetail> {
  return request<ConversationDetail>(`/api/conversations/${id}`);
}

export async function sendMessage(conversationId: string, content: string): Promise<SendMessageResult> {
  return request<SendMessageResult>(`/api/conversations/${conversationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ content }),
  });
}

export async function decideTask(
  taskId: string,
  action: 'approve' | 'cancel',
): Promise<{ status: string; taskId: string }> {
  return request<{ status: string; taskId: string }>(`/api/tasks/${taskId}/decide`, {
    method: 'POST',
    body: JSON.stringify({ action }),
  });
}

export function conversationWsUrl(id: string): string {
  return `${getWsBase()}/ws/conversations/${id}`;
}
