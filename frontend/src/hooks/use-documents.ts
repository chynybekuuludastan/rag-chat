"use client";

import { useCallback, useReducer } from "react";
import { api } from "@/lib/api";
import type { Document } from "@/types";

interface DocumentsState {
  documents: Document[];
  isLoading: boolean;
  isUploading: boolean;
  error: string | null;
}

type DocumentsAction =
  | { type: "FETCH_START" }
  | { type: "FETCH_SUCCESS"; documents: Document[] }
  | { type: "FETCH_ERROR"; error: string }
  | { type: "UPLOAD_START" }
  | { type: "UPLOAD_SUCCESS"; document: Document }
  | { type: "UPLOAD_ERROR"; error: string }
  | { type: "DELETE_SUCCESS"; id: string };

function documentsReducer(
  state: DocumentsState,
  action: DocumentsAction,
): DocumentsState {
  switch (action.type) {
    case "FETCH_START":
      return { ...state, isLoading: true, error: null };
    case "FETCH_SUCCESS":
      return { ...state, isLoading: false, documents: action.documents };
    case "FETCH_ERROR":
      return { ...state, isLoading: false, error: action.error };
    case "UPLOAD_START":
      return { ...state, isUploading: true, error: null };
    case "UPLOAD_SUCCESS":
      return {
        ...state,
        isUploading: false,
        documents: [action.document, ...state.documents],
      };
    case "UPLOAD_ERROR":
      return { ...state, isUploading: false, error: action.error };
    case "DELETE_SUCCESS":
      return {
        ...state,
        documents: state.documents.filter((d) => d.id !== action.id),
      };
  }
}

const initialState: DocumentsState = {
  documents: [],
  isLoading: false,
  isUploading: false,
  error: null,
};

export function useDocuments() {
  const [state, dispatch] = useReducer(documentsReducer, initialState);

  const fetchDocuments = useCallback(async () => {
    dispatch({ type: "FETCH_START" });
    try {
      const data = await api.request<{ documents: Document[] }>("/documents");
      dispatch({ type: "FETCH_SUCCESS", documents: data.documents });
    } catch (error) {
      dispatch({
        type: "FETCH_ERROR",
        error: error instanceof Error ? error.message : "Failed to load documents",
      });
    }
  }, []);

  const uploadDocument = useCallback(async (file: File) => {
    dispatch({ type: "UPLOAD_START" });
    try {
      const formData = new FormData();
      formData.append("file", file);
      const document = await api.upload<Document>("/documents", formData);
      dispatch({ type: "UPLOAD_SUCCESS", document });
      return document;
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Failed to upload document";
      dispatch({ type: "UPLOAD_ERROR", error: message });
      throw error;
    }
  }, []);

  const deleteDocument = useCallback(async (id: string) => {
    await api.request(`/documents/${id}`, { method: "DELETE" });
    dispatch({ type: "DELETE_SUCCESS", id });
  }, []);

  return { ...state, fetchDocuments, uploadDocument, deleteDocument };
}
