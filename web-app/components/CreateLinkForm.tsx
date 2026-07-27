"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const METHODS = ["card", "promptpay", "mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"];

export default function CreateLinkForm() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [amount, setAmount] = useState("");
  const [methods, setMethods] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");

  function toggle(m: string) {
    setMethods((cur) => (cur.includes(m) ? cur.filter((x) => x !== m) : [...cur, m]));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    const baht = Number(amount);
    if (!title.trim() || !(baht > 0)) {
      setErr("กรอกชื่อและจำนวนเงินให้ถูกต้อง");
      return;
    }
    setBusy(true);
    const res = await fetch("/api/payment-links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: title.trim(),
        amount_minor: Math.round(baht * 100),
        allowed_methods: methods,
      }),
    });
    setBusy(false);
    if (res.ok) {
      setTitle(""); setAmount(""); setMethods([]); setOpen(false);
      router.refresh();
    } else {
      const body = await res.json().catch(() => null);
      setErr(body?.message ?? "สร้างลิงก์ไม่สำเร็จ");
    }
  }

  if (!open) {
    return (
      <button onClick={() => setOpen(true)} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2">
        + สร้าง Payment Link
      </button>
    );
  }

  return (
    <form onSubmit={submit} className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-5 space-y-4">
      <div>
        <label className="block text-sm text-paycore-muted mb-1">ชื่อรายการ</label>
        <input value={title} onChange={(e) => setTitle(e.target.value)} className="w-full rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 focus:outline-none focus:border-paycore-primary" placeholder="เช่น กาแฟลาเต้" />
      </div>
      <div>
        <label className="block text-sm text-paycore-muted mb-1">จำนวนเงิน (บาท)</label>
        <input value={amount} onChange={(e) => setAmount(e.target.value)} inputMode="decimal" className="w-full rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 focus:outline-none focus:border-paycore-primary" placeholder="50.00" />
      </div>
      <div>
        <label className="block text-sm text-paycore-muted mb-2">ช่องทางจ่าย (ว่าง = ทุกช่องทาง)</label>
        <div className="flex flex-wrap gap-2">
          {METHODS.map((m) => (
            <button type="button" key={m} onClick={() => toggle(m)}
              className={`rounded-full px-3 py-1 text-sm border ${methods.includes(m) ? "bg-paycore-primary border-paycore-primary text-white" : "bg-paycore-surface2 border-paycore-line text-paycore-text2 hover:border-paycore-primary"}`}>
              {m}
            </button>
          ))}
        </div>
      </div>
      {err && <p className="text-paycore-danger text-sm">{err}</p>}
      <div className="flex gap-2">
        <button disabled={busy} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
          {busy ? "กำลังสร้าง…" : "สร้างลิงก์"}
        </button>
        <button type="button" onClick={() => setOpen(false)} className="rounded-lg border border-paycore-line px-4 py-2 text-paycore-text2 hover:bg-paycore-surface2">ยกเลิก</button>
      </div>
    </form>
  );
}
