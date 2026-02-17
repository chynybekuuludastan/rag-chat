"use client";

import { ChatSidebar } from "@/components/chat/chat-sidebar";
import { ChatView } from "@/components/chat/chat-view";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useChat } from "@/hooks/use-chat";
import { Menu } from "lucide-react";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";

export default function ChatPage() {
  const t = useTranslations("chat");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const chat = useChat();

  useEffect(() => {
    chat.fetchSessions();
  }, [chat.fetchSessions]);

  function handleSelectSession(id: string) {
    chat.selectSession(id);
    setSidebarOpen(false);
  }

  function handleNewChat() {
    chat.newChat();
    setSidebarOpen(false);
  }

  const sidebar = (
    <ChatSidebar
      sessions={chat.sessions}
      currentSessionId={chat.currentSessionId}
      onSelectSession={handleSelectSession}
      onNewChat={handleNewChat}
    />
  );

  return (
    <div className="flex flex-1 overflow-hidden">
      {/* Desktop sidebar */}
      <aside className="hidden w-64 shrink-0 border-r md:block">
        {sidebar}
      </aside>

      {/* Mobile sidebar */}
      <div className="flex items-center border-b px-2 md:hidden">
        <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon">
              <Menu className="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-64 p-0">
            <SheetTitle className="sr-only">{t("sessions")}</SheetTitle>
            {sidebar}
          </SheetContent>
        </Sheet>
        <span className="text-sm font-medium">{t("title")}</span>
      </div>

      {/* Main chat area */}
      <main className="flex flex-1 flex-col overflow-hidden">
        <ChatView
          messages={chat.messages}
          streamingContent={chat.streamingContent}
          streamingSources={chat.streamingSources}
          isStreaming={chat.isStreaming}
          isLoadingMessages={chat.isLoadingMessages}
          onSend={chat.sendMessage}
          onStop={chat.stopStreaming}
        />
      </main>
    </div>
  );
}
