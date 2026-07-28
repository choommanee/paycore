import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";
import {
  MethodChip,
  StatusPill,
  shortDate,
  type Transaction,
} from "@/components/TxPresentation";

const PAGE_SIZE = 25;

export default async function TransactionsPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = Math.max(1, parseInt(searchParams.page ?? "1", 10) || 1);
  const offset = (page - 1) * PAGE_SIZE;
  const res = await serverGet(`/transactions?limit=${PAGE_SIZE}&offset=${offset}`);
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`transactions failed: ${res.status}`);
  const items: Transaction[] = (await res.json()).data ?? [];

  return (
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <DashboardNav active="/transactions" />
      <h1 className="text-xl font-semibold mb-1">ธุรกรรมทั้งหมด</h1>
      <p className="text-paycore-muted text-xs mb-6">
        ทุกช่องทาง — บัตร, PromptPay และ e-wallet
      </p>

      <section className="space-y-2">
        {items.length === 0 && (
          <p className="text-paycore-muted text-sm">ยังไม่มีธุรกรรม</p>
        )}
        {items.map((t) => {
          // Only card transactions have a detail + refund page (/payments/:id).
          const isCard = t.source === "card";
          const Row = isCard ? "a" : "div";
          const rowProps = isCard ? { href: `/transactions/${t.id}` } : {};
          return (
            <Row
              key={t.id}
              {...rowProps}
              className={`rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-4 flex items-center justify-between gap-4 ${
                isCard ? "hover:bg-paycore-surface2 transition-colors" : ""
              }`}
            >
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <MethodChip tx={t} />
                  <StatusPill status={t.status} />
                </div>
                <p className="text-paycore-muted text-xs mt-2 font-mono">
                  {t.id.slice(0, 8)}…
                  {t.reference ? (
                    <span className="font-sans"> · {t.reference}</span>
                  ) : null}
                  {isCard ? (
                    <span className="font-sans text-paycore-primary"> · ดูรายละเอียด →</span>
                  ) : null}
                </p>
              </div>
              <div className="text-right shrink-0">
                <p className="font-semibold tabular-nums">
                  {formatMoney(t.amount_minor, t.currency)}
                </p>
                <p className="text-paycore-muted text-xs mt-1">{shortDate(t.created_at)}</p>
              </div>
            </Row>
          );
        })}
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
