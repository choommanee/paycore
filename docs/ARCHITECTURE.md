# Payment Gateway — เอกสารออกแบบสถาปัตยกรรม (Architecture Design)

เวอร์ชัน 0.1 · ประเภท: **Full Acquiring Payment Gateway** · ภาษา: **Go (Fiber)**

> เอกสารนี้ครอบคลุมสถาปัตยกรรมระบบ, การไหลของธุรกรรม, การออกแบบข้อมูล, ความปลอดภัย/PCI-DSS
> และการวางระบบให้สอดคล้องกับกฎหมายไทยเพื่อ **ยื่นขอใบอนุญาตจากธนาคารแห่งประเทศไทย (ธปท.)** ได้จริง
> รายละเอียดใบอนุญาต ดูเอกสารแยก `COMPLIANCE-TH.md`

---

## 1. ขอบเขตและนิยาม

Full Acquiring Gateway = ระบบที่ **รับ-ประมวลผลการชำระเงินด้วยบัตร** โดยตรงกับ card network / ธนาคาร ผู้ให้บริการต้อง:

- ถือ **ใบอนุญาต Acquiring Service** ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 (ทุนจดทะเบียนชำระแล้วขั้นต่ำ 50 ล้านบาท)
- ผ่านมาตรฐาน **PCI-DSS Level 1** (ประมวลผล > 6 ล้านรายการ/ปี หรือถูกกำหนดโดย acquirer/แบรนด์บัตร)
- ทำสัญญากับ **sponsoring bank / card scheme** (Visa, Mastercard) เพื่อเข้าถึง authorization & settlement network

หากยังไม่พร้อมลงทุนเต็มรูปแบบ แนะนำเริ่มจาก **Payment Facilitating Service** (ทุน 10 ล้านบาท) โดยต่อกับ acquirer ที่มีอยู่แล้วก่อน — สถาปัตยกรรมชุดนี้รองรับทั้งสองโหมดผ่านชั้น `external.Acquirer`

---

## 2. หลักการออกแบบ (Design Principles)

1. **Cardholder data ออกจาก scope ให้มากที่สุด** — ใช้ tokenization/vault แยกส่วน; ระบบหลักเห็นแค่ token + `card_last4`
2. **Money เป็น integer minor units (สตางค์)** — ไม่ใช้ float; คำนวณด้วย `decimal.Decimal`
3. **Ledger แบบ append-only (double-entry)** — เป็น source of truth สำหรับ settlement & reconciliation
4. **Idempotency ทุก endpoint ที่ขยับเงิน** — กันชาร์จซ้ำจาก retry
5. **Auditability** — ทุก state change ลง `audit_log` (บังคับตาม PCI/ธปท.)
6. **Clean Architecture** — Handler → Service → Repository; แทน acquirer/3DS/vault ด้วย interface เพื่อสลับผู้ให้บริการและเทสได้
7. **Fail closed** — เมื่อไม่แน่ใจสถานะธุรกรรม ให้ถือว่ายังไม่สำเร็จ และ reconcile ภายหลัง

---

## 3. สถาปัตยกรรมระดับระบบ (System Context)

```
Merchant (web/app/POS)
        │  HTTPS + API key/JWT + Idempotency-Key
        ▼
┌─────────────────────────────────────────────────────────┐
│                  PAYMENT GATEWAY (Go/Fiber)              │
│  API Edge → Payment Core → Ledger → Webhook/Notify      │
│                     │            │                       │
│     Tokenization Vault (PCI)     │                       │
│     Risk/Fraud Engine            │                       │
└─────────┬───────────────────────┴───────────────────────┘
          │ ISO 8583 / acquirer API           │ 3DS 2.x
          ▼                                    ▼
   Acquirer / Card Switch              3-D Secure (ACS/DS)
          │
          ▼
   Card Networks (Visa / Mastercard) → Issuing Banks
```

องค์ประกอบภายนอก: ธนาคารผู้รับเชื่อม (sponsor), card scheme, ผู้ให้บริการ 3DS, HSM/KMS, ระบบ settlement/หักบัญชี (เช่น ITMX สำหรับ local rails)

---

## 4. องค์ประกอบภายใน (Components)

| Component | หน้าที่ | โฟลเดอร์ในโปรเจ็ค |
|-----------|--------|-------------------|
| **API Edge / Middleware** | rate limit, auth (API key/JWT), request-id, idempotency, logging, recovery | `internal/middleware` |
| **Payment Core (Service)** | authorize / capture / void / refund + state machine | `internal/service` |
| **Tokenization Vault** | เก็บ/คืน PAN แบบเข้ารหัสด้วย HSM/KMS (อยู่ใน PCI scope แยก) | `internal/pkg/crypto` (+ บริการแยก) |
| **Acquirer Adapter** | คุยกับ card switch (ISO 8583 / REST) | `internal/external` |
| **3DS Adapter** | 3-D Secure 2.x authentication | `internal/external` + `internal/pkg/threeds` |
| **Risk/Fraud Engine** | scoring, velocity, blacklist (เริ่มด้วย rule-based) | `internal/service` (แยก module ภายหลัง) |
| **Ledger** | บันทึกรายการเงินแบบ append-only | ตาราง `ledger_entries` |
| **Webhook/Notifier** | แจ้ง merchant แบบ at-least-once + retry + ลงลายเซ็น | ตาราง `webhook_events` |
| **Reconciliation & Settlement** | กระทบยอดกับ acquirer, สรุปยอดจ่าย merchant | worker แยก (Phase 3) |
| **Admin/Merchant API** | onboarding, คีย์, รายงาน | เพิ่มภายหลัง |

### เหตุผลที่เลือก Go (Fiber)
- Concurrency & latency ต่ำ เหมาะกับ authorization ที่ต้องตอบเร็วและ throughput สูง
- Static binary + distroless image → attack surface เล็ก ดีต่อการผ่าน PCI
- ตรงกับสแตกเดิมของทีม (LOS: Fiber + sqlc + zerolog + Viper) ลด learning curve และใช้ CI/CD ร่วมกันได้

---

## 5. การไหลของธุรกรรม (Transaction Flows)

### 5.1 Authorize + Capture (การ์ด, มี 3DS)
1. Merchant สร้าง **single-use payment token** ฝั่ง client กับ Vault (PAN ไม่ผ่าน server ของ merchant)
2. `POST /v1/payments` พร้อม `Idempotency-Key` + `payment_token`
3. Gateway: ตรวจ idempotency → detokenize (ใน PCI scope) → risk scoring
4. ถ้า issuer ต้อง challenge → คืน `requires_action` + `next_action_url` (3DS)
5. หลัง 3DS สำเร็จ → ส่ง authorization ไป acquirer → ได้ `auth_code`
6. เขียน `payments` + `ledger_entries(authorize)` ใน transaction เดียว
7. ถ้า `capture=true` → capture ทันที → `ledger_entries(capture)` → สถานะ `captured`
8. ยิง webhook `payment.captured` ให้ merchant

### 5.2 Refund
`POST /v1/payments/{id}/refund` → ตรวจ `refunded + amount ≤ captured` → ส่ง refund ไป acquirer → `ledger_entries(refund)` → `partial_refunded`/`refunded`

### 5.3 Void
`POST /v1/payments/{id}/void` → ใช้ก่อน capture เท่านั้น → reverse authorization → `voided`

### 5.4 State Machine
```
requires_action ──3DS ok──▶ authorized ──capture──▶ captured ──refund──▶ partial_refunded ──▶ refunded
       │                        │
       └── fail ──▶ failed      └── void ──▶ voided
```

---

## 6. การออกแบบข้อมูล (Data Model)

ตารางหลัก (ดู `migrations/000001_init_schema.up.sql`):

- **merchants** — ผู้ค้า, เก็บ `api_key_hash` (ไม่เก็บคีย์ดิบ), MCC, สกุลเงิน settlement
- **payments** — aggregate หลัก, เงินเป็น `*_minor` (BIGINT), unique `(merchant_id, idempotency_key)`
- **ledger_entries** — append-only, `entry_type` = authorize/capture/refund/void/fee/settlement
- **refunds** — refund หลายรายการต่อ 1 payment
- **webhook_events** — คิวส่ง webhook + retry
- **audit_log** — append-only audit trail

**ข้อห้ามสำคัญ:** ห้ามเก็บ full PAN, CVV/CVV2, PIN, full magnetic track ใน operational DB นี้เด็ดขาด — เก็บได้แค่ `card_brand` + `card_last4` เพื่อแสดงผล

---

## 7. ความปลอดภัยและ PCI-DSS

| หัวข้อ | แนวทาง |
|--------|--------|
| **Scope minimization** | Vault/tokenization แยก network segment; ระบบหลักไม่แตะ PAN |
| **Encryption at rest** | คีย์อยู่ใน HSM/KMS; PAN เข้ารหัสด้วย envelope encryption; ไม่มีคีย์ในโค้ด/คอนฟิก |
| **Encryption in transit** | TLS 1.2+; mTLS ระหว่าง service ภายใน |
| **Key management** | key rotation, dual control, split knowledge (PCI Req 3) |
| **Access control** | least privilege, RBAC, MFA สำหรับ admin (PCI Req 7-8) |
| **Logging & monitoring** | structured log ไม่มี card data; ห้าม log request body; SIEM + alert (PCI Req 10) |
| **Network** | WAF, DDoS protection, segmentation, IDS/IPS (PCI Req 1, 11) |
| **App security** | input validation, parameterized SQL (sqlc), rate limit, idempotency, secrets manager |
| **Vuln management** | dependency scan (`govulncheck`), SAST/DAST ใน CI, quarterly ASV scan, annual pentest |
| **Fraud** | velocity check, 3DS 2.x, blacklist, anomaly scoring |

ระดับที่ต้องได้: **PCI-DSS Level 1** → ต้องมี **QSA (Qualified Security Assessor)** ทำ audit และออก **RoC (Report on Compliance)** ประจำปี + quarterly network scan โดย ASV

---

## 8. Non-functional Requirements

| ด้าน | เป้าหมาย |
|------|---------|
| Availability | ≥ 99.95% (payment core) |
| Auth latency (p99) | < 800 ms รวม network hop ไป acquirer |
| Throughput | ออกแบบให้ scale แนวนอน (stateless API + connection pool) |
| RPO / RTO | RPO ≤ 5 นาที (streaming replica), RTO ≤ 30 นาที |
| Data residency | เก็บข้อมูลในไทยตามข้อกำหนด ธปท./PDPA |
| Observability | metrics (Prometheus), tracing (OTel), log (JSON) |

---

## 9. ความเสี่ยงหลัก (Key Risks)

1. **เงินทุน & เวลา** — ใบอนุญาต + PCI L1 ใช้เงินและเวลาหลายเดือน (ดู ROADMAP)
2. **การเชื่อม acquirer/scheme** — ต้องมี sponsor bank; การรับรอง (certification) กับ Visa/MC ใช้เวลานาน
3. **Reconciliation mismatch** — ต้องมี ledger + งานกระทบยอดที่รัดกุมตั้งแต่แรก
4. **Fraud & chargeback** — ต้องมี dispute/chargeback workflow และ 3DS ตั้งแต่ต้น
5. **Compliance ต่อเนื่อง** — PCI ต่ออายุทุกปี + รายงาน ธปท. เป็นงวด

---

## 10. เอกสารที่เกี่ยวข้อง
- `COMPLIANCE-TH.md` — กฎหมายไทย ประเภทใบอนุญาต และขั้นตอนยื่นขอ
- `ROADMAP.md` — เฟส timeline และประมาณการ
- `DIAGRAMS.md` — context / component / sequence / ERD (mermaid)
- `../README.md` — โครงสร้างโปรเจ็คและวิธีรัน
