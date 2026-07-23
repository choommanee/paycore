# แผนตอบสนองเหตุการณ์และการแจ้งเหตุข้อมูลรั่วไหล (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring Service)**
> ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท**
> ควบคู่กับมาตรฐาน **PCI-DSS v4.0 Level 1** (โดยเฉพาะ **Requirement 12.10 — Incident Response**) และภาระหน้าที่
> การแจ้งเหตุตาม **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA) มาตรา 37(4)** กำกับโดย **คณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC)**
>
> รหัสเอกสาร: `COMP-16` · เวอร์ชัน 0.1 · จัดทำ 2026-07-22 · ทบทวนทุก 12 เดือน (และหลังเหตุการณ์สำคัญทุกครั้ง)
> เจ้าของเอกสาร: **Chief Information Security Officer (CISO)** ร่วมกับ **Data Protection Officer (DPO)**
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`, `./10-dpa-templates.md`, `./11-data-retention-deletion.md`, `./07-sar-str-procedure.md`
>
> **ข้อจำกัดความรับผิด:** เอกสารนี้เป็นข้อมูลอ้างอิงเชิงโครงสร้าง/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย
> ต้องผ่านการทบทวนโดยที่ปรึกษากฎหมายด้านใบอนุญาต ธปท., ที่ปรึกษา PDPA และ QSA ก่อนยื่นจริง

---

> [!IMPORTANT]
> **สมมติฐานและรายการที่ยังไม่สรุป (Assumptions / TODO)** — ต้องเติมค่าจริงหรือยืนยันก่อนยื่น ธปท./PDPC และก่อน go-live
>
> | # | รายการ | สถานะ | ผู้รับผิดชอบ |
> |---|--------|-------|-------------|
> | A1 | **ชื่อบริษัทจริง** — ใช้ placeholder `[บริษัท / Company]` ทั้งเอกสาร | TODO | Corporate Secretary |
> | A2 | **Sponsor bank / Acquiring bank** — ยังไม่ลงนาม (เส้นทาง B ตาม ROADMAP) มีผลต่อ SLA การแจ้งเหตุ, ADC/PFI process ของ card scheme และช่องทาง escalation | ยังไม่สรุป | CEO / Head of Partnerships |
> | A3 | **QSA / PFI vendor** (Qualified Security Assessor / PCI Forensic Investigator) — ยังไม่คัดเลือก; ต้องมี PFI ที่ scheme ยอมรับ retained ล่วงหน้า | ยังไม่สรุป | CISO |
> | A4 | **DFIR retainer** (Digital Forensics & Incident Response) และผู้ให้บริการ MDR/SOC 24×7 | ยังไม่สรุป | CISO |
> | A5 | **ช่องทาง/แบบฟอร์มแจ้งเหตุจริงของ ธปท.** ตามประกาศ IT Risk / Cyber Resilience (SLA ที่ต้องยืนยัน — ดูข้อ 5) | TODO | CCO / ที่ปรึกษากฎหมาย |
> | A6 | **ช่องทางแจ้งเหตุ PDPC** (ระบบ e-Notification ของ PDPC) และผู้ลงนามที่ได้รับมอบอำนาจ | TODO | DPO |
> | A7 | **Cyber insurance** (breach response / third-party liability) — วงเงินและเงื่อนไขการแจ้ง | ยังไม่สรุป | CFO / CISO |
> | A8 | **รายชื่อ + เบอร์ติดต่อจริงของทีม CSIRT** (call tree / on-call roster) — จัดเก็บแยกใน runbook ที่ควบคุมการเข้าถึง | TODO | CISO |
> | A9 | **ที่ปรึกษาประชาสัมพันธ์/สื่อสารภาวะวิกฤต** (crisis comms) และ template แถลงการณ์ที่ผ่านฝ่ายกฎหมาย | ยังไม่สรุป | Head of Comms |
>
> ห้ามกรอกชื่อผู้ให้บริการ/เบอร์/วงเงินสมมติในช่องข้างต้นลงในเอกสารที่ยื่นจริง — ต้องเป็นข้อมูลที่ยืนยันได้เท่านั้น

---

## 1. วัตถุประสงค์และขอบเขต (Purpose & Scope)

เอกสารนี้กำหนด **แผนตอบสนองเหตุการณ์ความมั่นคงปลอดภัย (Incident Response Plan — IRP)** และ **ขั้นตอนการแจ้งเหตุข้อมูลรั่วไหล (Breach Notification Procedure)** ของ [บริษัท / Company] ในฐานะผู้ให้บริการ Acquiring Gateway เพื่อให้:

1. ตรวจจับ ควบคุม กำจัด และฟื้นฟูจากเหตุการณ์ได้อย่างรวดเร็วและมีระเบียบ (ตาม **PCI-DSS v4.0 Req 12.10.1–12.10.7**)
2. แจ้งเหตุต่อหน่วยงานกำกับและผู้มีส่วนได้เสียภายในกรอบเวลาที่กฎหมายกำหนด โดยเฉพาะ **PDPA มาตรา 37(4): แจ้ง PDPC ภายใน 72 ชั่วโมง** นับแต่ทราบเหตุ
3. รักษาหลักฐานทางนิติวิทยาศาสตร์ (chain of custody) เพื่อรองรับ **PFI investigation** ของ card scheme และการสอบสวนของทางการ

**ขอบเขต:** ครอบคลุมทุกระบบใน CDE (Cardholder Data Environment) และระบบสนับสนุนตาม `ARCHITECTURE.md` ได้แก่ API Edge, Payment Core, Tokenization Vault, Acquirer/3DS Adapter, Ledger, Webhook, ฐานข้อมูล, HSM/KMS, log/SIEM, เครือข่าย รวมถึงผู้ให้บริการภายนอก (outsourced/cloud) พนักงาน กรรมการ และผู้รับจ้างช่วง

> **หลักการสำคัญที่ยึดตลอดเอกสาร:** ตาม `ARCHITECTURE.md` ข้อ 6 และเอกสาร `11-data-retention-deletion.md` —
> ระบบ **ไม่จัดเก็บ full PAN, CVV/CVV2, PIN, full track (SAD)** ใน operational DB จึงลด "blast radius" ของเหตุการณ์ลงอย่างมาก
> ข้อมูลที่รั่วได้จริงในระบบหลักคือ token, `card_last4`, `card_brand` และ PII/ข้อมูล KYC ของ merchant เป็นหลัก

---

## 2. บทบาทและความรับผิดชอบ — CSIRT (Roles & Responsibilities)

ทีมตอบสนองเหตุการณ์เรียกว่า **CSIRT (Computer Security Incident Response Team)** เปิดใช้เมื่อประกาศเหตุการณ์ระดับ **SEV-1/SEV-2**

| บทบาท | หน้าที่หลักในการตอบสนองเหตุการณ์ |
|-------|--------------------------------|
| **Incident Commander (IC)** — โดยปกติคือ CISO หรือผู้ที่ได้รับมอบหมาย | ผู้มีอำนาจสูงสุดในเหตุการณ์: ประกาศระดับความรุนแรง, สั่งการ containment, อนุมัติ downtime/isolation, ตัดสินใจ escalate |
| **Deputy IC** | ปฏิบัติหน้าที่แทน IC เมื่อ IC ไม่พร้อม/มีผลประโยชน์ทับซ้อน |
| **CISO** | เจ้าของแผน IRP, กำกับด้านเทคนิคความมั่นคงปลอดภัย, จุดติดต่อ QSA/PFI/DFIR |
| **DPO (Data Protection Officer)** | ประเมินว่าเป็น "personal data breach" หรือไม่, ตัดสินใจแจ้ง PDPC ภายใน 72 ชม. และแจ้งเจ้าของข้อมูล, ดูแล PDPA |
| **CCO / Compliance** | จุดติดต่อ **ธปท.** และ (หากเข้าเงื่อนไข AML) ปปง., ประเมินภาระรายงานตามใบอนุญาต |
| **MLRO** | ประเมินความเชื่อมโยงกับการฟอกเงิน/ทุจริต; หากพบให้เริ่ม STR ตาม `07-sar-str-procedure.md` |
| **Forensics Lead (SOC/DFIR)** | เก็บหลักฐาน, chain of custody, triage, root cause, ประสาน PFI |
| **Engineering / SRE On-call** | containment, isolation, patch, restore จาก backup, ยืนยัน RPO/RTO |
| **Legal Counsel** | สิทธิ/หน้าที่ตามกฎหมาย, privilege, ทบทวนแถลงการณ์ก่อนเผยแพร่ |
| **Head of Comms / PR** | สื่อสารภายใน-ภายนอก, สื่อ, ลูกค้า, ภายใต้การอนุมัติของ IC + Legal |
| **Merchant Support Lead** | สื่อสารกับ merchant ที่ได้รับผลกระทบ, ช่องทาง reissue/monitoring |
| **HR** | เหตุการณ์ที่เกี่ยวข้องกับบุคลากร (insider), มาตรการทางวินัย |
| **Executive Sponsor (CEO/Board)** | รับทราบ SEV-1, อนุมัติทรัพยากร/งบฉุกเฉิน, ความรับผิดชอบต่อ ธปท. |

> **หลักสำคัญ:** มีเพียง **DPO** เท่านั้นที่ตัดสินใจแจ้ง PDPC, มีเพียง **CCO/Compliance** ที่ติดต่อ ธปท. อย่างเป็นทางการ และมีเพียง **IC + Legal + Comms** ที่อนุมัติการสื่อสารภายนอก เพื่อคุมความถูกต้องและป้องกัน tipping-off
> รายชื่อจริง เบอร์ติดต่อ และ call tree เก็บใน runbook แยก (รายการ A8) ไม่เปิดเผยในเอกสารยื่น

---

## 3. การจัดระดับความรุนแรง (Severity Classification)

| ระดับ | นิยาม | ตัวอย่าง | ผู้เปิดใช้ | เป้า Response |
|-------|-------|---------|-----------|--------------|
| **SEV-1 (Critical)** | มี/สงสัยว่ามีการเข้าถึงหรือรั่วไหลของ **CHD/SAD/PII จำนวนมาก**, ransomware, การเจาะ CDE, HSM/key compromise, ระบบ authorization ล่มทั้งระบบ | ผู้ไม่ประสงค์ดีเข้าถึง Tokenization Vault; PAN/PII ถูก exfiltrate | IC + Executive Sponsor + Board notify | **Acknowledge ≤ 15 นาที**, ตั้ง war room ≤ 30 นาที |
| **SEV-2 (High)** | เหตุการณ์ยืนยันแล้วที่กระทบความลับ/ความถูกต้อง/ความพร้อมใช้อย่างมีนัยสำคัญ แต่ยังจำกัดขอบเขต | account takeover ของ admin เดี่ยว; DDoS ทำ auth ล่มบางส่วน; malware บน endpoint ใน scope | IC | Acknowledge ≤ 30 นาที |
| **SEV-3 (Medium)** | เหตุการณ์จำกัดขอบเขต ควบคุมได้ ไม่มีหลักฐานการรั่วของข้อมูล | ความพยายาม brute-force ที่ถูกบล็อก; misconfiguration ที่ยังไม่ถูก exploit | On-call lead | Acknowledge ≤ 4 ชม. |
| **SEV-4 (Low)** | เหตุการณ์เล็กน้อย/near-miss/policy violation | phishing ที่ผู้ใช้รายงานและไม่ได้คลิก; log anomaly เดี่ยว | SOC analyst | ตามรอบงานปกติ |

**เกณฑ์ยกระดับ (escalation triggers):** ยกเป็น SEV-1 ทันทีเมื่อเข้าข่ายข้อใดข้อหนึ่ง — (ก) สงสัยว่ามีการเข้าถึง CHD/SAD, (ข) key/HSM ถูก compromise, (ค) ยืนยัน exfiltration ของ PII, (ง) ระบบชำระเงินหลักไม่พร้อมใช้เกิน RTO (30 นาที ตาม `ARCHITECTURE.md`), (จ) card scheme/acquirer/หน่วยงานภายนอกแจ้งว่าเราคือ CPP (Common Point of Purchase)

> **หมายเหตุ PCI:** เหตุการณ์ที่กระทบ **account data** ตาม PCI-DSS v4.0 ต้องถือเป็น SEV-1/SEV-2 เสมอ และเข้าสู่กระบวนการ **ADC (Account Data Compromise)** ของ scheme ผ่าน acquirer

---

## 4. วงจรตอบสนองเหตุการณ์ (Incident Lifecycle — PCI Req 12.10.1)

```
Detect ──▶ Triage/Classify ──▶ Contain ──▶ Eradicate ──▶ Recover ──▶ Notify ──▶ Post-Incident Review
  │            │(severity)        │           │            │           │(72h/ธปท.)      │(lessons learned)
  └── SIEM/IDS/WAF/anomaly       └── isolate  └── patch    └── restore └── regulators  └── update controls
      /merchant/scheme report        /revoke      /rebuild     (RPO≤5m)     /merchants
```

1. **Detect & Report** — แหล่งตรวจจับ: SIEM/alert (PCI Req 10), IDS/IPS, WAF, FIM, `audit_log`, anomaly ของ Risk Engine, การแจ้งจาก merchant/acquirer/scheme หรือ external. ทุกคนมีหน้าที่รายงานผ่านช่องทางเดียว (security@ / on-call hotline) ภายในเวลาที่กำหนด
2. **Triage & Classify** — IC/on-call ประเมินความรุนแรงตามข้อ 3, เปิด incident ticket, บันทึก timeline
3. **Contain** — isolate host/segment, revoke credentials/keys, block IP, disable API key ที่ถูกใช้, snapshot เพื่อ forensics **ก่อน** ทำลายหลักฐาน
4. **Eradicate** — ลบ malware/backdoor, patch ช่องโหว่, หมุนคีย์ (HSM/KMS key rotation, PCI Req 3), reset credential ทั้งหมดที่เกี่ยวข้อง
5. **Recover** — restore จาก backup ที่สะอาด (RPO ≤ 5 นาที, RTO ≤ 30 นาที), ตรวจ integrity, เฝ้าระวังเข้มข้น (heightened monitoring) อย่างน้อย 14 วัน
6. **Notify** — แจ้งหน่วยงานกำกับและผู้มีส่วนได้เสียตามข้อ 5–6
7. **Post-Incident Review** — RCA ภายใน 10 วันทำการ, lessons learned, ปรับ control, อัปเดตแผน (PCI Req 12.10.6)

---

## 5. กรอบเวลาการแจ้งเหตุ (Notification Timelines & Thresholds)

> [!WARNING]
> **72 ชั่วโมง (PDPA ม.37(4))** และกรอบเวลาการแจ้ง ธปท./scheme เป็นภาระตามกฎหมาย/สัญญาที่ห้ามพลาด
> เริ่มนับ "นาฬิกา 72 ชม." **ตั้งแต่เวลาที่ผู้ควบคุมข้อมูล 'ทราบเหตุ (become aware)'** ไม่ใช่ตั้งแต่เกิดเหตุ — ต้องบันทึกเวลานี้ให้ชัดใน timeline

| ผู้รับแจ้ง | เงื่อนไข/เกณฑ์ | กรอบเวลา | ผู้รับผิดชอบ | อ้างอิง |
|-----------|---------------|----------|-------------|---------|
| **PDPC (สคส.)** | มี **personal data breach** ที่ **มีความเสี่ยงต่อสิทธิและเสรีภาพของบุคคล** | **ภายใน 72 ชม.** นับแต่ทราบเหตุ (เว้นแต่ความเสี่ยงต่ำ) | DPO | PDPA ม.37(4) |
| **เจ้าของข้อมูล (Data Subjects)** | breach มี **ความเสี่ยงสูง** ต่อสิทธิและเสรีภาพ | **โดยไม่ชักช้า (without undue delay)** พร้อมมาตรการเยียวยา | DPO + Comms | PDPA ม.37(4) |
| **ธปท. (BOT)** | เหตุการณ์ IT/cyber ที่มีนัยสำคัญต่อระบบชำระเงิน/ผู้ใช้บริการ (SEV-1/SEV-2) | ตามประกาศ ธปท. ด้าน IT Risk/Cyber Resilience — **โดยเร็ว** (ยืนยัน SLA จริง — A5) | CCO | ประกาศ ธปท. |
| **Acquirer / Card Scheme (Visa/Mastercard)** | สงสัย/ยืนยัน **account data compromise (ADC)** | **โดยทันที / ภายใน 24 ชม.** ตามกฎ scheme (ยืนยันผ่าน sponsor bank — A2) | CISO + Compliance | PCI DSS Req 12.10; Visa/MC ADC rules |
| **PFI (Forensic Investigator)** | ADC ที่ scheme กำหนดให้สอบสวน | เริ่ม engage **ภายใน 5 วัน** ตามกฎ scheme | CISO | Scheme ADC program |
| **ปปง. (AMLO)** | หากเหตุการณ์เชื่อมโยงกับการฟอกเงิน/ฉ้อโกงเข้าเกณฑ์ STR | ตาม `07-sar-str-procedure.md` | MLRO | พ.ร.บ.ปปง. |
| **ผู้บังคับใช้กฎหมาย (ตำรวจ/ปอท.)** | อาชญากรรมไซเบอร์/ ransomware/ extortion | ตามคำแนะนำ Legal | Legal + CISO | — |
| **Merchants ที่ได้รับผลกระทบ** | ข้อมูล merchant/ผู้ถือบัตรของ merchant กระทบ | โดยไม่ชักช้า ตามสัญญา/SLA | Merchant Support | สัญญาบริการ |
| **Cyber insurer** | เข้าเงื่อนไขกรมธรรม์ | ตามเงื่อนไขแจ้งเหตุของกรมธรรม์ (A7) | CFO/CISO | กรมธรรม์ |

**เนื้อหาขั้นต่ำที่ต้องมีในการแจ้ง PDPC/เจ้าของข้อมูล:** ลักษณะของเหตุการณ์และประเภท/ปริมาณข้อมูลที่กระทบ, ผลกระทบที่อาจเกิด, มาตรการที่ดำเนินการ/จะดำเนินการเพื่อจัดการและลดความเสียหาย, ช่องทางติดต่อ (DPO) หากยังไม่ทราบครบสามารถแจ้งเป็นระยะ (phased notification) ได้

---

## 6. การสื่อสาร (Communications Plan)

| ช่องทาง | ผู้รับ | ผู้อนุมัติ | หมายเหตุ |
|--------|-------|-----------|---------|
| **Internal war room** (ช่องทางเข้ารหัส out-of-band เช่น bridge line) | CSIRT | IC | ห้ามใช้ระบบที่อาจถูก compromise |
| **Executive/Board brief** | CEO, Board | IC | SEV-1 แจ้งทันที |
| **Regulator notice** (PDPC, ธปท.) | หน่วยงานกำกับ | DPO/CCO + Legal | เอกสารทางการเท่านั้น |
| **Scheme/Acquirer notice** | sponsor bank, Visa/MC | CISO + Compliance | ผ่านช่องทางที่ scheme กำหนด |
| **Merchant notice** | merchant | Merchant Support Lead | template ผ่าน Legal |
| **Public/Press statement** | สื่อ/สาธารณะ | IC + Legal + Comms | ต้องผ่าน Legal ก่อนเสมอ |

**หลักการสื่อสาร:** (1) **out-of-band** — ไม่สื่อสารเรื่องเหตุการณ์ผ่านระบบที่อาจถูกเจาะ; (2) **single source of truth** — Comms เป็นผู้เผยแพร่แต่ผู้เดียว; (3) **no speculation** — สื่อสารเฉพาะข้อเท็จจริงที่ยืนยันแล้ว; (4) ป้องกัน **tipping-off** เมื่อมีมิติ AML/สอบสวน; (5) ทุกการสื่อสารภายนอกผ่าน Legal

---

## 7. การเก็บหลักฐานและนิติวิทยาศาสตร์ (Evidence & Forensics)

- **Chain of custody** — บันทึกผู้ครอบครอง เวลา และการเข้าถึงหลักฐานทุกชิ้น; เก็บ hash เพื่อ integrity
- **Preserve before eradicate** — snapshot ระบบ/ดิสก์/memory และสำเนา log ก่อน remediation
- **Log integrity** — `audit_log` เป็น append-only (ตาม `ARCHITECTURE.md`); ส่ง log ไป SIEM แบบ write-once
- **PFI engagement** — เมื่อ scheme กำหนด ให้ใช้ PFI ที่ได้รับการรับรอง (A3); ห้ามทำลายหลักฐานก่อน PFI ตรวจ
- **Retention หลักฐานเหตุการณ์** — เก็บอย่างน้อยตาม `11-data-retention-deletion.md` และตามคำแนะนำ Legal/PFI (โดยทั่วไป ≥ 12 เดือนหรือจนคดี/การสอบสวนสิ้นสุด)

---

## 8. การทดสอบ ฝึกซ้อม และการปรับปรุง (Testing & Maintenance — PCI Req 12.10.2)

- **Tabletop exercise** อย่างน้อย **ปีละ 1 ครั้ง** (สถานการณ์: PAN exfiltration, ransomware, key compromise, insider)
- **ทบทวนและปรับแผน** อย่างน้อยปีละครั้งและหลังทุกเหตุการณ์ SEV-1/SEV-2 (Req 12.10.6)
- **ฝึกอบรมทีม CSIRT** และ **security awareness** ทั้งองค์กรประจำปี
- **24×7 availability** ของบุคลากรตอบสนอง (on-call roster / MDR — A4) ตาม Req 12.10.3
- **ตรวจสอบความพร้อมของ detection controls** (IDS/IPS, FIM, anomaly) เป็นระยะ (Req 12.10.5)
- ผลการฝึกซ้อมและ RCA เก็บเป็นหลักฐานประกอบการตรวจ QSA และการรายงาน ธปท.

---
---

# Incident response + breach notification procedure (PCI Req 12, PDPA 72h): severity, roles, comms (English)

> Supporting document for the license application for **Acquiring Service** under the **Payment Systems Act B.E. 2560**,
> supervised by the **Bank of Thailand (BOT / ธปท.)**, minimum paid-up capital **THB 50M**.
> Aligned with **PCI-DSS v4.0 Level 1** (notably **Requirement 12.10 — Incident Response**) and the breach-notification
> duty under **PDPA Section 37(4)**, supervised by the **Personal Data Protection Committee (PDPC)**.
>
> Document ID: `COMP-16` · Version 0.1 · Issued 2026-07-22 · Reviewed every 12 months (and after every major incident)
> Owner: **Chief Information Security Officer (CISO)** jointly with **Data Protection Officer (DPO)**
> Related: `../COMPLIANCE-TH.md`, `../ARCHITECTURE.md`, `../ROADMAP.md`, `./04-org-chart-governance.md`, `./10-dpa-templates.md`, `./11-data-retention-deletion.md`, `./07-sar-str-procedure.md`
>
> **Disclaimer:** Structural/operational reference, not legal advice. Must be reviewed by BOT-licensing counsel,
> PDPA counsel, and the QSA before actual submission.

---

> [!IMPORTANT]
> **Assumptions / TODO** — must be filled in or confirmed before BOT/PDPC submission and before go-live.
>
> | # | Item | Status | Owner |
> |---|------|--------|-------|
> | A1 | **Actual company name** — placeholder `[บริษัท / Company]` used throughout | TODO | Corporate Secretary |
> | A2 | **Sponsor / acquiring bank** — not yet signed (Path B per ROADMAP); drives notification SLAs, scheme ADC/PFI process, and escalation channel | Open | CEO / Head of Partnerships |
> | A3 | **QSA / PFI vendor** (Qualified Security Assessor / PCI Forensic Investigator) — not yet selected; a scheme-approved PFI must be retained in advance | Open | CISO |
> | A4 | **DFIR retainer** and 24×7 MDR/SOC provider | Open | CISO |
> | A5 | **Actual BOT reporting channel/form** per IT Risk / Cyber Resilience notifications (SLA to confirm — see §5) | TODO | CCO / counsel |
> | A6 | **PDPC notification channel** (PDPC e-Notification) and authorized signatory | TODO | DPO |
> | A7 | **Cyber insurance** (breach response / third-party) — limits and notice conditions | Open | CFO / CISO |
> | A8 | **Real CSIRT contact list** (call tree / on-call roster) — stored in an access-controlled runbook | TODO | CISO |
> | A9 | **Crisis-comms/PR advisor** and legal-approved statement templates | Open | Head of Comms |
>
> Do not enter fictitious vendor names/numbers/limits above into any document actually submitted — verified data only.

---

## 1. Purpose & Scope

This document defines the **Incident Response Plan (IRP)** and **Breach Notification Procedure** for [บริษัท / Company] as an Acquiring Gateway, so as to:

1. Detect, contain, eradicate, and recover from incidents rapidly and in an orderly way (**PCI-DSS v4.0 Req 12.10.1–12.10.7**).
2. Notify regulators and stakeholders within statutory timelines, especially **PDPA Section 37(4): notify PDPC within 72 hours** of becoming aware.
3. Preserve forensic evidence (chain of custody) to support the card scheme's **PFI investigation** and any official inquiry.

**Scope:** All CDE (Cardholder Data Environment) and supporting systems per `ARCHITECTURE.md` — API Edge, Payment Core, Tokenization Vault, Acquirer/3DS adapters, Ledger, Webhook, databases, HSM/KMS, logging/SIEM, network — plus outsourced/cloud providers, staff, directors, and subcontractors.

> **Guiding principle throughout:** Per `ARCHITECTURE.md` §6 and `11-data-retention-deletion.md`, the platform **stores no full PAN, CVV/CVV2, PIN, or full track (SAD)** in the operational DB, which sharply reduces the "blast radius." The data actually exposable in the core systems is primarily tokens, `card_last4`, `card_brand`, and merchant PII/KYC.

---

## 2. Roles & Responsibilities — CSIRT

The response team is the **CSIRT (Computer Security Incident Response Team)**, activated for **SEV-1/SEV-2** incidents.

| Role | Primary incident-response duty |
|------|--------------------------------|
| **Incident Commander (IC)** — normally CISO or delegate | Ultimate authority in the incident: declares severity, orders containment, approves downtime/isolation, decides escalation |
| **Deputy IC** | Acts for the IC when unavailable / conflicted |
| **CISO** | Owns the IRP, leads technical security, liaison to QSA/PFI/DFIR |
| **DPO (Data Protection Officer)** | Assesses whether a "personal data breach" occurred, decides on PDPC notification within 72h and data-subject notice, owns PDPA |
| **CCO / Compliance** | Contact point for **BOT** and (if AML-relevant) AMLO; assesses license reporting duties |
| **MLRO** | Assesses money-laundering/fraud linkage; if present, initiates STR per `07-sar-str-procedure.md` |
| **Forensics Lead (SOC/DFIR)** | Evidence collection, chain of custody, triage, root cause, PFI coordination |
| **Engineering / SRE On-call** | Containment, isolation, patching, restore from backup, RPO/RTO confirmation |
| **Legal Counsel** | Legal rights/duties, privilege, review of statements before release |
| **Head of Comms / PR** | Internal/external communications, media, customers — under IC + Legal approval |
| **Merchant Support Lead** | Communications to affected merchants; reissue/monitoring channels |
| **HR** | Personnel-related (insider) incidents, disciplinary measures |
| **Executive Sponsor (CEO/Board)** | Acknowledges SEV-1, approves emergency resources/budget, BOT accountability |

> **Key rule:** Only the **DPO** decides on PDPC notification; only **CCO/Compliance** formally contacts BOT; and only **IC + Legal + Comms** approve external communications — to preserve accuracy and prevent tipping-off.
> Real names, contacts, and call tree live in a separate access-controlled runbook (A8), not in the submission document.

---

## 3. Severity Classification

| Level | Definition | Examples | Declared by | Response target |
|-------|-----------|----------|-------------|-----------------|
| **SEV-1 (Critical)** | Confirmed/suspected access to or exfiltration of **large-scale CHD/SAD/PII**, ransomware, CDE breach, HSM/key compromise, full authorization outage | Attacker reaches the Tokenization Vault; PAN/PII exfiltrated | IC + Executive Sponsor + Board notify | **Ack ≤ 15 min**, war room ≤ 30 min |
| **SEV-2 (High)** | Confirmed incident materially affecting confidentiality/integrity/availability but still scoped | Single admin account takeover; partial auth outage via DDoS; in-scope endpoint malware | IC | Ack ≤ 30 min |
| **SEV-3 (Medium)** | Contained, scoped incident with no evidence of data exposure | Blocked brute-force; unexploited misconfiguration | On-call lead | Ack ≤ 4 h |
| **SEV-4 (Low)** | Minor / near-miss / policy violation | User-reported phishing not clicked; isolated log anomaly | SOC analyst | Normal workflow |

**Escalation triggers:** Escalate to SEV-1 immediately if any of: (a) suspected access to CHD/SAD; (b) key/HSM compromise; (c) confirmed PII exfiltration; (d) core payment unavailable beyond RTO (30 min per `ARCHITECTURE.md`); (e) a card scheme/acquirer/third party flags us as a **CPP (Common Point of Purchase)**.

> **PCI note:** Any incident affecting **account data** under PCI-DSS v4.0 must be treated as SEV-1/SEV-2 and entered into the scheme's **ADC (Account Data Compromise)** process via the acquirer.

---

## 4. Incident Lifecycle (PCI Req 12.10.1)

```
Detect ──▶ Triage/Classify ──▶ Contain ──▶ Eradicate ──▶ Recover ──▶ Notify ──▶ Post-Incident Review
  │            │(severity)        │           │            │           │(72h/BOT)       │(lessons learned)
  └── SIEM/IDS/WAF/anomaly       └── isolate  └── patch    └── restore └── regulators  └── update controls
      /merchant/scheme report        /revoke      /rebuild     (RPO≤5m)     /merchants
```

1. **Detect & Report** — Sources: SIEM/alerts (PCI Req 10), IDS/IPS, WAF, FIM, `audit_log`, Risk-Engine anomalies, reports from merchant/acquirer/scheme or external. Everyone must report via a single channel (security@ / on-call hotline) within the defined time.
2. **Triage & Classify** — IC/on-call assigns severity per §3, opens the incident ticket, starts a timeline.
3. **Contain** — Isolate host/segment, revoke credentials/keys, block IPs, disable abused API keys; snapshot for forensics **before** destroying evidence.
4. **Eradicate** — Remove malware/backdoors, patch vulnerabilities, rotate keys (HSM/KMS, PCI Req 3), reset all related credentials.
5. **Recover** — Restore from clean backup (RPO ≤ 5 min, RTO ≤ 30 min), verify integrity, apply heightened monitoring for at least 14 days.
6. **Notify** — Notify regulators and stakeholders per §5–6.
7. **Post-Incident Review** — RCA within 10 business days, lessons learned, control updates, plan revision (PCI Req 12.10.6).

---

## 5. Notification Timelines & Thresholds

> [!WARNING]
> **72 hours (PDPA s.37(4))** and the BOT/scheme timelines are statutory/contractual duties that must not be missed.
> The 72-hour clock starts when the **controller "becomes aware"**, not when the incident occurred — record this timestamp precisely in the timeline.

| Recipient | Condition / threshold | Timeline | Owner | Reference |
|-----------|----------------------|----------|-------|-----------|
| **PDPC** | A **personal data breach** posing a **risk to the rights and freedoms** of individuals | **Within 72h** of becoming aware (unless low risk) | DPO | PDPA s.37(4) |
| **Data Subjects** | Breach posing a **high risk** to rights and freedoms | **Without undue delay**, with remedial measures | DPO + Comms | PDPA s.37(4) |
| **BOT (ธปท.)** | Significant IT/cyber incident affecting the payment system/users (SEV-1/SEV-2) | Per BOT IT Risk/Cyber Resilience notifications — **promptly** (confirm actual SLA — A5) | CCO | BOT notification |
| **Acquirer / Card Scheme (Visa/MC)** | Suspected/confirmed **account data compromise (ADC)** | **Immediately / within 24h** per scheme rules (confirm via sponsor bank — A2) | CISO + Compliance | PCI DSS Req 12.10; Visa/MC ADC |
| **PFI (Forensic Investigator)** | Scheme-mandated ADC investigation | Engage **within 5 days** per scheme rules | CISO | Scheme ADC program |
| **AMLO (ปปง.)** | Incident linked to money laundering/fraud meeting STR criteria | Per `07-sar-str-procedure.md` | MLRO | AMLA |
| **Law enforcement** | Cybercrime / ransomware / extortion | Per Legal advice | Legal + CISO | — |
| **Affected merchants** | Merchant/cardholder data of a merchant affected | Without undue delay, per contract/SLA | Merchant Support | Service agreement |
| **Cyber insurer** | Policy conditions met | Per policy notice conditions (A7) | CFO/CISO | Policy |

**Minimum content for PDPC / data-subject notice:** nature of the breach and the categories/volume of data affected, likely consequences, measures taken/to be taken to address and mitigate, and a contact point (DPO). Where full facts are not yet known, a **phased notification** is permitted.

---

## 6. Communications Plan

| Channel | Audience | Approver | Note |
|---------|----------|----------|------|
| **Internal war room** (encrypted out-of-band, e.g. bridge line) | CSIRT | IC | Never use a possibly-compromised system |
| **Executive/Board brief** | CEO, Board | IC | SEV-1 notified immediately |
| **Regulator notice** (PDPC, BOT) | Regulators | DPO/CCO + Legal | Official documents only |
| **Scheme/Acquirer notice** | Sponsor bank, Visa/MC | CISO + Compliance | Via scheme-mandated channel |
| **Merchant notice** | Merchants | Merchant Support Lead | Templates via Legal |
| **Public/Press statement** | Media/public | IC + Legal + Comms | Always Legal-approved first |

**Communications principles:** (1) **out-of-band** — never discuss the incident through a possibly-breached system; (2) **single source of truth** — Comms is the sole publisher; (3) **no speculation** — only confirmed facts; (4) prevent **tipping-off** when an AML/investigation dimension exists; (5) all external communications go through Legal.

---

## 7. Evidence & Forensics

- **Chain of custody** — Record custodian, time, and access for every artifact; keep hashes for integrity.
- **Preserve before eradicate** — Snapshot systems/disk/memory and copy logs before remediation.
- **Log integrity** — `audit_log` is append-only (per `ARCHITECTURE.md`); ship logs to a write-once SIEM.
- **PFI engagement** — When scheme-mandated, use an approved PFI (A3); do not destroy evidence before PFI review.
- **Evidence retention** — At minimum per `11-data-retention-deletion.md` and per Legal/PFI guidance (typically ≥ 12 months or until the case/investigation concludes).

---

## 8. Testing & Maintenance (PCI Req 12.10.2)

- **Tabletop exercise** at least **annually** (scenarios: PAN exfiltration, ransomware, key compromise, insider).
- **Review and update the plan** at least annually and after every SEV-1/SEV-2 incident (Req 12.10.6).
- **Train the CSIRT** and run organization-wide **security awareness** annually.
- **24×7 availability** of response personnel (on-call roster / MDR — A4) per Req 12.10.3.
- **Verify detection controls** (IDS/IPS, FIM, anomaly) periodically (Req 12.10.5).
- Exercise results and RCAs are kept as evidence for QSA assessment and BOT reporting.
