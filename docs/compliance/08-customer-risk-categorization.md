# การจัดระดับความเสี่ยงลูกค้า (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1**
>
> รหัสเอกสาร: `COMP-08` · เวอร์ชัน 0.1 · เจ้าของเอกสาร: Chief Compliance Officer (CCO) / MLRO
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `04-org-chart-governance.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. และที่ปรึกษาด้าน AML ก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงก่อนยื่น ธปท. ห้ามกรอกค่าสมมติลงในเอกสารที่ยื่นจริง
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม; เกณฑ์ MCC ต้องห้าม/ต้องขออนุมัติ อาจถูกกำหนดเพิ่มโดยข้อสัญญา sponsor bank และ card scheme | ยังไม่สรุป | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — ยังไม่คัดเลือก; ใช้กำหนด scope การตรวจ CDE ที่เกี่ยวกับ risk engine/monitoring | ยังไม่สรุป | CISO |
> | A4 | **ผู้ให้บริการ sanction/PEP/adverse-media screening** (เช่น Dow Jones, Refinitiv World-Check, LexisNexis) — ยังไม่คัดเลือก | ยังไม่สรุป | MLRO |
> | A5 | **ผู้ให้บริการ fraud scoring / device fingerprinting / 3DS 2.x (EMV 3DS)** — ยังไม่คัดเลือก | ยังไม่สรุป | Head of Risk / CTO |
> | A6 | **ค่า threshold ตัวเลขจริง** (ยอด, velocity, chargeback ratio) ในเอกสารนี้เป็น **ค่าเริ่มต้นที่เสนอ (proposed baseline)** — ต้อง calibrate กับข้อมูลจริงและข้อกำหนด scheme (เช่น Visa VDMP/VFMP, Mastercard ECP/EFM) | TODO | CRO / MLRO |
> | A7 | **แผนที่ MCC → หมวดความเสี่ยง (MCC risk map)** ฉบับสมบูรณ์ — ต้อง sync กับ prohibited/high-brand-risk list ของ sponsor bank (A2) | TODO | Compliance |
>
> เกณฑ์ scheme (Visa/Mastercard) และเงื่อนไข sponsor bank อาจ **เข้มกว่า** เกณฑ์ในเอกสารนี้ — ให้ยึดเกณฑ์ที่เข้มกว่าเสมอ

---

## 1. วัตถุประสงค์และขอบเขต (Purpose & Scope)

นโยบายนี้กำหนดวิธี **จัดระดับความเสี่ยง (risk categorization)** ของลูกค้าและร้านค้า (merchant) ของ [บริษัท / Company]
ในฐานะผู้ให้บริการ Acquiring เพื่อให้:

1. ปฏิบัติตาม **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (AMLO)** และประกาศ/แนวปฏิบัติของสำนักงาน ปปง. เรื่องการทำ **CDD/EDD** ตามความเสี่ยง (Risk-Based Approach)
2. ปฏิบัติตามหลักเกณฑ์ **ธปท.** ด้านการบริหารความเสี่ยงของผู้ให้บริการชำระเงิน (merchant risk, fraud, chargeback, IT risk)
3. สอดคล้องกับข้อกำหนด **card scheme** (Visa, Mastercard) เรื่อง merchant monitoring, high-risk/high-brand-risk MCC และ dispute/fraud program
4. คุ้มครองข้อมูลส่วนบุคคลตาม **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** ภายใต้การกำกับของ **PDPC** — การเก็บ/ประมวลผลข้อมูลเพื่อ AML/fraud อาศัยฐาน "หน้าที่ตามกฎหมาย" และ "ประโยชน์อันชอบธรรม"

**ขอบเขต:** ใช้กับ merchant ทุกรายที่ onboard, ผู้ถือผลประโยชน์ที่แท้จริง (Ultimate Beneficial Owner — UBO), กรรมการ/ผู้มีอำนาจลงนาม และ (ที่เกี่ยวข้อง) sub-merchant ในโหมด Payment Facilitating

---

## 2. หลักการจัดระดับความเสี่ยง (Risk-Rating Principles)

- **Risk-Based Approach (RBA):** ระดับความเข้มของการตรวจสอบ (CDD/EDD) และความถี่ในการทบทวน แปรผันตามระดับความเสี่ยง
- **ประเมินตอน onboarding และตลอดวงจร (ongoing):** ระดับความเสี่ยงไม่คงที่ — ต้อง re-score เมื่อเกิด trigger event
- **หลายมิติ (multi-factor):** คะแนนความเสี่ยงมาจากหลายปัจจัยถ่วงน้ำหนัก ไม่ตัดสินจากปัจจัยเดียว
- **Fail closed / conservative:** เมื่อข้อมูลไม่ครบหรือคลุมเครือ ให้จัดระดับสูงกว่าไว้ก่อน แล้วปรับลดเมื่อพิสูจน์ได้ (สอดคล้อง design principle "Fail closed" ใน `ARCHITECTURE.md`)
- **Four-eyes / segregation of duties:** การอนุมัติลูกค้า High Risk ต้องแยกจากฝ่ายขาย (สอดคล้อง Three Lines of Defense ใน `04-org-chart-governance.md`)
- **บันทึกได้ตรวจสอบได้ (auditable):** ทุกการจัดระดับ/เปลี่ยนระดับบันทึกใน `audit_log` (append-only) พร้อมเหตุผลและผู้อนุมัติ

---

## 3. โมเดลการให้คะแนนความเสี่ยง (Risk-Scoring Model)

คะแนนรวม = ผลรวมถ่วงน้ำหนักของ 6 ปัจจัย (0–100) จากนั้น map เป็น 3 ระดับ

| ปัจจัยความเสี่ยง | น้ำหนัก | ตัวอย่างตัวชี้วัด |
|------------------|--------|------------------|
| **1. ประเภทธุรกิจ / MCC** | 30% | MCC อยู่ในหมวด high-risk (ดูข้อ 5) |
| **2. ความเสี่ยงประเทศ/ภูมิศาสตร์** | 15% | ประเทศจดทะเบียน/ที่ตั้ง/ลูกค้าปลายทางอยู่ในบัญชี FATF high-risk หรือประเทศที่ถูกคว่ำบาตร |
| **3. โครงสร้างเจ้าของ / UBO** | 15% | โครงสร้างซับซ้อน, nominee, UBO เป็น PEP, ถือหุ้นข้ามหลายชั้น |
| **4. รูปแบบการชำระเงิน / channel** | 15% | Card-Not-Present (CNP), MOTO, cross-border, high-ticket, subscription/recurring |
| **5. ประวัติ/ชื่อเสียง** | 15% | Sanction/PEP/adverse-media hit, ประวัติ chargeback/fraud, เคยถูก terminate (MATCH/TMF list ของ Mastercard) |
| **6. ปริมาณธุรกิจที่คาดการณ์** | 10% | ยอดขายต่อเดือนที่คาดการณ์, average ticket, ระยะเวลาส่งมอบสินค้า (delivery lag = future delivery risk) |

### การ map คะแนน → ระดับ

| คะแนนรวม | ระดับความเสี่ยง |
|----------|-----------------|
| 0–39 | **Low (ต่ำ)** |
| 40–69 | **Medium (ปานกลาง)** |
| 70–100 | **High (สูง)** |

> **Override บังคับ (mandatory high):** ไม่ว่าคะแนนรวมเท่าใด ให้จัดเป็น **High** ทันที หากเข้าเงื่อนไขข้อใดข้อหนึ่ง:
> sanction hit ที่ยืนยันแล้ว, UBO เป็น PEP, MCC อยู่ในหมวด high-risk, อยู่ในประเทศ FATF high-risk (black/grey list),
> พบใน MATCH/Terminated Merchant File, หรือโครงสร้าง nominee ที่ระบุ UBO ไม่ได้
>
> **Prohibited (ปฏิเสธ):** sanction match ที่ยืนยัน (OFAC/UN/EU/รายชื่อ ปปง.), ธุรกิจผิดกฎหมายตามกฎหมายไทย, MCC ต้องห้ามตามสัญญา sponsor bank/scheme → **ไม่รับ onboard**

---

## 4. ระดับความเสี่ยง 3 ระดับ — นิยาม การตรวจสอบ และการควบคุม

| หัวข้อ | **Low (ต่ำ)** | **Medium (ปานกลาง)** | **High (สูง)** |
|--------|---------------|----------------------|----------------|
| **ตัวอย่าง merchant** | ร้านค้าปลีกจดทะเบียนในไทย, MCC ทั่วไป, ticket ต่ำ, ส่งมอบทันที | e-commerce ทั่วไป, subscription, ยอดปานกลาง, CNP บางส่วน | Cross-border, high-ticket, travel/ticketing, forex/crypto-adjacent, gaming, nutraceutical, MCC high-risk |
| **ระดับ CDD** | Standard CDD (SDD ไม่ใช้เพราะ acquiring มีความเสี่ยง inherent) | Standard CDD + ตรวจเพิ่มบางรายการ | **Enhanced Due Diligence (EDD)** เต็มรูปแบบ |
| **เอกสาร onboarding** | หนังสือรับรองบริษัท, บัตร ปชช. กรรมการ, บัญชีธนาคาร settlement, URL/หน้าร้าน | + งบการเงิน, โครงสร้างผู้ถือหุ้น, ประมาณการยอดขาย | + UBO declaration + หลักฐาน, source of funds/wealth, business model deep-dive, ใบอนุญาตเฉพาะธุรกิจ (ถ้ามี), site/website review |
| **Sanction/PEP/adverse-media screening** | ตอน onboard + rescreen รายวันเทียบ list ที่อัปเดต | เหมือน Low | เหมือน Low + ตรวจ adverse media เชิงลึก manual |
| **ผู้อนุมัติ** | Onboarding Officer (1st line) | Onboarding Manager + Compliance review | **MLRO/CCO + AML Committee** (four-eyes) |
| **Transaction monitoring** | rule-based มาตรฐาน | rule-based + velocity เข้มขึ้น | rule-based เข้มสุด + manual review + lower thresholds |
| **Reserve / holdback** | โดยทั่วไปไม่มี | อาจมี rolling reserve ตาม chargeback profile | **Rolling reserve / delayed settlement** เป็นค่าเริ่มต้น (เช่น 5–10% ค้าง 90–180 วัน) |
| **3DS (EMV 3DS 2.x)** | บังคับสำหรับ CNP | บังคับ | บังคับ + step-up challenge เข้มขึ้น |
| **รอบทบทวน (review cadence)** | ทุก **36 เดือน** | ทุก **12 เดือน** | ทุก **6 เดือน** (หรือถี่กว่าเมื่อมี trigger) |

> **หมายเหตุ:** ค่า reserve %, ระยะค้าง และ threshold เป็น proposed baseline (ดู TODO A6) — ต้อง calibrate กับ sponsor bank และข้อมูลจริง

---

## 5. ความเสี่ยงตาม MCC (Merchant Category Code Risk)

MCC เป็นปัจจัยน้ำหนักสูงสุด (30%) [บริษัท / Company] แบ่ง MCC เป็น 4 กลุ่ม โดยอ้างอิงแนวปฏิบัติ card scheme
และเงื่อนไข sponsor bank (แผนที่ฉบับสมบูรณ์ = TODO A7)

| กลุ่ม MCC | ตัวอย่าง MCC / ธุรกิจ | การจัดการ |
|-----------|----------------------|-----------|
| **Prohibited (ต้องห้าม)** | ธุรกิจผิดกฎหมายไทย, เนื้อหา CSAM, ยาเสพติด, อาวุธผิดกฎหมาย, MCC ที่ sponsor bank/scheme ห้าม | **ปฏิเสธ onboard เด็ดขาด** |
| **High-risk (สูง — EDD บังคับ)** | 7995 (การพนัน/betting), 6211 (securities/brokers), 5967 (direct marketing/inbound telemarketing), 4816/5734 (บริการดิจิทัล/hosting บางประเภท), travel & ticketing (4722, 4511), timeshare, nutraceutical/supplement, forex, crypto on/off-ramp, adult (ที่กฎหมายอนุญาต), MLM | จัด **High**, EDD, rolling reserve, monitoring เข้มสุด, ต้องอนุมัติเป็นราย + อาจต้อง scheme registration (เช่น Visa GBPP/High-Brand-Risk) |
| **Medium-risk** | e-commerce ทั่วไป (5399, 5999), subscription/SaaS, electronics (5732), furniture, บริการทั่วไป | จัดเริ่มต้น **Medium**, standard CDD + monitoring เพิ่ม |
| **Low-risk** | ร้านอาหาร (5812), grocery (5411), retail หน้าร้านทั่วไป, utility, การกุศลที่จดทะเบียน | จัดเริ่มต้น **Low** |

**หลักปฏิบัติ MCC:**
- MCC ที่ประกาศตอน onboarding ต้องตรงกับธุรกิจจริง — ห้าม **miscoding / transaction laundering** (ผ่านธุรกรรมของธุรกิจอื่นเข้ามา)
- ตรวจจับ MCC mismatch ด้วย monitoring (ดูข้อ 6) — สินค้าจริงไม่ตรง MCC = red flag
- High-brand-risk MCC อาจต้อง **ลงทะเบียนกับ scheme** และเสียค่าธรรมเนียม/ปฏิบัติตามโปรแกรมเฉพาะ (TODO A2)

---

## 6. กฎการเฝ้าระวัง (Monitoring Rules)

Transaction monitoring ทำงานบน rule engine (`internal/service` Risk/Fraud Engine ตาม `ARCHITECTURE.md`) แบบ near-real-time
สำหรับ fraud/authorization และ batch สำหรับ AML pattern การ tune threshold ต่อ MCC/ระดับความเสี่ยง

### 6.1 กฎ Fraud / Transaction (near-real-time, per authorization)

| Rule ID | เงื่อนไข (baseline — ดู A6) | การกระทำ |
|---------|----------------------------|----------|
| **VEL-01** | > 5 ธุรกรรมจากบัตรใบเดียวใน 10 นาที | soft-decline + review |
| **VEL-02** | > 3 บัตรต่างใบจาก device/IP เดียวใน 1 ชม. | challenge (3DS step-up) + flag |
| **AMT-01** | ธุรกรรมเดี่ยว > 3× average ticket ของ merchant | hold + manual review |
| **CBK-01** | chargeback ratio ของ merchant > 0.9% (นับต่อเดือน) | เข้า watchlist + review; ใกล้ Visa VDMP/Mastercard ECP threshold |
| **DECL-01** | authorization decline rate > 20% ใน 24 ชม. (สัญญาณ card testing) | throttle + alert |
| **GEO-01** | IP ผู้ซื้อ/BIN ประเทศ mismatch กับ shipping/MCC ในสัดส่วนสูง | flag + EDD trigger |
| **MCC-01** | สินค้า/พฤติกรรมธุรกรรมไม่ตรง MCC ที่แจ้ง (transaction laundering) | freeze payout + investigation |

### 6.2 กฎ AML (batch / pattern-based)

| Rule ID | เงื่อนไข | การกระทำ |
|---------|----------|----------|
| **AML-STR-01** | รูปแบบธุรกรรมผิดปกติ/ไม่สมเหตุผลกับ profile (เช่น structuring, ยอดพุ่งผิดปกติ) | สืบสวนภายใน → หากมีมูล **รายงาน STR ต่อ ปปง.** |
| **AML-SCR-01** | คู่ค้า/ผู้รับเงินตรงกับ sanction/PEP list (rescreen รายวัน) | freeze + escalate MLRO |
| **AML-VOL-01** | ยอดจริงเกินประมาณการที่ประกาศตอน onboard อย่างมีนัยสำคัญ (เช่น > 200%) | re-underwrite + rescore |
| **AML-RFD-01** | อัตรา refund/reversal สูงผิดปกติ (สัญญาณ money movement/bust-out) | hold + review |

> **การรายงานตามกฎหมาย:** ธุรกรรมที่มีเหตุอันควรสงสัยต้องรายงาน **STR** ต่อ **สำนักงาน ปปง. (AMLO)** และรายงานธุรกรรมตามเกณฑ์ (เงินสด/โอน) ตามที่กฎหมายกำหนด — MLRO เป็นผู้รับผิดชอบ ห้าม **tipping-off** ลูกค้า

### 6.3 การจัดการ alert (Case Management)

- ทุก alert เปิดเป็น **case** มี SLA: High-risk merchant ปิดภายใน **2 วันทำการ**, Medium ภายใน **5 วันทำการ**
- Escalation: 1st line (analyst) → 2nd line (Compliance/MLRO) → AML Committee
- ทุกการตัดสิน (clear/escalate/report/terminate) บันทึกใน `audit_log` พร้อมเหตุผล (auditable ตาม PCI Req 10 & เกณฑ์ ธปท.)

---

## 7. รอบการทบทวน (Review Cadence) และ Trigger Events

### 7.1 การทบทวนตามกำหนด (Periodic)

| ระดับ | รอบทบทวนเต็ม (Full KYC refresh + rescore) |
|-------|-------------------------------------------|
| Low | ทุก **36 เดือน** |
| Medium | ทุก **12 เดือน** |
| High | ทุก **6 เดือน** |

- **Sanction/PEP rescreening:** ทุกระดับ **รายวัน** เทียบ list ที่อัปเดต (ไม่ผูกกับรอบ periodic)
- **Portfolio review:** Compliance ทบทวนภาพรวม portfolio ราย **ไตรมาส** เสนอ Risk Committee

### 7.2 การทบทวนเมื่อเกิดเหตุ (Event-Driven / Trigger)

ต้อง re-score ทันที (ภายใน 5 วันทำการ) เมื่อเกิด:
- Sanction/PEP/adverse-media hit ใหม่
- chargeback/fraud/decline ratio ทะลุ threshold (ข้อ 6)
- เปลี่ยนแปลงสาระสำคัญ: MCC, UBO/กรรมการ, business model, ปริมาณธุรกรรมพุ่ง
- คำขอ/หมายจากหน่วยงาน (ธปท., ปปง., ตำรวจ, ศาล)
- ผลจาก STR/investigation ภายใน

### 7.3 ผลลัพธ์การทบทวน
คงระดับเดิม / ปรับขึ้น (เพิ่ม EDD, reserve, ลด limit) / ปรับลง (เมื่อมีหลักฐานสนับสนุน) / **exit (offboard/terminate)** พร้อมแจ้ง scheme (MATCH) หากเข้าเงื่อนไข

---

## 8. บทบาทและความรับผิดชอบ (Roles — RACI)

| กิจกรรม | 1st line (Onboarding/Ops) | 2nd line (Compliance/MLRO/Risk) | 3rd line (Internal Audit) |
|---------|:---:|:---:|:---:|
| เก็บเอกสาร/KYC | **R** | C | I |
| ให้คะแนน & จัดระดับ (Low) | **R/A** | C | I |
| อนุมัติ Medium | R | **A** | I |
| อนุมัติ High / EDD | R | **A** (MLRO/CCO + AML Committee) | I |
| Transaction monitoring & case | R | **A** | I |
| STR ต่อ ปปง. | I | **R/A** (MLRO) | I |
| ทบทวนตามรอบ | R | **A** | I |
| ตรวจสอบความเป็นอิสระของกระบวนการ | I | C | **R/A** |

(R=Responsible, A=Accountable, C=Consulted, I=Informed — สอดคล้อง Three Lines of Defense ใน `04-org-chart-governance.md`)

---

## 9. ข้อมูล บันทึก และ PDPA

- **การเก็บรักษา:** เอกสาร KYC/CDD และบันทึกธุรกรรมเก็บอย่างน้อย **10 ปี** นับแต่ยุติความสัมพันธ์ ตามเกณฑ์ AMLO (หรือนานกว่าตามที่กฎหมายกำหนด)
- **PDPA (PDPC):** เก็บเท่าที่จำเป็น (data minimization); ฐานการประมวลผล = หน้าที่ตามกฎหมาย (AML) + ประโยชน์อันชอบธรรม (fraud prevention); มี retention schedule และควบคุมการเข้าถึงแบบ least-privilege/RBAC
- **PCI-DSS v4.0:** risk/monitoring engine ที่แตะข้อมูลธุรกรรมต้องอยู่ในขอบเขตควบคุม, log ไม่มี PAN/CVV, access control + audit trail (Req 7, 8, 10) — ระบบหลักเห็นเพียง `card_brand` + `card_last4` ตาม `ARCHITECTURE.md`

---
---

# Customer/merchant risk categorization (low/medium/high), MCC risk, monitoring rules, review cadence (English)

> Supporting document for the **Acquiring Service** license application under the **Payment Systems Act B.E. 2560**,
> supervised by the **Bank of Thailand (BOT / ธปท.)**, paid-up capital **THB 50 million**, alongside **PCI-DSS v4.0 Level 1**.
>
> Document ID: `COMP-08` · Version 0.1 · Owner: Chief Compliance Officer (CCO) / MLRO
> Related: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `04-org-chart-governance.md`
>
> **Disclaimer:** Structural reference only, not legal advice. Must be reviewed by BOT-licensing counsel and AML advisors before submission.

---

> [!IMPORTANT]
> **Assumptions / TODO** — resolve with real values before BOT submission. Do NOT insert fabricated values into the filed document.
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Legal company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **Sponsor / acquiring bank** — not signed; prohibited/high-brand-risk MCC list may be tightened by sponsor & scheme contracts | Open | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — not selected; defines CDE scope for the risk/monitoring engine | Open | CISO |
> | A4 | **Sanction/PEP/adverse-media screening vendor** (e.g. Dow Jones, Refinitiv World-Check, LexisNexis) — not selected | Open | MLRO |
> | A5 | **Fraud scoring / device fingerprinting / EMV 3DS 2.x provider** — not selected | Open | Head of Risk / CTO |
> | A6 | **Actual numeric thresholds** (amounts, velocity, chargeback ratio) shown here are a **proposed baseline** — calibrate to real data and scheme rules (Visa VDMP/VFMP, Mastercard ECP/EFM) | TODO | CRO / MLRO |
> | A7 | **Full MCC → risk-tier map** — must sync with the sponsor bank's prohibited / high-brand-risk lists (A2) | TODO | Compliance |
>
> Scheme (Visa/Mastercard) rules and sponsor-bank terms may be **stricter** than this policy — always apply the stricter standard.

---

## 1. Purpose & Scope

This policy defines how [บริษัท / Company], as an Acquiring provider, performs **risk categorization** of its customers and merchants to:

1. Comply with the **Anti-Money Laundering Act (AMLO)** and AMLO Office guidance on risk-based **CDD/EDD** (Risk-Based Approach).
2. Meet **BOT** requirements for payment-provider risk management (merchant, fraud, chargeback, IT risk).
3. Align with **card scheme** (Visa, Mastercard) requirements on merchant monitoring, high-risk/high-brand-risk MCCs, and dispute/fraud programs.
4. Protect personal data under the **Personal Data Protection Act B.E. 2562 (PDPA)**, supervised by **PDPC** — AML/fraud processing relies on the "legal obligation" and "legitimate interest" bases.

**Scope:** all onboarded merchants, Ultimate Beneficial Owners (UBOs), directors/authorized signatories, and (where relevant) sub-merchants under the Payment Facilitating mode.

---

## 2. Risk-Rating Principles

- **Risk-Based Approach (RBA):** intensity of CDD/EDD and review frequency scale with risk level.
- **Onboarding + ongoing:** risk is not static — re-score on trigger events.
- **Multi-factor:** score is a weighted blend of factors; no single factor decides.
- **Fail closed / conservative:** when data is incomplete or ambiguous, rate higher and adjust down only on proof (mirrors the "Fail closed" principle in `ARCHITECTURE.md`).
- **Four-eyes / segregation of duties:** High-risk approvals are separated from sales (per Three Lines of Defense in `04-org-chart-governance.md`).
- **Auditable:** every rating/change is logged to append-only `audit_log` with reason and approver.

---

## 3. Risk-Scoring Model

Total score = weighted sum of 6 factors (0–100), then mapped to 3 tiers.

| Risk factor | Weight | Example indicators |
|-------------|--------|--------------------|
| **1. Business type / MCC** | 30% | MCC falls in a high-risk category (Section 5) |
| **2. Country / geographic risk** | 15% | Registration/operation/end-customer country on FATF high-risk or sanctioned lists |
| **3. Ownership / UBO structure** | 15% | Complex structure, nominees, PEP UBO, multi-layer holdings |
| **4. Payment channel** | 15% | Card-Not-Present (CNP), MOTO, cross-border, high-ticket, subscription/recurring |
| **5. History / reputation** | 15% | Sanction/PEP/adverse-media hit, chargeback/fraud history, prior termination (Mastercard MATCH/TMF) |
| **6. Projected volume** | 10% | Projected monthly sales, average ticket, delivery lag (future-delivery risk) |

### Score → tier mapping

| Total score | Risk level |
|-------------|------------|
| 0–39 | **Low** |
| 40–69 | **Medium** |
| 70–100 | **High** |

> **Mandatory-high override:** regardless of total score, rate **High** immediately if any of: confirmed sanction hit, PEP UBO, high-risk MCC, FATF high-risk country (black/grey list), MATCH/Terminated Merchant File presence, or an un-identifiable nominee structure.
>
> **Prohibited (reject):** confirmed sanction match (OFAC/UN/EU/AMLO lists), businesses illegal under Thai law, MCCs prohibited by sponsor-bank/scheme contract → **do not onboard**.

---

## 4. The Three Tiers — Definitions, Due Diligence, Controls

| Item | **Low** | **Medium** | **High** |
|------|---------|------------|----------|
| **Example merchants** | Thai-registered retailer, generic MCC, low ticket, instant delivery | General e-commerce, subscription, mid volume, some CNP | Cross-border, high-ticket, travel/ticketing, forex/crypto-adjacent, gaming, nutraceutical, high-risk MCC |
| **CDD level** | Standard CDD (no SDD — acquiring is inherently risk-bearing) | Standard CDD + selective extras | Full **Enhanced Due Diligence (EDD)** |
| **Onboarding docs** | Company registration, director ID, settlement bank account, storefront/URL | + financials, ownership structure, sales projection | + UBO declaration & evidence, source of funds/wealth, business-model deep-dive, sector licenses, site/website review |
| **Sanction/PEP/adverse-media screening** | At onboarding + daily rescreen vs updated lists | Same as Low | Same as Low + deep manual adverse-media review |
| **Approver** | Onboarding Officer (1st line) | Onboarding Manager + Compliance review | **MLRO/CCO + AML Committee** (four-eyes) |
| **Transaction monitoring** | Standard rule-based | Rule-based + tighter velocity | Strictest rule-based + manual review + lower thresholds |
| **Reserve / holdback** | Generally none | Possible rolling reserve per chargeback profile | **Rolling reserve / delayed settlement** by default (e.g. 5–10% held 90–180 days) |
| **3DS (EMV 3DS 2.x)** | Required for CNP | Required | Required + stricter step-up challenge |
| **Review cadence** | Every **36 months** | Every **12 months** | Every **6 months** (or sooner on trigger) |

> **Note:** reserve %, hold periods, and thresholds are a proposed baseline (see TODO A6) — calibrate with the sponsor bank and real data.

---

## 5. MCC (Merchant Category Code) Risk

MCC carries the highest weight (30%). [บริษัท / Company] groups MCCs into 4 buckets, referencing card-scheme guidance and sponsor-bank terms (full map = TODO A7).

| MCC bucket | Example MCC / business | Handling |
|------------|------------------------|----------|
| **Prohibited** | Businesses illegal under Thai law, CSAM, narcotics, illegal weapons, MCCs banned by sponsor/scheme | **Reject onboarding outright** |
| **High-risk (EDD mandatory)** | 7995 (gambling/betting), 6211 (securities/brokers), 5967 (inbound telemarketing/direct marketing), 4816/5734 (certain digital/hosting), travel & ticketing (4722, 4511), timeshare, nutraceutical/supplements, forex, crypto on/off-ramp, adult (where legal), MLM | Rate **High**, EDD, rolling reserve, strictest monitoring, per-case approval + possible scheme registration (e.g. Visa GBPP / high-brand-risk) |
| **Medium-risk** | General e-commerce (5399, 5999), subscription/SaaS, electronics (5732), furniture, general services | Default **Medium**, standard CDD + extra monitoring |
| **Low-risk** | Restaurants (5812), grocery (5411), general storefront retail, utilities, registered charities | Default **Low** |

**MCC practices:**
- The onboarding MCC must match the real business — no **miscoding / transaction laundering** (channeling another business's transactions through the account).
- Detect MCC mismatch via monitoring (Section 6) — actual goods not matching MCC = red flag.
- High-brand-risk MCCs may require **scheme registration**, fees, and program-specific controls (TODO A2).

---

## 6. Monitoring Rules

Transaction monitoring runs on the rule engine (Risk/Fraud Engine in `internal/service`, per `ARCHITECTURE.md`): near-real-time for fraud/authorization and batch for AML patterns. Thresholds are tuned per MCC / risk tier.

### 6.1 Fraud / transaction rules (near-real-time, per authorization)

| Rule ID | Condition (baseline — see A6) | Action |
|---------|-------------------------------|--------|
| **VEL-01** | > 5 transactions on one card in 10 min | Soft-decline + review |
| **VEL-02** | > 3 distinct cards from one device/IP in 1 hr | Challenge (3DS step-up) + flag |
| **AMT-01** | Single transaction > 3× merchant average ticket | Hold + manual review |
| **CBK-01** | Merchant chargeback ratio > 0.9%/month | Watchlist + review; near Visa VDMP / Mastercard ECP thresholds |
| **DECL-01** | Authorization decline rate > 20% in 24 hr (card-testing signal) | Throttle + alert |
| **GEO-01** | High share of buyer IP/BIN-country mismatch vs shipping/MCC | Flag + EDD trigger |
| **MCC-01** | Goods/transaction pattern inconsistent with declared MCC (transaction laundering) | Freeze payout + investigation |

### 6.2 AML rules (batch / pattern-based)

| Rule ID | Condition | Action |
|---------|-----------|--------|
| **AML-STR-01** | Transaction pattern abnormal/unjustified vs profile (e.g. structuring, unexplained spikes) | Internal investigation → if grounded, **file STR to AMLO** |
| **AML-SCR-01** | Counterparty/payee matches sanction/PEP list (daily rescreen) | Freeze + escalate to MLRO |
| **AML-VOL-01** | Actual volume materially exceeds declared projection (e.g. > 200%) | Re-underwrite + rescore |
| **AML-RFD-01** | Abnormally high refund/reversal rate (money-movement / bust-out signal) | Hold + review |

> **Statutory reporting:** suspicious transactions must be reported via **STR to the AMLO Office**, plus threshold-based transaction reports as required by law — the MLRO owns this. **No tipping-off** the customer.

### 6.3 Case management

- Every alert opens a **case** with SLA: High-risk merchant closed within **2 business days**, Medium within **5 business days**.
- Escalation: 1st line (analyst) → 2nd line (Compliance/MLRO) → AML Committee.
- Every decision (clear/escalate/report/terminate) is logged to `audit_log` with rationale (auditable per PCI Req 10 and BOT requirements).

---

## 7. Review Cadence & Trigger Events

### 7.1 Periodic review

| Tier | Full review (KYC refresh + rescore) |
|------|-------------------------------------|
| Low | Every **36 months** |
| Medium | Every **12 months** |
| High | Every **6 months** |

- **Sanction/PEP rescreening:** **daily** for all tiers vs updated lists (independent of the periodic cycle).
- **Portfolio review:** Compliance reviews the whole portfolio **quarterly** and reports to the Risk Committee.

### 7.2 Event-driven / trigger review

Re-score immediately (within 5 business days) on:
- New sanction/PEP/adverse-media hit
- Chargeback/fraud/decline ratio breaching thresholds (Section 6)
- Material change: MCC, UBO/directors, business model, transaction-volume spike
- Request/order from an authority (BOT, AMLO, police, court)
- Outcome of an STR / internal investigation

### 7.3 Review outcomes
Keep tier / upgrade (add EDD, reserve, lower limits) / downgrade (with supporting evidence) / **exit (offboard/terminate)** with scheme notification (MATCH) where applicable.

---

## 8. Roles & Responsibilities (RACI)

| Activity | 1st line (Onboarding/Ops) | 2nd line (Compliance/MLRO/Risk) | 3rd line (Internal Audit) |
|----------|:---:|:---:|:---:|
| Collect KYC docs | **R** | C | I |
| Score & rate (Low) | **R/A** | C | I |
| Approve Medium | R | **A** | I |
| Approve High / EDD | R | **A** (MLRO/CCO + AML Committee) | I |
| Transaction monitoring & cases | R | **A** | I |
| STR to AMLO | I | **R/A** (MLRO) | I |
| Periodic review | R | **A** | I |
| Independent assurance of process | I | C | **R/A** |

(R=Responsible, A=Accountable, C=Consulted, I=Informed — aligns with Three Lines of Defense in `04-org-chart-governance.md`.)

---

## 9. Data, Records & PDPA

- **Retention:** KYC/CDD documents and transaction records kept at least **10 years** after relationship ends per AMLO rules (or longer where required).
- **PDPA (PDPC):** data minimization; processing bases = legal obligation (AML) + legitimate interest (fraud prevention); documented retention schedule with least-privilege/RBAC access control.
- **PCI-DSS v4.0:** any risk/monitoring component touching transaction data stays within scope; logs contain no PAN/CVV; access control + audit trail (Req 7, 8, 10). The core system only sees `card_brand` + `card_last4` per `ARCHITECTURE.md`.
