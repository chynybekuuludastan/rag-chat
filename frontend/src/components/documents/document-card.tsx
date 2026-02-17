"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { FileText, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import type { Document } from "@/types";

interface DocumentCardProps {
  document: Document;
  onDelete: (id: string) => Promise<void>;
}

export function DocumentCard({ document, onDelete }: DocumentCardProps) {
  const t = useTranslations("documents");
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleDelete() {
    setIsDeleting(true);
    try {
      await onDelete(document.id);
    } finally {
      setIsDeleting(false);
    }
  }

  const fileSize =
    document.file_size < 1024
      ? `${document.file_size} B`
      : document.file_size < 1024 * 1024
        ? `${(document.file_size / 1024).toFixed(1)} KB`
        : `${(document.file_size / (1024 * 1024)).toFixed(1)} MB`;

  const uploadDate = new Date(document.created_at).toLocaleDateString();

  return (
    <div className="flex items-center gap-3 rounded-lg border p-3">
      <div className="flex size-10 shrink-0 items-center justify-center rounded-md bg-muted">
        <FileText className="size-5 text-muted-foreground" />
      </div>

      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{document.filename}</p>
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span>{fileSize}</span>
          <span>&middot;</span>
          <span>
            {document.chunk_count} {t("chunks")}
          </span>
          <span>&middot;</span>
          <span>{uploadDate}</span>
        </div>
      </div>

      <Badge variant="secondary" className="shrink-0 uppercase">
        {document.file_type}
      </Badge>

      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="shrink-0 text-muted-foreground hover:text-destructive"
            disabled={isDeleting}
          >
            <Trash2 className="size-4" />
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("delete_confirm")}</AlertDialogTitle>
            <AlertDialogDescription>
              {t("delete_description")}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("cancel")}</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>
              {t("delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
