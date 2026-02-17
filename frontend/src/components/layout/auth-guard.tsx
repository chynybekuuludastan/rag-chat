"use client";

import { useEffect, useState } from "react";
import { useAuth, tryRefresh } from "@/hooks/use-auth";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();
  const [isReady, setIsReady] = useState(isAuthenticated);

  useEffect(() => {
    if (isAuthenticated) {
      setIsReady(true);
      return;
    }

    tryRefresh().then((ok) => {
      if (ok) {
        setIsReady(true);
      } else {
        window.location.href = "/auth/login";
      }
    });
  }, [isAuthenticated]);

  if (!isReady) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="size-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    );
  }

  return <>{children}</>;
}
