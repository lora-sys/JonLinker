export type MessageRole = 'user' | 'assistant' | 'system';

export type MessageStatus = 'streaming' | 'pending' | 'done' | 'error';

export interface ToolCall {
  id: string;
  toolName: string;
  args: Record<string, unknown>;
  result?: unknown;
  status: 'pending' | 'in_progress' | 'done' | 'error';
  error?: string;
}

export interface Attachment {
  id: string;
  type: 'image' | 'pdf' | 'document';
  url: string;
  name: string;
  size?: number;
}

export interface ChatMessage {
  id: string;
  matchId: string;
  role: MessageRole;
  content: string;
  toolInvocations?: ToolCall[];
  attachments?: Attachment[];
  createdAt: Date;
  status?: MessageStatus;
}

export interface ConversationThread {
  id: string;
  messages: ChatMessage[];
  createdAt: Date;
  updatedAt: Date;
  status: 'active' | 'archived';
}

export interface AIStreamState {
  messageId: string;
  partialContent: string;
  toolCalls: ToolCall[];
  isComplete: boolean;
  error?: string;
}

export interface ChatRequest {
  matchId: string;
  messages: Array<{
    id: string;
    role: MessageRole;
    content: string;
  }>;
  toolCall?: {
    toolName: string;
    args: Record<string, unknown>;
  };
}

export interface APIError {
  code: string;
  message: string;
  details?: unknown;
}
