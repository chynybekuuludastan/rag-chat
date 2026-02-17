"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/i18n/navigation";
import { MessageSquare, FileText, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "./theme-toggle";
import { LocaleSwitcher } from "./locale-switcher";
import { useAuth } from "@/hooks/use-auth";
import { cn } from "@/lib/utils";

export function Header() {
  const t = useTranslations("nav");
  const pathname = usePathname();
  const { isAuthenticated, logout } = useAuth();

  const navItems = [
    { href: "/chat" as const, label: t("chat"), icon: MessageSquare },
    { href: "/documents" as const, label: t("documents"), icon: FileText },
  ];

  return (
    <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-14 items-center gap-4 px-4">
        <Link href="/chat" className="text-sm font-semibold tracking-tight">
          RAG Chat
        </Link>

        {isAuthenticated && (
          <nav className="flex items-center gap-1">
            {navItems.map(({ href, label, icon: Icon }) => (
              <Link key={href} href={href}>
                <Button
                  variant="ghost"
                  size="sm"
                  className={cn(
                    "gap-2 text-muted-foreground",
                    pathname.startsWith(href) &&
                      "bg-accent text-accent-foreground",
                  )}
                >
                  <Icon className="size-4" />
                  {label}
                </Button>
              </Link>
            ))}
          </nav>
        )}

        <div className="ml-auto flex items-center gap-1">
          <LocaleSwitcher />
          <ThemeToggle />
          {isAuthenticated && (
            <Button
              variant="ghost"
              size="icon"
              onClick={logout}
              aria-label="Sign out"
            >
              <LogOut className="size-4" />
            </Button>
          )}
        </div>
      </div>
    </header>
  );
}
