"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function WebhookForm({ initialUrl }: { initialUrl: string }) {
  const router = useRouter();
  const [url, setUrl] = useState(initialUrl);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [secret, setSecret] = useState<string | null>(null);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    const res = await fetch("/api/me/webhook", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (res.ok) {
      setSecret(env?.data?.signing_secret ?? null);
      router.refresh();
      return;
    }
    setErr(env?.message ?? `บันทึกไม่สำเร็จ (${res.status})`);
  }

  return (
    <form onSubmit={save} className="space-y-3">
      <input
        type="url"
        required
        maxLength={2048}
        placeholder="https://example.com/webhooks/paycore"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        className="w-full rounded-lg bg-paycore-surface2 border border-paycore-line px-3 py-2 text-sm focus:outline-none focus:border-paycore-primary"
      />
      {err && <p className="text-paycore-danger text-sm">{err}</p>}
      {secret && (
        <div className="rounded-lg bg-paycore-surface2 border border-paycore-line p-3">
          <p className="text-paycore-muted text-xs mb-1">Signing secret ใหม่ (แสดงครั้งเดียว)</p>
          <code className="text-sm break-all">{secret}</code>
        </div>
      )}
      <button
        type="submit"
        disabled={busy}
        className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm disabled:opacity-60"
      >
        {busy ? "กำลังบันทึก…" : "บันทึก Webhook"}
      </button>
    </form>
  );
}
