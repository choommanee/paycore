# การบริหารความเสี่ยงผู้ให้บริการภายนอก (Outsourcing) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 · ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท · PCI-DSS v4.0 Level 1
> เอกสารเลขที่ **30 — การบริหารความเสี่ยงผู้ให้บริการภายนอก: บัญชีผู้ให้บริการสำคัญ + การประเมินความเสี่ยง + ข้อสัญญา SLA/สิทธิตรวจสอบ ตามหลักเกณฑ์ Outsourcing ของ ธปท.**
> เจ้าของเอกสาร: **[บริษัท / Company]** · เวอร์ชัน 1.0 · จัดทำ 2026-07-22 · ทบทวนทุก 12 เดือน
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงเทคนิค/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย ต้องผ่านการตรวจ
> โดยที่ปรึกษากฎหมายและ QSA ก่อนยื่นจริง เอกสารอ้างอิงร่วม: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`,
> `13-it-risk-management.md`, `14-cyber-resilience.md`, `15-bcp-dr.md`, `18-segregation-of-duties.md`, `21-tokenization-hsm-keymgmt.md`

---

## 1. วัตถุประสงค์และขอบเขต

นโยบายนี้กำหนดวิธีที่ [บริษัท / Company] **คัดเลือก ประเมิน ทำสัญญา กำกับ ติดตาม และยุติ** การใช้ผู้ให้บริการภายนอก
(outsourcing / third-party service providers) โดยเฉพาะผู้ให้บริการที่รองรับ **งานสำคัญ (critical / material outsourcing)**
ของธุรกิจรับชำระเงินด้วยบัตร โดยยึดหลัก **"การมอบหมายงานให้ผู้อื่นทำ ไม่ได้เป็นการมอบความรับผิดชอบตามกฎหมายไปด้วย"
([บริษัท / Company] ยังคงรับผิดชอบเต็มต่อ ธปท. และลูกค้าเสมอ)**

นโยบายนี้ตอบข้อกำหนดหลักดังนี้:

- **ประกาศ ธปท. ว่าด้วยการใช้บริการจากผู้ให้บริการภายนอก (Outsourcing)** สำหรับสถาบันการเงิน/ผู้ประกอบธุรกิจ
  ระบบและบริการการชำระเงินภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 — หลักการ due diligence ก่อนใช้บริการ,
  สัญญาที่ครอบคลุมความเสี่ยง, สิทธิเข้าตรวจสอบ (audit right) ของบริษัทและของ ธปท., แผนสำรอง/exit plan,
  และการกำกับผู้ให้บริการช่วง (sub-outsourcing)
- **ประกาศ ธปท. ด้าน IT risk / cyber resilience** — การจัดการความเสี่ยงด้านเทคโนโลยีสารสนเทศจากบุคคลภายนอก
  และ concentration risk / cloud
- **PCI-DSS v4.0 Requirement 12.8 (12.8.1–12.8.5)** — การบริหาร Third-Party Service Providers (TPSP):
  รายการ TPSP, การระบุว่าใครรับผิดชอบ PCI requirement ข้อใด (responsibility matrix), การตรวจสอบสถานะ PCI
  ของผู้ให้บริการอย่างน้อยปีละครั้ง, และ due diligence ก่อนผูกพัน
- **PCI-DSS v4.0 Requirement 12.9** — ข้อผูกพันฝั่ง TPSP ที่ต้องยอมรับความรับผิดชอบต่อ CHD ที่ตนถือ/กระทบ
- **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA) / PDPC** — เมื่อผู้ให้บริการเป็น **ผู้ประมวลผลข้อมูล (Data Processor)**
  ต้องมี **Data Processing Agreement (DPA)** ตาม ม.40 (ดู `10-dpa-templates.md`) และควบคุมการโอนข้อมูลข้ามพรมแดน (ม.28–29)
- **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (ปปง./AMLO)** — ผู้ให้บริการที่ทำงาน KYC/CDD/sanction screening แทน
  ต้องอยู่ใต้การกำกับและตรวจสอบได้ ([บริษัท / Company] ยังรับผิดชอบการรายงาน STR/SAR เอง)

> **ขอบเขต:** ครอบคลุมผู้ให้บริการทุกรายที่ (ก) เข้าถึง/ประมวลผล/จัดเก็บ Cardholder Data (CHD) หรือข้อมูลส่วนบุคคล,
> (ข) รองรับฟังก์ชันที่กระทบความต่อเนื่องของบริการชำระเงิน, หรือ (ค) เชื่อมต่อเข้ากับ Cardholder Data Environment (CDE)
> ตาม `ARCHITECTURE.md` การซื้อสินค้า/บริการทั่วไปที่ไม่กระทบข้อ (ก)–(ค) ใช้กระบวนการจัดซื้อมาตรฐาน ไม่อยู่ในนโยบายนี้

---

## 2. บทบาทและความรับผิดชอบ (Roles & Responsibilities)

| บทบาท | ความรับผิดชอบด้าน vendor risk |
|-------|-------------------------------|
| **Business Owner (เจ้าของงาน)** | ระบุความจำเป็นทางธุรกิจ, นิยาม service scope/SLA, เป็นเจ้าของความสัมพันธ์รายวัน |
| **Vendor Risk Manager / Third-Party Risk (TPRM)** | เจ้าของทะเบียนผู้ให้บริการ, จัด due diligence, ให้คะแนนความเสี่ยง, ติดตามรอบทบทวน |
| **CISO / DevSecOps** | ประเมินความเสี่ยงด้านความปลอดภัย/CDE, ตรวจ PCI AoC ของผู้ให้บริการ, อนุมัติการเชื่อมต่อทางเทคนิค |
| **DPO** | ตรวจว่าเป็น Data Processor หรือไม่, จัดทำ/ตรวจ DPA, ประเมินการโอนข้อมูลข้ามพรมแดน (PDPA) |
| **Compliance / AML Officer** | ตรวจผู้ให้บริการงาน AML/KYC/screening, sanction check ตัวผู้ให้บริการเอง |
| **Legal** | ร่าง/ตรวจสัญญา, ฝัง audit-right / exit / sub-outsourcing / liability clauses |
| **Procurement** | กระบวนการจัดซื้อ, ตรวจฐานะการเงินผู้ให้บริการ, จัดการ PO/สัญญา |
| **Outsourcing Risk Committee (ORC)** | อนุมัติการใช้ผู้ให้บริการระดับ **Critical/Material**, ทบทวนความเสี่ยงรวม (concentration) |
| **Internal Audit** | ตรวจอิสระว่ากระบวนการนี้ถูกปฏิบัติจริง, ทดสอบตัวอย่างสัญญา/การประเมิน |
| **Board / คณะกรรมการ** | อนุมัตินโยบาย outsourcing, รับทราบความเสี่ยง critical outsourcing รายปี |

> **หลัก Separation of Duties (SoD):** ผู้ที่คัดเลือก/เจรจากับผู้ให้บริการ (Business Owner/Procurement) **ต้องไม่ใช่**
> ผู้อนุมัติความเสี่ยงขั้นสุดท้าย (ORC) การอนุมัติ Critical outsourcing ต้องมี CISO + DPO + Compliance ลงนามร่วม

---

## 3. การจัดชั้นความสำคัญของผู้ให้บริการ (Vendor Criticality Tiering)

| ชั้น | นิยาม | ตัวอย่าง | การกำกับ |
|------|-------|---------|----------|
| **Tier 1 — Critical / Material** | หากล้มเหลว/รั่วไหล กระทบบริการชำระเงินหรือ CHD อย่างมีนัยสำคัญ หรือเข้าถึง CDE | Sponsor bank/acquirer, cloud (IaaS), HSM/KMS, tokenization vault, 3DS server (ACS/DS), KYC/screening | Due diligence เต็ม + อนุมัติ ORC + สัญญามี audit-right + exit plan + on-site ได้ + ทบทวนทุก 12 เดือน |
| **Tier 2 — Important** | กระทบการดำเนินงานแต่มีทางเลี่ยง/สำรอง หรือถือข้อมูลส่วนบุคคลจำกัด | Email/notification (SMS/LINE), monitoring/SIEM SaaS, log storage, ASV vendor | Due diligence มาตรฐาน + อนุมัติ CISO/DPO + DPA + ทบทวนทุก 12–24 เดือน |
| **Tier 3 — Low** | ไม่แตะ CHD/PII สำคัญ กระทบต่ำ | เครื่องมือ productivity, ที่ปรึกษาไม่แตะข้อมูลลูกค้า | Due diligence แบบเบา + ทบทวนตามสัญญา |

**เกณฑ์ Threshold ที่ทำให้เป็น Tier 1 (เข้าเงื่อนไขข้อใดข้อหนึ่งถือเป็น Critical):**
- เข้าถึง/จัดเก็บ/ส่งผ่าน **CHD หรือ authentication data** หรืออยู่ใน/ติดกับ **CDE**
- รองรับฟังก์ชันที่ downtime > **30 นาที** กระทบ authorization/settlement (สอดคล้อง RTO ใน `ARCHITECTURE.md`)
- ประมวลผลข้อมูลส่วนบุคคลของผู้ถือบัตร > **10,000 ราย** หรือมีการโอนข้อมูลออกนอกประเทศ
- ทำหน้าที่ตามกฎหมายแทนบริษัท (KYC/CDD, sanction screening, การรายงาน)

---

## 4. บัญชีผู้ให้บริการสำคัญ (Critical Service Provider List / TPSP Inventory — PCI Req 12.8.1)

> ตารางนี้เป็น **living document** ปรับปรุงเมื่อมีการเพิ่ม/เปลี่ยนผู้ให้บริการ และทบทวนอย่างน้อยปีละครั้ง
> ช่อง "PCI responsibility" อ้างอิง responsibility matrix (ข้อ 8)

| # | ประเภทบริการ | ผู้ให้บริการ (ตัวอย่าง/ที่วางแผน) | Tier | เข้าถึง CHD? | Data Processor (PDPA)? | ในประเทศ/ต่างประเทศ | หลักฐานที่ต้องเก็บ | รอบทบทวน |
|---|---------------|-----------------------------------|------|-------------|------------------------|----------------------|--------------------|-----------|
| 1 | Sponsor bank / Acquirer / Card switch | **TODO — ยังไม่สรุป (ดู callout §12)** | 1 | ทางอ้อม (routing) | ไม่ (controller-to-controller) | ในประเทศ | สัญญา acquiring, scheme cert | 12 เดือน |
| 2 | Cloud / IaaS (compute, DB, network) | เช่น AWS/GCP ap-southeast (region ในไทยเมื่อพร้อม) — **TODO ยืนยัน data residency** | 1 | มี (host CDE) | ใช่ (sub-processor) | **ต้องยืนยันว่าเก็บในไทย** (ARCHITECTURE §8) | PCI AoC (SP), SOC 2 Type II, ISO 27001, DPA | 12 เดือน |
| 3 | HSM / KMS | เช่น cloud HSM หรือ on-prem PCI-PTS HSM — **TODO** | 1 | มี (คีย์เข้ารหัส PAN) | ไม่ (ถือคีย์ ไม่ถือ PII) | ในประเทศ | FIPS 140-2/3 Lvl 3+, PCI AoC | 12 เดือน |
| 4 | Tokenization Vault (ถ้าใช้ external) | **TODO — ตัดสินใจ build vs buy** (ARCHITECTURE §4) | 1 | มี (PAN) | ใช่ | ต้องในไทย | PCI AoC (SP), token scheme | 12 เดือน |
| 5 | 3-D Secure (ACS/DS/3DS Server) | ผู้ให้บริการ EMV 3DS v2.x — **TODO** | 1 | authentication data | ใช่ | ต้องยืนยัน | EMVCo approval, PCI 3DS SDK/Core AoC | 12 เดือน |
| 6 | KYC / CDD / e-KYC + Sanction screening | ผู้ให้บริการ screening (เทียบ UN/OFAC/ปปง.) — **TODO** | 1 | PII (identity) | ใช่ | ต้องยืนยัน | DPA, ISO 27001, list coverage | 12 เดือน |
| 7 | QSA (PCI audit) + ASV (scan) + Pentest | **TODO — ยังไม่จ้าง (ดู `23-qsa-asv-pentest-plan.md`)** | 1/2 | อาจเห็น scope | จำกัด (ภายใต้ NDA) | — | PCI SSC listing (QSA/ASV) | ต่อ engagement |
| 8 | SIEM / Log management / Monitoring | SaaS observability (Prometheus/OTel sink, log store) | 2 | ไม่ควรมี CHD (masking) | ใช่ (log อาจมี PII) | ต้องยืนยัน | SOC 2, DPA, log masking config | 12–24 เดือน |
| 9 | Notification (email/SMS/LINE) | ผู้ให้บริการส่งข้อความ merchant/ลูกค้า | 2 | ไม่ | ใช่ | ต้องยืนยัน | DPA | 24 เดือน |
| 10 | WAF / DDoS / CDN | ผู้ให้บริการ edge security | 2 | TLS termination? (ต้องเลี่ยง) | จำกัด | ต้องยืนยัน | SOC 2, PCI AoC ถ้า terminate TLS | 12 เดือน |

> **TODO (ระบุจริงก่อนยื่น):** ต้องเติมชื่อผู้ให้บริการจริง, เลขที่สัญญา, วันหมดอายุ PCI AoC, DPO contact ของผู้ให้บริการ
> และ concentration flag (ผู้ให้บริการรายเดียวรองรับหลายฟังก์ชัน) — ดู callout §12

---

## 5. กระบวนการก่อนผูกพัน (Pre-engagement Due Diligence)

การใช้ผู้ให้บริการ **Tier 1/Tier 2 ทุกราย** ต้องผ่าน due diligence ก่อนลงนาม โดยรวบรวมและประเมิน:

| ด้าน | สิ่งที่ตรวจ | หลักฐาน |
|------|-----------|---------|
| ฐานะการเงิน | ความมั่นคง, ความสามารถให้บริการต่อเนื่อง | งบการเงิน, credit check |
| ความปลอดภัย/PCI | สถานะ PCI ปัจจุบัน, scope, ข้อยกเว้น | **PCI AoC/RoC**, SOC 2 Type II, ISO 27001 |
| PDPA | เป็น Data Processor?, มาตรการคุ้มครองข้อมูล, การโอนข้ามพรมแดน | DPA readiness, บันทึกกิจกรรมประมวลผล |
| ความต่อเนื่อง (BCP/DR) | RTO/RPO ของผู้ให้บริการ vs ของเรา, ผลทดสอบ DR | BCP/DR summary, ผล DR test ล่าสุด |
| AML/reputation | sanction screening ตัวผู้ให้บริการ/UBO, ข่าวเชิงลบ | ผล screening, adverse media |
| Sub-outsourcing | ผู้ให้บริการช่วง (4th party) และการควบคุม | รายการ sub-processor |
| Concentration | ซ้ำกับผู้ให้บริการเดิมหรือไม่ (single point of failure) | mapping ภายใน |

**ผลลัพธ์:** สรุปเป็น **Vendor Risk Assessment Report** ให้คะแนนความเสี่ยง (ข้อ 6) และเสนอผู้อนุมัติตาม Tier
Tier 1 ต้องอนุมัติโดย **ORC** ก่อนลงนามสัญญา

---

## 6. การให้คะแนนความเสี่ยง (Vendor Risk Scoring)

ให้คะแนน 6 มิติ มิติละ 1 (ต่ำ) ถึง 5 (สูง) แล้วถ่วงน้ำหนัก:

| มิติ | น้ำหนัก |
|------|---------|
| Data sensitivity (CHD/PII ที่แตะ) | 30% |
| Security posture (PCI/SOC2/gap) | 25% |
| Operational criticality (impact ถ้าล่ม) | 20% |
| Compliance/PDPA/AML | 10% |
| Financial stability | 10% |
| Concentration / sub-outsourcing | 5% |

**Residual score → Rating:** 1.0–2.0 Low · 2.1–3.5 Medium · 3.6–5.0 High
ผู้ให้บริการ **High** ต้องมีแผนลดความเสี่ยง (mitigation) + อนุมัติ ORC + ทบทวนถี่ขึ้น (ทุก 6 เดือน)

---

## 7. ข้อสัญญาบังคับ (Mandatory Contract Clauses — SLA / Audit-Rights / Exit)

ทุกสัญญา **Tier 1** ต้องมีข้อสัญญาต่อไปนี้ (Tier 2 ปรับตามความเหมาะสม):

1. **ขอบเขตบริการ & SLA** — availability, response/resolution time, capacity, ระบุ metric วัดผลได้ (ตาราง §7.1)
2. **สิทธิเข้าตรวจสอบ (Audit Rights)** — [บริษัท / Company] **และ ธปท./ผู้ตรวจที่ ธปท. แต่งตั้ง** มีสิทธิเข้าตรวจ
   (ทั้ง on-site และเอกสาร) รวมถึงการขอ report การตรวจสอบอิสระ (SOC 2/PCI AoC) โดยไม่จำกัดสิทธินี้
3. **การคุ้มครองข้อมูล (PDPA/DPA)** — เป็นภาคผนวก DPA ตาม ม.40 (ดู `10-dpa-templates.md`), ข้อจำกัดการใช้ข้อมูล,
   การแจ้งเหตุละเมิดข้อมูลภายในเวลากำหนด (§7.2), การคืน/ทำลายข้อมูลเมื่อสิ้นสุด
4. **ความปลอดภัย & PCI** — ผู้ให้บริการต้องคง PCI DSS ตลอดสัญญา, ส่ง AoC ปีละครั้ง, ยอมรับความรับผิดชอบต่อ CHD
   ที่ตนถือ (PCI 12.9), แจ้งเมื่อ scope/สถานะ PCI เปลี่ยน
5. **การแจ้งเหตุ (Incident/Breach Notification)** — แจ้งเหตุด้านความปลอดภัย/ข้อมูลรั่วไหลภายในเวลาที่กำหนด (§7.2)
6. **Sub-outsourcing (4th party)** — ห้ามใช้ผู้ให้บริการช่วงในงานสำคัญโดยไม่แจ้ง/ขออนุมัติล่วงหน้า, ต้องส่งต่อ
   ข้อผูกพันเดียวกัน (flow-down)
7. **ความต่อเนื่องธุรกิจ (BCP/DR)** — ผู้ให้บริการต้องมี BCP/DR, ทดสอบรายปี, และให้เข้าร่วม DR drill ของเรา
8. **Exit / Termination & Transition** — สิทธิบอกเลิกเมื่อผิดสัญญา/ผลตรวจไม่ผ่าน/ ธปท. สั่ง, ระยะเวลา transition,
   การส่งมอบและทำลายข้อมูล, **Exit Plan** (§9)
9. **Liability / Indemnity** — ความรับผิดกรณีข้อมูลรั่ว/ผิด SLA, การประกันภัย (cyber insurance) ถ้าเข้าเกณฑ์
10. **Data residency / Cross-border** — ระบุที่ตั้งข้อมูลตามข้อกำหนด ธปท./PDPA, การโอนข้ามพรมแดนต้องมีฐานทางกฎหมาย
11. **สิทธิของ ธปท.** — ยอมรับให้ ธปท. เข้าถึงข้อมูล/สถานที่/ระบบที่เกี่ยวข้องกับบริการที่ให้แก่บริษัท
12. **การรายงาน** — ส่ง SLA report / security report ตามรอบ (§7.1)

### 7.1 ตัวอย่าง SLA & Service Credit (Tier 1)

| Metric | เป้าหมาย | การวัด | Service credit เมื่อไม่ถึง |
|--------|---------|--------|-----------------------------|
| Availability (บริการหลัก) | ≥ 99.95%/เดือน | uptime monitoring | credit แบบขั้นบันได ตามชั่วโมงที่ขาด |
| P1 incident acknowledgement | ≤ 15 นาที | ticketing | credit + escalation |
| P1 resolution / workaround | ≤ 4 ชั่วโมง | ticketing | credit + RCA ภายใน 5 วันทำการ |
| Breach notification | ≤ 24 ชั่วโมง (§7.2) | log/email | ถือเป็นเหตุผิดสัญญาสำคัญ |
| PCI AoC ประจำปี | ส่งก่อนหมดอายุ | tracker | ระงับ/บอกเลิกได้ |
| SLA report | รายเดือน | report | escalation |

### 7.2 ระยะเวลาแจ้งเหตุ (Notification Timelines)

- **ข้อมูลส่วนบุคคลรั่วไหล (PDPA):** ผู้ให้บริการแจ้ง [บริษัท / Company] **โดยไม่ชักช้า (≤ 24 ชม.)** เพื่อให้บริษัท
  แจ้ง **PDPC ภายใน 72 ชั่วโมง** ตาม ม.37(4) (ดู `16-incident-response-breach.md`)
- **เหตุด้านความปลอดภัย/CDE (PCI):** แจ้ง ≤ 24 ชม., ให้ความร่วมมือในการสอบสวน (forensics)
- **เหตุกระทบบริการชำระเงิน:** แจ้งตาม incident response ของเรา และรายงาน ธปท. ตามเกณฑ์ที่กำหนด

---

## 8. Responsibility Matrix (PCI Req 12.8.5)

จัดทำ matrix ระบุว่า PCI-DSS แต่ละ requirement ใครรับผิดชอบ — [บริษัท / Company], ผู้ให้บริการ, หรือร่วมกัน (shared)
ตัวอย่าง (สรุป):

| PCI Req (ย่อ) | [บริษัท / Company] | Cloud/IaaS | Tokenization/Vault | 3DS |
|---------------|:---:|:---:|:---:|:---:|
| Req 1 Network security controls | Shared | ✔ (infra) | — | — |
| Req 3 Protect stored account data | Shared | — | ✔ | — |
| Req 6 Secure software | ✔ | — | Shared | Shared |
| Req 10 Logging & monitoring | ✔ | Shared | Shared | Shared |
| Req 11 Testing (ASV/pentest) | ✔ (จ้าง) | Shared | Shared | Shared |
| Req 12 Policy & TPSP mgmt | ✔ | — | — | — |

> matrix ฉบับเต็มเก็บแยกและทบทวนพร้อม AoC ประจำปีของแต่ละผู้ให้บริการ

---

## 9. การกำกับต่อเนื่อง & Exit Plan (Ongoing Monitoring & Exit)

**Ongoing monitoring:**
- ทบทวน SLA report รายเดือน (Tier 1), รายไตรมาส (Tier 2)
- เก็บ/ตรวจ **PCI AoC และ SOC 2** ปีละครั้ง (PCI 12.8.4) — ติดตามวันหมดอายุใน tracker
- ประเมินความเสี่ยงใหม่ตามรอบ (§3) หรือเมื่อมี trigger: incident, เปลี่ยน sub-processor, M&A, ผลตรวจไม่ผ่าน
- ทบทวน concentration risk รายปีต่อ ORC

**Exit Plan (บังคับสำหรับ Tier 1):** ต้องมีก่อนลงนาม ระบุ (ก) ผู้ให้บริการสำรอง/ทางเลือก, (ข) ขั้นตอนย้ายข้อมูล/
ระบบและระยะเวลา, (ค) การคืน/ทำลายข้อมูลพร้อมหลักฐาน, (ง) เงื่อนไข trigger การ exit (ธปท. สั่ง, ผิดสัญญาร้ายแรง,
ผู้ให้บริการล้มละลาย) เพื่อไม่ให้เกิด vendor lock-in ที่กระทบความต่อเนื่องของบริการชำระเงิน

---

## 10. ตารางสรุปกระบวนการ (End-to-End)

```
ระบุความจำเป็น ─▶ จัดชั้น Tier ─▶ Due Diligence ─▶ Risk Scoring ─▶ อนุมัติ (ORC/CISO/DPO)
      │                                                                      │
      │                                                                      ▼
      └───────────────  ทบทวน/Exit ◀── Ongoing Monitoring ◀── ลงนามสัญญา (clauses §7) + onboarding
```

---

## 11. เอกสาร/หลักฐานที่เก็บ (Evidence for BOT / QSA)

- ทะเบียนผู้ให้บริการ (TPSP Inventory) + criticality tier
- Vendor Risk Assessment Report ต่อราย + คะแนน
- สัญญา + ภาคผนวก DPA + SLA + audit-right clause
- PCI AoC / SOC 2 / ISO 27001 ของผู้ให้บริการ (ล่าสุด)
- Responsibility matrix (PCI 12.8.5)
- บันทึกการอนุมัติ ORC + minutes ทบทวน concentration
- Exit plan ต่อ Tier 1

---

## 12. ข้อสมมติและงานค้าง (Assumptions & TODO)

> **CALLOUT — ต้องปิดก่อนยื่นจริง:**
> - **Sponsor bank / acquirer:** ยังไม่สรุปคู่สัญญา (สอดคล้อง `ROADMAP.md` §5 critical path) — ต้องเติมชื่อ, สัญญา,
>   scheme certification timeline
> - **Cloud data residency:** ต้องยืนยันว่า region/ข้อมูลอยู่ในไทยตาม ธปท./PDPA (`ARCHITECTURE.md` §8) ก่อนลงนาม
> - **Tokenization vault:** ยังไม่ตัดสินใจ build vs buy — ถ้าใช้ external ต้องมี PCI AoC + DPA
> - **QSA / ASV / Pentest vendor:** ยังไม่จ้าง (ดู `23-qsa-asv-pentest-plan.md`) — ล็อกคิวก่อนตาม critical path
> - **ทุนจดทะเบียนชำระแล้วจริง 50 ล้านบาท:** อ้างอิงตามแผน ต้องยืนยันจากเอกสารนิติบุคคลจริง (`02-financial-projections-capital.md`)
> - ชื่อผู้ให้บริการทั้งหมดในตาราง §4 เป็นตัวอย่าง/แผน ยังไม่ผูกพัน จนกว่าจะผ่าน due diligence + ORC

---
---

# Critical service provider list + vendor risk assessment + SLA/audit-rights clauses per BOT outsourcing rules (English)

> Supporting document for the **Acquiring (Electronic Payment Acceptance) license** application
> under the Payment Systems Act B.E. 2560 · Paid-up capital THB 50M · PCI-DSS v4.0 Level 1
> Document No. **30 — Vendor Risk Management: critical service provider list + vendor risk assessment + SLA/audit-rights clauses per BOT outsourcing rules**
> Owner: **[บริษัท / Company]** · Version 1.0 · Prepared 2026-07-22 · Reviewed every 12 months
>
> **Note:** This is a technical/operational document, not legal advice. It must be reviewed by legal counsel and the QSA
> before submission. Related docs: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`, `13-it-risk-management.md`,
> `14-cyber-resilience.md`, `15-bcp-dr.md`, `18-segregation-of-duties.md`, `21-tokenization-hsm-keymgmt.md`

---

## 1. Purpose & Scope

This policy defines how [บริษัท / Company] **selects, assesses, contracts, governs, monitors and exits** third-party
service providers (outsourcing / TPSPs), especially those supporting **critical / material outsourcing** of the card
acquiring business, on the principle that **outsourcing a function does not outsource the legal accountability**
— [บริษัท / Company] remains fully responsible to the BOT and to customers at all times.

It satisfies:

- **BOT Outsourcing notification** for financial institutions / payment operators under the Payment Systems Act B.E. 2560
  — pre-engagement due diligence, risk-covering contracts, **audit rights** for the company **and for the BOT**,
  contingency/exit plans, and governance of **sub-outsourcing (4th parties)**.
- **BOT IT risk / cyber resilience notifications** — managing third-party technology and concentration/cloud risk.
- **PCI-DSS v4.0 Req 12.8 (12.8.1–12.8.5)** — TPSP inventory, responsibility matrix, at-least-annual monitoring of
  provider PCI status, and due diligence before engagement.
- **PCI-DSS v4.0 Req 12.9** — providers must acknowledge responsibility for CHD they store/process/impact.
- **PDPA B.E. 2562 / PDPC** — where a provider is a **Data Processor**, a **Data Processing Agreement (DPA)** under
  s.40 is required (see `10-dpa-templates.md`), plus cross-border transfer controls (s.28–29).
- **AMLA (AMLO/ปปง.)** — providers performing KYC/CDD/sanction screening on our behalf must remain governable and
  auditable; [บริษัท / Company] retains STR/SAR reporting responsibility.

> **Scope:** every provider that (a) accesses/processes/stores Cardholder Data (CHD) or personal data, (b) supports a
> function affecting payment-service continuity, or (c) connects to the CDE per `ARCHITECTURE.md`. General procurement
> not touching (a)–(c) uses the standard purchasing process and is out of scope here.

---

## 2. Roles & Responsibilities

| Role | Vendor-risk responsibility |
|------|-----------------------------|
| **Business Owner** | Business need, service scope/SLA, day-to-day relationship |
| **Vendor Risk / TPRM Manager** | Owns vendor register, runs due diligence, risk scoring, review cadence |
| **CISO / DevSecOps** | Security/CDE risk, reviews provider PCI AoC, approves technical connectivity |
| **DPO** | Determines processor status, drafts/reviews DPA, assesses cross-border transfers |
| **Compliance / AML Officer** | Reviews AML/KYC/screening providers, sanction-screens the provider itself |
| **Legal** | Drafts/reviews contracts; embeds audit-right / exit / sub-outsourcing / liability clauses |
| **Procurement** | Purchasing, financial vetting, PO/contract admin |
| **Outsourcing Risk Committee (ORC)** | Approves Critical/Material providers; reviews concentration risk |
| **Internal Audit** | Independently tests that the process is followed; samples contracts/assessments |
| **Board** | Approves outsourcing policy; annual review of critical outsourcing risk |

> **Separation of Duties:** whoever selects/negotiates (Business Owner/Procurement) must **not** be the final risk
> approver (ORC). Critical outsourcing requires joint sign-off by CISO + DPO + Compliance.

---

## 3. Vendor Criticality Tiering

| Tier | Definition | Examples | Governance |
|------|-----------|----------|-----------|
| **Tier 1 — Critical / Material** | Failure/breach materially impacts payment service or CHD, or accesses the CDE | Sponsor bank/acquirer, cloud IaaS, HSM/KMS, tokenization vault, 3DS (ACS/DS), KYC/screening | Full due diligence + ORC approval + audit-right + exit plan + on-site right + 12-month review |
| **Tier 2 — Important** | Operationally impactful but with fallback, or limited PII | Email/SMS/LINE, SIEM SaaS, log storage, ASV vendor | Standard due diligence + CISO/DPO approval + DPA + 12–24-month review |
| **Tier 3 — Low** | No sensitive CHD/PII, low impact | Productivity tools, advisors with no customer data | Light due diligence + contractual review |

**Tier 1 thresholds (any one triggers Critical):**
- Accesses/stores/transmits **CHD or authentication data**, or is in/connected to the **CDE**
- Supports a function where downtime > **30 min** impacts authorization/settlement (aligns with RTO in `ARCHITECTURE.md`)
- Processes personal data of > **10,000** cardholders or transfers data cross-border
- Performs a statutory function on our behalf (KYC/CDD, sanction screening, reporting)

---

## 4. Critical Service Provider List (TPSP Inventory — PCI Req 12.8.1)

> Living document; updated on add/change and reviewed at least annually. "PCI responsibility" ties to the matrix (§8).

| # | Service | Provider (example/planned) | Tier | CHD access? | Data Processor (PDPA)? | Domestic/Cross-border | Evidence held | Review |
|---|---------|-----------------------------|------|-------------|------------------------|-----------------------|---------------|--------|
| 1 | Sponsor bank / Acquirer / Card switch | **TODO — not finalized (see callout §12)** | 1 | Indirect (routing) | No (controller-to-controller) | Domestic | Acquiring contract, scheme cert | 12 mo |
| 2 | Cloud / IaaS | e.g. AWS/GCP ap-southeast (TH region when available) — **TODO confirm data residency** | 1 | Yes (hosts CDE) | Yes (sub-processor) | **Must confirm TH residency** (ARCH §8) | PCI AoC, SOC 2 Type II, ISO 27001, DPA | 12 mo |
| 3 | HSM / KMS | Cloud HSM or on-prem PCI-PTS HSM — **TODO** | 1 | Yes (PAN keys) | No (holds keys, not PII) | Domestic | FIPS 140-2/3 L3+, PCI AoC | 12 mo |
| 4 | Tokenization Vault (if external) | **TODO — build vs buy** (ARCH §4) | 1 | Yes (PAN) | Yes | Must be TH | PCI AoC, token scheme | 12 mo |
| 5 | 3-D Secure (ACS/DS/3DS Server) | EMV 3DS v2.x provider — **TODO** | 1 | Auth data | Yes | Must confirm | EMVCo approval, PCI 3DS AoC | 12 mo |
| 6 | KYC/CDD/e-KYC + Sanction screening | Screening provider (UN/OFAC/AMLO lists) — **TODO** | 1 | PII (identity) | Yes | Must confirm | DPA, ISO 27001, list coverage | 12 mo |
| 7 | QSA + ASV + Pentest | **TODO — not engaged (see `23-qsa-asv-pentest-plan.md`)** | 1/2 | May view scope | Limited (under NDA) | — | PCI SSC listing | per engagement |
| 8 | SIEM / Log / Monitoring | SaaS observability | 2 | Should hold no CHD (masking) | Yes (logs may hold PII) | Must confirm | SOC 2, DPA, masking config | 12–24 mo |
| 9 | Notification (email/SMS/LINE) | Messaging provider | 2 | No | Yes | Must confirm | DPA | 24 mo |
| 10 | WAF / DDoS / CDN | Edge security provider | 2 | TLS termination? (avoid) | Limited | Must confirm | SOC 2, PCI AoC if TLS-terminating | 12 mo |

> **TODO before submission:** populate real provider names, contract numbers, PCI AoC expiry, provider DPO contact,
> and concentration flags (one provider covering multiple functions) — see §12.

---

## 5. Pre-engagement Due Diligence

All **Tier 1/Tier 2** providers pass due diligence before signing, assessing:

| Area | Check | Evidence |
|------|-------|----------|
| Financial | Stability, ability to sustain service | Financials, credit check |
| Security/PCI | Current PCI status, scope, exceptions | **PCI AoC/RoC**, SOC 2 Type II, ISO 27001 |
| PDPA | Processor status, safeguards, cross-border | DPA readiness, RoPA |
| BCP/DR | Provider RTO/RPO vs ours, DR test results | BCP/DR summary, latest DR test |
| AML/reputation | Sanction-screen provider/UBO, adverse media | Screening results |
| Sub-outsourcing | 4th parties and controls | Sub-processor list |
| Concentration | Overlap with existing providers (SPOF) | Internal mapping |

Output: a **Vendor Risk Assessment Report** with a risk score (§6). Tier 1 requires **ORC** approval before signing.

---

## 6. Vendor Risk Scoring

Score six dimensions 1 (low)–5 (high), weighted:

| Dimension | Weight |
|-----------|--------|
| Data sensitivity (CHD/PII touched) | 30% |
| Security posture (PCI/SOC2/gaps) | 25% |
| Operational criticality | 20% |
| Compliance/PDPA/AML | 10% |
| Financial stability | 10% |
| Concentration / sub-outsourcing | 5% |

**Rating:** 1.0–2.0 Low · 2.1–3.5 Medium · 3.6–5.0 High. High providers require a mitigation plan + ORC approval +
6-month review cadence.

---

## 7. Mandatory Contract Clauses (SLA / Audit-Rights / Exit)

Every **Tier 1** contract must include (Tier 2 as appropriate):

1. **Service scope & SLA** — availability, response/resolution, capacity, measurable metrics (§7.1)
2. **Audit Rights** — [บริษัท / Company] **and the BOT (or BOT-appointed examiners)** may audit (on-site and documents),
   including independent-assurance reports (SOC 2/PCI AoC); this right is not to be restricted
3. **Data protection (PDPA/DPA)** — DPA annex per s.40 (see `10-dpa-templates.md`): purpose limitation, breach
   notification within set timelines (§7.2), return/destruction on termination
4. **Security & PCI** — maintain PCI DSS throughout the term, deliver AoC annually, accept responsibility for CHD held
   (PCI 12.9), notify on scope/status change
5. **Incident/Breach Notification** — notify within defined timelines (§7.2)
6. **Sub-outsourcing (4th party)** — no material sub-outsourcing without prior notice/approval; flow-down of the same
   obligations
7. **BCP/DR** — provider maintains and annually tests BCP/DR and participates in our DR drills
8. **Exit / Termination & Transition** — right to terminate on breach/failed audit/BOT direction; transition period,
   data handover and destruction, and an **Exit Plan** (§9)
9. **Liability / Indemnity** — liability for breach/SLA failure; cyber insurance where applicable
10. **Data residency / Cross-border** — data location per BOT/PDPA; lawful basis for any cross-border transfer
11. **BOT rights** — provider accepts BOT access to relevant data/premises/systems
12. **Reporting** — SLA/security reports per cadence (§7.1)

### 7.1 Example SLA & Service Credits (Tier 1)

| Metric | Target | Measurement | Service credit if missed |
|--------|--------|-------------|--------------------------|
| Availability (core) | ≥ 99.95%/mo | uptime monitoring | tiered credit by hours lost |
| P1 acknowledgement | ≤ 15 min | ticketing | credit + escalation |
| P1 resolution/workaround | ≤ 4 hours | ticketing | credit + RCA within 5 business days |
| Breach notification | ≤ 24 hours (§7.2) | log/email | material breach of contract |
| Annual PCI AoC | before expiry | tracker | suspend/terminate |
| SLA report | monthly | report | escalation |

### 7.2 Notification Timelines

- **Personal-data breach (PDPA):** provider notifies [บริษัท / Company] **without delay (≤ 24 h)** so we can notify
  **the PDPC within 72 hours** per s.37(4) (see `16-incident-response-breach.md`)
- **Security/CDE incident (PCI):** notify ≤ 24 h; cooperate with forensics
- **Payment-impacting incident:** per our incident response and BOT reporting thresholds

---

## 8. Responsibility Matrix (PCI Req 12.8.5)

A matrix records which party owns each PCI requirement — [บริษัท / Company], provider, or shared. Summary example:

| PCI Req (short) | [บริษัท / Company] | Cloud/IaaS | Tokenization/Vault | 3DS |
|-----------------|:---:|:---:|:---:|:---:|
| Req 1 Network security controls | Shared | ✔ (infra) | — | — |
| Req 3 Protect stored account data | Shared | — | ✔ | — |
| Req 6 Secure software | ✔ | — | Shared | Shared |
| Req 10 Logging & monitoring | ✔ | Shared | Shared | Shared |
| Req 11 Testing (ASV/pentest) | ✔ (engaged) | Shared | Shared | Shared |
| Req 12 Policy & TPSP mgmt | ✔ | — | — | — |

> The full matrix is maintained separately and reviewed alongside each provider's annual AoC.

---

## 9. Ongoing Monitoring & Exit Plan

**Ongoing monitoring:**
- Review SLA reports monthly (Tier 1), quarterly (Tier 2)
- Collect/verify **PCI AoC and SOC 2 annually** (PCI 12.8.4); track expiry in a tracker
- Re-assess on cadence (§3) or on triggers: incident, sub-processor change, M&A, failed audit
- Annual concentration-risk review by ORC

**Exit Plan (mandatory for Tier 1):** in place before signing — (a) alternate/backup providers, (b) data/system
migration steps and timeline, (c) data return/destruction with evidence, (d) exit triggers (BOT direction, material
breach, provider insolvency) — to avoid vendor lock-in that threatens payment continuity.

---

## 10. End-to-End Process

```
Identify need ─▶ Assign Tier ─▶ Due Diligence ─▶ Risk Scoring ─▶ Approve (ORC/CISO/DPO)
      │                                                                 │
      │                                                                 ▼
      └──────────────  Review/Exit ◀── Ongoing Monitoring ◀── Sign contract (§7) + onboarding
```

---

## 11. Evidence Held (for BOT / QSA)

- TPSP inventory + criticality tier
- Per-provider Vendor Risk Assessment Report + score
- Contracts + DPA annex + SLA + audit-right clause
- Provider PCI AoC / SOC 2 / ISO 27001 (latest)
- Responsibility matrix (PCI 12.8.5)
- ORC approval records + concentration review minutes
- Exit plans for Tier 1

---

## 12. Assumptions & TODO

> **CALLOUT — close before actual submission:**
> - **Sponsor bank / acquirer:** counterparty not finalized (per `ROADMAP.md` §5 critical path) — add name, contract,
>   scheme-certification timeline
> - **Cloud data residency:** confirm region/data resides in Thailand per BOT/PDPA (`ARCHITECTURE.md` §8) before signing
> - **Tokenization vault:** build vs buy undecided — if external, requires PCI AoC + DPA
> - **QSA / ASV / Pentest vendor:** not yet engaged (see `23-qsa-asv-pentest-plan.md`) — lock queue early per critical path
> - **Actual paid-up capital THB 50M:** per plan; confirm against corporate documents (`02-financial-projections-capital.md`)
> - All provider names in §4 are examples/planned and non-binding until they pass due diligence + ORC
