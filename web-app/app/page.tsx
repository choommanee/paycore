import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";

type AuthMe = {
  user_id: string;
  merchant_id: string;
  email: string;
  name: string;
  merchant_name: string;
};

export default async function DashboardHome() {
  const res = await serverGet("/auth/me");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`auth/me failed: ${res.status}`);
  const env = await res.json();
  const me: AuthMe = env.data;

  return (
    <main className="min-h-screen p-8">
      <header className="flex items-center justify-between mb-8">
        <h1 className="text-xl font-semibold">PayCore Dashboard</h1>
        <form action="/api/auth/logout" method="post">
          <button className="text-sm text-paycore-muted hover:text-paycore-text">ออกจากระบบ</button>
        </form>
      </header>
      <section className="rounded-xl2 bg-paycore-surface p-6">
        <p className="text-paycore-muted text-sm">เข้าสู่ระบบในชื่อ</p>
        <p className="text-lg font-medium">{me.name || me.email}</p>
        <p className="text-paycore-muted text-xs mt-1">merchant: {me.merchant_id}</p>
        <a href="/links" className="inline-block mt-4 text-paycore-primary hover:underline">จัดการ Payment Links →</a>
      </section>
    </main>
  );
}
