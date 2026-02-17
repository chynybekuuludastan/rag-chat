import type { ApiErrorResponse } from "@/types";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api";

export class ApiError extends Error {
  constructor(
    public status: number,
    public data: ApiErrorResponse,
  ) {
    super(data.message || data.error);
    this.name = "ApiError";
  }
}

class ApiClient {
  private accessToken: string | null = null;
  private refreshPromise: Promise<boolean> | null = null;

  setToken(token: string | null) {
    this.accessToken = token;
  }

  getToken() {
    return this.accessToken;
  }

  async request<T>(path: string, options?: RequestInit): Promise<T> {
    const res = await fetch(`${API_URL}${path}`, {
      ...options,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(this.accessToken && {
          Authorization: `Bearer ${this.accessToken}`,
        }),
        ...options?.headers,
      },
    });

    if (res.status === 401) {
      const refreshed = await this.refresh();
      if (refreshed) return this.request<T>(path, options);
      if (typeof window !== "undefined") {
        window.location.href = "/auth/login";
      }
      throw new ApiError(401, {
        error: "unauthorized",
        message: "Session expired",
      });
    }

    if (!res.ok) {
      const error: ApiErrorResponse = await res.json().catch(() => ({
        error: "unknown",
        message: "An unexpected error occurred",
      }));
      throw new ApiError(res.status, error);
    }

    return res.json() as Promise<T>;
  }

  async upload<T>(path: string, formData: FormData): Promise<T> {
    const res = await fetch(`${API_URL}${path}`, {
      method: "POST",
      credentials: "include",
      headers: {
        ...(this.accessToken && {
          Authorization: `Bearer ${this.accessToken}`,
        }),
      },
      body: formData,
    });

    if (res.status === 401) {
      const refreshed = await this.refresh();
      if (refreshed) return this.upload<T>(path, formData);
      if (typeof window !== "undefined") {
        window.location.href = "/auth/login";
      }
      throw new ApiError(401, {
        error: "unauthorized",
        message: "Session expired",
      });
    }

    if (!res.ok) {
      const error: ApiErrorResponse = await res.json().catch(() => ({
        error: "unknown",
        message: "An unexpected error occurred",
      }));
      throw new ApiError(res.status, error);
    }

    return res.json() as Promise<T>;
  }

  private async refresh(): Promise<boolean> {
    if (this.refreshPromise) return this.refreshPromise;

    this.refreshPromise = (async () => {
      try {
        const res = await fetch(`${API_URL}/auth/refresh`, {
          method: "POST",
          credentials: "include",
        });

        if (!res.ok) return false;

        const data = await res.json();
        this.accessToken = data.access_token;
        return true;
      } catch {
        return false;
      } finally {
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }
}

export const api = new ApiClient();
