# กฎบัตรและแผนการตรวจสอบภายใน (ไทย)

> เอกสาร: `31-internal-audit-charter.md` · เวอร์ชัน 0.1 · เจ้าของเอกสาร: คณะกรรมการตรวจสอบ (Audit Committee)
> ประเภทกิจการ: **Full Acquiring Payment Gateway** · ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> อ้างอิงกรอบกฎหมาย: พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 (กำกับโดย **ธปท.**), พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (**ปปง./AMLO**), พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (**PDPC/PDPA**), มาตรฐาน **PCI-DSS v4.0**, **EMV 3-D Secure (3DS)**
> เอกสารประกอบการยื่นขอใบอนุญาตต่อ ธปท. (ดู `../COMPLIANCE-TH.md` ข้อ 5 — นโยบายบริหารความเสี่ยง/การกำกับดูแล)

> **ข้อสงวน:** เอกสารนี้เป็นแม่แบบเชิงเทคนิค/การกำกับ ไม่ใช่คำแนะนำทางกฎหมาย ควรให้ที่ปรึกษากฎหมายและ QSA ตรวจทานก่อนยื่นจริง

---

## ⚠️ สมมติฐานและรายการที่ยังไม่ยืนยัน (Assumptions / TODO)

> รายการต่อไปนี้ยังขึ้นกับ external dependency ที่ยังไม่ปิด — **ห้ามถือเป็นข้อเท็จจริง** จนกว่าจะยืนยัน

| # | รายการ | สถานะ | ผู้รับผิดชอบปิดประเด็น |
|---|--------|-------|------------------------|
| A1 | **ชื่อบริษัทและเลขทะเบียนนิติบุคคล** — ใช้ placeholder "[บริษัท / Company]" | TODO | เลขานุการบริษัท |
| A2 | **Sponsor bank / acquiring partner** — ยังไม่ลงนามสัญญา ขอบเขตงาน settlement/chargeback ที่ต้องตรวจจึงยังไม่ล็อก | TODO | ประธานเจ้าหน้าที่บริหาร (CEO) |
| A3 | **QSA vendor (PCI-DSS v4.0)** — ยังไม่เลือกผู้ประเมิน RoC/ASV; วันประเมินจริงยังไม่ยืนยัน | TODO | CISO / Compliance |
| A4 | **ทุนจดทะเบียนชำระแล้วที่แท้จริง** — เอกสารสมมติ 50 ล้านบาทตามเกณฑ์ acquiring; ต้องแนบหลักฐานชำระจริงและคง ≥ 75% | TODO | CFO |
| A5 | **ผู้ตรวจสอบภายในอิสระ (co-source firm)** — ปีแรกคาดจ้าง co-source เพราะทีมในยังไม่พร้อม; ยังไม่เลือกสำนักงาน | TODO | Audit Committee |
| A6 | **DPO (Data Protection Officer)** ตาม PDPA — ยังไม่แต่งตั้งอย่างเป็นทางการ | TODO | CEO / Legal |

---

## 1. บทนำและวัตถุประสงค์

กฎบัตรฉบับนี้กำหนดพันธกิจ อำนาจหน้าที่ ความรับผิดชอบ และความเป็นอิสระของ **ฝ่ายตรวจสอบภายใน (Internal Audit — IA)** ของ **[บริษัท / Company]** ซึ่งประกอบธุรกิจให้บริการรับชำระเงินด้วยวิธีทางอิเล็กทรอนิกส์ (Acquiring) ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560

วัตถุประสงค์:
- ให้ความเชื่อมั่นอย่างเป็นอิสระ (independent assurance) ต่อคณะกรรมการบริษัทและคณะกรรมการตรวจสอบเกี่ยวกับความเพียงพอและประสิทธิผลของระบบควบคุมภายใน การบริหารความเสี่ยง และการกำกับดูแล
- สนับสนุนการปฏิบัติตามข้อกำหนดของ ธปท. ด้าน IT risk, cyber resilience, business continuity และ outsourcing
- ตรวจสอบการปฏิบัติตาม PCI-DSS v4.0, กฎหมาย AML/CFT ของ ปปง., และ PDPA

IA ยึดกรอบมาตรฐานสากล **IIA Global Internal Audit Standards (2024)** และแนวปฏิบัติของ ธปท. ว่าด้วยการควบคุมภายในและการตรวจสอบภายในของสถาบันภายใต้การกำกับ

---

## 2. โครงสร้างการกำกับ — Three Lines Model

| แนวป้องกัน | ผู้รับผิดชอบ | บทบาท |
|-----------|-------------|-------|
| **แนวที่ 1** | หน่วยปฏิบัติการ (Payment Ops, Engineering, Merchant Onboarding) | เป็นเจ้าของความเสี่ยงและควบคุมในกระบวนการ |
| **แนวที่ 2** | Compliance, Risk Management, CISO/Security, AML Officer, DPO | กำหนดนโยบาย ติดตาม และท้าทายแนวที่ 1 |
| **แนวที่ 3** | **ฝ่ายตรวจสอบภายใน (IA)** | ให้ความเชื่อมั่นอิสระต่อแนวที่ 1 และ 2 รายงานตรงต่อคณะกรรมการตรวจสอบ |

**สายรายงาน (Reporting Lines):**
- **สายงาน (functional):** หัวหน้าฝ่ายตรวจสอบภายใน (Chief Audit Executive — CAE) รายงานตรงต่อ **คณะกรรมการตรวจสอบ (Audit Committee)**
- **สายบริหาร (administrative):** รายงานต่อ CEO เฉพาะเรื่องธุรการ (งบประมาณ ทรัพยากร) เท่านั้น เพื่อรักษาความเป็นอิสระ
- คณะกรรมการตรวจสอบเป็นผู้อนุมัติการแต่งตั้ง/ถอดถอน ประเมินผลงาน และกำหนดค่าตอบแทนของ CAE

---

## 3. ความเป็นอิสระและความเที่ยงธรรม (Independence & Objectivity)

- IA ไม่มีอำนาจหรือความรับผิดชอบในการปฏิบัติงานที่ตนตรวจสอบ (no operational responsibility)
- ผู้ตรวจสอบต้องไม่ตรวจงานที่ตนเคยรับผิดชอบภายใน **12 เดือน** ที่ผ่านมา
- CAE ต้องยืนยันความเป็นอิสระขององค์กรต่อคณะกรรมการตรวจสอบเป็นลายลักษณ์อักษร **อย่างน้อยปีละ 1 ครั้ง**
- หากเกิดความขัดแย้งทางผลประโยชน์ (conflict of interest) ต้องเปิดเผยต่อคณะกรรมการตรวจสอบทันที และมอบหมายผู้ตรวจสอบรายอื่นหรือใช้ co-source (ดู TODO A5)

---

## 4. อำนาจหน้าที่ (Authority)

IA ได้รับอนุมัติจากคณะกรรมการให้มีอำนาจ:
1. เข้าถึงข้อมูล ระบบ บันทึก ทรัพย์สิน และบุคลากรทุกส่วนที่จำเป็นต่อการตรวจสอบ **โดยไม่มีข้อจำกัด** รวมถึงระบบใน PCI scope (vault/tokenization), `ledger_entries`, `audit_log`, และ SIEM
2. เข้าถึงเอกสารการประชุมคณะกรรมการและคณะกรรมการชุดย่อย
3. ขอคำชี้แจงและความร่วมมือจากพนักงานทุกระดับ
4. จัดหาผู้เชี่ยวชาญภายนอก (เช่น pentester, forensic) เมื่อจำเป็น

**ขอบเขต:** IA ครอบคลุมทุกหน่วยงาน กระบวนการ และผู้ให้บริการภายนอก (outsourcing/critical vendor รวมถึง QSA, sponsor bank interface, cloud, HSM/KMS) ตามสิทธิ์ audit-right ในสัญญา

---

## 5. ความรับผิดชอบ (Responsibilities)

- จัดทำ **แผนตรวจสอบประจำปีแบบ risk-based** และแผนหมุนเวียน 3 ปี เสนอคณะกรรมการตรวจสอบอนุมัติ
- ดำเนินการตรวจสอบตามแผน ประเมินความเพียงพอของการควบคุมด้าน payment ops, security, AML, PCI, PDPA
- ติดตามการแก้ไขข้อสังเกต (finding tracking) จนปิดประเด็น
- รายงานผลต่อคณะกรรมการตรวจสอบทุกไตรมาส
- ประสานงานกับผู้สอบบัญชีภายนอกและ QSA เพื่อลดความซ้ำซ้อน (combined assurance)
- แจ้งเบาะแสการทุจริต/เหตุการณ์สำคัญ (fraud, data breach, AML STR) ต่อคณะกรรมการตรวจสอบทันที

---

## 6. แผนตรวจสอบประจำปีแบบอิงความเสี่ยง (Risk-Based Annual Audit Plan)

### 6.1 วิธีจัดทำแผน
คะแนนความเสี่ยง = **ผลกระทบ (Impact) × โอกาสเกิด (Likelihood)** ปรับด้วย regulatory sensitivity โดยพิจารณา: มูลค่าธุรกรรม, การสัมผัสข้อมูลบัตร (CHD), ผลกระทบต่อ ธปท./ปปง./PDPC, ประวัติข้อสังเกต, และการเปลี่ยนแปลงระบบ

| ระดับความเสี่ยง | คะแนน | ความถี่ตรวจขั้นต่ำ |
|-----------------|-------|---------------------|
| สูงมาก (Critical) | 20–25 | ทุก 6–12 เดือน |
| สูง (High) | 12–19 | ปีละ 1 ครั้ง |
| ปานกลาง (Medium) | 6–11 | ทุก 18–24 เดือน |
| ต่ำ (Low) | 1–5 | ทุก 36 เดือน |

### 6.2 จักรวาลการตรวจสอบและแผนปี 2570 (2027)

| # | หน่วยตรวจสอบ (Audit Universe) | โดเมน | ระดับความเสี่ยง | ไตรมาสที่วางแผน | มาตรฐาน/ข้อกำหนดอ้างอิง |
|---|-------------------------------|-------|-----------------|------------------|--------------------------|
| 1 | Authorization / Capture / Refund / Void & idempotency | Payment Ops | Critical | Q1 | Architecture §5, ธปท. IT risk |
| 2 | Ledger integrity & reconciliation (append-only, double-entry) | Payment Ops | Critical | Q2 | Architecture §6 |
| 3 | Settlement & merchant payout กับ sponsor bank | Payment Ops | High | Q3 | TODO A2 |
| 4 | Chargeback / dispute workflow | Payment Ops | High | Q3 | EMV 3DS, scheme rules |
| 5 | Tokenization vault & key management (HSM/KMS, dual control) | Security/PCI | Critical | Q1 | PCI-DSS v4.0 Req 3, 4 |
| 6 | Access control, RBAC, MFA, privileged access | Security/PCI | Critical | Q2 | PCI-DSS v4.0 Req 7, 8 |
| 7 | Logging, monitoring, SIEM, no-CHD-in-logs | Security/PCI | High | Q2 | PCI-DSS v4.0 Req 10 |
| 8 | Network segmentation, WAF, IDS/IPS, ASV scan review | Security/PCI | High | Q3 | PCI-DSS v4.0 Req 1, 11 |
| 9 | Vulnerability & patch mgmt, SAST/DAST, pentest closure | Security/PCI | High | Q4 | PCI-DSS v4.0 Req 6, 11 |
| 10 | 3-D Secure (EMV 3DS 2.x) authentication flow & fraud rules | Security/Ops | High | Q4 | EMV 3DS |
| 11 | KYC/CDD, merchant onboarding, beneficial ownership | AML | Critical | Q1 | พ.ร.บ. ปปง. |
| 12 | Sanction/PEP screening & name-matching | AML | Critical | Q2 | ปปง./AMLO, UN/OFAC lists |
| 13 | Transaction monitoring & STR/CTR reporting | AML | High | Q3 | ปปง. (รายงานภายใน 7 วัน) |
| 14 | PDPA — data subject rights, consent, retention, cross-border | Privacy | High | Q4 | PDPA / PDPC |
| 15 | Business continuity (BCP) & DR drill (RPO ≤ 5m, RTO ≤ 30m) | Resilience | High | Q4 | ธปท. BCP, Architecture §8 |
| 16 | Outsourcing / critical vendor governance (cloud, QSA, HSM) | Governance | Medium | Q3 | ธปท. outsourcing |
| 17 | Follow-up ข้อสังเกตค้าง (finding remediation) | ทุกโดเมน | High | ทุกไตรมาส | §8 |

> **หมายเหตุ:** จำนวนวันตรวจ (audit days) และการจัดสรรทรัพยากรจะยืนยันเมื่อปิด TODO A5 (co-source firm) และ A3 (วันประเมิน QSA) เพื่อ align กับรอบ RoC ประจำปี

### 6.3 การปรับแผนระหว่างปี
CAE อาจเสนอปรับแผน (dynamic/agile audit plan) เมื่อมีเหตุ เช่น go-live ระบบใหม่, incident สำคัญ, การเปลี่ยน sponsor bank, หรือคำสั่ง ธปท. — โดยขออนุมัติจากคณะกรรมการตรวจสอบ

---

## 7. กระบวนการตรวจสอบ (Audit Lifecycle)

| ขั้น | กิจกรรม | มาตรฐานเวลา (SLA) |
|-----|---------|--------------------|
| 1. Planning | ประกาศเริ่มตรวจ (engagement notice), กำหนด scope & objective, risk & control matrix | ล่วงหน้า ≥ 10 วันทำการ |
| 2. Fieldwork | ทดสอบการควบคุม, walkthrough, sampling, สัมภาษณ์, เก็บหลักฐาน (workpaper) | ตามแผน engagement |
| 3. Reporting | ร่างรายงาน → ตอบข้อสังเกต (management response) → รายงานฉบับสมบูรณ์ | ร่างภายใน 10 วันทำการหลังจบ fieldwork |
| 4. Follow-up | ติดตามผลการแก้ไขจนปิดประเด็น | ตามวันครบกำหนดของแต่ละ finding |

การเก็บ workpaper และรายงานต้องรักษาความลับ และเก็บรักษาไว้ **อย่างน้อย 5 ปี** (สอดคล้องแนวเก็บเอกสารของ ธปท./ปปง.)

---

## 8. การจัดระดับและการติดตามข้อสังเกต (Finding Rating & Tracking)

### 8.1 เกณฑ์การจัดระดับ

| ระดับ | นิยาม | กำหนดแก้ไข (Target) | ผู้อนุมัติปิด |
|-------|-------|---------------------|--------------|
| **Critical** | เสี่ยงต่อการสูญเสียเงิน/ข้อมูลบัตรรั่วไหล/ผิดกฎ ธปท.-ปปง. อย่างมีนัยสำคัญ | ≤ 30 วัน | คณะกรรมการตรวจสอบ |
| **High** | ควบคุมล้มเหลวที่อาจนำสู่ผลกระทบรุนแรง | ≤ 60 วัน | CAE |
| **Medium** | จุดอ่อนควบคุมที่ควรแก้ไข | ≤ 90 วัน | CAE |
| **Low** | โอกาสปรับปรุงประสิทธิภาพ | ≤ 180 วัน | หัวหน้าฝ่ายตรวจสอบอาวุโส |

### 8.2 ทะเบียนติดตามข้อสังเกต (Finding Register)
ทุกข้อสังเกตบันทึกในทะเบียนกลาง (issue-tracking) พร้อมฟิลด์: `finding_id`, โดเมน, ระดับความรุนแรง, รายละเอียด, root cause, action plan, เจ้าของ (owner), วันครบกำหนด, สถานะ (Open / In-Progress / Overdue / Closed-Verified), หลักฐานการปิด

- IA เป็นผู้ **ยืนยันการปิด (validate)** ด้วยหลักฐาน — เจ้าของงานปิดเองไม่ได้
- ข้อสังเกตที่ **เกินกำหนด (Overdue)** และ Critical/High ทั้งหมด ต้องรายงานต่อคณะกรรมการตรวจสอบทุกไตรมาส
- Critical ที่เกินกำหนด > 30 วัน ต้อง escalate ถึงคณะกรรมการบริษัท

---

## 9. การรายงาน (Reporting)

- **รายไตรมาส:** สรุปผลตรวจสอบ, สถานะแผน, ข้อสังเกตค้าง/เกินกำหนด, KPI → คณะกรรมการตรวจสอบ
- **รายปี:** ความเห็นภาพรวม (overall opinion) ต่อความเพียงพอของระบบควบคุมภายใน + ยืนยันความเป็นอิสระ
- **ทันที (ad-hoc):** เหตุการณ์สำคัญ — data breach, สงสัยฟอกเงิน (STR), การทุจริต, PCI non-compliance ร้ายแรง

---

## 10. ตัวชี้วัดผลงาน (KPI / QAIP)

| ตัวชี้วัด | เป้าหมาย |
|----------|---------|
| การตรวจสอบตามแผนสำเร็จ | ≥ 90% ของแผนอนุมัติ |
| ข้อสังเกต Critical/High ปิดตรงเวลา | ≥ 95% |
| ข้อสังเกตเกินกำหนด (Overdue) | ≤ 5% ของทั้งหมด |
| ประเมินคุณภาพภายนอก (External Quality Assessment) | อย่างน้อยทุก 5 ปี ตาม IIA |

CAE จัดทำ **Quality Assurance & Improvement Program (QAIP)** ประเมินตนเองประจำปี และให้ประเมินภายนอกทุก 5 ปี

---

## 11. การทบทวนกฎบัตร
คณะกรรมการตรวจสอบทบทวนและอนุมัติกฎบัตรฉบับนี้ **อย่างน้อยปีละ 1 ครั้ง** หรือเมื่อมีการเปลี่ยนแปลงกฎเกณฑ์ ธปท./ปปง./PDPC หรือโครงสร้างธุรกิจอย่างมีนัยสำคัญ

---
---

# Internal audit charter + annual audit plan covering payment ops, security, AML, PCI, finding tracking (English)

> Document: `31-internal-audit-charter.md` · Version 0.1 · Owner: Audit Committee
> Business: **Full Acquiring Payment Gateway** · Paid-up capital **THB 50M**
> Regulatory frame: Payment Systems Act B.E. 2560 (supervised by **BOT/ธปท.**), Anti-Money Laundering Act (**AMLO/ปปง.**), Personal Data Protection Act B.E. 2562 (**PDPC/PDPA**), **PCI-DSS v4.0**, **EMV 3-D Secure (3DS)**
> Supporting document for the BOT license application (see `../COMPLIANCE-TH.md` §5 — risk-management / governance policy)

> **Disclaimer:** This is a technical/governance template, not legal advice. Legal counsel and the QSA must review before submission.

---

## ⚠️ Assumptions / Open TODOs (unresolved external dependencies)

> The items below depend on external dependencies that are not yet closed — **do NOT treat as fact** until confirmed.

| # | Item | Status | Owner |
|---|------|--------|-------|
| A1 | **Legal entity name & registration no.** — placeholder "[บริษัท / Company]" used | TODO | Company Secretary |
| A2 | **Sponsor bank / acquiring partner** — not yet contracted; settlement/chargeback audit scope not locked | TODO | CEO |
| A3 | **QSA vendor (PCI-DSS v4.0)** — RoC/ASV assessor not selected; actual assessment dates unconfirmed | TODO | CISO / Compliance |
| A4 | **Actual paid-up capital** — THB 50M assumed per acquiring threshold; proof of payment and ≥ 75% maintenance evidence required | TODO | CFO |
| A5 | **Independent internal auditor (co-source firm)** — first-year co-source likely (in-house team not yet staffed); firm not selected | TODO | Audit Committee |
| A6 | **Data Protection Officer (DPO)** under PDPA — not yet formally appointed | TODO | CEO / Legal |

---

## 1. Purpose

This charter defines the mission, authority, responsibilities and independence of the **Internal Audit (IA)** function of **[บริษัท / Company]**, which operates an electronic acquiring service under the Payment Systems Act B.E. 2560.

Objectives:
- Provide independent assurance to the Board and Audit Committee on the adequacy and effectiveness of internal controls, risk management and governance.
- Support compliance with BOT requirements on IT risk, cyber resilience, business continuity and outsourcing.
- Assess compliance with PCI-DSS v4.0, AMLO's AML/CFT law, and PDPA.

IA adheres to the **IIA Global Internal Audit Standards (2024)** and BOT guidance on internal control and internal audit for supervised institutions.

---

## 2. Governance — Three Lines Model

| Line | Owner | Role |
|------|-------|------|
| **1st** | Operating units (Payment Ops, Engineering, Merchant Onboarding) | Own the risks and controls in the process |
| **2nd** | Compliance, Risk, CISO/Security, AML Officer, DPO | Set policy, monitor and challenge the 1st line |
| **3rd** | **Internal Audit (IA)** | Independent assurance over the 1st and 2nd lines; reports to the Audit Committee |

**Reporting lines:**
- **Functional:** the Chief Audit Executive (CAE) reports directly to the **Audit Committee**.
- **Administrative:** reports to the CEO for administrative matters only (budget, resources) to preserve independence.
- The Audit Committee approves the appointment/removal, evaluates and sets the remuneration of the CAE.

---

## 3. Independence & Objectivity

- IA holds no authority or responsibility over the operations it audits (no operational responsibility).
- Auditors must not audit work they were responsible for within the previous **12 months**.
- The CAE confirms organizational independence to the Audit Committee in writing **at least annually**.
- Any conflict of interest must be disclosed to the Audit Committee immediately; the engagement is reassigned or co-sourced (see TODO A5).

---

## 4. Authority

IA is authorized by the Board to:
1. Access, **without restriction**, all data, systems, records, assets and personnel needed — including PCI-scope systems (vault/tokenization), `ledger_entries`, `audit_log`, and the SIEM.
2. Access Board and committee meeting records.
3. Require explanations and cooperation from staff at all levels.
4. Engage external specialists (e.g., penetration testers, forensics) as needed.

**Scope:** IA covers all units, processes and outsourced/critical vendors (including the QSA, sponsor-bank interface, cloud, HSM/KMS), subject to contractual audit rights.

---

## 5. Responsibilities

- Prepare a **risk-based annual audit plan** and a 3-year rotational plan for Audit Committee approval.
- Execute audits per plan; assess control adequacy across payment ops, security, AML, PCI and PDPA.
- Track remediation of findings through to closure (finding tracking).
- Report quarterly to the Audit Committee.
- Coordinate with external auditors and the QSA to reduce duplication (combined assurance).
- Immediately escalate significant events (fraud, data breach, AML STR) to the Audit Committee.

---

## 6. Risk-Based Annual Audit Plan

### 6.1 Planning method
Risk score = **Impact × Likelihood**, adjusted for regulatory sensitivity, considering: transaction value, cardholder-data (CHD) exposure, BOT/AMLO/PDPC impact, prior findings, and system change.

| Risk level | Score | Minimum frequency |
|-----------|-------|-------------------|
| Critical | 20–25 | Every 6–12 months |
| High | 12–19 | Annually |
| Medium | 6–11 | Every 18–24 months |
| Low | 1–5 | Every 36 months |

### 6.2 Audit universe & FY2027 plan

| # | Auditable entity | Domain | Risk | Planned quarter | Standard / requirement |
|---|------------------|--------|------|------------------|------------------------|
| 1 | Authorization / Capture / Refund / Void & idempotency | Payment Ops | Critical | Q1 | Architecture §5, BOT IT risk |
| 2 | Ledger integrity & reconciliation (append-only, double-entry) | Payment Ops | Critical | Q2 | Architecture §6 |
| 3 | Settlement & merchant payout with sponsor bank | Payment Ops | High | Q3 | TODO A2 |
| 4 | Chargeback / dispute workflow | Payment Ops | High | Q3 | EMV 3DS, scheme rules |
| 5 | Tokenization vault & key management (HSM/KMS, dual control) | Security/PCI | Critical | Q1 | PCI-DSS v4.0 Req 3, 4 |
| 6 | Access control, RBAC, MFA, privileged access | Security/PCI | Critical | Q2 | PCI-DSS v4.0 Req 7, 8 |
| 7 | Logging, monitoring, SIEM, no-CHD-in-logs | Security/PCI | High | Q2 | PCI-DSS v4.0 Req 10 |
| 8 | Network segmentation, WAF, IDS/IPS, ASV scan review | Security/PCI | High | Q3 | PCI-DSS v4.0 Req 1, 11 |
| 9 | Vulnerability & patch mgmt, SAST/DAST, pentest closure | Security/PCI | High | Q4 | PCI-DSS v4.0 Req 6, 11 |
| 10 | 3-D Secure (EMV 3DS 2.x) authentication & fraud rules | Security/Ops | High | Q4 | EMV 3DS |
| 11 | KYC/CDD, merchant onboarding, beneficial ownership | AML | Critical | Q1 | AMLO Act |
| 12 | Sanction/PEP screening & name-matching | AML | Critical | Q2 | AMLO, UN/OFAC lists |
| 13 | Transaction monitoring & STR/CTR reporting | AML | High | Q3 | AMLO (report within 7 days) |
| 14 | PDPA — data-subject rights, consent, retention, cross-border | Privacy | High | Q4 | PDPA / PDPC |
| 15 | Business continuity (BCP) & DR drill (RPO ≤ 5m, RTO ≤ 30m) | Resilience | High | Q4 | BOT BCP, Architecture §8 |
| 16 | Outsourcing / critical-vendor governance (cloud, QSA, HSM) | Governance | Medium | Q3 | BOT outsourcing |
| 17 | Follow-up on open findings (remediation) | All domains | High | Every quarter | §8 |

> **Note:** Audit-day budgets and resource allocation will be confirmed once TODO A5 (co-source firm) and A3 (QSA assessment dates) are closed, to align with the annual RoC cycle.

### 6.3 In-year plan changes
The CAE may propose dynamic/agile plan changes for events such as a major system go-live, a significant incident, a change of sponsor bank, or a BOT directive — subject to Audit Committee approval.

---

## 7. Audit Lifecycle

| Stage | Activity | SLA |
|-------|----------|-----|
| 1. Planning | Engagement notice, scope & objectives, risk & control matrix | ≥ 10 business days ahead |
| 2. Fieldwork | Control testing, walkthroughs, sampling, interviews, workpapers | Per engagement plan |
| 3. Reporting | Draft report → management response → final report | Draft within 10 business days of fieldwork close |
| 4. Follow-up | Track remediation to closure | Per each finding's due date |

Workpapers and reports are confidential and retained for **at least 5 years** (consistent with BOT/AMLO record-keeping expectations).

---

## 8. Finding Rating & Tracking

### 8.1 Rating criteria

| Rating | Definition | Target remediation | Closure approver |
|--------|-----------|--------------------|------------------|
| **Critical** | Significant risk of financial loss / CHD breach / BOT-AMLO breach | ≤ 30 days | Audit Committee |
| **High** | Control failure that could cause severe impact | ≤ 60 days | CAE |
| **Medium** | Control weakness that should be remediated | ≤ 90 days | CAE |
| **Low** | Efficiency-improvement opportunity | ≤ 180 days | Senior IA lead |

### 8.2 Finding register
Every finding is logged in a central issue-tracking register with: `finding_id`, domain, severity, description, root cause, action plan, owner, due date, status (Open / In-Progress / Overdue / Closed-Verified), and closure evidence.

- IA **validates** closure with evidence — owners cannot self-close.
- All **Overdue** items and all Critical/High findings are reported to the Audit Committee quarterly.
- Critical findings overdue > 30 days escalate to the Board.

---

## 9. Reporting

- **Quarterly:** audit results, plan status, open/overdue findings, KPIs → Audit Committee.
- **Annual:** overall opinion on internal-control adequacy + independence confirmation.
- **Ad-hoc / immediate:** significant events — data breach, suspected money laundering (STR), fraud, serious PCI non-compliance.

---

## 10. KPIs / QAIP

| Metric | Target |
|--------|--------|
| Plan completion | ≥ 90% of approved plan |
| Critical/High findings closed on time | ≥ 95% |
| Overdue findings | ≤ 5% of total |
| External Quality Assessment | At least every 5 years (IIA) |

The CAE maintains a **Quality Assurance & Improvement Program (QAIP)** with annual self-assessment and an external assessment every 5 years.

---

## 11. Charter Review
The Audit Committee reviews and approves this charter **at least annually**, or upon significant changes in BOT/AMLO/PDPC regulation or business structure.
