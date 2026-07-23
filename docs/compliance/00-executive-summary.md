# บทสรุปผู้บริหาร (ยื่นขอใบอนุญาต Full Acquiring) (ไทย)

> เอกสารเลขที่ `docs/compliance/00-executive-summary.md` · เวอร์ชัน 0.1 · วันที่จัดทำ 22 ก.ค. 2569
> ประเภทคำขอ: **ใบอนุญาตให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service) แบบเต็มรูปแบบ**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)**
> ผู้ขอ: **[บริษัท / Company]** · ทุนจดทะเบียนชำระแล้ว: **50,000,000 บาท**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อการภายใน (pre-submission draft) — ต้องผ่านการทานโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ก่อนยื่นจริง เอกสารนี้เป็นบทสรุปเชิงบริหาร มิใช่คำแนะนำทางกฎหมาย

---

## 1. วัตถุประสงค์ของคำขอ

[บริษัท / Company] ประสงค์ยื่นขอ **ใบอนุญาตให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)** ในกลุ่ม **บริการการชำระเงินภายใต้การกำกับ (Designated Payment Services)** ตาม พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 โดยใบอนุญาตออกโดย **รัฐมนตรีว่าการกระทรวงการคลัง ตามคำแนะนำของ ธปท.**

บริษัทเลือก **เส้นทาง B — Full Acquiring เต็มรูปแบบ** (ตาม `ROADMAP.md`) กล่าวคือ รับ-ประมวลผล authorization/clearing/settlement ของธุรกรรมบัตรโดยตรงผ่าน sponsor bank และ card scheme มิใช่การต่อพ่วง acquirer ที่มีอยู่ (ซึ่งจะเป็นเส้นทาง A — Payment Facilitating ทุน 10 ล้านบาท)

---

## 2. รูปแบบธุรกิจ (Business Model)

| หัวข้อ | สาระสำคัญ |
|--------|-----------|
| บริการหลัก | รับชำระเงินด้วยบัตร (card acquiring) สำหรับร้านค้า e-commerce, in-app และ POS |
| วิธีชำระเงินที่รองรับ | บัตรเครดิต/เดบิต Visa, Mastercard (+ local schemes ในเฟสถัดไป) พร้อม EMV 3DS 2.x |
| แหล่งรายได้ | Merchant Discount Rate (MDR), ค่าธรรมเนียมต่อรายการ, ค่าธรรมเนียม settlement/payout, ค่าบริการเสริม (tokenization, รายงาน) |
| การรับเงินและจ่ายเงิน | เงินคำนวณเป็นจำนวนเต็มหน่วยย่อย (สตางค์) บันทึกใน **ledger แบบ double-entry append-only** เป็น source of truth; จ่ายร้านค้าตามรอบ settlement |
| กลุ่มลูกค้าเป้าหมาย | ผู้ประกอบการไทยที่มีปริมาณธุรกรรมสูง ต้องการคุมต้นทุนต่อรายการ |
| การไม่เก็บข้อมูลบัตร | ระบบหลัก **ไม่เก็บ full PAN/CVV/PIN/track** — เก็บได้เฉพาะ `card_brand` + `card_last4`; ใช้ tokenization vault แยก network segment เพื่อลด PCI scope |

**หลักการเชิงบริหารความเสี่ยงในการออกแบบ** (จาก `ARCHITECTURE.md`): idempotency ทุก endpoint ที่ขยับเงิน, fail-closed, auditability ครบทุก state change ลง `audit_log`

---

## 3. ขนาดและประมาณการ (Scale)

> **[สมมติฐาน / TODO]** ตัวเลขปริมาณธุรกรรมและ GMV ด้านล่างเป็นประมาณการเพื่อวางแผน ต้องแทนที่ด้วยตัวเลขจาก business plan/financial model ที่ผ่านการรับรองก่อนยื่น ธปท.

| ตัวชี้วัด | ปีที่ 1 | ปีที่ 2 | ปีที่ 3 |
|----------|--------|--------|--------|
| จำนวนร้านค้า (active) | 300 | 1,200 | 3,500 |
| ปริมาณธุรกรรม (รายการ/ปี) | > 6 ล้าน | ~25 ล้าน | ~60 ล้าน |
| GMV (ประมาณ) | *[TODO]* | *[TODO]* | *[TODO]* |
| ระดับ PCI-DSS | **Level 1** | Level 1 | Level 1 |

เนื่องจากประมาณการปริมาณ **เกิน 6 ล้านรายการ/ปี** ตั้งแต่ปีแรก บริษัทจึงเข้าเกณฑ์ **PCI-DSS Level 1** โดยตรง (ต้องมี QSA ทำ audit ออก RoC + quarterly ASV scan + annual penetration test)

**เป้าหมายด้านคุณภาพบริการ (NFR):** ความพร้อมใช้งาน ≥ 99.95% (payment core), auth latency p99 < 800 ms, RPO ≤ 5 นาที, RTO ≤ 30 นาที, จัดเก็บข้อมูลในไทยตามข้อกำหนด ธปท./PDPA

---

## 4. พันธมิตรและผู้ให้บริการภายนอก (Partners & Outsourcing)

| บทบาท | สถานะ | หมายเหตุ |
|-------|-------|---------|
| Sponsor bank / acquiring bank | **[TODO — ยังไม่สรุป]** | จำเป็นสำหรับเส้นทาง B; กำหนดเวลา scheme certification |
| Card schemes (Visa, Mastercard) | **[TODO — ระหว่างเจรจา]** | ต้องผ่าน certification ก่อน go-live |
| ผู้ให้บริการ 3DS (ACS/DS) — EMV 3DS 2.x | **[TODO — คัดเลือก vendor]** | รองรับ challenge/frictionless flow |
| HSM/KMS provider | **[TODO — คัดเลือก]** | key management ตาม PCI Req 3 (dual control, split knowledge) |
| QSA (Qualified Security Assessor) | **[TODO — คัดเลือก vendor]** | ออก RoC ประจำปี — เป็น critical path ล็อกคิวแต่เนิ่น ๆ |
| ASV (Approved Scanning Vendor) | **[TODO — คัดเลือก]** | quarterly external scan |
| ที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. | **[TODO — แต่งตั้ง]** | ทานคำขอและเอกสารก่อนยื่น |
| Local settlement rails (เช่น ITMX) | อยู่ในแผน | สำหรับ local payment ในเฟสถัดไป |

> **หมายเหตุสำคัญ (callout):** ยังไม่มีการลงนามสัญญากับ **sponsor bank**, **QSA** และ **3DS vendor** ณ วันจัดทำเอกสาร — รายการทั้งหมดที่ทำเครื่องหมาย [TODO] ต้องได้ข้อสรุป (ชื่อคู่สัญญา + หนังสือแสดงเจตจำนง/สัญญา) ก่อนยื่นคำขอต่อ ธปท. บริษัทจะไม่ระบุชื่อคู่สัญญาที่ยังไม่ได้ยืนยันในคำขอฉบับจริง

---

## 5. ความพร้อม (Readiness)

### 5.1 ความพร้อมด้านเทคนิค
สถาปัตยกรรมออกแบบตาม Clean Architecture (Go/Fiber) รองรับการยื่นขอใบอนุญาต ครอบคลุม payment core (authorize/capture/void/refund + state machine), tokenization vault, acquirer/3DS adapter, ledger, webhook, reconciliation & settlement (ดู `ARCHITECTURE.md`)

### 5.2 ความพร้อมด้านการกำกับดูแล (Compliance Readiness)

| ด้าน | มาตรฐาน/หน่วยงาน | สถานะ |
|------|------------------|-------|
| นโยบายบริหารความเสี่ยง IT / Cyber resilience | ประกาศ ธปท. | มีเอกสารนโยบาย (ทานต่อ) |
| AML/KYC/CDD + sanction screening | **ปปง. (AMLO)** | มีนโยบาย + ระบบ screening (ตั้งค่า) |
| คุ้มครองข้อมูลส่วนบุคคล | **PDPA / PDPC** | มีนโยบาย + DPO *(TODO แต่งตั้ง)* |
| PCI-DSS v4.0 (Level 1) | QSA/ASV | เตรียม scope + segmentation |
| BCP/DR | ประกาศ ธปท. | มีแผน + ต้องซ้อม DR drill |
| Audit trail | ธปท./PCI Req 10 | `audit_log` append-only พร้อมใช้ |

### 5.3 Timeline (คู่ขนาน engineering + compliance)

| ระยะ | ช่วงเวลา | ผลลัพธ์ |
|------|---------|--------|
| เตรียมนิติบุคคล/ทุน | ส.ค.–ก.ย. 2569 | ทุนชำระแล้ว 50 ล้านบาท, แก้วัตถุประสงค์, โครงสร้างกรรมการ |
| จัดทำเอกสาร + ยื่น ธปท. | ก.ย. 2569 – ม.ค. 2570 | ชุดคำขอครบถ้วน |
| PCI-DSS L1 (QSA→RoC) | ต.ค. 2569 – ก.พ. 2570 | RoC + quarterly ASV scan |
| Scheme certification | หลังได้ sponsor bank | cert Visa/MC |
| Go-live | หลังได้ใบอนุญาต + RoC + cert | production cutover |

> ระยะพิจารณาของ ธปท. โดยทั่วไปเป็นหลัก **หลายเดือน** ขึ้นกับความครบถ้วนของเอกสารและความพร้อมของระบบ (on-site inspection ได้)

---

## 6. ทุนและคุณสมบัติผู้ขอ (Capital & Eligibility)

| หัวข้อ | เกณฑ์ | สถานะบริษัท |
|--------|-------|-------------|
| ประเภทนิติบุคคล | บริษัทจำกัด/มหาชน จดทะเบียนในไทย | ✔ *(ตรวจวัตถุประสงค์)* |
| ทุนจดทะเบียนชำระแล้ว | ≥ **50 ล้านบาท** (Acquiring) | **50 ล้านบาท** |
| การรักษาทุน | คงไว้ **ไม่น้อยกว่า 75%** ตลอดการดำเนินงาน | รักษา ≥ 37.5 ล้านบาท |
| กรรมการสัญชาติไทย | ≥ 1 คน มีถิ่นที่อยู่ในไทย | ✔ *(ยืนยัน)* |
| Fit & proper กรรมการ/ผู้บริหาร | ไม่มีลักษณะต้องห้าม | เตรียมเอกสาร |
| Foreign Business License | หากผู้ถือหุ้นข้างมากต่างชาติ | *[TODO ตรวจโครงสร้างผู้ถือหุ้น]* |

> **[สมมติฐาน / TODO]** ทุนจดทะเบียนชำระแล้ว 50 ล้านบาทเป็นตัวเลขตามเกณฑ์ขั้นต่ำของบริการ Acquiring — ต้องยืนยันด้วยหลักฐานการชำระทุนจริง (bank confirmation) และงบการเงินก่อนยื่น หากขอหลายบริการพร้อมกันให้ยึดจำนวน "สูงสุด" ของบริการที่ขอ

---

## 7. สรุปและขั้นถัดไป

[บริษัท / Company] มีสถาปัตยกรรมและกรอบการกำกับดูแลที่พร้อมสนับสนุนการขอใบอนุญาต Full Acquiring โดยความเสี่ยงตามลำดับ critical path คือ (1) การได้ **sponsor bank** และ scheme certification (2) **PCI-DSS L1 (QSA)** และ (3) ความครบถ้วนของคำขอต่อ ธปท. **ขั้นถัดไป:** แต่งตั้งที่ปรึกษากฎหมาย, ล็อก QSA, ยืนยัน sponsor bank และปิดรายการ [TODO] ทั้งหมดก่อนยื่น

---
---

# Executive summary for the BOT Full-Acquiring license application: business model, scale, partners, readiness, capital (English)

> Document `docs/compliance/00-executive-summary.md` · Version 0.1 · Prepared 22 Jul 2026
> Application type: **Full Acquiring Service license** under the **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT / ธปท.)**
> Applicant: **[บริษัท / Company]** · Paid-up registered capital: **THB 50,000,000**
>
> **Document status:** Internal pre-submission draft — must be reviewed by BOT-licensing legal counsel before actual filing. This is an executive summary, not legal advice.

---

## 1. Purpose of the application

[บริษัท / Company] seeks a **Designated Payment Service license for Acquiring Service** under the Payment Systems Act B.E. 2560. The license is granted by the **Minister of Finance on the recommendation of the BOT**.

The company pursues **Path B — Full Acquiring** (per `ROADMAP.md`): directly acquiring, authorizing, clearing and settling card transactions via a sponsor bank and card schemes, rather than fronting an existing acquirer (which would be Path A — Payment Facilitating, THB 10M capital).

---

## 2. Business model

| Item | Substance |
|------|-----------|
| Core service | Card acquiring for e-commerce, in-app and POS merchants |
| Payment methods | Visa, Mastercard credit/debit (+ local schemes in a later phase) with EMV 3DS 2.x |
| Revenue | Merchant Discount Rate (MDR), per-transaction fees, settlement/payout fees, value-added services (tokenization, reporting) |
| Money handling | Amounts held as integer minor units (satang), recorded in an **append-only double-entry ledger** as source of truth; merchants paid on settlement cycles |
| Target segment | High-volume Thai merchants seeking lower per-transaction cost |
| No card storage | Core system **stores no full PAN/CVV/PIN/track** — only `card_brand` + `card_last4`; a segregated tokenization vault minimizes PCI scope |

**Risk-driven design principles** (from `ARCHITECTURE.md`): idempotency on every money-moving endpoint, fail-closed behavior, and full auditability with every state change written to `audit_log`.

---

## 3. Scale and projections

> **[ASSUMPTION / TODO]** Transaction-count and GMV figures below are planning estimates and must be replaced with numbers from an approved business plan / financial model before BOT submission.

| Metric | Year 1 | Year 2 | Year 3 |
|--------|--------|--------|--------|
| Active merchants | 300 | 1,200 | 3,500 |
| Transactions / year | > 6M | ~25M | ~60M |
| GMV (approx.) | *[TODO]* | *[TODO]* | *[TODO]* |
| PCI-DSS level | **Level 1** | Level 1 | Level 1 |

Because projected volume **exceeds 6M transactions/year from Year 1**, the company falls directly under **PCI-DSS Level 1** (QSA-issued RoC + quarterly ASV scan + annual penetration test).

**Service-quality targets (NFRs):** availability ≥ 99.95% (payment core), auth latency p99 < 800 ms, RPO ≤ 5 min, RTO ≤ 30 min, data residency in Thailand per BOT/PDPA requirements.

---

## 4. Partners & outsourcing

| Role | Status | Note |
|------|--------|------|
| Sponsor / acquiring bank | **[TODO — unresolved]** | Required for Path B; drives scheme certification timeline |
| Card schemes (Visa, Mastercard) | **[TODO — in negotiation]** | Certification required before go-live |
| 3DS provider (ACS/DS) — EMV 3DS 2.x | **[TODO — vendor selection]** | Supports challenge / frictionless flows |
| HSM/KMS provider | **[TODO — selection]** | Key management per PCI Req 3 (dual control, split knowledge) |
| QSA (Qualified Security Assessor) | **[TODO — vendor selection]** | Annual RoC — critical path; book slot early |
| ASV (Approved Scanning Vendor) | **[TODO — selection]** | Quarterly external scan |
| BOT-licensing legal counsel | **[TODO — engagement]** | Review application before filing |
| Local settlement rails (e.g. ITMX) | Planned | For local payments in a later phase |

> **Important callout:** As of this draft, no agreements are signed with the **sponsor bank**, **QSA**, or **3DS vendor**. All items marked [TODO] must be resolved (counterparty name + LOI/contract) before filing with the BOT. The company will not name unconfirmed counterparties in the actual application.

---

## 5. Readiness

### 5.1 Technical readiness
The architecture follows Clean Architecture (Go/Fiber) and supports the license application, covering payment core (authorize/capture/void/refund + state machine), tokenization vault, acquirer/3DS adapters, ledger, webhooks, and reconciliation & settlement (see `ARCHITECTURE.md`).

### 5.2 Compliance readiness

| Area | Standard / authority | Status |
|------|----------------------|--------|
| IT risk / cyber-resilience policy | BOT notifications | Policy drafted (under review) |
| AML/KYC/CDD + sanction screening | **AMLO (ปปง.)** | Policy + screening (being configured) |
| Personal data protection | **PDPA / PDPC** | Policy + DPO *(TODO appoint)* |
| PCI-DSS v4.0 (Level 1) | QSA/ASV | Scope + segmentation in preparation |
| BCP/DR | BOT notifications | Plan in place; DR drill pending |
| Audit trail | BOT / PCI Req 10 | Append-only `audit_log` ready |

### 5.3 Timeline (parallel engineering + compliance)

| Stage | Window | Deliverable |
|-------|--------|-------------|
| Entity/capital prep | Aug–Sep 2026 | THB 50M paid up, objectives amended, board structure |
| Documentation + BOT filing | Sep 2026 – Jan 2027 | Complete application package |
| PCI-DSS L1 (QSA→RoC) | Oct 2026 – Feb 2027 | RoC + quarterly ASV scan |
| Scheme certification | After sponsor bank | Visa/MC certification |
| Go-live | After license + RoC + cert | Production cutover |

> BOT review typically takes **several months**, depending on completeness of documents and system readiness (on-site inspection possible).

---

## 6. Capital & eligibility

| Item | Requirement | Company status |
|------|-------------|----------------|
| Legal entity | Limited/public company registered in Thailand | ✔ *(verify objectives)* |
| Paid-up registered capital | ≥ **THB 50M** (Acquiring) | **THB 50M** |
| Capital maintenance | Maintain **≥ 75%** at all times | Maintain ≥ THB 37.5M |
| Thai-national director | ≥ 1, resident in Thailand | ✔ *(confirm)* |
| Fit & proper directors/execs | No prohibited characteristics | Documents in prep |
| Foreign Business License | If majority foreign-owned | *[TODO check ownership]* |

> **[ASSUMPTION / TODO]** THB 50M paid-up capital reflects the minimum threshold for the Acquiring service — must be confirmed with evidence of actual payment (bank confirmation) and financial statements before filing. If multiple services are requested together, the **highest** applicable threshold governs.

---

## 7. Conclusion & next steps

[บริษัท / Company] has an architecture and compliance framework ready to support a Full Acquiring license application. The critical-path risks, in order, are (1) securing a **sponsor bank** and scheme certification, (2) **PCI-DSS L1 (QSA)**, and (3) completeness of the BOT application. **Next steps:** engage legal counsel, book the QSA, confirm the sponsor bank, and close all [TODO] items before filing.
