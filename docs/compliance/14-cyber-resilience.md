# แผนความมั่นคงปลอดภัยไซเบอร์ (Cyber Resilience) (ไทย)

> เอกสารเลขที่ **14** ในชุดเอกสาร Compliance สำหรับการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** และคู่ขนานกับ **PCI-DSS Level 1**
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อยื่นขออนุมัติ (submission draft) · เวอร์ชัน 0.1 · วันที่ปรับปรุง 2026-07-22
> **เจ้าของเอกสาร:** ประธานเจ้าหน้าที่ความมั่นคงปลอดภัยสารสนเทศ (CISO) ของ **[บริษัท / Company]** โดยรายงานต่อคณะกรรมการบริหารความเสี่ยง (Risk Committee)
>
> เอกสารนี้เป็นนโยบายภายในและเอกสารประกอบคำขอใบอนุญาต **มิใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจทานโดยที่ปรึกษากฎหมาย, CISO และ QSA ก่อนบังคับใช้จริง เนื่องจากประกาศ/หลักเกณฑ์ของ ธปท. และมาตรฐาน PCI-DSS อาจปรับปรุงได้

---

## บทสรุปสำหรับผู้บริหาร (Executive Summary)

[บริษัท / Company] ในฐานะผู้ให้บริการรับชำระเงินด้วยบัตร (acquiring payment gateway) ถือครองข้อมูลบัตร (Cardholder Data — CHD) และประมวลผลธุรกรรมทางการเงินปริมาณสูง จึงตกเป็นเป้าหมายลำดับต้นของภัยคุกคามไซเบอร์ เอกสารฉบับนี้กำหนด **แผนความมั่นคงปลอดภัยไซเบอร์ (Cyber Resilience Plan)** ที่ครอบคลุมวงจรครบถ้วนตามหลัก **Identify → Protect → Detect → Respond → Recover** (NIST CSF 2.0) และสอดคล้องกับ **PCI-DSS v4.0** และแนวปฏิบัติด้าน IT Risk / Cyber Resilience ของ ธปท. โดยมี 5 เสาหลัก:

1. **การบริหารจัดการภัยคุกคาม (Threat Management)** — threat intelligence, threat modeling, การบริหารพื้นผิวการโจมตี (attack surface)
2. **ศูนย์เฝ้าระวังความมั่นคงปลอดภัย (SOC) และการเฝ้าติดตาม (Monitoring)** — SIEM, log ครบตาม PCI Req 10, การเฝ้าระวัง 24×7
3. **การบริหารจัดการช่องโหว่ (Vulnerability Management)** — scan, ASV, patch SLA, การจัดลำดับความรุนแรง
4. **การทดสอบเจาะระบบและ Red Team** — penetration test ประจำปี, purple team, breach & attack simulation
5. **การกู้คืนและความต่อเนื่อง (Recovery & Resilience)** — incident response, RTO/RPO, DR drill, ransomware playbook

> **หลักการชี้นำ (Guiding Principle):** *Assume breach* — ออกแบบระบบโดยสมมติว่าถูกเจาะได้เสมอ เน้น defense-in-depth, least privilege, การตรวจจับที่รวดเร็ว และความสามารถในการกู้คืนที่พิสูจน์ได้ (proven recoverability)

---

## 1. ฐานกฎหมายและมาตรฐานอ้างอิง

| กฎหมาย / มาตรฐาน | สาระสำคัญที่เกี่ยวข้องกับแผนนี้ |
|---|---|
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** | เงื่อนไขการมีระบบบริหารความเสี่ยงด้าน IT และ cyber resilience เป็นเงื่อนไขใบอนุญาต กำกับโดย **ธปท.** |
| **ประกาศ ธปท. ว่าด้วยหลักเกณฑ์การกำกับดูแลความเสี่ยงด้านเทคโนโลยีสารสนเทศ (IT Risk) และ Cyber Resilience** | การกำกับดูแล (governance), การควบคุมภายใน, การรายงานเหตุการณ์ทางไซเบอร์ต่อ ธปท. |
| **PCI-DSS v4.0** | Req 1 (network security), Req 5 (anti-malware), Req 6 (secure development/patch), Req 10 (logging & monitoring), Req 11 (scan/pentest), Req 12.10 (incident response) |
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — กำกับโดย **คณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC)** | การแจ้งเหตุละเมิดข้อมูลส่วนบุคคลต่อ PDPC ภายใน 72 ชั่วโมง (มาตรา 37(4)) |
| **พ.ร.บ. การรักษาความมั่นคงปลอดภัยไซเบอร์ พ.ศ. 2562** | กรอบการจัดการภัยคุกคามไซเบอร์และการประสานงานกับหน่วยงานที่เกี่ยวข้อง (สกมช./NCSA) หากเข้าข่าย CII |
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (ปปง./AMLO)** | ความเชื่อมโยงกับการเฝ้าระวังธุรกรรมและการรักษาความสมบูรณ์ของ audit trail |
| **NIST CSF 2.0 / NIST SP 800-61r2 / ISO/IEC 27001:2022 / ISO 22301** | กรอบอ้างอิงสากลด้าน cyber resilience, incident handling และ business continuity |
| **EMV 3-D Secure (3DS) 2.x** | การควบคุมความเสี่ยงการฉ้อโกงในชั้น authentication ของธุรกรรมบัตร |

> **[TODO/ข้อสมมติ]** ต้องยืนยันกับที่ปรึกษากฎหมายว่าประกาศ ธปท. ด้าน IT Risk / Cyber Resilience ฉบับล่าสุด ณ วันยื่นคำขอ มีเลขที่/ปีใด และปรับอ้างอิงให้ตรง รวมถึงพิจารณาว่า [บริษัท / Company] เข้าข่ายเป็นโครงสร้างพื้นฐานสำคัญทางสารสนเทศ (CII) ตาม พ.ร.บ. ไซเบอร์ฯ หรือไม่

---

## 2. โครงสร้างการกำกับดูแลและบทบาทหน้าที่ (Governance & Roles)

| บทบาท | ความรับผิดชอบหลักด้าน Cyber Resilience |
|---|---|
| **คณะกรรมการบริษัท (Board)** | อนุมัตินโยบายความมั่นคงปลอดภัยไซเบอร์, รับทราบรายงานความเสี่ยงระดับสูงรายไตรมาส |
| **คณะกรรมการบริหารความเสี่ยง (Risk Committee)** | กำกับ risk appetite, ทบทวน KRI, อนุมัติแผน remediation ของช่องโหว่ระดับ Critical |
| **CISO** | เจ้าของแผนฉบับนี้, กำกับ SOC, VM, red team, IR; รายงานตรงต่อ Risk Committee (แยกสายจาก CTO เพื่อ segregation of duties) |
| **SOC Lead / SecOps** | เฝ้าระวัง 24×7, triage alert, จัดการ incident ระดับ L1–L2 |
| **Vulnerability Management Lead** | บริหาร scan, patch SLA, ประสาน ASV และ QSA |
| **Incident Response Manager (IRM)** | สั่งการ CSIRT, ตัดสินใจ containment, ประสาน ธปท./PDPC/ปปง. |
| **DPO (Data Protection Officer)** | ประเมินและแจ้งเหตุละเมิดข้อมูลส่วนบุคคลต่อ PDPC (มาตรา 37(4)) |
| **DevSecOps / SRE** | ฝัง security ใน CI/CD, จัดการ patch โครงสร้างพื้นฐาน, ดำเนิน DR drill |
| **QSA (ภายนอก)** | ตรวจประเมิน PCI-DSS ประจำปี, ออก RoC |

> **หลักการแยกหน้าที่:** CISO ต้องมีสายรายงานที่เป็นอิสระจากทีมพัฒนา/ปฏิบัติการ (independent reporting line) เพื่อไม่ให้เกิดผลประโยชน์ทับซ้อนในการรายงานความเสี่ยง

---

## 3. เสาที่ 1 — การบริหารจัดการภัยคุกคาม (Threat Management)

### 3.1 Threat Intelligence (TI)
- สมัครแหล่งข่าวกรองภัยคุกคาม: **FS-ISAC** (ภาคการเงิน), card scheme alerts (Visa/Mastercard), CERT ในประเทศ (**ThaiCERT/สกมช.**), commercial TI feed
- นำ Indicator of Compromise (IOC) เข้าสู่ SIEM และ firewall/WAF แบบอัตโนมัติผ่าน threat feed (STIX/TAXII)
- ประเมินภัยคุกคามเฉพาะภาคชำระเงิน: **Magecart/e-skimming, credential stuffing, BIN attack, account takeover, ransomware, supply-chain attack, API abuse**

### 3.2 Threat Modeling
- ทำ threat modeling (แนวทาง **STRIDE**) ทุกครั้งที่ออกแบบ feature ใหม่ที่แตะเงินหรือ CHD ก่อนขึ้น production
- รักษา data flow diagram ของ CHD environment (CDE) ให้เป็นปัจจุบันตาม PCI Req 1.2 / 12.5.2

### 3.3 การบริหารพื้นผิวการโจมตี (Attack Surface Management)
- จัดทำ asset inventory ครบถ้วน (PCI Req 12.5.1) — ทุก system component ใน CDE มีเจ้าของและระดับความสำคัญ
- External Attack Surface Management (EASM) สแกน external IP/domain อย่างน้อยทุกสัปดาห์เพื่อหา shadow IT / exposed service

### 3.4 การจัดระดับภัยคุกคาม (Threat Severity Matrix)

| ระดับ | นิยาม | ตัวอย่าง | การตอบสนอง |
|---|---|---|---|
| **Critical (P1)** | กระทบ CHD / ระบบชำระเงินหลัก / มีการรั่วไหลข้อมูล | ransomware active, PAN exfiltration, RCE บน CDE | เรียก CSIRT ทันที, แจ้ง Board/ธปท. |
| **High (P2)** | ช่องโหว่รุนแรงที่ยัง exploit ไม่สำเร็จ | exposed admin panel, exploitable CVE (CVSS ≥ 9.0) | containment ภายใน 4 ชม. |
| **Medium (P3)** | ความเสี่ยงจำกัดวง | phishing พนักงานเดี่ยว, misconfiguration ไม่กระทบ CDE | ภายใน 24–72 ชม. |
| **Low (P4)** | ผลกระทบต่ำ | port scan, ช่องโหว่ระดับ informational | ตาม backlog |

---

## 4. เสาที่ 2 — SOC และการเฝ้าติดตาม (SOC / Monitoring)

### 4.1 รูปแบบการดำเนินงาน SOC
- **โหมดการทำงาน:** เฝ้าระวัง **24×7×365** ครอบคลุม CDE และระบบสนับสนุน
- **[TODO/ข้อสมมติ]** ต้องตัดสินใจระหว่าง **in-house SOC** หรือ **MSSP (co-managed SOC)** — เอกสารนี้ตั้งสมมติฐานเป็น **hybrid**: SIEM/analytics อยู่ในไทย (data residency) + MSSP เสริมเวรกลางคืน; ต้องระบุชื่อผู้ให้บริการจริงก่อนยื่น

### 4.2 Logging (สอดคล้อง PCI Req 10)
- เก็บ audit log ของทุก access ถึง CHD, ทุก admin action, ทุก authentication (สำเร็จ/ล้มเหลว), การเปลี่ยนแปลง config/สิทธิ์
- ตั้งเวลาให้ตรงกันด้วย **NTP** (Req 10.6) และ **ป้องกัน log ถูกแก้ไข** (write-once / append-only, ส่งไป log server แยก segment)
- **ระยะเก็บ log:** อย่างน้อย **12 เดือน** โดย **3 เดือนล่าสุดพร้อมเรียกดูทันที (immediately available)** ตาม PCI Req 10.5.1
- **ห้าม log ข้อมูลบัตร** — ไม่ log PAN เต็ม, CVV, track data, PIN (สอดคล้อง ARCHITECTURE.md §7)

### 4.3 SIEM & Detection
- SIEM รวม log จาก WAF, IDS/IPS, firewall, application (Go/Fiber structured JSON log), database, HSM/KMS, cloud audit
- **File Integrity Monitoring (FIM)** บนไฟล์สำคัญ (PCI Req 11.5) — แจ้งเตือนเมื่อมีการเปลี่ยนแปลงที่ไม่ได้รับอนุมัติ
- Detection rule ตาม **MITRE ATT&CK** — ครอบคลุมอย่างน้อย: brute force, privilege escalation, data exfiltration, anomalous DB query, การเข้าถึง vault นอกเวลาปกติ

### 4.4 SLA การตอบสนองต่อ Alert

| ความรุนแรง | เวลาตอบรับ (MTTA) | เวลาเริ่ม containment | ช่องทาง |
|---|---|---|---|
| **P1 Critical** | ≤ 15 นาที | ≤ 1 ชม. | on-call page + โทร IRM |
| **P2 High** | ≤ 30 นาที | ≤ 4 ชม. | on-call + ticket |
| **P3 Medium** | ≤ 4 ชม. | ≤ 24 ชม. | ticket |
| **P4 Low** | ≤ 1 วันทำการ | ตาม backlog | ticket |

---

## 5. เสาที่ 3 — การบริหารจัดการช่องโหว่ (Vulnerability Management)

### 5.1 กิจกรรมและความถี่

| กิจกรรม | ความถี่ | อ้างอิง |
|---|---|---|
| **Internal vulnerability scan** | อย่างน้อยทุก 3 เดือน + หลังการเปลี่ยนแปลงสำคัญ | PCI Req 11.3.1 |
| **External ASV scan** (ผู้ให้บริการที่ PCI SSC รับรอง) | ทุกไตรมาส ต้องได้ผล **passing** | PCI Req 11.3.2 |
| **Authenticated scan** | ทุก 3 เดือน | PCI Req 11.3.1.1 |
| **Dependency / SCA scan** (`govulncheck`, SBOM) | ทุก build ใน CI | PCI Req 6.3.1 |
| **SAST / DAST** | ทุก PR (SAST) / ทุก release (DAST) | PCI Req 6.2.3 |
| **Container image scan** | ทุก build (distroless base) | PCI Req 6.3.1 |

### 5.2 SLA การแก้ไขช่องโหว่ (Patch/Remediation SLA)

| ความรุนแรง (CVSS v3.1) | SLA ในระบบ CDE | SLA นอก CDE |
|---|---|---|
| **Critical (9.0–10.0)** | ≤ 14 วัน (emergency patch ทันทีหากถูก exploit) | ≤ 30 วัน |
| **High (7.0–8.9)** | ≤ 30 วัน | ≤ 60 วัน |
| **Medium (4.0–6.9)** | ≤ 90 วัน | ≤ 90 วัน |
| **Low (0.1–3.9)** | ตาม risk-based (Req 6.3.3) | ตาม backlog |

> การจัดลำดับความรุนแรงใช้ **risk ranking** ตาม PCI Req 6.3.1 (พิจารณา CVSS + exploitability + ตำแหน่งใน CDE) หากไม่สามารถ patch ในเวลาได้ ต้องมี **compensating control** ที่บันทึกและอนุมัติโดย CISO

### 5.3 Secure Development (SSDLC)
- Peer review ทุก PR, secret scanning ใน pre-commit + CI (ป้องกันคีย์หลุดตาม ARCHITECTURE.md §7)
- เทคนิคป้องกันตาม OWASP Top 10 / OWASP ASVS — parameterized SQL (sqlc), input validation, rate limit, idempotency
- แยก key management ผ่าน HSM/KMS, ไม่มีคีย์ในโค้ด/คอนฟิก (PCI Req 3, 6)

---

## 6. เสาที่ 4 — การทดสอบเจาะระบบและ Red Team

### 6.1 โปรแกรมทดสอบ

| ประเภทการทดสอบ | ความถี่ | ขอบเขต | ผู้ดำเนินการ |
|---|---|---|---|
| **External penetration test** | อย่างน้อยปีละครั้ง + หลังเปลี่ยนแปลงสำคัญ | perimeter, public API, 3DS flow | ผู้ทดสอบอิสระ (PCI Req 11.4.3) |
| **Internal penetration test** | ปีละครั้ง | segmentation, lateral movement | ทีมภายใน/บุคคลภายนอก (Req 11.4.2) |
| **Segmentation test** | อย่างน้อยทุก 6 เดือน | พิสูจน์ว่า CDE ถูกแยกจริง | Req 11.4.5 |
| **Red team exercise** | ปีละครั้ง (objective-based) | จำลอง APT: จาก phishing → CHD | Red team อิสระ |
| **Purple team** | ทุกไตรมาส | ปรับ detection rule ร่วมกับ SOC | Red + Blue team |
| **Breach & Attack Simulation (BAS)** | ต่อเนื่อง (automated) | ทดสอบ control ตาม MITRE ATT&CK | เครื่องมืออัตโนมัติ |

### 6.2 หลักการ Red Team
- ใช้แนวทาง **objective-based / TIBER-EU-style** — สมมติเป้าหมาย เช่น "เข้าถึง token vault และดึง test PAN 100 รายการ" โดยไม่ถูกตรวจจับ
- ช่องโหว่ที่พบต้อง exploit ให้เห็นจริง (Req 11.4.4) แล้ว **retest** หลัง remediation จนปิดได้
- รายงานสรุปเสนอ CISO และ Risk Committee พร้อม remediation timeline

> **[TODO/ข้อสมมติ]** ต้องคัดเลือกผู้ให้บริการ pentest/red team ที่มีคุณสมบัติ (เช่น CREST/OSCP) และเป็นอิสระจากทีมพัฒนา ยังไม่ล็อกผู้ให้บริการจริง

---

## 7. เสาที่ 5 — การตอบสนองเหตุการณ์และการกู้คืน (Incident Response & Recovery)

### 7.1 กระบวนการตอบสนองเหตุการณ์ (IR — ตาม NIST SP 800-61 & PCI Req 12.10)
**Prepare → Detect → Analyze → Contain → Eradicate → Recover → Post-incident review**

- **CSIRT** พร้อมเรียกใช้ 24×7 มี IRM เป็นผู้บัญชาการเหตุการณ์ (Incident Commander)
- Playbook เฉพาะเรื่อง: **ransomware, PAN/data breach, DDoS, account takeover, insider threat, third-party/supply-chain compromise**
- **ทดสอบ IR plan อย่างน้อยปีละครั้ง** (tabletop + technical drill) ตาม PCI Req 12.10.2

### 7.2 หน้าที่การแจ้งเหตุ (Regulatory Notification)

| หน่วยงาน | เหตุการณ์ที่ต้องแจ้ง | กรอบเวลา |
|---|---|---|
| **ธปท.** | เหตุการณ์ไซเบอร์/IT ที่กระทบบริการชำระเงินอย่างมีนัยสำคัญ | โดยเร็วตามหลักเกณฑ์ ธปท. (**[TODO]** ยืนยันกรอบเวลาตามประกาศฉบับล่าสุด) |
| **PDPC** | เหตุละเมิดข้อมูลส่วนบุคคลที่เสี่ยงต่อสิทธิเสรีภาพบุคคล | ภายใน **72 ชั่วโมง** (PDPA มาตรา 37(4)) |
| **ปปง. (AMLO)** | หากเหตุการณ์เชื่อมโยงกับธุรกรรมต้องสงสัย/ฟอกเงิน | ตามหลักเกณฑ์ ปปง. |
| **Card schemes / Sponsor bank** | สงสัยว่ามีการรั่วไหลของ account data (ADC) | ตามข้อกำหนด scheme (มักภายใน 24–72 ชม.) |
| **สกมช./ThaiCERT** | หากเข้าข่าย CII ตาม พ.ร.บ. ไซเบอร์ฯ | ตามกฎหมาย |

### 7.3 เป้าหมายการกู้คืน (Recovery Objectives)

| ตัวชี้วัด | เป้าหมาย | อ้างอิง |
|---|---|---|
| **RPO** (จุดกู้คืนข้อมูล) | ≤ 5 นาที (streaming replica) | ARCHITECTURE.md §8 |
| **RTO** (เวลากู้คืนบริการ) | ≤ 30 นาที (payment core) | ARCHITECTURE.md §8 |
| **Availability** | ≥ 99.95% (payment core) | ARCHITECTURE.md §8 |

### 7.4 ความยืดหยุ่นด้านโครงสร้างพื้นฐาน (Infrastructure Resilience)
- Multi-AZ deployment ในไทย (data residency ตาม ธปท./PDPA) + DR site
- **Immutable / air-gapped backup** และทดสอบ restore เป็นระยะ (ป้องกัน ransomware เข้ารหัส backup)
- **DR drill / failover test อย่างน้อยปีละครั้ง** พร้อมบันทึกผลและ RTO/RPO ที่วัดได้จริง
- Ledger แบบ append-only (double-entry) เป็น source of truth สำหรับ reconciliation หลังกู้คืน (สอดคล้อง ARCHITECTURE.md §2)
- **Fail closed** — เมื่อสถานะธุรกรรมไม่แน่นอน ถือว่ายังไม่สำเร็จ และ reconcile ภายหลัง

### 7.5 Ransomware Readiness (เฉพาะเรื่อง)
- ตรวจจับพฤติกรรมเข้ารหัสจำนวนมากผ่าน SIEM/EDR
- แยกส่วนเครือข่าย (segmentation) เพื่อจำกัดการแพร่กระจาย
- นโยบาย **ไม่จ่ายค่าไถ่เป็นทางเลือกแรก** — ยึดการกู้คืนจาก backup ที่สะอาด และประสานหน่วยงานบังคับใช้กฎหมาย

---

## 8. ตัวชี้วัดและการทบทวน (Metrics, KRI & Review)

| ตัวชี้วัด (KRI/KPI) | เป้าหมาย |
|---|---|
| MTTA (Critical alert) | ≤ 15 นาที |
| MTTR (Critical incident) | ≤ 4 ชม. |
| Critical patch ภายใน SLA | ≥ 95% |
| ASV quarterly scan | ผ่าน (passing) ทุกไตรมาส |
| Coverage log ของ system ใน CDE | 100% |
| ผลการทดสอบ DR drill | ผ่านตาม RTO/RPO |

- **ทบทวนแผนอย่างน้อยปีละครั้ง** และหลังเหตุการณ์สำคัญ / การเปลี่ยนแปลงกฎหมาย
- รายงานสถานะ cyber resilience ต่อ Risk Committee รายไตรมาส และต่อ Board รายปี

---

## 9. รายการข้อสมมติและสิ่งที่ต้องดำเนินการ (Assumptions & TODO)

> รวมประเด็นที่ยังพึ่งพาปัจจัยภายนอกและต้องปิดก่อนยื่นจริง

1. **[TODO]** เลือกและระบุชื่อ **QSA** ที่จะออก RoC
2. **[TODO]** ตัดสินใจ **in-house SOC vs MSSP** และระบุผู้ให้บริการ (พร้อมข้อตกลง data residency)
3. **[TODO]** เลือกผู้ให้บริการ **pentest / red team** อิสระที่มีใบรับรอง
4. **[TODO]** ยืนยันเลขที่/ปีของ **ประกาศ ธปท.** ด้าน IT Risk / Cyber Resilience ฉบับล่าสุด และกรอบเวลาการแจ้งเหตุ
5. **[TODO]** ยืนยันข้อกำหนดการแจ้งเหตุกับ **sponsor bank / card scheme** (ยังไม่สรุป sponsor bank)
6. **[ข้อสมมติ]** ทุนจดทะเบียนชำระแล้ว **50 ล้านบาท** (ต้องยืนยันตัวเลขจริงกับเอกสารเลขที่ 02)
7. **[TODO]** ประเมินสถานะ **CII** ตาม พ.ร.บ. ไซเบอร์ฯ

---

## 10. เอกสารที่เกี่ยวข้อง
- `../COMPLIANCE-TH.md` — กฎหมายไทยและขั้นตอนยื่นขอใบอนุญาต
- `../ARCHITECTURE.md` — สถาปัตยกรรม, security/PCI, NFR (RPO/RTO)
- `../ROADMAP.md` — เฟส Phase 2 (Security), Phase 4 (Certification, pentest, DR drill)
- `05-aml-kyc-cdd-policy.md`, `06-sanctions-screening.md` — การเชื่อมโยงกับ AML/audit trail

---
---

# Cyber resilience plan: threat management, SOC/monitoring, vulnerability mgmt, red-team, recovery (English)

> Document **No. 14** in the Compliance set for the **Full Acquiring payment service** license application under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT / ธปท.)**, in parallel with **PCI-DSS Level 1**.
>
> **Status:** Submission draft · Version 0.1 · Last updated 2026-07-22
> **Owner:** Chief Information Security Officer (CISO) of **[บริษัท / Company]**, reporting to the Risk Committee.
>
> This is an internal policy and a license-application supporting document — **not legal advice**. It must be reviewed by legal counsel, the CISO, and the QSA before it takes effect, as BOT notifications and the PCI-DSS standard may be updated.

---

## Executive Summary

As an acquiring payment gateway, **[บริษัท / Company]** holds Cardholder Data (CHD) and processes high-volume financial transactions, making it a priority target for cyber threats. This document sets out a **Cyber Resilience Plan** spanning the full **Identify → Protect → Detect → Respond → Recover** lifecycle (NIST CSF 2.0), aligned with **PCI-DSS v4.0** and BOT's IT Risk / Cyber Resilience guidance, across five pillars:

1. **Threat Management** — threat intelligence, threat modeling, attack surface management
2. **SOC / Monitoring** — SIEM, logging per PCI Req 10, 24×7 monitoring
3. **Vulnerability Management** — scanning, ASV, patch SLAs, severity ranking
4. **Penetration Testing & Red Team** — annual pentests, purple team, breach & attack simulation
5. **Recovery & Resilience** — incident response, RTO/RPO, DR drills, ransomware playbook

> **Guiding Principle — *Assume breach*:** design as if compromise is inevitable; emphasize defense-in-depth, least privilege, fast detection, and proven recoverability.

---

## 1. Legal Basis & Reference Standards

| Law / Standard | Relevance to this plan |
|---|---|
| **Payment Systems Act B.E. 2560** | Requires an IT risk and cyber resilience program as a licensing condition; supervised by **BOT (ธปท.)** |
| **BOT notifications on IT Risk & Cyber Resilience** | Governance, internal controls, cyber-incident reporting to BOT |
| **PCI-DSS v4.0** | Req 1 (network security), Req 5 (anti-malware), Req 6 (secure dev/patch), Req 10 (logging & monitoring), Req 11 (scan/pentest), Req 12.10 (incident response) |
| **Personal Data Protection Act B.E. 2562 (PDPA)** — enforced by the **PDPC** | Personal-data breach notification to PDPC within 72 hours (Sec. 37(4)) |
| **Cybersecurity Act B.E. 2562** | Threat-handling framework and coordination with authorities (NCSA/สกมช.) if classified as CII |
| **Anti-Money Laundering Act (AMLO / ปปง.)** | Linkage to transaction monitoring and integrity of the audit trail |
| **NIST CSF 2.0 / NIST SP 800-61r2 / ISO/IEC 27001:2022 / ISO 22301** | International references for cyber resilience, incident handling, and business continuity |
| **EMV 3-D Secure (3DS) 2.x** | Fraud risk control at the card-transaction authentication layer |

> **[TODO/Assumption]** Confirm with legal counsel the number/year of the latest BOT IT Risk / Cyber Resilience notification at filing time and align references; assess whether **[บริษัท / Company]** qualifies as Critical Information Infrastructure (CII) under the Cybersecurity Act.

---

## 2. Governance & Roles

| Role | Key cyber-resilience responsibilities |
|---|---|
| **Board of Directors** | Approves the cybersecurity policy; receives quarterly top-risk reports |
| **Risk Committee** | Governs risk appetite, reviews KRIs, approves remediation of Critical findings |
| **CISO** | Owns this plan; oversees SOC, VM, red team, IR; reports directly to the Risk Committee (independent of the CTO for segregation of duties) |
| **SOC Lead / SecOps** | 24×7 monitoring, alert triage, L1–L2 incident handling |
| **Vulnerability Management Lead** | Manages scanning, patch SLAs, liaises with ASV and QSA |
| **Incident Response Manager (IRM)** | Commands the CSIRT, decides containment, coordinates with BOT/PDPC/AMLO |
| **DPO** | Assesses and notifies personal-data breaches to PDPC (Sec. 37(4)) |
| **DevSecOps / SRE** | Embeds security in CI/CD, patches infrastructure, runs DR drills |
| **QSA (external)** | Annual PCI-DSS assessment, issues the RoC |

> **Segregation:** The CISO must have an independent reporting line separate from development/operations to avoid conflicts of interest in risk reporting.

---

## 3. Pillar 1 — Threat Management

### 3.1 Threat Intelligence (TI)
- Subscribe to threat feeds: **FS-ISAC** (financial sector), card-scheme alerts (Visa/Mastercard), local CERT (**ThaiCERT/NCSA**), commercial TI feeds
- Auto-ingest IOCs into SIEM and firewall/WAF via threat feeds (STIX/TAXII)
- Assess payment-specific threats: **Magecart/e-skimming, credential stuffing, BIN attacks, account takeover, ransomware, supply-chain attacks, API abuse**

### 3.2 Threat Modeling
- Perform threat modeling (**STRIDE**) for every new feature touching money or CHD before production
- Keep the CHD environment (CDE) data-flow diagram current per PCI Req 1.2 / 12.5.2

### 3.3 Attack Surface Management
- Maintain a complete asset inventory (PCI Req 12.5.1) — every CDE component has an owner and criticality
- Run External Attack Surface Management (EASM) weekly against external IPs/domains to detect shadow IT / exposed services

### 3.4 Threat Severity Matrix

| Level | Definition | Example | Response |
|---|---|---|---|
| **Critical (P1)** | Impacts CHD / core payments / data leakage | active ransomware, PAN exfiltration, RCE in CDE | Invoke CSIRT immediately; notify Board/BOT |
| **High (P2)** | Severe but not yet exploited | exposed admin panel, exploitable CVE (CVSS ≥ 9.0) | Contain within 4 hrs |
| **Medium (P3)** | Limited scope | single-user phishing, misconfig outside CDE | Within 24–72 hrs |
| **Low (P4)** | Low impact | port scan, informational finding | Per backlog |

---

## 4. Pillar 2 — SOC / Monitoring

### 4.1 SOC Operating Model
- **Coverage:** **24×7×365** monitoring across the CDE and supporting systems
- **[TODO/Assumption]** Decide between **in-house SOC** or **MSSP (co-managed SOC)**. This document assumes a **hybrid**: SIEM/analytics hosted in Thailand (data residency) with an MSSP augmenting overnight shifts; the actual provider must be named before filing.

### 4.2 Logging (PCI Req 10)
- Log all access to CHD, all admin actions, all authentications (success/failure), and all config/privilege changes
- Time-synchronize with **NTP** (Req 10.6) and **protect logs from tampering** (write-once / append-only, shipped to a segmented log server)
- **Retention:** at least **12 months**, with the **most recent 3 months immediately available** (PCI Req 10.5.1)
- **Never log card data** — no full PAN, CVV, track data, or PIN (consistent with ARCHITECTURE.md §7)

### 4.3 SIEM & Detection
- SIEM aggregates logs from WAF, IDS/IPS, firewall, application (Go/Fiber structured JSON logs), database, HSM/KMS, cloud audit
- **File Integrity Monitoring (FIM)** on critical files (PCI Req 11.5) — alerts on unauthorized changes
- Detection rules mapped to **MITRE ATT&CK** — at minimum: brute force, privilege escalation, data exfiltration, anomalous DB queries, out-of-hours vault access

### 4.4 Alert Response SLAs

| Severity | Ack (MTTA) | Containment start | Channel |
|---|---|---|---|
| **P1 Critical** | ≤ 15 min | ≤ 1 hr | on-call page + call IRM |
| **P2 High** | ≤ 30 min | ≤ 4 hrs | on-call + ticket |
| **P3 Medium** | ≤ 4 hrs | ≤ 24 hrs | ticket |
| **P4 Low** | ≤ 1 business day | per backlog | ticket |

---

## 5. Pillar 3 — Vulnerability Management

### 5.1 Activities & Frequency

| Activity | Frequency | Reference |
|---|---|---|
| **Internal vulnerability scan** | At least quarterly + after significant changes | PCI Req 11.3.1 |
| **External ASV scan** (PCI SSC-approved vendor) | Quarterly, must achieve a **passing** result | PCI Req 11.3.2 |
| **Authenticated scan** | Quarterly | PCI Req 11.3.1.1 |
| **Dependency / SCA scan** (`govulncheck`, SBOM) | Every build in CI | PCI Req 6.3.1 |
| **SAST / DAST** | Every PR (SAST) / every release (DAST) | PCI Req 6.2.3 |
| **Container image scan** | Every build (distroless base) | PCI Req 6.3.1 |

### 5.2 Remediation / Patch SLAs

| Severity (CVSS v3.1) | SLA in CDE | SLA outside CDE |
|---|---|---|
| **Critical (9.0–10.0)** | ≤ 14 days (emergency patch immediately if exploited) | ≤ 30 days |
| **High (7.0–8.9)** | ≤ 30 days | ≤ 60 days |
| **Medium (4.0–6.9)** | ≤ 90 days | ≤ 90 days |
| **Low (0.1–3.9)** | Risk-based (Req 6.3.3) | Per backlog |

> Severity uses a **risk ranking** per PCI Req 6.3.1 (CVSS + exploitability + position in the CDE). If a patch cannot meet SLA, a **compensating control** must be documented and approved by the CISO.

### 5.3 Secure Development (SSDLC)
- Peer review on every PR; secret scanning in pre-commit + CI (prevents key leakage per ARCHITECTURE.md §7)
- OWASP Top 10 / OWASP ASVS controls — parameterized SQL (sqlc), input validation, rate limiting, idempotency
- Key management via HSM/KMS; no keys in code/config (PCI Req 3, 6)

---

## 6. Pillar 4 — Penetration Testing & Red Team

### 6.1 Testing Program

| Test type | Frequency | Scope | Performed by |
|---|---|---|---|
| **External penetration test** | At least annually + after significant changes | perimeter, public API, 3DS flow | Independent tester (PCI Req 11.4.3) |
| **Internal penetration test** | Annually | segmentation, lateral movement | Internal/external (Req 11.4.2) |
| **Segmentation test** | At least every 6 months | prove CDE isolation | Req 11.4.5 |
| **Red team exercise** | Annually (objective-based) | simulate APT: phishing → CHD | Independent red team |
| **Purple team** | Quarterly | tune detection rules with SOC | Red + Blue team |
| **Breach & Attack Simulation (BAS)** | Continuous (automated) | test controls against MITRE ATT&CK | Automated tooling |

### 6.2 Red Team Principles
- Objective-based / TIBER-EU-style — e.g., "reach the token vault and extract 100 test PANs undetected"
- Findings must be exploited to demonstrate impact (Req 11.4.4), then **retested** after remediation until closed
- Summary reports to the CISO and Risk Committee with remediation timelines

> **[TODO/Assumption]** Select qualified pentest/red-team providers (e.g., CREST/OSCP), independent of the development team; no provider locked yet.

---

## 7. Pillar 5 — Incident Response & Recovery

### 7.1 IR Process (NIST SP 800-61 & PCI Req 12.10)
**Prepare → Detect → Analyze → Contain → Eradicate → Recover → Post-incident review**

- **CSIRT** available 24×7 with the IRM as Incident Commander
- Topic-specific playbooks: **ransomware, PAN/data breach, DDoS, account takeover, insider threat, third-party/supply-chain compromise**
- **Test the IR plan at least annually** (tabletop + technical drill) per PCI Req 12.10.2

### 7.2 Regulatory Notification Duties

| Authority | Reportable event | Timeframe |
|---|---|---|
| **BOT (ธปท.)** | Cyber/IT incident materially affecting payment services | Promptly per BOT rules (**[TODO]** confirm timeframe in latest notification) |
| **PDPC** | Personal-data breach risking individuals' rights | Within **72 hours** (PDPA Sec. 37(4)) |
| **AMLO (ปปง.)** | If linked to suspicious transactions/laundering | Per AMLO rules |
| **Card schemes / Sponsor bank** | Suspected Account Data Compromise (ADC) | Per scheme rules (often within 24–72 hrs) |
| **NCSA / ThaiCERT** | If classified as CII under the Cybersecurity Act | Per law |

### 7.3 Recovery Objectives

| Metric | Target | Reference |
|---|---|---|
| **RPO** | ≤ 5 min (streaming replica) | ARCHITECTURE.md §8 |
| **RTO** | ≤ 30 min (payment core) | ARCHITECTURE.md §8 |
| **Availability** | ≥ 99.95% (payment core) | ARCHITECTURE.md §8 |

### 7.4 Infrastructure Resilience
- Multi-AZ deployment in Thailand (data residency per BOT/PDPA) + a DR site
- **Immutable / air-gapped backups** with periodic restore tests (protects backups against ransomware encryption)
- **DR drill / failover test at least annually** with documented, measured RTO/RPO
- Append-only double-entry ledger as the source of truth for post-recovery reconciliation (per ARCHITECTURE.md §2)
- **Fail closed** — when transaction state is uncertain, treat it as not completed and reconcile later

### 7.5 Ransomware Readiness
- Detect mass-encryption behavior via SIEM/EDR
- Network segmentation to limit spread
- Policy of **no-ransom-as-first-option** — rely on clean backups and coordinate with law enforcement

---

## 8. Metrics, KRIs & Review

| Metric (KRI/KPI) | Target |
|---|---|
| MTTA (Critical alert) | ≤ 15 min |
| MTTR (Critical incident) | ≤ 4 hrs |
| Critical patches within SLA | ≥ 95% |
| ASV quarterly scan | Passing every quarter |
| Log coverage of CDE systems | 100% |
| DR drill result | Meets RTO/RPO |

- **Review the plan at least annually** and after major incidents / legal changes
- Report cyber-resilience status to the Risk Committee quarterly and to the Board annually

---

## 9. Assumptions & TODO

> External dependencies to close before actual filing

1. **[TODO]** Select and name the **QSA** that will issue the RoC
2. **[TODO]** Decide **in-house SOC vs MSSP** and name the provider (with data-residency terms)
3. **[TODO]** Select an independent, certified **pentest / red-team** provider
4. **[TODO]** Confirm the number/year of the latest **BOT notification** on IT Risk / Cyber Resilience and the incident-reporting timeframe
5. **[TODO]** Confirm notification requirements with the **sponsor bank / card scheme** (sponsor bank not yet finalized)
6. **[Assumption]** Registered paid-up capital of **THB 50 million** (verify actual figure against Document No. 02)
7. **[TODO]** Assess **CII** status under the Cybersecurity Act

---

## 10. Related Documents
- `../COMPLIANCE-TH.md` — Thai law and license-application steps
- `../ARCHITECTURE.md` — architecture, security/PCI, NFRs (RPO/RTO)
- `../ROADMAP.md` — Phase 2 (Security), Phase 4 (Certification, pentest, DR drill)
- `05-aml-kyc-cdd-policy.md`, `06-sanctions-screening.md` — linkage to AML / audit trail
