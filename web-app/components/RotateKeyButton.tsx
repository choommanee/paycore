"use client";

import { useState } from "react";

export default function RotateKeyButton() {
  const [busy, setBusy] = useState(false);
  const [key, setKey] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);

  async function rotate() {
    setBusy(true);
    setErr(null);
    const res = await fetch("/api/me/rotate-key", { method: "POST" });
    const env = await res.json().catch(() => null);
    setBusy(false);
    setConfirming(false);
    if (res.ok) {
      setKey(env?.data?.api_key ?? null);
      return;
    }
    setErr(env?.message ?? `หมุนคีย์ไม่สำเร็จ (${res.status})`);
  }

  return (
    <div className="space-y-3">
      <p className="text-paycore-muted text-sm">
        API key จะแสดงเพียงครั้งเดียวตอนสร้างหรือหมุนคีย์ ระบบเก็บเฉพาะค่าแฮชจึงแสดงคีย์เดิมซ้ำไม่ได้
      </p>
      {key ? (
        <div className="rounded-lg bg-paycore-bg border border-white/10 p-3">
          <p className="text-paycore-muted text-xs mb-1">คีย์ใหม่ (บันทึกทันที จะไม่แสดงอีก)</p>
          <code className="text-sm break-all">{key}</code>
        </div>
      ) : confirming ? (
        <div className="flex gap-2">
          <button
            onClick={rotate}
            disabled={busy}
            className="rounded-lg bg-red-500/90 hover:bg-red-500 text-white px-4 py-2 text-sm disabled:opacity-60"
          >
            {busy ? "กำลังหมุนคีย์…" : "ยืนยันหมุนคีย์ (คีย์เดิมใช้ไม่ได้ทันที)"}
          </button>
          <button onClick={() => setConfirming(false)} className="rounded-lg border border-white/15 px-4 py-2 text-sm">
            ยกเลิก
          </button>
        </div>
      ) : (
        <button
          onClick={() => setConfirming(true)}
          className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm"
        >
          หมุน API key
        </button>
      )}
      {err && <p className="text-red-400 text-sm">{err}</p>}
    </div>
  );
}
