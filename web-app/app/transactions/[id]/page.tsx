import { redirect, notFound } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";
import RefundForm from "@/components/RefundForm";

type Payment = {
  id: string;
  amount: string;
  captured_amount: string;
  refunded_amount: string;
  currency: string;
  status: string;
  card_brand?: string;
  card_last4?: string;
  acquirer_ref?: string;
  auth_code?: string;
  reference?: string;
  created_at: string;
  updated_at: string;
};

// Refund is only allowed by the service from captured / partial_refunded.
const REFUNDABLE = new Set(["captured", "partial_refunded"]);

export default async function TransactionDetail({ params }: { params: { id: string } }) {
  const res = await serverGet(`/payments/${params.id}`);
  if (res.status === 401) redirect("/login");
  if (res.status === 404) notFound();
  if (!res.ok) throw new Error(`payment failed: ${res.status}`);
  const p: Payment = (await res.json()).data;

  const captured = parseFloat(p.captured_amount || "0");
  const refunded = parseFloat(p.refunded_amount || "0");
  const remaining = Math.max(0, captured - refunded);

  const rows: [string, string][] = [
    ["ยอด", formatDecimalMoney(p.amount, p.currency)],
    ["เรียกเก็บแล้ว", formatDecimalMoney(p.captured_amount, p.currency)],
    ["คืนแล้ว", formatDecimalMoney(p.refunded_amount, p.currency)],
    ["บัตร", `${p.card_brand || "card"}${p.card_last4 ? ` ····${p.card_last4}` : ""}`],
    ["อ้างอิงผู้รับชำระ", p.acquirer_ref || "—"],
    ["รหัสอนุมัติ", p.auth_code || "—"],
    ["อ้างอิงร้านค้า", p.reference || "—"],
    ["สร้างเมื่อ", new Date(p.created_at).toLocaleString("th-TH")],
  ];

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <DashboardNav active="/transactions" />
      <a href="/transactions" className="text-sm text-paycore-muted hover:text-paycore-text">← ธุรกรรม</a>

      <div className="rounded-xl2 bg-paycore-surface p-6 mt-4">
        <div className="flex items-start justify-between">
          <h1 className="text-2xl font-semibold">{formatDecimalMoney(p.amount, p.currency)}</h1>
          <span className="rounded-full px-3 py-1 text-xs bg-paycore-bg border border-white/10">{p.status}</span>
        </div>

        <dl className="mt-6 space-y-2 text-sm">
          {rows.map(([k, v]) => (
            <div key={k} className="flex justify-between">
              <dt className="text-paycore-muted">{k}</dt>
              <dd>{v}</dd>
            </div>
          ))}
        </dl>

        {REFUNDABLE.has(p.status) ? (
          <RefundForm paymentId={p.id} remaining={remaining} currency={p.currency} />
        ) : (
          <p className="mt-6 border-t border-white/10 pt-6 text-paycore-muted text-sm">
            สถานะนี้คืนเงินไม่ได้
          </p>
        )}
      </div>
    </main>
  );
}
