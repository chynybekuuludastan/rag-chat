import { AuthGuard } from "@/components/layout/auth-guard";
import { Header } from "@/components/layout/header";

export default function ChatLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthGuard>
      <div className="flex h-screen flex-col">
        <Header />
        {children}
      </div>
    </AuthGuard>
  );
}
