import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import CreateLinkForm from "@/components/CreateLinkForm";

type PaymentLink = {
  id: string;
  public_id: string;
  title: string;
  amount_minor: number;
  currency: string;
  status: string;
  url: string;
};

export default async function LinksPage() {
  const res = await serverGet("/payment-links");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`payment-links failed: ${res.status}`);
  const env = await res.json();
  const links: PaymentLink[] = env.data ?? [];

  return (
    <main className="min-h-screen p-8 max-w-3xl mx-auto">
      <header className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold">Payment Links</h1>
        <a href="/dashboard" className="text-sm text-paycore-muted hover:text-paycore-text">← Dashboard</a>
      </header>

      <CreateLinkForm />

      <section className="mt-8 space-y-3">
        {links.length === 0 && (
          <p className="text-paycore-muted text-sm">ยังไม่มีลิงก์ — สร้างอันแรกด้านบน</p>
        )}
        {links.map((l) => (
          <div key={l.id} className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-4 flex items-center justify-between">
            <div>
              <a href={`/links/${l.id}`} className="font-medium hover:underline">{l.title}</a>
              <p className="text-paycore-muted text-sm">{formatMoney(l.amount_minor, l.currency)} · {l.status}</p>
            </div>
            <a href={l.url} target="_blank" className="text-paycore-primary text-sm hover:underline">เปิดลิงก์ ↗</a>
          </div>
        ))}
      </section>
    </main>
  );
}
