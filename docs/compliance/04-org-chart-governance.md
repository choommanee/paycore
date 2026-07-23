# ผังองค์กรและธรรมาภิบาล (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1**
>
> รหัสเอกสาร: `COMP-04` · เวอร์ชัน 0.1 · เจ้าของเอกสาร: Chief Compliance Officer (CCO)
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. ก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงก่อนยื่น ธปท.
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม (เส้นทาง B ตาม ROADMAP) | ยังไม่สรุป | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — ยังไม่คัดเลือก | ยังไม่สรุป | CISO |
> | A4 | **จำนวนทุนจดทะเบียนชำระแล้วจริง ณ วันยื่น** — ต้อง ≥ 50 ล้านบาท และคงไว้ ≥ 75% ตลอดการดำเนินงาน | TODO | CFO |
> | A5 | **รายชื่อกรรมการ/ผู้บริหารจริง + เอกสาร fit & proper** | TODO | Corporate Secretary / CCO |
> | A6 | **External independent audit firm** (co-source internal audit) | ยังไม่สรุป | Audit Committee |
> | A7 | **ASV vendor** สำหรับ quarterly network scan | ยังไม่สรุป | CISO |
>
> ห้ามกรอกตัวเลข/ชื่อสมมติในช่องข้างต้นลงในเอกสารที่ยื่นจริง — ต้องเป็นข้อมูลที่ยืนยันได้เท่านั้น

---

## 1. หลักธรรมาภิบาล (Governance Principles)

[บริษัท / Company] ยึดหลักธรรมาภิบาลตามแนวทางประกาศ ธปท. ว่าด้วยการกำกับดูแลกิจการที่ดี
การบริหารความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk) และ Cyber Resilience โดยมีหลักการดังนี้

1. **แบ่งแยกหน้าที่ (Segregation of Duties)** — แยกหน่วยงานที่ "ทำธุรกิจ (revenue-generating)" ออกจากหน่วยงานที่ "ควบคุม (control)" และหน่วยงานที่ "ตรวจสอบอิสระ (independent assurance)" อย่างชัดเจน
2. **สามแนวป้องกัน (Three Lines of Defense)** — ดูข้อ 4
3. **ความเป็นอิสระของหน่วยตรวจสอบ (Independence)** — Internal Audit และ Compliance รายงานตรงต่อคณะกรรมการ ไม่ขึ้นกับฝ่ายปฏิบัติการ
4. **Fit & Proper** — กรรมการและผู้บริหารระดับสูง (โดยเฉพาะผู้ดำรงตำแหน่งควบคุม) ต้องมีคุณสมบัติเหมาะสมและไม่มีลักษณะต้องห้ามตามเกณฑ์ ธปท.
5. **Accountability** — ทุกความเสี่ยงมี "เจ้าของความเสี่ยง (risk owner)" ที่ระบุตัวได้ ทุก control มีผู้รับผิดชอบ
6. **Escalation & Transparency** — มีเส้นทางรายงานเหตุการณ์/ข้อบกพร่องขึ้นสู่คณะกรรมการภายในกรอบเวลาที่กำหนด

---

## 2. ผังองค์กร (Organization Chart)

```
                        ┌───────────────────────────────┐
                        │   คณะกรรมการบริษัท (Board)      │
                        └───────────────┬───────────────┘
        ┌───────────────────┬───────────┼───────────────────┬────────────────────┐
        ▼                   ▼            ▼                   ▼                    ▼
┌───────────────┐  ┌──────────────┐ ┌──────────┐  ┌──────────────────┐ ┌──────────────────┐
│ Audit         │  │ Risk         │ │ Nomination│ │ IT Steering /    │ │ AML Committee     │
│ Committee     │  │ Committee    │ │ & Remun.  │ │ Cyber Committee  │ │ (ปปง./AMLO)       │
│ (คณะตรวจสอบ)  │  │ (คณะบริหาร   │ │ Committee │ │                  │ │                   │
│               │  │  ความเสี่ยง)  │ │          │  │                  │ │                   │
└───────┬───────┘  └──────┬───────┘ └──────────┘  └──────────────────┘ └──────────────────┘
        │ (สายรายงานอิสระ)        │
        │                         │
        │                 ┌───────▼─────────────────────────────────────────┐
        │                 │            ประธานเจ้าหน้าที่บริหาร (CEO)          │
        │                 └───────┬─────────────────────────────────────────┘
        │      ┌──────────┬───────┼──────────┬──────────────┬───────────────┐
        │      ▼          ▼       ▼          ▼              ▼               ▼
        │  ┌───────┐ ┌────────┐ ┌──────┐ ┌────────┐  ┌──────────┐  ┌──────────────┐
        │  │ CFO   │ │ CTO /  │ │ CISO │ │ CCO    │  │ CRO      │  │ Head of Ops  │
        │  │       │ │ Head   │ │      │ │(Compl. │  │ (Risk)   │  │ / Settlement │
        │  │       │ │ Eng.   │ │      │ │ +MLRO) │  │          │  │              │
        │  └───────┘ └────────┘ └──┬───┘ └───┬────┘  └────┬─────┘  └──────────────┘
        │                          │         │           │
        │                  แนวป้องกันที่ 2 (2nd Line: Risk, Compliance, Security)
        │
        └──────────▶ Internal Audit (Head of Internal Audit / Chief Audit Executive)
                     แนวป้องกันที่ 3 (3rd Line) — รายงานตรงต่อ Audit Committee
                     (administrative reporting ถึง CEO เท่านั้น)
```

> **สายรายงานสำคัญ:**
> - **Chief Audit Executive (CAE)** รายงาน functionally ต่อ **Audit Committee**, administratively ต่อ CEO
> - **CCO / MLRO** รายงาน functionally ต่อ **Board / Audit Committee**, administratively ต่อ CEO — มี direct access ถึงประธานคณะกรรมการ
> - **CRO** รายงานต่อ **Risk Committee** และ CEO
> - **CISO** รายงานต่อ CEO และรายงานเรื่อง cyber ต่อ **IT Steering / Cyber Committee** และ Board

---

## 3. บทบาทและความรับผิดชอบของตำแหน่งควบคุมหลัก (Control Function Roles)

### 3.1 Chief Compliance Officer (CCO) — ประธานเจ้าหน้าที่กำกับการปฏิบัติงาน

| หัวข้อ | รายละเอียด |
|--------|-----------|
| แนวป้องกัน | แนวที่ 2 |
| สายรายงาน | Functional → Board/Audit Committee; Administrative → CEO |
| ความเป็นอิสระ | ไม่รับผิดชอบหน่วยงานที่สร้างรายได้; ไม่มีเป้า KPI ด้านยอดขาย |
| หน้าที่หลัก | จัดทำ/ทบทวน **compliance framework** ให้สอดคล้อง พ.ร.บ. ระบบการชำระเงิน 2560, ประกาศ ธปท., PDPA (พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล 2562) |
| | ติดตามการเปลี่ยนแปลงกฎเกณฑ์ (regulatory change management) และประเมินผลกระทบภายใน **30 วัน** นับจากประกาศมีผล |
| | บริหารความสัมพันธ์กับ ธปท. เป็นจุดติดต่อหลัก (single point of contact) สำหรับการรายงานเป็นงวดและการตรวจ on-site |
| | จัดทำ **compliance risk assessment** ประจำปี และ compliance monitoring plan |
| | ดูแลการรายงานเหตุการณ์สำคัญต่อ ธปท. ภายในกรอบเวลาตามประกาศ (เช่น incident รุนแรงภายใน **24 ชั่วโมง** — ดูข้อ 6) |
| คุณสมบัติ | ประสบการณ์ด้าน compliance/กฎหมายในสถาบันการเงิน/ผู้ให้บริการชำระเงิน ≥ 5 ปี, ผ่าน fit & proper |

### 3.2 MLRO — Money Laundering Reporting Officer (ปปง./AMLO)

> ในระยะแรกอาจรวมกับ CCO ได้ แต่ต้องมีทีม AML แยกและมีแผนแยกตำแหน่งเมื่อปริมาณธุรกรรมโต

| หัวข้อ | รายละเอียด |
|--------|-----------|
| ฐานกฎหมาย | พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน + ประกาศสำนักงาน ปปง. |
| หน้าที่หลัก | ดูแล **KYC/CDD/EDD**, sanction & PEP screening, transaction monitoring |
| | จัดทำและส่ง **STR/SAR** (รายงานธุรกรรมที่มีเหตุอันควรสงสัย) และ **CTR** (ธุรกรรมเงินสด/โอนที่ถึงเกณฑ์) ต่อสำนักงาน ปปง. ตามกรอบเวลาที่กฎหมายกำหนด |
| | เก็บรักษาบันทึก CDD/ธุรกรรมไม่น้อยกว่า **10 ปี** ตามที่กฎหมายกำหนด |
| | ทบทวน risk-based AML program ประจำปี และรายงานต่อ AML Committee/Board |

### 3.3 CISO — Chief Information Security Officer

| หัวข้อ | รายละเอียด |
|--------|-----------|
| แนวป้องกัน | แนวที่ 2 (security governance) — เป็นอิสระจากทีมพัฒนา/ปฏิบัติการ IT (แนวที่ 1) |
| สายรายงาน | CEO + Cyber Committee/Board |
| หน้าที่หลัก | เป็นเจ้าของ **information security policy**, **PCI-DSS v4.0 compliance program**, และ cyber resilience ตามประกาศ ธปท. |
| | บริหาร scope PCI-DSS, network segmentation, tokenization vault, HSM/KMS key management (PCI Req 3), access control & MFA (Req 7-8), logging/SIEM (Req 10) |
| | ดูแล **RoC (Report on Compliance)** ประจำปีร่วมกับ **QSA** (TODO A3) และ **quarterly ASV scan** (TODO A7) + **annual penetration test** |
| | เป็นเจ้าของ **Incident Response Plan** และเป็น incident commander สำหรับ security/data breach |
| | ทบทวนสิทธิ์เข้าถึง (access review) ราย **ไตรมาส**; ทดสอบ 3DS (EMV 3DS / 3-D Secure 2.x) และการควบคุม fraud |
| คุณสมบัติ | ประสบการณ์ security ในระบบชำระเงิน/PCI ≥ 5 ปี, วุฒิ CISSP/CISM หรือเทียบเท่า (แนะนำ) |

### 3.4 CRO — Chief Risk Officer (Head of Risk)

| หัวข้อ | รายละเอียด |
|--------|-----------|
| แนวป้องกัน | แนวที่ 2 |
| สายรายงาน | Risk Committee + CEO |
| หน้าที่หลัก | เป็นเจ้าของ **Enterprise Risk Management (ERM) framework**: operational, credit/settlement, liquidity, fraud/chargeback, IT & third-party/outsourcing risk |
| | กำหนดและติดตาม **risk appetite** และ **key risk indicators (KRIs)** — ดูข้อ 5 |
| | ดูแล **BCP/DR** ให้สอดคล้อง RPO ≤ 5 นาที / RTO ≤ 30 นาที (ตาม ARCHITECTURE §8) และจัด DR drill อย่างน้อย **ปีละ 1 ครั้ง** |
| | ประเมินความเสี่ยง outsourcing/third-party (รวม sponsor bank, QSA, cloud) ตามประกาศ ธปท. ว่าด้วย outsourcing |

### 3.5 Chief Audit Executive (CAE) / Head of Internal Audit

| หัวข้อ | รายละเอียด |
|--------|-----------|
| แนวป้องกัน | แนวที่ 3 |
| สายรายงาน | **Functional → Audit Committee** (แต่งตั้ง/ถอดถอน/ประเมิน/อนุมัติงบผ่าน Audit Committee); Administrative → CEO เท่านั้น |
| ความเป็นอิสระ | **ห้ามมีหน้าที่ปฏิบัติการ (operational duties)** ในหน่วยงานที่ตนตรวจ; ไม่ออกแบบ control ที่ตนต้องตรวจ |
| หน้าที่หลัก | จัดทำ **risk-based annual audit plan** อนุมัติโดย Audit Committee |
| | ตรวจสอบความเพียงพอและประสิทธิผลของ control ในแนวที่ 1 และ 2 (compliance, AML, IT security, PCI, ledger/reconciliation, settlement) |
| | ติดตามการแก้ไขข้อบกพร่อง (audit findings remediation tracking) และรายงานความคืบหน้าต่อ Audit Committee ทุกไตรมาส |
| | อาจใช้รูปแบบ **co-source** กับสำนักงานตรวจสอบภายนอกอิสระ (TODO A6) สำหรับงานเฉพาะทาง (เช่น IT/PCI audit) |

---

## 4. สามแนวป้องกัน (Three Lines of Defense)

| | **แนวที่ 1 (1st Line)** | **แนวที่ 2 (2nd Line)** | **แนวที่ 3 (3rd Line)** |
|---|---|---|---|
| บทบาท | เจ้าของและบริหารความเสี่ยง | กำกับดูแลและควบคุมความเสี่ยง | ให้ความเชื่อมั่นอิสระ |
| หน่วยงาน | Engineering, Operations, Settlement, Merchant Onboarding, Customer Support | Risk (CRO), Compliance (CCO), AML (MLRO), Information Security (CISO) | Internal Audit (CAE) |
| หน้าที่ | ปฏิบัติตาม control ในงานประจำ, ทำ self-assessment, บันทึก `audit_log` ทุก state change | ตั้งกรอบนโยบาย/limit, ติดตาม KRI, ท้าทาย (challenge) แนวที่ 1, รายงานความเสี่ยง | ตรวจสอบอิสระว่าแนวที่ 1 และ 2 ทำงานได้ผล |
| รายงานถึง | สายบริหาร (CEO ผ่าน CTO/Head of Ops) | คณะกรรมการชุดย่อยที่เกี่ยวข้อง | Audit Committee |
| ความเป็นอิสระ | — | อิสระจากหน่วยธุรกิจ | อิสระจากทั้งแนวที่ 1 และ 2 |

**ตัวอย่างการทำงานร่วมกัน (ธุรกรรม chargeback):**
1. **แนวที่ 1** — ทีม Operations ดำเนินการ dispute workflow และบันทึกลง ledger/audit log
2. **แนวที่ 2** — Risk ติดตาม KRI chargeback ratio เทียบ threshold; Compliance ตรวจว่าตรงเกณฑ์ scheme/ธปท.
3. **แนวที่ 3** — Internal Audit สุ่มตรวจตัวอย่าง dispute ย้อนหลังว่ากระบวนการและ control ทำงานถูกต้อง

---

## 5. Risk Appetite และ Key Risk Indicators (ตัวอย่าง threshold)

| KRI | เจ้าของ | ระดับปกติ (Green) | เฝ้าระวัง (Amber) | Escalate (Red) |
|-----|--------|------------------|-------------------|----------------|
| Chargeback ratio | CRO | < 0.5% | 0.5–0.9% | ≥ 0.9% |
| Fraud rate (auth) | CISO/CRO | < 0.1% | 0.1–0.2% | ≥ 0.2% |
| Authorization success rate | Head of Ops | > 95% | 90–95% | < 90% |
| Reconciliation break (T+1 unmatched) | CFO/CRO | 0 รายการ | 1–5 รายการ | > 5 หรือค้าง > 48 ชม. |
| Critical security patch SLA | CISO | ภายใน 14 วัน | 15–30 วัน | > 30 วัน |
| Payment core availability | Head of Ops | ≥ 99.95% | 99.9–99.95% | < 99.9% |
| ทุนที่คงไว้ (% ของทุนชำระแล้ว) | CFO | ≥ 90% | 75–90% | < 75% (ผิดเกณฑ์ ธปท.) |

> เกิน Amber → รายงานต่อ Risk Committee ในการประชุมถัดไป; แตะ Red → escalate ทันทีต่อ CEO และประธานคณะกรรมการชุดที่เกี่ยวข้อง

---

## 6. การรายงานเหตุการณ์และเส้นทาง Escalation (Incident & Escalation)

| ประเภทเหตุการณ์ | ผู้รับผิดชอบ | Escalate ภายใน | รายงานภายนอก |
|-----------------|-------------|----------------|--------------|
| Security/data breach (ข้อมูลบัตร/ส่วนบุคคล) | CISO → CEO/Board | ภายใน 1 ชม. (ภายใน) | ธปท. ตามประกาศ cyber; PDPC ภายใน **72 ชม.** (PDPA); card scheme ตาม PFI process |
| ระบบล่มกระทบบริการ (major outage) | Head of Ops → CRO/CEO | ภายใน 1 ชม. | ธปท. ตามเกณฑ์การรายงานเหตุการณ์สำคัญ |
| ธุรกรรมน่าสงสัย (AML) | MLRO | ทันทีที่พบ | ปปง. (STR/SAR) ตามกรอบเวลากฎหมาย |
| ข้อบกพร่องจาก audit ระดับสูง | CAE → Audit Committee | การประชุมถัดไป / ทันทีถ้า critical | — |

---

## 7. รอบการประชุมและการรายงานเชิงธรรมาภิบาล (Governance Cadence)

| องค์คณะ | ความถี่ | สาระหลัก |
|---------|---------|----------|
| คณะกรรมการบริษัท (Board) | ไตรมาสละครั้ง (อย่างน้อย) | อนุมัตินโยบาย, รับรายงานความเสี่ยง/compliance/audit |
| Audit Committee | ไตรมาสละครั้ง | audit findings, remediation, งบตรวจสอบ |
| Risk Committee | รายเดือน | KRI dashboard, risk appetite, incident |
| Cyber/IT Steering Committee | รายเดือน | PCI status, security posture, patch/vuln |
| AML Committee | ไตรมาสละครั้ง | STR/CTR trend, screening effectiveness |
| Management Risk & Control Forum | รายสัปดาห์ | operational incident, KRI ที่ Amber/Red |

---
---

# Organization chart + governance: roles for compliance officer, CISO, internal audit, risk, three-lines-of-defense (English)

> Supporting document for the **Acquiring Service** license application under the
> **Payment Systems Act B.E. 2560 (2017)**, regulated by the **Bank of Thailand (BOT)**, with a paid-up
> registered capital of **THB 50 million**, alongside **PCI-DSS v4.0 Level 1**.
>
> Document ID: `COMP-04` · Version 0.1 · Owner: Chief Compliance Officer (CCO)
> Related documents: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **Disclaimer:** This document is a structural reference, not legal advice. It must be reviewed by
> qualified BOT-licensing counsel before submission.

---

> [!IMPORTANT]
> **Assumptions / TODO** — real values must be filled in before BOT submission.
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Legal company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **Sponsor / acquiring bank** — not yet signed (Track B per ROADMAP) | Unresolved | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — not yet selected | Unresolved | CISO |
> | A4 | **Actual paid-up capital at submission** — must be ≥ THB 50M and maintained ≥ 75% at all times | TODO | CFO |
> | A5 | **Actual directors/executives + fit & proper documentation** | TODO | Corporate Secretary / CCO |
> | A6 | **External independent audit firm** (co-source internal audit) | Unresolved | Audit Committee |
> | A7 | **ASV vendor** for quarterly network scans | Unresolved | CISO |
>
> Do NOT populate the fields above with fictitious names/numbers in the actual submission — only verifiable data.

---

## 1. Governance Principles

[บริษัท / Company] adopts governance aligned with BOT notifications on good corporate governance, IT
risk management, and cyber resilience, based on the following principles:

1. **Segregation of Duties** — clear separation between revenue-generating (business) functions, control functions, and independent assurance.
2. **Three Lines of Defense** — see §4.
3. **Independence of assurance** — Internal Audit and Compliance report functionally to the Board, not to operations.
4. **Fit & Proper** — directors and senior control-function executives must meet BOT fit-and-proper criteria and carry no disqualifying attributes.
5. **Accountability** — every risk has a named risk owner; every control has a named responsible party.
6. **Escalation & Transparency** — defined escalation paths carry incidents and control deficiencies to the Board within set timeframes.

---

## 2. Organization Chart

```
                        ┌───────────────────────────────┐
                        │        Board of Directors      │
                        └───────────────┬───────────────┘
        ┌───────────────────┬───────────┼───────────────────┬────────────────────┐
        ▼                   ▼            ▼                   ▼                    ▼
┌───────────────┐  ┌──────────────┐ ┌──────────┐  ┌──────────────────┐ ┌──────────────────┐
│ Audit         │  │ Risk         │ │ Nomination│ │ IT Steering /    │ │ AML Committee     │
│ Committee     │  │ Committee    │ │ & Remun.  │ │ Cyber Committee  │ │ (AMLO)            │
└───────┬───────┘  └──────┬───────┘ └──────────┘  └──────────────────┘ └──────────────────┘
        │ (independent reporting line)   │
        │                        ┌───────▼──────────────────────────────────────┐
        │                        │       Chief Executive Officer (CEO)          │
        │                        └───────┬──────────────────────────────────────┘
        │      ┌──────────┬───────┼──────────┬──────────────┬───────────────┐
        │      ▼          ▼       ▼          ▼              ▼               ▼
        │  ┌───────┐ ┌────────┐ ┌──────┐ ┌────────┐  ┌──────────┐  ┌──────────────┐
        │  │ CFO   │ │ CTO /  │ │ CISO │ │ CCO    │  │ CRO      │  │ Head of Ops  │
        │  │       │ │ Head   │ │      │ │(Compl. │  │ (Risk)   │  │ / Settlement │
        │  │       │ │ Eng.   │ │      │ │ +MLRO) │  │          │  │              │
        │  └───────┘ └────────┘ └──┬───┘ └───┬────┘  └────┬─────┘  └──────────────┘
        │                          │         │           │
        │                  2nd Line (Risk, Compliance, Security)
        │
        └──────────▶ Internal Audit (Chief Audit Executive)
                     3rd Line — functional reporting to Audit Committee
                     (administrative reporting to CEO only)
```

> **Key reporting lines:**
> - **Chief Audit Executive (CAE)** — functional to the **Audit Committee**, administrative to the CEO.
> - **CCO / MLRO** — functional to the **Board / Audit Committee**, administrative to the CEO, with direct access to the Board Chair.
> - **CRO** — reports to the **Risk Committee** and CEO.
> - **CISO** — reports to the CEO and, on cyber matters, to the **IT Steering / Cyber Committee** and the Board.

---

## 3. Control Function Roles & Responsibilities

### 3.1 Chief Compliance Officer (CCO)

| Item | Detail |
|------|--------|
| Line of defense | 2nd |
| Reporting | Functional → Board/Audit Committee; Administrative → CEO |
| Independence | No responsibility for revenue-generating units; no sales-based KPIs |
| Core duties | Own and maintain the **compliance framework** aligned with the Payment Systems Act B.E. 2560, BOT notifications, and PDPA (B.E. 2562) |
| | Regulatory change management — assess impact within **30 days** of a notification taking effect |
| | Act as the **single point of contact** with BOT for periodic reporting and on-site inspections |
| | Produce the annual **compliance risk assessment** and monitoring plan |
| | Oversee mandatory external incident reporting to BOT within regulatory timeframes (e.g. severe incident within **24 hours** — see §6) |
| Qualifications | ≥ 5 years compliance/legal experience in FIs/payment providers; passes fit & proper |

### 3.2 MLRO — Money Laundering Reporting Officer (AMLO)

> In the early stage the MLRO may be combined with the CCO, but a dedicated AML team is required, with a plan to split the role as volumes grow.

| Item | Detail |
|------|--------|
| Legal basis | Anti-Money Laundering Act + AMLO Office notifications |
| Core duties | Own **KYC/CDD/EDD**, sanction & PEP screening, transaction monitoring |
| | Prepare and file **STR/SAR** and **CTR** to the AMLO Office within statutory timeframes |
| | Retain CDD/transaction records for at least **10 years** as required by law |
| | Review the risk-based AML program annually; report to the AML Committee/Board |

### 3.3 CISO — Chief Information Security Officer

| Item | Detail |
|------|--------|
| Line of defense | 2nd (security governance) — independent from IT engineering/operations (1st line) |
| Reporting | CEO + Cyber Committee/Board |
| Core duties | Own the **information security policy**, **PCI-DSS v4.0 program**, and cyber resilience per BOT notifications |
| | Manage PCI scope, network segmentation, tokenization vault, HSM/KMS key management (PCI Req 3), access control & MFA (Req 7-8), logging/SIEM (Req 10) |
| | Own the annual **RoC (Report on Compliance)** with the **QSA** (TODO A3), **quarterly ASV scans** (TODO A7), and **annual penetration test** |
| | Own the **Incident Response Plan** and act as incident commander for security/data breaches |
| | Perform **quarterly** access reviews; test EMV 3DS (3-D Secure 2.x) and fraud controls |
| Qualifications | ≥ 5 years security experience in payments/PCI; CISSP/CISM or equivalent (recommended) |

### 3.4 CRO — Chief Risk Officer

| Item | Detail |
|------|--------|
| Line of defense | 2nd |
| Reporting | Risk Committee + CEO |
| Core duties | Own the **Enterprise Risk Management (ERM) framework**: operational, credit/settlement, liquidity, fraud/chargeback, IT and third-party/outsourcing risk |
| | Define and monitor **risk appetite** and **key risk indicators (KRIs)** — see §5 |
| | Own **BCP/DR** aligned to RPO ≤ 5 min / RTO ≤ 30 min (per ARCHITECTURE §8); run a DR drill at least **annually** |
| | Assess outsourcing/third-party risk (sponsor bank, QSA, cloud) per BOT outsourcing notifications |

### 3.5 Chief Audit Executive (CAE) / Head of Internal Audit

| Item | Detail |
|------|--------|
| Line of defense | 3rd |
| Reporting | **Functional → Audit Committee** (appointment/removal/evaluation/budget approval); Administrative → CEO only |
| Independence | **No operational duties** in areas audited; does not design controls it must audit |
| Core duties | Produce the **risk-based annual audit plan** approved by the Audit Committee |
| | Test adequacy and effectiveness of 1st- and 2nd-line controls (compliance, AML, IT security, PCI, ledger/reconciliation, settlement) |
| | Track remediation of findings and report progress to the Audit Committee quarterly |
| | May **co-source** with an external independent audit firm (TODO A6) for specialist work (e.g. IT/PCI audit) |

---

## 4. Three Lines of Defense

| | **1st Line** | **2nd Line** | **3rd Line** |
|---|---|---|---|
| Role | Own and manage risk | Oversee and control risk | Independent assurance |
| Functions | Engineering, Operations, Settlement, Merchant Onboarding, Support | Risk (CRO), Compliance (CCO), AML (MLRO), Security (CISO) | Internal Audit (CAE) |
| Duties | Apply controls in daily work, self-assessment, write `audit_log` on every state change | Set policy/limits, monitor KRIs, challenge the 1st line, report risk | Independently assure that 1st and 2nd lines work |
| Reports to | Management (CEO via CTO/Head of Ops) | Relevant board sub-committees | Audit Committee |
| Independence | — | Independent of business units | Independent of both 1st and 2nd lines |

**Worked example (chargeback):**
1. **1st line** — Operations runs the dispute workflow and records to ledger/audit log.
2. **2nd line** — Risk monitors the chargeback-ratio KRI against threshold; Compliance checks scheme/BOT alignment.
3. **3rd line** — Internal Audit samples historical disputes to confirm the process and controls operated correctly.

---

## 5. Risk Appetite & Key Risk Indicators (example thresholds)

| KRI | Owner | Green | Amber | Red (escalate) |
|-----|-------|-------|-------|----------------|
| Chargeback ratio | CRO | < 0.5% | 0.5–0.9% | ≥ 0.9% |
| Fraud rate (auth) | CISO/CRO | < 0.1% | 0.1–0.2% | ≥ 0.2% |
| Authorization success rate | Head of Ops | > 95% | 90–95% | < 90% |
| Reconciliation breaks (T+1 unmatched) | CFO/CRO | 0 items | 1–5 items | > 5 or aging > 48h |
| Critical security patch SLA | CISO | within 14 days | 15–30 days | > 30 days |
| Payment core availability | Head of Ops | ≥ 99.95% | 99.9–99.95% | < 99.9% |
| Capital maintained (% of paid-up) | CFO | ≥ 90% | 75–90% | < 75% (BOT breach) |

> Amber → report to the Risk Committee at the next meeting; Red → escalate immediately to the CEO and the relevant board sub-committee chair.

---

## 6. Incident Reporting & Escalation

| Incident type | Owner | Internal escalation | External reporting |
|---------------|-------|---------------------|--------------------|
| Security/data breach (cardholder/personal data) | CISO → CEO/Board | within 1h (internal) | BOT per cyber notification; PDPC within **72h** (PDPA); card scheme per PFI process |
| Major service outage | Head of Ops → CRO/CEO | within 1h | BOT per significant-incident reporting criteria |
| Suspicious transaction (AML) | MLRO | immediately on detection | AMLO (STR/SAR) within statutory timeframe |
| High-severity audit finding | CAE → Audit Committee | next meeting / immediately if critical | — |

---

## 7. Governance Cadence

| Body | Frequency | Focus |
|------|-----------|-------|
| Board of Directors | at least quarterly | Approve policy; receive risk/compliance/audit reports |
| Audit Committee | quarterly | Audit findings, remediation, audit budget |
| Risk Committee | monthly | KRI dashboard, risk appetite, incidents |
| Cyber/IT Steering Committee | monthly | PCI status, security posture, patch/vuln |
| AML Committee | quarterly | STR/CTR trends, screening effectiveness |
| Management Risk & Control Forum | weekly | Operational incidents, Amber/Red KRIs |

---

## 8. Cross-reference

| Topic | Document |
|-------|----------|
| License categories, capital, application steps | `../COMPLIANCE-TH.md` |
| System architecture, PCI scope, NFR (RPO/RTO/availability) | `../ARCHITECTURE.md` |
| Phases, timeline, team, cost | `../ROADMAP.md` |
