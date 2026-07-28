import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

// A dispute (chargeback) as returned by GET /v1/disputes. amount is a MAJOR-unit
// decimal string (e.g. "499.00"), not minor units — use formatDecimalMoney.
type Dispute = {
  id: string;
  payment_id: string;
  amount: string;
  currency: string;
  reason?: string;
  status: string;
  network_ref?: string;
  opened_at: string;
  resolved_at?: string;
  created_at: string;
};

type Tone = "warn" | "accent" | "success" | "danger" | "muted";
const TONE_CLASS: Record<Tone, string> = {
  warn: "bg-paycore-warnBg text-paycore-warn",
  accent: "bg-paycore-accentBg text-paycore-accent",
  success: "bg-paycore-successBg text-paycore-success",
  danger: "bg-paycore-dangerBg text-paycore-danger",
  muted: "bg-paycore-line2 text-paycore-muted",
};

// Dispute state machine: opened → under_review → won | lost | accepted.
const DISPUTE_STATUS: Record<string, { label: string; tone: Tone }> = {
  opened: { label: "เปิดเรื่อง", tone: "warn" },
  under_review: { label: "กำลังพิจารณา", tone: "accent" },
  won: { label: "ชนะข้อโต้แย้ง", tone: "success" },
  lost: { label: "แพ้ข้อโต้แย้ง", tone: "danger" },
  accepted: { label: "ยอมรับ", tone: "muted" },
};

function DisputePill({ status }: { status: string }) {
  const m = DISPUTE_STATUS[status] ?? { label: status, tone: "muted" as Tone };
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold ${TONE_CLASS[m.tone]}`}
    >
      <span className="h-1.5 w-1.5 rounded-full bg-current opacity-90" />
      {m.label}
    </span>
  );
}

function shortDate(iso: string): string {
  return new Intl.DateTimeFormat("th-TH", {
    day: "numeric",
    month: "short",
    year: "2-digit",
  }).format(new Date(iso));
}

export default async function DisputesPage() {
  const res = await serverGet("/disputes?limit=100");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`disputes failed: ${res.status}`);
  const rows: Dispute[] = (await res.json()).data ?? [];

  const currency = rows[0]?.currency ?? "THB";
  const open = rows.filter((r) => r.status === "opened" || r.status === "under_review");
  const openAmount = open.reduce((sum, r) => sum + (parseFloat(r.amount) || 0), 0);
  const won = rows.filter((r) => r.status === "won").length;
  const resolved = rows.filter(
    (r) => r.status === "won" || r.status === "lost" || r.status === "accepted",
  ).length;
  // Win rate over resolved disputes only (undecided ones don't count either way).
  const winRate = resolved > 0 ? Math.round((won / resolved) * 100) : null;

  return (
    <main className="min-h-screen bg-paycore-bg p-6 md:p-8">
      <div className="mx-auto max-w-5xl">
        <DashboardNav active="/disputes" />

        {/* Header */}
        <div className="mb-6">
          <h1 className="text-lg font-semibold text-paycore-text">
            ข้อโต้แย้ง (Disputes / Chargebacks)
          </h1>
          <p className="mt-0.5 text-xs text-paycore-muted">
            เมื่อผู้ถือบัตรโต้แย้งรายการผ่านธนาคาร ระบบจะเปิดเรื่องให้ที่นี่ — ตอบกลับพร้อมหลักฐาน
            ก่อนครบกำหนด
          </p>
        </div>

        {/* Summary tiles */}
        <section className="grid grid-cols-2 gap-4 md:grid-cols-3">
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              เปิดอยู่ / รอตอบกลับ
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {open.length.toLocaleString("th-TH")}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              มูลค่ารวม {formatDecimalMoney(openAmount, currency)}
            </p>
          </div>

          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              อัตราชนะ
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-success">
              {winRate === null ? "—" : `${winRate}%`}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              จาก {resolved.toLocaleString("th-TH")} เรื่องที่ตัดสินแล้ว
            </p>
          </div>

          <div className="col-span-2 rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card md:col-span-1">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              ทั้งหมด
            </p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {rows.length.toLocaleString("th-TH")}
            </p>
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              ข้อโต้แย้งที่บันทึกไว้
            </p>
          </div>
        </section>

        {/* Disputes table */}
        <section className="mt-6 rounded-xl2 border border-paycore-line bg-paycore-surface shadow-card">
          <div className="flex items-center justify-between px-5 py-4">
            <h2 className="text-sm font-semibold text-paycore-text">รายการข้อโต้แย้ง</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    วันที่เปิด
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    การชำระเงิน
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    เหตุผล
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    จำนวนเงิน
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    สถานะ
                  </th>
                </tr>
              </thead>
              <tbody>
                {rows.length === 0 && (
                  <tr>
                    <td colSpan={5} className="px-5 py-14 text-center">
                      <p className="text-sm font-medium text-paycore-text">
                        ยังไม่มีข้อโต้แย้ง 🎉
                      </p>
                      <p className="mx-auto mt-1 max-w-md text-xs text-paycore-muted">
                        เยี่ยมมาก — ยังไม่มีผู้ถือบัตรโต้แย้งรายการของคุณ เมื่อมีการเปิด chargeback
                        จากธนาคาร รายการจะปรากฏที่นี่พร้อมกำหนดเวลาตอบกลับ
                      </p>
                    </td>
                  </tr>
                )}
                {rows.map((r) => (
                  <tr key={r.id} className="border-b border-paycore-line2 last:border-b-0">
                    <td className="px-5 py-3">
                      <div className="font-medium text-paycore-text">{shortDate(r.opened_at)}</div>
                      <div className="font-mono text-[11px] text-paycore-muted">
                        {r.id.slice(0, 8)}…
                      </div>
                    </td>
                    <td className="px-5 py-3 font-mono text-[12px] text-paycore-text2">
                      {r.payment_id.slice(0, 12)}…
                    </td>
                    <td className="px-5 py-3 text-paycore-text2">
                      {r.reason ? (
                        <span className="line-clamp-1">{r.reason}</span>
                      ) : (
                        <span className="text-paycore-muted">—</span>
                      )}
                      {r.network_ref && (
                        <div className="font-mono text-[11px] text-paycore-muted">
                          ref {r.network_ref}
                        </div>
                      )}
                    </td>
                    <td className="px-5 py-3 text-right font-semibold tabular-nums text-paycore-text">
                      {formatDecimalMoney(r.amount, r.currency)}
                    </td>
                    <td className="px-5 py-3">
                      <DisputePill status={r.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        <p className="mt-4 text-xs text-paycore-muted">
          ข้อโต้แย้งที่ไม่ตอบกลับก่อนกำหนดจะถือว่ายอมรับโดยอัตโนมัติ · อัตราชนะคำนวณจากเรื่องที่ตัดสินแล้ว
          (ชนะ ÷ ชนะ+แพ้+ยอมรับ)
        </p>
      </div>
    </main>
  );
}
