"use client";

import { useEffect, useRef } from "react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Skeleton } from "@/components/ui/skeleton";
import { ChatMessage } from "./chat-message";
import { ChatEmpty } from "./chat-empty";
import { ChatInput } from "./chat-input";
import type { Message, SourceChunk } from "@/types";

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
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, streamingContent]);

  return (
    <div className="flex h-full flex-col">
      <ScrollArea className="flex-1">
        <div className="mx-auto max-w-3xl px-4">
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
      </ScrollArea>
      <ChatInput onSend={onSend} onStop={onStop} isStreaming={isStreaming} />
    </div>
  );
}
