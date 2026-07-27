import { redirect, notFound } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import LinkActions from "@/components/LinkActions";

type PaymentLink = {
  id: string;
  public_id: string;
  title: string;
  description?: string;
  amount_minor: number;
  currency: string;
  allowed_methods: string[];
  status: string;
  url: string;
};

export default async function LinkDetail({ params }: { params: { id: string } }) {
  const res = await serverGet(`/payment-links/${params.id}`);
  if (res.status === 401) redirect("/login");
  if (res.status === 404) notFound();
  if (!res.ok) throw new Error(`link failed: ${res.status}`);
  const env = await res.json();
  const link: PaymentLink = env.data;

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <a href="/links" className="text-sm text-paycore-muted hover:text-paycore-text">← Payment Links</a>
      <div className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-6 mt-4">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{link.title}</h1>
            <p className="text-paycore-muted mt-1">{formatMoney(link.amount_minor, link.currency)}</p>
          </div>
          <span className={`inline-block rounded-full px-3 py-1 text-xs font-medium ${
            link.status === "active" ? "bg-paycore-successBg text-paycore-success" : "bg-paycore-line2 text-paycore-text2"
          }`}>{link.status}</span>
        </div>

        {link.description && <p className="mt-4 text-sm text-paycore-muted">{link.description}</p>}

        <div className="mt-4 flex flex-wrap gap-2">
          {(link.allowed_methods.length ? link.allowed_methods : ["ทุกช่องทาง"]).map((m) => (
            <span key={m} className="rounded-full px-3 py-1 text-xs border border-paycore-line bg-paycore-surface2 text-paycore-text2">{m}</span>
          ))}
        </div>

        <LinkActions id={link.id} url={link.url} status={link.status} />
      </div>
    </main>
  );
}
