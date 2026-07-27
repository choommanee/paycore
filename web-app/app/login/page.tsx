"use client";

import { useState } from "react";

export default function LoginPage() {
  const [busy, setBusy] = useState(false);

  async function devLogin() {
    setBusy(true);
    const res = await fetch("/api/auth/dev-login", { method: "POST" });
    if (res.ok) window.location.href = "/dashboard";
    else setBusy(false);
  }

  return (
    <main className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-xl2 bg-paycore-surface border border-paycore-line p-8 shadow-cardlg">
        <h1 className="text-2xl font-semibold mb-1">PayCore</h1>
        <p className="text-paycore-muted mb-6 text-sm">เข้าสู่ระบบร้านค้า</p>

        <a
          href="/api/auth/google/start"
          className="block w-full text-center rounded-lg bg-paycore-surface2 border border-paycore-line text-paycore-text font-medium py-2.5 mb-3 hover:bg-paycore-line2"
        >
          เข้าสู่ระบบด้วย Google
        </a>

        <button
          onClick={devLogin}
          disabled={busy}
          className="block w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium py-2.5 disabled:opacity-60"
        >
          {busy ? "กำลังเข้าสู่ระบบ…" : "Dev login (sandbox)"}
        </button>

        <p className="text-paycore-muted text-xs mt-4">
          ปุ่ม Google ใช้ได้เมื่อ backend ตั้งค่า GOOGLE_CLIENT_ID แล้ว
        </p>
      </div>
    </main>
  );
}
