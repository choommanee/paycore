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
};

const METHOD_LABEL: Record<string, string> = {
  card: "บัตรเครดิต/เดบิต",
  promptpay: "PromptPay",
};

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
      if (v.allowed_methods.length === 1) setMethod(v.allowed_methods[0]);
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
    return <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10">{err}</div>;
  }
  if (!view) {
    return <div className="text-paycore-muted mt-10">กำลังโหลด…</div>;
  }

  // Terminal / QR states are rendered by the PromptPay + status component (Task 8).
  if (view.status === "paid" || view.status === "expired" || view.status === "failed" || view.selected_method === "promptpay") {
    return <CheckoutStatusView token={token} initial={view} />;
  }

  return (
    <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10 space-y-5">
      <header>
        <p className="text-paycore-muted text-sm">{view.merchant_name}</p>
        <h1 className="text-xl font-semibold">{view.title}</h1>
        <p className="text-2xl font-bold mt-1">{formatMoney(view.amount_minor, view.currency)}</p>
        {view.description && <p className="text-paycore-muted text-sm mt-1">{view.description}</p>}
      </header>

      <div>
        <p className="text-sm text-paycore-muted mb-2">เลือกวิธีชำระเงิน</p>
        <div className="flex flex-wrap gap-2">
          {view.allowed_methods.map((m) => (
            <button
              key={m}
              onClick={() => setMethod(m)}
              className={`rounded-full px-4 py-2 text-sm border ${
                method === m ? "bg-paycore-primary border-paycore-primary text-white" : "border-white/15 text-paycore-muted"
              }`}
            >
              {METHOD_LABEL[m] ?? m}
            </button>
          ))}
        </div>
      </div>

      {method === "promptpay" && (
        <PayPromptPayButton token={token} onPaid={setView} setErr={setErr} />
      )}

      {method === "card" && view.sandbox && (
        <form onSubmit={payCard} className="space-y-3">
          <p className="text-xs rounded-lg bg-yellow-500/10 text-yellow-300 px-3 py-2">
            โหมดทดสอบ (Sandbox) — ใช้บัตรทดสอบเท่านั้น เช่น 4111 1111 1111 1111
          </p>
          <input value={cardNumber} onChange={(e) => setCardNumber(e.target.value)} inputMode="numeric"
            placeholder="หมายเลขบัตร" className="w-full rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
          <div className="flex gap-2">
            <input value={expMonth} onChange={(e) => setExpMonth(e.target.value)} inputMode="numeric"
              placeholder="MM" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
            <input value={expYear} onChange={(e) => setExpYear(e.target.value)} inputMode="numeric"
              placeholder="YYYY" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
            <input value={cvv} onChange={(e) => setCvv(e.target.value)} inputMode="numeric"
              placeholder="CVV" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
          </div>
          {err && <p className="text-red-400 text-sm">{err}</p>}
          <button disabled={busy} className="w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
            {busy ? "กำลังชำระเงิน…" : `ชำระ ${formatMoney(view.amount_minor, view.currency)}`}
          </button>
        </form>
      )}

      {method === "card" && !view.sandbox && (
        <p className="text-sm text-paycore-muted">การชำระด้วยบัตรยังไม่พร้อมใช้งานบนระบบนี้</p>
      )}
    </div>
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

// PayPromptPayButton initiates the PromptPay charge, then hands off to the status
// view (which renders the QR + polls).
function PayPromptPayButton({ token, onPaid, setErr }: { token: string; onPaid: (v: CheckoutView) => void; setErr: (s: string) => void }) {
  const [busy, setBusy] = useState(false);
  async function start() {
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ method: "promptpay" }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "สร้าง QR ไม่สำเร็จ");
      return;
    }
    onPaid(env.data as CheckoutView); // status becomes requires_action -> status view takes over
  }
  return (
    <button onClick={start} disabled={busy} className="w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
      {busy ? "กำลังสร้าง QR…" : "สร้าง QR PromptPay"}
    </button>
  );
}

// CheckoutStatusView renders the QR (for PromptPay) and polls session status
// until a terminal state, then shows success + optional return_url.
function CheckoutStatusView({ token, initial }: { token: string; initial: CheckoutView }) {
  const [view, setView] = useState<CheckoutView>(initial);
  const qrBox = useRef<HTMLDivElement>(null);

  // Render the QR whenever a payload is present and not yet paid.
  useEffect(() => {
    if (!view.qr_payload || view.status === "paid") return;
    let cancelled = false;
    (async () => {
      try {
        const QR = await waitForQRCode();
        if (cancelled || !qrBox.current) return;
        qrBox.current.innerHTML = "";
        // Use the resolved global's own CorrectLevel.M rather than a hardcoded
        // numeric literal: davidshimjs/qrcodejs defines the enum as
        // { L:1, M:0, Q:3, H:2 }, so a literal "1" is actually level L, not M.
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

  if (view.status === "paid") {
    return (
      <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-8 mt-10 text-center space-y-4">
        <div className="text-4xl">✓</div>
        <h1 className="text-xl font-semibold">ชำระเงินสำเร็จ</h1>
        <p className="text-paycore-muted">{formatMoney(view.amount_minor, view.currency)}</p>
        {view.return_url && (
          <a href={view.return_url} className="inline-block rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2">
            กลับไปที่ร้านค้า
          </a>
        )}
      </div>
    );
  }
  if (view.status === "expired" || view.status === "failed") {
    return (
      <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-8 mt-10 text-center space-y-3">
        <div className="text-4xl">⚠️</div>
        <h1 className="text-lg font-semibold">
          {view.status === "expired" ? "หมดเวลาชำระเงิน" : "ชำระเงินไม่สำเร็จ"}
        </h1>
        <p className="text-paycore-muted text-sm">โปรดลองอีกครั้ง</p>
      </div>
    );
  }

  // PromptPay awaiting: show QR + amount.
  return (
    <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10 text-center space-y-4">
      <p className="text-paycore-muted text-sm">{view.merchant_name}</p>
      <h1 className="text-lg font-semibold">สแกนเพื่อชำระด้วย PromptPay</h1>
      <p className="text-2xl font-bold">{formatMoney(view.amount_minor, view.currency)}</p>
      <div ref={qrBox} className="mx-auto bg-white p-3 rounded-lg inline-block" style={{ minHeight: 220, minWidth: 220 }} />
      <p className="text-paycore-muted text-xs">รอการยืนยันการชำระเงิน…</p>
    </div>
  );
}
