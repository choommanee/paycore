"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LinkActions({ id, url, status }: { id: string; url: string; status: string }) {
  const router = useRouter();
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  async function disable() {
    setBusy(true);
    const res = await fetch(`/api/payment-links/${id}`, { method: "PATCH" });
    setBusy(false);
    if (res.ok) router.refresh();
  }

  return (
    <div className="mt-6 space-y-3">
      <div className="flex items-center gap-2">
        <input readOnly value={url} className="flex-1 rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 text-sm text-paycore-text2" />
        <button onClick={copy} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm">
          {copied ? "คัดลอกแล้ว ✓" : "คัดลอก"}
        </button>
      </div>
      <div className="flex gap-2">
        <a href={url} target="_blank" className="rounded-lg border border-paycore-line px-4 py-2 text-sm text-paycore-text2 hover:bg-paycore-surface2">เปิดหน้าจ่าย ↗</a>
        {status === "active" && (
          <button onClick={disable} disabled={busy} className="rounded-lg border border-paycore-danger/40 text-paycore-danger px-4 py-2 text-sm hover:bg-paycore-dangerBg disabled:opacity-60">
            {busy ? "…" : "ปิดลิงก์"}
          </button>
        )}
      </div>
    </div>
  );
}
