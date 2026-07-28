// Shared presentation for the unified transactions feed (card + PromptPay +
// e-wallet). Kept in one place so the /transactions list and the dashboard's
// "recent transactions" table render identical method chips and status pills.

export type Transaction = {
  id: string;
  source: "card" | "promptpay" | "wallet" | string;
  method: string;
  amount_minor: number;
  currency: string;
  status: string;
  reference?: string;
  created_at: string;
};

// --- method chip ------------------------------------------------------------
// Tasteful colored chips echoing the landing-page method badges. `bg`/`fg` are
// literal hex (brand colors live outside the paycore palette) applied via inline
// style so they survive Tailwind's purge.
type MethodMeta = { label: string; bg: string; fg: string; card?: boolean };

const WALLET_METHODS: Record<string, MethodMeta> = {
  truemoney: { label: "TrueMoney", bg: "#fff2e6", fg: "#c85a12" },
  shopeepay: { label: "ShopeePay", bg: "#fff0e9", fg: "#d23f1f" },
  alipay: { label: "Alipay", bg: "#e8f1ff", fg: "#1155cc" },
  wechat: { label: "WeChat Pay", bg: "#e7f7ec", fg: "#0a8f31" },
  wechatpay: { label: "WeChat Pay", bg: "#e7f7ec", fg: "#0a8f31" },
  mobile_banking: { label: "Mobile Banking", bg: "#eef1f5", fg: "#374151" },
  card_installment: { label: "ผ่อนบัตร", bg: "#eef1f5", fg: "#374151", card: true },
};

function methodMeta(t: Transaction): MethodMeta {
  if (t.source === "card" || t.method === "card") {
    return { label: "บัตร", bg: "#eef1f5", fg: "#374151", card: true };
  }
  if (t.source === "promptpay" || t.method.startsWith("promptpay")) {
    return { label: "PromptPay", bg: "#eaf1fb", fg: "#1a4b8f" };
  }
  const w = WALLET_METHODS[t.method];
  if (w) return w;
  // Unknown wallet/method: honest neutral fallback using the raw method name.
  return { label: t.method || t.source || "อื่น ๆ", bg: "#eef1f5", fg: "#374151" };
}

function CardMark({ color }: { color: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      width="13"
      height="13"
      aria-hidden="true"
      fill="none"
      stroke={color}
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="shrink-0"
    >
      <rect x="2.5" y="5" width="19" height="14" rx="2.5" />
      <path d="M2.5 9.5h19" />
    </svg>
  );
}

export function MethodChip({ tx }: { tx: Transaction }) {
  const m = methodMeta(tx);
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold"
      style={{ backgroundColor: m.bg, color: m.fg }}
    >
      {m.card && <CardMark color={m.fg} />}
      {m.label}
    </span>
  );
}

// --- status pill ------------------------------------------------------------
// Broad status set → semantic paycore tokens. tone drives the Tailwind classes.
type Tone = "success" | "warn" | "danger" | "accent" | "muted";

const TONE_CLASS: Record<Tone, string> = {
  success: "bg-paycore-successBg text-paycore-success",
  warn: "bg-paycore-warnBg text-paycore-warn",
  danger: "bg-paycore-dangerBg text-paycore-danger",
  accent: "bg-paycore-accentBg text-paycore-accentInk",
  muted: "bg-paycore-line2 text-paycore-muted",
};

const STATUS_META: Record<string, { label: string; tone: Tone }> = {
  captured: { label: "เรียกเก็บแล้ว", tone: "success" },
  paid: { label: "สำเร็จ", tone: "success" },
  authorized: { label: "อนุมัติแล้ว", tone: "accent" },
  partial_refunded: { label: "คืนบางส่วน", tone: "warn" },
  refunded: { label: "คืนเงิน", tone: "warn" },
  failed: { label: "ล้มเหลว", tone: "danger" },
  requires_action: { label: "รอดำเนินการ", tone: "warn" },
  awaiting_payment: { label: "รอชำระเงิน", tone: "warn" },
  processing: { label: "กำลังดำเนินการ", tone: "warn" },
  expired: { label: "หมดอายุ", tone: "muted" },
  voided: { label: "ยกเลิก", tone: "muted" },
};

function statusMeta(status: string) {
  return STATUS_META[status] ?? { label: status, tone: "muted" as Tone };
}

export function StatusPill({ status, dot = false }: { status: string; dot?: boolean }) {
  const m = statusMeta(status);
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold ${TONE_CLASS[m.tone]}`}
    >
      {dot && <span className="h-1.5 w-1.5 rounded-full bg-current opacity-90" />}
      {m.label}
    </span>
  );
}

// --- date -------------------------------------------------------------------
// Short, relative-ish Thai date for compact rows.
export function shortDate(iso: string): string {
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
