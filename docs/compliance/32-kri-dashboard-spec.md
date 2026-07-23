# ข้อกำหนด Dashboard ตัวชี้วัดความเสี่ยง (KRI) (ไทย)

> เอกสารเลขที่ **32** ในชุดเอกสาร Compliance สำหรับการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** และคู่ขนานกับ **PCI-DSS Level 1**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อยื่นขออนุมัติ (submission draft) · เวอร์ชัน 0.1 · วันที่ปรับปรุง 2026-07-22
> **เจ้าของเอกสาร:** ประธานเจ้าหน้าที่บริหารความเสี่ยง (CRO) ร่วมกับ CISO และ Head of Payment Operations ของ **[บริษัท / Company]**
> **ผู้ใช้เอกสาร:** คณะกรรมการบริหารความเสี่ยง (Risk Committee), คณะทำงาน IT Risk & Security, ทีม Fraud & Risk Operations, MLRO/AMLCO, Internal Audit
>
> เอกสารนี้เป็นข้อกำหนดเชิงระบบ (specification) และเอกสารประกอบคำขอใบอนุญาต **มิใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจทานโดยที่ปรึกษากฎหมาย, CRO, CISO, MLRO และ QSA ก่อนบังคับใช้จริง เนื่องจากประกาศ/หลักเกณฑ์ด้าน IT risk, cyber resilience และการกำกับ acquirer ของ ธปท. อาจปรับปรุงได้

---

## บทสรุปสำหรับผู้บริหาร (Executive Summary)

**[บริษัท / Company]** ในฐานะผู้ให้บริการรับชำระเงินด้วยบัตร (acquiring payment gateway) จำเป็นต้องเฝ้าระวังความเสี่ยงด้านการดำเนินงาน (operational), การฉ้อโกง (fraud), การฟอกเงิน (AML) และเสถียรภาพของระบบ (availability) แบบใกล้เวลาจริง (near-real-time) เอกสารนี้กำหนด **ข้อกำหนดของ Dashboard ตัวชี้วัดความเสี่ยงหลัก (Key Risk Indicator — KRI Dashboard)** ที่รวบรวมตัวชี้วัด 4 กลุ่มหลักและตัวชี้วัดสนับสนุน เพื่อให้ผู้บริหารและคณะกรรมการมองเห็นสถานะความเสี่ยงต่อ **Risk Appetite** ที่คณะกรรมการอนุมัติ

ตัวชี้วัดหลัก 4 กลุ่มตามขอบเขตเอกสารนี้:

1. **ปริมาณธุรกรรม (Transaction Volume)** — จำนวนรายการและมูลค่า (บาท) เพื่อตรวจจับความผิดปกติเชิงปริมาณและ capacity risk
2. **อัตรา authorization ล้มเหลว (Failed-Authorization Rate)** — สัดส่วนการปฏิเสธ/ล้มเหลว บ่งชี้ปัญหาเทคนิค, การโจมตี (card testing/enumeration) และคุณภาพ routing
3. **อัตราการคืนเงิน/ข้อโต้แย้ง (Refund / Chargeback Ratio)** — สัดส่วน refund และ chargeback ต่อยอดขาย ตัวชี้วัดความเสี่ยง merchant, การฉ้อโกง และการเข้าโครงการเฝ้าระวังของแบรนด์บัตร
4. **ผลขาดทุนจากการฉ้อโกง (Fraud Loss)** — มูลค่าและอัตราส่วน (bps) ของความสูญเสียจากการฉ้อโกงที่ยืนยันแล้ว

Dashboard นี้เป็น **ชั้นการเฝ้าระวัง (monitoring layer)** ที่ดึงข้อมูลจากตารางในสถาปัตยกรรม (`payments`, `ledger_entries`, `refunds`, `webhook_events`, `audit_log`) และผูกกับกลไก **สัญญาณเตือน (alerting)** และ **การยกระดับ (escalation)** ที่ระบุไว้ด้านล่าง สอดคล้องกับกรอบ Three Lines of Defense และการรายงานความเสี่ยงต่อคณะกรรมการในเอกสาร 13 (IT Risk Management) และ 05/07 (AML/SAR)

---

## 1. ฐานกฎหมายและมาตรฐานอ้างอิง

| กฎหมาย / มาตรฐาน | สาระสำคัญที่เกี่ยวข้องกับ KRI Dashboard |
|---|---|
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** | เงื่อนไขใบอนุญาตกำหนดให้ผู้ให้บริการต้องมีระบบบริหารความเสี่ยง การควบคุมภายใน และการรายงานที่ได้มาตรฐาน กำกับโดย **ธปท.** — KRI Dashboard เป็นหลักฐานการติดตามความเสี่ยงเชิงปริมาณ |
| **แนวนโยบาย/ประกาศ ธปท. ด้าน IT Risk & Cyber Resilience** | กำหนดให้มีการเฝ้าระวัง ตัวชี้วัด และการรายงานต่อคณะกรรมการ รวมถึงการบริหารความจุ (capacity) และเหตุการณ์ผิดปกติ |
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (ปปง./AMLO)** | KRI ด้าน velocity, structuring, และธุรกรรมผิดปกติ ป้อนเข้าสู่กระบวนการ SAR/STR (เชื่อมกับเอกสาร 05, 07) |
| **PCI-DSS v4.0** — Req 10 & 12.10 | Logging, การเฝ้าระวัง และการตรวจจับความผิดปกติ; ข้อมูลบน dashboard ห้ามมี PAN/CHD; รายงานเหตุจากตัวชี้วัดต้องเชื่อมกับ incident response |
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — **PDPC** | ข้อมูลรวม (aggregate) บน dashboard ต้องผ่านการลดทอนข้อมูลส่วนบุคคล (data minimization) และควบคุมการเข้าถึงตามหลักการ least privilege |
| **EMV 3-D Secure (3DS) 2.x** | อัตรา authentication, challenge rate และผล frictionless เป็นอินพุตของ KRI ด้าน fraud และ failed-auth |
| **Visa VAMP / Mastercard Excessive Chargeback Program (ECP/ECM)** | เกณฑ์ (threshold) ของ chargeback/fraud ที่แบรนด์บัตรใช้จัดโปรแกรมเฝ้าระวัง — เป็นฐานในการตั้งค่า threshold ของ KRI |
| **ISO/IEC 27001:2022, NIST CSF 2.0** | กรอบการตรวจจับ (Detect) และการวัดประสิทธิผลการควบคุม |

> **[TODO/ข้อสมมติ]** ต้องยืนยันกับที่ปรึกษากฎหมายและ ธปท. ว่าประกาศ/แนวปฏิบัติด้าน IT risk และการรายงานฉบับล่าสุด ณ วันยื่นคำขอ มีเลขที่/ปีใด และปรับอ้างอิงให้ตรง

> **[TODO/ข้อสมมติ]** เกณฑ์ (threshold) ของ **Visa VAMP** และ **Mastercard ECP/ECM** อ้างอิงตามเอกสารสาธารณะของแบรนด์บัตร ณ ปัจจุบัน — ต้องยืนยันตัวเลขที่บังคับใช้จริงกับ **sponsor bank / scheme** เมื่อทำสัญญา (ดูเอกสาร 24 Scheme Certification)

---

## 2. ขอบเขต วัตถุประสงค์ และหลักการออกแบบ

**วัตถุประสงค์:** ให้มุมมองเดียว (single pane of glass) ต่อความเสี่ยงเชิงปริมาณ เพื่อ (1) ตรวจจับสัญญาณเตือนล่วงหน้า (early warning) ก่อนความเสี่ยงกลายเป็นเหตุการณ์, (2) เปรียบเทียบสถานะกับ Risk Appetite, (3) สนับสนุนการรายงานต่อ ธปท. และคณะกรรมการ, (4) ป้อนหลักฐานให้กระบวนการ AML/SAR และ chargeback management

**ขอบเขต:** ครอบคลุมตัวชี้วัด 4 กลุ่มหลัก (txn volume, failed-auth rate, refund/chargeback ratio, fraud loss) และตัวชี้วัดสนับสนุน (3DS authentication rate, decline reason breakdown, velocity, system availability/latency) ทั้งในระดับ **portfolio รวม**, ระดับ **merchant**, และระดับ **BIN/ประเทศ/MCC/ช่องทาง**

### หลักการออกแบบ (Design Principles)

1. **ไม่มี CHD บน dashboard** — แสดงได้เพียง `card_brand`, `card_last4`, `bin (6-8 หลักแรก)`, MCC, ประเทศ; ห้ามแสดง PAN/CVV/PIN เด็ดขาด (สอดคล้อง ARCHITECTURE.md ข้อ 6)
2. **แหล่งข้อมูลเดียวที่เชื่อถือได้ (single source of truth)** — คำนวณจาก `ledger_entries` (append-only) และ `payments` เป็นหลัก เพื่อให้กระทบยอดกับ reconciliation ได้
3. **เงินเป็น integer minor units (สตางค์)** — คำนวณอัตราส่วนด้วย decimal ไม่ใช้ float
4. **แยกชั้นข้อมูล (read replica)** — dashboard/analytics อ่านจาก replica แยกจาก payment core เพื่อไม่กระทบ latency ของ authorization
5. **Traffic-light + threshold** — ทุก KRI มีสถานะ เขียว/เหลือง/แดง (Green/Amber/Red) เทียบกับเกณฑ์ที่คณะกรรมการอนุมัติ
6. **Auditability** — การเปลี่ยน threshold, การรับทราบ alert (acknowledge) และการยกเว้น (risk acceptance) ทุกครั้งลง `audit_log`
7. **Near-real-time + batch** — ตัวชี้วัดเชิงปฏิบัติการ (volume, failed-auth, velocity) รีเฟรชแบบ streaming/นาที; ตัวชี้วัดที่ต้องรอ settlement (chargeback, fraud loss ยืนยัน) รีเฟรชแบบ batch รายวัน

---

## 3. โครงสร้างการกำกับดูแลและบทบาท (Governance & RACI)

| บทบาท | ความรับผิดชอบต่อ KRI Dashboard |
|---|---|
| **คณะกรรมการบริษัท (Board)** | อนุมัติ Risk Appetite และเกณฑ์ KRI ระดับ Red; รับรายงานสรุปรายไตรมาส |
| **คณะกรรมการบริหารความเสี่ยง (Risk Committee)** | ทบทวน KRI รายไตรมาส, อนุมัติการเปลี่ยน threshold, กำกับแผนแก้ไข (remediation) |
| **CRO (เจ้าของเอกสาร)** | **Accountable** ต่อกรอบ KRI ทั้งหมด; รายงานต่อ Risk Committee |
| **CISO** | **Responsible** ต่อ KRI ด้าน availability/latency/security และความถูกต้องของ log |
| **Head of Payment Operations** | **Responsible** ต่อ KRI ด้าน volume, failed-auth, refund/chargeback |
| **Fraud & Risk Operations (analyst)** | เฝ้าระวังรายวัน, triage alert, ดำเนินการชั้นแรก, บันทึกผล |
| **MLRO / AMLCO** | รับสัญญาณจาก KRI ด้าน velocity/structuring ป้อนเข้ากระบวนการ SAR/STR (เอกสาร 07) |
| **Internal Audit** | ตรวจสอบอิสระต่อความถูกต้องของสูตร, การตั้ง threshold และการตอบสนอง alert |
| **Data/Platform Engineering** | **Responsible** ต่อ pipeline, ความถูกต้องของข้อมูล, uptime ของ dashboard |

### 3.1 รอบการทบทวน (Review Cadence)

| กิจกรรม | ความถี่ | ผู้รับผิดชอบ |
|---|---|---|
| เฝ้าระวัง real-time + triage alert | ต่อเนื่อง (24/7 on-call) | Fraud & Risk Ops |
| สรุปสถานะ KRI ประจำวัน (daily stand-up) | รายวัน | Payment Ops + Fraud Ops |
| ทบทวน KRI + เหตุการณ์ | รายเดือน (IT Risk & Security Committee) | CISO/CRO |
| รายงานความเสี่ยงต่อ Risk Committee | รายไตรมาส | CRO |
| ทบทวน/ปรับ threshold + backtest | รายไตรมาส (หรือเมื่อมีเหตุ) | CRO + Fraud Ops + Audit |

---

## 4. ข้อกำหนด KRI หลัก (Core KRI Definitions)

ทุก KRI ระบุ: นิยาม, สูตร, แหล่งข้อมูล, มิติการซอย (dimensions), ความถี่รีเฟรช, และเกณฑ์ Green/Amber/Red

### 4.1 KRI-01 — ปริมาณธุรกรรม (Transaction Volume)

- **นิยาม:** จำนวนรายการ authorization (count) และมูลค่ารวม (บาท) ในช่วงเวลา พร้อมการเปรียบเทียบกับค่าคาดการณ์ (baseline)
- **สูตร:**
  - `txn_count = COUNT(payments WHERE created_at ∈ window)`
  - `txn_value_thb = SUM(amount_minor)/100 WHERE ...`
  - `volume_deviation = (actual − forecast_baseline) / forecast_baseline`
- **แหล่งข้อมูล:** `payments`, `ledger_entries(authorize)`
- **มิติ:** merchant, MCC, ช่องทาง (card-present/CNP), ประเทศ BIN, เวลา (นาที/ชม./วัน)
- **ความถี่:** streaming (รีเฟรช ≤ 1 นาที)
- **เกณฑ์:** ดูตารางข้อ 5

> ความผิดปกติเชิงปริมาณ (spike/drop) อาจบ่งชี้: การโจมตี card-testing, merchant ผิดปกติ, ระบบล่ม, หรือ capacity risk ต่อ SLA (p99 < 800 ms ตาม ARCHITECTURE.md ข้อ 8)

### 4.2 KRI-02 — อัตรา authorization ล้มเหลว (Failed-Authorization Rate)

- **นิยาม:** สัดส่วนของ authorization ที่ถูกปฏิเสธหรือล้มเหลว (decline/error/timeout) ต่อจำนวนที่พยายามทั้งหมด
- **สูตร:** `failed_auth_rate = (declined + errored + timed_out) / total_auth_attempts`
- **แยกย่อยตามเหตุ (decline reason):** insufficient funds, do-not-honor, invalid card, suspected fraud, 3DS failed, technical/timeout
- **แหล่งข้อมูล:** `payments(status=failed)`, response code จาก acquirer/3DS
- **มิติ:** merchant, BIN, issuer country, decline reason, ช่องทาง
- **ความถี่:** streaming (≤ 1 นาที) + ค่าเฉลี่ยเคลื่อนที่ 15 นาที
- **สัญญาณเฉพาะ:** อัตราสูงผิดปกติเฉพาะ `invalid card` + volume พุ่ง = สัญญาณ **card enumeration/BIN attack** → ยกระดับทันที

### 4.3 KRI-03 — อัตราการคืนเงิน/ข้อโต้แย้ง (Refund / Chargeback Ratio)

- **นิยาม:**
  - `refund_ratio` = มูลค่า refund ต่อมูลค่าที่ capture
  - `chargeback_ratio (count)` = จำนวน chargeback ต่อจำนวนธุรกรรมที่ settle
  - `chargeback_ratio (value)` = มูลค่า chargeback ต่อมูลค่าที่ settle
- **สูตร:**
  - `refund_ratio = SUM(refund_amount)/SUM(captured_amount)`
  - `cb_ratio_count = COUNT(chargebacks)/COUNT(settled_txns)`
  - `cb_ratio_value = SUM(chargeback_amount)/SUM(settled_amount)`
- **แหล่งข้อมูล:** `refunds`, `ledger_entries(refund/capture/settlement)`, chargeback feed จาก acquirer/scheme
- **มิติ:** merchant (สำคัญที่สุด), MCC, BIN, reason code (fraud/dispute/processing error)
- **ความถี่:** batch รายวัน (chargeback ตามรอบ settlement) + rolling 30/60/120 วันเพื่อเทียบเกณฑ์แบรนด์
- **ผูกกับโปรแกรมแบรนด์บัตร:** เทียบกับเกณฑ์ **Visa VAMP** และ **Mastercard ECP/ECM** เพื่อเตือนก่อนถึงระดับที่แบรนด์เข้าเฝ้าระวัง (ค่าปรับ/บทลงโทษ)

### 4.4 KRI-04 — ผลขาดทุนจากการฉ้อโกง (Fraud Loss)

- **นิยาม:** มูลค่าความสูญเสียสุทธิจากการฉ้อโกงที่ยืนยันแล้ว และอัตราส่วนต่อยอดขาย (แสดงเป็น bps = จุดเบสิส)
- **สูตร:**
  - `fraud_loss_thb = SUM(confirmed_fraud_amount − recovered)`
  - `fraud_bps = (fraud_loss_thb / total_sales_thb) × 10,000`
  - `fraud_rate_count = COUNT(confirmed_fraud_txns)/COUNT(total_txns)`
- **แหล่งข้อมูล:** case management (fraud confirmed), chargeback reason=fraud, ผล 3DS liability shift
- **มิติ:** merchant, MCC, BIN, ช่องทาง, ประเภทฉ้อโกง (CNP, account takeover, friendly fraud)
- **ความถี่:** batch รายวัน + rolling 30/90 วัน
- **เชื่อมโยง:** liability shift จาก 3DS 2.x (เอกสาร 22) ลดผลขาดทุนที่ตกกับ [บริษัท / Company]

### 4.5 KRI สนับสนุน (Supporting KRIs)

| รหัส | ตัวชี้วัด | สูตรย่อ | เหตุผล |
|---|---|---|---|
| KRI-05 | 3DS authentication success rate | `authenticated / 3ds_initiated` | คุณภาพ 3DS, friction, liability |
| KRI-06 | Velocity (per card/BIN/IP/device) | จำนวนธุรกรรมต่อหน่วยเวลา | AML structuring, card testing |
| KRI-07 | System availability | `uptime / total_time` | SLA ≥ 99.95% (ARCHITECTURE.md) |
| KRI-08 | Auth latency p99 | percentile ของ response time | เป้าหมาย < 800 ms |
| KRI-09 | Reconciliation break rate | รายการ mismatch / รายการทั้งหมด | ความถูกต้องของ ledger/settlement |

---

## 5. เกณฑ์ (Thresholds) และสถานะ Traffic-Light

> เกณฑ์เริ่มต้น (initial baseline) — ต้องปรับเทียบ (calibrate) ด้วยข้อมูลจริงในช่วง pilot และอนุมัติโดย Risk Committee

| KRI | หน่วย | 🟢 Green (ปกติ) | 🟡 Amber (เฝ้าระวัง) | 🔴 Red (ยกระดับ) | ระดับ |
|---|---|---|---|---|---|
| KRI-01 Volume deviation | % จาก baseline | ±25% | ±25–50% | > ±50% | portfolio + merchant |
| KRI-02 Failed-auth rate | % | < 8% | 8–15% | > 15% | portfolio + merchant + BIN |
| KRI-02b Card-testing signal (invalid-card spike) | เท่าของ baseline | < 3× | 3–5× | > 5× | BIN/IP |
| KRI-03 Refund ratio | % ของ captured | < 5% | 5–10% | > 10% | merchant |
| KRI-03 Chargeback ratio (count) | % ของ settled | < 0.65% | 0.65–0.9% | > 0.9% | merchant |
| KRI-04 Fraud loss (value) | bps ของยอดขาย | < 10 bps | 10–20 bps | > 20 bps | portfolio + merchant |
| KRI-05 3DS auth success | % | > 90% | 80–90% | < 80% | portfolio |
| KRI-07 Availability | % (rolling 30d) | ≥ 99.95% | 99.9–99.95% | < 99.9% | ระบบ |
| KRI-08 Auth latency p99 | ms | < 800 | 800–1200 | > 1200 | ระบบ |

> **[TODO/ข้อสมมติ]** เกณฑ์ chargeback 0.9% และ fraud bps อ้างอิงแนวเกณฑ์ทั่วไปของ Visa/Mastercard ต้องยืนยันตัวเลขที่บังคับใช้จริงกับ **sponsor bank/scheme** — ตั้ง Amber ต่ำกว่าเกณฑ์แบรนด์เพื่อเตือนล่วงหน้าเสมอ

---

## 6. สัญญาณเตือนและการยกระดับ (Alerting & Escalation)

| สถานะ | ช่องทางแจ้ง | ผู้รับ | SLA ตอบสนอง (triage) | การยกระดับ |
|---|---|---|---|---|
| 🟡 Amber | Dashboard + อีเมล/Slack | Fraud & Risk Ops | ภายใน 4 ชม. (เวลาทำการ) | บันทึก + เฝ้าระวังถี่ขึ้น |
| 🔴 Red | Dashboard + PagerDuty/โทร + Slack | Fraud Ops on-call + Head of Payment Ops | ภายใน 30 นาที (24/7) | เปิด incident, แจ้ง CRO/CISO |
| 🔴 Red (card-testing/BIN attack) | อัตโนมัติ | Fraud Ops + SecOps | ทันที (auto) | เปิด rate-limit/velocity block ตาม runbook |
| 🔴 Red (fraud/AML pattern) | อัตโนมัติ | MLRO/AMLCO | ภายใน 24 ชม. | ประเมิน SAR/STR (เอกสาร 07) |

**การตอบสนองอัตโนมัติ (automated controls):** เมื่อ KRI-02b (card-testing) เข้า Red ระบบสามารถทริกเกอร์ velocity block, step-up 3DS challenge หรือ temporary merchant hold ตาม runbook โดยบันทึกทุกการกระทำลง `audit_log`

**การเชื่อมกับ Incident Response (เอกสาร 16):** ทุก Red ที่เข้าข่ายเหตุการณ์ความมั่นคงปลอดภัยหรือ availability ต้องเปิด incident ticket และเข้าสู่กระบวนการ IR รวมถึงการแจ้ง ธปท./PDPC ตามกำหนดเวลาหากเข้าเงื่อนไข

---

## 7. สถาปัตยกรรมข้อมูลและการนำไปปฏิบัติ (Data & Implementation)

```
payment core (Go/Fiber) ──write──▶ Postgres (payments, ledger_entries, refunds, audit_log)
                                          │  logical replication / CDC
                                          ▼
                              Read Replica / Analytics store
                                          │
                          ┌───────────────┴───────────────┐
                     streaming (นาที)                batch (รายวัน)
                     volume, failed-auth,           chargeback, fraud loss,
                     velocity, latency              reconciliation break
                                          │
                                          ▼
                              KRI compute service ──▶ threshold engine ──▶ alerting
                                          │
                                          ▼
                                   KRI Dashboard (RBAC + MFA)
```

- **การเข้าถึง:** RBAC + MFA (PCI Req 7-8); merchant เห็นเฉพาะข้อมูลตนเอง; ผู้บริหารเห็น portfolio
- **การเก็บรักษา:** ข้อมูล KRI aggregate เก็บ ≥ 5 ปี (สอดคล้องเอกสาร 11 data retention และข้อกำหนด ธปท./AML)
- **ความเป็นส่วนตัว (PDPA):** ข้อมูลบน dashboard เป็น aggregate/pseudonymized; การเข้าถึง raw ต้องมีสิทธิ์และลง log
- **Data quality:** KRI-09 (reconciliation break) และการตรวจ completeness ของ pipeline เป็นตัวควบคุมความถูกต้องของ dashboard เอง

> **[TODO/ข้อสมมติ]** การเลือก stack เฉพาะ (เช่น ClickHouse/Grafana/Metabase/managed BI) ยังไม่ผูกขาด — ต้องยืนยันในสถาปัตยกรรม data platform และให้อยู่ในไทยตามข้อกำหนด data residency ของ ธปท./PDPA

---

## 8. การทบทวน ปรับเทียบ และหลักฐานการตรวจสอบ (Review & Audit)

- **Calibration:** ทบทวน threshold ทุกไตรมาส ด้วย backtest ต่อข้อมูลจริง; บันทึกเหตุผลการเปลี่ยนทุกครั้งลง `audit_log`
- **Backtesting:** ทดสอบว่าเกณฑ์จับเหตุการณ์จริง (true positive) และลด noise (false positive) ได้เพียงพอ
- **หลักฐานสำหรับ ธปท./QSA:** เก็บสแนปช็อต dashboard รายเดือน, log การ acknowledge alert, บันทึก risk acceptance และรายงานต่อคณะกรรมการ
- **การเชื่อมโยงเอกสาร:** เอกสาร 05/07 (AML/SAR), 13 (IT Risk), 16 (Incident Response), 22 (3DS), 24 (Scheme Certification)

---

# Key Risk Indicator dashboard spec: txn volume, failed-auth rate, refund/chargeback ratio, fraud loss (English)

> Document **No. 32** in the Compliance document set for the **Full Acquiring payment-service license application** under the **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT)**, in parallel with **PCI-DSS Level 1**.
>
> **Status:** Submission draft · Version 0.1 · Last updated 2026-07-22
> **Owner:** Chief Risk Officer (CRO), jointly with the CISO and Head of Payment Operations of **[บริษัท / Company]**
> **Audience:** Risk Committee, IT Risk & Security Working Group, Fraud & Risk Operations, MLRO/AMLCO, Internal Audit
>
> This is a system specification and license-application artefact — **not legal advice**. It must be reviewed by legal counsel, the CRO, CISO, MLRO and the QSA before it takes effect, as BOT rules on IT risk, cyber resilience and acquirer supervision may change.

---

## Executive Summary

As a card **acquiring payment gateway**, **[บริษัท / Company]** must monitor operational, fraud, AML and availability risks in near-real-time. This document specifies the requirements for a **Key Risk Indicator (KRI) Dashboard** that consolidates four core indicator families plus supporting indicators, giving management and the Board visibility of risk posture against the Board-approved **Risk Appetite**.

The four core indicators in scope:

1. **Transaction Volume** — count and value (THB) to detect volumetric anomalies and capacity risk.
2. **Failed-Authorization Rate** — share of declines/failures, signalling technical faults, attacks (card testing / enumeration) and routing quality.
3. **Refund / Chargeback Ratio** — refund and chargeback ratios versus sales; an indicator of merchant risk, fraud and exposure to card-scheme monitoring programs.
4. **Fraud Loss** — value and ratio (bps) of confirmed fraud losses.

The dashboard is a **monitoring layer** sourced from architecture tables (`payments`, `ledger_entries`, `refunds`, `webhook_events`, `audit_log`) and wired to the **alerting** and **escalation** mechanisms below, consistent with the Three Lines of Defense and Board risk reporting described in Docs 13 (IT Risk) and 05/07 (AML/SAR).

---

## 1. Legal & Standards Basis

| Law / Standard | Relevance to the KRI Dashboard |
|---|---|
| **Payment Systems Act B.E. 2560** | License conditions require a sound risk-management, internal-control and reporting system, supervised by **BOT** — the KRI Dashboard evidences quantitative risk monitoring. |
| **BOT IT Risk & Cyber Resilience guidelines** | Require monitoring, indicators and Board reporting, including capacity management and anomaly handling. |
| **Anti-Money Laundering Act (AMLO)** | Velocity, structuring and anomalous-transaction KRIs feed the SAR/STR process (see Docs 05, 07). |
| **PCI-DSS v4.0** — Req 10 & 12.10 | Logging, monitoring and anomaly detection; the dashboard must contain no PAN/CHD; indicator-driven events link to incident response. |
| **PDPA B.E. 2562** — **PDPC** | Aggregated dashboard data must apply data minimization; access controlled on least-privilege basis. |
| **EMV 3-D Secure (3DS) 2.x** | Authentication rate, challenge rate and frictionless outcomes feed fraud and failed-auth KRIs. |
| **Visa VAMP / Mastercard ECP/ECM** | Scheme chargeback/fraud thresholds used to set KRI thresholds. |
| **ISO/IEC 27001:2022, NIST CSF 2.0** | Detect function and control-effectiveness measurement. |

> **[TODO/ASSUMPTION]** Confirm with counsel and BOT the exact number/year of the latest applicable IT-risk and reporting notices at filing time and align references.

> **[TODO/ASSUMPTION]** **Visa VAMP** and **Mastercard ECP/ECM** thresholds reference current public scheme documentation — confirm the actually-binding figures with the **sponsor bank / scheme** at contracting (see Doc 24, Scheme Certification).

---

## 2. Scope, Objectives & Design Principles

**Objective:** Provide a single pane of glass over quantitative risk to (1) detect early-warning signals before risks become incidents, (2) benchmark against Risk Appetite, (3) support BOT and Board reporting, and (4) feed evidence into AML/SAR and chargeback management.

**Scope:** The four core indicators (txn volume, failed-auth rate, refund/chargeback ratio, fraud loss) plus supporting indicators (3DS authentication rate, decline-reason breakdown, velocity, system availability/latency), at **portfolio**, **merchant**, and **BIN/country/MCC/channel** levels.

### Design Principles

1. **No CHD on the dashboard** — only `card_brand`, `card_last4`, `bin (first 6-8)`, MCC, country may be shown; never PAN/CVV/PIN (per ARCHITECTURE.md §6).
2. **Single source of truth** — computed primarily from `ledger_entries` (append-only) and `payments` so figures reconcile with settlement.
3. **Money as integer minor units (satang)** — ratios computed with decimal, never float.
4. **Separate read layer (replica)** — analytics reads from a replica isolated from the payment core to avoid impacting authorization latency.
5. **Traffic-light + thresholds** — every KRI carries Green/Amber/Red status against Board-approved thresholds.
6. **Auditability** — every threshold change, alert acknowledgement and risk acceptance is written to `audit_log`.
7. **Near-real-time + batch** — operational KRIs (volume, failed-auth, velocity) refresh via streaming/minute; settlement-dependent KRIs (chargeback, confirmed fraud loss) refresh via daily batch.

---

## 3. Governance & RACI

| Role | Responsibility for the KRI Dashboard |
|---|---|
| **Board** | Approves Risk Appetite and Red-level KRI thresholds; receives quarterly summary. |
| **Risk Committee** | Reviews KRIs quarterly, approves threshold changes, oversees remediation. |
| **CRO (owner)** | **Accountable** for the whole KRI framework; reports to Risk Committee. |
| **CISO** | **Responsible** for availability/latency/security KRIs and log integrity. |
| **Head of Payment Operations** | **Responsible** for volume, failed-auth, refund/chargeback KRIs. |
| **Fraud & Risk Operations (analyst)** | Daily monitoring, alert triage, first-line action, logging. |
| **MLRO / AMLCO** | Receives velocity/structuring signals, feeds the SAR/STR process (Doc 07). |
| **Internal Audit** | Independent assurance over formula correctness, threshold setting and alert response. |
| **Data/Platform Engineering** | **Responsible** for pipeline, data accuracy and dashboard uptime. |

### 3.1 Review Cadence

| Activity | Frequency | Owner |
|---|---|---|
| Real-time monitoring + alert triage | Continuous (24/7 on-call) | Fraud & Risk Ops |
| Daily KRI stand-up | Daily | Payment Ops + Fraud Ops |
| KRI + incident review | Monthly (IT Risk & Security Committee) | CISO/CRO |
| Risk report to Risk Committee | Quarterly | CRO |
| Threshold review + backtest | Quarterly (or event-driven) | CRO + Fraud Ops + Audit |

---

## 4. Core KRI Definitions

Each KRI states definition, formula, data source, dimensions, refresh cadence and Green/Amber/Red thresholds.

### 4.1 KRI-01 — Transaction Volume

- **Definition:** Authorization count and total value (THB) per window, compared against a forecast baseline.
- **Formula:**
  - `txn_count = COUNT(payments WHERE created_at ∈ window)`
  - `txn_value_thb = SUM(amount_minor)/100 WHERE ...`
  - `volume_deviation = (actual − forecast_baseline) / forecast_baseline`
- **Source:** `payments`, `ledger_entries(authorize)`
- **Dimensions:** merchant, MCC, channel (card-present/CNP), BIN country, time (min/hour/day)
- **Cadence:** streaming (refresh ≤ 1 min)
- **Thresholds:** see §5

> Volume spikes/drops may indicate card-testing attacks, anomalous merchants, outages, or capacity risk to SLA (p99 < 800 ms per ARCHITECTURE.md §8).

### 4.2 KRI-02 — Failed-Authorization Rate

- **Definition:** Share of declined/failed authorizations (decline/error/timeout) over total attempts.
- **Formula:** `failed_auth_rate = (declined + errored + timed_out) / total_auth_attempts`
- **By decline reason:** insufficient funds, do-not-honor, invalid card, suspected fraud, 3DS failed, technical/timeout.
- **Source:** `payments(status=failed)`, acquirer/3DS response codes.
- **Dimensions:** merchant, BIN, issuer country, decline reason, channel.
- **Cadence:** streaming (≤ 1 min) + 15-min moving average.
- **Special signal:** abnormally high `invalid card` + volume surge = **card enumeration / BIN attack** signal → escalate immediately.

### 4.3 KRI-03 — Refund / Chargeback Ratio

- **Definition:**
  - `refund_ratio` = refund value over captured value.
  - `chargeback_ratio (count)` = chargeback count over settled count.
  - `chargeback_ratio (value)` = chargeback value over settled value.
- **Formula:**
  - `refund_ratio = SUM(refund_amount)/SUM(captured_amount)`
  - `cb_ratio_count = COUNT(chargebacks)/COUNT(settled_txns)`
  - `cb_ratio_value = SUM(chargeback_amount)/SUM(settled_amount)`
- **Source:** `refunds`, `ledger_entries(refund/capture/settlement)`, chargeback feed from acquirer/scheme.
- **Dimensions:** merchant (most important), MCC, BIN, reason code (fraud/dispute/processing error).
- **Cadence:** daily batch (chargebacks follow settlement) + rolling 30/60/120 days for scheme benchmarking.
- **Scheme linkage:** compared to **Visa VAMP** and **Mastercard ECP/ECM** thresholds to warn before scheme monitoring (fines/penalties) triggers.

### 4.4 KRI-04 — Fraud Loss

- **Definition:** Net loss from confirmed fraud and its ratio to sales (shown in bps = basis points).
- **Formula:**
  - `fraud_loss_thb = SUM(confirmed_fraud_amount − recovered)`
  - `fraud_bps = (fraud_loss_thb / total_sales_thb) × 10,000`
  - `fraud_rate_count = COUNT(confirmed_fraud_txns)/COUNT(total_txns)`
- **Source:** case management (fraud confirmed), chargeback reason=fraud, 3DS liability-shift outcomes.
- **Dimensions:** merchant, MCC, BIN, channel, fraud type (CNP, account takeover, friendly fraud).
- **Cadence:** daily batch + rolling 30/90 days.
- **Linkage:** 3DS 2.x liability shift (Doc 22) reduces loss borne by **[บริษัท / Company]**.

### 4.5 Supporting KRIs

| Code | Indicator | Formula (short) | Rationale |
|---|---|---|---|
| KRI-05 | 3DS authentication success rate | `authenticated / 3ds_initiated` | 3DS quality, friction, liability |
| KRI-06 | Velocity (per card/BIN/IP/device) | txns per unit time | AML structuring, card testing |
| KRI-07 | System availability | `uptime / total_time` | SLA ≥ 99.95% (ARCHITECTURE.md) |
| KRI-08 | Auth latency p99 | percentile of response time | Target < 800 ms |
| KRI-09 | Reconciliation break rate | mismatched / total entries | Ledger/settlement accuracy |

---

## 5. Thresholds & Traffic-Light Status

> Initial baseline — must be calibrated with live pilot data and approved by the Risk Committee.

| KRI | Unit | 🟢 Green | 🟡 Amber | 🔴 Red | Level |
|---|---|---|---|---|---|
| KRI-01 Volume deviation | % from baseline | ±25% | ±25–50% | > ±50% | portfolio + merchant |
| KRI-02 Failed-auth rate | % | < 8% | 8–15% | > 15% | portfolio + merchant + BIN |
| KRI-02b Card-testing signal (invalid-card spike) | × baseline | < 3× | 3–5× | > 5× | BIN/IP |
| KRI-03 Refund ratio | % of captured | < 5% | 5–10% | > 10% | merchant |
| KRI-03 Chargeback ratio (count) | % of settled | < 0.65% | 0.65–0.9% | > 0.9% | merchant |
| KRI-04 Fraud loss (value) | bps of sales | < 10 bps | 10–20 bps | > 20 bps | portfolio + merchant |
| KRI-05 3DS auth success | % | > 90% | 80–90% | < 80% | portfolio |
| KRI-07 Availability | % (rolling 30d) | ≥ 99.95% | 99.9–99.95% | < 99.9% | system |
| KRI-08 Auth latency p99 | ms | < 800 | 800–1200 | > 1200 | system |

> **[TODO/ASSUMPTION]** The 0.9% chargeback and fraud-bps thresholds reflect general Visa/Mastercard guidance; confirm binding figures with the **sponsor bank/scheme**. Always set Amber below the scheme threshold for early warning.

---

## 6. Alerting & Escalation

| Status | Channel | Recipient | Triage SLA | Escalation |
|---|---|---|---|---|
| 🟡 Amber | Dashboard + email/Slack | Fraud & Risk Ops | Within 4h (business hours) | Log + increased monitoring |
| 🔴 Red | Dashboard + PagerDuty/call + Slack | Fraud Ops on-call + Head of Payment Ops | Within 30 min (24/7) | Open incident, notify CRO/CISO |
| 🔴 Red (card-testing/BIN attack) | Automated | Fraud Ops + SecOps | Immediate (auto) | Trigger rate-limit/velocity block per runbook |
| 🔴 Red (fraud/AML pattern) | Automated | MLRO/AMLCO | Within 24h | Assess SAR/STR (Doc 07) |

**Automated controls:** when KRI-02b (card-testing) goes Red, the system can trigger velocity blocks, step-up 3DS challenges or a temporary merchant hold per runbook, logging every action to `audit_log`.

**Incident Response linkage (Doc 16):** every Red qualifying as a security or availability event opens an incident ticket and enters the IR process, including BOT/PDPC notification within required timelines where conditions are met.

---

## 7. Data Architecture & Implementation

```
payment core (Go/Fiber) ──write──▶ Postgres (payments, ledger_entries, refunds, audit_log)
                                          │  logical replication / CDC
                                          ▼
                              Read Replica / Analytics store
                                          │
                          ┌───────────────┴───────────────┐
                     streaming (minute)              batch (daily)
                     volume, failed-auth,           chargeback, fraud loss,
                     velocity, latency              reconciliation break
                                          │
                                          ▼
                              KRI compute service ──▶ threshold engine ──▶ alerting
                                          │
                                          ▼
                                   KRI Dashboard (RBAC + MFA)
```

- **Access:** RBAC + MFA (PCI Req 7-8); merchants see only their own data; executives see portfolio.
- **Retention:** aggregate KRI data retained ≥ 5 years (per Doc 11 retention and BOT/AML requirements).
- **Privacy (PDPA):** dashboard data is aggregate/pseudonymized; raw access is entitled and logged.
- **Data quality:** KRI-09 (reconciliation break) and pipeline completeness checks control the dashboard's own accuracy.

> **[TODO/ASSUMPTION]** The specific stack (e.g. ClickHouse/Grafana/Metabase/managed BI) is not yet fixed — confirm in the data-platform architecture and keep it in Thailand per BOT/PDPA data-residency requirements.

---

## 8. Review, Calibration & Audit Evidence

- **Calibration:** review thresholds quarterly via backtesting against live data; log every change rationale to `audit_log`.
- **Backtesting:** verify thresholds catch real events (true positives) while limiting noise (false positives).
- **Evidence for BOT/QSA:** retain monthly dashboard snapshots, alert-acknowledgement logs, risk-acceptance records and Board reports.
- **Cross-references:** Docs 05/07 (AML/SAR), 13 (IT Risk), 16 (Incident Response), 22 (3DS), 24 (Scheme Certification).
