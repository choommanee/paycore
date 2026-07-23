# แม่แบบรายงานต่อ ธปท. เป็นงวด (ไทย)

> เอกสารเลขที่ **35** ในชุดเอกสาร Compliance ของ **[บริษัท / Company]**
> หัวข้อ: **แม่แบบรายงานต่อธนาคารแห่งประเทศไทย (ธปท.) เป็นงวด** — ปริมาณธุรกรรม, ตัวชี้วัดการทุจริต (fraud), เคส AML, เหตุการณ์ผิดปกติ (incidents) และสถานะ RoC (PCI-DSS)
> บริบท: ผู้ให้บริการ **รับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** — ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท** และมาตรฐาน **PCI-DSS v4.0 Level 1**
>
> **หมายเหตุ:** เอกสารนี้เป็นแม่แบบเชิงปฏิบัติการภายในเพื่อเตรียมความพร้อมยื่นรายงาน มิใช่คำแนะนำทางกฎหมาย รูปแบบ/ความถี่/ช่องทางที่แท้จริงต้องยึดตามประกาศและหนังสือเวียนล่าสุดของ ธปท. และเงื่อนไขแนบท้ายใบอนุญาต

---

## 0. ข้อสมมติและสิ่งที่ยังไม่สรุป (Assumptions / TODO)

> ⚠️ **CALLOUT — รายการที่ต้องยืนยันกับหน่วยงานภายนอกก่อนใช้จริง (ห้ามสมมติเอง)**
>
> 1. **แม่แบบทางการของ ธปท.** — ธปท. อาจกำหนดแบบฟอร์ม/สคีมาไฟล์ (เช่น XML/Excel/DMS) และช่องทางส่งเฉพาะ (เช่น ระบบ DMS — Data Management System / e-Payment reporting portal) — **TODO:** ขอชุดแม่แบบและ data dictionary ทางการจากฝ่ายกำกับระบบการชำระเงิน ธปท. และแทนที่ตารางในเอกสารนี้ให้ตรงฟิลด์
> 2. **Sponsor bank / Acquiring partner** — ยังไม่ลงนามสัญญา (ดู ROADMAP เส้นทาง B) — ช่องรายงาน settlement/chargeback บางส่วนขึ้นกับ interface ของธนาคารผู้สนับสนุน — **TODO:** ยืนยันชื่อธนาคาร, cut-off time, รูปแบบไฟล์ recon
> 3. **QSA vendor (PCI-DSS)** — ยังไม่เลือกผู้ประเมิน (Qualified Security Assessor) — วันครบกำหนด RoC/AoC และรอบ ASV scan จริงขึ้นกับสัญญากับ QSA/ASV — **TODO:** ระบุชื่อ QSA, เลขที่ RoC, วันหมดอายุจริง
> 4. **ทุนจดทะเบียนชำระแล้ว** — ตั้งไว้ที่ **50 ล้านบาท** ตามเกณฑ์ acquiring; ต้องคง **≥ 75% (37.5 ล้านบาท)** ตลอดเวลา — **TODO:** ยืนยันยอดชำระแล้วจริงจากหนังสือรับรอง/งบการเงินล่าสุด
> 5. **ความถี่/กำหนดส่งที่ระบุในเอกสารนี้** (รายเดือน/รายไตรมาส/รายปี/ทันที) เป็นค่าตั้งต้นตามแนวปฏิบัติทั่วไป — **TODO:** ล็อกกับหนังสือเงื่อนไขแนบท้ายใบอนุญาตฉบับจริง
>
> ทุกจุดที่มีเครื่องหมาย `«…»` ในเอกสารคือค่าที่ต้องเติมจากข้อมูลจริง ณ เวลารายงาน

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนด **แม่แบบมาตรฐาน กระบวนการจัดทำ ผู้รับผิดชอบ และกำหนดเวลา** สำหรับรายงานเป็นงวดที่ **[บริษัท / Company]** ต้องนำส่ง ธปท. หลังได้รับใบอนุญาต ครอบคลุม 5 กลุ่มรายงาน:

| # | กลุ่มรายงาน | สาระ | ความถี่ (ตั้งต้น) |
|---|-------------|------|-------------------|
| R1 | ปริมาณธุรกรรม (Transaction Volume) | มูลค่า/จำนวนรายการ authorize, capture, refund, void, settlement | **รายเดือน** |
| R2 | ตัวชี้วัดการทุจริต (Fraud Metrics) | fraud rate (bps), chargeback, 3DS, fraud ที่ป้องกันได้ | **รายเดือน** + สรุป **รายไตรมาส** |
| R3 | เคส AML/CFT | STR/SAR, sanction hit, CDD/EDD, freeze | **รายเดือน** (เคสเร่งด่วนส่ง ปปง. **ทันที**) |
| R4 | เหตุการณ์ผิดปกติ (Incidents) | เหตุ IT/ไซเบอร์/ระบบล่ม/ข้อมูลรั่ว | **ทันที (≤ เวลาที่กำหนด)** + สรุป **รายไตรมาส** |
| R5 | สถานะ RoC / PCI-DSS | RoC, AoC, ASV scan, pentest, ช่องว่าง | **รายไตรมาส** + **รายปี** |

**หลักการควบคุมคุณภาพข้อมูล:** ทุกตัวเลขต้องกระทบยอด (reconcile) กับ **double-entry ledger** (`ledger_entries`) และ **audit_log** ที่เป็น source of truth ตามสถาปัตยกรรมระบบ ก่อนนำส่ง

---

## 2. บทบาทและความรับผิดชอบ (RACI)

| บทบาท | หน้าที่ในการรายงาน |
|-------|---------------------|
| **Compliance Officer (เจ้าหน้าที่กำกับ)** | เจ้าของรายงาน R3/R5, ผู้ลงนามรับรอง, ผู้ประสาน ธปท./ปปง./PDPC |
| **MLRO (Money Laundering Reporting Officer)** | เจ้าของรายงาน R3, ตัดสินใจส่ง STR ต่อ ปปง. |
| **Head of Risk & Fraud** | เจ้าของรายงาน R2, กำหนด threshold, ทบทวน rule |
| **Head of Engineering / SRE** | เจ้าของข้อมูล R1/R4, สรุป incident timeline, RCA |
| **CISO / DPO** | R4 (ไซเบอร์/ข้อมูลรั่ว), ประสาน PDPC, เจ้าของ R5 ฝั่งเทคนิค |
| **Finance / Settlement** | กระทบยอด settlement/chargeback ใน R1/R2 |
| **CEO / กรรมการผู้มีอำนาจ** | ผู้ลงนามอนุมัติขั้นสุดท้ายก่อนนำส่ง ธปท. |

**Maker-Checker:** ทุกรายงานต้องผ่าน 2 ชั้น (ผู้จัดทำ → ผู้ตรวจทาน) และบันทึกใน `audit_log` (actor, timestamp, hash ของไฟล์ที่นำส่ง)

---

## 3. ปฏิทินการนำส่ง (Reporting Calendar)

| รายงาน | รอบข้อมูล | กำหนดส่ง (ตั้งต้น) | ช่องทาง (TODO ยืนยัน) |
|--------|-----------|---------------------|------------------------|
| R1 Transaction Volume | เดือนปฏิทิน | ภายในวันทำการที่ **15** ของเดือนถัดไป | DMS / e-Payment portal ธปท. |
| R2 Fraud (รายเดือน) | เดือนปฏิทิน | ภายในวันที่ **15** ของเดือนถัดไป | DMS ธปท. |
| R2 Fraud (รายไตรมาส) | ไตรมาส | ภายใน **30 วัน** หลังสิ้นไตรมาส | DMS ธปท. |
| R3 AML (รายเดือน) | เดือนปฏิทิน | ภายในวันที่ **15** ของเดือนถัดไป | ธปท. |
| R3 STR/เร่งด่วน | ต่อเคส | **ทันทีที่พบ** ตามเกณฑ์ ปปง. | ระบบ ปปง. (AMLO) |
| R4 Incident (ทันที) | ต่อเหตุการณ์ | **แจ้งเบื้องต้น ≤ 24 ชม.** / รายงานฉบับเต็ม ≤ **72 ชม.** (ดูข้อ 7) | ช่องทางแจ้งเหตุ ธปท./ThaiCERT |
| R4 Incident (สรุป) | ไตรมาส | ภายใน **30 วัน** หลังสิ้นไตรมาส | DMS ธปท. |
| R5 RoC/PCI | ไตรมาส/ปี | ไตรมาส: ภายใน **30 วัน**; ปี: ภายใน **60 วัน** หลังออก RoC | ธปท. |

> ปิดงวดข้อมูลใช้เวลาไทย (Asia/Bangkok, UTC+7); ตัวเลขการเงินเป็นสกุล THB และหน่วย **สตางค์ (minor units, integer)** ภายในระบบ แสดงผลรายงานเป็นบาททศนิยม 2 ตำแหน่ง

---

## 4. R1 — แม่แบบรายงานปริมาณธุรกรรม (Transaction Volume)

**แหล่งข้อมูล:** `payments`, `ledger_entries`, `refunds` — กระทบยอดกับไฟล์ settlement จาก acquirer/sponsor bank

### 4.1 สรุประดับพอร์ต (รายเดือน)

| ฟิลด์ | คำอธิบาย | ตัวอย่าง |
|-------|----------|----------|
| `report_period` | งวด (YYYY-MM) | 2026-06 |
| `txn_count_authorized` | จำนวนรายการอนุมัติสำเร็จ | «12,345» |
| `txn_count_captured` | จำนวนรายการ capture | «11,980» |
| `gross_amount_captured_thb` | มูลค่ารวม capture (บาท) | «45,200,150.00» |
| `refund_count` / `refund_amount_thb` | จำนวน/มูลค่า refund | «210 / 1,050,000.00» |
| `void_count` | จำนวน void (ก่อน capture) | «95» |
| `net_settled_amount_thb` | ยอด settle สุทธิ (capture − refund − fee) | «43,900,120.00» |
| `avg_ticket_thb` | มูลค่าเฉลี่ยต่อรายการ | «3,772.00» |
| `decline_rate_pct` | อัตราปฏิเสธ (declined / attempted) | «4.8%» |
| `merchant_count_active` | จำนวนร้านค้าที่มีรายการ | «318» |

### 4.2 แยกตามมิติ (แนบเป็นตารางย่อย)

- **ตามช่องทาง:** e-commerce (CNP) / POS (card-present) / QR / recurring
- **ตามแบรนด์บัตร:** Visa / Mastercard / JCB / UnionPay / local scheme
- **ตามสกุลเงิน:** THB / อื่น ๆ (ระบุ FX)
- **ตาม MCC / กลุ่มธุรกิจร้านค้า** (ระบุ high-risk MCC แยก)
- **Top 10 merchant** ตามมูลค่า (เพื่อดู concentration risk)

### 4.3 การกระทบยอด

ต้องแนบ **reconciliation statement**: `ledger net` เทียบ `acquirer settlement file` — ผลต่างต้อง **= 0** หรืออธิบายทุกบรรทัดที่ต่าง (unmatched > «0.5%» ต้องมี remediation note)

---

## 5. R2 — แม่แบบตัวชี้วัดการทุจริต (Fraud Metrics)

**แหล่งข้อมูล:** Risk/Fraud Engine, `payments`, ข้อมูล chargeback จาก acquirer, ผล 3DS (EMV 3DS / 3-D Secure 2.x)

### 5.1 ตัวชี้วัดหลัก (รายเดือน)

| ฟิลด์ | คำอธิบาย | เกณฑ์เฝ้าระวัง (ตั้งต้น) |
|-------|----------|--------------------------|
| `fraud_txn_count` / `fraud_amount_thb` | จำนวน/มูลค่ารายการทุจริตที่ยืนยัน | — |
| `fraud_rate_bps` | มูลค่า fraud ÷ มูลค่า capture (basis points) | **แจ้งเตือนภายในเมื่อ > 15 bps; วิกฤตเมื่อ > 40 bps** (สอดคล้องเกณฑ์ scheme fraud monitoring) |
| `chargeback_count` / `chargeback_amount_thb` | ปริมาณ chargeback | — |
| `chargeback_ratio_pct` | chargeback ÷ รายการ capture | **เฝ้าระวัง > 0.65%; วิกฤต > 0.9%** (แนว Visa VDMP/Mastercard ECM) |
| `3ds_attempt_rate_pct` | สัดส่วนรายการที่ผ่าน 3DS | «96.5%» |
| `3ds_challenge_rate_pct` | สัดส่วนที่ถูก challenge | «12.0%» |
| `fraud_prevented_count` | รายการที่ถูก rule/model บล็อก | — |
| `false_positive_rate_pct` | อัตราบล็อกผิดพลาด | «< 2%» |
| `velocity_block_count` | บล็อกจาก velocity rule | — |
| `blacklist_hit_count` | ตรงบัญชีดำ (BIN/card/device) | — |

### 5.2 การจำแนกประเภททุจริต

CNP fraud / lost-stolen / counterfeit / account takeover / friendly fraud / merchant collusion — พร้อม **แนวโน้ม 3 งวดย้อนหลัง** และ **top merchant ที่มี fraud สูง**

### 5.3 มาตรการแก้ไข

ทุกไตรมาสต้องสรุป: rule ที่เพิ่ม/ปรับ, การขึ้น 3DS challenge, การพักร้านค้าเสี่ยง, แผนลด chargeback ratio ให้ต่ำกว่าเกณฑ์ scheme

---

## 6. R3 — แม่แบบเคส AML/CFT

**กรอบกฎหมาย:** พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน + พ.ร.บ. ป้องกันและปราบปรามการสนับสนุนทางการเงินแก่การก่อการร้ายฯ กำกับโดย **สำนักงาน ปปง. (AMLO)** — รายงานเชิงกำกับด้านความเสี่ยง AML สรุปให้ ธปท. ส่วน STR/CTR ที่เป็นเคสส่งตรง **ปปง.**

### 6.1 สรุปเชิงตัวเลข (รายเดือน — ส่ง ธปท.)

| ฟิลด์ | คำอธิบาย |
|-------|----------|
| `str_filed_count` | จำนวนรายงานธุรกรรมน่าสงสัย (STR) ที่ส่ง ปปง. ในงวด |
| `ctr_filed_count` | รายงานธุรกรรมเงินสด/ตามเกณฑ์มูลค่า (ถ้ามี) |
| `sanction_screening_hits` | จำนวน hit จากการคัดกรอง (UN, OFAC, ปปง. designated list) |
| `true_positive_hits` | hit ที่ยืนยันเป็นบุคคล/นิติบุคคลต้องห้าม |
| `accounts_frozen_count` | จำนวนบัญชี/ร้านค้าที่ถูกอายัด/ระงับ |
| `edd_cases_count` | เคสที่เข้าสู่ Enhanced Due Diligence |
| `pep_relationships_count` | ความสัมพันธ์กับบุคคลที่มีสถานภาพทางการเมือง (PEP) |
| `cdd_overdue_count` | ร้านค้าที่ทบทวน CDD เกินกำหนด |

### 6.2 ทะเบียนเคส (แนบ, ปกปิดข้อมูลตาม PDPA)

| ฟิลด์ | หมายเหตุ |
|-------|----------|
| `case_id` | รหัสภายใน (ไม่เปิดเผยตัวตนในเวอร์ชันสรุป) |
| `case_type` | STR / sanction / EDD / PEP / structuring |
| `opened_at` / `status` | เปิด / อยู่ระหว่างสอบ / ส่ง ปปง. / ปิด |
| `amlo_ref_no` | เลขอ้างอิงที่ ปปง. ตอบรับ (ถ้ามี) |
| `disposition` | ผลการพิจารณา + มาตรการ |

> ⚠️ **CALLOUT — ความเร่งด่วน:** STR ต้องส่ง ปปง. **ทันทีตามเกณฑ์ของ ปปง.** ห้ามรอรอบรายเดือน; รายงานรายเดือนต่อ ธปท. เป็นเพียงสรุปภาพรวมความเสี่ยง ไม่ใช่ช่องทางส่ง STR การส่งข้อมูลลูกค้าต้องเป็นไปตาม PDPA และฐานทางกฎหมายที่เกี่ยวข้อง

---

## 7. R4 — แม่แบบรายงานเหตุการณ์ผิดปกติ (Incidents)

**ครอบคลุม:** ระบบล่ม (availability < SLA 99.95%), การละเมิดความปลอดภัยไซเบอร์, ข้อมูลรั่วไหล (รวมข้อมูลส่วนบุคคล/ข้อมูลบัตร), การทุจริตเชิงระบบ, ความล้มเหลวของ acquirer/settlement

### 7.1 ระดับความรุนแรง (Severity)

| ระดับ | นิยาม | ตัวอย่าง |
|-------|-------|----------|
| **SEV-1 (วิกฤต)** | กระทบวงกว้าง/เงิน/ข้อมูลบัตรรั่ว/ระบบชำระเงินหยุด | payment core ล่ม, ต้องสงสัยข้อมูล PAN รั่ว |
| **SEV-2 (สูง)** | บริการบางส่วนเสีย/degraded | 3DS ล่มบางส่วน, latency เกิน SLA ต่อเนื่อง |
| **SEV-3 (กลาง)** | กระทบจำกัด มี workaround | webhook delay, node เดียวล่ม |
| **SEV-4 (ต่ำ)** | ไม่กระทบลูกค้า | alert เชิงป้องกัน |

### 7.2 ไทม์ไลน์การแจ้ง (ตั้งต้น — TODO ยืนยันกับเงื่อนไขใบอนุญาต)

| ขั้น | เวลา | ผู้รับ |
|-----|------|--------|
| **แจ้งเบื้องต้น (initial notification)** | ≤ **24 ชั่วโมง** หลังตรวจพบ (SEV-1/2) | ธปท. (ช่องทางแจ้งเหตุ), ThaiCERT ถ้าเป็นไซเบอร์ |
| **รายงานฉบับเต็ม** | ≤ **72 ชั่วโมง** | ธปท. |
| **แจ้ง PDPC (ข้อมูลส่วนบุคคลรั่ว)** | ≤ **72 ชั่วโมง** ตาม PDPA | PDPC + แจ้งเจ้าของข้อมูลถ้าเสี่ยงสูง |
| **RCA / Post-mortem** | ≤ **14 วัน** | ธปท. + ภายใน |

### 7.3 ฟิลด์รายงานเหตุการณ์

| ฟิลด์ | คำอธิบาย |
|-------|----------|
| `incident_id` / `severity` | รหัส + ระดับ |
| `detected_at` / `contained_at` / `resolved_at` | ไทม์ไลน์ |
| `category` | availability / security / data-breach / fraud / third-party |
| `impact` | จำนวน merchant/รายการ/มูลค่าที่กระทบ, ข้อมูลบัตร/บุคคลเกี่ยวข้องหรือไม่ |
| `root_cause` | สาเหตุราก (จาก RCA) |
| `remediation` / `preventive_actions` | มาตรการแก้ + ป้องกันซ้ำ |
| `pci_impact` | กระทบ CDE / PCI scope หรือไม่ |
| `regulator_notified_at` | เวลาที่แจ้ง ธปท./ปปง./PDPC |

> **PCI-DSS v4.0 Req 12.10** กำหนดให้มี incident response plan และการทดสอบประจำปี — รายงานสรุปรายไตรมาสต้องอ้างอิงผลการซ้อม (tabletop/DR drill) ด้วย

---

## 8. R5 — แม่แบบสถานะ RoC / PCI-DSS

**มาตรฐาน:** PCI-DSS **v4.0**, ระดับ **Level 1** (ประมวลผล > 6 ล้านรายการ/ปี หรือกำหนดโดย scheme) — ต้องมี **QSA** ทำ assessment ออก **RoC (Report on Compliance)** + **AoC (Attestation of Compliance)** ประจำปี, **ASV scan รายไตรมาส**, และ **penetration test รายปี**

### 8.1 แดชบอร์ดสถานะ (รายไตรมาส)

| ฟิลด์ | คำอธิบาย | สถานะตัวอย่าง |
|-------|----------|----------------|
| `pci_version` | เวอร์ชันที่ประเมิน | v4.0 |
| `assessment_level` | ระดับ | Level 1 |
| `qsa_vendor` | ชื่อ QSA | «TODO — ยังไม่เลือก» |
| `roc_status` | สถานะ RoC | in-progress / compliant / remediation |
| `roc_issue_date` / `roc_expiry_date` | วันออก/หมดอายุ | «TBD» |
| `aoc_on_file` | มี AoC ล่าสุดหรือไม่ | yes/no |
| `asv_scan_last` / `asv_result` | ASV scan ล่าสุด | «2026-Qx / pass» |
| `pentest_last` / `pentest_findings_open` | pentest ล่าสุด + findings ค้าง | «วันที่ / จำนวน critical=0» |
| `open_findings_by_severity` | ช่องว่างค้าง แยกระดับ | critical/high/medium/low |
| `saq_or_roc_scope_changes` | การเปลี่ยน scope/segmentation | — |

### 8.2 การติดตามช่องว่าง (Gap Tracker)

| ฟิลด์ | หมายเหตุ |
|-------|----------|
| `finding_id` / `pci_requirement` | อ้างข้อกำหนด (เช่น Req 3, 8, 10, 11) |
| `severity` / `due_date` / `owner` | ระดับ / กำหนดปิด / ผู้รับผิดชอบ |
| `remediation_status` | open / in-progress / closed |
| `compensating_control` | มาตรการชดเชย (ถ้ามี) |

> ⚠️ **CALLOUT (TODO):** วันออก/หมดอายุ RoC จริงและรอบ ASV จริงขึ้นกับสัญญากับ QSA/ASV ที่ยังไม่ลงนาม — ห้ามกรอกวันสมมติในรายงานจริง

---

## 9. การควบคุมคุณภาพ ความปลอดภัย และการเก็บรักษา

1. **แหล่งความจริงเดียว:** ทุกตัวเลขดึงจาก ledger/audit_log/fraud engine ผ่าน query ที่ version-controlled; ห้ามแก้มือใน Excel โดยไม่มี trail
2. **ไม่มีข้อมูลบัตรในรายงาน:** ห้ามใส่ full PAN/CVV/track — ใช้ได้แค่ `card_brand`+`card_last4`; รายงานลูกค้าต้อง minimize ตาม PDPA
3. **การเก็บรักษา:** เก็บสำเนารายงานและหลักฐานการนำส่ง **≥ 5–10 ปี** (ตามเกณฑ์ AML/ธปท. — TODO ยืนยันปีที่แน่นอน)
4. **ความลับและการเข้าถึง:** ไฟล์รายงานเก็บใน storage ที่เข้ารหัส, access แบบ least-privilege + MFA (PCI Req 7-8), ลง `audit_log` ทุกครั้งที่เข้าถึง/นำส่ง
5. **การลงนามและ integrity:** ไฟล์ที่นำส่งบันทึก hash (SHA-256) ใน audit trail เพื่อพิสูจน์ความครบถ้วน

---

# Periodic BOT reporting templates: transaction volume, fraud metrics, AML cases, incidents, RoC status (English)

> Document **#35** in the **[บริษัท / Company]** compliance set.
> Subject: **Periodic Bank of Thailand (BOT) reporting templates** — transaction volume, fraud metrics, AML cases, incidents, and RoC (PCI-DSS) status.
> Context: **Full Acquiring** payment service provider under the **Payment Systems Act B.E. 2560 (2017)** — paid-up registered capital **THB 50M** and **PCI-DSS v4.0 Level 1**.
>
> **Note:** This is an internal operational template to prepare regulatory submissions; it is not legal advice. Actual format, frequency, and channel must follow the latest BOT notifications/circulars and the conditions attached to the license.

---

## 0. Assumptions / TODO

> ⚠️ **CALLOUT — items to confirm with external parties before production use (do not invent facts)**
>
> 1. **Official BOT templates** — BOT may prescribe specific forms/file schemas (XML/Excel/DMS) and submission channels (e.g., the DMS — Data Management System / e-Payment reporting portal). **TODO:** obtain the official template set and data dictionary from BOT's Payment Systems Policy Department and align the field tables herein.
> 2. **Sponsor bank / acquiring partner** — not yet contracted (see ROADMAP Track B). Some settlement/chargeback fields depend on the sponsor bank interface. **TODO:** confirm bank name, cut-off times, recon file format.
> 3. **QSA vendor (PCI-DSS)** — not yet selected. Actual RoC/AoC due dates and ASV scan cadence depend on the QSA/ASV contract. **TODO:** record QSA name, RoC number, real expiry date.
> 4. **Paid-up registered capital** — set at **THB 50M** per the acquiring threshold; must remain **≥ 75% (THB 37.5M)** at all times. **TODO:** confirm actual paid-up amount from the latest company affidavit/financials.
> 5. **Frequencies/deadlines stated here** (monthly/quarterly/annual/immediate) are defaults per common practice. **TODO:** lock against the actual license condition letter.
>
> Every `«…»` placeholder must be filled with real data at reporting time.

---

## 1. Purpose and Scope

This document defines **standard templates, preparation process, owners, and deadlines** for the periodic reports **[บริษัท / Company]** must file with BOT after licensing, covering 5 report families:

| # | Report family | Substance | Frequency (default) |
|---|---------------|-----------|----------------------|
| R1 | Transaction Volume | value/count of authorize, capture, refund, void, settlement | **Monthly** |
| R2 | Fraud Metrics | fraud rate (bps), chargebacks, 3DS, fraud prevented | **Monthly** + **quarterly** summary |
| R3 | AML/CFT cases | STR/SAR, sanction hits, CDD/EDD, freezes | **Monthly** (urgent cases to AMLO **immediately**) |
| R4 | Incidents | IT/cyber/outage/data-breach events | **Immediate (≤ deadline)** + **quarterly** summary |
| R5 | RoC / PCI-DSS status | RoC, AoC, ASV scan, pentest, gaps | **Quarterly** + **annual** |

**Data quality principle:** every figure must reconcile to the **double-entry ledger** (`ledger_entries`) and the `audit_log` — the system's source of truth per the architecture — before submission.

---

## 2. Roles and Responsibilities (RACI)

| Role | Reporting responsibility |
|------|--------------------------|
| **Compliance Officer** | Owner of R3/R5, certifying signatory, liaison to BOT/AMLO/PDPC |
| **MLRO** | Owner of R3, decides STR filing to AMLO |
| **Head of Risk & Fraud** | Owner of R2, sets thresholds, rule reviews |
| **Head of Engineering / SRE** | Data owner for R1/R4, incident timelines, RCA |
| **CISO / DPO** | R4 (cyber/data breach), PDPC liaison, technical owner of R5 |
| **Finance / Settlement** | Reconciles settlement/chargeback in R1/R2 |
| **CEO / authorized director** | Final approving signatory before BOT submission |

**Maker-checker:** every report passes 2 layers (preparer → reviewer) and is logged in `audit_log` (actor, timestamp, hash of submitted file).

---

## 3. Reporting Calendar

| Report | Data period | Deadline (default) | Channel (TODO confirm) |
|--------|-------------|---------------------|-------------------------|
| R1 Transaction Volume | calendar month | by business day **15** of next month | BOT DMS / e-Payment portal |
| R2 Fraud (monthly) | calendar month | by day **15** of next month | BOT DMS |
| R2 Fraud (quarterly) | quarter | within **30 days** of quarter-end | BOT DMS |
| R3 AML (monthly) | calendar month | by day **15** of next month | BOT |
| R3 STR/urgent | per case | **immediately** per AMLO rules | AMLO system |
| R4 Incident (immediate) | per event | **initial ≤ 24h** / full report ≤ **72h** (see §7) | BOT incident channel / ThaiCERT |
| R4 Incident (summary) | quarter | within **30 days** of quarter-end | BOT DMS |
| R5 RoC/PCI | quarter/year | quarter: within **30 days**; annual: within **60 days** of RoC issuance | BOT |

> Data cut-offs use Thailand time (Asia/Bangkok, UTC+7); financial figures are THB, stored internally as **satang (integer minor units)** and presented as baht with 2 decimals.

---

## 4. R1 — Transaction Volume Template

**Sources:** `payments`, `ledger_entries`, `refunds` — reconciled against the acquirer/sponsor-bank settlement file.

### 4.1 Portfolio summary (monthly)

| Field | Description | Example |
|-------|-------------|---------|
| `report_period` | period (YYYY-MM) | 2026-06 |
| `txn_count_authorized` | successful authorizations | «12,345» |
| `txn_count_captured` | captures | «11,980» |
| `gross_amount_captured_thb` | total captured (THB) | «45,200,150.00» |
| `refund_count` / `refund_amount_thb` | refund count/value | «210 / 1,050,000.00» |
| `void_count` | voids (pre-capture) | «95» |
| `net_settled_amount_thb` | net settled (capture − refund − fee) | «43,900,120.00» |
| `avg_ticket_thb` | average ticket value | «3,772.00» |
| `decline_rate_pct` | declined / attempted | «4.8%» |
| `merchant_count_active` | merchants with activity | «318» |

### 4.2 Breakdowns (attached sub-tables)

- **By channel:** e-commerce (CNP) / POS (card-present) / QR / recurring
- **By card brand:** Visa / Mastercard / JCB / UnionPay / local scheme
- **By currency:** THB / others (state FX)
- **By MCC / merchant category** (list high-risk MCCs separately)
- **Top 10 merchants** by value (concentration risk)

### 4.3 Reconciliation

Attach a **reconciliation statement**: `ledger net` vs `acquirer settlement file` — the difference must be **0** or every variance line explained (unmatched > «0.5%» requires a remediation note).

---

## 5. R2 — Fraud Metrics Template

**Sources:** Risk/Fraud Engine, `payments`, acquirer chargeback data, 3DS results (EMV 3DS / 3-D Secure 2.x).

### 5.1 Core metrics (monthly)

| Field | Description | Watch threshold (default) |
|-------|-------------|----------------------------|
| `fraud_txn_count` / `fraud_amount_thb` | confirmed fraud count/value | — |
| `fraud_rate_bps` | fraud value ÷ captured value (basis points) | **internal alert > 15 bps; critical > 40 bps** (aligned to scheme fraud monitoring) |
| `chargeback_count` / `chargeback_amount_thb` | chargeback volume | — |
| `chargeback_ratio_pct` | chargebacks ÷ captures | **watch > 0.65%; critical > 0.9%** (Visa VDMP / Mastercard ECM guidance) |
| `3ds_attempt_rate_pct` | share going through 3DS | «96.5%» |
| `3ds_challenge_rate_pct` | share challenged | «12.0%» |
| `fraud_prevented_count` | txns blocked by rule/model | — |
| `false_positive_rate_pct` | wrongful block rate | «< 2%» |
| `velocity_block_count` | velocity-rule blocks | — |
| `blacklist_hit_count` | BIN/card/device blacklist hits | — |

### 5.2 Fraud typology

CNP fraud / lost-stolen / counterfeit / account takeover / friendly fraud / merchant collusion — with a **3-period trend** and **top merchants by fraud**.

### 5.3 Remediation

Each quarter summarize: rules added/tuned, 3DS challenge escalation, high-risk merchant suspensions, and the plan to keep chargeback ratio below scheme thresholds.

---

## 6. R3 — AML/CFT Cases Template

**Legal frame:** the Anti-Money Laundering Act and the Counter-Terrorism & Proliferation Financing Act, supervised by the **Anti-Money Laundering Office (AMLO)**. AML risk oversight is summarized to BOT; **STRs/CTRs are filed directly to AMLO**.

### 6.1 Numeric summary (monthly — to BOT)

| Field | Description |
|-------|-------------|
| `str_filed_count` | STRs filed to AMLO in the period |
| `ctr_filed_count` | cash/threshold-value reports (if any) |
| `sanction_screening_hits` | screening hits (UN, OFAC, AMLO designated list) |
| `true_positive_hits` | confirmed prohibited persons/entities |
| `accounts_frozen_count` | accounts/merchants frozen or suspended |
| `edd_cases_count` | cases escalated to Enhanced Due Diligence |
| `pep_relationships_count` | Politically Exposed Person relationships |
| `cdd_overdue_count` | merchants with overdue CDD review |

### 6.2 Case register (attached, PDPA-redacted)

| Field | Note |
|-------|------|
| `case_id` | internal ID (de-identified in summary version) |
| `case_type` | STR / sanction / EDD / PEP / structuring |
| `opened_at` / `status` | open / investigating / filed to AMLO / closed |
| `amlo_ref_no` | AMLO acknowledgement reference (if any) |
| `disposition` | outcome + action taken |

> ⚠️ **CALLOUT — urgency:** STRs must be filed to AMLO **immediately per AMLO rules** — never held for the monthly cycle. The monthly BOT report is a risk overview only, not an STR channel. Any transfer of customer data must comply with PDPA and the applicable lawful basis.

---

## 7. R4 — Incident Report Template

**Covers:** outages (availability < 99.95% SLA), cybersecurity breaches, data leakage (including personal/cardholder data), systemic fraud, and acquirer/settlement failures.

### 7.1 Severity levels

| Level | Definition | Example |
|-------|-----------|---------|
| **SEV-1 (critical)** | broad impact / funds / cardholder-data leak / payments halted | payment core down, suspected PAN exposure |
| **SEV-2 (high)** | partial/degraded service | partial 3DS outage, sustained latency breach |
| **SEV-3 (medium)** | limited impact, workaround exists | webhook delay, single-node failure |
| **SEV-4 (low)** | no customer impact | preventive alert |

### 7.2 Notification timeline (default — TODO confirm against license conditions)

| Step | Time | Recipient |
|------|------|-----------|
| **Initial notification** | ≤ **24 hours** after detection (SEV-1/2) | BOT (incident channel), ThaiCERT if cyber |
| **Full report** | ≤ **72 hours** | BOT |
| **PDPC notification (personal-data breach)** | ≤ **72 hours** per PDPA | PDPC + affected data subjects if high risk |
| **RCA / Post-mortem** | ≤ **14 days** | BOT + internal |

### 7.3 Incident fields

| Field | Description |
|-------|-------------|
| `incident_id` / `severity` | ID + level |
| `detected_at` / `contained_at` / `resolved_at` | timeline |
| `category` | availability / security / data-breach / fraud / third-party |
| `impact` | merchants/txns/value affected, whether cardholder/personal data involved |
| `root_cause` | root cause (from RCA) |
| `remediation` / `preventive_actions` | fix + prevent-recurrence measures |
| `pci_impact` | whether CDE / PCI scope affected |
| `regulator_notified_at` | when BOT/AMLO/PDPC were notified |

> **PCI-DSS v4.0 Req 12.10** requires an incident response plan and annual testing — the quarterly summary must reference drill results (tabletop / DR drill).

---

## 8. R5 — RoC / PCI-DSS Status Template

**Standard:** PCI-DSS **v4.0**, **Level 1** (>6M txns/year or scheme-designated) — requires a **QSA** assessment producing a **RoC (Report on Compliance)** + **AoC (Attestation of Compliance)** annually, **quarterly ASV scans**, and an **annual penetration test**.

### 8.1 Status dashboard (quarterly)

| Field | Description | Example status |
|-------|-------------|----------------|
| `pci_version` | assessed version | v4.0 |
| `assessment_level` | level | Level 1 |
| `qsa_vendor` | QSA name | «TODO — not selected» |
| `roc_status` | RoC status | in-progress / compliant / remediation |
| `roc_issue_date` / `roc_expiry_date` | issue/expiry | «TBD» |
| `aoc_on_file` | latest AoC on file? | yes/no |
| `asv_scan_last` / `asv_result` | last ASV scan | «2026-Qx / pass» |
| `pentest_last` / `pentest_findings_open` | last pentest + open findings | «date / critical=0» |
| `open_findings_by_severity` | open gaps by severity | critical/high/medium/low |
| `saq_or_roc_scope_changes` | scope/segmentation changes | — |

### 8.2 Gap tracker

| Field | Note |
|-------|------|
| `finding_id` / `pci_requirement` | referenced requirement (e.g., Req 3, 8, 10, 11) |
| `severity` / `due_date` / `owner` | severity / close date / owner |
| `remediation_status` | open / in-progress / closed |
| `compensating_control` | compensating control (if any) |

> ⚠️ **CALLOUT (TODO):** Real RoC issue/expiry dates and ASV cadence depend on the not-yet-signed QSA/ASV contract — do not enter assumed dates in the actual report.

---

## 9. Quality, Security, and Retention Controls

1. **Single source of truth:** every figure is pulled from the ledger/audit_log/fraud engine via version-controlled queries; no manual Excel edits without a trail.
2. **No card data in reports:** never include full PAN/CVV/track — only `card_brand`+`card_last4`; customer data minimized per PDPA.
3. **Retention:** keep report copies and submission evidence **≥ 5–10 years** (per AML/BOT rules — TODO confirm exact years).
4. **Confidentiality and access:** reports stored in encrypted storage, least-privilege access + MFA (PCI Req 7-8), logged in `audit_log` on every access/submission.
5. **Signing and integrity:** submitted files record a SHA-256 hash in the audit trail to prove completeness.

---

## 10. Related documents

- `COMPLIANCE-TH.md` — Thai law, license categories, application steps
- `ARCHITECTURE.md` — system architecture, data model, PCI controls
- `ROADMAP.md` — phases, timeline, cost estimates
- Other files in `docs/compliance/` — full compliance document set
