"use client";

import { useTranslations } from "next-intl";
import { MessageSquare } from "lucide-react";

export function ChatEmpty() {
  const t = useTranslations("chat");

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-3 text-center">
      <MessageSquare className="size-10 text-muted-foreground/50" />
      <div>
        <p className="font-medium">{t("empty_title")}</p>
        <p className="text-sm text-muted-foreground">{t("empty_description")}</p>
      </div>
    </div>
  );
}
