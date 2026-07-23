# ขั้นตอนการรายงานธุรกรรมที่น่าสงสัย (SAR/STR) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1** และภาระหน้าที่ตาม **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน พ.ศ. 2542** (และที่แก้ไขเพิ่มเติม) กำกับโดย **สำนักงาน ปปง. (AMLO / Thai FIU)**
>
> รหัสเอกสาร: `COMP-07` · เวอร์ชัน 0.1 · เจ้าของเอกสาร: Money Laundering Reporting Officer (MLRO) / Chief Compliance Officer (CCO)
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้าน AML และด้านใบอนุญาต ธปท. ก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงก่อนยื่น ธปท./ปปง.
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **เลขทะเบียนผู้มีหน้าที่รายงานต่อ ปปง.** และการลงทะเบียนระบบ **AERB / e-Filing** ของ ปปง. | ยังไม่สรุป | MLRO |
> | A3 | **รายชื่อ + คำสั่งแต่งตั้ง MLRO และ Deputy MLRO อย่างเป็นทางการ** (แจ้งชื่อต่อ ปปง.) | TODO | CCO / Board |
> | A4 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม (เส้นทาง B ตาม ROADMAP) มีผลต่อ layered reporting และ information sharing | ยังไม่สรุป | CEO / Head of Partnerships |
> | A5 | **ผู้ให้บริการ Sanction/PEP/Adverse-media screening** (เช่น Dow Jones, LexisNexis, ComplyAdvantage) — ยังไม่คัดเลือก | ยังไม่สรุป | MLRO / CISO |
> | A6 | **เกณฑ์มูลค่าธุรกรรมเงินสด/โอน ตามกฎกระทรวง** — ต้องยืนยันตัวเลขล่าสุด ณ วันยื่น (ดูข้อ 3) | TODO | MLRO / ที่ปรึกษากฎหมาย AML |
> | A7 | **แบบรายงานทางการของ ปปง.** (ปปง. 1-01 / 1-02 / 1-03 หรือฉบับล่าสุด) — ยืนยันเวอร์ชันแบบฟอร์มปัจจุบัน | TODO | MLRO |
> | A8 | **ช่องทางยื่นจริง** (ระบบอิเล็กทรอนิกส์ ปปง. vs. เอกสารกระดาษ) และใบรับรองดิจิทัลที่ใช้ | ยังไม่สรุป | MLRO / IT |
>
> ห้ามกรอกตัวเลข/ชื่อ/เลขทะเบียนสมมติในช่องข้างต้นลงในเอกสารที่ยื่นจริง — ต้องเป็นข้อมูลที่ยืนยันได้เท่านั้น
> ตัวเลขเกณฑ์ (threshold) ในเอกสารนี้อ้างอิงกฎหมายที่บังคับใช้ทั่วไป **ต้องตรวจสอบกับกฎกระทรวงฉบับล่าสุดก่อนยื่น**

---

## 1. วัตถุประสงค์และขอบเขต (Purpose & Scope)

เอกสารนี้กำหนด **นโยบายและขั้นตอนปฏิบัติงาน (policy & SOP)** สำหรับการตรวจจับ วิเคราะห์ ยกระดับ และรายงาน **ธุรกรรมที่มีเหตุอันควรสงสัย (Suspicious Transaction Report — STR)** และ **ธุรกรรมที่มีมูลค่าตามเกณฑ์ (Cash/Threshold Transaction Report — CTR/TTR)** ต่อ **สำนักงาน ปปง. (AMLO)** ซึ่งทำหน้าที่เป็น **หน่วยข่าวกรองทางการเงินของประเทศไทย (Thai Financial Intelligence Unit — FIU)**

**ขอบเขตการบังคับใช้:** ใช้กับทุกหน่วยงานของ [บริษัท / Company] ที่เกี่ยวข้องกับการรับสมัครร้านค้า (merchant onboarding), การประมวลผลธุรกรรมบัตร/QR, การกระทบยอดและ settlement, การจ่ายเงินให้ร้านค้า (payout) และการจัดการข้อพิพาท/chargeback รวมถึงพนักงาน ผู้บริหาร กรรมการ และผู้ให้บริการภายนอก (outsourced) ที่เข้าถึงข้อมูลธุรกรรม

**ความสัมพันธ์กับสถาปัตยกรรมระบบ:** สัญญาณตรวจจับดึงจากตาราง `payments`, `ledger_entries`, `refunds`, `webhook_events`, `merchants` และ `audit_log` ตาม `ARCHITECTURE.md` โดยเครื่องมือ Risk/Fraud Engine ทำ velocity/anomaly scoring เป็นชั้นแรก ทั้งนี้ **ห้าม** ระบบ AML เข้าถึงหรือจัดเก็บ PAN/CVV/PIN โดยเด็ดขาด — ใช้ได้เพียง `card_brand`, `card_last4`, token และข้อมูล KYC ของร้านค้า

---

## 2. บทบาทและความรับผิดชอบ (Roles & Responsibilities)

| บทบาท | หน้าที่หลักด้าน AML/STR |
|-------|------------------------|
| **คณะกรรมการบริษัท (Board)** | อนุมัตินโยบาย AML/CFT, รับทราบรายงานสรุปเชิงสถิติ (ไม่รวมรายละเอียดที่เป็นความลับของ STR), จัดสรรทรัพยากร |
| **AML Committee** | กำกับดูแลระบบ AML ในระดับปฏิบัติการ, ทบทวนเกณฑ์ (rules/thresholds), อนุมัติ case ที่มีความเสี่ยงสูง |
| **MLRO (Money Laundering Reporting Officer)** | **ผู้มีอำนาจตัดสินใจสูงสุดในการยื่น/ไม่ยื่น STR ต่อ ปปง.**, ลงนามรายงาน, เป็นจุดติดต่อทางการกับ ปปง., ดูแลการรักษาความลับ (tipping-off) |
| **Deputy MLRO** | ปฏิบัติหน้าที่แทน MLRO เมื่อไม่อยู่/มีผลประโยชน์ทับซ้อน, ทบทวน case ชั้นที่สอง |
| **AML Analyst / Compliance Ops (1st review)** | รับ alert, สืบค้นข้อมูล, จัดทำ case file, เสนอความเห็นเบื้องต้น |
| **Frontline (Merchant Onboarding / Ops)** | ยื่น **Internal Suspicious Activity Report (iSAR)** เมื่อพบพฤติกรรมผิดปกติ, ทำ KYC/CDD/EDD |
| **CISO / IT** | ดูแลระบบ screening, log integrity, ความมั่นคงปลอดภัยของ case file, การส่งข้อมูลแบบเข้ารหัส |
| **Internal Audit (3rd line)** | ตรวจสอบอิสระประสิทธิผลของกระบวนการ STR อย่างน้อยปีละครั้ง |
| **DPO (Data Protection Officer)** | ดูแลความสอดคล้อง PDPA โดยเฉพาะข้อยกเว้นเพื่อการปฏิบัติตามกฎหมาย AML |

> **หลักสำคัญ:** พนักงานทุกคนมี "หน้าที่รายงานภายใน (duty to report internally)" แต่ **มีเพียง MLRO/Deputy MLRO เท่านั้น** ที่มีอำนาจยื่นรายงานออกไปยัง ปปง. เพื่อคุมความลับและความสอดคล้อง

---

## 3. ประเภทรายงานและเกณฑ์มูลค่า (Report Types & Thresholds)

> [!WARNING]
> ตัวเลขเกณฑ์ด้านล่างเป็นค่าอ้างอิงตามกฎกระทรวงที่ใช้บังคับโดยทั่วไป **ต้องยืนยันกับกฎกระทรวง/ประกาศ ปปง. ฉบับล่าสุด** (รายการ A6) ก่อนนำไปตั้งค่าในระบบจริง

| ประเภท | เกณฑ์ (Threshold) | กรอบเวลายื่นต่อ ปปง. | แบบรายงาน (อ้างอิง) |
|--------|-------------------|----------------------|----------------------|
| **STR — ธุรกรรมที่มีเหตุอันควรสงสัย** | **ไม่มีเกณฑ์มูลค่าขั้นต่ำ** — ยื่นเมื่อมี "เหตุอันควรสงสัย" ไม่ว่าจำนวนเงินเท่าใด แม้ธุรกรรมยังไม่สำเร็จ | ยื่น **ภายในกำหนดที่กฎหมายกำหนด** นับแต่วันที่มีเหตุอันควรสงสัย (โดยทั่วไปภายใน 7 วันทำการ — ยืนยัน A6) | ปปง. 1-03 (STR) หรือฉบับล่าสุด (A7) |
| **CTR — ธุรกรรมเงินสด** | ตั้งแต่ **2,000,000 บาท** ขึ้นไป (เงินสด) | ตามกำหนดของ ปปง. (โดยทั่วไปภายใน 15 วันของเดือนถัดไป — ยืนยัน A6) | ปปง. 1-01 หรือฉบับล่าสุด |
| **TTR — ธุรกรรมที่เกี่ยวกับทรัพย์สิน** | ตั้งแต่ **5,000,000 บาท** ขึ้นไป | ตามกำหนดของ ปปง. | ปปง. 1-02 หรือฉบับล่าสุด |
| **Asset-freeze / Designated persons (CFT/PF)** | รายชื่อตามประกาศ (UNSCR / กฎหมายไทย) — **ไม่มีเกณฑ์มูลค่า** | **ทันที (immediate)** เมื่อพบว่าตรงรายชื่อ — ระงับและรายงาน ปปง. โดยไม่ชักช้า | ตามแนวทาง ปปง. ด้าน CFT/PF |

**บริบทของ Acquiring Gateway:** โดยธุรกิจแล้วเราประมวลผล **การชำระด้วยบัตร/อิเล็กทรอนิกส์เป็นหลัก ไม่ใช่เงินสด** ดังนั้น **STR คือช่องทางรายงานหลัก** ส่วน CTR/TTR จะเกี่ยวข้องเมื่อมี cash-based flow หรือ payout ขนาดใหญ่เข้าเงื่อนไข ทั้งนี้ต้อง aggregate ธุรกรรมที่ดูเหมือนถูกแบ่งย่อย (structuring/smurfing) เพื่อเลี่ยงเกณฑ์ด้วย

### ตัวอย่างสัญญาณเตือน (Red Flags) เฉพาะธุรกิจ Acquiring

- **Transaction laundering / factoring** — ร้านค้ารูด/รับชำระแทนธุรกิจอื่นที่ไม่ได้ลงทะเบียน (undisclosed MCC mismatch)
- **Structuring** — แบ่งยอดหลายรายการต่ำกว่าเกณฑ์ หรือต่ำกว่า limit ของ velocity rule อย่างจงใจ
- **Bust-out fraud** — ยอดพุ่งผิดปกติหลัง onboarding แล้วขอ payout เร่งด่วนแล้วเงียบหาย
- **Refund/chargeback abuse** — refund ไปยังบัตร/บัญชีที่ไม่ใช่ต้นทาง, refund สูงผิดสัดส่วนกับ capture
- **High-risk MCC / จำหน่ายสินค้าต้องห้าม** — การพนัน crypto ไม่ได้รับอนุญาต สินค้าผิดกฎหมาย
- **Sanction/PEP hit** — beneficial owner หรือ settlement account ตรงรายชื่อ sanction/PEP/adverse media
- **Geographic anomaly** — IP/BIN/ที่อยู่ settlement ขัดแย้งกันหรือมาจากเขตความเสี่ยงสูง (FATF)

---

## 4. ขั้นตอน (Workflow) — จากการตรวจจับสู่การยื่นรายงาน

```
[ตรวจจับ] ──▶ [รายงานภายใน iSAR] ──▶ [ประเมินชั้น 1] ──▶ [สืบสวน/EDD]
   │ automated (Risk Engine)              (AML Analyst)      (case file)
   │ + manual (Frontline)                                        │
   └────────────────────────────────────────────────────────────┘
                                                                 ▼
                                                 [ตัดสินใจโดย MLRO]
                                             ┌───────────┴───────────┐
                                             ▼                       ▼
                                     [ไม่ยื่น → บันทึกเหตุผล]   [ยื่น STR ต่อ ปปง.]
                                             │                       │
                                             ▼                       ▼
                                      [เก็บ ≥ 5–10 ปี]        [ระงับ/ติดตาม + เก็บหลักฐาน]
```

**ขั้นที่ 1 — ตรวจจับ (Detection).** ผสมสองแหล่ง (ก) อัตโนมัติ: Risk/Fraud Engine ยิง alert จาก velocity, anomaly score, sanction/PEP screening ที่ทำ ณ onboarding และแบบ real-time/batch (ข) manual: พนักงานหน้างานพบพฤติกรรมผิดปกติ

**ขั้นที่ 2 — รายงานภายใน (Internal SAR / iSAR).** ผู้พบเหตุยื่น iSAR ผ่านระบบ case management ภายในทันที (เป้าหมาย ≤ 24 ชม.) โดย **ห้ามแจ้งลูกค้า/ร้านค้า** (ดูข้อ 5 — tipping-off) การยิง iSAR ทุกครั้งบันทึกลง `audit_log`

**ขั้นที่ 3 — ประเมินชั้นที่ 1 (Triage).** AML Analyst รวบรวมข้อมูล KYC, ประวัติธุรกรรม, ledger, ผล screening แล้วจัดทำ **case file** พร้อมประเมินว่ามี "เหตุอันควรสงสัย" หรือไม่ (SLA: ปิด triage ภายใน 3 วันทำการ)

**ขั้นที่ 4 — สืบสวน/EDD (Investigation).** กรณีน่าสงสัยจริงทำ Enhanced Due Diligence: ตรวจ beneficial ownership, source of funds, ความสมเหตุสมผลของธุรกิจ, จัดทำ transaction timeline

**ขั้นที่ 5 — ตัดสินใจโดย MLRO (Decision).** MLRO/Deputy MLRO ทบทวน case แล้วตัดสิน "ยื่น" หรือ "ไม่ยื่น" — **ทั้งสองกรณีต้องบันทึกเหตุผลเป็นลายลักษณ์อักษร** พร้อมลงนามและเวลา

**ขั้นที่ 6 — ยื่นรายงานต่อ ปปง. (Filing).** หากยื่น STR: กรอกแบบ ปปง. 1-03 (A7) ส่งผ่านช่องทางอิเล็กทรอนิกส์ที่ ปปง. กำหนด (A8) **ภายในกรอบเวลาตามกฎหมาย** (ข้อ 3) เก็บหลักฐานการยื่น (acknowledgement) ไว้ในระบบ

**ขั้นที่ 7 — ดำเนินการต่อเนื่อง (Post-filing).** ประเมินความสัมพันธ์ทางธุรกิจ (ระงับ/เลิกร้านค้า/hold payout ตามความเหมาะสมและตามคำสั่ง ปปง.), เฝ้าติดตาม (ongoing monitoring), ตอบคำขอข้อมูลเพิ่มเติมจาก ปปง.

### ตาราง SLA / ผู้รับผิดชอบ

| ขั้น | ผู้รับผิดชอบ | เป้าหมายเวลา |
|------|-------------|--------------|
| ยื่น iSAR หลังพบเหตุ | ผู้พบเหตุ | ≤ 24 ชั่วโมง |
| Triage ชั้น 1 | AML Analyst | ≤ 3 วันทำการ |
| EDD / สืบสวน | AML Analyst + MLRO | ≤ 7 วันทำการ |
| ตัดสินใจยื่น/ไม่ยื่น | MLRO | ก่อนครบกำหนดตามกฎหมาย |
| ยื่น STR ต่อ ปปง. | MLRO | ภายในกำหนดกฎหมาย (ข้อ 3) |
| Sanction/asset-freeze hit | MLRO (escalate ทันที) | **ทันที (same day)** |

---

## 5. การรักษาความลับและการห้ามเปิดเผย (Confidentiality & Tipping-off)

1. **ห้ามเปิดเผย (No tipping-off).** ห้ามพนักงาน กรรมการ หรือผู้ใดแจ้ง/บอกใบ้แก่ลูกค้า ร้านค้า หรือบุคคลภายนอกว่ามีหรือกำลังจะมีการยื่น STR หรือกำลังถูกสืบสวน การฝ่าฝืนเป็นความผิดตามกฎหมายและวินัยร้ายแรง
2. **Need-to-know.** เข้าถึง case file / STR ได้เฉพาะ MLRO, Deputy MLRO, AML Analyst ที่รับผิดชอบ และผู้ที่ได้รับอนุญาตเป็นการเฉพาะเท่านั้น — บังคับด้วย RBAC และ MFA (สอดคล้อง PCI-DSS v4.0 Req 7–8)
3. **การคุ้มครองผู้รายงาน (Safe harbor).** พนักงานที่รายงานโดยสุจริตได้รับความคุ้มครองตามกฎหมาย AML และนโยบายบริษัท ห้ามตอบโต้ (no retaliation)
4. **PDPA vs. AML.** การประมวลผลข้อมูลส่วนบุคคลเพื่อยื่น STR/ทำ CDD อาศัย **ฐานการปฏิบัติตามกฎหมาย** ตาม PDPA จึง **ไม่ต้องขอความยินยอม** และ **ไม่แจ้ง data subject** เมื่อการแจ้งจะกระทบการสืบสวน (ข้อยกเว้นสิทธิ) โดยประสานงานกับ DPO และ PDPC ตามความจำเป็น
5. **ความปลอดภัยของข้อมูล.** จัดเก็บ case file แบบเข้ารหัส, การส่งไป ปปง. ใช้ช่องทาง/ใบรับรองที่กำหนด, ทุกการเข้าถึงบันทึกใน audit log ที่แก้ไขไม่ได้ (append-only)

---

## 6. การเก็บรักษาเอกสาร การฝึกอบรม และการตรวจสอบ (Recordkeeping, Training, Assurance)

- **เก็บรักษา (Retention).** เก็บ case file, iSAR, บันทึกการตัดสินใจ (ทั้งยื่นและไม่ยื่น), หลักฐานการยื่น และเอกสาร CDD/EDD เป็นเวลา **อย่างน้อย 5 ปี** (แนะนำ 10 ปีเพื่อรองรับการสอบสวน) นับแต่ปิด case หรือสิ้นสุดความสัมพันธ์ทางธุรกิจ ตามที่กฎหมาย AML กำหนด
- **ฝึกอบรม (Training).** พนักงานที่เกี่ยวข้องทุกคนอบรม AML/CFT + การรับรู้ tipping-off อย่างน้อย **ปีละครั้ง** และ onboarding สำหรับพนักงานใหม่ พร้อมบันทึกผล
- **ทบทวนเกณฑ์ (Rule tuning).** AML Committee ทบทวนกฎ/threshold ของ Risk Engine อย่างน้อยทุกไตรมาส เพื่อลด false positive/negative
- **การตรวจสอบอิสระ (Independent testing).** Internal Audit (แนวป้องกันที่ 3) ตรวจประสิทธิผลของกระบวนการ STR อย่างน้อยปีละครั้ง และรายงานต่อ Audit Committee
- **การรายงานต่อผู้กำกับ.** สรุปสถิติ (จำนวน alert/iSAR/STR ที่ยื่น) รายงานต่อ AML Committee และ Board เป็นงวด **โดยไม่เปิดเผยรายละเอียดที่เป็นความลับของ STR**

---

# Suspicious Activity/Transaction Reporting procedure to the Thai FIU/AMLO, thresholds, workflow, confidentiality (English)

> Supporting document for the license application for **Electronic Payment Acquiring Service**
> under the **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT)**, paid-up registered capital **THB 50 million**,
> alongside **PCI-DSS v4.0 Level 1**, and obligations under the **Anti-Money Laundering Act B.E. 2542** (as amended), regulated by the **Anti-Money Laundering Office (AMLO — the Thai Financial Intelligence Unit / FIU)**.
>
> Document code: `COMP-07` · Version 0.1 · Owner: Money Laundering Reporting Officer (MLRO) / Chief Compliance Officer (CCO)
> Related documents: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`
>
> **Disclaimer:** This document is a structural reference, not legal advice. It must be reviewed by AML and BOT-licensing legal counsel before actual submission.

---

> [!IMPORTANT]
> **Assumptions / TODO** — real values must be filled in before submission to BOT/AMLO.
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Actual company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **AMLO reporting-entity registration number** and enrolment in AMLO's **electronic filing (e-Filing / AERB)** system | Open | MLRO |
> | A3 | **Formal appointment letters for MLRO and Deputy MLRO** (names notified to AMLO) | TODO | CCO / Board |
> | A4 | **Sponsor / Acquiring bank** — not yet signed (Path B per ROADMAP); affects layered reporting & information sharing | Open | CEO / Head of Partnerships |
> | A5 | **Sanction/PEP/adverse-media screening vendor** (e.g., Dow Jones, LexisNexis, ComplyAdvantage) — not yet selected | Open | MLRO / CISO |
> | A6 | **Cash/asset transaction thresholds per ministerial regulation** — confirm latest figures as of filing date (see §3) | TODO | MLRO / AML counsel |
> | A7 | **Official AMLO report forms** (Por Por Ngor 1-01 / 1-02 / 1-03 or latest) — confirm current form version | TODO | MLRO |
> | A8 | **Actual filing channel** (AMLO electronic system vs. paper) and the digital certificate used | Open | MLRO / IT |
>
> Do not enter fabricated numbers/names/registration IDs into the actual submission — only verifiable data is permitted.
> Threshold figures in this document reference generally-applicable law and **must be verified against the latest ministerial regulation before filing.**

---

## 1. Purpose & Scope

This document defines the **policy and standard operating procedure (SOP)** for detecting, analysing, escalating, and reporting **Suspicious Transaction Reports (STRs)** and **Cash/Threshold Transaction Reports (CTR/TTR)** to the **Anti-Money Laundering Office (AMLO)**, which acts as **Thailand's Financial Intelligence Unit (FIU)**.

**Applicability:** All units of [บริษัท / Company] involved in merchant onboarding, card/QR transaction processing, reconciliation and settlement, merchant payout, and dispute/chargeback handling — including employees, executives, directors, and outsourced service providers with access to transaction data.

**Relationship to system architecture:** Detection signals are drawn from the `payments`, `ledger_entries`, `refunds`, `webhook_events`, `merchants`, and `audit_log` tables per `ARCHITECTURE.md`, with the Risk/Fraud Engine providing first-layer velocity/anomaly scoring. Under no circumstances may the AML system access or store PAN/CVV/PIN — only `card_brand`, `card_last4`, tokens, and merchant KYC data may be used.

---

## 2. Roles & Responsibilities

| Role | Primary AML/STR responsibility |
|------|--------------------------------|
| **Board of Directors** | Approve AML/CFT policy; receive aggregate statistical reporting (excluding confidential STR details); allocate resources |
| **AML Committee** | Operational oversight of the AML program; review rules/thresholds; approve high-risk cases |
| **MLRO (Money Laundering Reporting Officer)** | **Final authority to file/not file an STR with AMLO**; signs reports; official AMLO point of contact; guards confidentiality (tipping-off) |
| **Deputy MLRO** | Acts for the MLRO on absence/conflict of interest; second-line case review |
| **AML Analyst / Compliance Ops (1st review)** | Receive alerts, research, build the case file, provide preliminary opinion |
| **Frontline (Merchant Onboarding / Ops)** | File an **Internal Suspicious Activity Report (iSAR)** on abnormal behaviour; perform KYC/CDD/EDD |
| **CISO / IT** | Maintain screening systems, log integrity, case-file security, encrypted transmission |
| **Internal Audit (3rd line)** | Independently test STR-process effectiveness at least annually |
| **DPO (Data Protection Officer)** | Ensure PDPA alignment, especially the legal-obligation exemption for AML |

> **Key principle:** Every employee has a *duty to report internally*, but **only the MLRO/Deputy MLRO** may file a report externally to AMLO — to preserve confidentiality and consistency.

---

## 3. Report Types & Thresholds

> [!WARNING]
> The threshold figures below reflect generally-applicable ministerial regulations and **must be confirmed against the latest AMLO regulation/notification** (item A6) before configuring production systems.

| Type | Threshold | Filing deadline to AMLO | Report form (ref.) |
|------|-----------|-------------------------|--------------------|
| **STR — Suspicious Transaction Report** | **No minimum monetary threshold** — file whenever "reasonable grounds to suspect" exist, regardless of amount, even if the transaction was not completed | File **within the statutory deadline** from the date grounds arose (commonly within 7 business days — confirm A6) | Por Por Ngor 1-03 (STR) or latest (A7) |
| **CTR — Cash Transaction Report** | **THB 2,000,000** or more (cash) | Per AMLO schedule (commonly by the 15th of the following month — confirm A6) | Por Por Ngor 1-01 or latest |
| **TTR — Asset-related Transaction Report** | **THB 5,000,000** or more | Per AMLO schedule | Por Por Ngor 1-02 or latest |
| **Asset-freeze / Designated persons (CFT/PF)** | Listed names (UNSCR / Thai law) — **no monetary threshold** | **Immediate** upon a name match — freeze and report to AMLO without delay | Per AMLO CFT/PF guidance |

**Acquiring-gateway context:** Our business processes predominantly **card/electronic payments, not cash**, so the **STR is the primary reporting channel**. CTR/TTR become relevant where cash-based flows or large payouts meet the thresholds. Transactions that appear deliberately split (structuring/smurfing) to evade thresholds must be aggregated.

### Business-specific red flags (Acquiring)

- **Transaction laundering / factoring** — a merchant processing on behalf of an undisclosed business (MCC mismatch)
- **Structuring** — deliberately splitting amounts below a threshold or below velocity-rule limits
- **Bust-out fraud** — abnormal volume spike shortly after onboarding, urgent payout request, then disappearance
- **Refund/chargeback abuse** — refunds to a card/account other than the original, refunds disproportionate to captures
- **High-risk MCC / prohibited goods** — unlicensed gambling/crypto, illegal goods
- **Sanction/PEP hit** — beneficial owner or settlement account matching sanction/PEP/adverse-media lists
- **Geographic anomaly** — conflicting IP/BIN/settlement-address data or origins in FATF high-risk jurisdictions

---

## 4. Workflow — from detection to filing

```
[Detect] ──▶ [Internal iSAR] ──▶ [Tier-1 triage] ──▶ [Investigate / EDD]
   │ automated (Risk Engine)         (AML Analyst)        (case file)
   │ + manual (Frontline)                                     │
   └──────────────────────────────────────────────────────────┘
                                                               ▼
                                                   [MLRO decision]
                                          ┌───────────┴───────────┐
                                          ▼                       ▼
                                 [No file → log rationale]  [File STR to AMLO]
                                          │                       │
                                          ▼                       ▼
                                  [Retain ≥ 5–10 yrs]     [Restrict/monitor + retain evidence]
```

**Step 1 — Detection.** Two sources: (a) automated — the Risk/Fraud Engine raises alerts from velocity, anomaly scores, and sanction/PEP screening performed at onboarding and on a real-time/batch basis; (b) manual — frontline staff observing abnormal behaviour.

**Step 2 — Internal report (iSAR).** The observer files an iSAR in the internal case-management system immediately (target ≤ 24h). The customer/merchant **must not be informed** (see §5 — tipping-off). Every iSAR is recorded in `audit_log`.

**Step 3 — Tier-1 triage.** The AML Analyst gathers KYC, transaction history, ledger data, and screening results, builds a **case file**, and assesses whether "reasonable grounds to suspect" exist (SLA: close triage within 3 business days).

**Step 4 — Investigation / EDD.** Genuinely suspicious cases undergo Enhanced Due Diligence: beneficial-ownership review, source of funds, business rationale, and a transaction timeline.

**Step 5 — MLRO decision.** The MLRO/Deputy MLRO reviews the case and decides to file or not file — **both outcomes must be documented in writing**, signed and timestamped.

**Step 6 — File to AMLO.** If filing an STR: complete form Por Por Ngor 1-03 (A7), submit via the AMLO electronic channel (A8) **within the statutory deadline** (§3), and retain the acknowledgement in the system.

**Step 7 — Post-filing.** Reassess the business relationship (restrict/offboard the merchant / hold payout as appropriate and as directed by AMLO), continue ongoing monitoring, and respond to AMLO requests for further information.

### SLA / ownership table

| Step | Owner | Target time |
|------|-------|-------------|
| File iSAR after observation | Observer | ≤ 24 hours |
| Tier-1 triage | AML Analyst | ≤ 3 business days |
| EDD / investigation | AML Analyst + MLRO | ≤ 7 business days |
| File / no-file decision | MLRO | Before statutory deadline |
| File STR to AMLO | MLRO | Within statutory deadline (§3) |
| Sanction / asset-freeze hit | MLRO (immediate escalation) | **Immediate (same day)** |

---

## 5. Confidentiality & Tipping-off

1. **No tipping-off.** No employee, director, or any person may inform or hint to a customer, merchant, or third party that an STR has been or will be filed, or that an investigation is under way. Breach is a legal offence and a serious disciplinary matter.
2. **Need-to-know.** Access to case files / STRs is restricted to the MLRO, Deputy MLRO, the assigned AML Analyst, and specifically authorised persons — enforced via RBAC and MFA (consistent with PCI-DSS v4.0 Req 7–8).
3. **Safe harbour.** Employees reporting in good faith are protected under AML law and company policy; no retaliation is permitted.
4. **PDPA vs. AML.** Processing personal data to file STRs / perform CDD relies on the **legal-obligation basis** under the PDPA, so **no consent is required** and the data subject is **not notified** where notification would prejudice the investigation (exercise-of-rights exception), coordinated with the DPO and PDPC (Personal Data Protection Committee) as needed.
5. **Data security.** Case files are stored encrypted; transmission to AMLO uses the prescribed channel/certificate; every access is recorded in an immutable (append-only) audit log.

---

## 6. Recordkeeping, Training & Assurance

- **Retention.** Retain case files, iSARs, decision records (both filed and not filed), filing acknowledgements, and CDD/EDD documents for **at least 5 years** (10 years recommended to support investigations) from case closure or end of the business relationship, as required by AML law.
- **Training.** All relevant staff receive AML/CFT and tipping-off awareness training at least **annually**, plus onboarding for new hires, with records kept.
- **Rule tuning.** The AML Committee reviews Risk-Engine rules/thresholds at least quarterly to reduce false positives/negatives.
- **Independent testing.** Internal Audit (third line of defence) tests STR-process effectiveness at least annually and reports to the Audit Committee.
- **Regulatory reporting.** Aggregate statistics (numbers of alerts/iSARs/STRs filed) are reported periodically to the AML Committee and the Board **without disclosing confidential STR details**.
