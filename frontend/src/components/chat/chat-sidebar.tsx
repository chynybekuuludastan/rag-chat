"use client";

import { useTranslations } from "next-intl";
import { Plus, MessageSquare } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import type { ChatSession } from "@/types";

interface ChatSidebarProps {
  sessions: ChatSession[];
  currentSessionId: string | null;
  onSelectSession: (id: string) => void;
  onNewChat: () => void;
}

export function ChatSidebar({
  sessions,
  currentSessionId,
  onSelectSession,
  onNewChat,
}: ChatSidebarProps) {
  const t = useTranslations("chat");

  return (
    <div className="flex h-full flex-col">
      <div className="p-3">
        <Button
          variant="outline"
          size="sm"
          className="w-full justify-start gap-2"
          onClick={onNewChat}
        >
          <Plus className="size-4" />
          {t("new_chat")}
        </Button>
      </div>

      <div className="px-3 pb-1">
        <p className="text-xs font-medium text-muted-foreground">
          {t("sessions")}
        </p>
      </div>

      <ScrollArea className="flex-1 px-2">
        {sessions.length === 0 ? (
          <p className="px-3 py-6 text-center text-xs text-muted-foreground">
            {t("no_sessions")}
          </p>
        ) : (
          <div className="grid gap-0.5">
            {sessions.map((session) => (
              <button
                key={session.id}
                onClick={() => onSelectSession(session.id)}
                className={cn(
                  "flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm transition-colors hover:bg-accent",
                  currentSessionId === session.id && "bg-accent",
                )}
              >
                <MessageSquare className="size-4 shrink-0 text-muted-foreground" />
                <span className="truncate">{session.title}</span>
              </button>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}
