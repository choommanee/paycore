import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

// A settlement row is one payout period: the platform sums a merchant's payments
// over a settlement window, deducts refunds + fees, and transfers net_minor to
// the merchant's bank account. All amounts are minor units (สตางค์).
type Settlement = {
  id: string;
  currency: string;
  gross_minor: number;
  refunded_minor: number;
  fee_minor: number;
  net_minor: number;
  payment_count: number;
  status: string;
  period_start: string;
  period_end: string;
  created_at: string;
};

// Settlement-specific status pill. The shared TxPresentation StatusPill labels a
// payment as "สำเร็จ" and has no "pending" entry; here we want payout semantics
// (จ่ายแล้ว / รอโอน) mapped to the paycore semantic tokens.
type Tone = "success" | "warn" | "muted";
const TONE_CLASS: Record<Tone, string> = {
  success: "bg-paycore-successBg text-paycore-success",
  warn: "bg-paycore-warnBg text-paycore-warn",
  muted: "bg-paycore-line2 text-paycore-muted",
};
const PAYOUT_STATUS: Record<string, { label: string; tone: Tone }> = {
  paid: { label: "จ่ายแล้ว", tone: "success" },
  pending: { label: "รอโอน", tone: "warn" },
  processing: { label: "กำลังโอน", tone: "warn" },
  failed: { label: "ล้มเหลว", tone: "muted" },
  canceled: { label: "ยกเลิก", tone: "muted" },
};

function PayoutPill({ status }: { status: string }) {
  const m = PAYOUT_STATUS[status] ?? { label: status, tone: "muted" as Tone };
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold ${TONE_CLASS[m.tone]}`}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-90" />
      {m.label}
    </span>
  );
}

// Compact day-month date for the settlement window (e.g. "1 ก.ค.").
function periodDate(iso: string): string {
  return new Intl.DateTimeFormat("th-TH", { day: "numeric", month: "short" }).format(
    new Date(iso),
  );
}

export default async function PayoutsPage() {
  const res = await serverGet("/settlements?limit=50");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`settlements failed: ${res.status}`);
  const rows: Settlement[] = (await res.json()).data ?? [];

  const currency = rows[0]?.currency ?? "THB";
  const totalNet = rows.reduce((sum, r) => sum + r.net_minor, 0);
  const paidNet = rows
    .filter((r) => r.status === "paid")
    .reduce((sum, r) => sum + r.net_minor, 0);

  return (
    <main className="min-h-screen bg-paycore-bg p-6 md:p-8">
      <div className="mx-auto max-w-5xl">
        <DashboardNav active="/payouts" />

        {/* Page header */}
        <div className="mb-6">
          <h1 className="text-lg font-semibold text-paycore-text">
            การจ่ายเงินเข้าบัญชี (Payouts)
          </h1>
          <p className="mt-0.5 text-xs text-paycore-muted">
            ยอดรับสุทธิที่โอนเข้าบัญชีธนาคารของคุณตามรอบ settlement
          </p>
        </div>

        {/* Summary tiles */}
        <section className="grid grid-cols-2 gap-4 md:grid-cols-3">
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              รับสุทธิรวม
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {formatMoney(totalNet, currency)}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              รวมทุกรอบที่แสดง
            </p>
          </div>

          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              โอนเข้าบัญชีแล้ว
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-success">
              {formatMoney(paidNet, currency)}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              เฉพาะรอบที่จ่ายแล้ว
            </p>
          </div>

          <div className="col-span-2 rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card md:col-span-1">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              รอบจ่ายเงิน
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {rows.length.toLocaleString("th-TH")}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              รอบ settlement ทั้งหมด
            </p>
          </div>
        </section>

        {/* Payouts table */}
        <section className="mt-6 rounded-xl2 border border-paycore-line bg-paycore-surface shadow-card">
          <div className="flex items-center justify-between px-5 py-4">
            <h2 className="text-sm font-semibold text-paycore-text">รอบจ่ายเงินทั้งหมด</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    รอบ
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    จำนวนรายการ
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    ยอดรวม
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    คืนเงิน
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    ค่าธรรมเนียม
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    รับสุทธิ
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    สถานะ
                  </th>
                </tr>
              </thead>
              <tbody>
                {rows.length === 0 && (
                  <tr>
                    <td
                      colSpan={7}
                      className="px-5 py-12 text-center text-sm text-paycore-muted"
                    >
                      ยังไม่มีรอบจ่ายเงิน — ระบบสรุปยอดและโอนเข้าบัญชีตามรอบ settlement
                    </td>
                  </tr>
                )}
                {rows.map((r) => (
                  <tr key={r.id} className="border-b border-paycore-line2 last:border-b-0">
                    <td className="px-5 py-3">
                      <div className="font-medium text-paycore-text">
                        {periodDate(r.period_start)} – {periodDate(r.period_end)}
                      </div>
                      <div className="font-mono text-[11px] text-paycore-muted">
                        {r.id.slice(0, 8)}…
                      </div>
                    </td>
                    <td className="px-5 py-3 text-right tabular-nums text-paycore-text2">
                      {r.payment_count.toLocaleString("th-TH")}
                    </td>
                    <td className="px-5 py-3 text-right tabular-nums text-paycore-text2">
                      {formatMoney(r.gross_minor, r.currency)}
                    </td>
                    <td className="px-5 py-3 text-right tabular-nums text-paycore-muted">
                      {r.refunded_minor > 0
                        ? `−${formatMoney(r.refunded_minor, r.currency)}`
                        : formatMoney(0, r.currency)}
                    </td>
                    <td className="px-5 py-3 text-right tabular-nums text-paycore-muted">
                      {r.fee_minor > 0
                        ? `−${formatMoney(r.fee_minor, r.currency)}`
                        : formatMoney(0, r.currency)}
                    </td>
                    <td className="px-5 py-3 text-right font-semibold tabular-nums text-paycore-text">
                      {formatMoney(r.net_minor, r.currency)}
                    </td>
                    <td className="px-5 py-3">
                      <PayoutPill status={r.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <p className="mt-4 text-xs text-paycore-muted">
          ยอดรับสุทธิจะถูกโอนเข้าบัญชีตามรอบ settlement · รับสุทธิ = ยอดรวม − คืนเงิน − ค่าธรรมเนียม
          (ค่าธรรมเนียมหักจากยอดรวม)
        </p>
      </div>
    </main>
  );
}
