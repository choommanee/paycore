import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import DashboardNav from "@/components/DashboardNav";
import CopyCode from "@/components/CopyCode";

// Merchant profile as returned by GET /v1/me. Only the fields this page shows.
type Profile = {
  id: string;
  name: string;
  settlement_currency: string;
  status: string;
  webhook_url?: string;
};

// Documented API base. This is the illustrative host shown in code samples;
// server-to-server clients call the merchant's own PayCore API domain. From the
// dashboard's own origin the same routes are reachable under /api (the Next proxy
// rewrites /api/* → the backend's /v1/*).
const API_BASE = "https://api.paycore.app/v1";

// The real, registered v1 routes (internal/router/router.go) grouped for the
// reference table. Auth column: "API key" = X-API-Key; "Public" = token/no key.
const ENDPOINTS: { method: string; path: string; desc: string; auth: string }[] = [
  { method: "POST", path: "/payment-links", desc: "สร้างลิงก์ชำระเงิน", auth: "API key" },
  { method: "GET", path: "/payment-links", desc: "รายการลิงก์ชำระเงิน", auth: "API key" },
  { method: "POST", path: "/checkout/sessions", desc: "เปิด hosted checkout session", auth: "Public" },
  { method: "GET", path: "/checkout/sessions/:token", desc: "ดึงสถานะ session", auth: "Public" },
  { method: "POST", path: "/payments", desc: "สร้างการชำระเงิน (card)", auth: "API key" },
  { method: "GET", path: "/payments", desc: "รายการการชำระเงิน", auth: "API key" },
  { method: "POST", path: "/payments/:id/refund", desc: "คืนเงิน (ต้องมี Idempotency-Key)", auth: "API key" },
  { method: "POST", path: "/qr-payments", desc: "สร้าง PromptPay / QR", auth: "API key" },
  { method: "GET", path: "/transactions", desc: "ธุรกรรมรวมทุกช่องทาง", auth: "API key" },
  { method: "GET", path: "/settlements", desc: "รอบจ่ายเงิน (payouts)", auth: "API key" },
];

const METHOD_TONE: Record<string, string> = {
  GET: "bg-paycore-successBg text-paycore-success",
  POST: "bg-paycore-accentBg text-paycore-accent",
  PUT: "bg-paycore-warnBg text-paycore-warn",
  PATCH: "bg-paycore-warnBg text-paycore-warn",
};

function MethodBadge({ method }: { method: string }) {
  return (
    <span
      className={`inline-block rounded px-1.5 py-0.5 font-mono text-[10px] font-bold ${
        METHOD_TONE[method] ?? "bg-paycore-line2 text-paycore-muted"
      }`}
    >
      {method}
    </span>
  );
}

export default async function DevelopersPage() {
  const res = await serverGet("/me");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`me failed: ${res.status}`);
  const p: Profile = (await res.json()).data;

  const createLink = `curl ${API_BASE}/payment-links \\
  -H "X-API-Key: sk_live_xxxxxxxxxxxxxxxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 49900,
    "currency": "THB",
    "description": "เสื้อยืด PayCore",
    "reference": "order-1042"
  }'`;

  const createCheckout = `curl ${API_BASE}/checkout/sessions \\
  -H "Content-Type: application/json" \\
  -d '{
    "amount": 49900,
    "currency": "THB",
    "success_url": "https://shop.example.com/thanks"
  }'
# → { "token": "...", "url": "https://pay.paycore.app/pay/<publicId>" }
# เปลี่ยนเส้นทางลูกค้าไปที่ url เพื่อชำระเงิน`;

  const refund = `curl ${API_BASE}/payments/pay_123/refund \\
  -H "X-API-Key: sk_live_xxxxxxxxxxxxxxxx" \\
  -H "Idempotency-Key: refund-order-1042" \\
  -H "Content-Type: application/json" \\
  -d '{ "amount": 49900 }'`;

  const verifyNode = `import crypto from "node:crypto";

// Verify a PayCore webhook: recompute v1 over "<t>.<rawBody>", compare in
// constant time, and reject deliveries whose timestamp is outside ±5 min.
function verifyPayCoreWebhook(rawBody, headers, secret, tolerance = 300) {
  const ts = headers["x-paycore-timestamp"];
  const sig = headers["x-paycore-signature"]; // "t=<unix>,v1=<hex>"
  if (!ts || !sig) return false;
  if (Math.abs(Date.now() / 1000 - Number(ts)) > tolerance) return false; // replay

  const parts = Object.fromEntries(sig.split(",").map((p) => p.split("=")));
  const signed = Buffer.concat([Buffer.from(parts.t + "."), rawBody]);
  const expected = crypto.createHmac("sha256", secret).update(signed).digest("hex");

  const a = Buffer.from(expected);
  const b = Buffer.from(parts.v1 ?? "");
  return a.length === b.length && crypto.timingSafeEqual(a, b);
}`;

  const verifyPython = `import hmac, hashlib, time

def verify(raw_body: bytes, header_sig: str, header_ts: str, secret: str, tol=300) -> bool:
    if abs(time.time() - int(header_ts)) > tol:      # reject stale / replayed
        return False
    parts = dict(p.split("=", 1) for p in header_sig.split(","))
    expected = hmac.new(secret.encode(), f"{parts['t']}.".encode() + raw_body,
                        hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, parts["v1"])`;

  return (
    <main className="min-h-screen bg-paycore-bg p-6 md:p-8">
      <div className="mx-auto max-w-4xl">
        <DashboardNav active="/developers" />

        {/* Header */}
        <div className="mb-6">
          <h1 className="text-lg font-semibold text-paycore-text">
            นักพัฒนา (Developers)
          </h1>
          <p className="mt-0.5 text-xs text-paycore-muted">
            รวม REST API, ตัวอย่างโค้ด และวิธีตรวจสอบลายเซ็น webhook แบบมาตรฐาน (HMAC-SHA256)
          </p>
        </div>

        {/* Account facts */}
        <section className="mb-8 grid grid-cols-1 gap-4 md:grid-cols-3">
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              API base URL
            </p>
            <p className="mt-1 break-all font-mono text-[13px] text-paycore-text">{API_BASE}</p>
          </div>
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              Merchant ID
            </p>
            <p className="mt-1 break-all font-mono text-[13px] text-paycore-text">{p.id}</p>
          </div>
          <div className="rounded-xl2 border border-paycore-line bg-paycore-surface p-4 shadow-card">
            <p className="text-[11px] font-semibold uppercase tracking-wide text-paycore-muted">
              Webhook URL
            </p>
            <p className="mt-1 break-all font-mono text-[13px] text-paycore-text">
              {p.webhook_url || "— ยังไม่ได้ตั้ง (ตั้งค่า → Webhook)"}
            </p>
          </div>
        </section>

        {/* Authentication */}
        <section className="mb-8">
          <h2 className="mb-2 text-sm font-semibold text-paycore-text">1 · การยืนยันตัวตน</h2>
          <p className="mb-3 text-[13px] leading-relaxed text-paycore-text2">
            ทุกคำขอฝั่งเซิร์ฟเวอร์แนบ API key ผ่านเฮดเดอร์{" "}
            <code className="rounded bg-paycore-surface2 px-1 py-0.5 font-mono text-[12px]">
              X-API-Key
            </code>
            . สร้างหรือหมุน key ได้ที่หน้า{" "}
            <a href="/settings" className="font-medium text-paycore-accent hover:underline">
              ตั้งค่า → API key
            </a>{" "}
            (แสดงเต็มครั้งเดียวตอนสร้าง — เก็บไว้ให้ปลอดภัย ห้ามฝังในโค้ดฝั่งเบราว์เซอร์).
          </p>
          <CopyCode
            label="HTTP header"
            code={`X-API-Key: sk_live_xxxxxxxxxxxxxxxx`}
          />
        </section>

        {/* Endpoint reference */}
        <section className="mb-8">
          <h2 className="mb-2 text-sm font-semibold text-paycore-text">2 · Endpoints</h2>
          <div className="overflow-x-auto rounded-xl2 border border-paycore-line bg-paycore-surface shadow-card">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr>
                  <th className="border-b border-paycore-line bg-paycore-surface2 px-4 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    Method
                  </th>
                  <th className="border-b border-paycore-line bg-paycore-surface2 px-4 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    Path
                  </th>
                  <th className="border-b border-paycore-line bg-paycore-surface2 px-4 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    คำอธิบาย
                  </th>
                  <th className="border-b border-paycore-line bg-paycore-surface2 px-4 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-paycore-muted">
                    Auth
                  </th>
                </tr>
              </thead>
              <tbody>
                {ENDPOINTS.map((e) => (
                  <tr
                    key={`${e.method} ${e.path}`}
                    className="border-b border-paycore-line2 last:border-b-0"
                  >
                    <td className="px-4 py-2.5">
                      <MethodBadge method={e.method} />
                    </td>
                    <td className="px-4 py-2.5 font-mono text-[12.5px] text-paycore-text">
                      {e.path}
                    </td>
                    <td className="px-4 py-2.5 text-[13px] text-paycore-text2">{e.desc}</td>
                    <td className="px-4 py-2.5 text-[12px] text-paycore-muted">{e.auth}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        {/* Quick examples */}
        <section className="mb-8">
          <h2 className="mb-3 text-sm font-semibold text-paycore-text">3 · ตัวอย่างการใช้งาน</h2>
          <p className="mb-3 text-[13px] leading-relaxed text-paycore-text2">
            จำนวนเงินเป็น{" "}
            <span className="font-medium text-paycore-text">หน่วยย่อย (สตางค์)</span> — เช่น
            ฿499.00 = <code className="font-mono">49900</code>.
          </p>
          <div className="space-y-4">
            <div>
              <p className="mb-1.5 text-[13px] font-medium text-paycore-text">สร้างลิงก์ชำระเงิน</p>
              <CopyCode label="cURL" code={createLink} />
            </div>
            <div>
              <p className="mb-1.5 text-[13px] font-medium text-paycore-text">
                เปิด hosted checkout
              </p>
              <CopyCode label="cURL" code={createCheckout} />
            </div>
            <div>
              <p className="mb-1.5 text-[13px] font-medium text-paycore-text">คืนเงิน</p>
              <CopyCode label="cURL" code={refund} />
            </div>
          </div>
        </section>

        {/* Webhooks + HMAC */}
        <section className="mb-8">
          <h2 className="mb-2 text-sm font-semibold text-paycore-text">
            4 · Webhooks &amp; การตรวจลายเซ็น (HMAC-SHA256)
          </h2>
          <p className="mb-3 text-[13px] leading-relaxed text-paycore-text2">
            ทุก webhook ที่ PayCore ส่งออกจะถูกเซ็นเพื่อให้คุณยืนยันว่ามาจาก PayCore จริงและไม่ถูก
            replay. เฮดเดอร์ที่แนบมา:
          </p>
          <div className="mb-4 space-y-2 rounded-xl2 border border-paycore-line bg-paycore-surface p-4 text-[12.5px] shadow-card">
            <div>
              <code className="font-mono text-paycore-accent">X-PayCore-Timestamp</code>
              <span className="text-paycore-text2"> — unix seconds ตอนส่ง</span>
            </div>
            <div>
              <code className="font-mono text-paycore-accent">X-PayCore-Signature</code>
              <span className="text-paycore-text2">
                {" "}
                — <span className="font-mono">t=&lt;unix&gt;,v1=&lt;hex HMAC(secret, &quot;&lt;t&gt;.&lt;rawBody&gt;&quot;)&gt;</span>{" "}
                (มาตรฐาน กัน replay — timestamp ผูกอยู่ใน MAC)
              </span>
            </div>
            <div>
              <code className="font-mono text-paycore-muted">X-Signature</code>
              <span className="text-paycore-muted">
                {" "}
                — <span className="font-mono">sha256=&lt;hex HMAC(secret, rawBody)&gt;</span> (legacy
                body-only เก็บไว้เพื่อความเข้ากันได้)
              </span>
            </div>
          </div>
          <p className="mb-3 text-[13px] leading-relaxed text-paycore-text2">
            <span className="font-medium text-paycore-text">วิธีตรวจ:</span> คำนวณ{" "}
            <code className="font-mono">v1</code> ใหม่บน{" "}
            <code className="font-mono">&quot;&lt;t&gt;.&lt;rawBody&gt;&quot;</code> ด้วย{" "}
            <span className="font-medium">signing secret</span> ของคุณ (แสดงครั้งเดียวตอนตั้ง
            Webhook), เปรียบเทียบแบบ <span className="font-medium">constant-time</span>, และปฏิเสธ
            ถ้า <code className="font-mono">t</code> อยู่นอกกรอบ ±5 นาที. ต้องคำนวณจาก{" "}
            <span className="font-medium">raw body</span> ก่อน parse JSON.
          </p>
          <div className="space-y-4">
            <CopyCode label="Node.js" code={verifyNode} />
            <CopyCode label="Python" code={verifyPython} />
          </div>
        </section>

        <p className="text-xs text-paycore-muted">
          ต้องการทดสอบ? ใช้ค่าตั้งต้น sandbox — บัตร{" "}
          <code className="font-mono">4242 4242 4242 4242</code> ผ่านเสมอ, PromptPay/e-wallet
          อนุมัติผ่านหน้าจำลอง. ดูสถานะจริงได้ที่{" "}
          <a href="/transactions" className="font-medium text-paycore-accent hover:underline">
            ธุรกรรม
          </a>
          .
        </p>
      </div>
    </main>
  );
}
