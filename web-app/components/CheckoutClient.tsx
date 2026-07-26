"use client";

import { useEffect, useRef, useState } from "react";
import { formatMoney } from "@/lib/format";

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

// Placeholder components implemented fully in Task 8; declared here so this file
// compiles independently. Task 8 REPLACES these with the QR + polling versions.
function PayPromptPayButton(_: { token: string; onPaid: (v: CheckoutView) => void; setErr: (s: string) => void }) {
  return null;
}
function CheckoutStatusView(_: { token: string; initial: CheckoutView }) {
  return null;
}
