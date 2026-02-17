"use client";

import { useCallback, useReducer, useRef } from "react";
import { api } from "@/lib/api";
import type { ChatEvent, ChatSession, Message, SourceChunk } from "@/types";

interface ChatState {
  messages: Message[];
  sessions: ChatSession[];
  currentSessionId: string | null;
  isStreaming: boolean;
  streamingContent: string;
  streamingSources: SourceChunk[];
  isLoadingSessions: boolean;
  isLoadingMessages: boolean;
}

type ChatAction =
  | { type: "SET_SESSIONS"; sessions: ChatSession[] }
  | { type: "SET_SESSIONS_LOADING"; loading: boolean }
  | { type: "SET_MESSAGES"; messages: Message[] }
  | { type: "SET_MESSAGES_LOADING"; loading: boolean }
  | { type: "SET_SESSION"; sessionId: string }
  | { type: "STREAM_START"; question: string }
  | { type: "STREAM_CHUNK"; content: string }
  | { type: "STREAM_SOURCES"; sources: SourceChunk[] }
  | {
      type: "STREAM_DONE";
      assistantMessage: Message;
      sessionId: string;
      sessionTitle?: string;
    }
  | { type: "STREAM_ERROR" }
  | { type: "NEW_CHAT" };

function chatReducer(state: ChatState, action: ChatAction): ChatState {
  switch (action.type) {
    case "SET_SESSIONS":
      return { ...state, sessions: action.sessions, isLoadingSessions: false };
    case "SET_SESSIONS_LOADING":
      return { ...state, isLoadingSessions: action.loading };
    case "SET_MESSAGES":
      return { ...state, messages: action.messages, isLoadingMessages: false };
    case "SET_MESSAGES_LOADING":
      return { ...state, isLoadingMessages: action.loading };
    case "SET_SESSION":
      return {
        ...state,
        currentSessionId: action.sessionId,
        messages: [],
        streamingContent: "",
        streamingSources: [],
      };
    case "STREAM_START":
      return {
        ...state,
        isStreaming: true,
        streamingContent: "",
        streamingSources: [],
        messages: [
          ...state.messages,
          {
            id: `temp-user-${Date.now()}`,
            session_id: state.currentSessionId ?? "",
            role: "user",
            content: action.question,
            created_at: new Date().toISOString(),
          },
        ],
      };
    case "STREAM_CHUNK":
      return {
        ...state,
        streamingContent: state.streamingContent + action.content,
      };
    case "STREAM_SOURCES":
      return { ...state, streamingSources: action.sources };
    case "STREAM_DONE": {
      const updatedSessions = state.sessions.some(
        (s) => s.id === action.sessionId,
      )
        ? state.sessions.map((s) =>
            s.id === action.sessionId && action.sessionTitle
              ? { ...s, title: action.sessionTitle }
              : s,
          )
        : [
            {
              id: action.sessionId,
              user_id: "",
              title: action.sessionTitle ?? "New Chat",
              created_at: new Date().toISOString(),
            },
            ...state.sessions,
          ];

      return {
        ...state,
        isStreaming: false,
        streamingContent: "",
        streamingSources: [],
        currentSessionId: action.sessionId,
        messages: [...state.messages, action.assistantMessage],
        sessions: updatedSessions,
      };
    }
    case "STREAM_ERROR":
      return {
        ...state,
        isStreaming: false,
        streamingContent: "",
        streamingSources: [],
      };
    case "NEW_CHAT":
      return {
        ...state,
        currentSessionId: null,
        messages: [],
        streamingContent: "",
        streamingSources: [],
      };
  }
}

const initialState: ChatState = {
  messages: [],
  sessions: [],
  currentSessionId: null,
  isStreaming: false,
  streamingContent: "",
  streamingSources: [],
  isLoadingSessions: false,
  isLoadingMessages: false,
};

export function useChat() {
  const [state, dispatch] = useReducer(chatReducer, initialState);
  const abortRef = useRef<AbortController | null>(null);

  const fetchSessions = useCallback(async () => {
    dispatch({ type: "SET_SESSIONS_LOADING", loading: true });
    try {
      const data = await api.request<{ sessions: ChatSession[] }>(
        "/chat/history",
      );
      dispatch({ type: "SET_SESSIONS", sessions: data.sessions ?? [] });
    } catch {
      dispatch({ type: "SET_SESSIONS_LOADING", loading: false });
    }
  }, []);

  const selectSession = useCallback(async (sessionId: string) => {
    dispatch({ type: "SET_SESSION", sessionId });
    dispatch({ type: "SET_MESSAGES_LOADING", loading: true });
    try {
      const data = await api.request<{ messages: Message[] }>(
        `/chat/history/${sessionId}`,
      );
      dispatch({ type: "SET_MESSAGES", messages: data.messages ?? [] });
    } catch {
      dispatch({ type: "SET_MESSAGES_LOADING", loading: false });
    }
  }, []);

  const sendMessage = useCallback(
    async (question: string) => {
      if (state.isStreaming || !question.trim()) return;

      dispatch({ type: "STREAM_START", question: question.trim() });

      const abortController = new AbortController();
      abortRef.current = abortController;

      let fullContent = "";
      let sources: SourceChunk[] = [];
      let sessionId = state.currentSessionId;

      try {
        const API_URL =
          process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api";
        const token = api.getToken();

        const res = await fetch(`${API_URL}/chat`, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            ...(token && { Authorization: `Bearer ${token}` }),
          },
          body: JSON.stringify({
            question: question.trim(),
            ...(state.currentSessionId && {
              session_id: state.currentSessionId,
            }),
          }),
          signal: abortController.signal,
        });

        if (!res.ok || !res.body) {
          dispatch({ type: "STREAM_ERROR" });
          return;
        }

        const responseSessionId = res.headers.get("X-Session-ID");
        if (responseSessionId) {
          sessionId = responseSessionId;
        }

        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() ?? "";

          for (const line of lines) {
            if (!line.startsWith("data: ")) continue;
            const jsonStr = line.slice(6).trim();
            if (!jsonStr) continue;

            try {
              const event: ChatEvent = JSON.parse(jsonStr);

              switch (event.type) {
                case "chunk":
                  fullContent += event.content;
                  dispatch({ type: "STREAM_CHUNK", content: event.content });
                  break;
                case "sources":
                  sources = event.chunks;
                  dispatch({ type: "STREAM_SOURCES", sources: event.chunks });
                  break;
                case "done":
                  break;
              }
            } catch {
              // Skip malformed JSON
            }
          }
        }

        const assistantMessage: Message = {
          id: `msg-${Date.now()}`,
          session_id: sessionId ?? "",
          role: "assistant",
          content: fullContent,
          source_chunks: sources.map((s) => s.id),
          created_at: new Date().toISOString(),
        };

        dispatch({
          type: "STREAM_DONE",
          assistantMessage,
          sessionId: sessionId ?? "",
          sessionTitle:
            question.trim().length > 50
              ? question.trim().slice(0, 50) + "..."
              : question.trim(),
        });
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") return;
        dispatch({ type: "STREAM_ERROR" });
      }
    },
    [state.isStreaming, state.currentSessionId],
  );

  const stopStreaming = useCallback(() => {
    abortRef.current?.abort();
    dispatch({ type: "STREAM_ERROR" });
  }, []);

  const newChat = useCallback(() => {
    dispatch({ type: "NEW_CHAT" });
  }, []);

  return {
    ...state,
    fetchSessions,
    selectSession,
    sendMessage,
    stopStreaming,
    newChat,
  };
}
