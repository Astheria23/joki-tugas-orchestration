export interface TaskHistoryItem {
  step: number;
  agentKey: string;
  status: string;
  input?: unknown;
  output?: unknown;
  error?: string;
  timestamp: string;
}

export interface Task {
  id: string;
  userId?: string;
  conversationId?: string;
  prompt: string;
  pipeline: string[];
  currentStep: number;
  status: 'pending' | 'running' | 'completed' | 'failed' | string;
  result?: string;
  history: TaskHistoryItem[];
  createdAt: string;
  updatedAt: string;
}

export interface User {
  id: string;
  username: string;
  createdAt?: string;
  chatUsed?: number;
}

export interface Me {
  id: string;
  username: string;
  chatUsed: number;
  chatLimit: number;
}

export interface Conversation {
  id: string;
  userId: string;
  title: string;
  createdAt: string;
  updatedAt: string;
}

export interface ChatMessage {
  id: string;
  conversationId: string;
  userId: string;
  role: 'user' | 'assistant' | 'system' | string;
  content: string;
  kind?: string;
  taskId?: string;
  pipeline?: string[];
  approvalStatus?: 'awaiting' | 'approved' | 'cancelled' | string;
  createdAt: string;
}

export interface ConversationDetail {
  conversation: Conversation;
  messages: ChatMessage[];
}

export interface SendMessageResult {
  userMessage: ChatMessage;
  assistantMessage: ChatMessage;
  chatUsed: number;
  chatLimit: number;
  taskId?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code?: number;
    message: string;
  };
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ChatWSPayload {
  type: string;
  message: ChatMessage;
}
