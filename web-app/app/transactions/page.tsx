import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

type Payment = {
  id: string;
  amount: string;
  currency: string;
  status: string;
  card_brand?: string;
  card_last4?: string;
  reference?: string;
  created_at: string;
};

const PAGE_SIZE = 25;

const STATUS_LABEL: Record<string, string> = {
  authorized: "อนุมัติแล้ว",
  captured: "เรียกเก็บแล้ว",
  partial_refunded: "คืนบางส่วน",
  refunded: "คืนเงินแล้ว",
  voided: "ยกเลิก",
  failed: "ล้มเหลว",
  requires_action: "รอยืนยัน",
};

// Semantic pill colors on a light surface.
const STATUS_PILL: Record<string, string> = {
  authorized: "bg-paycore-accentBg text-paycore-accentInk",
  captured: "bg-paycore-successBg text-paycore-success",
  partial_refunded: "bg-paycore-warnBg text-paycore-warn",
  refunded: "bg-paycore-warnBg text-paycore-warn",
  voided: "bg-paycore-dangerBg text-paycore-danger",
  failed: "bg-paycore-dangerBg text-paycore-danger",
  requires_action: "bg-paycore-warnBg text-paycore-warn",
};
const pillClass = (s: string) => STATUS_PILL[s] ?? "bg-paycore-line2 text-paycore-text2";

export default async function TransactionsPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = Math.max(1, parseInt(searchParams.page ?? "1", 10) || 1);
  const offset = (page - 1) * PAGE_SIZE;
  const res = await serverGet(`/payments?limit=${PAGE_SIZE}&offset=${offset}`);
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`payments failed: ${res.status}`);
  const items: Payment[] = (await res.json()).data ?? [];

  return (
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <DashboardNav active="/transactions" />
      <h1 className="text-xl font-semibold mb-1">ธุรกรรม (บัตร)</h1>
      <p className="text-paycore-muted text-xs mb-6">
        แสดงเฉพาะการชำระด้วยบัตร PromptPay และ e-wallet ยังไม่รวมในรายการนี้
      </p>

      <section className="space-y-2">
        {items.length === 0 && (
          <p className="text-paycore-muted text-sm">ยังไม่มีธุรกรรม</p>
        )}
        {items.map((p) => (
          <a
            key={p.id}
            href={`/transactions/${p.id}`}
            className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-4 flex items-center justify-between hover:bg-paycore-surface2 transition-colors"
          >
            <div>
              <p className="font-medium">{formatDecimalMoney(p.amount, p.currency)}</p>
              <p className="text-paycore-muted text-xs mt-1">
                {(p.card_brand || "card")}{p.card_last4 ? ` ····${p.card_last4}` : ""}
                {p.reference ? ` · ${p.reference}` : ""}
              </p>
            </div>
            <div className="text-right">
              <span className={`inline-block rounded-full px-3 py-1 text-xs font-medium ${pillClass(p.status)}`}>
                {STATUS_LABEL[p.status] ?? p.status}
              </span>
              <p className="text-paycore-muted text-xs mt-1">
                {new Date(p.created_at).toLocaleString("th-TH")}
              </p>
            </div>
          </a>
        ))}
      </section>

      <nav className="mt-6 flex items-center justify-between text-sm">
        {page > 1 ? (
          <a href={`/transactions?page=${page - 1}`} className="text-paycore-primary hover:underline">← ก่อนหน้า</a>
        ) : (
          <span />
        )}
        <span className="text-paycore-muted">หน้า {page}</span>
        {items.length === PAGE_SIZE ? (
          <a href={`/transactions?page=${page + 1}`} className="text-paycore-primary hover:underline">ถัดไป →</a>
        ) : (
          <span />
        )}
      </nav>
    </main>
  );
}
