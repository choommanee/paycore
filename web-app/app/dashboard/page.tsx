import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

type AuthMe = { email: string; name: string; merchant_name: string; merchant_id: string };
type Stats = {
  count: number;
  volumeMinor: number;
  byStatus: { authorized: number; captured: number; refunded: number; failed: number };
  successRate: number;
  refundRatio: number;
};

function pct(fraction: number): string {
  return `${(fraction * 100).toFixed(1)}%`;
}

export default async function DashboardHome() {
  const [meRes, statsRes] = await Promise.all([serverGet("/auth/me"), serverGet("/stats")]);
  if (meRes.status === 401 || statsRes.status === 401) redirect("/login");
  if (!meRes.ok) throw new Error(`auth/me failed: ${meRes.status}`);
  if (!statsRes.ok) throw new Error(`stats failed: ${statsRes.status}`);
  const me: AuthMe = (await meRes.json()).data;
  const s: Stats = (await statsRes.json()).data;

  const tiles = [
    { label: "ยอดรับชำระ (30 วัน)", value: formatMoney(s.volumeMinor) },
    { label: "จำนวนธุรกรรม", value: String(s.count) },
    { label: "อัตราสำเร็จ", value: pct(s.successRate) },
    { label: "อัตราคืนเงิน", value: pct(s.refundRatio) },
  ];

  return (
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <DashboardNav active="/dashboard" />
      <p className="text-paycore-muted text-sm mb-6">สวัสดี {me.name || me.email} · {me.merchant_name}</p>

      <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {tiles.map((t) => (
          <div key={t.label} className="rounded-xl2 bg-paycore-surface p-5">
            <p className="text-paycore-muted text-xs">{t.label}</p>
            <p className="text-2xl font-semibold mt-2">{t.value}</p>
          </div>
        ))}
      </section>

      <p className="text-paycore-muted text-xs mt-6">
        ตัวเลขคำนวณจากการชำระด้วยบัตร (card) ในรอบ 30 วันล่าสุด
      </p>

      <div className="mt-6 flex gap-4">
        <a href="/transactions" className="text-paycore-primary hover:underline">ดูธุรกรรมทั้งหมด →</a>
        <a href="/links" className="text-paycore-primary hover:underline">จัดการลิงก์ชำระเงิน →</a>
      </div>
    </main>
  );
}
