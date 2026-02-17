export interface User {
  id: string;
  email: string;
  created_at: string;
}

export interface Document {
  id: string;
  user_id: string;
  filename: string;
  file_type: string;
  file_size: number;
  chunk_count: number;
  created_at: string;
}

export interface ChatSession {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
}

export interface Message {
  id: string;
  session_id: string;
  role: "user" | "assistant";
  content: string;
  source_chunks?: string[];
  sources?: SourceChunk[];
  created_at: string;
}

export interface SourceChunk {
  id: string;
  document: string;
  content: string;
}

export type ChatEventChunk = {
  type: "chunk";
  content: string;
};

export type ChatEventSources = {
  type: "sources";
  chunks: SourceChunk[];
};

export type ChatEventDone = {
  type: "done";
};

export type ChatEvent = ChatEventChunk | ChatEventSources | ChatEventDone;

export interface LoginResponse {
  access_token: string;
  expires_at: number;
}

export interface RegisterResponse {
  id: string;
  email: string;
  created_at: string;
}

export interface ApiErrorResponse {
  error: string;
  message: string;
}
