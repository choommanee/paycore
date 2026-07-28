import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney, formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

type AuthMe = { email: string; name: string; merchant_name: string; merchant_id: string };

type Stats = {
  count: number;
  volumeMinor: number;
  byStatus: { authorized: number; captured: number; refunded: number; failed: number };
  successRate: number;
  refundRatio: number;
};

type SeriesPoint = { date: string; volume_minor: number; count: number };
type Series = {
  days: number;
  series: SeriesPoint[];
  totals: { volume_minor: number; count: number };
  trend: { volume_pct: number; count_pct: number };
};

type Payment = {
  id: string;
  amount: string;
  status: string;
  card_brand?: string;
  card_last4?: string;
  created_at: string;
};

function pct(fraction: number): string {
  return `${(fraction * 100).toFixed(1)}%`;
}

// --- sparkline geometry -----------------------------------------------------
// Builds a stroke path + an area path from a numeric series, normalized into a
// 100x28 mini viewport. Straight segments with round joins (matches the mockup).
function spark(values: number[], w = 100, h = 28, pad = 4) {
  const vals = values.length >= 2 ? values : values.length === 1 ? [values[0], values[0]] : [0, 0];
  const min = Math.min(...vals);
  const max = Math.max(...vals);
  const range = max - min || 1;
  const n = vals.length;
  const pts = vals.map((v, i) => {
    const x = (i / (n - 1)) * w;
    const y = pad + (1 - (v - min) / range) * (h - 2 * pad);
    return [Number(x.toFixed(2)), Number(y.toFixed(2))] as const;
  });
  const line = pts.map((p, i) => `${i === 0 ? "M" : "L"}${p[0]},${p[1]}`).join(" ");
  const area = `${line} L${w},${h} L0,${h} Z`;
  return { line, area, last: pts[pts.length - 1] };
}

function Sparkline({
  values,
  color,
  id,
}: {
  values: number[];
  color: string;
  id: string;
}) {
  const { line, area, last } = spark(values);
  return (
    <svg
      viewBox="0 0 100 28"
      preserveAspectRatio="none"
      aria-hidden="true"
      className="mt-2.5 block h-7 w-full overflow-visible"
    >
      <defs>
        <linearGradient id={id} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor={color} stopOpacity="0.22" />
          <stop offset="1" stopColor={color} stopOpacity="0" />
        </linearGradient>
      </defs>
      <path d={area} fill={`url(#${id})`} />
      <path
        d={line}
        fill="none"
        stroke={color}
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
        vectorEffect="non-scaling-stroke"
      />
      <circle cx={last[0]} cy={last[1]} r={2.6} fill={color} />
    </svg>
  );
}

// A small horizontal meter for the rate tiles (honest — we have no per-day
// success/refund breakdown, so we show the current level rather than a fake line).
function Meter({ value, color }: { value: number; color: string }) {
  const w = Math.max(2, Math.min(100, value * 100));
  return (
    <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-paycore-line2" aria-hidden="true">
      <div className="h-full rounded-full" style={{ width: `${w}%`, backgroundColor: color }} />
    </div>
  );
}

function TrendChip({ pctValue, invert = false }: { pctValue: number; invert?: boolean }) {
  const rounded = Math.round(pctValue * 1000) / 10; // one decimal
  if (rounded === 0) {
    return <span className="text-[11px] font-semibold text-paycore-muted">— 0.0% เทียบ 30 วันก่อน</span>;
  }
  const up = rounded > 0;
  // For most metrics up = good (green). For refund-style metrics, up = bad (invert).
  const good = invert ? !up : up;
  const color = good ? "#0f6e56" : "#a32d2d";
  const arrow = up ? "▲" : "▼";
  return (
    <span className="text-[11px] font-semibold" style={{ color }}>
      {arrow} {Math.abs(rounded).toFixed(1)}%
      <span className="ml-1 font-normal text-paycore-muted">เทียบ 30 วันก่อน</span>
    </span>
  );
}

// --- payment presentation ---------------------------------------------------
const STATUS_META: Record<string, { label: string; bg: string; fg: string }> = {
  captured: { label: "Captured", bg: "#e1f5ee", fg: "#0f6e56" },
  authorized: { label: "Authorized", bg: "#e6f1fb", fg: "#0c447c" },
  requires_action: { label: "Pending", bg: "#fdf3e1", fg: "#8a5a0b" },
  refunded: { label: "Refunded", bg: "#f3e2f6", fg: "#7a2a86" },
  partial_refunded: { label: "Partial refund", bg: "#f3e2f6", fg: "#7a2a86" },
  failed: { label: "Failed", bg: "#fcebeb", fg: "#a32d2d" },
  voided: { label: "Voided", bg: "#fcebeb", fg: "#a32d2d" },
};

function statusMeta(status: string) {
  return STATUS_META[status] ?? { label: status, bg: "#eef1f5", fg: "#374151" };
}

function methodLabel(p: Payment): string {
  if (p.card_brand && p.card_last4) {
    const brand = p.card_brand.charAt(0).toUpperCase() + p.card_brand.slice(1);
    return `${brand} •••• ${p.card_last4}`;
  }
  if (p.card_last4) return `บัตร •••• ${p.card_last4}`;
  return "บัตร";
}

function shortDate(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const sameDay = d.toDateString() === now.toDateString();
  if (sameDay) {
    return `วันนี้ ${new Intl.DateTimeFormat("th-TH", { hour: "2-digit", minute: "2-digit" }).format(d)}`;
  }
  const diffDays = Math.floor((now.getTime() - d.getTime()) / 86_400_000);
  if (diffDays >= 1 && diffDays <= 6) return `${diffDays} วันก่อน`;
  return new Intl.DateTimeFormat("th-TH", { day: "numeric", month: "short" }).format(d);
}

export default async function DashboardHome() {
  const [meRes, statsRes, seriesRes, paymentsRes] = await Promise.all([
    serverGet("/auth/me"),
    serverGet("/stats"),
    serverGet("/stats/series?days=30"),
    serverGet("/payments?limit=6"),
  ]);
  if (meRes.status === 401 || statsRes.status === 401) redirect("/login");
  if (!meRes.ok) throw new Error(`auth/me failed: ${meRes.status}`);
  if (!statsRes.ok) throw new Error(`stats failed: ${statsRes.status}`);
  const me: AuthMe = (await meRes.json()).data;
  const s: Stats = (await statsRes.json()).data;
  const series: Series | null = seriesRes.ok ? (await seriesRes.json()).data : null;
  const payments: Payment[] = paymentsRes.ok ? ((await paymentsRes.json()).data ?? []) : [];

  const volumeSeries = series?.series.map((p) => p.volume_minor) ?? [];
  const countSeries = series?.series.map((p) => p.count) ?? [];
  const volumePct = series?.trend.volume_pct ?? 0;
  const countPct = series?.trend.count_pct ?? 0;
  const volUp = volumePct >= 0;
  const cntUp = countPct >= 0;

  return (
    <main className="min-h-screen bg-paycore-bg p-6 md:p-8">
      <div className="mx-auto max-w-5xl">
        <DashboardNav active="/dashboard" />

        {/* Page header */}
        <div className="mb-6 flex items-center gap-3">
          <div>
            <h1 className="text-lg font-semibold text-paycore-text">ภาพรวมร้านค้า</h1>
            <p className="text-xs text-paycore-muted">
              {me.merchant_name} · 30 วันล่าสุด · สวัสดี {me.name || me.email}
            </p>
          </div>
          <span
            className="ml-auto inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold"
            style={{ backgroundColor: "#e1f5ee", color: "#0f6e56" }}
          >
            <span className="h-1.5 w-1.5 rounded-full bg-current" />
            Live
          </span>
        </div>

        {/* Stat tiles */}
        <section className="grid grid-cols-2 gap-4 md:grid-cols-4">
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">ยอดรับชำระ (30 วัน)</p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {formatMoney(s.volumeMinor)}
            </p>
            <Sparkline values={volumeSeries} color={volUp ? "#185fa5" : "#a32d2d"} id="spk-vol" />
            <div className="mt-1.5">
              <TrendChip pctValue={volumePct} />
            </div>
          </div>

          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">จำนวนธุรกรรม</p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {s.count.toLocaleString("th-TH")}
            </p>
            <Sparkline values={countSeries} color={cntUp ? "#0f6e56" : "#a32d2d"} id="spk-cnt" />
            <div className="mt-1.5">
              <TrendChip pctValue={countPct} />
            </div>
          </div>

          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">อัตราสำเร็จ</p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {pct(s.successRate)}
            </p>
            <Meter value={s.successRate} color="#0f6e56" />
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              {s.byStatus.captured}/{s.count} รายการสำเร็จ
            </p>
          </div>

          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">อัตราคืนเงิน</p>
            <p className="mt-1 text-xl font-semibold tabular-nums tracking-tight text-paycore-text">
              {pct(s.refundRatio)}
            </p>
            <Meter value={s.refundRatio} color="#8a5a0b" />
            <p className="mt-2 text-[11px] font-semibold text-paycore-muted">
              {s.byStatus.refunded} รายการคืนเงิน
            </p>
          </div>
        </section>

        {/* Recent transactions */}
        <section className="mt-6 rounded-xl2 border border-paycore-line bg-paycore-surface shadow-card">
          <div className="flex items-center justify-between px-5 py-4">
            <h2 className="text-sm font-semibold text-paycore-text">ธุรกรรมล่าสุด</h2>
            <a href="/transactions" className="text-xs font-medium text-paycore-primary hover:underline">
              ดูทั้งหมด →
            </a>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    ธุรกรรม
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    ช่องทาง
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    สถานะ
                  </th>
                  <th className="border-y border-paycore-line bg-paycore-surface2 px-5 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    ยอด
                  </th>
                </tr>
              </thead>
              <tbody>
                {payments.length === 0 && (
                  <tr>
                    <td colSpan={4} className="px-5 py-8 text-center text-sm text-paycore-muted">
                      ยังไม่มีธุรกรรม
                    </td>
                  </tr>
                )}
                {payments.map((p) => {
                  const meta = statusMeta(p.status);
                  return (
                    <tr key={p.id} className="border-b border-paycore-line2 last:border-b-0">
                      <td className="px-5 py-3">
                        <div className="font-mono text-xs text-paycore-text">{p.id.slice(0, 8)}…</div>
                        <div className="text-[11px] text-paycore-muted">{shortDate(p.created_at)}</div>
                      </td>
                      <td className="px-5 py-3 text-paycore-text2">{methodLabel(p)}</td>
                      <td className="px-5 py-3">
                        <span
                          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold"
                          style={{ backgroundColor: meta.bg, color: meta.fg }}
                        >
                          <span className="h-1.5 w-1.5 rounded-full bg-current opacity-90" />
                          {meta.label}
                        </span>
                      </td>
                      <td className="px-5 py-3 text-right font-semibold tabular-nums text-paycore-text">
                        {formatDecimalMoney(p.amount)}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </section>

        <p className="mt-4 text-xs text-paycore-muted">
          ตัวเลขและรายการคำนวณจากการชำระด้วยบัตร (card) ในรอบ 30 วันล่าสุด
        </p>

        <div className="mt-4 flex gap-4">
          <a href="/transactions" className="text-sm text-paycore-primary hover:underline">
            ดูธุรกรรมทั้งหมด →
          </a>
          <a href="/links" className="text-sm text-paycore-primary hover:underline">
            จัดการลิงก์ชำระเงิน →
          </a>
        </div>
      </div>
    </main>
  );
}
