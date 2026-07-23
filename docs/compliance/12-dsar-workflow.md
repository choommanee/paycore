# ขั้นตอนคำขอใช้สิทธิของเจ้าของข้อมูล (DSAR) (ไทย)

> เอกสารเลขที่ `docs/compliance/12-dsar-workflow.md` · เวอร์ชัน 0.1 · วันที่จัดทำ 22 ก.ค. 2569
> ประเภทคำขอ: **ใบอนุญาตให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service) แบบเต็มรูปแบบ**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)**
> ผู้ขอ: **[บริษัท / Company]** · ทุนจดทะเบียนชำระแล้ว: **50,000,000 บาท**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อการภายใน (pre-submission draft) — ต้องผ่านการทานโดยที่ปรึกษากฎหมาย/DPO และที่ปรึกษาด้าน PDPA ก่อนใช้จริง เอกสารนี้เป็นกระบวนการปฏิบัติงาน (operating procedure) มิใช่คำแนะนำทางกฎหมาย
>
> เอกสารประกอบ: `COMPLIANCE-TH.md` (กฎหมาย/ใบอนุญาต) · `ARCHITECTURE.md` (สถาปัตยกรรม/PCI scope) · `03-shareholder-board-fit-proper.md` · `04-org-chart-governance.md` (บทบาทและธรรมาภิบาล)

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนดขั้นตอนมาตรฐาน (workflow) สำหรับการรับและดำเนินการ **คำขอใช้สิทธิของเจ้าของข้อมูลส่วนบุคคล (Data Subject Access Request — DSAR)** ตาม **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** ครอบคลุมตั้งแต่ **การรับคำขอ (intake) → การพิสูจน์ตัวตน (verification) → การดำเนินการตามกรอบเวลา (fulfillment SLA) → ข้อยกเว้น (exceptions)**

**สิทธิของเจ้าของข้อมูลที่ครอบคลุม** (มาตรา 30–36 PDPA):

| สิทธิ | มาตรา (โดยประมาณ) | สาระสำคัญ |
|-------|-------------------|-----------|
| สิทธิเข้าถึงและขอรับสำเนา (Access) | ม.30 | ขอเข้าถึงและขอสำเนาข้อมูลของตน + ที่มาของข้อมูล |
| สิทธิให้โอนย้ายข้อมูล (Portability) | ม.31 | ขอรับข้อมูลในรูปแบบที่อ่าน/ใช้งานได้ด้วยเครื่องมืออัตโนมัติ |
| สิทธิคัดค้าน (Object) | ม.32 | คัดค้านการเก็บ/ใช้/เปิดเผยข้อมูล |
| สิทธิให้ลบ/ทำลาย (Erasure) | ม.33 | ขอให้ลบ ทำลาย หรือทำให้ไม่ระบุตัวตน |
| สิทธิให้ระงับการใช้ (Restriction) | ม.34 | ขอให้ระงับการใช้ชั่วคราว |
| สิทธิให้แก้ไข (Rectification) | ม.35–36 | ขอให้แก้ไขให้ถูกต้อง เป็นปัจจุบัน สมบูรณ์ |
| สิทธิถอนความยินยอม (Withdraw consent) | ม.19 | ถอนความยินยอมได้ทุกเมื่อ (เท่าที่ไม่มีฐานอื่นรองรับ) |

**ขอบเขตข้อมูล:** ข้อมูลของ (ก) ผู้ถือบัตร/ผู้ชำระเงิน (cardholder) (ข) ผู้ติดต่อฝั่งร้านค้า (merchant contacts) (ค) ผู้เข้าเว็บ/แอป และ (ง) พนักงาน/ผู้สมัครงาน

**นอกขอบเขต:** ข้อมูลบัตรเต็ม (full PAN/CVV/PIN/track) — ระบบหลัก **ไม่จัดเก็บ** ตามหลัก scope minimization ใน `ARCHITECTURE.md` ข้อ 6 (เก็บได้เฉพาะ `card_brand` + `card_last4`)

---

## 2. บทบาทและความรับผิดชอบ (RACI)

| บทบาท | หน้าที่ในกระบวนการ DSAR |
|-------|-------------------------|
| **DPO (เจ้าหน้าที่คุ้มครองข้อมูลส่วนบุคคล)** | เจ้าของกระบวนการ (process owner); อนุมัติการตอบ; ตัดสินข้อยกเว้น; จุดติดต่อ PDPC |
| **DSAR Coordinator** | รับคำขอ, ลงทะเบียน, ติดตาม SLA, ประสานทีม, ร่างหนังสือตอบ |
| **Identity Verification (IAM/Support)** | พิสูจน์ตัวตนผู้ยื่นคำขอ |
| **Engineering / Data Team** | ค้นหา ดึง (extract) ลบ/แก้ไขข้อมูลจากระบบ (payments, merchants, audit_log, logs) |
| **Legal Counsel** | ประเมินข้อยกเว้นทางกฎหมาย, การชนกันกับ AML/ธปท. |
| **Security (DevSecOps)** | จัดส่งข้อมูลแบบเข้ารหัส, บันทึก access, ตรวจ PCI scope ของ export |
| **Compliance / AMLO Liaison** | ตรวจข้อขัดแย้งกับหน้าที่เก็บรักษาข้อมูลตาม ปปง./ธปท. |

**หลักการแบ่งแยกหน้าที่:** ผู้พิสูจน์ตัวตน ผู้ดึงข้อมูล และผู้อนุมัติการส่งออกต้องไม่ใช่บุคคลเดียวกันสำหรับคำขอที่มีข้อมูลอ่อนไหวหรือปริมาณมาก (สอดคล้องหลัก dual control ตาม PCI Req 7)

> **[สมมติฐาน / TODO]** ยังมิได้แต่งตั้ง **DPO** อย่างเป็นทางการ และยังมิได้ระบุตัวบุคคลใน DSAR Coordinator / Legal Counsel ณ วันจัดทำ ต้องระบุชื่อ-ตำแหน่ง-ช่องทางติดต่อ และแนบคำสั่งแต่งตั้งก่อนใช้จริงและก่อนยื่น ธปท. (สอดคล้อง `04-org-chart-governance.md`)

---

## 3. ช่องทางรับคำขอ (Intake)

| ช่องทาง | รายละเอียด | SLA รับเรื่อง (acknowledge) |
|---------|-----------|----------------------------|
| อีเมล DPO เฉพาะ | `[TODO: dpo@company.co.th]` | ภายใน 3 วันทำการ |
| แบบฟอร์มออนไลน์ (privacy portal) | ฟอร์มในเว็บ/แอป + reCAPTCHA | ทันที (ระบบออกเลขคำขออัตโนมัติ) |
| ไปรษณีย์ลงทะเบียน | ที่อยู่สำนักงานจดทะเบียน | ภายใน 5 วันทำการหลังรับ |
| ผ่านร้านค้า (สำหรับ cardholder) | ร้านค้าส่งต่อมายัง DPO ภายใน 2 วันทำการ | นับ SLA จากวันที่ DPO รับ |

**ข้อมูลขั้นต่ำที่ต้องระบุในคำขอ:** (1) ชื่อ-นามสกุล และช่องทางติดต่อกลับ (2) ความสัมพันธ์กับ [บริษัท / Company] (cardholder / merchant / พนักงาน) (3) สิทธิที่ประสงค์ใช้ (4) ขอบเขต/ช่วงเวลาข้อมูล (ถ้าทราบ) (5) หลักฐานพิสูจน์ตัวตน

**การลงทะเบียน:** ทุกคำขอบันทึกใน **DSAR Register** พร้อม `request_id` (รูปแบบ `DSAR-YYYY-NNNN`), วันรับ, ประเภทสิทธิ, สถานะ, deadline, ผู้รับผิดชอบ, และผลลัพธ์ — เก็บเป็นหลักฐานการปฏิบัติตาม PDPA และแสดงต่อ PDPC/ธปท. เมื่อร้องขอ ทุกการเข้าถึง register บันทึกลง `audit_log` (append-only ตาม `ARCHITECTURE.md`)

---

## 4. การพิสูจน์ตัวตน (Verification)

การพิสูจน์ตัวตนใช้แนวทาง **risk-based** — ยิ่งข้อมูลอ่อนไหว/ปริมาณมาก ยิ่งเข้มขึ้น เพื่อป้องกันการปลอมตัว (impersonation) ซึ่งเป็นความเสี่ยงหลักของ DSAR

| ระดับ | เมื่อใด | หลักฐานที่ต้องใช้ |
|-------|--------|-------------------|
| **L1 — พื้นฐาน** | คำขอเข้าถึง/แก้ไขข้อมูลติดต่อทั่วไป | ยืนยันการควบคุมอีเมล/เบอร์ที่ลงทะเบียน (OTP) |
| **L2 — ยกระดับ** | ข้อมูลธุรกรรมบัตร, portability, erasure | L1 + สำเนาบัตรประชาชน (ปิดบังเลขที่ไม่จำเป็น) + จับคู่ข้อมูลในระบบ เช่น `card_last4` + วันที่/ยอดธุรกรรม |
| **L3 — เข้มสุด** | ข้อมูลจำนวนมาก / ข้อมูลอ่อนไหว / มีข้อสงสัยว่าปลอมตัว | L2 + video KYC หรือยืนยันผ่านช่องทางที่ลงทะเบียนไว้ 2 ทาง |

**หลักปฏิบัติ:**
- **ห้าม** ใช้ข้อมูลบัตรเต็ม (full PAN/CVV) เป็นหลักฐานพิสูจน์ตัวตนเด็ดขาด — ขัด PCI-DSS v4.0
- ผู้ยื่นแทน (ผู้แทนโดยชอบธรรม/ผู้รับมอบอำนาจ) ต้องแนบหนังสือมอบอำนาจ + บัตรของทั้งสองฝ่าย
- หากพิสูจน์ตัวตนไม่ผ่านภายใน **14 วัน** หลังร้องขอเอกสารเพิ่ม → ปิดคำขอชั่วคราวและแจ้งเหตุ (นาฬิกา SLA หยุดระหว่างรอเอกสารจากผู้ยื่น — clock-stop)
- เอกสารพิสูจน์ตัวตนเก็บแยก เข้ารหัส และ**ทำลายภายใน 90 วัน**หลังปิดคำขอ (data minimization)

---

## 5. กรอบเวลาดำเนินการ (Fulfillment SLA)

**มาตรฐาน:** ตอบสนองคำขอโดยไม่ชักช้าและ**ภายใน 30 วัน**นับแต่วันที่พิสูจน์ตัวตนสำเร็จ ขยายได้อีก **ไม่เกิน 30 วัน** สำหรับคำขอที่ซับซ้อน/จำนวนมาก โดยต้องแจ้งเหตุผลเป็นหนังสือก่อนครบกำหนดเดิม

| ประเภทสิทธิ | เป้าหมายภายใน (internal target) | ผลลัพธ์ที่ส่งมอบ |
|-------------|-------------------------------|-------------------|
| เข้าถึง/ขอสำเนา (Access) | 20 วัน | รายงานข้อมูล + ที่มา + วัตถุประสงค์ + ผู้รับข้อมูล + ระยะเก็บรักษา |
| โอนย้าย (Portability) | 20 วัน | ไฟล์ machine-readable (JSON/CSV) |
| แก้ไข (Rectification) | 15 วัน | ยืนยันการแก้ไข + แจ้งผู้รับข้อมูลปลายทางที่เกี่ยวข้อง |
| ลบ/ทำลาย (Erasure) | 25 วัน | ยืนยันการลบ/anonymize + รายการที่คงไว้ตามข้อยกเว้น |
| ระงับ/คัดค้าน (Restriction/Object) | 15 วัน | ยืนยันการระงับ/ยุติการประมวลผล |
| ถอนความยินยอม | 7 วัน | ยืนยันการถอน + ผลกระทบต่อบริการ |

**การส่งมอบอย่างปลอดภัย:** ผลลัพธ์ส่งผ่านช่องทางเข้ารหัส (portal ที่ต้อง authenticate หรือไฟล์เข้ารหัสรหัสผ่านส่งแยกช่องทาง) — **ไม่ส่งข้อมูลส่วนบุคคลเป็น attachment แบบ plaintext** และ export ต้องผ่าน redaction ให้ไม่มี CHD ตาม PCI-DSS v4.0

**ค่าบริการ:** ครั้งแรก**ไม่คิดค่าใช้จ่าย**; คำขอซ้ำ ๆ หรือเกินสมควร (excessive) อาจคิดค่าธรรมเนียมตามต้นทุนจริงที่สมเหตุสมผล หรือปฏิเสธพร้อมเหตุผล

**การหยุดนาฬิกา (clock-stop):** SLA หยุดนับระหว่างรอข้อมูลเพิ่มจากผู้ยื่น และนับต่อเมื่อได้รับครบ

---

## 6. ขั้นตอนปฏิบัติแบบ end-to-end

```
รับคำขอ (intake) ──▶ ลงทะเบียน DSAR Register (request_id) ──▶ acknowledge ผู้ยื่น
     │
     ▼
พิสูจน์ตัวตน (L1/L2/L3) ──ไม่ผ่าน──▶ ขอเอกสารเพิ่ม (clock-stop) ──เกิน 14 วัน──▶ ปิดชั่วคราว
     │ ผ่าน
     ▼
ค้นหา/รวบรวมข้อมูล (payments, merchants, audit_log, logs, backups)
     │
     ▼
ตรวจข้อยกเว้น (AML/ธปท. retention, สิทธิบุคคลที่สาม, PCI) ── Legal/Compliance review
     │
     ▼
DPO อนุมัติ ──▶ ส่งมอบผ่านช่องทางเข้ารหัส ──▶ ปิดคำขอ + บันทึกผลใน Register + audit_log
```

**การค้นหาข้อมูล** ครอบคลุมทุกที่จัดเก็บ: ฐานข้อมูลปฏิบัติการ (`payments`, `merchants`, `refunds`, `webhook_events`, `audit_log`), log/observability store, สำเนาสำรอง (backups/replica), และข้อมูลที่อยู่กับผู้ประมวลผลภายนอก (processors เช่น 3DS/vault vendor) — ต้องแจ้ง processor ให้ดำเนินการตามสัญญา DPA

---

## 7. ข้อยกเว้นและการปฏิเสธ (Exceptions)

การใช้สิทธิบางกรณี **ถูกจำกัดหรือปฏิเสธได้โดยชอบด้วยกฎหมาย** โดยเฉพาะเมื่อชนกับหน้าที่เก็บรักษาข้อมูลของสถาบันการชำระเงิน:

| ข้อยกเว้น | ฐาน | ผลต่อสิทธิ |
|-----------|-----|-----------|
| **หน้าที่เก็บข้อมูลตาม ปปง./AMLO** | พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (KYC/CDD, บันทึกธุรกรรม เก็บ ≥ 5 ปีหลังยุติความสัมพันธ์) | **ปฏิเสธการลบ**ในส่วนนี้ได้ — anonymize ส่วนที่ลบได้เท่านั้น |
| **หน้าที่รายงาน/เก็บหลักฐานตาม ธปท.** | พ.ร.บ. ระบบการชำระเงิน 2560 + ประกาศ ธปท. (audit trail, reconciliation, settlement) | คงไว้ตามระยะที่ ธปท. กำหนด |
| **บันทึกที่ PCI-DSS บังคับ** | PCI-DSS v4.0 Req 10 (audit log ≥ 1 ปี, 3 เดือนพร้อมใช้ทันที) | คงไว้ตามมาตรฐาน |
| **สิทธิของบุคคลที่สาม** | PDPA | redact ส่วนที่กระทบผู้อื่น |
| **การก่อตั้ง/ใช้สิทธิเรียกร้องตามกฎหมาย** | PDPA ม.33 | คงข้อมูลเท่าที่จำเป็น (dispute/chargeback) |
| **คำขอที่ไม่มีมูล/เกินสมควร** | PDPA | คิดค่าธรรมเนียม หรือปฏิเสธพร้อมเหตุผล |

**หลักปฏิบัติเมื่อปฏิเสธ (บางส่วนหรือทั้งหมด):**
- แจ้ง**เป็นหนังสือ** พร้อมเหตุผลและฐานกฎหมายที่ชัดเจน
- แจ้งสิทธิ**ร้องเรียนต่อคณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC)**
- บันทึกเหตุผลการปฏิเสธใน DSAR Register และ `audit_log`
- กรณีลบไม่ได้เพราะ AML/ธปท. → ดำเนินการ **restriction/anonymization** เท่าที่ทำได้แทนการลบทั้งหมด

> **หมายเหตุสำคัญ (callout):** เมื่อ **สิทธิลบข้อมูล (erasure)** ชนกับ **หน้าที่เก็บรักษาข้อมูลตาม ปปง./ธปท./PCI** ให้หน้าที่ตามกฎหมาย/มาตรฐานเหล่านี้ **มีน้ำหนักเหนือกว่า** เฉพาะขอบเขตข้อมูลที่กฎหมายบังคับ ส่วนข้อมูลนอกเหนือหน้าที่บังคับต้องดำเนินการลบตามคำขอ — Legal + Compliance ต้องระบุเป็นลายลักษณ์อักษรว่าข้อมูลใดคงไว้ด้วยฐานใด

---

## 8. เหตุการณ์ละเมิดข้อมูลที่เกี่ยวข้อง

หากระหว่างดำเนินการ DSAR พบการละเมิดข้อมูลส่วนบุคคล ให้เข้าสู่กระบวนการ **แจ้งเหตุต่อ PDPC ภายใน 72 ชั่วโมง** ตาม PDPA และแจ้งเจ้าของข้อมูลหากมีความเสี่ยงสูง พร้อมพิจารณาแจ้ง ธปท. ตามประกาศด้าน cyber resilience (ดูเอกสาร incident response แยกในชุด compliance)

---

## 9. การวัดผลและทบทวน (Metrics & Review)

| ตัวชี้วัด | เป้าหมาย |
|----------|---------|
| % คำขอที่ตอบภายใน SLA | ≥ 98% |
| เวลาเฉลี่ยในการปิดคำขอ | ≤ 20 วัน |
| จำนวนคำขอที่ต้องขยายเวลา | รายงานทุกไตรมาส |
| ข้อร้องเรียนที่ยกระดับถึง PDPC | 0 (เป้าหมาย) |

ทบทวนกระบวนการนี้**อย่างน้อยปีละครั้ง** หรือเมื่อกฎหมาย/ประกาศ PDPC/ธปท. เปลี่ยน โดย DPO เป็นผู้รับผิดชอบ

---

## 10. สรุปสมมติฐาน/สิ่งที่ต้องยืนยันก่อนยื่น (TODO)

- [ ] แต่งตั้ง **DPO** และระบุช่องทางติดต่อ (อีเมล/โทรศัพท์) จริง
- [ ] จัดตั้ง **privacy portal** และอีเมล `dpo@[บริษัท / Company]`
- [ ] จัดทำ **DPA** กับ processors ทุกราย (3DS vendor, vault, cloud) ให้ครอบคลุมหน้าที่สนับสนุน DSAR
- [ ] ยืนยันระยะเก็บรักษาข้อมูลตาม ปปง./ธปท. กับที่ปรึกษากฎหมาย (ตัวเลข ≥ 5 ปีเป็นค่าอ้างอิง ต้องยืนยัน)
- [ ] เชื่อม DSAR Register เข้ากับ `audit_log` และกำหนด retention ของเอกสารพิสูจน์ตัวตน

---
---

# Data Subject Access Request workflow: intake, verification, fulfillment SLA, exceptions (English)

> Document ID `docs/compliance/12-dsar-workflow.md` · Version 0.1 · Dated 22 Jul 2026
> Application type: **Full Acquiring Service license**
> under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT / ธปท.)**
> Applicant: **[บริษัท / Company]** · Paid-up registered capital: **THB 50,000,000**
>
> **Document status:** internal pre-submission draft — must be reviewed by legal counsel / the DPO and a PDPA advisor before it is operational. This is an operating procedure, not legal advice.
>
> Related docs: `COMPLIANCE-TH.md` (law/licensing) · `ARCHITECTURE.md` (architecture/PCI scope) · `03-shareholder-board-fit-proper.md` · `04-org-chart-governance.md` (roles and governance)

---

## 1. Purpose and Scope

This document defines the standard workflow for receiving and handling **Data Subject Access Requests (DSAR)** under Thailand's **Personal Data Protection Act B.E. 2562 (PDPA)**, covering **intake → verification → fulfillment SLA → exceptions**.

**Rights in scope** (PDPA sections 30–36):

| Right | Section (approx.) | Substance |
|-------|-------------------|-----------|
| Access / obtain a copy | s.30 | Access and receive a copy of one's data plus its source |
| Data portability | s.31 | Receive data in a machine-readable format |
| Object | s.32 | Object to collection/use/disclosure |
| Erasure | s.33 | Request deletion, destruction, or anonymization |
| Restriction | s.34 | Request temporary suspension of use |
| Rectification | s.35–36 | Request correction to be accurate, current, complete |
| Withdraw consent | s.19 | Withdraw consent at any time (where no other lawful basis applies) |

**Data scope:** data of (a) cardholders/payers, (b) merchant contacts, (c) website/app visitors, and (d) employees/applicants.

**Out of scope:** full card data (full PAN/CVV/PIN/track) — the core system **does not store** it per the scope-minimization principle in `ARCHITECTURE.md` §6 (only `card_brand` + `card_last4` are retained).

---

## 2. Roles and Responsibilities (RACI)

| Role | Responsibility in the DSAR process |
|------|------------------------------------|
| **DPO** | Process owner; approves responses; adjudicates exceptions; PDPC point of contact |
| **DSAR Coordinator** | Receives, logs, tracks SLA, coordinates teams, drafts response letters |
| **Identity Verification (IAM/Support)** | Verifies the requester's identity |
| **Engineering / Data Team** | Searches, extracts, deletes/rectifies data across systems (payments, merchants, audit_log, logs) |
| **Legal Counsel** | Assesses legal exceptions and conflicts with AML/BOT obligations |
| **Security (DevSecOps)** | Encrypted delivery, access logging, PCI-scope check on exports |
| **Compliance / AMLO Liaison** | Checks conflicts with AMLO/BOT record-retention duties |

**Separation of duties:** for sensitive or high-volume requests, the identity verifier, data extractor, and export approver must not be the same person (consistent with dual control, PCI Req 7).

> **[Assumption / TODO]** A **DPO** has not yet been formally appointed, and named individuals for DSAR Coordinator / Legal Counsel are not yet fixed. Names, titles, contact channels, and appointment letters must be filled in before this is operational and before BOT submission (aligns with `04-org-chart-governance.md`).

---

## 3. Intake Channels

| Channel | Detail | Acknowledgement SLA |
|---------|--------|---------------------|
| Dedicated DPO email | `[TODO: dpo@company.co.th]` | Within 3 business days |
| Online form (privacy portal) | Web/app form + reCAPTCHA | Immediate (auto-issues request ID) |
| Registered mail | Registered office address | Within 5 business days of receipt |
| Via merchant (for cardholders) | Merchant forwards to DPO within 2 business days | SLA counts from DPO receipt |

**Minimum information required:** (1) full name and reply channel; (2) relationship to [บริษัท / Company] (cardholder / merchant / employee); (3) the right invoked; (4) data scope/time range (if known); (5) identity evidence.

**Logging:** every request is recorded in the **DSAR Register** with a `request_id` (format `DSAR-YYYY-NNNN`), receipt date, right type, status, deadline, owner, and outcome — retained as PDPA compliance evidence and producible to PDPC/BOT on request. Every access to the register is written to `audit_log` (append-only per `ARCHITECTURE.md`).

---

## 4. Identity Verification

Verification is **risk-based** — the more sensitive/voluminous the data, the stronger the check — to prevent impersonation, the primary DSAR risk.

| Tier | When | Evidence required |
|------|------|-------------------|
| **L1 — Basic** | Access/rectify ordinary contact data | Prove control of the registered email/phone (OTP) |
| **L2 — Elevated** | Card transaction data, portability, erasure | L1 + national ID copy (redact non-essential digits) + match in-system data such as `card_last4` + transaction date/amount |
| **L3 — Strongest** | High-volume / sensitive data / suspected impersonation | L2 + video KYC or confirmation via two registered channels |

**Rules:**
- **Never** use full card data (full PAN/CVV) as identity evidence — violates PCI-DSS v4.0.
- Requests by a representative (legal guardian / attorney-in-fact) require a power of attorney plus IDs of both parties.
- If verification is not completed within **14 days** of requesting additional documents, the request is provisionally closed with notice (SLA clock stops while awaiting the requester's documents — clock-stop).
- Verification documents are stored separately, encrypted, and **destroyed within 90 days** of closure (data minimization).

---

## 5. Fulfillment SLA

**Standard:** respond without undue delay and **within 30 days** of successful verification, extendable by **no more than 30 additional days** for complex/high-volume requests, with written justification given before the original deadline.

| Right | Internal target | Deliverable |
|-------|-----------------|-------------|
| Access / copy | 20 days | Data report + source + purpose + recipients + retention period |
| Portability | 20 days | Machine-readable file (JSON/CSV) |
| Rectification | 15 days | Confirmation of correction + notice to relevant downstream recipients |
| Erasure | 25 days | Confirmation of deletion/anonymization + list of items retained under exceptions |
| Restriction / Object | 15 days | Confirmation of suspension/cessation of processing |
| Withdraw consent | 7 days | Confirmation of withdrawal + service impact |

**Secure delivery:** results are delivered over an encrypted channel (an authenticated portal, or a password-protected file with the password sent out of band) — **no plaintext personal-data attachments** — and every export is redacted to contain no CHD per PCI-DSS v4.0.

**Fees:** the first request is **free of charge**; repetitive or excessive requests may attract a reasonable cost-based fee or be refused with reasons.

**Clock-stop:** the SLA is paused while awaiting further information from the requester and resumes once received.

---

## 6. End-to-end Procedure

```
Intake ──▶ Log in DSAR Register (request_id) ──▶ Acknowledge requester
     │
     ▼
Verify identity (L1/L2/L3) ──fail──▶ Request more docs (clock-stop) ──>14 days──▶ Provisional close
     │ pass
     ▼
Search / assemble data (payments, merchants, audit_log, logs, backups)
     │
     ▼
Check exceptions (AML/BOT retention, third-party rights, PCI) ── Legal/Compliance review
     │
     ▼
DPO approves ──▶ Deliver over encrypted channel ──▶ Close + record outcome in Register + audit_log
```

**Data search** covers every store: operational databases (`payments`, `merchants`, `refunds`, `webhook_events`, `audit_log`), log/observability stores, backups/replicas, and data held by external processors (e.g. 3DS/vault vendors) — processors must be instructed to act per their DPA.

---

## 7. Exceptions and Refusals

Certain rights may be **lawfully limited or refused**, particularly where they conflict with a payment institution's record-retention duties:

| Exception | Basis | Effect on the right |
|-----------|-------|---------------------|
| **AMLO retention duty** | Anti-Money Laundering Act (KYC/CDD, transaction records kept ≥ 5 years after relationship ends) | **Erasure may be refused** for that data — anonymize only what can be removed |
| **BOT reporting/evidence duty** | Payment Systems Act 2560 + BOT notifications (audit trail, reconciliation, settlement) | Retain per BOT-mandated period |
| **PCI-DSS-mandated records** | PCI-DSS v4.0 Req 10 (audit logs ≥ 1 year, 3 months immediately available) | Retain per standard |
| **Third-party rights** | PDPA | Redact portions affecting others |
| **Establishment/exercise of legal claims** | PDPA s.33 | Retain data as necessary (dispute/chargeback) |
| **Manifestly unfounded/excessive requests** | PDPA | Charge a fee or refuse with reasons |

**Practice on refusal (partial or full):**
- Notify **in writing** with a clear reason and legal basis.
- Inform the requester of the right to **complain to the PDPC** (Personal Data Protection Committee).
- Record the refusal rationale in the DSAR Register and `audit_log`.
- Where erasure is impossible due to AML/BOT duties, apply **restriction/anonymization** to the extent possible instead of full deletion.

> **Important callout:** where the **right to erasure** conflicts with **AMLO/BOT/PCI retention duties**, those legal/standard obligations **prevail** only for the exact data they mandate; data beyond the mandated scope must still be deleted per the request. Legal + Compliance must document, in writing, which data is retained and under which basis.

---

## 8. Related Data Breaches

If a personal-data breach is discovered while handling a DSAR, enter the breach process: **notify the PDPC within 72 hours** per the PDPA, notify affected data subjects where the risk is high, and assess BOT notification under cyber-resilience notifications (see the separate incident-response document in this compliance set).

---

## 9. Metrics and Review

| Metric | Target |
|--------|--------|
| % of requests answered within SLA | ≥ 98% |
| Average time to close a request | ≤ 20 days |
| Requests requiring an extension | Reported quarterly |
| Complaints escalated to PDPC | 0 (target) |

This process is reviewed **at least annually** or whenever the law/PDPC/BOT notifications change; the DPO is accountable.

---

## 10. Assumptions / Pre-submission TODO

- [ ] Formally appoint a **DPO** and publish real contact details (email/phone).
- [ ] Stand up a **privacy portal** and a `dpo@[บริษัท / Company]` mailbox.
- [ ] Execute **DPAs** with all processors (3DS vendor, vault, cloud) covering DSAR support duties.
- [ ] Confirm AMLO/BOT retention periods with legal counsel (the ≥ 5-year figure is indicative and must be confirmed).
- [ ] Wire the DSAR Register into `audit_log` and set the retention period for verification documents.
