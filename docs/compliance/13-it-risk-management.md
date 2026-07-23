# นโยบายบริหารความเสี่ยงด้านเทคโนโลยีสารสนเทศ (ไทย)

> เอกสารเลขที่ **13** ในชุดเอกสาร Compliance สำหรับการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** และคู่ขนานกับ **PCI-DSS Level 1**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อยื่นขออนุมัติ (submission draft) · เวอร์ชัน 0.1 · วันที่ปรับปรุง 2026-07-22
> **เจ้าของเอกสาร:** ประธานเจ้าหน้าที่ด้านความมั่นคงปลอดภัยสารสนเทศ (CISO) และคณะกรรมการบริหารความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk Committee) ของ **[บริษัท / Company]**
>
> เอกสารนี้เป็นนโยบายภายในและเอกสารประกอบคำขอใบอนุญาต **มิใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจทานโดยที่ปรึกษากฎหมาย, CISO และ QSA ก่อนบังคับใช้จริง เนื่องจากประกาศ/หลักเกณฑ์ด้าน IT risk และ cyber resilience ของ ธปท. อาจปรับปรุงได้

---

## บทสรุปสำหรับผู้บริหาร (Executive Summary)

[บริษัท / Company] ในฐานะผู้ให้บริการรับชำระเงินด้วยบัตร (acquiring payment gateway) ที่ประมวลผลข้อมูลบัตร (cardholder data) และดำเนินธุรกรรมทางการเงินแบบเรียลไทม์ ถือว่า **ความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk) และภัยคุกคามทางไซเบอร์ (Cyber Risk) เป็นความเสี่ยงหลักในระดับองค์กร (enterprise-level principal risk)** นโยบายฉบับนี้กำหนดกรอบการกำกับดูแล การประเมิน การควบคุม และการรายงานความเสี่ยงด้าน IT ให้สอดคล้องกับ:

1. **แนวปฏิบัติด้านการบริหารความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk) และการรักษาความมั่นคงปลอดภัยไซเบอร์ (Cyber Resilience) ของ ธปท.**
2. **PCI-DSS v4.0** (12 requirements) สำหรับการคุ้มครองข้อมูลบัตร
3. **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** กำกับโดย **PDPC**
4. **พ.ร.บ. การรักษาความมั่นคงปลอดภัยไซเบอร์ พ.ศ. 2562** (Cybersecurity Act) — เนื่องจากบริการชำระเงินอาจถูกจัดเป็นโครงสร้างพื้นฐานสำคัญทางสารสนเทศ (CII)
5. มาตรฐานสากลที่ใช้อ้างอิง: **ISO/IEC 27001:2022, NIST CSF 2.0, ISO 22301 (BCM)**

นโยบายนี้ครอบคลุมกรอบการทำงาน **Three Lines of Defense**, การประเมินความเสี่ยงเชิงระบบ, การควบคุม (controls) ที่ผูกกับความเสี่ยง, ตัวชี้วัดความเสี่ยงหลัก (KRIs) พร้อมเกณฑ์ (thresholds), และกลไกการรายงานถึงคณะกรรมการบริษัท

> **ความเชื่อมโยงกับสถาปัตยกรรม:** นโยบายนี้อ้างอิงการควบคุมทางเทคนิคที่ได้ออกแบบไว้ในเอกสาร `ARCHITECTURE.md` (tokenization vault แยก scope, ledger แบบ append-only, idempotency, audit_log, mTLS ภายใน, HSM/KMS) — นโยบายนี้เป็นชั้นการกำกับดูแล (governance layer) ที่ครอบการควบคุมเหล่านั้น

---

## 1. ฐานกฎหมายและมาตรฐานอ้างอิง

| กฎหมาย / มาตรฐาน | สาระสำคัญที่เกี่ยวข้องกับนโยบายนี้ |
|---|---|
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** | เงื่อนไขใบอนุญาตกำหนดให้ต้องมีระบบบริหารความเสี่ยงด้าน IT และการควบคุมภายในที่ได้มาตรฐาน กำกับโดย **ธปท.** |
| **แนวนโยบาย/ประกาศ ธปท. ด้าน IT Risk & Cyber Resilience** | การกำกับดูแลโดยคณะกรรมการ, การประเมินความเสี่ยง, การควบคุมการเข้าถึง, การบริหารเหตุการณ์, การทดสอบเจาะระบบ, การบริหารผู้ให้บริการภายนอก |
| **PCI-DSS v4.0** | Req 1-12 การคุ้มครองข้อมูลบัตร (CHD/SAD), network security, encryption, access control, logging, vulnerability & risk management, incident response |
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — **PDPC** | มาตรการรักษาความมั่นคงปลอดภัยของข้อมูลส่วนบุคคล, การแจ้งเหตุละเมิดข้อมูลภายใน 72 ชั่วโมง |
| **พ.ร.บ. การรักษาความมั่นคงปลอดภัยไซเบอร์ พ.ศ. 2562** | การเตรียมความพร้อม รับมือ และรายงานภัยคุกคามไซเบอร์ (กรณีเข้าข่าย CII) |
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (ปปง./AMLO)** | ความเสี่ยงด้าน IT ที่กระทบความสมบูรณ์ของระบบ AML/screening (เชื่อมกับเอกสาร 05) |
| **ISO/IEC 27001:2022 / 27002** | กรอบระบบบริหารความมั่นคงปลอดภัยสารสนเทศ (ISMS) และชุดการควบคุม |
| **NIST Cybersecurity Framework 2.0** | ฟังก์ชัน Govern–Identify–Protect–Detect–Respond–Recover |
| **ISO 22301** | การบริหารความต่อเนื่องทางธุรกิจ (เชื่อมกับเอกสาร BCP/DR) |
| **EMV 3-D Secure (3DS) 2.x** | การควบคุมความเสี่ยงการฉ้อโกงธุรกรรมด้วยการยืนยันตัวตนผู้ถือบัตร |

> **[TODO/ข้อสมมติ]** ต้องยืนยันกับที่ปรึกษากฎหมายและ ธปท. ว่าประกาศ/แนวปฏิบัติด้าน IT risk และ cyber resilience ฉบับล่าสุด ณ วันยื่นคำขอ มีเลขที่/ปีใด และปรับอ้างอิงให้ตรง รวมถึงยืนยันสถานะการเป็น CII ภายใต้ พ.ร.บ. ไซเบอร์ฯ

---

## 2. ขอบเขตและนิยาม (Scope & Definitions)

**ขอบเขต:** ครอบคลุมสินทรัพย์สารสนเทศทั้งหมดของ [บริษัท / Company] ได้แก่ ระบบ payment core (Go/Fiber), tokenization vault, ledger, ฐานข้อมูล, HSM/KMS, โครงสร้างพื้นฐานคลาวด์/on-prem, เครือข่าย, endpoint, ซอฟต์แวร์ที่พัฒนาเอง, ผู้ให้บริการภายนอก (acquirer, sponsor bank, 3DS provider, cloud) และบุคลากรที่เข้าถึงระบบ

| คำ | นิยาม |
|---|---|
| **IT Risk** | ความเสี่ยงที่เกิดจากความล้มเหลว/ไม่เพียงพอของกระบวนการ ระบบ หรือบุคลากรด้าน IT รวมถึงเหตุการณ์ภายนอก ที่กระทบความลับ (Confidentiality) ความถูกต้อง (Integrity) และความพร้อมใช้ (Availability) — CIA |
| **Cyber Risk** | ส่วนย่อยของ IT Risk ที่เกิดจากภัยคุกคามโดยเจตนา (attack) หรือช่องโหว่ที่ถูกใช้ประโยชน์ |
| **Inherent Risk** | ระดับความเสี่ยงก่อนใช้การควบคุม |
| **Residual Risk** | ระดับความเสี่ยงหลังใช้การควบคุม |
| **Risk Appetite** | ระดับความเสี่ยงที่องค์กรยอมรับได้ กำหนดโดยคณะกรรมการบริษัท |
| **KRI** | Key Risk Indicator — ตัวชี้วัดเชิงปริมาณที่ส่งสัญญาณเตือนล่วงหน้าก่อนความเสี่ยงจะกลายเป็นเหตุการณ์จริง |
| **CDE** | Cardholder Data Environment — สภาพแวดล้อมที่จัดเก็บ/ประมวลผล/ส่งข้อมูลบัตร ตาม PCI-DSS |

---

## 3. โครงสร้างการกำกับดูแล (Governance) — Three Lines of Defense

[บริษัท / Company] ใช้แบบจำลอง **Three Lines of Defense**:

| แนวป้องกัน | ผู้รับผิดชอบ | หน้าที่ด้าน IT Risk |
|---|---|---|
| **แนวที่ 1 (Risk Ownership)** | ทีม Engineering, SRE/Infra, DevSecOps, Product | เป็นเจ้าของความเสี่ยง ดำเนินการควบคุมในกระบวนการประจำวัน (secure SDLC, patching, monitoring) |
| **แนวที่ 2 (Risk Oversight)** | **CISO**, IT Risk & Compliance Function, DPO | กำหนดนโยบาย/กรอบ ประเมินอิสระ ติดตาม KRI ท้าทายแนวที่ 1 |
| **แนวที่ 3 (Independent Assurance)** | **Internal Audit** (+ QSA/ผู้ตรวจภายนอก) | ตรวจสอบอิสระต่อคณะกรรมการตรวจสอบ ยืนยันประสิทธิผลของการควบคุม |

### 3.1 คณะกรรมการและบทบาท

| หน่วยงาน | องค์ประกอบ | ความถี่ประชุม | ความรับผิดชอบหลัก |
|---|---|---|---|
| **คณะกรรมการบริษัท (Board)** | กรรมการทั้งหมด | รายไตรมาส | อนุมัติ Risk Appetite, รับรองนโยบาย, รับรายงานความเสี่ยง IT |
| **คณะกรรมการบริหารความเสี่ยง (Risk Committee)** | กรรมการอิสระ + ผู้บริหารระดับสูง | รายไตรมาส (หรือเมื่อมีเหตุ) | กำกับกรอบ ERM/IT Risk, ทบทวน risk register |
| **คณะทำงาน IT Risk & Security (IT Risk Committee)** | CISO (ประธาน), CTO, Head of SRE, DPO, MLRO, Compliance | **รายเดือน** | ทบทวน KRI, เหตุการณ์, ผลสแกน/pentest, การยกเว้นความเสี่ยง (risk acceptance) |
| **คณะกรรมการตรวจสอบ (Audit Committee)** | กรรมการอิสระ | รายไตรมาส | กำกับ Internal Audit, ผลการตรวจ, การแก้ไขข้อบกพร่อง |

> **RACI แบบย่อ:** CISO = **Accountable** สำหรับ IT Risk Framework · Head of Engineering/SRE = **Responsible** สำหรับการควบคุมทางเทคนิค · Risk Committee = **Consulted** · Board = **Informed/Approve**

> **[TODO/ข้อสมมติ]** ต้องแต่งตั้ง CISO และ DPO อย่างเป็นทางการ พร้อมมติคณะกรรมการ และระบุชื่อในเอกสารก่อนยื่น — ปัจจุบันเป็นบทบาทที่ออกแบบไว้ตามโครงสร้างในเอกสาร 04 (org chart)

---

## 4. กระบวนการประเมินความเสี่ยง (Risk Assessment)

### 4.1 วงจรการประเมิน

1. **Identify** — ระบุสินทรัพย์ (asset inventory + data classification) และภัยคุกคาม/ช่องโหว่
2. **Analyze** — ประเมิน **Likelihood × Impact** ได้ inherent risk
3. **Evaluate** — เทียบกับ risk appetite และ threshold
4. **Treat** — เลือกวิธีจัดการ (Mitigate / Transfer / Avoid / Accept)
5. **Monitor & Review** — ติดตามผ่าน KRI และทบทวนตามรอบ

**ความถี่:** ประเมินระดับองค์กร **อย่างน้อยปีละ 1 ครั้ง** และประเมินเฉพาะกิจ (event-driven) เมื่อ: มีการเปลี่ยนแปลงสถาปัตยกรรมสำคัญ, ก่อน go-live ฟีเจอร์ที่ขยับเงิน, มีเหตุการณ์ความมั่นคงปลอดภัยรุนแรง, หรือมีการเปลี่ยนผู้ให้บริการภายนอกใน CDE

### 4.2 เกณฑ์โอกาสเกิด (Likelihood)

| ระดับ | คำอธิบาย | ความถี่โดยประมาณ |
|---|---|---|
| 5 – Almost Certain | คาดว่าเกิดได้ตลอด | > 1 ครั้ง/เดือน |
| 4 – Likely | น่าจะเกิด | 1 ครั้ง/ไตรมาส |
| 3 – Possible | อาจเกิด | 1 ครั้ง/ปี |
| 2 – Unlikely | ไม่น่าเกิด | 1 ครั้ง/1-3 ปี |
| 1 – Rare | เกิดยากมาก | < 1 ครั้ง/3 ปี |

### 4.3 เกณฑ์ผลกระทบ (Impact)

| ระดับ | การเงิน | การดำเนินงาน | กฎเกณฑ์/ชื่อเสียง | ข้อมูล |
|---|---|---|---|---|
| 5 – Critical | > 20 ล้านบาท | หยุดบริการ > 4 ชม. | เพิกถอนใบอนุญาต/PCI, ข่าวเชิงลบวงกว้าง | ข้อมูลบัตร/PII รั่วไหลจำนวนมาก |
| 4 – Major | 5–20 ล้านบาท | หยุด 1–4 ชม. | ธปท./PDPC สั่งการ, ปรับ | รั่วไหลจำกัด |
| 3 – Moderate | 1–5 ล้านบาท | เสื่อมประสิทธิภาพ | ต้องรายงานหน่วยงานกำกับ | เข้าถึงโดยไม่ได้รับอนุญาต (contained) |
| 2 – Minor | 0.1–1 ล้านบาท | กระทบผู้ใช้บางส่วน | ข้อสังเกตภายใน | เกือบเกิด (near miss) |
| 1 – Insignificant | < 0.1 ล้านบาท | แทบไม่กระทบ | ไม่มี | ไม่มี |

### 4.4 Risk Matrix (5×5) และระดับความเสี่ยง

| Likelihood \ Impact | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| **5** | M | H | H | E | E |
| **4** | M | M | H | H | E |
| **3** | L | M | M | H | H |
| **2** | L | L | M | M | H |
| **1** | L | L | L | M | M |

**เกณฑ์การจัดการตามระดับ (Risk Appetite):**

| ระดับ | ชื่อ | การอนุมัติ/จัดการ | SLA แก้ไข |
|---|---|---|---|
| **E** | Extreme | ต้องแจ้ง Board/Risk Committee ทันที, จัดการทันที | ทันที / ≤ 24 ชม. |
| **H** | High | อนุมัติแผนโดย CISO + Risk Committee | ≤ 30 วัน |
| **M** | Medium | เจ้าของความเสี่ยงจัดการ, ติดตามโดย IT Risk Committee | ≤ 90 วัน |
| **L** | Low | ยอมรับได้ ทบทวนตามรอบปกติ | รอบประเมินถัดไป |

> **Risk Appetite Statement:** [บริษัท / Company] มี **zero tolerance** ต่อการจัดเก็บ full PAN/CVV/PIN นอก vault, ต่อการเข้าถึง CDE โดยไม่มี MFA, และต่อการไม่รายงานเหตุละเมิดข้อมูลตามกำหนดเวลาของกฎหมาย

### 4.5 ทะเบียนความเสี่ยง (Risk Register) — ตัวอย่างรายการหลัก

| # | ความเสี่ยง | หมวด | Inherent | การควบคุมหลัก | Residual | เจ้าของ |
|---|---|---|---|---|---|---|
| R-01 | ข้อมูลบัตร (CHD) รั่วไหลจาก CDE | Cyber/Data | E (5×4) | Tokenization vault แยก segment, envelope encryption, HSM, DLP, PCI Req 3/4 | H | CISO |
| R-02 | บัญชี admin ถูกยึด (account takeover) | Access | H (4×4) | MFA, RBAC least-privilege, PAM, session timeout (PCI Req 7/8) | M | Head of SRE |
| R-03 | ระบบ authorization ล่ม (availability) | Operational | H (3×5) | HA multi-AZ, auto-scaling, circuit breaker, RPO≤5m/RTO≤30m | M | Head of SRE |
| R-04 | ช่องโหว่ในโค้ด/dependency (SQLi, RCE) | AppSec | H (4×4) | sqlc parameterized, SAST/DAST, govulncheck, code review, WAF | M | DevSecOps |
| R-05 | Fraud/chargeback เกินเกณฑ์ | Fraud | H (4×3) | 3DS 2.x, velocity check, risk scoring, blacklist | M | Risk/Fraud Lead |
| R-06 | ผู้ให้บริการภายนอก (acquirer/cloud) ขัดข้อง | Third-party | H (3×4) | สัญญา/SLA, exit plan, การประเมิน vendor, multi-region | M | CTO |
| R-07 | Ransomware / มัลแวร์ | Cyber | H (3×5) | EDR, backup แบบ immutable/offline, network segmentation, patching | M | CISO |
| R-08 | Insider threat / การใช้สิทธิ์เกิน | People | M (2×4) | Segregation of duties, dual control, log monitoring, background check | M | HR + CISO |
| R-09 | Data residency / PDPA non-compliance | Compliance | H (3×4) | เก็บข้อมูลในไทย, consent, DPA กับ vendor, breach notification | M | DPO |
| R-10 | Key compromise (HSM/KMS) | Crypto | M (2×5) | Dual control, split knowledge, key rotation, HSM FIPS 140-2/3 | L | CISO |

---

## 5. การควบคุม (Controls) — ผูกกับ PCI-DSS v4.0 และ ธปท.

| โดเมนการควบคุม | มาตรการหลัก | อ้างอิงมาตรฐาน |
|---|---|---|
| **การจำแนกและปกป้องข้อมูล** | Data classification, tokenization, envelope encryption, ห้ามเก็บ SAD หลัง authorization | PCI Req 3, PDPA |
| **การเข้ารหัสข้อมูลระหว่างส่ง** | TLS 1.2+/1.3, mTLS ระหว่าง service ภายใน, ห้าม protocol ที่อ่อนแอ | PCI Req 4 |
| **ความปลอดภัยเครือข่าย** | Network segmentation แยก CDE, firewall/WAF, DDoS protection, IDS/IPS, deny-by-default | PCI Req 1, 11 |
| **การควบคุมการเข้าถึง** | RBAC least-privilege, **MFA บังคับสำหรับทุกการเข้าถึง CDE และ admin**, PAM, JIT access | PCI Req 7, 8 |
| **การบริหารกุญแจเข้ารหัส** | HSM/KMS, dual control, split knowledge, key rotation ตามรอบ, ไม่มีคีย์ใน code/config | PCI Req 3 |
| **Secure SDLC** | Threat modeling, secure coding, peer review, SAST/DAST/SCA ใน CI/CD, ไม่มี secret ใน repo | PCI Req 6 |
| **การบริหารช่องโหว่** | `govulncheck`, dependency scan, **quarterly ASV scan**, **annual penetration test**, patch SLA | PCI Req 6, 11 |
| **Logging & Monitoring** | Structured log (ไม่มี card data), centralized SIEM, alerting, log retention ≥ 1 ปี (3 เดือน online) | PCI Req 10 |
| **การบริหารเหตุการณ์** | Incident Response Plan, on-call, runbook, forensic readiness | PCI Req 12.10 |
| **การควบคุมการเปลี่ยนแปลง** | Change management, approval gate, rollback plan, ห้ามเปลี่ยนแปลง prod โดยไม่ผ่านกระบวนการ | ธปท. / ITIL |
| **การบริหารผู้ให้บริการภายนอก** | Due diligence, DPA/สัญญา, PCI AoC ของ vendor, การติดตามต่อเนื่อง | PCI Req 12.8, ธปท. Outsourcing |
| **ความต่อเนื่องทางธุรกิจ** | BCP/DR, backup immutable, DR drill ≥ ปีละ 1 ครั้ง | ISO 22301 |

### 5.1 การจัดการ Patch / Vulnerability ตามความรุนแรง

| ความรุนแรง (CVSS) | SLA ติดตั้ง patch |
|---|---|
| Critical (9.0–10.0) | ภายใน **72 ชั่วโมง** (emergency change) |
| High (7.0–8.9) | ภายใน **30 วัน** |
| Medium (4.0–6.9) | ภายใน **90 วัน** |
| Low (< 4.0) | ตามรอบ maintenance |

---

## 6. ตัวชี้วัดความเสี่ยงหลัก (Key Risk Indicators — KRIs)

KRI ทุกตัวถูกวัดอัตโนมัติจาก observability stack (Prometheus/SIEM) และรายงานใน dashboard ของ IT Risk Committee รายเดือน โดยใช้ระบบสัญญาณ **เขียว/เหลือง/แดง (RAG)**:

| # | KRI | เขียว (ปกติ) | เหลือง (เฝ้าระวัง) | แดง (เกินเกณฑ์ → escalate) | รอบวัด |
|---|---|---|---|---|---|
| K-01 | Availability ของ payment core | ≥ 99.95% | 99.9–99.95% | < 99.9% | รายเดือน |
| K-02 | Authorization latency (p99) | < 800 ms | 800–1200 ms | > 1200 ms | รายวัน |
| K-03 | จำนวน critical/high vuln ที่เกิน SLA | 0 | 1–2 | ≥ 3 | รายสัปดาห์ |
| K-04 | % ระบบใน CDE ที่ patch เป็นปัจจุบัน | ≥ 99% | 95–99% | < 95% | รายสัปดาห์ |
| K-05 | Failed login / brute-force ต่อ admin | baseline | +50% เหนือ baseline | +100% หรือมี lockout | เรียลไทม์ |
| K-06 | Fraud rate (มูลค่า) | < 0.10% | 0.10–0.20% | > 0.20% | รายวัน |
| K-07 | Chargeback ratio | < 0.65% | 0.65–0.90% | > 0.90% (เกณฑ์ scheme) | รายเดือน |
| K-08 | จำนวน security incident (Sev1/Sev2) | 0 | 1 | ≥ 2 | รายเดือน |
| K-09 | Mean Time To Detect (MTTD) | < 15 นาที | 15–60 นาที | > 60 นาที | ต่อเหตุการณ์ |
| K-10 | Mean Time To Respond/Recover (MTTR) | < 30 นาที | 30–120 นาที | > 120 นาที | ต่อเหตุการณ์ |
| K-11 | % พนักงานผ่าน security awareness training | 100% | 90–99% | < 90% | รายไตรมาส |
| K-12 | Privileged access ที่ไม่มี MFA | 0 | — | ≥ 1 (zero tolerance) | เรียลไทม์ |
| K-13 | Backup/DR restore test สำเร็จ | 100% | — | ล้มเหลว | รายไตรมาส |
| K-14 | Reconciliation mismatch ค้างเกิน T+1 | 0 | 1–5 รายการ | > 5 รายการ | รายวัน |

> เมื่อ KRI เข้าเขต **แดง** ระบบต้องสร้าง alert อัตโนมัติไปยัง on-call และ CISO และเปิด case ติดตามจนปิด (พร้อม root cause)

---

## 7. การบริหารเหตุการณ์ (Incident Management) และการรายงานตามกฎหมาย

### 7.1 การจัดระดับเหตุการณ์

| ระดับ | นิยาม | ตัวอย่าง |
|---|---|---|
| **Sev1 (Critical)** | กระทบบริการทั้งระบบ / ข้อมูลบัตร-PII รั่วไหล | CDE ถูกเจาะ, ระบบล่มทั้งหมด, ransomware |
| **Sev2 (High)** | กระทบบางส่วน / เสี่ยงบานปลาย | latency spike รุนแรง, การเข้าถึงผิดปกติ |
| **Sev3 (Medium)** | กระทบจำกัด | เสื่อมประสิทธิภาพบางฟีเจอร์ |
| **Sev4 (Low)** | ไม่กระทบผู้ใช้ | near miss, ข้อสังเกตจาก log |

### 7.2 กระบวนการ (NIST CSF: Detect → Respond → Recover)

Detect → Triage/จัดระดับ → Contain → Eradicate → Recover → **Post-Incident Review (blameless) ภายใน 5 วันทำการ**

### 7.3 กรอบเวลาการแจ้งเหตุตามกฎหมาย (สำคัญมาก)

| หน่วยงาน | เงื่อนไข | กรอบเวลา |
|---|---|---|
| **PDPC** | เหตุละเมิดข้อมูลส่วนบุคคลที่เสี่ยงต่อสิทธิเสรีภาพ | **ภายใน 72 ชั่วโมง** นับแต่ทราบเหตุ (ตาม PDPA) |
| **เจ้าของข้อมูล (Data Subject)** | กรณีเสี่ยงสูงต่อบุคคล | โดยไม่ชักช้า |
| **ธปท.** | เหตุการณ์ IT/cyber ที่กระทบบริการชำระเงินอย่างมีนัยสำคัญ | ตามกรอบเวลาที่ประกาศ ธปท. กำหนด (**[TODO/ยืนยันชั่วโมงที่แน่นอน]**) |
| **card scheme / sponsor bank** | สงสัยข้อมูลบัตรรั่ว (suspected account data compromise) | ตามกฎ Visa/Mastercard (โดยทันที) |
| **ปปง./AMLO** | หากเหตุกระทบความสมบูรณ์ของธุรกรรม/ระบบ AML | ตามที่กฎหมายกำหนด |

> **[TODO/ข้อสมมติ]** ต้องยืนยันกรอบเวลาการรายงานเหตุ IT/cyber ต่อ ธปท. และช่องทางที่แน่นอน (จำนวนชั่วโมง/แบบฟอร์ม) กับที่ปรึกษากฎหมายก่อนยื่น

---

## 8. การรายงาน (Reporting Cadence)

| รายงาน | ผู้จัดทำ | ผู้รับ | ความถี่ |
|---|---|---|---|
| IT Risk & Security Dashboard (KRI RAG) | CISO | IT Risk Committee | **รายเดือน** |
| รายงานความเสี่ยง IT + risk register | CISO | Risk Committee | **รายไตรมาส** |
| รายงานต่อคณะกรรมการบริษัท | Risk Committee | Board | **รายไตรมาส** |
| Incident report (Sev1/Sev2) | Incident Commander | CISO → Board (Sev1) | ต่อเหตุการณ์ + สรุปรายเดือน |
| ผลสแกน ASV | ASV vendor | CISO | **รายไตรมาส** |
| ผล penetration test | ผู้ตรวจ + CISO | Risk/Audit Committee | **รายปี** (+ เมื่อเปลี่ยนแปลงสำคัญ) |
| PCI-DSS RoC / AoC | **QSA** | ธปท. / scheme / Board | **รายปี** |
| รายงานเชิงกำกับดูแลต่อ ธปท. | Compliance + CISO | ธปท. | ตามงวดที่ ธปท. กำหนด |

---

## 9. การทบทวนนโยบายและข้อสมมติที่ยังไม่ปิด

- **รอบทบทวน:** นโยบายนี้ทบทวน **อย่างน้อยปีละ 1 ครั้ง** และเมื่อมีการเปลี่ยนแปลงกฎเกณฑ์/สถาปัตยกรรมสำคัญ อนุมัติโดยคณะกรรมการบริษัท

### สรุปข้อสมมติ / TODO ที่ต้องปิดก่อนยื่นจริง

| # | ประเด็นที่ยังไม่ resolve | ผู้รับผิดชอบปิด |
|---|---|---|
| A-1 | **Sponsor bank / acquirer** ยังไม่ลงนาม → ยังไม่มี SLA/exit terms จริง สำหรับ R-06 | CTO / Legal |
| A-2 | **QSA vendor** ยังไม่ว่าจ้าง → กำหนดการ RoC และ scope validation ยังเป็นประมาณการ | CISO |
| A-3 | **ASV vendor** และผู้ให้บริการ pentest ยังไม่เลือก | CISO |
| A-4 | **ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท** ต้องยืนยันชำระจริงและคงไว้ ≥ 75% (เชื่อมเอกสาร 02) | CFO |
| A-5 | การแต่งตั้ง **CISO/DPO** อย่างเป็นทางการพร้อมมติคณะกรรมการ | Board / HR |
| A-6 | กรอบเวลารายงานเหตุ cyber ต่อ **ธปท.** และสถานะ **CII** ภายใต้ พ.ร.บ. ไซเบอร์ฯ | Legal |

---

# IT Risk Management policy per BOT IT-risk guidelines: risk assessment, controls, KRIs, reporting (English)

> Document **#13** in the Compliance document set for the license application for **Acquiring Payment Service (Full Acquiring)** under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT)**, in parallel with **PCI-DSS Level 1**.
>
> **Status:** submission draft · v0.1 · updated 2026-07-22
> **Owner:** Chief Information Security Officer (CISO) and the IT Risk Committee of **[บริษัท / Company]**
>
> This is an internal policy and a license-application supporting document. **It is not legal advice** — it must be reviewed by legal counsel, the CISO, and the QSA before it takes effect, as BOT's IT-risk and cyber-resilience notifications may be revised.

---

## Executive Summary

As an acquiring payment gateway that processes cardholder data and executes real-time financial transactions, **[บริษัท / Company]** treats **IT Risk and Cyber Risk as an enterprise-level principal risk**. This policy establishes the governance, assessment, control, and reporting framework for IT risk, aligned with:

1. **BOT IT-risk management and cyber-resilience guidelines**
2. **PCI-DSS v4.0** (12 requirements) for cardholder data protection
3. **Personal Data Protection Act B.E. 2562 (PDPA)**, supervised by the **PDPC**
4. **Cybersecurity Act B.E. 2562** — as payment services may be designated Critical Information Infrastructure (CII)
5. International reference standards: **ISO/IEC 27001:2022, NIST CSF 2.0, ISO 22301 (BCM)**

It covers the **Three Lines of Defense** model, systematic risk assessment, risk-linked controls, Key Risk Indicators (KRIs) with defined thresholds, and reporting to the Board.

> **Architecture linkage:** This policy governs the technical controls already designed in `ARCHITECTURE.md` (segregated tokenization vault, append-only ledger, idempotency, audit_log, internal mTLS, HSM/KMS). It is the governance layer wrapping those controls.

---

## 1. Legal & Standards Basis

| Law / Standard | Relevance to this policy |
|---|---|
| **Payment Systems Act B.E. 2560** | License conditions require a standards-grade IT risk-management and internal-control system; supervised by **BOT** |
| **BOT IT-risk & cyber-resilience notifications** | Board oversight, risk assessment, access control, incident management, penetration testing, third-party management |
| **PCI-DSS v4.0** | Req 1-12: CHD/SAD protection, network security, encryption, access control, logging, vulnerability & risk management, incident response |
| **PDPA B.E. 2562** — **PDPC** | Security measures for personal data; breach notification within 72 hours |
| **Cybersecurity Act B.E. 2562** | Preparedness, response, and reporting of cyber threats (if designated CII) |
| **AML Act (AMLO)** | IT risks affecting integrity of AML/screening systems (links to Doc 05) |
| **ISO/IEC 27001:2022 / 27002** | ISMS framework and control set |
| **NIST CSF 2.0** | Govern–Identify–Protect–Detect–Respond–Recover functions |
| **ISO 22301** | Business continuity management (links to BCP/DR doc) |
| **EMV 3-D Secure (3DS) 2.x** | Fraud risk control via cardholder authentication |

> **[TODO/ASSUMPTION]** Confirm with legal counsel and BOT the exact number/year of the latest IT-risk and cyber-resilience notifications as of filing, and confirm CII status under the Cybersecurity Act.

---

## 2. Scope & Definitions

**Scope:** All information assets of [บริษัท / Company] — the payment core (Go/Fiber), tokenization vault, ledger, databases, HSM/KMS, cloud/on-prem infrastructure, network, endpoints, in-house software, third parties (acquirer, sponsor bank, 3DS provider, cloud), and personnel with system access.

| Term | Definition |
|---|---|
| **IT Risk** | Risk from inadequate/failed IT processes, systems, or people, plus external events, affecting Confidentiality, Integrity, Availability (CIA) |
| **Cyber Risk** | Subset of IT Risk arising from intentional threats or exploited vulnerabilities |
| **Inherent Risk** | Risk level before controls |
| **Residual Risk** | Risk level after controls |
| **Risk Appetite** | Risk level the organization accepts, set by the Board |
| **KRI** | Key Risk Indicator — quantitative early-warning metric |
| **CDE** | Cardholder Data Environment (per PCI-DSS) |

---

## 3. Governance — Three Lines of Defense

| Line | Owner | IT Risk role |
|---|---|---|
| **1st (Risk Ownership)** | Engineering, SRE/Infra, DevSecOps, Product | Own risks; operate day-to-day controls (secure SDLC, patching, monitoring) |
| **2nd (Risk Oversight)** | **CISO**, IT Risk & Compliance, DPO | Set policy/framework, independent assessment, KRI monitoring, challenge Line 1 |
| **3rd (Independent Assurance)** | **Internal Audit** (+ QSA/external) | Independent assurance to the Audit Committee |

### 3.1 Committees & Roles

| Body | Composition | Cadence | Core responsibility |
|---|---|---|---|
| **Board** | All directors | Quarterly | Approve Risk Appetite, endorse policy, receive IT risk reports |
| **Risk Committee** | Independent directors + senior mgmt | Quarterly (or ad hoc) | Oversee ERM/IT Risk framework, review risk register |
| **IT Risk & Security Committee** | CISO (chair), CTO, Head of SRE, DPO, MLRO, Compliance | **Monthly** | Review KRIs, incidents, scan/pentest results, risk acceptances |
| **Audit Committee** | Independent directors | Quarterly | Oversee Internal Audit, findings, remediation |

> **RACI (condensed):** CISO = **Accountable** for the IT Risk Framework · Head of Engineering/SRE = **Responsible** for technical controls · Risk Committee = **Consulted** · Board = **Informed/Approve**

> **[TODO/ASSUMPTION]** CISO and DPO must be formally appointed by board resolution and named before filing; currently designed per Doc 04 (org chart).

---

## 4. Risk Assessment Process

### 4.1 Cycle
1. **Identify** — asset inventory + data classification; threats/vulnerabilities
2. **Analyze** — Likelihood × Impact → inherent risk
3. **Evaluate** — compare against appetite/thresholds
4. **Treat** — Mitigate / Transfer / Avoid / Accept
5. **Monitor & Review** — via KRIs and scheduled reviews

**Frequency:** Enterprise-wide **at least annually**, plus event-driven assessments on major architecture changes, before go-live of money-moving features, after a severe security event, or on changing a third party in the CDE.

### 4.2 Likelihood

| Level | Description | Approx. frequency |
|---|---|---|
| 5 – Almost Certain | Expected constantly | > 1/month |
| 4 – Likely | Probably occurs | 1/quarter |
| 3 – Possible | May occur | 1/year |
| 2 – Unlikely | Improbable | 1/1–3 years |
| 1 – Rare | Very unlikely | < 1/3 years |

### 4.3 Impact

| Level | Financial | Operational | Regulatory/Reputation | Data |
|---|---|---|---|---|
| 5 – Critical | > THB 20M | Outage > 4h | License/PCI revocation, wide negative press | Large CHD/PII breach |
| 4 – Major | THB 5–20M | Outage 1–4h | BOT/PDPC action, fines | Limited breach |
| 3 – Moderate | THB 1–5M | Degradation | Reportable to regulator | Contained unauthorized access |
| 2 – Minor | THB 0.1–1M | Some users affected | Internal finding | Near miss |
| 1 – Insignificant | < THB 0.1M | Negligible | None | None |

### 4.4 Risk Matrix (5×5)

| Likelihood \ Impact | 1 | 2 | 3 | 4 | 5 |
|---|---|---|---|---|---|
| **5** | M | H | H | E | E |
| **4** | M | M | H | H | E |
| **3** | L | M | M | H | H |
| **2** | L | L | M | M | H |
| **1** | L | L | L | M | M |

**Treatment by level (Risk Appetite):**

| Level | Name | Approval / treatment | Remediation SLA |
|---|---|---|---|
| **E** | Extreme | Notify Board/Risk Committee immediately; treat now | Immediate / ≤ 24h |
| **H** | High | Plan approved by CISO + Risk Committee | ≤ 30 days |
| **M** | Medium | Risk owner treats; tracked by IT Risk Committee | ≤ 90 days |
| **L** | Low | Acceptable; review at next cycle | Next assessment |

> **Risk Appetite Statement:** [บริษัท / Company] holds **zero tolerance** for storing full PAN/CVV/PIN outside the vault, for CDE access without MFA, and for failing to report a data breach within statutory deadlines.

### 4.5 Risk Register (key entries)

| # | Risk | Category | Inherent | Key controls | Residual | Owner |
|---|---|---|---|---|---|---|
| R-01 | CHD leakage from CDE | Cyber/Data | E (5×4) | Segregated tokenization vault, envelope encryption, HSM, DLP, PCI Req 3/4 | H | CISO |
| R-02 | Admin account takeover | Access | H (4×4) | MFA, least-privilege RBAC, PAM, session timeout (PCI Req 7/8) | M | Head of SRE |
| R-03 | Authorization outage (availability) | Operational | H (3×5) | HA multi-AZ, auto-scaling, circuit breaker, RPO≤5m/RTO≤30m | M | Head of SRE |
| R-04 | Code/dependency vulns (SQLi, RCE) | AppSec | H (4×4) | sqlc parameterized, SAST/DAST, govulncheck, code review, WAF | M | DevSecOps |
| R-05 | Fraud/chargeback over threshold | Fraud | H (4×3) | 3DS 2.x, velocity checks, risk scoring, blacklist | M | Risk/Fraud Lead |
| R-06 | Third-party (acquirer/cloud) outage | Third-party | H (3×4) | Contracts/SLA, exit plan, vendor assessment, multi-region | M | CTO |
| R-07 | Ransomware / malware | Cyber | H (3×5) | EDR, immutable/offline backups, segmentation, patching | M | CISO |
| R-08 | Insider threat / privilege misuse | People | M (2×4) | Segregation of duties, dual control, log monitoring, background checks | M | HR + CISO |
| R-09 | Data residency / PDPA non-compliance | Compliance | H (3×4) | In-Thailand data storage, consent, vendor DPA, breach notification | M | DPO |
| R-10 | Key compromise (HSM/KMS) | Crypto | M (2×5) | Dual control, split knowledge, key rotation, FIPS 140-2/3 HSM | L | CISO |

---

## 5. Controls — mapped to PCI-DSS v4.0 and BOT

| Control domain | Key measures | Standard ref. |
|---|---|---|
| **Data classification & protection** | Classification, tokenization, envelope encryption, no SAD retention post-auth | PCI Req 3, PDPA |
| **Encryption in transit** | TLS 1.2+/1.3, internal mTLS, no weak protocols | PCI Req 4 |
| **Network security** | CDE segmentation, firewall/WAF, DDoS protection, IDS/IPS, deny-by-default | PCI Req 1, 11 |
| **Access control** | Least-privilege RBAC, **MFA mandatory for all CDE and admin access**, PAM, JIT access | PCI Req 7, 8 |
| **Key management** | HSM/KMS, dual control, split knowledge, key rotation, no keys in code/config | PCI Req 3 |
| **Secure SDLC** | Threat modeling, secure coding, peer review, SAST/DAST/SCA in CI/CD, no secrets in repo | PCI Req 6 |
| **Vulnerability management** | govulncheck, dependency scan, **quarterly ASV scan**, **annual penetration test**, patch SLA | PCI Req 6, 11 |
| **Logging & monitoring** | Structured logs (no card data), centralized SIEM, alerting, retention ≥ 1 year (3 months online) | PCI Req 10 |
| **Incident management** | IR plan, on-call, runbooks, forensic readiness | PCI Req 12.10 |
| **Change management** | Approval gates, rollback plans, no unapproved prod changes | BOT / ITIL |
| **Third-party management** | Due diligence, DPA/contracts, vendor PCI AoC, ongoing monitoring | PCI Req 12.8, BOT Outsourcing |
| **Business continuity** | BCP/DR, immutable backups, DR drill ≥ annually | ISO 22301 |

### 5.1 Patch / Vulnerability SLA by severity

| Severity (CVSS) | Patch SLA |
|---|---|
| Critical (9.0–10.0) | Within **72 hours** (emergency change) |
| High (7.0–8.9) | Within **30 days** |
| Medium (4.0–6.9) | Within **90 days** |
| Low (< 4.0) | Next maintenance cycle |

---

## 6. Key Risk Indicators (KRIs)

All KRIs are measured automatically from the observability stack (Prometheus/SIEM) and reported on the IT Risk Committee's monthly dashboard using a **Red/Amber/Green (RAG)** scheme:

| # | KRI | Green | Amber | Red (escalate) | Cadence |
|---|---|---|---|---|---|
| K-01 | Payment core availability | ≥ 99.95% | 99.9–99.95% | < 99.9% | Monthly |
| K-02 | Authorization latency (p99) | < 800 ms | 800–1200 ms | > 1200 ms | Daily |
| K-03 | Critical/high vulns past SLA | 0 | 1–2 | ≥ 3 | Weekly |
| K-04 | % CDE systems fully patched | ≥ 99% | 95–99% | < 95% | Weekly |
| K-05 | Failed/brute-force admin logins | baseline | +50% over baseline | +100% or lockout | Real-time |
| K-06 | Fraud rate (value) | < 0.10% | 0.10–0.20% | > 0.20% | Daily |
| K-07 | Chargeback ratio | < 0.65% | 0.65–0.90% | > 0.90% (scheme threshold) | Monthly |
| K-08 | Security incidents (Sev1/Sev2) | 0 | 1 | ≥ 2 | Monthly |
| K-09 | Mean Time To Detect (MTTD) | < 15 min | 15–60 min | > 60 min | Per event |
| K-10 | Mean Time To Respond/Recover (MTTR) | < 30 min | 30–120 min | > 120 min | Per event |
| K-11 | % staff completing security awareness training | 100% | 90–99% | < 90% | Quarterly |
| K-12 | Privileged access without MFA | 0 | — | ≥ 1 (zero tolerance) | Real-time |
| K-13 | Backup/DR restore test success | 100% | — | Failed | Quarterly |
| K-14 | Reconciliation mismatches aged > T+1 | 0 | 1–5 | > 5 | Daily |

> On a **Red** breach, the system must auto-alert on-call and the CISO and open a tracked case until closure with root cause.

---

## 7. Incident Management & Statutory Reporting

### 7.1 Severity classification

| Level | Definition | Example |
|---|---|---|
| **Sev1 (Critical)** | System-wide impact / CHD-PII breach | CDE breach, full outage, ransomware |
| **Sev2 (High)** | Partial impact / escalation risk | Severe latency spike, anomalous access |
| **Sev3 (Medium)** | Limited impact | Feature degradation |
| **Sev4 (Low)** | No user impact | Near miss, log finding |

### 7.2 Process (NIST CSF: Detect → Respond → Recover)

Detect → Triage/classify → Contain → Eradicate → Recover → **Blameless post-incident review within 5 business days**

### 7.3 Statutory reporting timelines (critical)

| Authority | Trigger | Timeline |
|---|---|---|
| **PDPC** | Personal-data breach risking rights/freedoms | **Within 72 hours** of becoming aware (PDPA) |
| **Data Subject** | High risk to individuals | Without delay |
| **BOT** | Significant IT/cyber event impacting payment service | Per BOT-defined timeline (**[TODO/confirm exact hours]**) |
| **Card scheme / sponsor bank** | Suspected account data compromise | Per Visa/Mastercard rules (immediately) |
| **AMLO** | Event affecting transaction/AML system integrity | As required by law |

> **[TODO/ASSUMPTION]** Confirm the exact BOT reporting timeline and channel for IT/cyber events (hours/form) with legal counsel before filing.

---

## 8. Reporting Cadence

| Report | Prepared by | Recipient | Frequency |
|---|---|---|---|
| IT Risk & Security Dashboard (KRI RAG) | CISO | IT Risk Committee | **Monthly** |
| IT risk report + risk register | CISO | Risk Committee | **Quarterly** |
| Board report | Risk Committee | Board | **Quarterly** |
| Incident report (Sev1/Sev2) | Incident Commander | CISO → Board (Sev1) | Per event + monthly summary |
| ASV scan results | ASV vendor | CISO | **Quarterly** |
| Penetration test results | Tester + CISO | Risk/Audit Committee | **Annually** (+ on major change) |
| PCI-DSS RoC / AoC | **QSA** | BOT / scheme / Board | **Annually** |
| Regulatory reporting to BOT | Compliance + CISO | BOT | Per BOT-defined periods |

---

## 9. Policy Review & Open Assumptions

- **Review cycle:** Reviewed **at least annually** and upon material regulatory/architecture change; approved by the Board.

### Open assumptions / TODOs to close before filing

| # | Unresolved item | Owner to close |
|---|---|---|
| A-1 | **Sponsor bank / acquirer** not yet signed → no real SLA/exit terms for R-06 | CTO / Legal |
| A-2 | **QSA vendor** not engaged → RoC schedule and scope validation are estimates | CISO |
| A-3 | **ASV vendor** and pentest provider not selected | CISO |
| A-4 | **Paid-up capital THB 50M** must be confirmed as actually paid and maintained ≥ 75% (links to Doc 02) | CFO |
| A-5 | Formal **CISO/DPO** appointment by board resolution | Board / HR |
| A-6 | BOT cyber-incident reporting timeline and **CII** status under the Cybersecurity Act | Legal |
