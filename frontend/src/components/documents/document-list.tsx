"use client";

import { useTranslations } from "next-intl";
import { FileText } from "lucide-react";
import { DocumentCard } from "./document-card";
import { Skeleton } from "@/components/ui/skeleton";
import type { Document } from "@/types";

interface DocumentListProps {
  documents: Document[];
  isLoading: boolean;
  onDelete: (id: string) => Promise<void>;
}

export function DocumentList({
  documents,
  isLoading,
  onDelete,
}: DocumentListProps) {
  const t = useTranslations("documents");

  if (isLoading) {
    return (
      <div className="grid gap-2">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-16 w-full rounded-lg" />
        ))}
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 py-12 text-center">
        <FileText className="size-10 text-muted-foreground/50" />
        <p className="text-sm text-muted-foreground">{t("empty")}</p>
      </div>
    );
  }

  return (
    <div className="grid gap-2">
      {documents.map((doc) => (
        <DocumentCard key={doc.id} document={doc} onDelete={onDelete} />
      ))}
    </div>
  );
}
