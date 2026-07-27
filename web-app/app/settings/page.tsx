import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import DashboardNav from "@/components/DashboardNav";
import RotateKeyButton from "@/components/RotateKeyButton";
import WebhookForm from "@/components/WebhookForm";

type Profile = {
  id: string;
  name: string;
  mcc?: string;
  settlement_currency: string;
  status: string;
  webhook_url?: string;
};

export default async function SettingsPage() {
  const res = await serverGet("/me");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`me failed: ${res.status}`);
  const p: Profile = (await res.json()).data;

  const rows: [string, string][] = [
    ["ชื่อร้านค้า", p.name],
    ["Merchant ID", p.id],
    ["สกุลเงินตั้งจ่าย", p.settlement_currency],
    ["MCC", p.mcc || "—"],
    ["สถานะ", p.status],
  ];

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <DashboardNav active="/settings" />
      <h1 className="text-xl font-semibold mb-6">ตั้งค่า</h1>

      <section className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-6 mb-6">
        <h2 className="font-medium mb-4">ข้อมูลร้านค้า</h2>
        <dl className="space-y-2 text-sm">
          {rows.map(([k, v]) => (
            <div key={k} className="flex justify-between">
              <dt className="text-paycore-muted">{k}</dt>
              <dd className="break-all text-right">{v}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-6 mb-6">
        <h2 className="font-medium mb-4">API key</h2>
        <RotateKeyButton />
      </section>

      <section className="rounded-xl2 bg-paycore-surface border border-paycore-line shadow-card p-6">
        <h2 className="font-medium mb-4">Webhook</h2>
        {p.webhook_url && (
          <p className="text-paycore-muted text-xs mb-3">ปัจจุบัน: {p.webhook_url}</p>
        )}
        <WebhookForm initialUrl={p.webhook_url ?? ""} />
      </section>
    </main>
  );
}
