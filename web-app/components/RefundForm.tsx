"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function RefundForm({
  paymentId,
  remaining,
  currency,
}: {
  paymentId: string;
  remaining: number; // major units still refundable
  currency: string;
}) {
  const router = useRouter();
  const [amount, setAmount] = useState(remaining.toFixed(2));
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  if (remaining <= 0) {
    return <p className="text-paycore-muted text-sm">คืนเงินครบแล้ว — ไม่มียอดคงเหลือให้คืน</p>;
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    const res = await fetch(`/api/payments/${paymentId}/refund`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": crypto.randomUUID(),
      },
      body: JSON.stringify({ amount, reason: reason || undefined }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (res.ok) {
      setDone(true);
      router.refresh();
      return;
    }
    setErr(env?.message ?? `คืนเงินไม่สำเร็จ (${res.status})`);
  }

  return (
    <form onSubmit={submit} className="mt-6 space-y-3 border-t border-paycore-line pt-6">
      <h2 className="font-medium">คืนเงิน</h2>
      <div className="flex gap-2">
        <input
          type="number"
          step="0.01"
          min="0.01"
          max={remaining}
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="w-40 rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 text-sm focus:outline-none focus:border-paycore-primary"
        />
        <input
          type="text"
          placeholder="เหตุผล (ไม่บังคับ)"
          maxLength={140}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          className="flex-1 rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 text-sm focus:outline-none focus:border-paycore-primary"
        />
      </div>
      <p className="text-paycore-muted text-xs">คืนได้สูงสุด {remaining.toFixed(2)} {currency}</p>
      {err && <p className="text-paycore-danger text-sm">{err}</p>}
      {done && <p className="text-paycore-success text-sm">คืนเงินสำเร็จ</p>}
      <button
        type="submit"
        disabled={busy}
        className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm disabled:opacity-60"
      >
        {busy ? "กำลังคืนเงิน…" : "ยืนยันคืนเงิน"}
      </button>
    </form>
  );
}
