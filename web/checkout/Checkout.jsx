import { useState } from "react";

// Hosted checkout component (Fintech clean theme).
// The PAN is tokenized client-side into a single-use vault token; only that
// token + an Idempotency-Key are sent to POST /v1/payments. The raw card number
// never reaches the gateway API (it stays outside PCI-DSS scope — see the note on
// domain.CreatePaymentRequest.PaymentToken, "vault token, not PAN").

// Base URL is configurable via the `apiBase` prop; defaults to the same origin
// the app is served from (works locally and on a hosted deploy).
const DEFAULT_API_BASE =
  typeof window !== "undefined" && window.location ? window.location.origin : "";
// Demo merchant id (a real UUID). In production this comes from your session.
const MERCHANT_ID = "00000000-0000-0000-0000-000000000001";

// tokenize emulates the client-side vault: there is no separate tokenization
// endpoint in the router, so we mint an opaque single-use `tok_...` value here.
// The gateway detokenizes payment_token -> PAN internally (HSM-backed vault).
// Replace this with your PCI-scoped vault/iframe SDK in production; the POST below
// is unchanged.
async function tokenize({ pan }) {
  return {
    token: "tok_" + crypto.randomUUID().replace(/-/g, ""),
    last4: pan.slice(-4),
  };
}

// Unwrap the standard { success, code, data, ... } API envelope.
async function apiJson(res) {
  let body = null;
  try { body = await res.json(); } catch (_) {}
  if (!res.ok || (body && body.success === false)) {
    const msg = (body && (body.message || body.code)) || `HTTP ${res.status}`;
    const err = new Error(msg);
    err.status = res.status;
    throw err;
  }
  return body ? body.data : null;
}

const luhn = (n) => {
  let sum = 0, alt = false;
  for (let i = n.length - 1; i >= 0; i--) {
    let d = +n[i];
    if (alt) { d *= 2; if (d > 9) d -= 9; }
    sum += d; alt = !alt;
  }
  return n.length >= 13 && sum % 10 === 0;
};
const brandOf = (n) =>
  /^4/.test(n) ? "VISA" : /^(5[1-5]|2[2-7])/.test(n) ? "MASTERCARD" : /^3[47]/.test(n) ? "AMEX" : "";

export default function Checkout({
  merchant = "Acme Store",
  amountMinor = 129000,
  currency = "THB",
  reference = "INV-2026-00841",
  apiBase = DEFAULT_API_BASE,
  onPaid = () => {},
}) {
  const [card, setCard] = useState("");
  const [exp, setExp] = useState("");
  const [cvv, setCvv] = useState("");
  const [name, setName] = useState("");
  const [errors, setErrors] = useState({});
  // idle | tokenizing | authenticating (3DS) | confirming | done | error
  const [stage, setStage] = useState("idle");
  const [failMsg, setFailMsg] = useState("");
  const [result, setResult] = useState(null); // the terminal Payment from the API

  // domain.CreatePaymentRequest.Amount is a decimal (JSON string "1290.00"),
  // NOT integer minor units. amountMinor is only for ฿ display formatting.
  const amountDecimal = (amountMinor / 100).toFixed(2);
  const amount = (amountMinor / 100).toLocaleString("th-TH", { minimumFractionDigits: 2 });
  const brand = brandOf(card.replace(/\s/g, ""));

  const onCard = (v) =>
    setCard(v.replace(/\D/g, "").slice(0, 16).replace(/(.{4})/g, "$1 ").trim());
  const onExp = (v) => {
    const d = v.replace(/\D/g, "").slice(0, 4);
    setExp(d.length > 2 ? `${d.slice(0, 2)} / ${d.slice(2)}` : d);
  };

  const validate = () => {
    const e = {};
    const num = card.replace(/\s/g, "");
    const [mm, yy] = exp.split("/").map((s) => s && s.trim());
    if (!luhn(num)) e.card = "Enter a valid card number.";
    const m = +mm, y = 2000 + +yy;
    if (!m || m < 1 || m > 12 || !yy || new Date(y, m) <= new Date())
      e.exp = "Card expiry is invalid or past.";
    if (cvv.length < 3) e.cvv = "Invalid CVV.";
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  // Finish on a terminal payment state returned by the gateway.
  const finish = (p) => {
    setResult(p);
    setStage("done");
    onPaid(p);
  };

  // 3DS: gateway returned requires_action + next_action_url. Drive the issuer ACS,
  // then POST the 3DS result to /v1/payments/{id}/3ds/return and re-read the payment.
  const handleRequiresAction = async (p) => {
    setStage("authenticating");
    if (p.next_action_url) {
      // In production the ACS redirects the browser to return_url; here we open it.
      window.open(p.next_action_url, "_blank", "noopener");
    }
    const body = new URLSearchParams({ transStatus: "Y" }); // Y = issuer authenticated
    const res = await fetch(`${apiBase}/v1/payments/${p.id}/3ds/return`, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body.toString(),
    });
    finish(await apiJson(res));
  };

  const pay = async (ev) => {
    ev.preventDefault();
    if (!validate()) return;
    setFailMsg("");
    const pan = card.replace(/\s/g, "");
    const [mm, yy] = exp.split("/").map((s) => s && s.trim());

    try {
      // 1) Tokenize client-side — raw PAN never leaves the browser for our API.
      setStage("tokenizing");
      const tok = await tokenize({ pan, expMonth: +mm, expYear: 2000 + +yy });

      // 2) Create the payment with the single-use token + an Idempotency-Key.
      setStage("confirming");
      const res = await fetch(`${apiBase}/v1/payments`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Idempotency-Key": crypto.randomUUID(), // safe to retry without double-charge
        },
        body: JSON.stringify({
          merchant_id: MERCHANT_ID,
          amount: amountDecimal,     // decimal string, e.g. "1290.00"
          currency,
          payment_token: tok.token,  // vault token, not the PAN
          capture: true,             // true = charge; false = authorize only
          reference,
          return_url: typeof location !== "undefined" ? location.href : "",
        }),
      });
      const p = await apiJson(res);

      // 3) Reflect the returned state.
      if (p.status === "requires_action") {
        await handleRequiresAction(p);          // 3DS challenge step
      } else if (p.status === "authorized" || p.status === "captured") {
        finish(p);
      } else {
        setFailMsg(`Payment ${p.status || "failed"}.`);
        setStage("error");
      }
    } catch (err) {
      setFailMsg(err.message || "Something went wrong.");
      setStage("error");
    }
  };

  const inp = "w-full h-11 border rounded-lg px-3 text-sm outline-none focus:border-blue-700 focus:ring-2 focus:ring-blue-100";
  const bad = "border-red-600 ring-2 ring-red-100";

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-5">
      <form onSubmit={pay} className="w-full max-w-sm bg-white border border-slate-200 rounded-2xl p-6">
        <div className="flex items-center gap-2 mb-5">
          <div className="w-7 h-7 rounded-lg bg-blue-50 text-blue-700 flex items-center justify-center font-semibold">P</div>
          <b className="text-slate-900">{merchant}</b>
        </div>

        <div className="flex justify-between items-baseline pb-4 mb-4 border-b border-slate-200">
          <span className="text-slate-500 text-sm">Total to pay</span>
          <strong className="text-2xl">฿{amount}</strong>
        </div>

        <label className="block text-xs text-slate-500 mb-1">Card number</label>
        <div className="relative mb-1">
          <input className={`${inp} ${errors.card ? bad : "border-slate-200"}`} inputMode="numeric"
            placeholder="1234 1234 1234 1234" value={card} onChange={(e) => onCard(e.target.value)} />
          <span className="absolute right-3 top-3 text-xs font-semibold text-blue-700">{brand}</span>
        </div>
        {errors.card && <p className="text-red-600 text-xs mb-2">{errors.card}</p>}

        <div className="flex gap-2.5 mt-2">
          <div className="flex-1">
            <label className="block text-xs text-slate-500 mb-1">Expiry</label>
            <input className={`${inp} ${errors.exp ? bad : "border-slate-200"}`} inputMode="numeric"
              placeholder="MM / YY" value={exp} onChange={(e) => onExp(e.target.value)} />
          </div>
          <div className="flex-1">
            <label className="block text-xs text-slate-500 mb-1">CVV</label>
            <input className={`${inp} ${errors.cvv ? bad : "border-slate-200"}`} inputMode="numeric"
              placeholder="123" value={cvv} maxLength={4}
              onChange={(e) => setCvv(e.target.value.replace(/\D/g, ""))} />
          </div>
        </div>
        {errors.exp && <p className="text-red-600 text-xs mt-1">{errors.exp}</p>}

        <label className="block text-xs text-slate-500 mb-1 mt-3">Name on card</label>
        <input className={`${inp} border-slate-200`} placeholder="SAKDA C."
          value={name} onChange={(e) => setName(e.target.value)} />

        <button type="submit" disabled={stage !== "idle"}
          className="w-full h-12 rounded-lg bg-blue-700 hover:bg-blue-900 text-white font-semibold mt-4 disabled:opacity-60">
          {stage === "idle" ? `Pay ฿${amount}` : "Processing…"}
        </button>

        <div className="flex items-center justify-center gap-1.5 mt-3 text-slate-500 text-xs">
          🔒 Secured · PCI-DSS · 3-D Secure
        </div>

        {stage !== "idle" && (
          <div className="fixed inset-0 bg-slate-900/50 flex items-center justify-center p-5">
            <div className="bg-white rounded-2xl p-7 max-w-xs text-center">
              {stage === "done" ? (
                <>
                  <div className="w-12 h-12 rounded-full bg-emerald-50 text-emerald-700 text-2xl flex items-center justify-center mx-auto mb-3">✓</div>
                  <h3 className="text-base font-medium mb-1">Payment successful</h3>
                  <p className="text-slate-500 text-sm">
                    ฿{amount} {result?.status === "authorized" ? "authorized" : "charged"}
                    {result?.card_last4 ? ` · ••${result.card_last4}` : ""}
                    {result?.id ? ` · ${result.id}` : ""}
                  </p>
                </>
              ) : stage === "error" ? (
                <>
                  <div className="w-12 h-12 rounded-full bg-red-50 text-red-600 text-2xl flex items-center justify-center mx-auto mb-3">!</div>
                  <h3 className="text-base font-medium mb-1">Payment failed</h3>
                  <p className="text-slate-500 text-sm mb-4">{failMsg}</p>
                  <button type="button" onClick={() => { setStage("idle"); }}
                    className="w-full h-11 rounded-lg bg-blue-700 hover:bg-blue-900 text-white font-semibold">
                    Try again
                  </button>
                </>
              ) : (
                <>
                  <div className="w-9 h-9 border-[3px] border-blue-100 border-t-blue-700 rounded-full animate-spin mx-auto mb-3" />
                  <h3 className="text-base font-medium mb-1">
                    {stage === "tokenizing" ? "Securing your card"
                      : stage === "authenticating" ? "Authenticating your card"
                      : "Confirming payment"}
                  </h3>
                  <p className="text-slate-500 text-sm">
                    {stage === "tokenizing" ? "Tokenizing…"
                      : stage === "authenticating" ? "Redirecting to 3-D Secure…"
                      : "Do not close this window."}
                  </p>
                </>
              )}
            </div>
          </div>
        )}
      </form>
    </div>
  );
}
