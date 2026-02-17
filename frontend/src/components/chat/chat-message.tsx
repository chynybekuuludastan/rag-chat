"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import ReactMarkdown from "react-markdown";
import rehypeHighlight from "rehype-highlight";
import { Copy, Check, ChevronDown, FileText, Bot, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { SourceChunk } from "@/types";

interface ChatMessageProps {
  role: "user" | "assistant";
  content: string;
  sources?: SourceChunk[];
  isStreaming?: boolean;
}

export function ChatMessage({
  role,
  content,
  sources,
  isStreaming,
}: ChatMessageProps) {
  const t = useTranslations("chat");
  const [copied, setCopied] = useState(false);
  const [sourcesOpen, setSourcesOpen] = useState(false);

  async function handleCopy() {
    await navigator.clipboard.writeText(content);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  return (
    <div
      className={cn(
        "group flex gap-3 py-4",
        role === "user" && "flex-row-reverse",
      )}
    >
      <div
        className={cn(
          "flex size-7 shrink-0 items-center justify-center rounded-full",
          role === "user"
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-muted-foreground",
        )}
      >
        {role === "user" ? (
          <User className="size-4" />
        ) : (
          <Bot className="size-4" />
        )}
      </div>

      <div
        className={cn(
          "min-w-0 max-w-[85%] space-y-2",
          role === "user" && "text-right",
        )}
      >
        <div
          className={cn(
            "inline-block rounded-lg text-sm",
            role === "user"
              ? "bg-muted px-3 py-2"
              : "prose prose-sm dark:prose-invert max-w-none",
          )}
        >
          {role === "user" ? (
            <p className="whitespace-pre-wrap">{content}</p>
          ) : (
            <ReactMarkdown rehypePlugins={[rehypeHighlight]}>
              {content}
            </ReactMarkdown>
          )}
          {isStreaming && (
            <span className="ml-0.5 inline-block h-4 w-1.5 animate-pulse bg-foreground" />
          )}
        </div>

        {role === "assistant" && !isStreaming && content && (
          <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 gap-1 px-2 text-xs text-muted-foreground"
              onClick={handleCopy}
            >
              {copied ? (
                <>
                  <Check className="size-3" />
                  {t("copied")}
                </>
              ) : (
                <>
                  <Copy className="size-3" />
                  {t("copy")}
                </>
              )}
            </Button>
          </div>
        )}

        {sources && sources.length > 0 && !isStreaming && (
          <div className="mt-2">
            <button
              onClick={() => setSourcesOpen(!sourcesOpen)}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
            >
              <ChevronDown
                className={cn(
                  "size-3 transition-transform",
                  sourcesOpen && "rotate-180",
                )}
              />
              {sources.length} {t("sources")}
            </button>
            {sourcesOpen && (
              <div className="mt-2 grid gap-1.5">
                {sources.map((source) => (
                  <div
                    key={source.id}
                    className="flex items-start gap-2 rounded-md border p-2 text-xs"
                  >
                    <FileText className="mt-0.5 size-3 shrink-0 text-muted-foreground" />
                    <div className="min-w-0">
                      <p className="font-medium">{source.document}</p>
                      <p className="line-clamp-2 text-muted-foreground">
                        {source.content}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
