"use client";

import { Skeleton } from "@/components/ui/skeleton";
import type { Message, SourceChunk } from "@/types";
import { useEffect, useRef } from "react";
import { ChatEmpty } from "./chat-empty";
import { ChatInput } from "./chat-input";
import { ChatMessage } from "./chat-message";

interface ChatViewProps {
  messages: Message[];
  streamingContent: string;
  streamingSources: SourceChunk[];
  isStreaming: boolean;
  isLoadingMessages: boolean;
  onSend: (message: string) => void;
  onStop: () => void;
}

export function ChatView({
  messages,
  streamingContent,
  streamingSources,
  isStreaming,
  isLoadingMessages,
  onSend,
  onStop,
}: ChatViewProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const shouldAutoScroll = useRef(true);

  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;

    function handleScroll() {
      if (!el) return;
      const { scrollTop, scrollHeight, clientHeight } = el;
      shouldAutoScroll.current = scrollHeight - scrollTop - clientHeight < 80;
    }

    el.addEventListener("scroll", handleScroll);
    return () => el.removeEventListener("scroll", handleScroll);
  }, []);

  useEffect(() => {
    if (shouldAutoScroll.current) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, streamingContent]);

  return (
    <div className="flex h-full flex-col">
      <div ref={scrollRef} className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl px-4 py-4">
          {isLoadingMessages ? (
            <div className="grid gap-4 py-8">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : messages.length === 0 && !isStreaming ? (
            <ChatEmpty />
          ) : (
            <div className="py-4">
              {messages.map((msg) => (
                <ChatMessage
                  key={msg.id}
                  role={msg.role}
                  content={msg.content}
                  sources={msg.sources}
                />
              ))}
              {isStreaming && (
                <ChatMessage
                  role="assistant"
                  content={streamingContent}
                  sources={streamingSources}
                  isStreaming
                />
              )}
              <div ref={bottomRef} />
            </div>
          )}
        </div>
      </div>
      <ChatInput onSend={onSend} onStop={onStop} isStreaming={isStreaming} />
    </div>
  );
}
