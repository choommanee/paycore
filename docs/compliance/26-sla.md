# ข้อตกลงระดับการให้บริการ (SLA) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 ต่อธนาคารแห่งประเทศไทย (ธปท.) และเป็นเอกสารแนบข้อกำหนดระดับการให้บริการ (Service Level Agreement) สำหรับผู้ค้า (Merchant)
> เอกสารชุดเดียวกันนี้ใช้ประกอบการประเมิน **PCI-DSS v4.0 Level 1** ในหัวข้อความพร้อมใช้งาน (availability) การตอบสนองเหตุการณ์ (incident response) และการจัดการผู้ให้บริการภายนอก
> เจ้าของเอกสาร: [บริษัท / Company] · เวอร์ชัน 0.1 · ปรับปรุงล่าสุด: 2026-07-22 · รอบทบทวน: ทุก 12 เดือน หรือเมื่อมีการเปลี่ยนแปลงสำคัญ

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนดคำมั่นด้านคุณภาพบริการที่ [บริษัท / Company] ("ผู้ให้บริการ") ให้ไว้ต่อผู้ค้าที่ใช้บริการ Payment Gateway โดยครอบคลุม:

- **Availability (ความพร้อมใช้งาน)** ของบริการหลักและบริการรอง
- **Latency (เวลาตอบสนอง)** ของ API สำคัญ
- **Support tiers (ระดับการสนับสนุน)** และช่องทางติดต่อ
- **Incident response times (เวลาการตอบสนองเหตุการณ์)** ตามระดับความรุนแรง
- **Service credits (เครดิตชดเชย)** เมื่อไม่เป็นไปตามเป้าหมาย

**ขอบเขต:** บริการที่ [บริษัท / Company] ควบคุมโดยตรง ได้แก่ Payment Core (authorize/capture/void/refund), Tokenization Vault, Merchant/Admin API, Webhook/Notifier, Dashboard และงาน Reconciliation/Settlement ภายใน

**นอกขอบเขต (out of scope):** เหตุขัดข้องที่มีต้นเหตุจากผู้ให้บริการต้นน้ำที่อยู่นอกการควบคุมของผู้ให้บริการ เช่น sponsor bank/acquirer, card scheme (Visa/Mastercard), ผู้ให้บริการ 3-D Secure (ACS/DS), ระบบหักบัญชีท้องถิ่น (เช่น ITMX), ผู้ให้บริการโครงข่ายของผู้ค้า, force majeure และช่วงบำรุงรักษาตามแผน (planned maintenance) ที่แจ้งล่วงหน้า

> **หมายเหตุการกำกับ:** SLA ฉบับนี้เป็นข้อผูกพันเชิงพาณิชย์ระหว่างผู้ให้บริการกับผู้ค้า และ **ไม่ลดทอน** หน้าที่ตามกฎหมายของผู้ให้บริการต่อ ธปท. (เช่น การรายงานเหตุการณ์ด้าน IT/ไซเบอร์), ต่อ ปปง./AMLO (การรายงานธุรกรรม/STR-CTR) และต่อ PDPC (การแจ้งเหตุละเมิดข้อมูลส่วนบุคคลภายใน 72 ชั่วโมงตาม PDPA)

---

## 2. นิยามและวิธีวัด (Definitions & Measurement)

| คำ | นิยาม |
|----|-------|
| **Uptime %** | (นาทีทั้งหมดในเดือน − นาที Downtime) ÷ นาทีทั้งหมดในเดือน × 100 |
| **Downtime** | ช่วงที่บริการหลักไม่สามารถประมวลผลคำขอที่ถูกต้องได้สำเร็จ วัดจากอัตราความล้มเหลว (error rate จาก HTTP 5xx/timeout ที่มีต้นเหตุจากผู้ให้บริการ) เกิน **5%** ต่อเนื่องเกิน **60 วินาที** ในหน้าต่างวัด 1 นาที |
| **รอบวัด (Measurement window)** | รายเดือนตามปฏิทิน เขตเวลา Asia/Bangkok (UTC+7) |
| **แหล่งข้อมูลวัด (System of record)** | ระบบ observability ภายใน (Prometheus metrics, OpenTelemetry traces, synthetic health check ทุก 30 วินาทีจากอย่างน้อย 2 จุดวัด) — เก็บ log ≥ 12 เดือน (สอดคล้อง PCI-DSS Req 10) |
| **Latency p99** | เปอร์เซ็นไทล์ที่ 99 ของเวลาประมวลผลฝั่งเซิร์ฟเวอร์ (server-side) ไม่รวมเวลา network ฝั่งผู้ค้า วัดที่ API Edge |
| **Planned maintenance** | งานบำรุงรักษาที่แจ้งล่วงหน้า ≥ 5 วันทำการ ดำเนินการในหน้าต่างมาตรฐาน วันอาทิตย์ 02:00–06:00 (Asia/Bangkok) ไม่นับเป็น Downtime |

---

## 3. ความพร้อมใช้งาน (Availability / Uptime)

| บริการ | เป้าหมาย Uptime/เดือน | Downtime สูงสุดต่อเดือน (โดยประมาณ) |
|--------|----------------------|-------------------------------------|
| **Payment Core API** (authorize/capture/void/refund) | **≥ 99.95%** | ~21.9 นาที |
| **Tokenization Vault** (detokenize/tokenize) | **≥ 99.95%** | ~21.9 นาที |
| **Merchant/Admin API + Dashboard** | ≥ 99.9% | ~43.8 นาที |
| **Webhook/Notifier** (การส่งครั้งแรก) | ≥ 99.5% (best-effort + retry) | — |
| **Reporting/Analytics (แบบ batch)** | ≥ 99.0% | — |

**สถาปัตยกรรมรองรับเป้าหมาย** (อ้างอิง `ARCHITECTURE.md` ข้อ 8): stateless API + horizontal scaling, HA แบบหลาย availability zone, streaming replica (RPO ≤ 5 นาที), เป้าหมายกู้คืน RTO ≤ 30 นาที, health check + auto-failover, WAF/DDoS protection

**Webhook เป็น at-least-once:** หากส่งครั้งแรกไม่สำเร็จ ระบบ retry แบบ exponential backoff อย่างน้อย 24 ชั่วโมง และมี endpoint ให้ผู้ค้าดึงสถานะย้อนหลัง (reconciliation) ความล่าช้าของ webhook ไม่ถือเป็น Downtime ของ Payment Core หากธุรกรรมสำเร็จและ query ได้

---

## 4. เวลาตอบสนอง (Latency)

เป้าหมาย latency ฝั่งเซิร์ฟเวอร์ (ไม่รวม hop ไป acquirer เว้นแต่ระบุ) วัดรายเดือน:

| Endpoint / การทำงาน | p50 | p95 | p99 |
|---------------------|-----|-----|-----|
| `POST /v1/payments` (authorize, **ไม่รวม** เวลา 3DS/acquirer) | ≤ 120 ms | ≤ 300 ms | ≤ 500 ms |
| `POST /v1/payments` (**รวม** hop ไป acquirer) | — | — | ≤ 800 ms |
| `POST /v1/payments/{id}/capture` | ≤ 100 ms | ≤ 250 ms | ≤ 450 ms |
| `POST /v1/payments/{id}/refund` | ≤ 120 ms | ≤ 300 ms | ≤ 550 ms |
| Tokenize / Detokenize (Vault) | ≤ 40 ms | ≤ 120 ms | ≤ 200 ms |
| Read API (GET payment/report) | ≤ 60 ms | ≤ 150 ms | ≤ 300 ms |
| Webhook dispatch (ตั้งแต่เหตุการณ์ถึงส่งครั้งแรก) | ≤ 2 s | ≤ 5 s | ≤ 10 s |

> เวลาที่ขึ้นกับผู้ให้บริการภายนอก (3-D Secure challenge, authorization ของ issuer) อยู่นอกการควบคุมของ latency SLA แต่จะรายงานแยกเพื่อความโปร่งใส

---

## 5. ระดับการสนับสนุน (Support Tiers)

| แพ็กเกจ | ช่องทาง | เวลาให้บริการ | ผู้ค้าเป้าหมาย |
|---------|--------|--------------|----------------|
| **Standard** | อีเมล, ระบบ ticket, เอกสาร/สถานะระบบ (status page) | จันทร์–ศุกร์ 09:00–18:00 (เว้นวันหยุดธนาคาร) | ผู้ค้าทั่วไป |
| **Business** | + โทรศัพท์สายด่วน, ช่อง chat | 07:00–22:00 ทุกวัน | ผู้ค้าปริมาณกลาง |
| **Enterprise** | + Technical Account Manager (TAM), Slack/Line ร่วม, on-call | 24×7×365 | ผู้ค้าปริมาณสูง/mission-critical |

**เหตุการณ์ระดับ Sev-1/Sev-2 (ดูข้อ 6) รับแจ้งและตอบสนอง 24×7 สำหรับทุกแพ็กเกจ** ผ่านสายด่วนเหตุฉุกเฉิน — โดยไม่ขึ้นกับ tier — เพราะเป็นเรื่องความมั่นคงของระบบชำระเงิน

### บทบาทและความรับผิดชอบ (RACI ย่อ)

| บทบาท | หน้าที่ |
|-------|--------|
| **L1 Support** | รับ ticket, จำแนกระดับเบื้องต้น, ตอบคำถามใช้งาน, escalate |
| **L2 Engineering (on-call)** | วิเคราะห์เชิงเทคนิค, แก้ไขเหตุการณ์, ประสาน acquirer/vendor |
| **Incident Commander (IC)** | ควบคุมเหตุการณ์ Sev-1/Sev-2, ตัดสินใจ, สื่อสาร stakeholder |
| **SRE / Infra** | HA, failover, DR, กู้คืนระบบ |
| **Compliance/DPO** | ประเมินหน้าที่รายงาน ธปท./ปปง./PDPC, คุมเส้นตายทางกฎหมาย |
| **Comms Lead** | อัปเดต status page และแจ้งผู้ค้า |

---

## 6. การจำแนกระดับและเวลาตอบสนองเหตุการณ์ (Incident Severity & Response)

| ระดับ | คำนิยาม | ตัวอย่าง | เวลาตอบรับ (Acknowledge) | เวลาอัปเดต | เป้าหมายแก้ไข/บรรเทา (Mitigate) |
|-------|--------|----------|--------------------------|-----------|-------------------------------|
| **Sev-1 (Critical)** | บริการชำระเงินหลักล่มทั้งระบบ หรือสงสัยเหตุละเมิดข้อมูลบัตร/ข้อมูลส่วนบุคคล | authorize ล้มเหลวทั้งระบบ, Vault ไม่ตอบสนอง, สงสัย data breach | **≤ 15 นาที (24×7)** | ทุก 30 นาที | ≤ 4 ชั่วโมง |
| **Sev-2 (High)** | บริการหลักบางส่วนเสื่อม กระทบผู้ค้าจำนวนมาก | error rate สูงกว่าปกติมาก, latency เกินเป้าอย่างรุนแรง, webhook ค้างเป็นวงกว้าง | **≤ 30 นาที (24×7)** | ทุก 60 นาที | ≤ 8 ชั่วโมง |
| **Sev-3 (Medium)** | กระทบจำกัด มี workaround | ฟีเจอร์รองบางส่วนขัดข้อง, dashboard ช้า | ≤ 4 ชั่วโมงทำการ | รายวัน | ≤ 3 วันทำการ |
| **Sev-4 (Low)** | ผลกระทบเล็กน้อย/คำถามทั่วไป | ข้อสงสัยการใช้งาน, คำขอเชิงข้อมูล | ≤ 1 วันทำการ | ตามตกลง | ตาม backlog |

### กระบวนการตอบสนองเหตุการณ์ (Incident Response Procedure)

1. **Detect** — ตรวจพบจาก alert อัตโนมัติ (SLO burn-rate, error budget), synthetic monitor หรือแจ้งจากผู้ค้า
2. **Triage & Declare** — L1/on-call จำแนกระดับ; Sev-1/Sev-2 เรียก Incident Commander และเปิด war room ทันที
3. **Contain & Mitigate** — failover/rollback/circuit breaker (fail closed ตามหลักการใน `ARCHITECTURE.md`); หากสงสัยการรั่วไหลของข้อมูลบัตร ให้ **isolate** ระบบและปฏิบัติตาม PCI-DSS Req 12.10 (Incident Response Plan)
4. **Communicate** — อัปเดต status page และแจ้งผู้ค้าตามความถี่ในตาราง; Comms Lead เป็นเจ้าของ
5. **Resolve & Recover** — ยืนยันบริการกลับสู่ปกติ + reconcile ledger เพื่อความถูกต้องของธุรกรรม
6. **Post-Incident Review (PIR)** — สำหรับ Sev-1/Sev-2 จัดทำ blameless RCA ภายใน **5 วันทำการ** พร้อม corrective action และเจ้าของงาน

### หน้าที่รายงานตามกฎหมาย (แยกจาก SLA ทางพาณิชย์)

| หน่วยงาน | เหตุการณ์ | กรอบเวลา |
|----------|-----------|----------|
| **ธปท.** | เหตุขัดข้องด้าน IT/ไซเบอร์ที่มีนัยสำคัญต่อบริการชำระเงิน | แจ้งเบื้องต้นโดยเร็วตามหลักเกณฑ์ประกาศ ธปท. (ดู TODO ด้านล่าง) แล้วตามด้วยรายงานฉบับสมบูรณ์ |
| **PDPC** | เหตุละเมิดข้อมูลส่วนบุคคลที่เสี่ยงต่อสิทธิของเจ้าของข้อมูล | **ภายใน 72 ชั่วโมง** ตาม PDPA มาตรา 37(4) |
| **ปปง./AMLO** | ธุรกรรมที่มีเหตุอันควรสงสัย (STR) / ธุรกรรมเงินสด (CTR) | ตามกรอบ พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน |
| **Card scheme / Sponsor bank** | เหตุกระทบข้อมูลบัตร (Account Data Compromise) | ตามข้อกำหนดสัญญากับ scheme (โดยทั่วไปทันที) |

---

## 7. เครดิตชดเชย (Service Credits)

หากค่า **Payment Core Uptime** รายเดือนต่ำกว่าเป้า **99.95%** ผู้ค้าที่มีสิทธิ์จะได้รับเครดิตคำนวณจากค่าธรรมเนียมบริการรายเดือน (Monthly Service Fee, MSF) ของเดือนที่เกิดเหตุ:

| Uptime รายเดือนจริง | เครดิต (% ของ MSF เดือนนั้น) |
|---------------------|------------------------------|
| < 99.95% ถึง ≥ 99.90% | 5% |
| < 99.90% ถึง ≥ 99.50% | 10% |
| < 99.50% ถึง ≥ 99.00% | 25% |
| < 99.00% | 50% |

**เพดานรวม:** เครดิตรวมต่อเดือนไม่เกิน **50%** ของ MSF เดือนนั้น

**เงื่อนไขการขอเครดิต:**
- ผู้ค้าต้องยื่นคำขอผ่านระบบ ticket ภายใน **30 วัน** นับจากสิ้นเดือนที่เกิดเหตุ พร้อมรหัสธุรกรรม/ช่วงเวลาที่ได้รับผลกระทบ
- เครดิตเป็นรูปแบบ **หักลดค่าบริการรอบถัดไป** เท่านั้น ไม่คืนเป็นเงินสด
- ยกเว้นไม่นับรวมเป็น Downtime: planned maintenance (แจ้งล่วงหน้า), เหตุจากผู้ให้บริการต้นน้ำ/นอกขอบเขต (ข้อ 1), การใช้งานผิด/เกิน rate limit ของผู้ค้า, force majeure
- เครดิตเป็น **มาตรการเยียวยาเดียว (sole remedy)** สำหรับการไม่บรรลุเป้า Uptime ภายใต้ SLA นี้

---

## 8. ข้อยกเว้นและความรับผิดชอบของผู้ค้า

- ผู้ค้าต้องตั้งค่า retry/idempotency อย่างถูกต้อง (ใช้ `Idempotency-Key` ตามคู่มือ) และรองรับ webhook แบบ at-least-once (idempotent consumer)
- ผู้ค้าต้องเก็บรักษา API key อย่างปลอดภัยและใช้ TLS 1.2+ ในการเชื่อมต่อ
- ผู้ค้าต้องไม่ส่ง/จัดเก็บ PAN, CVV, PIN, full track ผ่านระบบตนเอง (ใช้ payment token ฝั่ง client) ตามข้อห้ามใน `ARCHITECTURE.md` ข้อ 6
- การไม่ปฏิบัติตามอาจทำให้ผู้ค้าเสียสิทธิ์เครดิตในเหตุการณ์ที่เกี่ยวข้อง

---

## 9. การกำกับ ทบทวน และรายงาน (Governance & Reporting)

- **รายงานประจำเดือน:** ค่า Uptime, latency percentile, สรุปเหตุการณ์ และสถานะเครดิต เผยแพร่ผ่าน Dashboard
- **ทบทวน SLA:** อย่างน้อยปีละ 1 ครั้ง หรือเมื่อเปลี่ยน sponsor bank/สถาปัตยกรรม/หลักเกณฑ์ ธปท.
- **สอดคล้อง PCI-DSS v4.0:** การเก็บ log (Req 10), incident response (Req 12.10), การจัดการผู้ให้บริการภายนอก (Req 12.8) และการทดสอบ (quarterly ASV scan, annual penetration test)
- **การจัดการผู้ให้บริการภายนอก:** สัญญากับ sponsor bank/acquirer, ผู้ให้บริการ 3DS, HSM/KMS และ QSA ต้องมี back-to-back SLA ที่รองรับคำมั่นข้างต้น

---

## 10. สมมติฐานและรายการที่ต้องดำเนินการ (Assumptions / TODO)

> **⚠️ กล่องสมมติฐาน — ต้องยืนยันก่อนยื่น/ก่อนมีผลผูกพัน:**
>
> - **[TODO] Sponsor bank / acquirer:** ยังไม่สรุปคู่สัญญา — ค่า availability/latency ที่รวม hop ต้นน้ำ (ข้อ 3–4) และหน้าที่แจ้งเหตุ Account Data Compromise (ข้อ 6) จะต้องปรับให้ตรงกับ SLA ของ sponsor bank ที่เลือกจริง
> - **[TODO] QSA vendor:** ยังไม่เลือกผู้ประเมิน PCI-DSS Level 1 — กรอบ incident response/logging ต้องผ่านการตรวจและอ้าง RoC เมื่อได้แล้ว
> - **[TODO] หลักเกณฑ์รายงานเหตุ IT/ไซเบอร์ของ ธปท.:** กรอบเวลาแจ้งเหตุที่แน่นอนต้องอ้างอิงประกาศ/หนังสือเวียน ธปท. ฉบับล่าสุดที่บังคับใช้ ณ เวลายื่น — ให้ที่ปรึกษากฎหมายด้านใบอนุญาตตรวจสอบ
> - **[TODO] ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท:** ต้องยืนยันว่าชำระครบและคงไว้ ≥ 75% ตลอดการดำเนินงาน (เงื่อนไข Full Acquiring) — ไม่ใช่ตัวเลขที่กรอกในเอกสารนี้เพื่อการอ้างอิงเท่านั้น
> - **[TODO] ค่าธรรมเนียมบริการ (MSF) และเพดานเครดิต:** ตัวเลข % เครดิตในข้อ 7 เป็นค่าตั้งต้น ต้องสอดคล้องกับสัญญาพาณิชย์และการอนุมัติภายในก่อนมีผลผูกพัน

---
---

# Service Level Agreement: uptime, latency, support tiers, incident response times, credits (English)

> Supporting document for the **Electronic Payment Acquiring Service (Full Acquiring)** license application under the Payment Systems Act B.E. 2560 to the Bank of Thailand (BOT / ธปท.), and the Service Level Agreement schedule offered to Merchants.
> The same document supports the **PCI-DSS v4.0 Level 1** assessment for availability, incident response, and third-party service provider management.
> Document owner: [บริษัท / Company] · Version 0.1 · Last updated: 2026-07-22 · Review cycle: every 12 months or upon material change.

---

## 1. Purpose and Scope

This document defines the service quality commitments [บริษัท / Company] ("the Provider") makes to Merchants using its Payment Gateway, covering:

- **Availability / Uptime** of primary and secondary services
- **Latency** of critical API operations
- **Support tiers** and contact channels
- **Incident response times** by severity
- **Service credits** when targets are missed

**In scope:** services under the Provider's direct control — Payment Core (authorize/capture/void/refund), Tokenization Vault, Merchant/Admin API, Webhook/Notifier, Dashboard, and internal Reconciliation/Settlement.

**Out of scope:** failures originating from upstream providers outside the Provider's control — sponsor bank/acquirer, card schemes (Visa/Mastercard), 3-D Secure providers (ACS/DS), local clearing rails (e.g., ITMX), the Merchant's own network, force majeure, and announced planned maintenance windows.

> **Regulatory note:** This SLA is a commercial commitment between Provider and Merchant and does **not** diminish the Provider's statutory obligations to the BOT (e.g., IT/cyber incident reporting), to AMLO/ปปง. (transaction reporting / STR-CTR), or to the PDPC (personal data breach notification within 72 hours under the PDPA).

---

## 2. Definitions & Measurement

| Term | Definition |
|------|-----------|
| **Uptime %** | (Total minutes in month − Downtime minutes) ÷ Total minutes in month × 100 |
| **Downtime** | Period during which the primary service cannot successfully process valid requests, measured as Provider-caused error rate (5xx/timeouts) exceeding **5%** for more than **60 continuous seconds** within a 1-minute window |
| **Measurement window** | Calendar month, Asia/Bangkok time zone (UTC+7) |
| **System of record** | Internal observability stack (Prometheus metrics, OpenTelemetry traces, synthetic health checks every 30 s from ≥ 2 vantage points) — logs retained ≥ 12 months (aligned with PCI-DSS Req 10) |
| **Latency p99** | 99th percentile of server-side processing time (excluding Merchant-side network), measured at the API Edge |
| **Planned maintenance** | Work announced ≥ 5 business days ahead, performed in the standard window Sunday 02:00–06:00 (Asia/Bangkok); not counted as Downtime |

---

## 3. Availability / Uptime

| Service | Monthly Uptime target | Max Downtime/month (approx.) |
|---------|-----------------------|------------------------------|
| **Payment Core API** (authorize/capture/void/refund) | **≥ 99.95%** | ~21.9 min |
| **Tokenization Vault** (detokenize/tokenize) | **≥ 99.95%** | ~21.9 min |
| **Merchant/Admin API + Dashboard** | ≥ 99.9% | ~43.8 min |
| **Webhook/Notifier** (first delivery) | ≥ 99.5% (best-effort + retry) | — |
| **Reporting/Analytics (batch)** | ≥ 99.0% | — |

**Architecture supports these targets** (per `ARCHITECTURE.md` §8): stateless API + horizontal scaling, multi-AZ HA, streaming replica (RPO ≤ 5 min), recovery target RTO ≤ 30 min, health checks + auto-failover, WAF/DDoS protection.

**Webhooks are at-least-once:** on first-delivery failure the system retries with exponential backoff for ≥ 24 hours and exposes a pull endpoint for reconciliation. Webhook lag is not counted as Payment Core Downtime if the transaction succeeded and is queryable.

---

## 4. Latency

Server-side latency targets (excluding the acquirer hop unless stated), measured monthly:

| Endpoint / Operation | p50 | p95 | p99 |
|----------------------|-----|-----|-----|
| `POST /v1/payments` (authorize, **excl.** 3DS/acquirer time) | ≤ 120 ms | ≤ 300 ms | ≤ 500 ms |
| `POST /v1/payments` (**incl.** acquirer hop) | — | — | ≤ 800 ms |
| `POST /v1/payments/{id}/capture` | ≤ 100 ms | ≤ 250 ms | ≤ 450 ms |
| `POST /v1/payments/{id}/refund` | ≤ 120 ms | ≤ 300 ms | ≤ 550 ms |
| Tokenize / Detokenize (Vault) | ≤ 40 ms | ≤ 120 ms | ≤ 200 ms |
| Read API (GET payment/report) | ≤ 60 ms | ≤ 150 ms | ≤ 300 ms |
| Webhook dispatch (event → first send) | ≤ 2 s | ≤ 5 s | ≤ 10 s |

> Times dependent on external providers (3-D Secure challenge, issuer authorization) fall outside the latency SLA but are reported separately for transparency.

---

## 5. Support Tiers

| Package | Channels | Hours | Target Merchants |
|---------|----------|-------|------------------|
| **Standard** | Email, ticketing, docs/status page | Mon–Fri 09:00–18:00 (excl. bank holidays) | General merchants |
| **Business** | + Phone hotline, chat | 07:00–22:00 daily | Mid-volume merchants |
| **Enterprise** | + Technical Account Manager (TAM), shared Slack/Line, on-call | 24×7×365 | High-volume / mission-critical |

**Sev-1/Sev-2 incidents (see §6) are received and responded to 24×7 for all packages** via the emergency hotline — regardless of tier — given the criticality of payment infrastructure.

### Roles & Responsibilities (RACI summary)

| Role | Responsibility |
|------|----------------|
| **L1 Support** | Receive tickets, initial triage, usage questions, escalate |
| **L2 Engineering (on-call)** | Technical analysis, remediation, coordinate with acquirer/vendors |
| **Incident Commander (IC)** | Command Sev-1/Sev-2 incidents, decisions, stakeholder comms |
| **SRE / Infra** | HA, failover, DR, system recovery |
| **Compliance/DPO** | Assess BOT/AMLO/PDPC reporting duties, own legal deadlines |
| **Comms Lead** | Update status page and notify merchants |

---

## 6. Incident Severity & Response

| Severity | Definition | Examples | Acknowledge | Update cadence | Mitigation target |
|----------|-----------|----------|-------------|----------------|-------------------|
| **Sev-1 (Critical)** | Full outage of core payment service, or suspected card/personal data breach | Total authorize failure, Vault unresponsive, suspected data breach | **≤ 15 min (24×7)** | Every 30 min | ≤ 4 hours |
| **Sev-2 (High)** | Partial degradation of core service affecting many merchants | Severely elevated error rate, severe latency breach, widespread webhook backlog | **≤ 30 min (24×7)** | Every 60 min | ≤ 8 hours |
| **Sev-3 (Medium)** | Limited impact with workaround | Secondary feature fault, slow dashboard | ≤ 4 business hours | Daily | ≤ 3 business days |
| **Sev-4 (Low)** | Minor impact / general query | Usage questions, information requests | ≤ 1 business day | As agreed | Per backlog |

### Incident Response Procedure

1. **Detect** — automated alerts (SLO burn-rate, error budget), synthetic monitors, or merchant reports
2. **Triage & Declare** — L1/on-call assigns severity; Sev-1/Sev-2 invoke the Incident Commander and open a war room immediately
3. **Contain & Mitigate** — failover/rollback/circuit breaker (fail-closed per `ARCHITECTURE.md`); on suspected cardholder data compromise, **isolate** systems and follow PCI-DSS Req 12.10 (Incident Response Plan)
4. **Communicate** — update status page and notify merchants at the cadence above; Comms Lead owns this
5. **Resolve & Recover** — confirm service restoration and reconcile the ledger for transaction integrity
6. **Post-Incident Review (PIR)** — for Sev-1/Sev-2, produce a blameless RCA within **5 business days** with corrective actions and owners

### Statutory reporting duties (separate from the commercial SLA)

| Authority | Event | Timeline |
|-----------|-------|----------|
| **BOT (ธปท.)** | Material IT/cyber incident affecting payment service | Initial notification promptly per applicable BOT notification (see TODO), followed by full report |
| **PDPC** | Personal data breach risking data subjects' rights | **Within 72 hours** per PDPA §37(4) |
| **AMLO (ปปง.)** | Suspicious transaction (STR) / cash transaction (CTR) | Per the Anti-Money Laundering Act framework |
| **Card scheme / Sponsor bank** | Account Data Compromise event | Per scheme contract (typically immediate) |

---

## 7. Service Credits

If monthly **Payment Core Uptime** falls below the **99.95%** target, eligible Merchants receive a credit calculated against that month's Monthly Service Fee (MSF):

| Actual monthly Uptime | Credit (% of that month's MSF) |
|-----------------------|--------------------------------|
| < 99.95% to ≥ 99.90% | 5% |
| < 99.90% to ≥ 99.50% | 10% |
| < 99.50% to ≥ 99.00% | 25% |
| < 99.00% | 50% |

**Aggregate cap:** total monthly credits shall not exceed **50%** of that month's MSF.

**Credit conditions:**
- The Merchant must submit a claim via the ticketing system within **30 days** of month-end, with affected transaction IDs / time windows.
- Credits are applied as a **discount on the next billing cycle** only; no cash refund.
- Excluded from Downtime: planned maintenance (announced), upstream/out-of-scope causes (§1), Merchant misuse / rate-limit breaches, and force majeure.
- Credits are the Merchant's **sole remedy** for failure to meet the Uptime target under this SLA.

---

## 8. Exclusions and Merchant Responsibilities

- Merchants must correctly configure retry/idempotency (use `Idempotency-Key` per the guide) and consume webhooks at-least-once (idempotent consumer).
- Merchants must safeguard API keys and connect over TLS 1.2+.
- Merchants must not transmit/store PAN, CVV, PIN, or full track through their own systems (use client-side payment tokens), per the prohibition in `ARCHITECTURE.md` §6.
- Non-compliance may forfeit credit eligibility for the related incidents.

---

## 9. Governance, Review & Reporting

- **Monthly report:** Uptime, latency percentiles, incident summary, and credit status published via the Dashboard.
- **SLA review:** at least annually, or upon change of sponsor bank / architecture / BOT rules.
- **PCI-DSS v4.0 alignment:** log retention (Req 10), incident response (Req 12.10), third-party service provider management (Req 12.8), and testing (quarterly ASV scan, annual penetration test).
- **Third-party management:** contracts with sponsor bank/acquirer, 3DS provider, HSM/KMS, and QSA must carry back-to-back SLAs supporting the commitments above.

---

## 10. Assumptions / TODO

> **⚠️ Assumptions box — confirm before submission / before this becomes binding:**
>
> - **[TODO] Sponsor bank / acquirer:** counterparty not yet finalized — availability/latency figures that include the upstream hop (§3–4) and the Account Data Compromise notification duty (§6) must be aligned to the chosen sponsor bank's actual SLA.
> - **[TODO] QSA vendor:** the PCI-DSS Level 1 assessor is not yet selected — the incident-response/logging framework must pass assessment and cite the resulting RoC.
> - **[TODO] BOT IT/cyber incident reporting rules:** the exact notification timeline must reference the latest BOT notification/circular in force at submission — to be verified by licensing counsel.
> - **[TODO] Paid-up registered capital of THB 50M:** must be confirmed as fully paid and maintained at ≥ 75% throughout operations (Full Acquiring condition) — the figure herein is for reference only.
> - **[TODO] Monthly Service Fee (MSF) and credit cap:** the credit % figures in §7 are defaults and must align with the commercial contract and internal approval before becoming binding.
