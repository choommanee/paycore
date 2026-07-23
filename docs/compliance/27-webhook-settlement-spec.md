# ข้อกำหนดทางเทคนิค Webhook และ Settlement (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 ต่อธนาคารแห่งประเทศไทย (ธปท.) และเป็นเอกสารประกอบการประเมิน PCI-DSS v4.0 Level 1
>
> เอกสารเลขที่: `COMP-27` · เวอร์ชัน 1.0 · เจ้าของเอกสาร: Head of Payment Engineering / CTO (ร่วมกับ Settlement & Reconciliation Lead)
> เอกสารอ้างอิง: `COMPLIANCE-TH.md`, `ARCHITECTURE.md` (§4, §5, §6, §8), `ROADMAP.md` (Phase 3), `docs/compliance/13-it-risk-management.md`, `docs/compliance/16-incident-response-breach.md`, `docs/compliance/20-network-segmentation-cde.md`
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงเทคนิค/นโยบายภายใน ไม่ใช่คำแนะนำทางกฎหมาย ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายและ QSA ก่อนยื่นจริง

---

> ### ⚠️ ข้อสมมติและสิ่งที่ยังต้องยืนยัน (Assumptions / TODO)
> รายการต่อไปนี้ยังขึ้นกับคู่สัญญา/ผู้ให้บริการภายนอกที่ยังไม่สรุป — ห้ามถือเป็นข้อเท็จจริงจนกว่าจะยืนยัน:
> - **[TODO — Sponsor Bank / Acquirer]** ยังไม่ลงนามธนาคารผู้รับเชื่อม (sponsoring bank) และ card scheme (Visa/Mastercard) — **รอบเวลา settlement จริง (settlement window/cut-off), สกุลเงิน settlement, ค่าธรรมเนียม interchange/scheme fee, รูปแบบไฟล์กระทบยอด (clearing/settlement file เช่น Visa TC33/VSS หรือ Mastercard IPM/T112) และ T+n ที่แท้จริง** จะกำหนดได้หลังลงนามและผ่าน scheme certification เท่านั้น ค่าที่ระบุในเอกสารนี้เป็น **ค่าตั้งต้นเชิงออกแบบ (design default)**
> - **[TODO — Local Rails / ITMX]** หากใช้ local settlement rails (เช่น ITMX สำหรับบัตรในประเทศ) รอบหักบัญชีและ cut-off ต้องยืนยันกับผู้ให้บริการ switching
> - **[TODO — QSA / ASV]** ยังไม่เลือกผู้ประเมิน PCI-DSS (QSA) — การรับรอง webhook payload ว่าไม่มี CHD/SAD และการควบคุม key management ต้องผ่านการทบทวนโดย QSA
> - **[TODO — ทุนจดทะเบียน]** ทุนจดทะเบียนชำระแล้วเป้าหมาย **50 ล้านบาท** (Full Acquiring) — ต้องยืนยันจำนวนที่ชำระจริงและรักษาไว้ ≥ 75% ตลอดการดำเนินงาน
> - **[TODO — ชื่อบริษัท / โดเมนจริง]** ชื่อ `[บริษัท / Company]`, โดเมน webhook (`https://webhook.[company].example` ในตัวอย่าง), และค่า threshold ปฏิบัติงานจริง ต้องเติมค่าจริงก่อนยื่นและก่อนการประเมิน on-site

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนดข้อกำหนดทางเทคนิคของ (ก) การลงลายเซ็น webhook ด้วย **HMAC-SHA256**, (ข) นโยบายการส่งซ้ำ (retry policy), (ค) ตารางการชำระบัญชีให้ร้านค้า (settlement schedule แบบ T+n) และ (ง) การกระทบยอด (reconciliation) ของ `[บริษัท / Company]` เพื่อรองรับข้อกำหนดด้าน **ความถูกต้องครบถ้วนของธุรกรรม (transaction integrity)**, **audit trail** ตามประกาศ ธปท. และ **PCI-DSS v4.0** (Req 4 การเข้ารหัสข้อมูลระหว่างส่ง, Req 10 การบันทึกและตรวจสอบ)

หลักการที่นำมาจาก `ARCHITECTURE.md`:
- **Ledger แบบ append-only (double-entry)** เป็น source of truth สำหรับ settlement และ reconciliation (§2, §6)
- **Idempotency** ทุก endpoint ที่ขยับเงิน (§2)
- **Webhook แบบ at-least-once + retry + ลงลายเซ็น** (§4)
- **Fail closed** — เมื่อไม่แน่ใจสถานะให้ถือว่ายังไม่สำเร็จ และ reconcile ภายหลัง (§2)
- **ห้ามมี CHD/SAD ใน payload** — webhook เห็นได้เพียง `card_brand` + `card_last4` (§6)

---

## 2. การลงลายเซ็น Webhook (HMAC-SHA256)

### 2.1 หลักการ
ทุก webhook ที่ `[บริษัท / Company]` ส่งไปยัง endpoint ของร้านค้า ต้องลงลายเซ็นด้วย **HMAC-SHA256** โดยใช้ **webhook signing secret** เฉพาะร้านค้า (per-merchant) เพื่อให้ร้านค้ายืนยันได้ว่า (ก) ข้อความมาจาก gateway จริง (authenticity) และ (ข) ไม่ถูกแก้ไขระหว่างทาง (integrity)

### 2.2 การประกอบข้อความที่ใช้เซ็น (signed payload)
```
signed_payload = timestamp + "." + raw_request_body
signature      = hex( HMAC_SHA256(signing_secret, signed_payload) )
```
- `timestamp` = Unix epoch (วินาที) ณ เวลาสร้างลายเซ็น
- `raw_request_body` = ไบต์ดิบของ JSON body (ต้องเซ็นและตรวจจากไบต์ดิบ ห้าม re-serialize)

### 2.3 HTTP headers ที่ส่งไปกับทุก webhook
| Header | ตัวอย่างค่า | ความหมาย |
|--------|-------------|----------|
| `Webhook-Id` | `evt_01HZX...` | รหัส event ไม่ซ้ำ (ใช้ทำ idempotency ฝั่งผู้รับ) |
| `Webhook-Timestamp` | `1753142400` | Unix epoch วินาที ใช้ตรวจ replay |
| `Webhook-Signature` | `v1=9f86d0...` | ลายเซ็น HMAC-SHA256 (รองรับหลายเวอร์ชันคั่นด้วยช่องว่าง เช่น `v1=... v1=...` ระหว่าง key rotation) |
| `Content-Type` | `application/json` | |
| `User-Agent` | `[Company]-Webhook/1.0` | |

### 2.4 ขั้นตอนการตรวจสอบฝั่งร้านค้า (ต้องระบุในเอกสาร integration)
1. อ่าน `Webhook-Timestamp` และเทียบกับเวลาปัจจุบัน — **ปฏิเสธถ้าเกิน ±5 นาที (300 วินาที)** เพื่อกัน replay attack
2. คำนวณ `HMAC_SHA256(signing_secret, timestamp + "." + raw_body)` ด้วย raw body
3. เปรียบเทียบกับค่าใน `Webhook-Signature` แบบ **constant-time compare** (กัน timing attack)
4. ตรวจ `Webhook-Id` กับที่เคยประมวลผล — ถ้าซ้ำให้ตอบ `200` แล้วหยุด (idempotent)

### 2.5 การจัดการ signing secret และ key rotation
| ประเด็น | นโยบาย |
|---------|--------|
| การสร้าง | สุ่มด้วย CSPRNG ความยาว ≥ 256-bit ต่อร้านค้า (per-merchant, per-endpoint) |
| การจัดเก็บ | เก็บแบบเข้ารหัส (envelope encryption ด้วย KMS) — ไม่เก็บ plaintext; แสดงเต็มครั้งเดียวตอนสร้างเท่านั้น |
| การหมุนคีย์ | รองรับ **overlapping keys** — ระหว่าง rotation ส่งลายเซ็นทั้งคีย์เก่าและใหม่ (ระยะทับซ้อน ≥ 24 ชม.) |
| ความถี่ | หมุนอย่างน้อยทุก **12 เดือน** และ **ทันที** เมื่อสงสัยว่ารั่ว |
| การเพิกถอน | ร้านค้าขอ rotate ได้ผ่าน Merchant API / Dashboard; ลง `audit_log` ทุกครั้ง |

### 2.6 ความปลอดภัยของช่องทาง
- ส่งผ่าน **HTTPS (TLS 1.2 ขึ้นไป)** เท่านั้น — ปฏิเสธ endpoint ที่เป็น `http://` (PCI Req 4)
- HMAC เป็น **การยืนยันตัวตนชั้นแอปพลิเคชัน** เพิ่มเติมจาก TLS ไม่ใช่ทดแทน
- **ห้ามใส่ CHD/SAD** ใน payload เด็ดขาด — payload มีได้เพียง id, สถานะ, จำนวนเงิน (minor units), `card_brand`, `card_last4`, timestamp (สอดคล้อง ARCHITECTURE §6)

---

## 3. นโยบายการส่งซ้ำ (Retry Policy)

### 3.1 หลักการส่งมอบ
Webhook เป็นแบบ **at-least-once delivery** — ร้านค้าต้องออกแบบ handler ให้ **idempotent** (ใช้ `Webhook-Id`) การส่งเข้าคิวจากตาราง `webhook_events` (ARCHITECTURE §6) ทำโดย worker แยก

### 3.2 เกณฑ์ว่าส่งสำเร็จ
- สำเร็จเมื่อได้ HTTP **2xx** ภายใน **timeout 10 วินาที**
- ถือว่าล้มเหลวเมื่อ: timeout, connection error, TLS error, หรือ HTTP **≥ 300** (รวม 3xx redirect ที่ไม่ตาม, 4xx, 5xx)

### 3.3 ตาราง backoff (exponential + jitter)
| ครั้งที่ | ระยะห่างโดยประมาณจากครั้งก่อน | เวลาสะสมโดยประมาณ |
|--------|-------------------------------|--------------------|
| 1 (ครั้งแรก) | ทันที | 0 |
| 2 | +1 นาที | ~1 นาที |
| 3 | +5 นาที | ~6 นาที |
| 4 | +30 นาที | ~36 นาที |
| 5 | +2 ชั่วโมง | ~2.5 ชั่วโมง |
| 6 | +6 ชั่วโมง | ~9 ชั่วโมง |
| 7–10 | +12 ชั่วโมง (ต่อครั้ง) | สูงสุด ~72 ชั่วโมง (3 วัน) |

- ใส่ **jitter** แบบสุ่ม ±20% กัน thundering herd
- ครบ **10 ครั้ง / 72 ชั่วโมง** แล้วยังไม่สำเร็จ → ตั้งสถานะ event เป็น `failed` และแจ้งเตือนร้านค้า (email/dashboard alert)

### 3.4 กลไกเสริม
| กลไก | รายละเอียด |
|------|-----------|
| **Dead-letter** | event ที่ล้มเหลวถาวรเก็บใน dead-letter สำหรับ replay ด้วยมือ (ต้องมีสิทธิ์ + ลง audit) |
| **Manual replay** | ร้านค้า/ผู้ดูแลสั่งส่งซ้ำผ่าน Dashboard/Admin API ได้ (ควบคุมด้วย RBAC) |
| **Circuit breaker** | ถ้า endpoint ร้านค้าล้มเหลวต่อเนื่อง ให้ลดอัตราส่งชั่วคราว (back-pressure) กันการถล่ม endpoint |
| **Ordering** | ไม่รับประกันลำดับ — ร้านค้าต้องใช้ timestamp/สถานะใน payload ตัดสิน แทนลำดับการมาถึง |
| **การเก็บบันทึก** | ทุกครั้งที่ส่ง (attempt no., response code, latency) บันทึกเพื่อ audit (PCI Req 10) เก็บอย่างน้อย 1 ปี online + 1 ปี archive (สอดคล้องนโยบาย log retention) |

---

## 4. ตารางการชำระบัญชี (Settlement Schedule — T+n)

### 4.1 นิยาม T+n
`T` = **วันทำการที่ธุรกรรมถูก capture และเข้ารอบ settlement (batch cut-off)** ไม่ใช่วันที่ authorize
`T+n` = จำนวน **วันทำการ (business days)** หลัง T ที่เงินสุทธิ (net) เข้าบัญชีร้านค้า โดยข้ามวันหยุดธนาคาร/วันหยุดนักขัตฤกษ์ตามปฏิทิน ธปท.

### 4.2 ตารางค่าตั้งต้นเชิงออกแบบ (design default — ต้องยืนยันกับ sponsor bank)
| ระดับความเสี่ยงร้านค้า (MCC/history) | รอบ settlement | หมายเหตุ |
|--------------------------------------|----------------|----------|
| ความเสี่ยงต่ำ (standard retail) | **T+2 วันทำการ** | ค่าตั้งต้นทั่วไป |
| ความเสี่ยงกลาง | T+3 ถึง T+5 วันทำการ | ตาม chargeback ratio / MCC |
| ความเสี่ยงสูง / ธุรกิจส่งมอบล่วงหน้า (เช่น ท่องเที่ยว, pre-order) | T+7 ขึ้นไป + **rolling reserve** | ถือ reserve เพื่อคุ้ม chargeback |
| ร้านค้าใหม่ (ช่วง onboarding แรก) | +hold ชั่วคราวตามนโยบายความเสี่ยง | ทบทวนหลังผ่านช่วงสังเกตการณ์ |

> **[TODO — Sponsor Bank]** ค่า T+n จริง ขึ้นกับรอบ funding ของ acquirer/scheme และ cut-off ของธนาคารผู้จ่าย

### 4.3 การคำนวณยอดสุทธิ (net settlement)
```
gross_captured       = Σ ยอด capture ในรอบ (minor units)
- refunds            = Σ refund ในรอบ
- chargebacks        = Σ chargeback/dispute ที่ตัดในรอบ
- processing_fee      = ค่าธรรมเนียม [บริษัท] (MDR)
- interchange/scheme  = ค่าธรรมเนียม network [TODO ยืนยันอัตรา]
- rolling_reserve     = ส่วนกันสำรอง (ถ้ามี)
+ reserve_release     = reserve ที่ครบกำหนดคืน
= net_payout          → โอนเข้าบัญชีร้านค้า
```
- ทุกยอดเป็น **integer minor units (สตางค์)** ห้ามใช้ float (ARCHITECTURE §2)
- ทุกองค์ประกอบลง `ledger_entries` (`entry_type` = capture/refund/fee/settlement) แบบ double-entry
- สร้าง **settlement statement** ต่อรอบต่อร้านค้า พร้อมรายการอ้างอิงกลับไปทุก payment

### 4.4 ขั้นตอนงาน (settlement worker — Phase 3)
1. Batch cut-off ณ เวลาที่กำหนด (เช่น 23:00 น. เวลาไทย) — **[TODO ยืนยัน cut-off กับ acquirer]**
2. รวมรายการ captured/refunded/chargeback ที่ยังไม่ settle
3. คำนวณ net ต่อร้านค้า → สร้าง settlement statement (สถานะ `pending`)
4. **กระทบยอดกับไฟล์ acquirer ก่อนจ่าย (ดู §5)** — จ่ายเฉพาะที่ตรงกัน (fail closed)
5. สร้างคำสั่งโอน (payout) → สถานะ `paid` เมื่อยืนยันการโอน
6. ยิง webhook `payout.paid` + ให้ statement ผ่าน Merchant API
7. ลง `audit_log` ทุกขั้น (maker-checker สำหรับการปล่อยจ่ายเกิน threshold)

### 4.5 การควบคุมการปล่อยจ่าย (payout controls)
| การควบคุม | นโยบาย |
|-----------|--------|
| Maker-checker | payout เกิน threshold (เช่น > 1,000,000 บาท/รอบ/ร้านค้า — **ค่าตั้งต้น ปรับได้**) ต้องมีผู้อนุมัติ 2 คน |
| Dual control | คำสั่งโอน batch ต้องอนุมัติโดยผู้มีสิทธิ์ต่างคนจากผู้สร้าง |
| AML/Sanction | ก่อนจ่าย ตรวจกับ sanction screening (ปปง./AMLO, OFAC/UN lists) — ดู `06-sanctions-screening.md`; พบ hit → ระงับ + escalate |
| Hold / reserve | ระงับหรือหักสำรองตามนโยบายความเสี่ยง (rolling reserve) |
| ธุรกรรมผิดปกติ | ยอดผิดปกติ/spike เข้าเกณฑ์รายงานตาม `07-sar-str-procedure.md` |

---

## 5. การกระทบยอด (Reconciliation)

### 5.1 หลักการ
กระทบยอด **3 ทาง (three-way match)** ทุกวันทำการ:
```
Internal Ledger (ledger_entries)  ⇄  Acquirer/Scheme settlement file  ⇄  Bank statement (เงินเข้าจริง)
```
ledger ภายในเป็น source of truth เชิงบัญชี — แต่ **เงินจริงต้องตรงกับที่ acquirer และธนาคารยืนยัน** ก่อนปิดรอบ

### 5.2 รอบและขอบเขต
| รอบ | สิ่งที่ทำ |
|-----|----------|
| **รายวัน (T+1 เช้า)** | ดึงไฟล์ settlement/clearing จาก acquirer → จับคู่รายการต่อรายการกับ `ledger_entries` ด้วย `acquirer_ref`/`auth_code`/RRN |
| **ต่อรอบ settlement** | กระทบ net payout กับ bank statement เงินเข้าจริง |
| **รายเดือน** | สรุปค่าธรรมเนียม, reserve, chargeback, ยอดคงค้าง; รายงานผู้บริหาร/ธปท. ตามงวด |

### 5.3 ประเภทความไม่ตรง (exceptions) และการจัดการ
| ประเภท | คำอธิบาย | การจัดการ |
|--------|----------|-----------|
| **Missing at acquirer** | มีใน ledger แต่ไม่มีในไฟล์ acquirer | สอบสวน; อาจ auth ไม่ถูก settle → void/expire |
| **Missing internally** | มีในไฟล์ acquirer แต่ไม่มีใน ledger | ตรวจ webhook/authorization ที่หลุด → บันทึกย้อนพร้อม audit |
| **Amount mismatch** | ยอดต่างกัน | ตรวจ FX/fee/partial capture; แก้ด้วย adjustment entry |
| **Duplicate** | รายการซ้ำ | ตรวจ idempotency; ตัดซ้ำ |
| **Fee mismatch** | ค่าธรรมเนียมต่างจากที่คำนวณ | ทบทวนอัตรา; ปรับปรุงตาราง fee |
| **Chargeback/dispute** | รายการโต้แย้งจาก issuer | เข้าสู่ dispute workflow; ปรับ ledger |

### 5.4 SLA และการยกระดับ (escalation)
| ระดับ | เกณฑ์ | เจ้าของ | กรอบเวลา |
|-------|-------|---------|----------|
| L1 | รายการ mismatch < threshold | Settlement/Reconciliation Analyst | แก้ภายในวันทำการเดียวกัน |
| L2 | mismatch มูลค่าสูง / เกินจำนวนที่กำหนด | Reconciliation Lead + Finance | ภายใน 2 วันทำการ |
| L3 | สงสัยทุจริต / ระบบผิดพลาดวงกว้าง | CFO / CISO + Incident Response (`16-incident-response-breach.md`) | เปิดเคสทันที |

- **นโยบาย fail closed:** ห้ามปล่อย payout สำหรับรายการที่ยัง reconcile ไม่ผ่าน
- **Segregation of duties:** ผู้ทำ reconciliation ต้องแยกจากผู้อนุมัติ payout
- ผลกระทบยอดและ exception ทั้งหมดลง `audit_log` (append-only) เพื่อ audit trail ตาม ธปท./PCI Req 10

### 5.5 บทบาทและความรับผิดชอบ (RACI ย่อ)
| กิจกรรม | R | A | C | I |
|--------|---|---|---|---|
| ดึง/นำเข้าไฟล์ acquirer | Settlement Engineer | Recon Lead | — | Finance |
| จับคู่ + จัดการ exception | Recon Analyst | Recon Lead | Payment Eng | Finance |
| อนุมัติ payout | Finance Approver | CFO | Compliance | Merchant |
| รายงาน ธปท. เป็นงวด | Compliance | CFO | Legal | ธปท. |

---

## 6. การเก็บรักษาบันทึกและการตรวจสอบ (Records & Audit)
- **Webhook delivery log** (ทุก attempt, response, latency): online ≥ 1 ปี, archive รวม ≥ 2 ปี
- **Settlement statements & reconciliation reports:** เก็บตามระยะเวลาที่ ธปท./กฎหมายบัญชีกำหนด (โดยทั่วไป ≥ 5–10 ปี — **[TODO ยืนยันกับที่ปรึกษากฎหมาย]**)
- **audit_log** (append-only): ทุก state change ของ payment/settlement/payout (ARCHITECTURE §6)
- บันทึกทั้งหมด **ห้ามมี CHD/SAD** และปกป้องด้วย access control + integrity monitoring (PCI Req 10)

---

# Technical spec: webhook signing (HMAC-SHA256), retry policy, settlement schedule (T+n), reconciliation (English)

> Supporting document for the **Electronic Payment Acquiring Service (Full Acquiring)** licence application
> under the Payment Systems Act B.E. 2560 to the Bank of Thailand (BOT), and supporting evidence for PCI-DSS v4.0 Level 1.
>
> Document no.: `COMP-27` · Version 1.0 · Owner: Head of Payment Engineering / CTO (with Settlement & Reconciliation Lead)
> References: `COMPLIANCE-TH.md`, `ARCHITECTURE.md` (§4, §5, §6, §8), `ROADMAP.md` (Phase 3), `docs/compliance/13-it-risk-management.md`, `docs/compliance/16-incident-response-breach.md`, `docs/compliance/20-network-segmentation-cde.md`
>
> **Note:** This is an internal technical/policy document, not legal advice. It must be reviewed by legal counsel and the QSA before formal submission.

---

> ### ⚠️ Assumptions / TODO
> The following depend on external counterparties not yet finalised — do NOT treat as fact until confirmed:
> - **[TODO — Sponsor Bank / Acquirer]** Sponsoring bank and card scheme (Visa/Mastercard) not yet signed — the **actual settlement window/cut-off, settlement currency, interchange/scheme fees, clearing/settlement file formats (e.g. Visa TC33/VSS, Mastercard IPM/T112) and true T+n** can only be set after signature and scheme certification. Values here are **design defaults**.
> - **[TODO — Local Rails / ITMX]** If local settlement rails (e.g. ITMX for domestic cards) are used, clearing cycles and cut-offs must be confirmed with the switching provider.
> - **[TODO — QSA / ASV]** PCI-DSS assessor (QSA) not yet engaged — attestation that webhook payloads carry no CHD/SAD and the key-management controls must be QSA-reviewed.
> - **[TODO — Paid-up capital]** Target paid-up registered capital **THB 50M** (Full Acquiring) — actual paid amount to be confirmed and maintained at ≥ 75% at all times.
> - **[TODO — Company name / real domains]** `[บริษัท / Company]`, the webhook domain (`https://webhook.[company].example` in examples), and operational thresholds must be populated before submission and the on-site assessment.

---

## 1. Purpose & Scope

This document specifies the technical requirements for (a) webhook signing with **HMAC-SHA256**, (b) the **retry policy**, (c) the merchant **settlement schedule (T+n)**, and (d) **reconciliation** for `[บริษัท / Company]`, in support of BOT transaction-integrity/audit-trail requirements and **PCI-DSS v4.0** (Req 4 – encryption in transit, Req 10 – logging and monitoring).

Principles inherited from `ARCHITECTURE.md`:
- **Append-only double-entry ledger** is the source of truth for settlement and reconciliation (§2, §6)
- **Idempotency** on every money-moving endpoint (§2)
- **Webhooks are at-least-once + retry + signed** (§4)
- **Fail closed** — if status is uncertain, treat as not-successful and reconcile later (§2)
- **No CHD/SAD in payloads** — only `card_brand` + `card_last4` are exposed (§6)

---

## 2. Webhook Signing (HMAC-SHA256)

### 2.1 Principle
Every webhook `[บริษัท / Company]` sends to a merchant endpoint is signed with **HMAC-SHA256** using a per-merchant **webhook signing secret**, allowing the merchant to verify (a) the message truly originates from the gateway (authenticity) and (b) it was not tampered with in transit (integrity).

### 2.2 Signed payload construction
```
signed_payload = timestamp + "." + raw_request_body
signature      = hex( HMAC_SHA256(signing_secret, signed_payload) )
```
- `timestamp` = Unix epoch (seconds) at signing time
- `raw_request_body` = the raw JSON body bytes (sign and verify against raw bytes; never re-serialize)

### 2.3 HTTP headers sent with every webhook
| Header | Example | Meaning |
|--------|---------|---------|
| `Webhook-Id` | `evt_01HZX...` | Unique event id (receiver-side idempotency key) |
| `Webhook-Timestamp` | `1753142400` | Unix epoch seconds, used for replay protection |
| `Webhook-Signature` | `v1=9f86d0...` | HMAC-SHA256 signature (multiple space-separated versions supported during key rotation, e.g. `v1=... v1=...`) |
| `Content-Type` | `application/json` | |
| `User-Agent` | `[Company]-Webhook/1.0` | |

### 2.4 Merchant-side verification steps (must be documented in the integration guide)
1. Read `Webhook-Timestamp` and compare to now — **reject if outside ±5 minutes (300s)** to prevent replay.
2. Compute `HMAC_SHA256(signing_secret, timestamp + "." + raw_body)` over the raw body.
3. Compare to `Webhook-Signature` using a **constant-time comparison** (defeats timing attacks).
4. Check `Webhook-Id` against processed ids — if seen, return `200` and stop (idempotent).

### 2.5 Signing-secret management & key rotation
| Item | Policy |
|------|--------|
| Generation | CSPRNG, ≥ 256-bit, per-merchant and per-endpoint |
| Storage | Encrypted (envelope encryption via KMS) — no plaintext at rest; shown in full only once at creation |
| Rotation | **Overlapping keys** — during rotation, sign with both old and new keys (overlap ≥ 24h) |
| Frequency | At least every **12 months**, and **immediately** on suspected compromise |
| Revocation | Merchant can rotate via Merchant API / Dashboard; every action written to `audit_log` |

### 2.6 Channel security
- Delivered over **HTTPS (TLS 1.2+) only** — `http://` endpoints are rejected (PCI Req 4).
- HMAC is **application-layer authentication** in addition to, not a replacement for, TLS.
- **No CHD/SAD in the payload** — only id, status, amount (minor units), `card_brand`, `card_last4`, timestamps (per ARCHITECTURE §6).

---

## 3. Retry Policy

### 3.1 Delivery guarantee
Webhooks are **at-least-once**; merchants must build **idempotent** handlers keyed on `Webhook-Id`. Delivery is dispatched from the `webhook_events` table (ARCHITECTURE §6) by a dedicated worker.

### 3.2 Success criteria
- Success = HTTP **2xx** within a **10-second timeout**.
- Failure = timeout, connection error, TLS error, or HTTP **≥ 300** (including un-followed 3xx redirects, 4xx, 5xx).

### 3.3 Backoff schedule (exponential + jitter)
| Attempt | Approx. delay from previous | Approx. cumulative |
|--------|------------------------------|--------------------|
| 1 (initial) | immediate | 0 |
| 2 | +1 min | ~1 min |
| 3 | +5 min | ~6 min |
| 4 | +30 min | ~36 min |
| 5 | +2 h | ~2.5 h |
| 6 | +6 h | ~9 h |
| 7–10 | +12 h (each) | up to ~72 h (3 days) |

- Apply random **jitter** of ±20% to avoid thundering-herd.
- After **10 attempts / 72 hours** without success → mark the event `failed` and alert the merchant (email/dashboard).

### 3.4 Supporting mechanisms
| Mechanism | Detail |
|-----------|--------|
| **Dead-letter** | Permanently failed events go to a dead-letter store for manual replay (requires privilege + audit) |
| **Manual replay** | Merchant/operator can re-send via Dashboard/Admin API (RBAC-controlled) |
| **Circuit breaker** | If a merchant endpoint fails repeatedly, throttle delivery temporarily (back-pressure) |
| **Ordering** | Ordering is not guaranteed — merchants must use payload timestamp/status, not arrival order |
| **Logging** | Every attempt (attempt no., response code, latency) is logged for audit (PCI Req 10); retain ≥ 1 year online + ≥ 1 year archive |

---

## 4. Settlement Schedule (T+n)

### 4.1 Definition of T+n
`T` = the **business day the transaction is captured and enters the settlement batch (cut-off)** — not the authorize date.
`T+n` = number of **business days** after T when net funds land in the merchant account, skipping bank/public holidays per the BOT calendar.

### 4.2 Design-default table (must be confirmed with the sponsor bank)
| Merchant risk tier (MCC/history) | Settlement cycle | Notes |
|----------------------------------|------------------|-------|
| Low risk (standard retail) | **T+2 business days** | General default |
| Medium risk | T+3 to T+5 business days | Per chargeback ratio / MCC |
| High risk / deferred delivery (e.g. travel, pre-order) | T+7+ and **rolling reserve** | Reserve held to cover chargebacks |
| New merchants (initial onboarding) | Temporary hold per risk policy | Reviewed after observation period |

> **[TODO — Sponsor Bank]** Actual T+n depends on the acquirer/scheme funding cycle and the paying bank's cut-off.

### 4.3 Net settlement calculation
```
gross_captured       = Σ captures in cycle (minor units)
- refunds            = Σ refunds in cycle
- chargebacks        = Σ chargebacks/disputes debited in cycle
- processing_fee      = [Company] fee (MDR)
- interchange/scheme  = network fees [TODO confirm rates]
- rolling_reserve     = reserve withheld (if any)
+ reserve_release     = reserves due for release
= net_payout          → transferred to merchant account
```
- All amounts are **integer minor units (satang)** — no floats (ARCHITECTURE §2).
- Every component is written to `ledger_entries` (`entry_type` = capture/refund/fee/settlement) as double entries.
- Produce a **settlement statement** per cycle per merchant, cross-referencing every payment.

### 4.4 Settlement worker flow (Phase 3)
1. Batch cut-off at a fixed time (e.g. 23:00 ICT) — **[TODO confirm cut-off with acquirer]**.
2. Aggregate not-yet-settled captured/refunded/chargeback items.
3. Compute per-merchant net → create statement (status `pending`).
4. **Reconcile against the acquirer file before paying (see §5)** — pay only matched items (fail closed).
5. Create payout instruction → status `paid` once the transfer is confirmed.
6. Emit `payout.paid` webhook + expose the statement via Merchant API.
7. Write every step to `audit_log` (maker-checker for payouts above threshold).

### 4.5 Payout controls
| Control | Policy |
|---------|--------|
| Maker-checker | Payouts above threshold (e.g. > THB 1,000,000 per cycle per merchant — **default, tunable**) need two approvers |
| Dual control | Batch transfer instructions approved by someone other than the creator |
| AML/Sanctions | Before payout, screen against sanctions lists (AMLO/ปปง., OFAC/UN) — see `06-sanctions-screening.md`; on a hit → block + escalate |
| Hold / reserve | Withhold or reserve per risk policy (rolling reserve) |
| Anomalous flows | Unusual amounts/spikes meeting thresholds reported per `07-sar-str-procedure.md` |

---

## 5. Reconciliation

### 5.1 Principle
Perform a **three-way match** every business day:
```
Internal Ledger (ledger_entries)  ⇄  Acquirer/Scheme settlement file  ⇄  Bank statement (actual funds in)
```
The internal ledger is the accounting source of truth — but **real money must match what the acquirer and bank confirm** before a cycle is closed.

### 5.2 Cycles & scope
| Cycle | Activity |
|-------|----------|
| **Daily (T+1 morning)** | Pull acquirer settlement/clearing file → match line-by-line to `ledger_entries` via `acquirer_ref`/`auth_code`/RRN |
| **Per settlement cycle** | Match net payout to actual funds on the bank statement |
| **Monthly** | Summarise fees, reserves, chargebacks, outstanding balances; periodic reports to management/BOT |

### 5.3 Exception types & handling
| Type | Description | Handling |
|------|-------------|----------|
| **Missing at acquirer** | In ledger, absent from acquirer file | Investigate; auth may be unsettled → void/expire |
| **Missing internally** | In acquirer file, absent from ledger | Check dropped webhook/authorization → back-post with audit |
| **Amount mismatch** | Amounts differ | Check FX/fee/partial capture; correct via adjustment entry |
| **Duplicate** | Duplicated line | Check idempotency; de-duplicate |
| **Fee mismatch** | Fee differs from computed | Review rates; update fee table |
| **Chargeback/dispute** | Issuer dispute | Enter dispute workflow; adjust ledger |

### 5.4 SLA & escalation
| Level | Trigger | Owner | Timeframe |
|-------|---------|-------|-----------|
| L1 | Mismatch < threshold | Settlement/Reconciliation Analyst | Same business day |
| L2 | High-value / over-threshold mismatch | Reconciliation Lead + Finance | Within 2 business days |
| L3 | Suspected fraud / wide-scale error | CFO / CISO + Incident Response (`16-incident-response-breach.md`) | Open case immediately |

- **Fail-closed policy:** never release payout for items that have not passed reconciliation.
- **Segregation of duties:** reconciliation performers are separate from payout approvers.
- All reconciliation results and exceptions are written to `audit_log` (append-only) for the BOT/PCI Req 10 audit trail.

### 5.5 Roles & responsibilities (condensed RACI)
| Activity | R | A | C | I |
|----------|---|---|---|---|
| Ingest acquirer file | Settlement Engineer | Recon Lead | — | Finance |
| Match + handle exceptions | Recon Analyst | Recon Lead | Payment Eng | Finance |
| Approve payout | Finance Approver | CFO | Compliance | Merchant |
| Periodic BOT reporting | Compliance | CFO | Legal | BOT |

---

## 6. Records & Audit
- **Webhook delivery logs** (every attempt, response, latency): ≥ 1 year online, ≥ 2 years total including archive.
- **Settlement statements & reconciliation reports:** retained per BOT/accounting-law requirements (generally ≥ 5–10 years — **[TODO confirm with legal counsel]**).
- **audit_log** (append-only): every payment/settlement/payout state change (ARCHITECTURE §6).
- All records contain **no CHD/SAD** and are protected by access controls + integrity monitoring (PCI Req 10).
