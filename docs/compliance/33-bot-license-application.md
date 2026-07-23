# แบบฟอร์มและเอกสารประกอบการยื่นขอใบอนุญาตต่อ ธปท. (ไทย)

> เอกสารเลขที่ `docs/compliance/33-bot-license-application.md` · เวอร์ชัน 0.1 · วันที่จัดทำ 22 ก.ค. 2569
> ประเภทคำขอ: **ใบอนุญาตให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service) แบบเต็มรูปแบบ**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ออกใบอนุญาตโดย **รัฐมนตรีว่าการกระทรวงการคลัง ตามคำแนะนำของ ธปท.**
> ผู้ขอ: **[บริษัท / Company]** · ทุนจดทะเบียนชำระแล้ว: **50,000,000 บาท** · เป้าหมายมาตรฐาน: **PCI-DSS v4.0 Level 1**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อการภายใน (pre-submission draft) — ต้องผ่านการทานโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ก่อนยื่นจริง เอกสารนี้เป็น **แพ็กเกจปะหน้า (cover package)** ที่รวบรวมและอ้างอิงเอกสารประกอบทั้งชุด (`00`–`25`) มิใช่คำแนะนำทางกฎหมาย

---

## 0. วัตถุประสงค์ของเอกสารนี้

เอกสารฉบับนี้เป็น **แพ็กเกจปะหน้าคำขอ (application package)** ที่ผู้ยื่นใช้ประกอบการยื่นต่อ ธปท. ประกอบด้วย 4 ส่วน:

1. **แบบฟอร์มปะหน้า (cover form)** — ข้อมูลผู้ขอ ประเภทบริการ และการรับรองความถูกต้อง
2. **สารบัญเอกสารแนบ (checklist of attachments)** — รายการเอกสารครบชุด พร้อมสถานะและผู้รับผิดชอบ
3. **บทสรุปผู้บริหาร (executive summary)** — สรุปย่อ อ้างอิงเอกสารเต็มที่ `00-executive-summary.md`
4. **คู่มือการยื่น (submission guide)** — ช่องทาง ลำดับขั้น timeline และการตอบคำถามเพิ่มเติมของ ธปท.

> **หมายเหตุ:** แบบฟอร์มทางการของ ธปท. อาจปรับปรุงเป็นครั้งคราว เอกสารนี้จัดโครงสร้างฟิลด์ให้ครอบคลุมข้อมูลที่ ธปท. ต้องการ แต่ต้องนำไปกรอกลงในแบบฟอร์มทางการฉบับล่าสุดที่ ธปท. เผยแพร่ ณ วันยื่น

---

## 1. แบบฟอร์มปะหน้าคำขอ (Cover Form)

### 1.1 ข้อมูลผู้ขออนุญาต

| ฟิลด์ | ข้อมูล |
|-------|--------|
| ชื่อนิติบุคคล (ไทย/อังกฤษ) | **[บริษัท / Company]** |
| เลขทะเบียนนิติบุคคล | *[TODO — เลข 13 หลัก จากหนังสือรับรอง DBD]* |
| ประเภทนิติบุคคล | บริษัทจำกัด / บริษัทมหาชนจำกัด (จดทะเบียนในไทย) |
| ที่ตั้งสำนักงานใหญ่ | *[TODO — ที่อยู่ตามหนังสือรับรอง]* |
| ทุนจดทะเบียน / ทุนชำระแล้ว | 50,000,000 บาท / **50,000,000 บาท (ชำระเต็ม)** |
| วันจดทะเบียนบริษัท | *[TODO]* |
| ผู้มีอำนาจลงนามผูกพัน | *[TODO — ตามหนังสือรับรอง]* |
| ผู้ประสานงาน (contact person) | *[TODO — ชื่อ / ตำแหน่ง / โทร / อีเมล]* |

### 1.2 ประเภทบริการที่ขออนุญาต

| รายการ | ค่า |
|--------|-----|
| กลุ่มบริการ | บริการการชำระเงินภายใต้การกำกับ (Designated Payment Services) |
| บริการที่ขอ | **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)** |
| ทุนขั้นต่ำตามเกณฑ์ | 50 ล้านบาท (ยึดจำนวนสูงสุดหากขอหลายบริการ) |
| ขอบเขต | รับ-ประมวลผล authorization / clearing / settlement ของธุรกรรมบัตร (Visa, Mastercard) พร้อม EMV 3DS 2.x |
| ช่องทางที่รองรับ | e-commerce, in-app, POS |
| แผนบริการเพิ่มเติมในอนาคต | local schemes / QR payment (เฟสถัดไป — ดู `../QR-PAYMENT.md`) |

### 1.3 คำรับรองของผู้ขอ (Declaration)

ข้าพเจ้าในนาม **[บริษัท / Company]** ขอรับรองว่า:

- ข้อมูลและเอกสารประกอบทั้งหมดถูกต้อง ครบถ้วน และเป็นความจริง
- บริษัทมีทุนจดทะเบียนชำระแล้ว **ไม่น้อยกว่า 50 ล้านบาท** และจะดำรงทุนไว้ **ไม่น้อยกว่า 75%** (≥ 37.5 ล้านบาท) ตลอดการดำเนินงาน
- กรรมการและผู้บริหารมีคุณสมบัติเหมาะสม (fit & proper) และไม่มีลักษณะต้องห้ามตามกฎหมาย
- บริษัทจะปฏิบัติตาม พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560, ประกาศ ธปท. ที่เกี่ยวข้อง, พ.ร.บ. ปปง., และ PDPA
- บริษัทยินยอมให้ ธปท. ขอข้อมูลเพิ่มเติมและเข้าตรวจสอบ ณ สถานที่ประกอบการ (on-site inspection)

ลงนาม: __________________________ (กรรมการผู้มีอำนาจ) · วันที่: ____________ · ประทับตราบริษัท

---

## 2. สารบัญเอกสารแนบ (Checklist of Attachments)

> คอลัมน์ "เอกสารอ้างอิง" ชี้ไปยังไฟล์ในชุด `docs/compliance/` ที่รองรับหัวข้อนั้น สถานะ: ✅ พร้อม · 🟡 ร่าง/ต้องทาน · 🔴 ยังไม่มี/รอ external

### หมวด A — นิติบุคคลและทุน

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| A1 | หนังสือรับรองบริษัท + วัตถุประสงค์ (ครอบคลุมธุรกิจ payment) | *(นิติบุคคล)* | Legal | 🔴 |
| A2 | หนังสือบริคณห์สนธิ / ข้อบังคับบริษัท | *(นิติบุคคล)* | Legal | 🔴 |
| A3 | หลักฐานการชำระทุน 50 ล้านบาท (bank confirmation) | `02-financial-projections-capital.md` | Finance | 🔴 |
| A4 | บัญชีรายชื่อผู้ถือหุ้น (บอจ.5) + โครงสร้างการถือหุ้น | `03-shareholder-board-fit-proper.md` | Legal | 🟡 |
| A5 | Foreign Business License (หากผู้ถือหุ้นข้างมากต่างชาติ) | `03-shareholder-board-fit-proper.md` | Legal | 🔴 (ตรวจสอบ) |

### หมวด B — ธรรมาภิบาลและบุคลากร

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| B1 | ผังโครงสร้างองค์กรและการกำกับดูแล | `04-org-chart-governance.md` | Compliance | 🟡 |
| B2 | ประวัติ + เอกสาร fit & proper กรรมการ/ผู้บริหาร | `03-shareholder-board-fit-proper.md` | Legal | 🟡 |
| B3 | นโยบายแบ่งแยกหน้าที่ (segregation of duties) | `18-segregation-of-duties.md` | Compliance | 🟡 |
| B4 | หนังสือแต่งตั้ง DPO และเจ้าหน้าที่กำกับ AML | `09-pdpa-privacy-policy.md`, `05-aml-kyc-cdd-policy.md` | Compliance | 🔴 (แต่งตั้ง) |

### หมวด C — แผนธุรกิจและการเงิน

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| C1 | แผนธุรกิจ (business plan) | `01-business-plan.md` | Product/Finance | 🟡 |
| C2 | ประมาณการการเงิน 3 ปี + สมมติฐาน | `02-financial-projections-capital.md` | Finance | 🟡 |
| C3 | บทสรุปผู้บริหาร | `00-executive-summary.md` | Compliance | ✅ |

### หมวด D — AML / KYC / CDD

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| D1 | นโยบาย AML/KYC/CDD | `05-aml-kyc-cdd-policy.md` | Compliance | 🟡 |
| D2 | กระบวนการ sanction/PEP screening | `06-sanctions-screening.md` | Compliance | 🟡 |
| D3 | ขั้นตอนรายงานธุรกรรมน่าสงสัย (SAR/STR) ต่อ ปปง. | `07-sar-str-procedure.md` | Compliance | 🟡 |
| D4 | การจัดระดับความเสี่ยงลูกค้า (risk categorization) | `08-customer-risk-categorization.md` | Compliance | 🟡 |

### หมวด E — คุ้มครองข้อมูลส่วนบุคคล (PDPA)

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| E1 | นโยบายความเป็นส่วนตัว (privacy policy) | `09-pdpa-privacy-policy.md` | DPO | 🟡 |
| E2 | แม่แบบข้อตกลงประมวลผลข้อมูล (DPA) | `10-dpa-templates.md` | Legal/DPO | 🟡 |
| E3 | นโยบายเก็บรักษาและลบข้อมูล (retention/deletion) | `11-data-retention-deletion.md` | DPO | 🟡 |
| E4 | ขั้นตอนสิทธิเจ้าของข้อมูล (DSAR) | `12-dsar-workflow.md` | DPO | 🟡 |

### หมวด F — บริหารความเสี่ยง IT / Cyber / ความต่อเนื่อง

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| F1 | นโยบายบริหารความเสี่ยงด้าน IT | `13-it-risk-management.md` | Security | 🟡 |
| F2 | แผน cyber resilience | `14-cyber-resilience.md` | Security | 🟡 |
| F3 | แผนความต่อเนื่องทางธุรกิจ + DR (BCP/DR) | `15-bcp-dr.md` | SRE | 🟡 |
| F4 | แผนตอบสนองเหตุการณ์ + การแจ้งเหตุข้อมูลรั่วไหล | `16-incident-response-breach.md` | Security | 🟡 |
| F5 | นโยบายการบริหารการเปลี่ยนแปลง (change management) | `17-change-management.md` | SRE | 🟡 |

### หมวด G — PCI-DSS และความปลอดภัยข้อมูลบัตร

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| G1 | แผนงาน PCI-DSS v4.0 Level 1 | `19-pci-dss-roadmap.md` | Security | 🟡 |
| G2 | ผัง network segmentation / CDE | `20-network-segmentation-cde.md` | Security | 🟡 |
| G3 | Tokenization + HSM/KMS + key management | `21-tokenization-hsm-keymgmt.md` | Security | 🟡 |
| G4 | กลยุทธ์ EMV 3DS 2.x | `22-3ds-strategy.md` | Product/Security | 🟡 |
| G5 | แผน QSA / ASV / penetration test | `23-qsa-asv-pentest-plan.md` | Security | 🟡 |
| G6 | แผน scheme certification (Visa/Mastercard) | `24-scheme-certification-plan.md` | Product | 🔴 (รอ sponsor bank) |

### หมวด H — สถาปัตยกรรมและการปฏิบัติงาน

| # | เอกสาร | เอกสารอ้างอิง | ผู้รับผิดชอบ | สถานะ |
|---|--------|----------------|--------------|-------|
| H1 | เอกสารสถาปัตยกรรมระบบ + การไหลของธุรกรรม/เงิน | `../ARCHITECTURE.md`, `../DIAGRAMS.md` | Engineering | ✅ |
| H2 | สัญญาผู้ค้า / ข้อกำหนดการใช้บริการ (TOS) | `25-merchant-agreement-tos.md` | Legal | 🟡 |
| H3 | Roadmap / timeline | `../ROADMAP.md` | Product | ✅ |
| H4 | แผน outsourcing / ผู้ให้บริการภายนอก | *(รวมใน `13`/`24`)* | Compliance | 🟡 |

> **[สมมติฐาน / TODO — external dependencies]** รายการที่เป็น 🔴 ส่วนใหญ่ผูกกับปัจจัยภายนอกที่ยังไม่สรุป ได้แก่ **(1) sponsor bank / acquiring bank** (กำหนดเวลา scheme certification), **(2) QSA vendor** (ออก RoC), **(3) 3DS vendor**, และ **(4) หลักฐานการชำระทุน 50 ล้านบาทจริง** ห้ามระบุชื่อคู่สัญญาที่ยังไม่ยืนยันในคำขอฉบับจริง และต้องปิดรายการเหล่านี้ก่อนยื่น

---

## 3. บทสรุปผู้บริหาร (Executive Summary — ย่อ)

> ฉบับเต็มอยู่ที่ `00-executive-summary.md` — ส่วนนี้เป็นบทสรุปย่อสำหรับปะหน้า

**[บริษัท / Company]** ขอ **ใบอนุญาต Acquiring Service แบบเต็มรูปแบบ (เส้นทาง B)** ตาม พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 ให้บริการรับชำระเงินด้วยบัตร Visa/Mastercard สำหรับร้านค้า e-commerce/in-app/POS พร้อม EMV 3DS 2.x รายได้จาก MDR + ค่าธรรมเนียมต่อรายการ + ค่า settlement

- **ขนาด:** ปีแรกคาดปริมาณ > 6 ล้านรายการ/ปี → เข้าเกณฑ์ **PCI-DSS Level 1** โดยตรง *(ตัวเลขเป็นประมาณการ — TODO ยืนยันจาก financial model)*
- **ทุน:** ชำระแล้ว 50 ล้านบาท ดำรงไว้ ≥ 75%
- **สถาปัตยกรรม:** Clean Architecture (Go/Fiber), ledger double-entry append-only, ไม่เก็บ full PAN/CVV/PIN (เก็บเฉพาะ `card_brand` + `card_last4`), tokenization vault แยก segment เพื่อลด PCI scope, idempotency + fail-closed + audit trail ครบทุก state change
- **NFR:** availability ≥ 99.95%, auth p99 < 800 ms, RPO ≤ 5 นาที, RTO ≤ 30 นาที, data residency ในไทย
- **Critical path:** (1) sponsor bank + scheme certification (2) PCI-DSS L1 (QSA→RoC) (3) ความครบถ้วนของคำขอ

---

## 4. คู่มือการยื่น (Submission Guide)

### 4.1 ช่องทางและผู้รับเรื่อง

| หัวข้อ | รายละเอียด |
|--------|-----------|
| หน่วยงานผู้กำกับ | ธนาคารแห่งประเทศไทย (ธปท.) — ฝ่ายที่รับผิดชอบระบบการชำระเงิน |
| ผู้ออกใบอนุญาต | รัฐมนตรีว่าการกระทรวงการคลัง (ตามคำแนะนำของ ธปท.) |
| ช่องทางยื่น | ยื่นตามช่องทางที่ ธปท. กำหนด (หนังสือถึง ธปท. / ระบบอิเล็กทรอนิกส์ที่ ธปท. ประกาศ) *[TODO ยืนยันช่องทางล่าสุดกับที่ปรึกษา]* |
| ภาษาเอกสาร | ไทยเป็นหลัก (เอกสารต่างประเทศแนบคำแปล) |
| จำนวนชุด | ตามที่ ธปท. กำหนด (ต้นฉบับ + สำเนา) *[TODO ยืนยัน]* |

### 4.2 ลำดับขั้นการยื่นและพิจารณา

1. **Pre-check ภายใน** — ทานเอกสารครบชุด (checklist หมวด A–H) โดย Compliance + ที่ปรึกษากฎหมาย; ปิดรายการ 🔴 ทั้งหมด
2. **ยื่นคำขอ + เอกสารประกอบ** ต่อ ธปท.
3. **ธปท. ตรวจความครบถ้วน (completeness check)** — อาจแจ้งขอเอกสารเพิ่ม
4. **การพิจารณาเนื้อหา + on-site inspection** — ธปท. อาจเข้าตรวจสถานที่/ระบบ (security, AML, BCP, audit trail)
5. **ธปท. เสนอความเห็นต่อรัฐมนตรีว่าการกระทรวงการคลัง** เพื่อออกใบอนุญาต
6. **ได้รับใบอนุญาต** → ปฏิบัติตามเงื่อนไข + รายงานเป็นงวด + ประเมินต่อเนื่อง

### 4.3 Timeline (คู่ขนาน engineering + compliance)

| ระยะ | ช่วงเวลา | ผลลัพธ์ |
|------|---------|--------|
| เตรียมนิติบุคคล/ทุน | ส.ค.–ก.ย. 2569 | ทุนชำระแล้ว 50 ล้านบาท, แก้วัตถุประสงค์, โครงสร้างกรรมการ |
| จัดทำเอกสาร + ยื่น ธปท. | ก.ย. 2569 – ม.ค. 2570 | ชุดคำขอครบถ้วน |
| PCI-DSS L1 (QSA→RoC) | ต.ค. 2569 – ก.พ. 2570 | RoC + quarterly ASV scan |
| Scheme certification | หลังได้ sponsor bank | cert Visa/MC |
| Go-live | หลังได้ใบอนุญาต + RoC + cert | production cutover |

> ระยะพิจารณาของ ธปท. โดยทั่วไปเป็นหลัก **หลายเดือน** ขึ้นกับความครบถ้วนของเอกสารและความพร้อมของระบบ

### 4.4 การตอบข้อสอบถามเพิ่มเติมของ ธปท. (RFI handling)

- กำหนดผู้ประสานงานเดียว (single point of contact) และ log ทุกคำถาม/คำตอบ
- ตอบภายในกรอบเวลาที่ ธปท. กำหนด (ปกติกำหนดเป็นวันทำการ) — หากต้องใช้เวลา ให้แจ้งขอขยายเป็นลายลักษณ์อักษร
- อ้างอิงเอกสารในชุด `docs/compliance/` ให้ตรงหมวดที่ถูกถาม

### 4.5 หลังได้รับใบอนุญาต (Post-license obligations)

- ดำรงทุนชำระแล้ว ≥ 75% และแจ้ง ธปท. เมื่อมีการเปลี่ยนแปลงสาระสำคัญ
- รายงานเป็นงวดตามที่ ธปท. กำหนด (operational / incident / financial)
- ต่ออายุ PCI-DSS ประจำปี (RoC + quarterly ASV + annual pentest)
- แจ้งเหตุการณ์สำคัญ (major incident / data breach) ตามกรอบเวลาของ ธปท. และ PDPC

---
---

# BOT license application package: cover form, checklist of attachments, executive summary, submission guide (English)

> Document `docs/compliance/33-bot-license-application.md` · Version 0.1 · Prepared 22 Jul 2026
> Application type: **Full Acquiring Service license** under the **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT / ธปท.)**, granted by the **Minister of Finance on the recommendation of the BOT**.
> Applicant: **[บริษัท / Company]** · Paid-up registered capital: **THB 50,000,000** · Target standard: **PCI-DSS v4.0 Level 1**
>
> **Document status:** Internal pre-submission draft — must be reviewed by BOT-licensing legal counsel before actual filing. This is the **cover package** consolidating and referencing the full attachment set (`00`–`25`); it is not legal advice.

---

## 0. Purpose of this document

This is the **application cover package** used to file with the BOT. It has four parts:

1. **Cover form** — applicant identity, requested service, and declaration of accuracy.
2. **Checklist of attachments** — the complete document set with status and owner.
3. **Executive summary** — a short abstract; the full version lives at `00-executive-summary.md`.
4. **Submission guide** — channel, sequence, timeline, and how to handle BOT follow-ups.

> **Note:** The BOT's official form may be revised from time to time. This document structures the fields to cover the information the BOT requires, but the data must be transcribed into the latest official BOT form published at the time of filing.

---

## 1. Application cover form

### 1.1 Applicant details

| Field | Value |
|-------|-------|
| Legal name (Thai/English) | **[บริษัท / Company]** |
| Company registration number | *[TODO — 13-digit number from DBD certificate]* |
| Entity type | Private / public limited company (registered in Thailand) |
| Registered head office | *[TODO — address per company affidavit]* |
| Registered / paid-up capital | THB 50,000,000 / **THB 50,000,000 (fully paid)** |
| Incorporation date | *[TODO]* |
| Authorized signatory | *[TODO — per company affidavit]* |
| Contact person | *[TODO — name / title / phone / email]* |

### 1.2 Service applied for

| Item | Value |
|------|-------|
| Service group | Designated Payment Services |
| Requested service | **Acquiring Service (electronic payment acceptance)** |
| Minimum capital threshold | THB 50M (highest applicable threshold governs if multiple services) |
| Scope | Acquiring, authorization, clearing and settlement of card transactions (Visa, Mastercard) with EMV 3DS 2.x |
| Channels | e-commerce, in-app, POS |
| Future services | local schemes / QR payment (later phase — see `../QR-PAYMENT.md`) |

### 1.3 Applicant declaration

On behalf of **[บริษัท / Company]**, we declare that:

- All information and supporting documents are accurate, complete, and true.
- The company has paid-up capital of **no less than THB 50M** and will maintain **no less than 75%** (≥ THB 37.5M) throughout operations.
- Directors and executives are fit and proper and have no legally prohibited characteristics.
- The company will comply with the Payment Systems Act B.E. 2560, relevant BOT notifications, the AMLO Act, and the PDPA.
- The company consents to the BOT requesting further information and conducting on-site inspection.

Signed: __________________________ (authorized director) · Date: ____________ · Company seal

---

## 2. Checklist of attachments

> The "Reference" column points to the file in `docs/compliance/` that supports each item. Status: ✅ ready · 🟡 draft/needs review · 🔴 missing/awaiting external.

### Section A — Entity & capital

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| A1 | Company affidavit + objectives (covering payment business) | *(corporate)* | Legal | 🔴 |
| A2 | Memorandum / Articles of Association | *(corporate)* | Legal | 🔴 |
| A3 | Evidence of THB 50M capital payment (bank confirmation) | `02-financial-projections-capital.md` | Finance | 🔴 |
| A4 | Shareholder list (Bor.Or.Jor.5) + ownership structure | `03-shareholder-board-fit-proper.md` | Legal | 🟡 |
| A5 | Foreign Business License (if majority foreign-owned) | `03-shareholder-board-fit-proper.md` | Legal | 🔴 (verify) |

### Section B — Governance & personnel

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| B1 | Org chart & governance structure | `04-org-chart-governance.md` | Compliance | 🟡 |
| B2 | Director/executive CVs + fit & proper documents | `03-shareholder-board-fit-proper.md` | Legal | 🟡 |
| B3 | Segregation-of-duties policy | `18-segregation-of-duties.md` | Compliance | 🟡 |
| B4 | DPO and AML compliance officer appointment letters | `09-pdpa-privacy-policy.md`, `05-aml-kyc-cdd-policy.md` | Compliance | 🔴 (appoint) |

### Section C — Business plan & finance

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| C1 | Business plan | `01-business-plan.md` | Product/Finance | 🟡 |
| C2 | 3-year financial projections + assumptions | `02-financial-projections-capital.md` | Finance | 🟡 |
| C3 | Executive summary | `00-executive-summary.md` | Compliance | ✅ |

### Section D — AML / KYC / CDD

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| D1 | AML/KYC/CDD policy | `05-aml-kyc-cdd-policy.md` | Compliance | 🟡 |
| D2 | Sanction/PEP screening process | `06-sanctions-screening.md` | Compliance | 🟡 |
| D3 | Suspicious transaction reporting (SAR/STR) to AMLO | `07-sar-str-procedure.md` | Compliance | 🟡 |
| D4 | Customer risk categorization | `08-customer-risk-categorization.md` | Compliance | 🟡 |

### Section E — Personal data protection (PDPA)

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| E1 | Privacy policy | `09-pdpa-privacy-policy.md` | DPO | 🟡 |
| E2 | Data Processing Agreement (DPA) templates | `10-dpa-templates.md` | Legal/DPO | 🟡 |
| E3 | Data retention/deletion policy | `11-data-retention-deletion.md` | DPO | 🟡 |
| E4 | DSAR (data subject rights) workflow | `12-dsar-workflow.md` | DPO | 🟡 |

### Section F — IT / cyber risk & continuity

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| F1 | IT risk management policy | `13-it-risk-management.md` | Security | 🟡 |
| F2 | Cyber resilience plan | `14-cyber-resilience.md` | Security | 🟡 |
| F3 | Business continuity + DR plan (BCP/DR) | `15-bcp-dr.md` | SRE | 🟡 |
| F4 | Incident response + breach notification | `16-incident-response-breach.md` | Security | 🟡 |
| F5 | Change management policy | `17-change-management.md` | SRE | 🟡 |

### Section G — PCI-DSS & cardholder data security

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| G1 | PCI-DSS v4.0 Level 1 roadmap | `19-pci-dss-roadmap.md` | Security | 🟡 |
| G2 | Network segmentation / CDE diagram | `20-network-segmentation-cde.md` | Security | 🟡 |
| G3 | Tokenization + HSM/KMS + key management | `21-tokenization-hsm-keymgmt.md` | Security | 🟡 |
| G4 | EMV 3DS 2.x strategy | `22-3ds-strategy.md` | Product/Security | 🟡 |
| G5 | QSA / ASV / penetration test plan | `23-qsa-asv-pentest-plan.md` | Security | 🟡 |
| G6 | Scheme certification plan (Visa/Mastercard) | `24-scheme-certification-plan.md` | Product | 🔴 (awaiting sponsor bank) |

### Section H — Architecture & operations

| # | Document | Reference | Owner | Status |
|---|----------|-----------|-------|--------|
| H1 | System architecture + transaction/money flows | `../ARCHITECTURE.md`, `../DIAGRAMS.md` | Engineering | ✅ |
| H2 | Merchant agreement / terms of service | `25-merchant-agreement-tos.md` | Legal | 🟡 |
| H3 | Roadmap / timeline | `../ROADMAP.md` | Product | ✅ |
| H4 | Outsourcing / third-party service plan | *(in `13`/`24`)* | Compliance | 🟡 |

> **[ASSUMPTION / TODO — external dependencies]** Most 🔴 items depend on unresolved external factors: **(1) sponsor / acquiring bank** (drives scheme certification timing), **(2) QSA vendor** (issues the RoC), **(3) 3DS vendor**, and **(4) actual evidence of THB 50M capital payment**. Do not name unconfirmed counterparties in the actual application, and close these items before filing.

---

## 3. Executive summary (abstract)

> Full version at `00-executive-summary.md` — this is a cover-page abstract.

**[บริษัท / Company]** applies for a **Full Acquiring Service license (Path B)** under the Payment Systems Act B.E. 2560, providing Visa/Mastercard card acceptance for e-commerce/in-app/POS merchants with EMV 3DS 2.x. Revenue is from MDR + per-transaction fees + settlement fees.

- **Scale:** Year-1 volume projected > 6M transactions/year → falls directly under **PCI-DSS Level 1** *(estimate — TODO confirm from financial model)*.
- **Capital:** THB 50M paid up, maintained ≥ 75%.
- **Architecture:** Clean Architecture (Go/Fiber), append-only double-entry ledger, no full PAN/CVV/PIN storage (only `card_brand` + `card_last4`), segregated tokenization vault to minimize PCI scope, idempotency + fail-closed + full audit trail on every state change.
- **NFRs:** availability ≥ 99.95%, auth p99 < 800 ms, RPO ≤ 5 min, RTO ≤ 30 min, data residency in Thailand.
- **Critical path:** (1) sponsor bank + scheme certification, (2) PCI-DSS L1 (QSA→RoC), (3) completeness of the application.

---

## 4. Submission guide

### 4.1 Channel and recipient

| Item | Detail |
|------|--------|
| Regulator | Bank of Thailand (BOT) — payment systems supervision function |
| License issuer | Minister of Finance (on BOT's recommendation) |
| Filing channel | As prescribed by the BOT (formal letter / BOT-published e-channel) *[TODO confirm current channel with counsel]* |
| Document language | Thai primary (foreign documents with certified translation) |
| Copies | Per BOT requirement (original + copies) *[TODO confirm]* |

### 4.2 Filing and review sequence

1. **Internal pre-check** — Compliance + legal counsel verify the full set (checklist Sections A–H); close all 🔴 items.
2. **Submit application + attachments** to the BOT.
3. **BOT completeness check** — may request additional documents.
4. **Substantive review + on-site inspection** — BOT may inspect premises/systems (security, AML, BCP, audit trail).
5. **BOT recommends to the Minister of Finance** for license issuance.
6. **License granted** → comply with conditions + periodic reporting + ongoing assessment.

### 4.3 Timeline (parallel engineering + compliance)

| Stage | Window | Deliverable |
|-------|--------|-------------|
| Entity/capital prep | Aug–Sep 2026 | THB 50M paid up, objectives amended, board structure |
| Documentation + BOT filing | Sep 2026 – Jan 2027 | Complete application package |
| PCI-DSS L1 (QSA→RoC) | Oct 2026 – Feb 2027 | RoC + quarterly ASV scan |
| Scheme certification | After sponsor bank | Visa/MC certification |
| Go-live | After license + RoC + cert | Production cutover |

> BOT review typically takes **several months**, depending on completeness of documents and system readiness.

### 4.4 Handling BOT requests for information (RFI)

- Assign a single point of contact and log every question/answer.
- Respond within the BOT's stated deadline (usually in business days); if more time is needed, request an extension in writing.
- Reference the exact `docs/compliance/` document for the topic asked.

### 4.5 Post-license obligations

- Maintain paid-up capital ≥ 75% and notify the BOT of material changes.
- File periodic reports as required by the BOT (operational / incident / financial).
- Renew PCI-DSS annually (RoC + quarterly ASV + annual pentest).
- Report major incidents / data breaches within BOT and PDPC timeframes.
