# ขั้นตอนการคัดกรองรายชื่อต้องห้าม (Sanctions) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **Acquiring Service** ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560
> (ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท) และมาตรฐาน **PCI-DSS v4.0 Level 1**
> เอกสารในชุด: `docs/compliance/06-sanctions-screening.md` · เวอร์ชัน 0.1
> **หมายเหตุ:** เอกสารนี้เป็นนโยบายและขั้นตอนปฏิบัติภายในของ **[บริษัท / Company]** ไม่ใช่คำแนะนำ
> ทางกฎหมาย ควรให้ที่ปรึกษากฎหมาย/ผู้เชี่ยวชาญด้าน AML และสำนักงาน ปปง. (AMLO) ตรวจทานก่อนยื่นจริง

---

## 1. วัตถุประสงค์และขอบเขต

เอกสารนี้กำหนดขั้นตอนการคัดกรองรายชื่อบุคคล/นิติบุคคลที่ถูกกำหนด (sanctions / designated persons)
สำหรับ **[บริษัท / Company]** ในฐานะผู้ให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)
เพื่อ:

- ป้องกันมิให้บริษัทถูกใช้เป็นช่องทางสนับสนุนการก่อการร้าย การแพร่ขยายอาวุธที่มีอานุภาพทำลายล้างสูง
  (Proliferation Financing) และการฟอกเงิน ตาม **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน พ.ศ. 2542**
  และ **พ.ร.บ. ป้องกันและปราบปรามการสนับสนุนทางการเงินแก่การก่อการร้ายและการแพร่ขยายอาวุธที่มี
  อานุภาพทำลายล้างสูง พ.ศ. 2559 (CTPF Act)**
- ปฏิบัติตามหลักเกณฑ์ของ **ธนาคารแห่งประเทศไทย (ธปท.)** ด้าน AML/CFT และการบริหารความเสี่ยง
- ปฏิบัติตามพันธะของ card scheme (Visa/Mastercard) และ sponsor bank

**ขอบเขต (Scope) ครอบคลุม:**

| กลุ่มที่ต้องคัดกรอง | ตัวอย่าง |
|---------------------|----------|
| ผู้ค้า (Merchant) และผู้มีอำนาจควบคุม (UBO ≥ 25%, กรรมการ, ผู้มีอำนาจลงนาม) | ตอน onboarding และตลอดความสัมพันธ์ |
| ผู้ถือบัตร / ผู้ชำระเงิน (Cardholder / Payer) | ในธุรกรรมที่ทราบชื่อ (เช่น payout, refund ไปบัญชีปลายทาง) |
| ผู้รับโอนเงิน / บัญชีรับ settlement / payout | ก่อนจ่ายเงินออก |
| พนักงาน คู่ค้า และผู้ให้บริการภายนอก (outsourcing) | ก่อนทำสัญญา |

> **หลักการ:** ระบบหลัก (payment core) เห็นข้อมูลบัตรเป็น token + `card_last4` เท่านั้น (ดู
> `ARCHITECTURE.md` ข้อ 6) การคัดกรองผู้ถือบัตรจึงใช้ข้อมูลระบุตัวตนที่ได้จาก 3DS/KYC ไม่ใช่ PAN เต็ม

---

## 2. รายการต้องห้ามที่ใช้ (Sanctions Lists)

| # | รายการ | แหล่ง | ลักษณะบังคับ |
|---|--------|-------|--------------|
| 1 | **UN Consolidated Sanctions List** | คณะมนตรีความมั่นคงแห่งสหประชาชาติ (UNSC 1267/1988/1718/2231) | ผูกพันไทยผ่าน CTPF Act |
| 2 | **รายชื่อบุคคลที่ถูกกำหนด (Designated Persons List) ของไทย** | สำนักงาน ปปง. (AMLO) ประกาศตาม CTPF Act มาตรา 5/6/7 | บังคับตามกฎหมายไทย |
| 3 | **OFAC SDN List + Consolidated (Non-SDN)** | U.S. Department of the Treasury (OFAC) | ผูกพันผ่าน card scheme / USD correspondent / สัญญา sponsor bank |
| 4 | **UK OFSI / EU Consolidated List** | HM Treasury / EU | ใช้ตามความเสี่ยงและข้อกำหนดของคู่ค้า |
| 5 | **Internal watchlist / blacklist ของบริษัท** | รวบรวมภายใน (เช่น merchant ที่เคยฉ้อโกง, chargeback สูง) | นโยบายภายใน |
| 6 | **PEP list** (Politically Exposed Persons) | ผู้ให้บริการข้อมูลเชิงพาณิชย์ | Enhanced Due Diligence (EDD) ไม่ใช่การบล็อกอัตโนมัติ |

> **การจัดการเวอร์ชันรายการ:** ทุกครั้งที่รับ list เข้าระบบ ต้องบันทึก `list_source`, `list_version`,
> `published_date`, `ingested_at`, และ checksum เพื่อให้ตรวจสอบย้อนหลังได้ว่าคัดกรองด้วยรายการเวอร์ชันใด
> (audit trail ตาม PCI-DSS Req 10 และหลักเกณฑ์ ธปท.)

> [!WARNING]
> **สมมติฐาน / TODO ที่ยังไม่ยุติ (ต้องยืนยันก่อนยื่น)**
> - **ผู้ให้บริการข้อมูล sanctions/PEP (screening data vendor):** ยังไม่เลือก (ตัวเลือก เช่น
>   Dow Jones, LexisNexis, ComplyAdvantage, Refinitiv World-Check) — SLA การอัปเดต, fuzzy-match engine
>   และราคายังไม่สรุป
> - **Sponsor bank / acquirer:** ยังไม่ลงนาม — ข้อกำหนด sanctions เพิ่มเติมของ sponsor (เช่น OFAC 50% Rule,
>   list เพิ่มเติม) จะ override เกณฑ์ขั้นต่ำในเอกสารนี้เมื่อทราบ
> - **QSA vendor (PCI-DSS L1):** ยังไม่เลือก — ขอบเขต logging/audit ของกระบวนการนี้อาจปรับตาม RoC
> - **ทุนจดทะเบียนชำระแล้วจริง:** ต้องยืนยันคงไว้ ≥ 50 ล้านบาท (≥ 75% ตลอดการดำเนินงาน)

---

## 3. เวลาและความถี่ในการคัดกรอง (Screening Frequency)

| เหตุการณ์ (Trigger) | สิ่งที่คัดกรอง | เวลา |
|---------------------|----------------|------|
| **Onboarding merchant** | merchant + UBO + กรรมการ + ผู้ลงนาม | ก่อนเปิดใช้งาน (blocking) — ต้องผ่านก่อน activate |
| **Real-time ก่อนอนุมัติธุรกรรม** | payer/beneficiary ที่ทราบชื่อ | Inline ใน authorization flow, งบเวลา < 300 ms (ส่วนหนึ่งของ p99 < 800 ms) |
| **ก่อน payout / settlement ออก** | บัญชี/ผู้รับปลายทาง | Blocking ก่อนปล่อยเงิน |
| **Batch re-screening ทั้งฐานข้อมูล** | merchant + UBO ที่ active ทั้งหมด | **ทุกวัน (daily)** เทียบกับ list เวอร์ชันล่าสุด |
| **Delta re-screening เมื่อ list เปลี่ยน** | เฉพาะรายการที่เพิ่ม/แก้ในลิสต์ เทียบทั้งฐาน | **ภายใน 24 ชม.** หลัง list ใหม่เผยแพร่ (เป้าหมาย: ทันทีสำหรับ UN/AMLO) |
| **เมื่อข้อมูลลูกค้าเปลี่ยน** (ชื่อ, UBO, ที่อยู่) | รายการที่เปลี่ยน | ทันทีที่บันทึกการเปลี่ยนแปลง |
| **Periodic KYC refresh** | ตามระดับความเสี่ยงลูกค้า | ความเสี่ยงสูง: ทุก 12 เดือน · กลาง: 24 เดือน · ต่ำ: 36 เดือน |

**การอัปเดต list:**

- **UN & AMLO Designated Persons:** ดึง/ตรวจสอบ **ทุกวันทำการ** และทุกครั้งที่มีประกาศใหม่ (บริษัท
  ตั้ง subscription/monitor ประกาศ ปปง.) — ถือเป็นลำดับความสำคัญสูงสุด (ผูกพันตามกฎหมายไทย)
- **OFAC/OFSI/EU:** ดึงอัตโนมัติ **ทุกวัน** ผ่าน vendor feed
- ทุก feed ต้องผ่าน integrity check (checksum + record count) ก่อนใช้งานจริง หาก feed ล้มเหลว → alert
  ทีม Compliance ภายใน 1 ชม. และใช้ **fail-closed** (ดูข้อ 7)

---

## 4. วิธีการจับคู่ (Matching Methodology)

- **Fuzzy / phonetic matching** รองรับการสะกดต่างและการทับศัพท์ไทย↔อังกฤษ (เช่น Jaro-Winkler,
  Levenshtein, Soundex/Metaphone, การถอดอักษรไทย-โรมัน)
- **Match score threshold** เริ่มต้น:

| Score | การจัดประเภท | การดำเนินการ |
|-------|--------------|--------------|
| ≥ 95% | Strong / near-exact | บล็อกทันที → Confirmed pending review (L1) |
| 85–94% | Probable | Hold → คัดกรองโดยเจ้าหน้าที่ (L1) ภายใน SLA |
| 70–84% | Possible | Hold แบบ soft → review (L1) |
| < 70% | ต่ำกว่าเกณฑ์ | ไม่แจ้งเตือน แต่บันทึก log |
| Exact ใน list UN/AMLO | Legal hit | บล็อก + freeze + รายงาน ปปง. (ดูข้อ 6) |

- ปัจจัยประกอบการจับคู่: ชื่อ + วันเดือนปีเกิด + สัญชาติ + เลขประจำตัว/เลขนิติบุคคล + ที่อยู่
  เพื่อลด false positive
- **การปรับ threshold** ต้องผ่านการอนุมัติจาก MLRO และบันทึกเหตุผล (change log) — ห้ามปรับเพื่อ
  หลบเลี่ยงการแจ้งเตือน
- **OFAC 50% Rule / UN ownership:** นิติบุคคลที่ถูกถือหุ้นรวม ≥ 50% โดยบุคคลที่ถูกกำหนด ให้ถือเป็น
  รายการต้องห้ามด้วย

---

## 5. การจัดการเมื่อพบรายการตรงกัน (Match Handling)

```
เกิด alert (real-time / batch / delta)
        │
        ▼
  Auto-hold ธุรกรรม/บัญชี (fail-closed) — ยังไม่ปล่อยเงิน/ยังไม่ activate
        │
        ▼
  L1 Analyst คัดกรองภายใน SLA ──► False positive? ──► ปิด alert + บันทึกเหตุผล + ปล่อย hold
        │                                              (whitelist มี TTL, ทบทวนเป็นงวด)
        │ True / ไม่ชัดเจน
        ▼
  ยกระดับ L2 (Compliance/MLRO) ──► ยืนยัน hit?
        │                              │ ใช่ (โดยเฉพาะ UN/AMLO)
        │ ไม่ใช่                       ▼
        └─► ปลด hold           Freeze ทรัพย์สิน/ระงับความสัมพันธ์
                                + รายงาน ปปง. + แจ้ง ธปท./sponsor bank
```

**หลักการสำคัญ:**

1. **ห้ามแจ้งเบาะแส (No tipping-off):** ตาม พ.ร.บ. ฟอกเงิน ห้ามแจ้งลูกค้าว่ากำลังถูกคัดกรอง/รายงาน
   ปปง. — สื่อสารกับลูกค้าตามสคริปต์ที่ฝ่ายกฎหมายอนุมัติเท่านั้น
2. **Fail-closed:** ระหว่างรอผล ธุรกรรมค้างสถานะ hold/`requires_action` และไม่มีการปล่อยเงินออก
3. **หลักฐานครบ:** ทุก alert เก็บ snapshot ของ list version, match score, ผู้ตัดสิน, เหตุผล และ
   timestamp (audit trail, immutable, retention ≥ 10 ปีตามข้อกำหนด AML ไทย)
4. **True match กับ UN/AMLO list = ต้อง freeze ทันทีและรายงานโดยไม่ชักช้า** (ไม่มีดุลพินิจให้ผ่าน)

---

## 6. การยกระดับและการรายงาน (Escalation & Reporting)

| ระดับ | ผู้รับผิดชอบ | หน้าที่ | SLA |
|-------|-------------|---------|-----|
| **L1** | Sanctions/AML Analyst (ทีม Compliance Operations) | คัดกรอง alert เบื้องต้น, ปิด false positive, ยกระดับ | เริ่มพิจารณาภายใน **4 ชม.ทำการ**; real-time hit ภายใน **1 ชม.** |
| **L2** | Compliance Manager / MLRO (Money Laundering Reporting Officer) | ยืนยัน true match, สั่ง freeze, อนุมัติปลด hold | ภายใน **24 ชม.** หลังรับเรื่อง |
| **L3** | Chief Compliance Officer + ที่ปรึกษากฎหมาย + กรรมการผู้จัดการ | ตัดสินใจกรณีซับซ้อน, การรายงานทางการ, การสื่อสารกับ ธปท./ปปง. | ตามกรอบกฎหมาย |

**การรายงานทางการ:**

- **True match (UN/AMLO/CTPF):** MLRO **ระงับทรัพย์สิน (freeze) ทันที** และรายงานต่อสำนักงาน ปปง.
  ตามแบบและกำหนดเวลาที่กฎหมายกำหนด **โดยไม่ชักช้า** — พร้อมส่ง **STR (Suspicious Transaction
  Report)** เมื่อเข้าเงื่อนไข
- **แจ้ง ธปท.** ตามหลักเกณฑ์การรายงานเหตุการณ์สำคัญ (material incident) หากกระทบต่อการดำเนินงาน/
  ชื่อเสียง หรือเข้าข่ายต้องแจ้ง
- **แจ้ง sponsor bank / card scheme** ตามสัญญา (เมื่อลงนามแล้ว)
- **PDPA / PDPC:** การประมวลผลข้อมูลส่วนบุคคลเพื่อการคัดกรองอาศัยฐาน "หน้าที่ตามกฎหมาย" — เก็บ
  เท่าที่จำเป็น, จำกัดการเข้าถึงตาม RBAC, และไม่เปิดเผยผลการคัดกรอง/การรายงานให้เจ้าของข้อมูล
  (ข้อยกเว้นตามกฎหมาย AML เหนือกว่าสิทธิการเข้าถึงของเจ้าของข้อมูล)

---

## 7. การควบคุมทางเทคนิคและ Fail-Closed

- คัดกรองเป็น service แยก (`sanctions-screening`) เรียกผ่าน interface — สลับ vendor ได้ตาม Clean
  Architecture (สอดคล้อง `ARCHITECTURE.md` ข้อ 4)
- **Fail-closed:** หาก screening service/feed ไม่พร้อม → real-time ธุรกรรมที่ต้องคัดกรองจะถูก hold
  ไม่ auto-approve; batch/payout หยุดรอ พร้อม alert ทีม Compliance ภายใน 1 ชม.
- **Audit:** ทุกการตัดสินใจลง `audit_log` (append-only), ไม่มีข้อมูลบัตร (PCI Req 10); เข้าถึงผลด้วย
  least-privilege + MFA (PCI Req 7–8)
- **Data residency:** ข้อมูลลูกค้า/ผลการคัดกรองเก็บในไทยตามข้อกำหนด ธปท./PDPA
- **การเข้ารหัส:** ข้อมูลระบุตัวตนเข้ารหัส at-rest (KMS/HSM) และ in-transit (TLS 1.2+/mTLS)

---

## 8. การกำกับ ทดสอบ และทบทวน (Governance & Assurance)

| กิจกรรม | ความถี่ | ผู้รับผิดชอบ |
|---------|---------|--------------|
| ทบทวนนโยบายฉบับนี้ | อย่างน้อยปีละ 1 ครั้ง หรือเมื่อกฎหมายเปลี่ยน | MLRO / CCO |
| Model/threshold tuning & false-positive review | ทุกไตรมาส | Compliance Analytics |
| ทดสอบระบบด้วยรายชื่อจำลอง (positive/negative test cases) | ทุกไตรมาส + หลังเปลี่ยน vendor/logic | QA + Compliance |
| Independent audit (internal/external) | ปีละ 1 ครั้ง | Internal Audit / QSA (ส่วนที่เกี่ยว PCI) |
| อบรมพนักงาน AML/Sanctions | แรกเข้า + ทบทวนปีละ 1 ครั้ง | Compliance |

---

# Sanctions screening procedure: UN/OFAC/Thai lists, screening frequency, match handling, escalation (English)

> Supporting document for the **Acquiring Service** license application under the Payment Systems
> Act B.E. 2560 (registered paid-up capital THB 50M) and **PCI-DSS v4.0 Level 1**.
> Document set: `docs/compliance/06-sanctions-screening.md` · Version 0.1
> **Note:** This is an internal policy/procedure of **[บริษัท / Company]**, not legal advice. Have it
> reviewed by AML counsel and validated against AMLO (ปปง.) requirements before submission.

---

## 1. Purpose & Scope

This document defines the procedure by which **[บริษัท / Company]**, as a licensed Acquiring payment
service provider, screens persons and entities against sanctions / designated-persons lists in order
to:

- Prevent the company from being used to finance terrorism, proliferation of weapons of mass
  destruction (Proliferation Financing), or launder money, under the **Anti-Money Laundering Act
  B.E. 2542** and the **Counter-Terrorism and Proliferation of Weapon of Mass Destruction Financing
  Act B.E. 2559 (CTPF Act)**.
- Comply with **Bank of Thailand (ธปท.)** AML/CFT and risk-management requirements.
- Meet card-scheme (Visa/Mastercard) and sponsor-bank obligations.

**Scope covers:**

| Population screened | Examples |
|---------------------|----------|
| Merchants and controllers (UBO ≥ 25%, directors, authorized signatories) | At onboarding and throughout the relationship |
| Cardholders / payers | Where identity is known (e.g., payouts, refunds to a destination account) |
| Payout / settlement beneficiaries | Before releasing funds |
| Employees, partners, outsourced service providers | Before contracting |

> **Principle:** The payment core sees card data only as a token + `card_last4` (see `ARCHITECTURE.md`
> §6). Cardholder screening therefore uses identity attributes obtained from 3DS/KYC, never the full PAN.

---

## 2. Sanctions Lists Used

| # | List | Source | Binding basis |
|---|------|--------|---------------|
| 1 | **UN Consolidated Sanctions List** | UN Security Council (1267/1988/1718/2231) | Binding on Thailand via the CTPF Act |
| 2 | **Thai Designated Persons List** | AMLO (ปปง.), issued under CTPF Act §5/6/7 | Binding under Thai law |
| 3 | **OFAC SDN + Consolidated (Non-SDN)** | U.S. Treasury (OFAC) | Via card scheme / USD correspondent / sponsor-bank contract |
| 4 | **UK OFSI / EU Consolidated List** | HM Treasury / EU | Risk-based and per counterparty requirements |
| 5 | **Internal watchlist / blacklist** | Compiled internally (fraud, excessive chargebacks) | Internal policy |
| 6 | **PEP list** | Commercial data vendor | Triggers Enhanced Due Diligence (EDD), not an automatic block |

> **List versioning:** Every ingested list records `list_source`, `list_version`, `published_date`,
> `ingested_at`, and a checksum so it is provable which list version screened a given record
> (audit trail per PCI-DSS Req 10 and ธปท. requirements).

> [!WARNING]
> **Open assumptions / TODO (confirm before submission)**
> - **Sanctions/PEP data vendor:** not yet selected (candidates: Dow Jones, LexisNexis,
>   ComplyAdvantage, Refinitiv World-Check). Update SLA, fuzzy-match engine, and pricing pending.
> - **Sponsor bank / acquirer:** not yet signed — additional sponsor sanctions requirements (e.g.,
>   OFAC 50% Rule, extra lists) will override the minimums here once known.
> - **QSA vendor (PCI-DSS L1):** not yet selected — logging/audit scope for this process may adjust per RoC.
> - **Actual paid-up capital:** must confirm and maintain ≥ THB 50M (≥ 75% throughout operations).

---

## 3. Screening Frequency

| Trigger | What is screened | Timing |
|---------|------------------|--------|
| **Merchant onboarding** | Merchant + UBO + directors + signatories | Before activation (blocking) — must clear first |
| **Real-time pre-authorization** | Named payer/beneficiary | Inline in the authorization flow, budget < 300 ms (part of p99 < 800 ms) |
| **Before payout / settlement** | Destination account / beneficiary | Blocking before funds release |
| **Full batch re-screening** | All active merchants + UBOs | **Daily**, against the latest list version |
| **Delta re-screening on list change** | Added/changed list entries vs. the whole base | **Within 24 h** of list publication (target: immediate for UN/AMLO) |
| **On customer data change** (name, UBO, address) | The changed record | Immediately on save |
| **Periodic KYC refresh** | Per customer risk rating | High: 12 mo · Medium: 24 mo · Low: 36 mo |

**List updates:**

- **UN & AMLO Designated Persons:** pulled/verified **every business day** and on every new
  publication (the company subscribes to / monitors AMLO announcements) — highest priority (binding
  under Thai law).
- **OFAC/OFSI/EU:** auto-pulled **daily** via vendor feed.
- Every feed passes an integrity check (checksum + record count) before use. If a feed fails →
  alert Compliance within 1 h and **fail-closed** (see §7).

---

## 4. Matching Methodology

- **Fuzzy / phonetic matching** handles spelling variants and Thai↔English transliteration
  (Jaro-Winkler, Levenshtein, Soundex/Metaphone, Thai-to-Roman transliteration).
- Initial **match-score thresholds:**

| Score | Classification | Action |
|-------|----------------|--------|
| ≥ 95% | Strong / near-exact | Immediate block → Confirmed, pending review (L1) |
| 85–94% | Probable | Hold → analyst review (L1) within SLA |
| 70–84% | Possible | Soft hold → review (L1) |
| < 70% | Below threshold | No alert, but logged |
| Exact on UN/AMLO list | Legal hit | Block + freeze + report to AMLO (see §6) |

- Corroborating attributes: name + DOB + nationality + ID/registration number + address, to reduce
  false positives.
- **Threshold changes** require MLRO approval and a documented rationale (change log) — thresholds
  must never be tuned to suppress legitimate alerts.
- **OFAC 50% Rule / UN ownership:** entities owned ≥ 50% in aggregate by designated persons are
  treated as designated as well.

---

## 5. Match Handling

```
Alert raised (real-time / batch / delta)
        │
        ▼
  Auto-hold transaction/account (fail-closed) — no funds released / no activation
        │
        ▼
  L1 Analyst reviews within SLA ──► False positive? ──► Close alert + record rationale + release hold
        │                                                (whitelist has TTL, reviewed periodically)
        │ True / unclear
        ▼
  Escalate to L2 (Compliance/MLRO) ──► Confirmed hit?
        │                                    │ Yes (esp. UN/AMLO)
        │ No                                 ▼
        └─► Release hold             Freeze assets / terminate relationship
                                     + report to AMLO + notify ธปท./sponsor bank
```

**Key principles:**

1. **No tipping-off:** Under the AML Act, the customer must not be told they are being screened /
   reported to AMLO. Communicate only via legally approved scripts.
2. **Fail-closed:** While pending, transactions stay in hold/`requires_action`; no funds leave.
3. **Full evidence:** Every alert stores a snapshot of list version, match score, decision-maker,
   rationale, and timestamp (immutable audit trail, retention ≥ 10 years per Thai AML requirements).
4. **A true match against UN/AMLO = immediate freeze and prompt reporting** (no discretion to pass).

---

## 6. Escalation & Reporting

| Tier | Owner | Responsibility | SLA |
|------|-------|----------------|-----|
| **L1** | Sanctions/AML Analyst (Compliance Operations) | Triage alerts, close false positives, escalate | Begin review within **4 business hours**; real-time hits within **1 hour** |
| **L2** | Compliance Manager / MLRO | Confirm true match, order freeze, approve hold release | Within **24 hours** of receipt |
| **L3** | Chief Compliance Officer + legal counsel + Managing Director | Complex cases, official reporting, ธปท./AMLO liaison | Per statutory deadlines |

**Official reporting:**

- **True match (UN/AMLO/CTPF):** the MLRO **freezes assets immediately** and reports to AMLO in the
  prescribed form and timeframe **without delay**, plus files an **STR (Suspicious Transaction
  Report)** where applicable.
- **Notify ธปท.** under material-incident reporting rules where operations/reputation are affected
  or reporting is otherwise required.
- **Notify sponsor bank / card scheme** per contract (once signed).
- **PDPA / PDPC:** personal-data processing for screening relies on the "legal obligation" lawful
  basis — collect only what is necessary, restrict access via RBAC, and do not disclose screening/
  reporting outcomes to the data subject (AML legal exemptions override data-subject access rights).

---

## 7. Technical Controls & Fail-Closed

- Screening runs as a dedicated service (`sanctions-screening`) behind an interface — vendor-swappable
  per Clean Architecture (consistent with `ARCHITECTURE.md` §4).
- **Fail-closed:** if the screening service/feed is unavailable, real-time transactions requiring
  screening are held (never auto-approved); batch/payout pauses, with a Compliance alert within 1 h.
- **Audit:** every decision is written to `audit_log` (append-only), with no card data (PCI Req 10);
  results accessed under least-privilege + MFA (PCI Req 7–8).
- **Data residency:** customer/screening data stored in Thailand per ธปท./PDPA.
- **Encryption:** identity data encrypted at rest (KMS/HSM) and in transit (TLS 1.2+/mTLS).

---

## 8. Governance & Assurance

| Activity | Frequency | Owner |
|----------|-----------|-------|
| Review this policy | At least annually, or on legal change | MLRO / CCO |
| Model/threshold tuning & false-positive review | Quarterly | Compliance Analytics |
| System testing with synthetic names (positive/negative cases) | Quarterly + after vendor/logic change | QA + Compliance |
| Independent audit (internal/external) | Annually | Internal Audit / QSA (PCI-relevant parts) |
| AML/Sanctions staff training | On hire + annual refresher | Compliance |

---

## Related documents

- `COMPLIANCE-TH.md` — Thai law, license categories, application process
- `ARCHITECTURE.md` — system architecture, PCI scope, data model
- `ROADMAP.md` — phases, timeline, cost estimates
