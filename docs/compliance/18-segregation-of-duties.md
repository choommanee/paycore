# เมทริกซ์การแบ่งแยกหน้าที่ (SoD) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1**
>
> รหัสเอกสาร: `COMP-18` · เวอร์ชัน 0.1 · เจ้าของเอกสาร: Chief Information Security Officer (CISO) ร่วมกับ Chief Compliance Officer (CCO)
> เอกสารที่เกี่ยวข้อง: `04-org-chart-governance.md`, `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท. และ QSA ก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงก่อนยื่น ธปท.
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม จึงยังไม่ระบุ role สำหรับ settlement instruction ปลายทาง (เส้นทาง B ตาม ROADMAP) | ยังไม่สรุป | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — ยังไม่คัดเลือก จึงยังไม่กำหนดผู้ทบทวน SoD matrix ภายนอก | ยังไม่สรุป | CISO |
> | A4 | **IAM/IdP และ PAM tooling จริง** (เช่น IdP + privileged access management + secrets vault) — ยังไม่จัดซื้อ ชื่อผลิตภัณฑ์เป็น placeholder | ยังไม่สรุป | CISO / Head of Infrastructure |
> | A5 | **ระบบ HSM/KMS จริง** สำหรับ dual control / split knowledge ของ key ceremony | ยังไม่สรุป | CISO |
> | A6 | **จำนวน headcount จริง ณ วันยื่น** — matrix นี้ตั้งบนสมมติฐานทีมขั้นต่ำตาม ROADMAP ข้อ 3; หากทีมเล็กเกินจนควบรวมบทบาทที่ขัดกัน ต้องประกาศ **compensating control** เป็นลายลักษณ์อักษร | TODO | CISO / CCO |
> | A7 | **External independent audit firm** (co-source สำหรับตรวจ SoD ประจำปี) | ยังไม่สรุป | Audit Committee |
>
> ห้ามกรอกชื่อผลิตภัณฑ์/ชื่อบุคคลสมมติในช่องข้างต้นลงในเอกสารที่ยื่นจริง — ต้องเป็นข้อมูลที่ยืนยันได้เท่านั้น

---

## 1. วัตถุประสงค์และขอบเขต (Purpose & Scope)

เอกสารนี้กำหนด **นโยบายการแบ่งแยกหน้าที่ (Segregation of Duties — SoD)**, **การให้สิทธิ์ตามหลักสิทธิ์ขั้นต่ำ (Least Privilege)** และ **การควบคุมสองชั้น (Dual Control / Four-Eyes)** สำหรับการดำเนินงานสำคัญของ [บริษัท / Company] ในฐานะผู้ให้บริการ Full Acquiring Payment Gateway

**วัตถุประสงค์:**

1. ป้องกันไม่ให้บุคคลใดบุคคลหนึ่งควบคุมธุรกรรม/กระบวนการที่มีความเสี่ยงตั้งแต่ต้นจนจบ (end-to-end) โดยลำพัง ซึ่งอาจนำไปสู่การทุจริต การฟอกเงิน หรือความผิดพลาดที่ไม่ถูกตรวจพบ
2. สอดคล้องกับหลักธรรมาภิบาลและ Three Lines of Defense ตามประกาศ ธปท. ว่าด้วยการกำกับดูแลความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk) และ Cyber Resilience (อ้างอิง `04-org-chart-governance.md`)
3. ปฏิบัติตาม **PCI-DSS v4.0** โดยเฉพาะ Requirement 7 (least privilege / need-to-know), Requirement 8 (identification & authentication), Requirement 3 (dual control & split knowledge สำหรับ cryptographic key), Requirement 10 (logging & accountability)
4. รองรับข้อกำหนดของ **ปปง./AMLO** (การแยกผู้ทำรายการออกจากผู้อนุมัติ SAR/STR และ sanction screening) และ **PDPC/PDPA** (least privilege ต่อข้อมูลส่วนบุคคล)

**ขอบเขต:** ครอบคลุมทุกระบบใน production, PCI Cardholder Data Environment (CDE), tokenization vault, ledger/settlement, merchant onboarding, key management, และการเข้าถึงฐานข้อมูล/โครงสร้างพื้นฐาน สำหรับพนักงานประจำ, ผู้รับเหมา และ service account ทั้งหมด

---

## 2. นิยามและหลักการ (Definitions & Principles)

| คำ | นิยาม |
|----|-------|
| **Segregation of Duties (SoD)** | การจัดสรรหน้าที่ให้ไม่มีบุคคลเดียวควบคุมทั้ง "การริเริ่ม (initiate)", "การอนุมัติ (authorize)", "การบันทึก (record)" และ "การกระทบยอด (reconcile)" ของธุรกรรมเดียวกัน |
| **Least Privilege** | ผู้ใช้/service account ได้รับสิทธิ์เพียงเท่าที่จำเป็นต่อหน้าที่ (need-to-know / need-to-use) เท่านั้น และในระยะเวลาที่จำเป็นเท่านั้น |
| **Dual Control (Four-Eyes)** | การกระทำสำคัญต้องมีบุคคล **2 คนที่แยกจากกัน** ร่วมกันทำให้สำเร็จ (maker–checker) โดยไม่มีใครทำครบทุกขั้นได้ลำพัง |
| **Split Knowledge** | ความลับ (เช่น cryptographic key component) ถูกแบ่งให้บุคคลต่างคน โดยไม่มีใครถือครบเพื่อ reconstruct key ได้คนเดียว (PCI Req 3.7) |
| **Maker–Checker** | ผู้ทำรายการ (maker) และผู้ตรวจสอบ/อนุมัติ (checker) ต้องเป็นคนละคนเสมอ |
| **JIT / Break-glass** | การให้สิทธิ์ระดับสูงแบบชั่วคราว (Just-In-Time) ที่หมดอายุอัตโนมัติ + บันทึกเหตุผลและอนุมัติล่วงหน้า |
| **Toxic Combination** | คู่สิทธิ์/หน้าที่ที่หากอยู่กับคนเดียวจะสร้างความเสี่ยงทุจริตร้ายแรง (ดูข้อ 5) |

**หลักการ 7 ข้อ:**

1. **แยก 4 บทบาทหลักของธุรกรรม** — initiate / authorize / record / reconcile ต้องกระจายอย่างน้อย 2 บุคคล
2. **แยก Dev ออกจาก Prod** — นักพัฒนา (developer) ห้ามมีสิทธิ์ deploy หรือแก้ข้อมูลใน production โดยลำพัง (PCI Req 6.5.3, 7)
3. **แยกผู้ปฏิบัติออกจากผู้ตรวจสอบ** — 1st line (ปฏิบัติการ) แยกจาก 2nd line (risk/compliance/security) และ 3rd line (internal audit)
4. **Deny by default** — ไม่มีสิทธิ์ใดถูกให้โดยปริยาย ทุกสิทธิ์ต้องขอ–อนุมัติ–ทบทวน
5. **สิทธิ์มีอายุ** — สิทธิ์ระดับสูงเป็นแบบชั่วคราว/หมุนเวียน ไม่ใช่ถาวร
6. **ทุกการกระทำระบุตัวได้** — ห้ามใช้ shared/generic account; ทุก action มี identity ที่ตรวจสอบย้อนได้ (PCI Req 8.2, 10)
7. **Compensating control เมื่อแยกไม่ได้จริง** — หากทีมเล็กจนต้องควบบทบาท ต้องมีมาตรการชดเชยที่ได้รับอนุมัติจาก Risk Committee และทบทวนทุกไตรมาส

---

## 3. บทบาทและกลุ่มสิทธิ์ (Roles / RBAC Groups)

การให้สิทธิ์ทั้งหมดยึด **Role-Based Access Control (RBAC)** ผ่าน IdP กลาง (A4) — ไม่มีการให้สิทธิ์รายบุคคลนอกโครงสร้าง role ยกเว้นผ่าน JIT ที่มีเวลาหมดอายุ

| รหัส Role | ชื่อบทบาท | Line of Defense | สิทธิ์โดยสรุป |
|-----------|----------|-----------------|--------------|
| R-ENG-DEV | Software Engineer (Developer) | 1st | เขียนโค้ด, เข้าถึง non-prod, อ่าน log ที่ปิดบังข้อมูล (masked); **ไม่มี** prod write |
| R-SRE-OPS | SRE / Platform Operator | 1st | deploy ผ่าน pipeline, จัดการ infra prod (แบบ break-glass), ไม่มีสิทธิ์แก้ business data โดยตรง |
| R-DBA | Database Administrator | 1st | จัดการ schema/สำรองข้อมูล; อ่าน/แก้ข้อมูล prod ต้องผ่าน JIT + dual approval |
| R-PAY-OPS | Payment Operations | 1st | ดูธุรกรรม, ริเริ่ม refund/void ภายใต้ threshold, จัดการ dispute/chargeback |
| R-SETTLE | Settlement / Treasury Ops | 1st | ริเริ่ม payout/settlement batch (maker); **ไม่มี** สิทธิ์อนุมัติ payout |
| R-SETTLE-APV | Settlement Approver | 1st (senior) | อนุมัติ payout/settlement batch (checker); **ไม่มี** สิทธิ์ริเริ่ม |
| R-MERCH-ONB | Merchant Onboarding Officer | 1st | สร้าง/แก้ merchant profile (maker), ทำ KYC/CDD; ไม่ activate เอง |
| R-MERCH-APV | Merchant Approver / MLRO delegate | 2nd | อนุมัติ activate merchant, อนุมัติผล KYC/EDD (checker) |
| R-RISK | Risk / Fraud Analyst | 2nd | ปรับ fraud rule (maker), review alert; deploy rule ต้องผ่าน checker |
| R-COMPL | Compliance / AML Officer | 2nd | ทำ sanction screening review, ริเริ่ม SAR/STR (maker) |
| R-MLRO | MLRO (Money Laundering Reporting Officer) | 2nd | อนุมัติและยื่น SAR/STR ต่อ ปปง. (checker); ดู `07-sar-str-procedure.md` |
| R-DPO | Data Protection Officer | 2nd | อนุมัติคำขอใช้ข้อมูลส่วนบุคคลนอกขอบเขตปกติ (PDPA), review DSAR |
| R-SEC | Security Engineer / SecOps | 2nd | จัดการ WAF/IDS/SIEM, review access; **ไม่มี** สิทธิ์อนุมัติ access ของตนเอง |
| R-KEY-CUST-A | Key Custodian A | 2nd | ถือ key component A (split knowledge) |
| R-KEY-CUST-B | Key Custodian B | 2nd | ถือ key component B (split knowledge) — คนละคนกับ A เสมอ |
| R-IAM-ADMIN | Identity & Access Admin | 2nd | สร้าง/ปิด account (maker); การให้ privileged role ต้องผ่าน checker |
| R-AUDIT | Internal Auditor | 3rd | อ่านอย่างเดียว (read-only) ทุก log/หลักฐาน; **ไม่มี** write ในระบบปฏิบัติการ |

> **service account** ทั้งหมดถือเป็น identity แยก มี owner ที่ระบุได้, ผูก scope น้อยที่สุด, หมุนเวียน credential อัตโนมัติผ่าน secrets vault และห้ามใช้ล็อกอินแบบ interactive

---

## 4. เมทริกซ์การแบ่งแยกหน้าที่ (SoD Matrix)

เครื่องหมาย: **M** = Maker (ริเริ่ม/ทำรายการ) · **C** = Checker (อนุมัติ/ตรวจ) · **R** = Review เท่านั้น (read-only) · **—** = ไม่มีสิทธิ์
กฎ: ในหนึ่งกระบวนการ **Maker และ Checker ต้องเป็นคนละบุคคล** เสมอ (ระบบบังคับด้วย IdP/แอป ไม่พึ่งวินัยส่วนบุคคล)

| กระบวนการสำคัญ | R-PAY-OPS | R-SETTLE | R-SETTLE-APV | R-MERCH-ONB | R-MERCH-APV | R-RISK | R-COMPL | R-MLRO | R-SRE-OPS | R-DBA | R-IAM-ADMIN | R-SEC | R-AUDIT |
|----------------|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|
| Merchant onboarding / activate | — | — | — | **M** | **C** | R | R | — | — | — | — | — | R |
| Refund / Void ≤ threshold | **M** | — | — | — | — | — | — | — | — | — | — | — | R |
| Refund / Void > threshold | **M** | — | — | — | **C** | — | — | — | — | — | — | — | R |
| Settlement / Payout batch | — | **M** | **C** | — | — | — | — | — | — | — | — | — | R |
| แก้ไข fraud/risk rule ใน prod | R | — | — | — | — | **M** | — | — | **C** | — | — | R | R |
| Sanction screening — clear a hit | — | — | — | — | — | R | **M** | **C** | — | — | — | — | R |
| ยื่น SAR/STR ต่อ ปปง. | — | — | — | — | — | — | **M** | **C** | — | — | — | — | R |
| Deploy โค้ดสู่ production | — | — | — | — | — | — | — | — | **M** | — | — | **C** | R |
| แก้ไขข้อมูลใน production DB | — | — | — | — | — | — | — | — | — | **M** | — | **C** | R |
| สร้าง/ให้ privileged access | — | — | — | — | — | — | — | — | — | — | **M** | **C** | R |
| Cryptographic key ceremony | — | — | — | — | — | — | — | — | — | — | — | **M/C** | R |
| ปรับ config WAF/firewall (CDE) | — | — | — | — | — | — | — | — | R | — | — | **M** | R |

> หมายเหตุ key ceremony: ต้องมี **Key Custodian A + Key Custodian B** (split knowledge) ร่วมกับ Security เป็นสักขีพยาน — ไม่มีบุคคลใดเข้าถึง key component ครบทั้งสองส่วน (ดูข้อ 6.3)

---

## 5. คู่หน้าที่ต้องห้าม (Toxic / Conflicting Combinations)

ระบบ IdP และ application-level check ต้อง **บล็อกเชิงป้องกัน (preventive)** ไม่ให้ role ต่อไปนี้อยู่กับ identity เดียว หากจำเป็นต้องควบต้องมี compensating control ที่อนุมัติแล้ว (ข้อ 8)

| # | คู่ที่ห้ามควบ | เหตุผล/ความเสี่ยง |
|---|---------------|-------------------|
| T1 | R-SETTLE **+** R-SETTLE-APV | ริเริ่มและอนุมัติจ่ายเงินเอง → ยักยอกเงิน settlement |
| T2 | R-MERCH-ONB **+** R-MERCH-APV | สร้างและอนุมัติ merchant ปลอมเอง → ฟอกเงิน/รับเงินเข้าบัญชีตน |
| T3 | R-ENG-DEV **+** สิทธิ์ deploy/แก้ prod DB | ฝัง backdoor แล้วปล่อยขึ้น prod เองโดยไม่มีผู้ตรวจ (PCI Req 6/7) |
| T4 | R-COMPL/R-MLRO ทำทั้ง maker และ checker ของ SAR เดียวกัน | ปกปิดธุรกรรมน่าสงสัยได้ลำพัง (ขัดเจตนารมณ์ ปปง.) |
| T5 | R-IAM-ADMIN อนุมัติสิทธิ์ให้ตนเอง | ยกระดับสิทธิ์ตนเองแบบไม่มีผู้ตรวจ |
| T6 | R-AUDIT **+** role ที่มี write ใด ๆ ในระบบปฏิบัติการ | ทำลายความเป็นอิสระของ 3rd line |
| T7 | R-KEY-CUST-A **+** R-KEY-CUST-B | ถือ key component ครบ → ถอดรหัส PAN ได้ลำพัง (PCI Req 3.7) |
| T8 | R-RISK maker **+** checker ของ rule เดียวกัน | ปิด fraud control เพื่อปล่อยธุรกรรมทุจริต |

---

## 6. การควบคุมสองชั้นสำหรับปฏิบัติการสำคัญ (Dual Control)

### 6.1 Threshold และ SLA ของ maker–checker

| ปฏิบัติการ | Threshold ที่ต้อง dual control | ผู้ทำ (Maker) | ผู้อนุมัติ (Checker) | SLA อนุมัติ |
|-----------|-------------------------------|---------------|---------------------|-------------|
| Refund/Void | > 50,000 บาท/รายการ หรือ > 200,000 บาท/วัน/merchant | R-PAY-OPS | R-MERCH-APV (senior ops) | ≤ 4 ชม. ทำการ |
| Settlement/Payout batch | **ทุก batch** (ไม่มีขั้นต่ำ) | R-SETTLE | R-SETTLE-APV | ≤ 2 ชม. ก่อน cut-off |
| Merchant activate | **ทุกราย** | R-MERCH-ONB | R-MERCH-APV | ตาม SLA onboarding |
| Deploy ขึ้น production | **ทุกครั้ง** | R-SRE-OPS | R-SEC (หรือ senior SRE คนละคน) | ตาม change window |
| แก้ไข production DB (นอก pipeline) | **ทุกครั้ง** | R-DBA | R-SEC | JIT ≤ 60 นาที |
| ให้ privileged access | **ทุกครั้ง** | R-IAM-ADMIN | Line manager + R-SEC | ≤ 1 วันทำการ |
| Key ceremony (generate/rotate) | **ทุกครั้ง** | Custodian A | Custodian B + Security witness | ตามกำหนดพิธี |
| ปรับ fraud rule ใน prod | **ทุกครั้ง** | R-RISK | R-SRE-OPS / senior risk | ≤ 8 ชม. |

> Threshold ข้างต้นเป็น **ค่าตั้งต้นที่แนะนำ** — ต้องปรับให้สอดคล้องกับ risk appetite ที่ Risk Committee อนุมัติ และทบทวนอย่างน้อยปีละครั้ง

### 6.2 กลไกบังคับ dual control

- **บังคับด้วยระบบ ไม่ใช่ด้วยวินัย** — แอปพลิเคชัน/CI-CD/IdP ปฏิเสธคำขอที่ maker = checker เสมอ
- ทุกการอนุมัติ dual control บันทึกลง `audit_log` (append-only, ดู ARCHITECTURE ข้อ 6) พร้อม identity ทั้งสองฝ่าย, timestamp, เหตุผล และ before/after state
- Checker ต้องยืนยันด้วย **MFA** ขณะอนุมัติ (PCI Req 8.4/8.5)
- ห้าม self-approval, ห้าม approval ย้อนหลัง, ห้ามใช้ delegation ที่ทำให้ maker กลายเป็น checker

### 6.3 Split knowledge & dual control ของ cryptographic key (PCI Req 3.5–3.7)

- Master/Key-Encryption-Key สร้างและหมุนเวียนใน **HSM/KMS** (A5) ผ่าน key ceremony ที่มี **Key Custodian A + B** (คนละคน) และ Security เป็นสักขีพยาน
- ไม่มีบุคคลใดเข้าถึง key component ครบทุกส่วน (split knowledge); ไม่มี clear-text key อยู่นอก HSM
- ทุกพิธีมี **script + ใบลงนาม (ceremony log)** เก็บเป็นหลักฐานสำหรับ QSA
- การหมุน key ตามรอบที่กำหนดในนโยบาย key management และเมื่อสงสัยว่ารั่วไหล

---

## 7. หลักสิทธิ์ขั้นต่ำและวงจรชีวิตสิทธิ์ (Least Privilege & Access Lifecycle)

### 7.1 วงจร Joiner–Mover–Leaver (JML)

| เหตุการณ์ | การดำเนินการ | SLA |
|-----------|--------------|-----|
| **Joiner** | ให้เฉพาะ role ตามตำแหน่ง (RBAC template) โดย line manager ร้อง + R-IAM-ADMIN maker + checker | ภายในวันเริ่มงาน |
| **Mover** (เปลี่ยนตำแหน่ง) | **เพิกถอนสิทธิ์เดิมทั้งหมด** แล้วให้ใหม่ตามตำแหน่งใหม่ (ห้าม accumulate/privilege creep) | ≤ 2 วันทำการ |
| **Leaver** | ปิด/ระงับทุก account, เพิกถอน token/คีย์, คืนอุปกรณ์ | **≤ 4 ชม.** สำหรับ CDE/privileged; ≤ 24 ชม. อื่น ๆ |

### 7.2 การให้สิทธิ์ระดับสูงแบบชั่วคราว (JIT / Break-glass)

- สิทธิ์ระดับสูง/เข้า CDE เป็นแบบ **Just-In-Time** ผ่าน PAM (A4): ขอ → ระบุเหตุผล → อนุมัติ (dual) → ใช้งาน → **หมดอายุอัตโนมัติ (เช่น ≤ 60 นาที)**
- Break-glass account สำหรับเหตุฉุกเฉิน: credential ปิดผนึกใน vault, การเปิดใช้ trigger alert ทันทีไปยัง CISO + on-call, ต้อง post-incident review ภายใน 24 ชม. และหมุน credential หลังใช้

### 7.3 การทบทวนสิทธิ์ (Access Recertification)

| ประเภทสิทธิ์ | ความถี่ทบทวน | ผู้ทบทวน |
|--------------|--------------|----------|
| Privileged / CDE / prod | **ทุกไตรมาส** | Line manager + R-SEC (PCI Req 7.2.4) |
| สิทธิ์ทั่วไป (non-privileged) | ทุก 6 เดือน | Line manager |
| Service account & scope | ทุกไตรมาส | Owner + R-SEC |
| SoD toxic-combination scan | ทุกเดือน (อัตโนมัติ) + ทบทวนรายไตรมาส | R-IAM-ADMIN + R-AUDIT |

> ผลการทบทวนและการเพิกถอนบันทึกเป็นหลักฐานให้ Internal Audit และ QSA; สิทธิ์ที่ไม่มีการใช้งาน (dormant) เกิน 90 วันจะถูกระงับอัตโนมัติ

---

## 8. Compensating Controls (เมื่อทีมเล็กเกินจะแยกได้เต็มรูปแบบ)

ในระยะเริ่มต้น (Phase 0–2 ตาม ROADMAP) headcount อาจไม่พอแยกทุกบทบาท (A6) หากจำเป็นต้องควบบทบาทที่ขัดกัน ต้องมี compensating control ที่ **อนุมัติเป็นลายลักษณ์อักษรโดย Risk Committee** และทบทวนทุกไตรมาส เช่น

- ให้ **checker เป็นบุคคลจาก 2nd/3rd line** (เช่น CISO/CCO) แทนเพื่อนร่วมทีมเมื่อ maker–checker ในทีมเดียวไม่พอ
- เพิ่ม **detective control**: การสอบทานรายวัน/สุ่มตรวจ, alert อัตโนมัติเมื่อเกิด single-person action บนปฏิบัติการสำคัญ, การกระทบยอด ledger รายวัน
- **บันทึกการควบบทบาท** ลงทะเบียน SoD exception register พร้อมวันหมดอายุ และแผนแยกบทบาทเมื่อ headcount เพิ่ม
- Internal Audit ทดสอบประสิทธิผลของ compensating control ทุกไตรมาสและรายงานต่อ Audit Committee

---

## 9. การบังคับใช้ ตรวจสอบ และการรายงาน (Enforcement, Audit & Reporting)

- **บังคับด้วยระบบ:** IdP/PAM/แอปเป็นจุดบังคับ SoD และ least privilege; นโยบายที่บังคับด้วยเอกสารอย่างเดียวถือว่าไม่เพียงพอ
- **Logging (PCI Req 10):** ทุกการให้สิทธิ์, การใช้สิทธิ์ระดับสูง, การอนุมัติ dual control, break-glass ลง `audit_log` แบบ append-only, ป้องกันการแก้ไข, เก็บอย่างน้อย **1 ปี online + รวม 12 เดือนพร้อมใช้ทันที** (จัดเก็บระยะยาวตาม `11-data-retention-deletion.md`)
- **การตรวจสอบภายใน:** Internal Audit (3rd line) ทดสอบ SoD matrix และ toxic-combination อย่างน้อยปีละครั้ง รายงานต่อ Audit Committee
- **การประเมินภายนอก:** QSA (A3) ทบทวน SoD/least privilege/dual control เป็นส่วนหนึ่งของ PCI-DSS v4.0 RoC ประจำปี
- **การละเมิด:** การหลบเลี่ยง SoD/least privilege ถือเป็นการละเมิดนโยบายความปลอดภัยและอาจนำสู่มาตรการทางวินัยและการรายงานต่อผู้กำกับหากเข้าข่าย

---

## 10. ความถี่การทบทวนเอกสาร (Document Review)

ทบทวนเอกสารนี้อย่างน้อย **ปีละครั้ง** และเมื่อ (ก) มีการเปลี่ยนโครงสร้างองค์กร/บทบาท (ข) มีการเปลี่ยน risk appetite (ค) มีเหตุการณ์/ผลตรวจสอบชี้ช่องว่าง (ง) PCI-DSS หรือประกาศ ธปท./ปปง./PDPC ปรับปรุง — เจ้าของเอกสาร: CISO ร่วมกับ CCO อนุมัติโดย Risk Committee

---
---

# Segregation of Duties matrix + least-privilege access control, dual control for key operations (English)

> Supporting document for the **Acquiring Service license application** under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT)**, with paid-up registered capital of **THB 50 million**, alongside **PCI-DSS v4.0 Level 1**.
>
> Document ID: `COMP-18` · Version 0.1 · Owner: Chief Information Security Officer (CISO) with the Chief Compliance Officer (CCO)
> Related: `04-org-chart-governance.md`, `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`
>
> **Disclaimer:** This is a structural reference document, not legal advice. It must be reviewed by BOT-licensing legal counsel and the QSA before submission.

---

> [!IMPORTANT]
> **Assumptions / Open TODOs** — real values must be completed before BOT submission.
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Legal company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **Sponsor / acquiring bank** — not yet signed, so the downstream settlement-instruction role is not finalized (Track B per ROADMAP) | Unresolved | CEO / Head of Partnerships |
> | A3 | **QSA vendor (PCI-DSS v4.0 L1)** — not yet selected; external SoD-matrix reviewer undefined | Unresolved | CISO |
> | A4 | **Actual IAM/IdP and PAM tooling** (IdP + privileged access management + secrets vault) — not procured; product names are placeholders | Unresolved | CISO / Head of Infrastructure |
> | A5 | **Actual HSM/KMS** for dual-control / split-knowledge key ceremonies | Unresolved | CISO |
> | A6 | **Actual headcount at submission** — this matrix assumes the minimum team in ROADMAP §3; if the team is too small and conflicting roles must be combined, a written **compensating control** is mandatory | TODO | CISO / CCO |
> | A7 | **External independent audit firm** (co-source for annual SoD testing) | Unresolved | Audit Committee |
>
> Do not enter fictitious product/person names into the filed version — only verifiable information is permitted.

---

## 1. Purpose & Scope

This document defines the **Segregation of Duties (SoD)** policy, **least-privilege access control**, and **dual control (four-eyes)** requirements for [บริษัท / Company]'s key operations as a Full Acquiring Payment Gateway.

**Objectives:**

1. Prevent any single individual from controlling a sensitive transaction/process end-to-end, which could enable fraud, money laundering, or undetected error.
2. Align with governance and the Three Lines of Defense per BOT notifications on IT Risk and Cyber Resilience (see `04-org-chart-governance.md`).
3. Comply with **PCI-DSS v4.0**, notably Requirement 7 (least privilege / need-to-know), Requirement 8 (identification & authentication), Requirement 3 (dual control & split knowledge for cryptographic keys), and Requirement 10 (logging & accountability).
4. Support **AMLO** requirements (separating the person who acts from the person who approves SAR/STR and sanction screening) and **PDPC/PDPA** (least privilege over personal data).

**Scope:** all production systems, the PCI Cardholder Data Environment (CDE), tokenization vault, ledger/settlement, merchant onboarding, key management, and database/infrastructure access — for all employees, contractors, and service accounts.

---

## 2. Definitions & Principles

| Term | Definition |
|------|------------|
| **Segregation of Duties (SoD)** | No single person controls the *initiate*, *authorize*, *record*, and *reconcile* steps of the same transaction. |
| **Least Privilege** | A user/service account holds only the access strictly needed for its function (need-to-know / need-to-use), and only for as long as needed. |
| **Dual Control (Four-Eyes)** | A key action requires **two separate individuals** to complete (maker–checker); no one can complete all steps alone. |
| **Split Knowledge** | A secret (e.g. a cryptographic key component) is divided across individuals so no one can reconstruct the key alone (PCI Req 3.7). |
| **Maker–Checker** | The person who performs an action (maker) and the person who reviews/approves it (checker) are always different people. |
| **JIT / Break-glass** | Just-In-Time elevated access that auto-expires, with pre-approval and a logged justification. |
| **Toxic Combination** | A pair of privileges/duties that, if held by one person, creates severe fraud risk (see §5). |

**Seven principles:**

1. **Separate the four transaction roles** — initiate / authorize / record / reconcile must span at least two people.
2. **Separate Dev from Prod** — developers must not be able to deploy or alter production data alone (PCI Req 6.5.3, 7).
3. **Separate doers from checkers** — 1st line (operations) is separate from 2nd line (risk/compliance/security) and 3rd line (internal audit).
4. **Deny by default** — no access is implicit; every entitlement is requested, approved, and recertified.
5. **Access has a lifespan** — elevated access is temporary/rotating, never permanent.
6. **Every action is attributable** — no shared/generic accounts; every action maps to a verifiable identity (PCI Req 8.2, 10).
7. **Compensating controls when true separation is impossible** — if the team is too small to split roles, a compensating control approved by the Risk Committee and reviewed quarterly is required.

---

## 3. Roles / RBAC Groups

All access is granted via **Role-Based Access Control (RBAC)** through a central IdP (A4). No per-user grants exist outside the role structure except via time-boxed JIT.

| Role ID | Role | Line of Defense | Summary of access |
|---------|------|-----------------|-------------------|
| R-ENG-DEV | Software Engineer (Developer) | 1st | Write code, access non-prod, read masked logs; **no** prod write |
| R-SRE-OPS | SRE / Platform Operator | 1st | Deploy via pipeline, manage prod infra (break-glass); no direct business-data edits |
| R-DBA | Database Administrator | 1st | Manage schema/backups; prod data read/write only via JIT + dual approval |
| R-PAY-OPS | Payment Operations | 1st | View transactions, initiate refund/void under threshold, handle dispute/chargeback |
| R-SETTLE | Settlement / Treasury Ops | 1st | Initiate payout/settlement batch (maker); **no** payout approval |
| R-SETTLE-APV | Settlement Approver | 1st (senior) | Approve payout/settlement batch (checker); **no** initiation |
| R-MERCH-ONB | Merchant Onboarding Officer | 1st | Create/edit merchant profile (maker), perform KYC/CDD; cannot self-activate |
| R-MERCH-APV | Merchant Approver / MLRO delegate | 2nd | Approve merchant activation and KYC/EDD outcome (checker) |
| R-RISK | Risk / Fraud Analyst | 2nd | Author fraud rules (maker), review alerts; deploying a rule needs a checker |
| R-COMPL | Compliance / AML Officer | 2nd | Sanction-screening review, initiate SAR/STR (maker) |
| R-MLRO | MLRO | 2nd | Approve and file SAR/STR to AMLO (checker); see `07-sar-str-procedure.md` |
| R-DPO | Data Protection Officer | 2nd | Approve out-of-scope personal-data use (PDPA), review DSARs |
| R-SEC | Security Engineer / SecOps | 2nd | Manage WAF/IDS/SIEM, review access; **cannot** approve their own access |
| R-KEY-CUST-A | Key Custodian A | 2nd | Holds key component A (split knowledge) |
| R-KEY-CUST-B | Key Custodian B | 2nd | Holds key component B — always a different person than A |
| R-IAM-ADMIN | Identity & Access Admin | 2nd | Create/disable accounts (maker); granting privileged roles needs a checker |
| R-AUDIT | Internal Auditor | 3rd | Read-only on all logs/evidence; **no** write in operational systems |

> All **service accounts** are treated as distinct identities with an identifiable owner, minimal scope, credentials auto-rotated via a secrets vault, and no interactive login.

---

## 4. Segregation of Duties Matrix

Legend: **M** = Maker (initiate/perform) · **C** = Checker (approve/verify) · **R** = Review only (read-only) · **—** = No access.
Rule: within a process, **Maker and Checker must always be different people** (enforced by IdP/application, not by personal discipline).

| Key process | R-PAY-OPS | R-SETTLE | R-SETTLE-APV | R-MERCH-ONB | R-MERCH-APV | R-RISK | R-COMPL | R-MLRO | R-SRE-OPS | R-DBA | R-IAM-ADMIN | R-SEC | R-AUDIT |
|-------------|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|
| Merchant onboarding / activate | — | — | — | **M** | **C** | R | R | — | — | — | — | — | R |
| Refund / Void ≤ threshold | **M** | — | — | — | — | — | — | — | — | — | — | — | R |
| Refund / Void > threshold | **M** | — | — | — | **C** | — | — | — | — | — | — | — | R |
| Settlement / Payout batch | — | **M** | **C** | — | — | — | — | — | — | — | — | — | R |
| Change fraud/risk rule in prod | R | — | — | — | — | **M** | — | — | **C** | — | — | R | R |
| Sanction screening — clear a hit | — | — | — | — | — | R | **M** | **C** | — | — | — | — | R |
| File SAR/STR to AMLO | — | — | — | — | — | — | **M** | **C** | — | — | — | — | R |
| Deploy code to production | — | — | — | — | — | — | — | — | **M** | — | — | **C** | R |
| Modify production database | — | — | — | — | — | — | — | — | — | **M** | — | **C** | R |
| Grant privileged access | — | — | — | — | — | — | — | — | — | — | **M** | **C** | R |
| Cryptographic key ceremony | — | — | — | — | — | — | — | — | — | — | — | **M/C** | R |
| Change WAF/firewall config (CDE) | — | — | — | — | — | — | — | — | R | — | — | **M** | R |

> Key-ceremony note: requires **Key Custodian A + Key Custodian B** (split knowledge) with Security as witness — no one accesses both key components (see §6.3).

---

## 5. Toxic / Conflicting Combinations

The IdP and application-level checks must **preventively block** the following role pairs on one identity. If combination is unavoidable, an approved compensating control is required (§8).

| # | Prohibited combination | Risk |
|---|------------------------|------|
| T1 | R-SETTLE **+** R-SETTLE-APV | Initiate and approve own payouts → settlement embezzlement |
| T2 | R-MERCH-ONB **+** R-MERCH-APV | Create and approve fake merchants → laundering / self-payout |
| T3 | R-ENG-DEV **+** deploy/prod-DB write | Plant a backdoor and push to prod unreviewed (PCI Req 6/7) |
| T4 | R-COMPL/R-MLRO acting as both maker and checker on the same SAR | Conceal suspicious activity alone (contrary to AMLO intent) |
| T5 | R-IAM-ADMIN approving own access | Unchecked self privilege escalation |
| T6 | R-AUDIT **+** any operational write role | Destroys 3rd-line independence |
| T7 | R-KEY-CUST-A **+** R-KEY-CUST-B | Hold both key components → decrypt PAN alone (PCI Req 3.7) |
| T8 | R-RISK maker **+** checker of the same rule | Disable fraud controls to pass fraudulent traffic |

---

## 6. Dual Control for Key Operations

### 6.1 Thresholds and maker–checker SLAs

| Operation | Threshold requiring dual control | Maker | Checker | Approval SLA |
|-----------|----------------------------------|-------|---------|--------------|
| Refund/Void | > THB 50,000 per item, or > THB 200,000/day/merchant | R-PAY-OPS | R-MERCH-APV (senior ops) | ≤ 4 business hours |
| Settlement/Payout batch | **every batch** (no minimum) | R-SETTLE | R-SETTLE-APV | ≤ 2 hours before cut-off |
| Merchant activation | **every merchant** | R-MERCH-ONB | R-MERCH-APV | per onboarding SLA |
| Deploy to production | **every time** | R-SRE-OPS | R-SEC (or a different senior SRE) | per change window |
| Modify production DB (out of pipeline) | **every time** | R-DBA | R-SEC | JIT ≤ 60 min |
| Grant privileged access | **every time** | R-IAM-ADMIN | Line manager + R-SEC | ≤ 1 business day |
| Key ceremony (generate/rotate) | **every time** | Custodian A | Custodian B + Security witness | per ceremony schedule |
| Change fraud rule in prod | **every time** | R-RISK | R-SRE-OPS / senior risk | ≤ 8 hours |

> These thresholds are **recommended defaults** — they must be aligned with the Risk-Committee-approved risk appetite and reviewed at least annually.

### 6.2 Enforcement mechanisms

- **System-enforced, not discipline-based** — the application/CI-CD/IdP rejects any request where maker = checker.
- Every dual-control approval is written to `audit_log` (append-only, see ARCHITECTURE §6) with both identities, timestamp, justification, and before/after state.
- The checker must confirm with **MFA** at approval time (PCI Req 8.4/8.5).
- No self-approval, no retroactive approval, and no delegation that turns a maker into the checker.

### 6.3 Split knowledge & dual control for cryptographic keys (PCI Req 3.5–3.7)

- Master / Key-Encryption-Keys are generated and rotated inside an **HSM/KMS** (A5) via a key ceremony with **Key Custodians A + B** (different people) and Security as witness.
- No single person accesses all key components (split knowledge); no clear-text key exists outside the HSM.
- Every ceremony has a **script + signed ceremony log** retained as QSA evidence.
- Keys are rotated on the schedule defined in the key-management policy and whenever compromise is suspected.

---

## 7. Least Privilege & Access Lifecycle

### 7.1 Joiner–Mover–Leaver (JML)

| Event | Action | SLA |
|-------|--------|-----|
| **Joiner** | Grant only the position's role (RBAC template) — manager requests, R-IAM-ADMIN maker + checker | by start date |
| **Mover** (role change) | **Revoke all prior access**, then re-grant per the new role (no accumulation/privilege creep) | ≤ 2 business days |
| **Leaver** | Disable/suspend all accounts, revoke tokens/keys, recover devices | **≤ 4 hours** for CDE/privileged; ≤ 24 hours otherwise |

### 7.2 Just-In-Time (JIT) / Break-glass

- Elevated/CDE access is **Just-In-Time** via PAM (A4): request → justify → dual-approve → use → **auto-expire (e.g. ≤ 60 minutes)**.
- Break-glass accounts for emergencies: credentials sealed in the vault; activation triggers an immediate alert to CISO + on-call; a post-incident review is required within 24 hours and credentials are rotated after use.

### 7.3 Access Recertification

| Access type | Review frequency | Reviewer |
|-------------|------------------|----------|
| Privileged / CDE / prod | **quarterly** | Line manager + R-SEC (PCI Req 7.2.4) |
| General (non-privileged) | every 6 months | Line manager |
| Service account & scope | quarterly | Owner + R-SEC |
| SoD toxic-combination scan | monthly (automated) + quarterly review | R-IAM-ADMIN + R-AUDIT |

> Recertification results and revocations are retained as evidence for Internal Audit and the QSA; access dormant for more than 90 days is auto-suspended.

---

## 8. Compensating Controls (when the team is too small to fully separate)

In early phases (Phase 0–2 per ROADMAP), headcount may be insufficient to split every role (A6). If conflicting roles must be combined, a compensating control **approved in writing by the Risk Committee** and reviewed quarterly is required, for example:

- Use a **checker from the 2nd/3rd line** (e.g. CISO/CCO) instead of a teammate when in-team maker–checker separation is not possible.
- Add **detective controls**: daily/sample reviews, automated alerts on any single-person action over a key operation, daily ledger reconciliation.
- **Log the combination** in an SoD exception register with an expiry date and a plan to separate roles as headcount grows.
- Internal Audit tests the effectiveness of compensating controls quarterly and reports to the Audit Committee.

---

## 9. Enforcement, Audit & Reporting

- **System-enforced:** IdP/PAM/application are the enforcement points for SoD and least privilege; document-only policy is insufficient.
- **Logging (PCI Req 10):** every access grant, privileged use, dual-control approval, and break-glass event is written to an append-only `audit_log`, tamper-protected, retained **at least 1 year (with 12 months immediately available)** and archived per `11-data-retention-deletion.md`.
- **Internal audit:** Internal Audit (3rd line) tests the SoD matrix and toxic combinations at least annually and reports to the Audit Committee.
- **External assessment:** the QSA (A3) reviews SoD/least privilege/dual control as part of the annual PCI-DSS v4.0 RoC.
- **Violations:** circumventing SoD/least privilege is a security-policy breach and may lead to disciplinary action and regulatory reporting where applicable.

---

## 10. Document Review

Review this document at least **annually** and whenever (a) organizational/role structures change, (b) risk appetite changes, (c) an incident/audit finding reveals a gap, or (d) PCI-DSS or BOT/AMLO/PDPC notifications are updated. Owner: CISO with the CCO; approved by the Risk Committee.
