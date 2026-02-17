"use client";

import { useEffect } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Header } from "@/components/layout/header";
import { UploadZone } from "@/components/documents/upload-zone";
import { DocumentList } from "@/components/documents/document-list";
import { useDocuments } from "@/hooks/use-documents";
import { handleApiError } from "@/hooks/use-auth";

export default function DocumentsPage() {
  const t = useTranslations("documents");
  const tErrors = useTranslations("errors");
  const {
    documents,
    isLoading,
    isUploading,
    fetchDocuments,
    uploadDocument,
    deleteDocument,
  } = useDocuments();

  useEffect(() => {
    fetchDocuments();
  }, [fetchDocuments]);

  async function handleUpload(file: File) {
    try {
      await uploadDocument(file);
    } catch (err) {
      const message = handleApiError(err);
      toast.error(message || tErrors("generic"));
    }
  }

  async function handleDelete(id: string) {
    try {
      await deleteDocument(id);
    } catch (err) {
      const message = handleApiError(err);
      toast.error(message || tErrors("generic"));
    }
  }

  return (
    <div className="flex min-h-screen flex-col">
      <Header />
      <main className="mx-auto w-full max-w-3xl flex-1 px-4 py-6">
        <h1 className="mb-6 text-lg font-semibold">{t("title")}</h1>
        <div className="grid gap-6">
          <UploadZone onUpload={handleUpload} isUploading={isUploading} />
          <DocumentList
            documents={documents}
            isLoading={isLoading}
            onDelete={handleDelete}
          />
        </div>
      </main>
    </div>
  );
}
