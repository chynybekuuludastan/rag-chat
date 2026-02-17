"use client";

import { useCallback, useMemo, useSyncExternalStore } from "react";
import { api, ApiError } from "@/lib/api";
import type { LoginResponse, RegisterResponse } from "@/types";

let listeners: (() => void)[] = [];

function emitChange() {
  for (const listener of listeners) {
    listener();
  }
}

function subscribe(listener: () => void) {
  listeners = [...listeners, listener];
  return () => {
    listeners = listeners.filter((l) => l !== listener);
  };
}

function getSnapshot() {
  return api.getToken();
}

function getServerSnapshot() {
  return null;
}

export function useAuth() {
  const token = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const isAuthenticated = token !== null;

  const login = useCallback(async (email: string, password: string) => {
    const data = await api.request<LoginResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    api.setToken(data.access_token);
    emitChange();
    return data;
  }, []);

  const register = useCallback(async (email: string, password: string) => {
    const data = await api.request<RegisterResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    return data;
  }, []);

  const logout = useCallback(() => {
    api.setToken(null);
    emitChange();
    window.location.href = "/auth/login";
  }, []);

  return useMemo(
    () => ({ isAuthenticated, login, register, logout }),
    [isAuthenticated, login, register, logout],
  );
}

export function handleApiError(error: unknown): string {
  if (error instanceof ApiError) {
    return error.data.message || error.data.error;
  }
  return "An unexpected error occurred";
}
