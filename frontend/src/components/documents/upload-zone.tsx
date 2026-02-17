"use client";

import { useCallback } from "react";
import { useDropzone } from "react-dropzone";
import { useTranslations } from "next-intl";
import { Upload } from "lucide-react";
import { cn } from "@/lib/utils";

const ACCEPTED_TYPES = {
  "text/plain": [".txt"],
  "text/markdown": [".md"],
  "application/pdf": [".pdf"],
};

const MAX_SIZE = 10 * 1024 * 1024; // 10MB

interface UploadZoneProps {
  onUpload: (file: File) => void;
  isUploading: boolean;
}

export function UploadZone({ onUpload, isUploading }: UploadZoneProps) {
  const t = useTranslations("documents");

  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      const file = acceptedFiles[0];
      if (file) onUpload(file);
    },
    [onUpload],
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: ACCEPTED_TYPES,
    maxSize: MAX_SIZE,
    multiple: false,
    disabled: isUploading,
  });

  return (
    <div
      {...getRootProps()}
      className={cn(
        "flex cursor-pointer flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-8 text-center transition-colors",
        isDragActive
          ? "border-primary bg-primary/5"
          : "border-muted-foreground/25 hover:border-primary/50",
        isUploading && "pointer-events-none opacity-50",
      )}
    >
      <input {...getInputProps()} />
      <Upload className="size-8 text-muted-foreground" />
      <p className="text-sm font-medium">
        {isUploading ? t("uploading") : t("upload")}
      </p>
      <p className="text-xs text-muted-foreground">{t("upload_hint")}</p>
    </div>
  );
}
