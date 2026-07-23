# แผนความต่อเนื่องทางธุรกิจและกู้คืนระบบ (BCP/DR) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **Acquiring Service** ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560
> (ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท) และมาตรฐาน **PCI-DSS v4.0 Level 1**
> เอกสารในชุด: `docs/compliance/15-bcp-dr.md` · เวอร์ชัน 0.1
> **หมายเหตุ:** เอกสารนี้เป็นนโยบายและขั้นตอนปฏิบัติภายในของ **[บริษัท / Company]** ไม่ใช่คำแนะนำ
> ทางกฎหมาย ควรให้ที่ปรึกษากฎหมาย ผู้เชี่ยวชาญด้าน IT resilience และ QSA ตรวจทานก่อนยื่นจริง อ้างอิง
> สถาปัตยกรรมใน `ARCHITECTURE.md` (ข้อ 8 Non-functional / RPO-RTO) และ timeline ใน `ROADMAP.md` (Phase 4 DR drill)

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนด **แผนความต่อเนื่องทางธุรกิจ (Business Continuity Plan — BCP)** และ **แผนกู้คืนระบบจากภัยพิบัติ
(Disaster Recovery — DR)** ของ **[บริษัท / Company]** ในฐานะผู้ให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์
(Full Acquiring) เพื่อ:

- ให้บริการ authorization / capture / refund / settlement **ต่อเนื่อง** และกลับสู่ภาวะปกติภายในเวลาที่กำหนด
  (RTO/RPO) เมื่อเกิดเหตุขัดข้อง — สอดคล้องเป้า Availability ≥ 99.95% ของ payment core
- ปฏิบัติตามหลักเกณฑ์ **ธนาคารแห่งประเทศไทย (ธปท.)** ด้าน IT risk, cyber resilience และ business continuity
  สำหรับผู้ประกอบธุรกิจภายใต้ พ.ร.บ. ระบบการชำระเงิน
- คุ้มครองความมั่นคงปลอดภัยและความพร้อมใช้ของข้อมูลตาม **PCI-DSS v4.0** (โดยเฉพาะ Req 12.10 Incident Response
  และ Req 10 Logging) และ **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** ในกรณีเหตุกระทบข้อมูลส่วนบุคคล
- รักษาพันธะต่อ card scheme (Visa/Mastercard), sponsor bank และผู้ค้า (merchant) ตาม SLA

**ขอบเขต (Scope) ครอบคลุม:**

| ด้าน | รวมอยู่ในแผน |
|------|--------------|
| ระบบสารสนเทศ | Payment Core, Ledger, Tokenization Vault, Acquirer/3DS Adapter, Webhook, DB, HSM/KMS, network edge |
| กระบวนการธุรกิจ | authorization, capture/void/refund, settlement/payout, reconciliation, merchant support, AML/sanctions |
| บุคลากรและสถานที่ | สำนักงานหลัก, ศูนย์ปฏิบัติการ (NOC/SOC), work-from-anywhere |
| ผู้ให้บริการภายนอก | cloud/IDC, sponsor bank, card switch, 3DS provider, screening vendor (ดู outsourcing risk) |

> **นอกขอบเขต:** ระบบภายในที่ไม่กระทบธุรกรรม (เช่น HR/marketing) จัดอยู่ในระดับ Tier 3 และไม่ครอบคลุม RTO ที่เข้มงวด

---

## 2. บทบาทและหน้าที่ (Roles & Responsibilities)

| บทบาท | ผู้รับผิดชอบ (ตัวอย่างตำแหน่ง) | หน้าที่ในภาวะวิกฤต |
|-------|-------------------------------|--------------------|
| **BCP Sponsor** | กรรมการผู้จัดการ / CEO | อนุมัติแผน, จัดสรรทรัพยากร, ตัดสินใจ invoke DR |
| **Crisis Manager (Incident Commander)** | Head of SRE / CTO | สั่งการภาพรวม, ประกาศ activation, ประสานทุกทีม |
| **DR Coordinator** | SRE Lead | รัน runbook failover/failback, ตรวจ RTO/RPO |
| **Technical Recovery Team** | SRE + Backend (Go) + DBA | กู้คืน service, DB, network, HSM/KMS |
| **Security Lead** | Head of Security / DevSecOps | ประเมิน cyber incident, containment, forensic, PCI Req 12.10 |
| **Compliance & Regulatory Liaison** | Compliance Officer | รายงาน **ธปท.**, **ปปง./AMLO** (ถ้าเกี่ยว AML), **PDPC** (ถ้าข้อมูลรั่ว), แจ้ง scheme/sponsor |
| **Comms / Merchant Support Lead** | Product / Support Lead | สื่อสารกับ merchant, สถานะหน้า status page |
| **Business Recovery Team** | Finance / Ops | ตรวจ settlement, reconciliation, ยอดค้าง |

> **การสืบทอดตำแหน่ง (Succession):** ทุกบทบาทหลักต้องมีผู้สำรอง (deputy) อย่างน้อย 1 คน พร้อม on-call rotation
> และรายชื่อ/ช่องทางติดต่อฉุกเฉิน (call tree) ปรับปรุงทุกไตรมาส เก็บทั้งแบบ online และ offline (พิมพ์)

---

## 3. การวิเคราะห์ผลกระทบทางธุรกิจ (Business Impact Analysis — BIA)

จัดลำดับความสำคัญของกระบวนการตาม **Maximum Tolerable Downtime (MTD)** และผลกระทบ

| ระดับ | กระบวนการ / บริการ | MTD | RTO เป้าหมาย | RPO เป้าหมาย | ผลกระทบหากล่ม |
|-------|--------------------|-----|--------------|--------------|----------------|
| **Tier 0 — Critical** | Authorization / Capture / Void (payment core online) | 1 ชม. | **≤ 30 นาที** | **≤ 5 นาที** | สูญเสียรายได้ทันที, SLA breach กับ merchant/scheme |
| **Tier 0 — Critical** | Ledger (source of truth) + Tokenization Vault + HSM/KMS | 1 ชม. | **≤ 30 นาที** | **0 (RPO≈0)** | สูญเสียความถูกต้องของเงิน = ห้ามเกิด (synchronous commit) |
| **Tier 1 — High** | Refund, Webhook/Notify, 3DS flow | 4 ชม. | ≤ 2 ชม. | ≤ 15 นาที | merchant ไม่ได้ callback, ผู้ถือบัตรไม่ได้เงินคืน |
| **Tier 2 — Medium** | Settlement / Payout, Reconciliation worker | 24 ชม. | ≤ 8 ชม. | ≤ 1 ชม. | จ่ายเงิน merchant ช้า (batch เลื่อนได้ในวันเดียวกัน) |
| **Tier 2 — Medium** | Merchant/Admin API, Dashboard, Reporting | 24 ชม. | ≤ 8 ชม. | ≤ 1 ชม. | กระทบการใช้งาน ไม่กระทบเงิน |
| **Tier 3 — Low** | ระบบภายใน (BI, internal tools) | 72 ชม. | ≤ 48 ชม. | ≤ 24 ชม. | กระทบต่ำ |

**เหตุผลของค่า RPO/RTO:**
- **RPO ≤ 5 นาที** สำหรับ operational DB มาจาก streaming replication (PostgreSQL) ไปยัง DR site
- **RPO ≈ 0** สำหรับ ledger/vault ใช้ **synchronous commit** ไปยัง standby อย่างน้อย 1 node (ยอมเสีย latency
  เล็กน้อยเพื่อไม่ให้เงินหาย — สอดคล้องหลัก "fail closed" ใน `ARCHITECTURE.md` ข้อ 2)
- **RTO ≤ 30 นาที** สำหรับ payment core ผ่าน hot/warm standby ที่ pre-provisioned ไว้

> [!WARNING]
> **สมมติฐาน / TODO ที่ยังไม่ยุติ (ต้องยืนยันก่อนยื่น)**
> - **Cloud/IDC provider และ region:** ยังไม่สรุปผู้ให้บริการ (ตัวเลือกที่ต้องมี region ในไทยตามข้อกำหนด data
>   residency ของ ธปท./PDPA เช่น AWS ap-southeast-7 (Bangkok), Google Cloud, หรือ IDC ในประเทศ) — ค่า RTO/RPO
>   จริงต้องยืนยันกับ SLA ของ provider ที่เลือก
> - **Sponsor bank / acquirer:** ยังไม่ลงนาม — SLA availability, ช่องทาง failover การเชื่อม (ISO 8583 / API),
>   และข้อกำหนด BCP ของ sponsor จะ override เกณฑ์ขั้นต่ำในเอกสารนี้เมื่อทราบ
> - **QSA vendor (PCI-DSS L1):** ยังไม่เลือก — ขอบเขต DR ที่อยู่ใน CDE (Cardholder Data Environment) และหลักฐาน
>   ที่ต้องแสดงใน RoC อาจปรับตามคำแนะนำ QSA
> - **3DS provider / HSM vendor:** ยังไม่เลือก — โหมด HA/DR ของ HSM (cluster/replication ของ key material)
>   ต้องยืนยันตามผลิตภัณฑ์จริง
> - **ทุนจดทะเบียนชำระแล้วจริง:** ต้องยืนยันคงไว้ ≥ 50 ล้านบาท (≥ 75% ตลอดการดำเนินงาน)

---

## 4. สถาปัตยกรรมความพร้อมใช้และ DR (Availability & DR Architecture)

### 4.1 หลักการ

- **Active–Active ภายใน region หลัก** — API stateless หลาย availability zone (AZ) หลัง load balancer + WAF
- **Warm/Hot standby ที่ DR site (คนละ AZ/region ในไทย)** — pre-provisioned, พร้อม promote
- **Data residency:** ข้อมูล production และ backup ทั้งหมดเก็บภายในประเทศไทยตามข้อกำหนด ธปท./PDPA
- **CDE segmentation คงอยู่ที่ DR site** — network segmentation, WAF, mTLS ต้องเหมือน production (PCI Req 1)

### 4.2 กลยุทธ์ตามชั้นข้อมูล

| ชั้น | กลไก | โหมด replicate | RPO |
|------|------|----------------|-----|
| API / Service (stateless) | multi-AZ, autoscale, distroless image | N/A (redeploy จาก image registry) | N/A |
| Operational DB (payments, webhook) | PostgreSQL primary + streaming replica | asynchronous | ≤ 5 นาที |
| **Ledger + Vault (critical)** | PostgreSQL + **synchronous_commit=remote_apply** ไป standby | synchronous | ≈ 0 |
| Object storage (logs, exports) | cross-AZ replicated bucket, versioned, immutable (WORM) | near-real-time | ≤ 15 นาที |
| HSM / KMS | HSM cluster / partition replication (dual site) | vendor-managed | ≈ 0 |
| Secrets | secrets manager replicated + break-glass offline copy (sealed) | N/A | N/A |

### 4.3 การเชื่อมต่อภายนอก

- **Acquirer / card switch:** ออกแบบให้มีเส้นทางสำรอง (redundant links) และ **fail closed** — เมื่อไม่แน่ใจสถานะ
  authorization ให้ถือว่ายังไม่สำเร็จ แล้ว reconcile ภายหลัง (ป้องกัน double charge)
- **3DS provider:** timeout + fallback rules; หาก ACS/DS ล่ม ให้ decline อย่างปลอดภัยตามนโยบายความเสี่ยง
- **Webhook:** at-least-once + retry + signature; คิว persist ที่ DB จึงไม่สูญหายเมื่อ failover

---

## 5. การสำรองข้อมูล (Backup Strategy)

| รายการ | ความถี่ | ประเภท | เก็บนาน (Retention) | ที่เก็บ | เข้ารหัส |
|--------|---------|--------|---------------------|---------|----------|
| PostgreSQL (ทุก DB) | ต่อเนื่อง (WAL) + full daily | PITR (Point-in-Time Recovery) | 35 วัน (PITR), full 90 วัน | in-country, cross-AZ | AES-256 at rest |
| Ledger / Vault | ต่อเนื่อง (WAL) + full daily | PITR + immutable snapshot | ≥ 1 ปี (audit) | in-country | AES-256 + KMS |
| Audit log (`audit_log`) | ต่อเนื่อง | append-only export → WORM bucket | **≥ 1 ปี online, ≥ 5 ปี archive** (ธปท./PCI Req 10.5) | immutable object store | AES-256 |
| Config / IaC / secrets metadata | ทุกการเปลี่ยนแปลง (git) | version control + snapshot | ตามนโยบาย git retention | repo + sealed break-glass | เข้ารหัส secret ทั้งหมด |
| HSM key material | ตาม vendor | secure key backup (dual control, split knowledge) | ตลอดอายุคีย์ | ตู้นิรภัย/HSM คู่ | ตาม PCI Req 3 |

**นโยบายสำคัญ**
- **กฎ 3-2-1:** อย่างน้อย 3 สำเนา, 2 media/AZ, 1 สำเนา offline/immutable — แต่ **ทั้งหมดต้องอยู่ในไทย**
- **ห้ามสำรอง prohibited data:** ห้าม backup full PAN / CVV / PIN / track — สอดคล้อง `ARCHITECTURE.md` ข้อ 6
  (operational DB เก็บได้แค่ `card_brand` + `card_last4`); PAN ที่ vault เก็บเป็น encrypted token เท่านั้น
- **Immutability:** backup ของ ledger/audit เป็น WORM ป้องกัน ransomware และการแก้ไขย้อนหลัง
- **ทดสอบ restore:** ทำ **restore test** อย่างน้อยเดือนละครั้ง (partial) และไตรมาสละครั้ง (full) พร้อมบันทึกผล
  restore duration เทียบกับ RTO
- **Key management ของ backup:** คีย์เข้ารหัส backup แยกจากคีย์ production, rotate ตามรอบ, dual control (PCI Req 3)

---

## 6. ระดับความรุนแรงและการเรียกใช้แผน (Severity & Activation)

| ระดับ | นิยาม | ตัวอย่าง | ผู้ประกาศ activate | เป้าเวลาแจ้งเตือน |
|-------|-------|----------|--------------------|-------------------|
| **SEV-1** | Tier 0 ล่มหรือข้อมูลเสี่ยงสูญ/รั่ว | payment core down, DB primary loss, สงสัย breach | Crisis Manager (แจ้ง CEO) | < 15 นาที |
| **SEV-2** | บริการหลักลดคุณภาพรุนแรง (degraded) | latency พุ่ง, error rate สูง, AZ เดียวล่ม | SRE Lead / on-call | < 30 นาที |
| **SEV-3** | เหตุจำกัด ไม่กระทบธุรกรรมโดยตรง | dashboard ช้า, non-critical worker ล่ม | on-call | < 1 ชม. |

**ทริกเกอร์อัตโนมัติ (ตัวอย่าง threshold ที่ตั้งใน alerting):**
- Authorization success rate < 98% ต่อเนื่อง 5 นาที → paging on-call (SEV-2)
- p99 auth latency > 800 ms ต่อเนื่อง 10 นาที → SEV-2
- DB replication lag > 5 นาที → SEV-2, > 15 นาที → SEV-1
- Primary DB heartbeat หาย > 60 วินาที → พิจารณา failover อัตโนมัติ/กึ่งอัตโนมัติ (SEV-1)
- Ledger write failure ใด ๆ → SEV-1 ทันที (fail closed)

---

## 7. ขั้นตอนการกู้คืน (Recovery Procedures / Runbooks — สรุป)

### 7.1 Failover ฐานข้อมูล (DB primary loss — SEV-1)

1. Alert ยิง → on-call ยืนยันภายใน 5 นาที; Crisis Manager ประกาศ SEV-1
2. ตรวจ replica lag ล่าสุด (ยืนยัน RPO ที่จะเกิด); ถ้า ledger/vault → ยืนยัน synchronous standby caught up
3. หยุด traffic เขียน (fence primary เดิม ป้องกัน split-brain)
4. **Promote standby** เป็น primary; อัปเดต service discovery / connection string
5. Redirect API → primary ใหม่; ตรวจ health + smoke test (authorize สินค้าจริงจำนวนน้อยใน canary)
6. เปิด traffic เต็ม; ประกาศ recovery; บันทึกเวลา (วัด RTO/RPO จริง)
7. ตั้ง standby ใหม่ทดแทน; เริ่ม reconciliation ตรวจธุรกรรมช่วงรอยต่อ

### 7.2 Region/Site DR (สูญเสีย site หลัก — SEV-1)

1. Crisis Manager ประกาศ DR invoke (มติร่วม CEO/CTO)
2. DR Coordinator รัน runbook promote DR site (DB, vault, HSM partition, service)
3. เปลี่ยน DNS/traffic ไป DR (TTL ต่ำ pre-set); ตรวจ segmentation/WAF/mTLS ที่ DR เทียบเท่า production
4. ตรวจ external links (acquirer/3DS) ที่ DR ทำงาน; ถ้าจำเป็นสลับ endpoint สำรอง
5. Smoke test end-to-end; เปิด traffic; comms แจ้ง merchant + status page
6. Compliance Liaison เตรียมรายงาน ธปท. (และ PDPC/ปปง. ถ้าเข้าเงื่อนไข)

### 7.3 Cyber incident / Ransomware (SEV-1)

1. **Containment** ก่อน — isolate ระบบที่กระทบ, revoke credential, เปลี่ยน key ที่เสี่ยง
2. รักษาหลักฐาน (forensic image, log) — ห้ามลบ; ประสาน Security Lead/QSA
3. กู้จาก **immutable/clean backup** ที่ยืนยันไม่ปนเปื้อน (WORM)
4. ประเมินการรั่วของข้อมูลส่วนบุคคล → PDPA breach notification (ดูข้อ 9)
5. Post-incident: root cause, remediation, ทบทวน PCI Req 12.10

### 7.4 Failback

- กลับสู่ site หลักเฉพาะเมื่อเสถียรและ replicate ครบ; ทำนอกช่วง peak; วางแผน controlled cutover
  พร้อม reconciliation ปิดรอยต่อทุกครั้ง

---

## 8. การซ้อมแผน DR (DR Drills / Testing)

| ประเภทการทดสอบ | ความถี่ | สิ่งที่ทดสอบ | เกณฑ์ผ่าน |
|-----------------|---------|--------------|-----------|
| **Tabletop exercise** | ทุกไตรมาส | ทีมรับมือ SEV-1 ตามสถานการณ์สมมติ, call tree, การตัดสินใจ | ทุกบทบาททราบหน้าที่, ช่องว่างถูกบันทึก |
| **Backup restore test** | เดือนละครั้ง (partial), ไตรมาสละครั้ง (full) | restore DB/ledger จาก backup + ตรวจ integrity | restore สำเร็จภายใน RTO, ข้อมูลตรง |
| **DB failover drill** | ทุกไตรมาส | promote standby จริงใน staging/แบบควบคุมใน prod | RTO ≤ 30 นาที, RPO ตามเป้า, ไม่มี data loss |
| **Full DR failover (site)** | **อย่างน้อยปีละ 1 ครั้ง** (ก่อน go-live 1 ครั้งใน Phase 4) | ยก workload ไป DR site เต็มรูปแบบ | ผ่าน RTO/RPO Tier 0, smoke test ผ่าน |
| **Chaos/component failure** | ทุกครึ่งปี | kill AZ/instance/dependency แบบสุ่ม | ระบบ self-heal, ไม่กระทบธุรกรรม |
| **Cyber incident simulation** | ปีละครั้ง | จำลอง ransomware/breach + PDPA/ธปท. notification | containment + report ภายในเวลาที่กำหนด |

**การบันทึกผลการซ้อม (บังคับสำหรับยื่น ธปท. และ PCI Req 12.10):**
- แต่ละ drill ต้องมี **After-Action Report (AAR)**: วันที่, ผู้เข้าร่วม, สถานการณ์, RTO/RPO ที่วัดได้จริง เทียบเป้า,
  ปัญหาที่พบ, action items พร้อมผู้รับผิดชอบและกำหนดปิด
- ทบทวนและปรับ BCP/DR อย่างน้อย **ปีละครั้ง** หรือเมื่อมีการเปลี่ยนแปลงสำคัญ (สถาปัตยกรรม, sponsor, provider)

> **หมายเหตุ ROADMAP:** DR drill รอบแรกกำหนดใน **Phase 4 — Certification & Go-live** ก่อน production cutover
> (ดู `ROADMAP.md`) และเป็นเงื่อนไขหนึ่งของ readiness ก่อนยื่น/เปิดบริการ

---

## 9. การสื่อสารและการรายงานหน่วยงานกำกับ (Communication & Regulatory Reporting)

| ผู้รับ | เมื่อใด | ผู้รับผิดชอบ |
|--------|---------|--------------|
| **ธปท. (BOT)** | เหตุ IT/cyber ที่กระทบบริการอย่างมีนัยสำคัญ ตามหลักเกณฑ์ ธปท. (แจ้งเบื้องต้นโดยเร็ว + รายงานเต็มภายในกรอบที่กำหนด) | Compliance Liaison |
| **PDPC** | เหตุข้อมูลส่วนบุคคลรั่วไหลที่มีความเสี่ยงต่อเจ้าของข้อมูล — **ภายใน 72 ชั่วโมง** ตาม PDPA | DPO / Compliance |
| **ปปง. (AMLO)** | หากเหตุเกี่ยวข้องกับธุรกรรมต้องสงสัย/สินทรัพย์ที่เกี่ยวข้อง | Compliance Liaison |
| **Card scheme / Sponsor bank** | ตามสัญญาและกฎ scheme (เช่น incident/breach notification) | Security Lead |
| **Merchant / ลูกค้า** | เมื่อบริการกระทบ — ผ่าน status page + อีเมล + support | Comms Lead |

- **ช่องทางสำรอง:** มี status page แยก infra, กลุ่มสื่อสารฉุกเฉิน (out-of-band), และ call tree แบบ offline
- **เทมเพลตสาร:** เตรียมไว้ล่วงหน้าสำหรับ SEV-1/SEV-2 เพื่อสื่อสารได้ทันเวลา

---

## 10. เอกสารและการทบทวนที่เกี่ยวข้อง

- ปรับปรุงเอกสารนี้ **อย่างน้อยปีละครั้ง** และหลังทุก SEV-1 / DR drill ใหญ่ (version control ใน repo)
- เอกสารที่เชื่อมโยง: `ARCHITECTURE.md` (ข้อ 7 Security/PCI, ข้อ 8 NFR), `ROADMAP.md` (Phase 4),
  `COMPLIANCE-TH.md`, และเอกสารชุด compliance อื่น (AML/KYC, sanctions, DPA)

---
---

# Business Continuity + Disaster Recovery plan: RTO/RPO, failover, backup, DR drills (English)

> Supporting document for the **Acquiring Service** license application under the Payment Systems Act
> B.E. 2560 (paid-up capital THB 50M) and **PCI-DSS v4.0 Level 1**.
> Document set: `docs/compliance/15-bcp-dr.md` · version 0.1
> **Note:** This is an internal policy/procedure of **[บริษัท / Company]**, not legal advice. It should be
> reviewed by legal counsel, IT-resilience experts, and the QSA before submission. See `ARCHITECTURE.md`
> (§8 Non-functional / RPO-RTO) and `ROADMAP.md` (Phase 4 DR drill).

---

## 1. Purpose & Scope

This document defines the **Business Continuity Plan (BCP)** and **Disaster Recovery (DR)** plan for
**[บริษัท / Company]** as a Full Acquiring payment service provider, in order to:

- Keep authorization / capture / refund / settlement services **running** and restore them within defined
  RTO/RPO after any disruption — aligned with the payment core availability target of ≥ 99.95%.
- Comply with **Bank of Thailand (BOT)** requirements on IT risk, cyber resilience, and business continuity
  for entities regulated under the Payment Systems Act.
- Protect availability and integrity per **PCI-DSS v4.0** (notably Req 12.10 Incident Response and Req 10
  Logging) and the **Personal Data Protection Act B.E. 2562 (PDPA)** where personal data is affected.
- Honor obligations to card schemes (Visa/Mastercard), the sponsor bank, and merchants per SLA.

**In scope:**

| Domain | Covered |
|--------|---------|
| Systems | Payment Core, Ledger, Tokenization Vault, Acquirer/3DS Adapter, Webhook, DB, HSM/KMS, network edge |
| Business processes | authorization, capture/void/refund, settlement/payout, reconciliation, merchant support, AML/sanctions |
| People & sites | HQ, operations center (NOC/SOC), work-from-anywhere |
| Third parties | cloud/IDC, sponsor bank, card switch, 3DS provider, screening vendor |

> **Out of scope:** Internal systems not affecting transactions (e.g., HR/marketing) are Tier 3, with relaxed RTO.

---

## 2. Roles & Responsibilities

| Role | Owner (example title) | Duty during crisis |
|------|-----------------------|--------------------|
| **BCP Sponsor** | Managing Director / CEO | Approve plan, allocate resources, authorize DR invocation |
| **Crisis Manager (Incident Commander)** | Head of SRE / CTO | Overall command, declare activation, coordinate teams |
| **DR Coordinator** | SRE Lead | Execute failover/failback runbooks, verify RTO/RPO |
| **Technical Recovery Team** | SRE + Backend (Go) + DBA | Recover services, DB, network, HSM/KMS |
| **Security Lead** | Head of Security / DevSecOps | Assess cyber incidents, containment, forensics, PCI Req 12.10 |
| **Compliance & Regulatory Liaison** | Compliance Officer | Report to **BOT**, **AMLO** (if AML-related), **PDPC** (if data breach); notify scheme/sponsor |
| **Comms / Merchant Support Lead** | Product / Support Lead | Communicate with merchants, maintain status page |
| **Business Recovery Team** | Finance / Ops | Verify settlement, reconciliation, outstanding balances |

> **Succession:** Every key role has at least one deputy plus on-call rotation. An emergency call tree is
> refreshed quarterly and stored both online and offline (printed).

---

## 3. Business Impact Analysis (BIA)

Processes are prioritized by **Maximum Tolerable Downtime (MTD)** and impact.

| Tier | Process / Service | MTD | Target RTO | Target RPO | Impact if down |
|------|-------------------|-----|------------|------------|----------------|
| **Tier 0 — Critical** | Authorization / Capture / Void (payment core online) | 1 h | **≤ 30 min** | **≤ 5 min** | Immediate revenue loss, SLA breach with merchant/scheme |
| **Tier 0 — Critical** | Ledger (source of truth) + Tokenization Vault + HSM/KMS | 1 h | **≤ 30 min** | **0 (RPO≈0)** | Loss of financial integrity = must not happen (sync commit) |
| **Tier 1 — High** | Refund, Webhook/Notify, 3DS flow | 4 h | ≤ 2 h | ≤ 15 min | Merchants miss callbacks, cardholders miss refunds |
| **Tier 2 — Medium** | Settlement / Payout, Reconciliation worker | 24 h | ≤ 8 h | ≤ 1 h | Delayed merchant payout (same-day batch shift tolerable) |
| **Tier 2 — Medium** | Merchant/Admin API, Dashboard, Reporting | 24 h | ≤ 8 h | ≤ 1 h | Usability impact, no money impact |
| **Tier 3 — Low** | Internal systems (BI, internal tools) | 72 h | ≤ 48 h | ≤ 24 h | Low impact |

**Rationale for RPO/RTO values:**
- **RPO ≤ 5 min** for the operational DB comes from PostgreSQL streaming replication to the DR site.
- **RPO ≈ 0** for ledger/vault uses **synchronous commit** to at least one standby node (accepting slightly
  higher latency so money is never lost — aligned with the "fail closed" principle in `ARCHITECTURE.md` §2).
- **RTO ≤ 30 min** for the payment core is achieved via a pre-provisioned hot/warm standby.

> [!WARNING]
> **Assumptions / open TODOs (must be confirmed before submission)**
> - **Cloud/IDC provider and region:** Not finalized. Must have an in-Thailand region per BOT/PDPA data-residency
>   requirements (e.g., AWS ap-southeast-7 Bangkok, Google Cloud, or a local IDC). Actual RTO/RPO must be
>   confirmed against the chosen provider's SLA.
> - **Sponsor bank / acquirer:** Not yet signed. The sponsor's availability SLA, failover connectivity path
>   (ISO 8583 / API), and BCP requirements will override the minimums here once known.
> - **QSA vendor (PCI-DSS L1):** Not yet selected. The DR scope inside the CDE (Cardholder Data Environment)
>   and evidence required for the RoC may change per QSA guidance.
> - **3DS provider / HSM vendor:** Not selected. HSM HA/DR mode (key-material clustering/replication) must be
>   confirmed against the actual product.
> - **Actual paid-up capital:** Must be confirmed maintained ≥ THB 50M (≥ 75% throughout operations).

---

## 4. Availability & DR Architecture

### 4.1 Principles

- **Active–Active within the primary region** — stateless API across multiple AZs behind LB + WAF.
- **Warm/Hot standby at a DR site** (separate AZ/region in Thailand), pre-provisioned and promotable.
- **Data residency:** All production data and backups reside within Thailand per BOT/PDPA.
- **CDE segmentation preserved at DR** — network segmentation, WAF, and mTLS must match production (PCI Req 1).

### 4.2 Strategy by data tier

| Tier | Mechanism | Replication mode | RPO |
|------|-----------|------------------|-----|
| API / Service (stateless) | multi-AZ, autoscale, distroless image | N/A (redeploy from image registry) | N/A |
| Operational DB (payments, webhook) | PostgreSQL primary + streaming replica | asynchronous | ≤ 5 min |
| **Ledger + Vault (critical)** | PostgreSQL + **synchronous_commit=remote_apply** to standby | synchronous | ≈ 0 |
| Object storage (logs, exports) | cross-AZ replicated bucket, versioned, immutable (WORM) | near-real-time | ≤ 15 min |
| HSM / KMS | HSM cluster / partition replication (dual site) | vendor-managed | ≈ 0 |
| Secrets | secrets manager replicated + sealed offline break-glass copy | N/A | N/A |

### 4.3 External connectivity

- **Acquirer / card switch:** Redundant links and **fail closed** — if authorization status is uncertain,
  treat as not successful and reconcile later (prevents double charging).
- **3DS provider:** Timeouts + fallback rules; if ACS/DS is down, decline safely per risk policy.
- **Webhook:** At-least-once + retry + signature; the queue is persisted in the DB, so nothing is lost on failover.

---

## 5. Backup Strategy

| Item | Frequency | Type | Retention | Location | Encryption |
|------|-----------|------|-----------|----------|------------|
| PostgreSQL (all DBs) | continuous (WAL) + daily full | PITR | 35 days (PITR), full 90 days | in-country, cross-AZ | AES-256 at rest |
| Ledger / Vault | continuous (WAL) + daily full | PITR + immutable snapshot | ≥ 1 year (audit) | in-country | AES-256 + KMS |
| Audit log (`audit_log`) | continuous | append-only export → WORM bucket | **≥ 1 year online, ≥ 5 years archive** (BOT/PCI Req 10.5) | immutable object store | AES-256 |
| Config / IaC / secrets metadata | every change (git) | version control + snapshot | per git retention | repo + sealed break-glass | all secrets encrypted |
| HSM key material | per vendor | secure key backup (dual control, split knowledge) | key lifetime | vault / paired HSM | per PCI Req 3 |

**Key policies**
- **3-2-1 rule:** at least 3 copies, 2 media/AZ, 1 offline/immutable — but **all must remain in Thailand**.
- **Never back up prohibited data:** No full PAN / CVV / PIN / track data — consistent with `ARCHITECTURE.md`
  §6 (operational DB holds only `card_brand` + `card_last4`); the vault stores PAN only as an encrypted token.
- **Immutability:** Ledger/audit backups are WORM to defend against ransomware and retroactive tampering.
- **Restore testing:** Perform a **restore test** at least monthly (partial) and quarterly (full), recording
  restore duration against RTO.
- **Backup key management:** Backup encryption keys are separate from production keys, rotated on schedule,
  under dual control (PCI Req 3).

---

## 6. Severity & Activation

| Level | Definition | Example | Declared by | Notification target |
|-------|-----------|---------|-------------|---------------------|
| **SEV-1** | Tier 0 down, or data at risk of loss/leak | payment core down, DB primary loss, suspected breach | Crisis Manager (informs CEO) | < 15 min |
| **SEV-2** | Severely degraded core service | latency spike, high error rate, single AZ down | SRE Lead / on-call | < 30 min |
| **SEV-3** | Limited event, no direct transaction impact | slow dashboard, non-critical worker down | on-call | < 1 h |

**Automated triggers (example alerting thresholds):**
- Authorization success rate < 98% sustained 5 min → page on-call (SEV-2)
- p99 auth latency > 800 ms sustained 10 min → SEV-2
- DB replication lag > 5 min → SEV-2; > 15 min → SEV-1
- Primary DB heartbeat lost > 60 s → consider (semi-)automated failover (SEV-1)
- Any ledger write failure → immediate SEV-1 (fail closed)

---

## 7. Recovery Procedures (Runbooks — summary)

### 7.1 Database failover (DB primary loss — SEV-1)

1. Alert fires → on-call confirms within 5 min; Crisis Manager declares SEV-1.
2. Check latest replica lag (confirm the resulting RPO); for ledger/vault, confirm synchronous standby is caught up.
3. Stop write traffic (fence the old primary to prevent split-brain).
4. **Promote the standby** to primary; update service discovery / connection strings.
5. Redirect API to the new primary; run health + smoke tests (small canary of real authorizations).
6. Reopen full traffic; declare recovery; log timings (measure actual RTO/RPO).
7. Provision a replacement standby; run reconciliation over the cutover window.

### 7.2 Region/Site DR (loss of primary site — SEV-1)

1. Crisis Manager declares DR invocation (joint CEO/CTO decision).
2. DR Coordinator runs the DR-site promotion runbook (DB, vault, HSM partition, services).
3. Switch DNS/traffic to DR (pre-set low TTL); verify segmentation/WAF/mTLS at DR match production.
4. Verify external links (acquirer/3DS) at DR; switch to backup endpoints if needed.
5. End-to-end smoke test; reopen traffic; comms to merchants + status page.
6. Compliance Liaison prepares BOT report (and PDPC/AMLO if criteria are met).

### 7.3 Cyber incident / Ransomware (SEV-1)

1. **Containment first** — isolate affected systems, revoke credentials, rotate at-risk keys.
2. Preserve evidence (forensic images, logs) — do not delete; engage Security Lead/QSA.
3. Restore from **immutable/clean backups** verified uncontaminated (WORM).
4. Assess personal-data exposure → PDPA breach notification (see §9).
5. Post-incident: root cause, remediation, review PCI Req 12.10.

### 7.4 Failback

- Return to the primary site only when stable and fully replicated; perform off-peak as a controlled cutover,
  always closing the window with reconciliation.

---

## 8. DR Drills / Testing

| Test type | Frequency | What is tested | Pass criteria |
|-----------|-----------|----------------|---------------|
| **Tabletop exercise** | quarterly | Team handling of a simulated SEV-1, call tree, decision-making | All roles know duties; gaps logged |
| **Backup restore test** | monthly (partial), quarterly (full) | Restore DB/ledger from backup + integrity check | Restore within RTO; data matches |
| **DB failover drill** | quarterly | Real standby promotion in staging / controlled in prod | RTO ≤ 30 min, RPO on target, no data loss |
| **Full DR failover (site)** | **at least annually** (plus one before go-live in Phase 4) | Full workload shift to DR site | Meets Tier 0 RTO/RPO; smoke tests pass |
| **Chaos/component failure** | semi-annually | Randomly kill AZ/instance/dependency | System self-heals; no transaction impact |
| **Cyber incident simulation** | annually | Simulated ransomware/breach + PDPA/BOT notification | Containment + reporting within target time |

**Drill recordkeeping (mandatory for BOT submission and PCI Req 12.10):**
- Each drill produces an **After-Action Report (AAR):** date, participants, scenario, measured RTO/RPO vs.
  target, issues found, action items with owners and due dates.
- Review and update the BCP/DR **at least annually** or upon significant change (architecture, sponsor, provider).

> **ROADMAP note:** The first DR drill is scheduled in **Phase 4 — Certification & Go-live** before production
> cutover (see `ROADMAP.md`) and is a readiness gate before submission/launch.

---

## 9. Communication & Regulatory Reporting

| Recipient | When | Owner |
|-----------|------|-------|
| **BOT** | Material IT/cyber events affecting service, per BOT rules (prompt initial notice + full report within the prescribed window) | Compliance Liaison |
| **PDPC** | Personal-data breach posing risk to data subjects — **within 72 hours** per PDPA | DPO / Compliance |
| **AMLO** | If the event involves suspicious transactions/related assets | Compliance Liaison |
| **Card scheme / Sponsor bank** | Per contract and scheme rules (e.g., incident/breach notification) | Security Lead |
| **Merchants / customers** | When service is impacted — via status page + email + support | Comms Lead |

- **Backup channels:** A separately-hosted status page, an out-of-band emergency comms group, and an offline call tree.
- **Message templates:** Prepared in advance for SEV-1/SEV-2 to communicate on time.

---

## 10. Related Documents & Review

- Update this document **at least annually** and after every SEV-1 / major DR drill (version-controlled in the repo).
- Linked documents: `ARCHITECTURE.md` (§7 Security/PCI, §8 NFR), `ROADMAP.md` (Phase 4), `COMPLIANCE-TH.md`,
  and the other compliance-set documents (AML/KYC, sanctions, DPA).
