"use client";

import { useEffect, useRef, useState } from "react";
import { formatMoney } from "@/lib/format";

declare global {
  interface Window {
    QRCode?: {
      new (el: HTMLElement, opts: { text: string; width: number; height: number; correctLevel: number }): unknown;
      CorrectLevel: { L: number; M: number; Q: number; H: number };
    };
  }
}

type CheckoutView = {
  id: string;
  status: string;
  amount_minor: number;
  currency: string;
  merchant_name: string;
  title: string;
  description?: string;
  image_url?: string;
  allowed_methods: string[];
  selected_method?: string;
  qr_payload?: string;
  next_action_url?: string;
  return_url?: string;
  expires_at: string;
  sandbox: boolean;
  session_token?: string;
  rail_instructions?: RailInstructions;
};

// On-chain payment instructions for a crypto/stablecoin rail (e.g. ThaiChain).
type RailInstructions = {
  asset: string;
  address: string;
  memo: string;
  chain_id: number;
  fee_sponsored: boolean;
  uri?: string;
  explorer_url?: string;
};

// Labels for every method the registry can surface (Phase 3 card/promptpay +
// Phase 4 wallet / redirect methods).
const METHOD_LABEL: Record<string, string> = {
  card: "บัตรเครดิต / เดบิต",
  promptpay: "PromptPay",
  mobile_banking: "Mobile Banking",
  truemoney: "TrueMoney Wallet",
  shopeepay: "ShopeePay",
  alipay: "Alipay",
  wechat: "WeChat Pay",
  card_installment: "ผ่อนชำระบัตร",
  thaichain: "คริปโต · Stablecoin (ThaiChain)",
};

// The six Phase 4 wallet / redirect methods share one mock flow.
const WALLET_METHODS = ["mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"];
const isWallet = (m: string) => WALLET_METHODS.includes(m);

// A small right-aligned brand mark for each method row, mirroring the landing
// checkout mockup (card-brand circles, coloured wordmarks).
function MethodMark({ method }: { method: string }) {
  if (method === "card") {
    return (
      <svg viewBox="0 0 40 24" width="34" height="20" aria-hidden="true">
        <circle cx="16" cy="12" r="9" fill="#EB001B" />
        <circle cx="24" cy="12" r="9" fill="#F79E1B" opacity=".85" />
      </svg>
    );
  }
  const WORDMARK: Record<string, { text: string; color: string }> = {
    promptpay: { text: "PromptPay", color: "#1a4b8f" },
    truemoney: { text: "TrueMoney", color: "#e8792b" },
    shopeepay: { text: "ShopeePay", color: "#ee4d2d" },
    alipay: { text: "Alipay", color: "#1677ff" },
    wechat: { text: "WeChat", color: "#09b83e" },
    mobile_banking: { text: "Bank", color: "#185fa5" },
    card_installment: { text: "ผ่อน", color: "#185fa5" },
    thaichain: { text: "Stablecoin", color: "#0f766e" },
  };
  const w = WORDMARK[method];
  if (!w) return null;
  return (
    <span className="font-extrabold text-[13px] tracking-tight" style={{ color: w.color }}>
      {w.text}
    </span>
  );
}

// Reusable outer shell so every checkout state (form, QR, wallet, terminal)
// shares the exact same centered white card + secure footer.
function CheckoutCard({ children, footer = true }: { children: React.ReactNode; footer?: boolean }) {
  return (
    <div className="w-full max-w-[440px] rounded-xl2 bg-paycore-surface border border-paycore-line shadow-cardlg p-6 sm:p-7">
      {children}
      {footer && <SecureFooter />}
    </div>
  );
}

// RailField renders a labelled on-chain instruction value (address / memo /
// amount) in a bordered box. highlight draws attention to the memo, which the
// payer MUST attach for the transfer to reconcile.
function RailField({
  label,
  value,
  mono = false,
  highlight = false,
}: {
  label: string;
  value: string;
  mono?: boolean;
  highlight?: boolean;
}) {
  return (
    <div
      className={`rounded-xl border px-3.5 py-2.5 ${
        highlight ? "border-paycore-warn/40 bg-paycore-warnBg" : "border-paycore-line bg-paycore-surface2"
      }`}
    >
      <div className="text-[10px] font-semibold uppercase tracking-wide text-paycore-muted">{label}</div>
      <div
        className={`mt-0.5 break-all text-[13px] text-paycore-text ${mono ? "font-mono" : "font-semibold tabular-nums"}`}
      >
        {value}
      </div>
    </div>
  );
}

function SecureFooter() {
  return (
    <div className="mt-5 flex items-center justify-center gap-1.5 text-[11px] text-paycore-muted">
      <LockIcon className="w-3 h-3" />
      <span>ชำระเงินอย่างปลอดภัยด้วย 3-D Secure · ข้อมูลบัตรผ่าน tokenization vault</span>
    </div>
  );
}

function LockIcon({ className = "w-4 h-4" }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <rect x="4" y="11" width="16" height="9" rx="2" />
      <path d="M8 11V8a4 4 0 0 1 8 0v3" />
    </svg>
  );
}

// Merchant header: gradient monogram + name + item/title, with a TEST badge in
// sandbox. Shared across the open + terminal states.
function MerchantHeader({ view }: { view: CheckoutView }) {
  const monogram = (view.merchant_name || "?").trim().charAt(0).toUpperCase();
  return (
    <div className="flex items-center gap-3">
      <div
        className="w-9 h-9 rounded-lg flex-none flex items-center justify-center text-white font-bold text-base"
        style={{ background: "linear-gradient(135deg,#f0a63c,#e07c2b)" }}
        aria-hidden="true"
      >
        {monogram}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-sm font-semibold text-paycore-text truncate">{view.merchant_name}</div>
        {view.title && <div className="text-[11px] text-paycore-muted truncate">{view.title}</div>}
      </div>
      {view.sandbox && (
        <span className="flex-none rounded-full bg-paycore-warnBg text-paycore-warn text-[10px] font-bold tracking-wide px-2.5 py-1 uppercase">
          Test Mode
        </span>
      )}
    </div>
  );
}

function AmountBlock({ view }: { view: CheckoutView }) {
  return (
    <div className="text-center my-6">
      <div className="text-xs text-paycore-muted">ยอดที่ต้องชำระ</div>
      <div className="text-[38px] leading-tight font-semibold tracking-tight tabular-nums text-paycore-text">
        {formatMoney(view.amount_minor, view.currency)}
      </div>
      {view.description && <p className="text-paycore-muted text-sm mt-1">{view.description}</p>}
    </div>
  );
}

export default function CheckoutClient({ publicId }: { publicId: string }) {
  const [view, setView] = useState<CheckoutView | null>(null);
  const [token, setToken] = useState("");
  const [method, setMethod] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  // Card form state (sandbox only).
  const [cardNumber, setCardNumber] = useState("");
  const [expMonth, setExpMonth] = useState("");
  const [expYear, setExpYear] = useState("");
  const [cvv, setCvv] = useState("");

  const createdRef = useRef(false);

  // Create the session once on mount.
  useEffect(() => {
    if (createdRef.current) return;
    createdRef.current = true;
    (async () => {
      const res = await fetch("/api/checkout/sessions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ link: publicId }),
      });
      const env = await res.json().catch(() => null);
      if (!res.ok) {
        setErr(env?.message ?? "ไม่พบลิงก์ชำระเงินนี้");
        return;
      }
      const v: CheckoutView = env.data;
      setView(v);
      setToken(v.session_token ?? "");
      // Preselect the first method so the card form (if card) shows by default,
      // matching the landing mockup.
      if (v.allowed_methods.length >= 1) setMethod(v.allowed_methods[0]);
    })();
  }, [publicId]);

  async function payCard(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        method: "card",
        card: {
          number: cardNumber.replace(/\s+/g, ""),
          exp_month: Number(expMonth),
          exp_year: Number(expYear),
          cvv,
        },
      }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "ชำระเงินไม่สำเร็จ");
      return;
    }
    const v: CheckoutView = env.data;
    setView(v);
    if (v.next_action_url) window.location.href = v.next_action_url; // 3DS redirect
  }

  if (err && !view) {
    return (
      <CheckoutCard footer={false}>
        <div className="flex items-start gap-3 text-paycore-text2 text-sm">
          <span className="text-lg leading-none">⚠️</span>
          <p>{err}</p>
        </div>
      </CheckoutCard>
    );
  }
  if (!view) {
    return (
      <CheckoutCard footer={false}>
        <div className="flex items-center justify-center gap-2 py-8 text-paycore-muted text-sm">
          <span className="w-4 h-4 rounded-full border-2 border-paycore-line border-t-paycore-primary animate-spin" />
          กำลังโหลด…
        </div>
      </CheckoutCard>
    );
  }

  // Any non-open session (requires_action for promptpay QR or a wallet mock, or a
  // terminal state) is driven by the status view.
  if (view.status !== "open") {
    return <CheckoutStatusView token={token} initial={view} />;
  }

  return (
    <CheckoutCard>
      <MerchantHeader view={view} />
      <AmountBlock view={view} />

      {/* Method selector — full-width radio rows. */}
      <div role="radiogroup" aria-label="เลือกวิธีชำระเงิน" className="flex flex-col gap-2">
        {view.allowed_methods.map((m) => {
          const selected = method === m;
          return (
            <button
              key={m}
              type="button"
              role="radio"
              aria-checked={selected}
              onClick={() => setMethod(m)}
              className={`flex items-center gap-3 rounded-xl border px-4 py-3 text-left transition focus:outline-none focus-visible:ring-2 focus-visible:ring-paycore-primary/60 ${
                selected
                  ? "border-paycore-primary bg-[#fbfdff] ring-2 ring-paycore-accentBg"
                  : "border-paycore-line bg-paycore-surface hover:border-[#c9d3df]"
              }`}
            >
              <span
                className={`w-[18px] h-[18px] rounded-full border-2 flex-none grid place-items-center ${
                  selected ? "border-paycore-primary" : "border-paycore-line"
                }`}
              >
                {selected && <span className="w-2 h-2 rounded-full bg-paycore-primary" />}
              </span>
              <span className="flex-1 text-[13px] font-semibold text-paycore-text">{METHOD_LABEL[m] ?? m}</span>
              <span className="flex-none flex items-center">
                <MethodMark method={m} />
              </span>
            </button>
          );
        })}
      </div>

      {/* Per-method action area. */}
      <div className="mt-5">
        {method === "promptpay" && (
          <PayMethodButton token={token} method="promptpay" label="สร้าง QR PromptPay" busyLabel="กำลังสร้าง QR…" onDone={setView} setErr={setErr} />
        )}

        {isWallet(method) && (
          <>
            {!view.sandbox && (
              <p className="text-xs rounded-lg bg-paycore-warnBg text-paycore-warn px-3 py-2">
                ช่องทางนี้ยังไม่พร้อมใช้งานบนระบบนี้
              </p>
            )}
            {view.sandbox && (
              <PayMethodButton
                token={token}
                method={method}
                label={`ดำเนินการต่อด้วย ${METHOD_LABEL[method] ?? method}`}
                busyLabel="กำลังดำเนินการ…"
                onDone={setView}
                setErr={setErr}
              />
            )}
          </>
        )}

        {method === "thaichain" && (
          <>
            {!view.sandbox && (
              <p className="text-xs rounded-lg bg-paycore-warnBg text-paycore-warn px-3 py-2">
                ช่องทางนี้ยังไม่พร้อมใช้งานบนระบบนี้
              </p>
            )}
            {view.sandbox && (
              <PayMethodButton
                token={token}
                method="thaichain"
                label="จ่ายด้วย Stablecoin (ThaiChain)"
                busyLabel="กำลังสร้างคำสั่งชำระ…"
                onDone={setView}
                setErr={setErr}
              />
            )}
          </>
        )}

        {method === "card" && view.sandbox && (
          <form onSubmit={payCard} className="space-y-3">
            <p className="flex items-start gap-2 text-xs rounded-lg bg-paycore-warnBg text-paycore-warn px-3 py-2">
              <span className="leading-none">🔧</span>
              <span>โหมดทดสอบ (Sandbox) — ใช้บัตรทดสอบเท่านั้น เช่น 4111 1111 1111 1111</span>
            </p>
            <div>
              <label className="block text-[11px] font-medium text-paycore-muted mb-1">หมายเลขบัตร</label>
              <input
                value={cardNumber}
                onChange={(e) => setCardNumber(e.target.value)}
                inputMode="numeric"
                placeholder="4242 4242 4242 4242"
                className="w-full rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2.5 text-sm tabular-nums placeholder:text-paycore-muted/70 focus:outline-none focus:border-paycore-primary focus:ring-2 focus:ring-paycore-accentBg"
              />
            </div>
            <div className="flex gap-3">
              <div className="flex-1">
                <label className="block text-[11px] font-medium text-paycore-muted mb-1">เดือน / ปี</label>
                <div className="flex gap-2">
                  <input
                    value={expMonth}
                    onChange={(e) => setExpMonth(e.target.value)}
                    inputMode="numeric"
                    placeholder="MM"
                    className="w-1/2 rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2.5 text-sm tabular-nums placeholder:text-paycore-muted/70 focus:outline-none focus:border-paycore-primary focus:ring-2 focus:ring-paycore-accentBg"
                  />
                  <input
                    value={expYear}
                    onChange={(e) => setExpYear(e.target.value)}
                    inputMode="numeric"
                    placeholder="YYYY"
                    className="w-1/2 rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2.5 text-sm tabular-nums placeholder:text-paycore-muted/70 focus:outline-none focus:border-paycore-primary focus:ring-2 focus:ring-paycore-accentBg"
                  />
                </div>
              </div>
              <div className="w-[92px]">
                <label className="block text-[11px] font-medium text-paycore-muted mb-1">CVC</label>
                <input
                  value={cvv}
                  onChange={(e) => setCvv(e.target.value)}
                  inputMode="numeric"
                  placeholder="•••"
                  className="w-full rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2.5 text-sm tabular-nums placeholder:text-paycore-muted/70 focus:outline-none focus:border-paycore-primary focus:ring-2 focus:ring-paycore-accentBg"
                />
              </div>
            </div>
            {err && <p className="text-paycore-danger text-sm">{err}</p>}
            <button
              disabled={busy}
              className="w-full flex items-center justify-center gap-2 rounded-xl bg-paycore-primary hover:bg-paycore-primaryHover text-white font-semibold text-[15px] h-12 transition disabled:opacity-60"
            >
              {busy ? (
                "กำลังชำระเงิน…"
              ) : (
                <>
                  <LockIcon className="w-[17px] h-[17px]" />
                  ชำระเงิน {formatMoney(view.amount_minor, view.currency)}
                </>
              )}
            </button>
          </form>
        )}

        {method === "card" && !view.sandbox && (
          <p className="text-sm text-paycore-muted">การชำระด้วยบัตรยังไม่พร้อมใช้งานบนระบบนี้</p>
        )}
      </div>
    </CheckoutCard>
  );
}

// waitForQRCode resolves once the vanilla qrcode.min.js global is available
// (loaded via <Script strategy="afterInteractive">). Gives up after ~5s.
function waitForQRCode(): Promise<NonNullable<Window["QRCode"]>> {
  return new Promise((resolve, reject) => {
    const started = Date.now();
    const tick = () => {
      if (typeof window !== "undefined" && window.QRCode) return resolve(window.QRCode);
      if (Date.now() - started > 5000) return reject(new Error("QR library not loaded"));
      setTimeout(tick, 50);
    };
    tick();
  });
}

// PayMethodButton POSTs /pay for a data-less method (promptpay or a wallet slug),
// then hands the returned requires_action view to the status view via onDone.
function PayMethodButton({
  token, method, label, busyLabel, onDone, setErr,
}: {
  token: string; method: string; label: string; busyLabel: string;
  onDone: (v: CheckoutView) => void; setErr: (s: string) => void;
}) {
  const [busy, setBusy] = useState(false);
  async function start() {
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ method }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "ดำเนินการไม่สำเร็จ");
      return;
    }
    onDone(env.data as CheckoutView);
  }
  return (
    <button
      onClick={start}
      disabled={busy}
      className="w-full flex items-center justify-center gap-2 rounded-xl bg-paycore-primary hover:bg-paycore-primaryHover text-white font-semibold text-[15px] h-12 transition disabled:opacity-60"
    >
      {busy ? busyLabel : label}
    </button>
  );
}

// CheckoutStatusView renders the PromptPay QR or the wallet mock approve/decline
// panel while awaiting action, polls session status until a terminal state, then
// shows success + optional return_url.
function CheckoutStatusView({ token, initial }: { token: string; initial: CheckoutView }) {
  const [view, setView] = useState<CheckoutView>(initial);
  const qrBox = useRef<HTMLDivElement>(null);

  const walletAwaiting =
    view.status === "requires_action" && !!view.selected_method && WALLET_METHODS.includes(view.selected_method);

  const railAwaiting =
    view.status === "requires_action" && view.selected_method === "thaichain" && !!view.rail_instructions;

  // Render the QR whenever a PromptPay payload is present and not yet paid.
  useEffect(() => {
    if (!view.qr_payload || view.status === "paid") return;
    let cancelled = false;
    (async () => {
      try {
        const QR = await waitForQRCode();
        if (cancelled || !qrBox.current) return;
        qrBox.current.innerHTML = "";
        new QR(qrBox.current, { text: view.qr_payload!, width: 220, height: 220, correctLevel: QR.CorrectLevel.M });
      } catch {
        /* leave the payload text visible as a fallback */
      }
    })();
    return () => { cancelled = true; };
  }, [view.qr_payload, view.status]);

  // Poll status until terminal.
  useEffect(() => {
    if (view.status === "paid" || view.status === "expired" || view.status === "failed") return;
    const id = setInterval(async () => {
      const res = await fetch(`/api/checkout/sessions/${token}`, { cache: "no-store" });
      if (!res.ok) return;
      const env = await res.json();
      setView(env.data as CheckoutView);
    }, 3000);
    return () => clearInterval(id);
  }, [token, view.status]);

  // Mock wallet approve/decline (sandbox only) → flips the session server-side.
  async function confirmMock(approve: boolean) {
    const res = await fetch(`/api/checkout/sessions/${token}/confirm-mock`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ approve }),
    });
    if (!res.ok) return;
    const env = await res.json();
    setView(env.data as CheckoutView);
  }

  if (view.status === "paid") {
    return (
      <CheckoutCard footer={false}>
        <div className="text-center py-2">
          <div className="mx-auto w-14 h-14 rounded-full bg-paycore-successBg text-paycore-success grid place-items-center">
            <svg viewBox="0 0 24 24" className="w-7 h-7" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M20 6 9 17l-5-5" />
            </svg>
          </div>
          <h1 className="text-xl font-semibold mt-4 text-paycore-text">ชำระเงินสำเร็จ</h1>
          <p className="text-[26px] font-semibold tabular-nums mt-1 text-paycore-text">{formatMoney(view.amount_minor, view.currency)}</p>
          <p className="text-paycore-muted text-sm mt-1">{view.merchant_name}</p>
          {view.return_url && (
            <a
              href={view.return_url}
              className="mt-5 inline-flex w-full items-center justify-center rounded-xl bg-paycore-primary hover:bg-paycore-primaryHover text-white font-semibold h-12 transition"
            >
              กลับไปที่ร้านค้า
            </a>
          )}
        </div>
      </CheckoutCard>
    );
  }
  if (view.status === "expired" || view.status === "failed") {
    return (
      <CheckoutCard footer={false}>
        <div className="text-center py-2">
          <div className="mx-auto w-14 h-14 rounded-full bg-paycore-dangerBg text-paycore-danger grid place-items-center">
            <svg viewBox="0 0 24 24" className="w-7 h-7" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M18 6 6 18M6 6l12 12" />
            </svg>
          </div>
          <h1 className="text-lg font-semibold mt-4 text-paycore-text">
            {view.status === "expired" ? "หมดเวลาชำระเงิน" : "ชำระเงินไม่สำเร็จ"}
          </h1>
          <p className="text-paycore-muted text-sm mt-1">โปรดลองอีกครั้ง</p>
        </div>
      </CheckoutCard>
    );
  }

  // Wallet mock: simulate the PSP approve/decline screen (sandbox only).
  if (walletAwaiting) {
    return (
      <CheckoutCard>
        <MerchantHeader view={view} />
        <AmountBlock view={view} />
        <p className="text-center text-sm font-semibold text-paycore-text -mt-3 mb-4">
          {METHOD_LABEL[view.selected_method ?? ""] ?? view.selected_method}
        </p>
        <p className="flex items-start gap-2 text-xs rounded-lg bg-paycore-warnBg text-paycore-warn px-3 py-2 mb-4">
          <span className="leading-none">🔧</span>
          <span>โหมดทดสอบ (Sandbox) — จำลองหน้าอนุมัติของผู้ให้บริการ</span>
        </p>
        <div className="flex gap-2">
          <button
            onClick={() => confirmMock(true)}
            className="flex-1 rounded-xl bg-paycore-primary hover:bg-paycore-primaryHover text-white font-semibold h-12 transition"
          >
            อนุมัติการชำระเงิน
          </button>
          <button
            onClick={() => confirmMock(false)}
            className="flex-1 rounded-xl border border-paycore-line text-paycore-text2 hover:bg-paycore-surface2 font-medium h-12 transition"
          >
            ปฏิเสธ
          </button>
        </div>
      </CheckoutCard>
    );
  }

  // Crypto/stablecoin rail awaiting: show the deposit address + memo + finality.
  if (railAwaiting && view.rail_instructions) {
    const ri = view.rail_instructions;
    const assetAmount = (view.amount_minor / 100).toLocaleString("en-US", {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    });
    return (
      <CheckoutCard>
        <MerchantHeader view={view} />
        <AmountBlock view={view} />
        <p className="text-center text-sm font-semibold text-paycore-text -mt-3 mb-3">
          ชำระด้วย {ri.asset} บน ThaiChain
        </p>
        <div className="flex items-center justify-center gap-2 mb-4">
          <span className="rounded-full bg-paycore-accentBg text-paycore-accent text-[11px] font-semibold px-2.5 py-1">
            Chain ID {ri.chain_id}
          </span>
          {ri.fee_sponsored && (
            <span className="rounded-full bg-paycore-successBg text-paycore-success text-[11px] font-semibold px-2.5 py-1">
              ค่าแก๊สฟรี · sponsored
            </span>
          )}
        </div>
        <div className="space-y-3">
          <RailField label={`ส่ง ${ri.asset} จำนวน`} value={`${assetAmount} ${ri.asset}`} />
          <RailField label="ที่อยู่ผู้รับ (Address)" value={ri.address} mono />
          <RailField label="Memo — ต้องแนบทุกครั้ง (ใช้ยืนยันออเดอร์)" value={ri.memo} mono highlight />
        </div>
        <p className="text-[11px] text-paycore-muted mt-3 leading-relaxed">
          โอน {ri.asset} ไปยังที่อยู่ด้านบนพร้อมแนบ memo — ระบบยืนยันอัตโนมัติเมื่อธุรกรรม on-chain
          ครบตามจำนวน confirmations
          {ri.explorer_url && (
            <>
              {" · "}
              <a href={ri.explorer_url} target="_blank" rel="noopener noreferrer" className="text-paycore-accent hover:underline">
                ดูบน explorer
              </a>
            </>
          )}
        </p>
        {view.sandbox && (
          <div className="mt-5">
            <p className="flex items-start gap-2 text-xs rounded-lg bg-paycore-warnBg text-paycore-warn px-3 py-2 mb-3">
              <span className="leading-none">🔧</span>
              <span>โหมดทดสอบ (Sandbox) — จำลองการโอน on-chain แทนการส่งจริง</span>
            </p>
            <div className="flex gap-2">
              <button
                onClick={() => confirmMock(true)}
                className="flex-1 rounded-xl bg-paycore-primary hover:bg-paycore-primaryHover text-white font-semibold h-12 transition"
              >
                จำลองว่าจ่ายแล้ว (ยืนยัน on-chain)
              </button>
              <button
                onClick={() => confirmMock(false)}
                className="flex-none px-4 rounded-xl border border-paycore-line text-paycore-text2 hover:bg-paycore-surface2 font-medium h-12 transition"
              >
                ยกเลิก
              </button>
            </div>
          </div>
        )}
        <p className="flex items-center justify-center gap-2 text-paycore-muted text-xs mt-4">
          <span className="w-3 h-3 rounded-full border-2 border-paycore-line border-t-paycore-primary animate-spin" />
          รอการยืนยันบนเชน…
        </p>
      </CheckoutCard>
    );
  }

  // PromptPay awaiting: show QR + amount.
  return (
    <CheckoutCard>
      <MerchantHeader view={view} />
      <AmountBlock view={view} />
      <p className="text-center text-sm font-semibold text-paycore-text -mt-3 mb-4">สแกนเพื่อชำระด้วย PromptPay</p>
      <div className="flex justify-center">
        <div ref={qrBox} className="bg-white p-3 rounded-xl border border-paycore-line" style={{ minHeight: 220, minWidth: 220 }} />
      </div>
      <p className="flex items-center justify-center gap-2 text-paycore-muted text-xs mt-4">
        <span className="w-3 h-3 rounded-full border-2 border-paycore-line border-t-paycore-primary animate-spin" />
        รอการยืนยันการชำระเงิน…
      </p>
    </CheckoutCard>
  );
}
