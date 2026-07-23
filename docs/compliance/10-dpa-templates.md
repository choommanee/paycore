# ข้อตกลงประมวลผลข้อมูล (DPA) กับผู้ให้บริการภายนอก (ไทย)

> เอกสารชุด: **การยื่นขอใบอนุญาต Full Acquiring** ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 (กำกับโดย **ธนาคารแห่งประเทศไทย — ธปท.**) + **PCI-DSS Level 1**
> เอกสารที่เกี่ยวข้อง: `../COMPLIANCE-TH.md` (กฎหมาย/ใบอนุญาต), `../ARCHITECTURE.md` (สถาปัตยกรรม/PCI scope), `../ROADMAP.md` (เฟส/timeline)
> **ข้อสงวน:** เอกสารนี้เป็นแม่แบบเชิงเทคนิค-นโยบาย ไม่ใช่คำแนะนำทางกฎหมาย ต้องให้ที่ปรึกษากฎหมายและ **QSA** ตรวจก่อนลงนามจริง

เอกสารนี้กำหนดแม่แบบ **ข้อตกลงประมวลผลข้อมูลส่วนบุคคล (Data Processing Agreement — DPA)** และเงื่อนไข outsourcing ที่ **[บริษัท / Company]** ต้องจัดทำกับผู้ให้บริการภายนอก 5 ประเภทที่อยู่บนเส้นทางการชำระเงินด้วยบัตร ได้แก่ **Acquirer/Sponsor Bank, ผู้ให้บริการ 3-D Secure (3DS), Tokenization Vault, ITMX (local rails), และ HSM/KMS provider**

---

## 1. กรอบกฎหมายและมาตรฐานที่ผูกกับ DPA

| กฎหมาย/มาตรฐาน | บทบาทต่อ DPA |
|---|---|
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — กำกับโดย **PDPC** | มาตรา 40 กำหนดหน้าที่ **ผู้ประมวลผลข้อมูล (Data Processor)** ต้องทำ **ข้อตกลงเป็นหนังสือ** กับผู้ควบคุมข้อมูล; มาตรา 37 มาตรการรักษาความมั่นคงปลอดภัย; มาตรา 28–29 การส่งข้อมูลออกนอกประเทศ |
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** — กำกับโดย **ธปท.** | หลักเกณฑ์ **การใช้บริการจากผู้ให้บริการภายนอก (outsourcing)**: ผู้รับใบอนุญาตยังคงรับผิดชอบเต็ม, ธปท. ต้องเข้าตรวจสอบผู้ให้บริการช่วงได้ (right to audit) |
| **PCI-DSS v4.0** | Req 12.8 (จัดการ **TPSP — Third-Party Service Providers**), Req 12.9 (ผู้ให้บริการต้องยอมรับความรับผิดชอบเป็นหนังสือ), Req 3 (คุ้มครอง SAD/PAN), Req 4 (เข้ารหัสระหว่างส่ง) |
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (AMLO / ปปง.)** | การส่งต่อ/เก็บข้อมูล KYC/CDD และการรายงานธุรกรรมต้องผูกในสัญญากับคู่ค้าที่เกี่ยวข้อง |
| **ประกาศ ธปท. ด้าน IT Risk / Cyber Resilience / BCP** | ผูกข้อกำหนด SLA, การแจ้งเหตุ, BCP/DR, exit plan เข้าใน DPA/outsourcing agreement |

**บทบาทตาม PDPA (สำคัญมาก — กำหนดชนิดของสัญญา):**

| คู่สัญญา | บทบาทของคู่สัญญา | บทบาทของ [บริษัท / Company] | ชนิดสัญญาหลัก |
|---|---|---|---|
| Acquirer / Sponsor Bank | ผู้ควบคุมข้อมูลร่วม / อิสระ (independent/joint controller) | ผู้ควบคุมข้อมูล | Data Sharing + Outsourcing |
| ผู้ให้บริการ 3DS | ผู้ประมวลผล (มักเป็น processor ให้ scheme/issuer ด้วย) | ผู้ควบคุมข้อมูล | DPA (Art. 40) |
| Tokenization Vault | ผู้ประมวลผล | ผู้ควบคุมข้อมูล | DPA + PCI TPSP |
| ITMX | ผู้ควบคุมข้อมูล (ตัวกลางระบบชาติ) | ผู้ควบคุมข้อมูล | Interchange/Membership + Data Sharing |
| HSM/KMS provider | ผู้ประมวลผล (ไม่เห็น plaintext PAN หากออกแบบถูก) | ผู้ควบคุมข้อมูล | DPA + PCI TPSP |

---

## 2. โครงมาตรฐานของ DPA (ใช้ซ้ำได้ทุกคู่สัญญา)

ทุก DPA ต้องมีอย่างน้อย 14 ข้อต่อไปนี้ (สอดคล้อง PDPA ม.40 + PCI Req 12.8/12.9):

1. **นิยามและบทบาท** — ระบุ controller/processor, ขอบเขต scope, PCI CDE ที่เกี่ยวข้อง
2. **วัตถุประสงค์และขอบเขตการประมวลผล** — ประมวลผลได้เฉพาะตามคำสั่งเป็นลายลักษณ์อักษรของ [บริษัท / Company] เท่านั้น
3. **ประเภทข้อมูลและเจ้าของข้อมูล** (ดูตารางข้อ 3 ของแต่ละคู่สัญญา)
4. **มาตรการความมั่นคงปลอดภัย (technical & organizational)** — เข้ารหัส, access control, logging, key management
5. **การใช้ผู้ประมวลผลช่วง (sub-processor)** — ต้องขออนุมัติล่วงหน้า + ผูกเงื่อนไขเทียบเท่า
6. **การแจ้งเหตุละเมิดข้อมูล (breach notification)** — ตามตาราง timeline ข้อ 4
7. **สิทธิของเจ้าของข้อมูล** — ช่วยตอบคำขอ (access/rectify/erase/portability) ภายใน SLA
8. **การส่งข้อมูลข้ามพรมแดน** — ต้องเป็นไปตาม PDPA ม.28–29 และ data residency ของ ธปท.
9. **สิทธิในการตรวจสอบ (audit right)** — [บริษัท / Company], QSA และ **ธปท.** เข้าตรวจได้
10. **การรายงานการปฏิบัติตาม** — AoC/RoC (PCI), SOC 2 Type II, ISO 27001 รายปี
11. **BCP/DR และ SLA** — RTO/RPO, availability, ความต่อเนื่องบริการ
12. **การคืน/ทำลายข้อมูลเมื่อสิ้นสุด** — พร้อมหลักฐาน certificate of destruction
13. **ความรับผิดและการชดใช้ (liability & indemnity)** — รวมค่าปรับ regulator, ค่าปรับ scheme
14. **Exit / transition plan** — แผนถอนออกโดยไม่กระทบบริการและ compliance

---

## 3. ประเภทข้อมูลและฐานการประมวลผลรายคู่สัญญา

| คู่สัญญา | ข้อมูลที่ส่ง/ประมวลผล | ห้ามส่ง | ฐานการประมวลผล (PDPA) |
|---|---|---|---|
| **Acquirer / Sponsor Bank** | PAN (ผ่าน network token/ISO 8583), auth data, จำนวนเงิน, MCC, merchant ID | CVV2 หลัง authorization, PIN | สัญญา (ม.24(3)) + ประโยชน์โดยชอบด้วยกฎหมาย |
| **3DS provider** | PAN/DPAN, device data, browser fingerprint, cardholder name, billing | CVV2 (ไม่ผ่าน 3DS flow), track data | สัญญา + หน้าที่ตามกฎหมาย (SCA/scheme mandate) |
| **Tokenization Vault** | full PAN (ชั่วคราวเพื่อสร้าง token), expiry | CVV2, PIN, track เก็บถาวร | สัญญา |
| **ITMX** | บัญชี/พร้อมเพย์ proxy, จำนวนเงิน, PAN สำหรับ local card switch | CVV2, PIN block นอก HSM | สัญญา + หน้าที่ตามกฎหมาย (ระบบชำระเงินชาติ) |
| **HSM/KMS provider** | ciphertext, key material (ภายใน FIPS boundary) | **plaintext PAN/PII** (ไม่ควรออกจาก HSM) | สัญญา |

> **หลักการ:** ตาม `../ARCHITECTURE.md` §6 ระบบหลักเห็นแค่ `card_brand` + `card_last4`; **ห้ามเก็บ full PAN, CVV/CVV2, PIN, full track** ใน operational DB — สอดคล้อง PCI Req 3.2 (ห้ามเก็บ SAD หลัง authorization)

---

## 4. Timeline การแจ้งเหตุและ SLA (ผูกในทุก DPA)

| เหตุการณ์ | ผู้ให้บริการต้องแจ้ง [บริษัท / Company] ภายใน | [บริษัท / Company] แจ้ง regulator ภายใน |
|---|---|---|
| สงสัย/ยืนยันการละเมิดข้อมูล (data breach) | **≤ 24 ชม.** นับแต่ตรวจพบ | **PDPC ภายใน 72 ชม.** (PDPA ม.37(4)); **ธปท.** ตามประกาศ cyber incident (โดยทั่วไป ≤ 24 ชม. สำหรับเหตุร้ายแรง) |
| Account Data Compromise (ADC) ตาม PCI | **ทันที (immediate)** | แจ้ง scheme (Visa/Mastercard) + acquirer + QSA ตามขั้นตอน ADC |
| เหตุกระทบ availability (major outage) | **≤ 1 ชม.** | ธปท. ตามเกณฑ์ operational disruption |
| ผลตรวจ ASV scan ล้มเหลว / vuln วิกฤต | **≤ 48 ชม.** | ภายในแผน remediation PCI |
| การเปลี่ยน sub-processor / material change | **≥ 30 วันล่วงหน้า** | แจ้ง ธปท. หากเป็น material outsourcing |

**SLA ขั้นต่ำที่แนะนำ:** Availability ≥ 99.95% (สอดคล้อง `../ARCHITECTURE.md` §8), RPO ≤ 5 นาที, RTO ≤ 30 นาที

---

## 5. แม่แบบเฉพาะรายคู่สัญญา

### 5.1 Acquirer / Sponsor Bank

- **ชนิดสัญญา:** Merchant Acquiring / Sponsorship Agreement + Data Sharing Addendum
- **ประเด็นหลัก:** สิทธิเข้าถึง authorization & settlement network, chargeback/dispute allocation, การส่ง PAN ผ่าน network token, การจัดสรรความรับผิดจาก fraud/chargeback
- **PCI:** acquirer ต้องให้ **AoC** และรับ [บริษัท / Company] เข้าโปรแกรม TPSP monitoring
- **ธปท.:** ผูกข้อ right-to-audit ให้ ธปท. เข้าตรวจ sponsor ได้ผ่าน [บริษัท / Company]

> **[TODO / สมมติฐาน]** ยังไม่ได้เลือก **sponsor bank/acquirer** จริง (ดู `../ROADMAP.md` §5 critical path) — ต้องเติมชื่อคู่สัญญา, BIN sponsorship, และ scheme certification timeline หลังลงนาม LOI

### 5.2 ผู้ให้บริการ 3-D Secure (EMV 3DS)

- **ขอบเขต:** ACS/DS integration ตาม **EMV 3DS 2.x** เพื่อ SCA และโอน liability shift ไปยัง issuer
- **ข้อมูล:** device/browser data, DPAN — ต้องผูก data minimization และ retention ≤ ที่ scheme กำหนด
- **PCI:** 3DS server อยู่ใน scope; ผู้ให้บริการต้องส่ง AoC/SDP compliance
- **ผูก fallback:** เมื่อ 3DS ไม่พร้อม ให้ fail closed ตาม `../ARCHITECTURE.md` §2

### 5.3 Tokenization Vault

- **ขอบเขต:** สร้าง/คืน token, envelope encryption ด้วยคีย์จาก HSM (`../ARCHITECTURE.md` §7)
- **ข้อกำหนด:** vault ต้องอยู่ใน **network segment แยก** จากระบบหลัก; single-use payment token ฝั่ง client
- **PCI:** เป็น TPSP หลักใน CDE — ต้องมี **RoC/AoC Level 1**, quarterly ASV, annual pentest
- **การทำลายข้อมูล:** crypto-erase คีย์เมื่อสิ้นสุดสัญญา + certificate

### 5.4 ITMX (National Interbank Transaction Management and Exchange)

- **ชนิดสัญญา:** Membership/Interchange Agreement + Data Sharing (ทั้งคู่เป็น controller)
- **ขอบเขต:** local card switch / PromptPay proxy, การหักบัญชีและ settlement ในประเทศ
- **Data residency:** ข้อมูลอยู่ในไทยตามเกณฑ์ ธปท. (`../ARCHITECTURE.md` §8)
- **AML:** ผูกการส่งต่อข้อมูลธุรกรรมเพื่อการรายงาน ปปง. ตามที่กฎหมายกำหนด

> **[TODO / สมมติฐาน]** เงื่อนไข membership และ data flow ของ **ITMX** ต้องยืนยันตามข้อกำหนดสมาชิกจริง หลังได้รับการรับรองเชื่อมต่อ

### 5.5 HSM / KMS Provider

- **ขอบเขต:** จัดเก็บ/ป้องกันคีย์ระดับ **FIPS 140-2/3 Level 3**, dual control, split knowledge (PCI Req 3.6–3.7)
- **หลักการ:** **plaintext PAN/PII ต้องไม่ออกจาก HSM boundary**; provider เห็นแค่ ciphertext
- **บริการที่ผูก:** key rotation, key ceremony, HSM firmware update, HA/DR ของ key store
- **PCI:** provider ต้องมี AoC หรือ FIPS certificate + ผูก right-to-audit

> **[TODO / สมมติฐาน]** ยังไม่ได้เลือก **QSA vendor** และ **HSM/KMS provider** จริง (cloud KMS+CloudHSM หรือ on-prem) — ตัวเลข **ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท** อ้างตามเกณฑ์ Acquiring ใน `../COMPLIANCE-TH.md` §2 ให้ยืนยันจำนวนที่ชำระจริงกับฝ่ายการเงิน

---

## 6. Checklist ก่อนลงนาม (ทุกคู่สัญญา)

- [ ] ระบุบทบาท controller/processor ชัดเจนตาม PDPA ม.40
- [ ] แนบ TOMs (technical & organizational measures) เป็นภาคผนวก
- [ ] มี AoC/RoC (PCI) หรือ SOC 2 / ISO 27001 ล่าสุด
- [ ] ข้อ breach notification ≤ 24 ชม. + timeline regulator
- [ ] right-to-audit ครอบคลุม [บริษัท / Company], QSA และ **ธปท.**
- [ ] รายชื่อ sub-processor + สิทธิคัดค้าน
- [ ] cross-border transfer ผ่านเกณฑ์ PDPA ม.28–29 + data residency ธปท.
- [ ] คืน/ทำลายข้อมูล + certificate เมื่อสิ้นสุด
- [ ] exit/transition plan และ SLA/BCP ผูกครบ

---

# Data Processing Agreement templates for acquirer, 3DS, tokenization vault, ITMX, HSM providers (English)

> Document set: **Full Acquiring license** application under the **Payment Systems Act B.E. 2560** (supervised by the **Bank of Thailand — BOT**) + **PCI-DSS Level 1**
> Related docs: `../COMPLIANCE-TH.md` (law/licensing), `../ARCHITECTURE.md` (architecture/PCI scope), `../ROADMAP.md` (phases/timeline)
> **Disclaimer:** This is a policy/technical template, not legal advice. Legal counsel and a **QSA** must review before execution.

This document defines the **Data Processing Agreement (DPA)** and outsourcing templates that **[บริษัท / Company]** must execute with the five categories of external providers on the card payment path: **Acquirer/Sponsor Bank, 3-D Secure (3DS) provider, Tokenization Vault, ITMX (local rails), and HSM/KMS provider.**

---

## 1. Legal & Standards Framework Binding the DPA

| Law / Standard | Role in the DPA |
|---|---|
| **Personal Data Protection Act B.E. 2562 (PDPA)** — regulated by **PDPC** | Section 40 requires a **written agreement** between the **Data Processor** and controller; Section 37 security measures; Sections 28–29 cross-border transfers |
| **Payment Systems Act B.E. 2560** — regulated by **BOT** | **Outsourcing** rules: the licensee remains fully accountable; BOT retains a right to audit the service provider |
| **PCI-DSS v4.0** | Req 12.8 (manage **TPSPs**), Req 12.9 (provider written acknowledgement of responsibility), Req 3 (protect SAD/PAN), Req 4 (encrypt in transit) |
| **Anti-Money Laundering Act (AMLO)** | KYC/CDD data sharing and transaction reporting obligations bound into relevant contracts |
| **BOT notifications on IT Risk / Cyber Resilience / BCP** | Bind SLA, incident reporting, BCP/DR and exit plan into the DPA/outsourcing agreement |

**PDPA roles (critical — determines the contract type):**

| Counterparty | Their role | [บริษัท / Company] role | Primary contract |
|---|---|---|---|
| Acquirer / Sponsor Bank | Independent / joint controller | Controller | Data Sharing + Outsourcing |
| 3DS provider | Processor (often also processor for scheme/issuer) | Controller | DPA (s.40) |
| Tokenization Vault | Processor | Controller | DPA + PCI TPSP |
| ITMX | Controller (national scheme intermediary) | Controller | Interchange/Membership + Data Sharing |
| HSM/KMS provider | Processor (no plaintext PAN if correctly designed) | Controller | DPA + PCI TPSP |

---

## 2. Standard DPA Skeleton (reusable across all counterparties)

Every DPA must contain at minimum these 14 clauses (aligned with PDPA s.40 + PCI Req 12.8/12.9):

1. **Definitions & roles** — controller/processor, scope, applicable PCI CDE
2. **Purpose & scope of processing** — process only on [บริษัท / Company]'s documented instructions
3. **Data categories & data subjects** (see §3 per counterparty)
4. **Technical & organizational security measures** — encryption, access control, logging, key management
5. **Sub-processors** — prior approval + flow-down of equivalent terms
6. **Breach notification** — per §4 timeline
7. **Data subject rights** — assist with access/rectify/erase/portability within SLA
8. **Cross-border transfer** — per PDPA ss.28–29 and BOT data residency
9. **Audit right** — [บริษัท / Company], QSA and **BOT** may audit
10. **Compliance reporting** — annual AoC/RoC (PCI), SOC 2 Type II, ISO 27001
11. **BCP/DR & SLA** — RTO/RPO, availability, service continuity
12. **Return/destruction on termination** — with certificate of destruction
13. **Liability & indemnity** — including regulator and scheme fines
14. **Exit / transition plan** — offboarding without service or compliance impact

---

## 3. Data Categories & Lawful Basis per Counterparty

| Counterparty | Data processed | Prohibited | Lawful basis (PDPA) |
|---|---|---|---|
| **Acquirer / Sponsor Bank** | PAN (via network token/ISO 8583), auth data, amount, MCC, merchant ID | CVV2 post-authorization, PIN | Contract (s.24(3)) + legitimate interest |
| **3DS provider** | PAN/DPAN, device data, browser fingerprint, cardholder name, billing | CVV2 (outside 3DS flow), track data | Contract + legal obligation (SCA/scheme mandate) |
| **Tokenization Vault** | full PAN (transient, to mint token), expiry | CVV2, PIN, persisted track data | Contract |
| **ITMX** | account/PromptPay proxy, amount, PAN for local card switch | CVV2, PIN block outside HSM | Contract + legal obligation (national payment system) |
| **HSM/KMS provider** | ciphertext, key material (inside FIPS boundary) | **plaintext PAN/PII** (must not leave HSM) | Contract |

> **Principle:** Per `../ARCHITECTURE.md` §6 the core system only sees `card_brand` + `card_last4`; **never store full PAN, CVV/CVV2, PIN, or full track** in the operational DB — aligned with PCI Req 3.2 (no SAD storage post-authorization).

---

## 4. Notification & SLA Timelines (bound into every DPA)

| Event | Provider must notify [บริษัท / Company] within | [บริษัท / Company] notifies regulator within |
|---|---|---|
| Suspected/confirmed data breach | **≤ 24 h** from detection | **PDPC within 72 h** (PDPA s.37(4)); **BOT** per cyber-incident notification (typically ≤ 24 h for severe events) |
| PCI Account Data Compromise (ADC) | **Immediate** | Notify scheme (Visa/Mastercard) + acquirer + QSA per ADC procedure |
| Major availability outage | **≤ 1 h** | BOT per operational-disruption thresholds |
| Failed ASV scan / critical vuln | **≤ 48 h** | Within PCI remediation plan |
| Sub-processor / material change | **≥ 30 days in advance** | Notify BOT if material outsourcing |

**Minimum recommended SLA:** Availability ≥ 99.95% (per `../ARCHITECTURE.md` §8), RPO ≤ 5 min, RTO ≤ 30 min.

---

## 5. Counterparty-Specific Templates

### 5.1 Acquirer / Sponsor Bank

- **Contract type:** Merchant Acquiring / Sponsorship Agreement + Data Sharing Addendum
- **Key terms:** access to authorization & settlement network, chargeback/dispute allocation, PAN transmission via network token, fraud/chargeback liability allocation
- **PCI:** acquirer provides **AoC** and enrolls [บริษัท / Company] in TPSP monitoring
- **BOT:** include right-to-audit allowing BOT to inspect the sponsor via [บริษัท / Company]

> **[TODO / ASSUMPTION]** No real **sponsor bank/acquirer** selected yet (see `../ROADMAP.md` §5 critical path) — fill counterparty name, BIN sponsorship, and scheme certification timeline after LOI signing.

### 5.2 3-D Secure Provider (EMV 3DS)

- **Scope:** ACS/DS integration per **EMV 3DS 2.x** for SCA and liability shift to issuer
- **Data:** device/browser data, DPAN — bind data minimization and retention ≤ scheme limits
- **PCI:** 3DS server in scope; provider must furnish AoC/SDP compliance
- **Fallback:** if 3DS unavailable, fail closed per `../ARCHITECTURE.md` §2

### 5.3 Tokenization Vault

- **Scope:** mint/detokenize, envelope encryption with HSM-held keys (`../ARCHITECTURE.md` §7)
- **Requirements:** vault in a **segmented network** separate from the core; client-side single-use payment token
- **PCI:** primary CDE TPSP — must hold **Level 1 RoC/AoC**, quarterly ASV, annual pentest
- **Destruction:** crypto-erase keys at termination + certificate

### 5.4 ITMX (National Interbank Transaction Management and Exchange)

- **Contract type:** Membership/Interchange Agreement + Data Sharing (both parties are controllers)
- **Scope:** local card switch / PromptPay proxy, domestic clearing and settlement
- **Data residency:** data kept in Thailand per BOT requirements (`../ARCHITECTURE.md` §8)
- **AML:** bind transaction-data sharing for AMLO reporting as required by law

> **[TODO / ASSUMPTION]** **ITMX** membership terms and data flow to be confirmed against actual membership rules after connectivity certification.

### 5.5 HSM / KMS Provider

- **Scope:** key protection at **FIPS 140-2/3 Level 3**, dual control, split knowledge (PCI Req 3.6–3.7)
- **Principle:** **plaintext PAN/PII must not leave the HSM boundary**; provider sees ciphertext only
- **Bound services:** key rotation, key ceremony, HSM firmware update, HA/DR of the key store
- **PCI:** provider must furnish AoC or FIPS certificate + right-to-audit

> **[TODO / ASSUMPTION]** No real **QSA vendor** or **HSM/KMS provider** selected yet (cloud KMS+CloudHSM vs on-prem). The **paid-up capital of THB 50M** figure cites the Acquiring threshold in `../COMPLIANCE-TH.md` §2 — confirm the actual paid-up amount with Finance.

---

## 6. Pre-Signing Checklist (all counterparties)

- [ ] Controller/processor roles clearly stated per PDPA s.40
- [ ] TOMs (technical & organizational measures) attached as an annex
- [ ] Current AoC/RoC (PCI) or SOC 2 / ISO 27001
- [ ] Breach notification ≤ 24 h + regulator timeline
- [ ] Right-to-audit covers [บริษัท / Company], QSA and **BOT**
- [ ] Sub-processor list + right to object
- [ ] Cross-border transfer meets PDPA ss.28–29 + BOT data residency
- [ ] Data return/destruction + certificate at termination
- [ ] Exit/transition plan and SLA/BCP fully bound
