# ตารางการเก็บรักษาและลบข้อมูล (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 · ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท · PCI-DSS v4.0 Level 1
> เอกสารเลขที่ **11 — นโยบายและตารางการเก็บรักษาและลบข้อมูล (Data Retention & Deletion Schedule)**
> เจ้าของเอกสาร: **[บริษัท / Company]** · เวอร์ชัน 1.0 · จัดทำ 2026-07-22 · ทบทวนทุก 12 เดือน
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงเทคนิค/ปฏิบัติการ ไม่ใช่คำแนะนำทางกฎหมาย ต้องผ่านการตรวจ
> โดยที่ปรึกษากฎหมายและ QSA ก่อนยื่นจริง เอกสารอ้างอิงร่วม: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`

---

## 1. วัตถุประสงค์และขอบเขต

นโยบายนี้กำหนด **ระยะเวลาการเก็บรักษา (retention)** และ **วิธีการลบข้อมูลอย่างปลอดภัย (secure deletion)**
สำหรับข้อมูลทุกประเภทที่ [บริษัท / Company] จัดเก็บในฐานะผู้ให้บริการ Acquiring โดยยึดหลัก **เก็บให้น้อยที่สุด
เท่าที่จำเป็นตามกฎหมายและการปฏิบัติงาน (data minimization / storage limitation)** ตามมาตรา 37 แห่ง PDPA และ
PCI-DSS v4.0 Requirement 3.2 (เก็บ cardholder data เท่าที่จำเป็นต่อธุรกิจ/กฎหมาย)

ขอบเขตครอบคลุม: ระบบ operational DB, tokenization vault, ledger, audit log, log ระบบ, ข้อมูล KYC/AML,
เอกสารสัญญา merchant, สำรองข้อมูล (backup), และข้อมูลที่ outsource ให้ผู้ให้บริการภายนอก

> **หลักสำคัญที่ยึดตลอดทั้งเอกสาร:** ตามที่ระบุใน `ARCHITECTURE.md` ข้อ 6 —
> **ห้ามเก็บ full PAN, CVV/CVV2/CVC2, PIN/PIN block, full magnetic track (SAD) ใน operational DB โดยเด็ดขาด**
> เก็บได้เพียง `card_brand` + `card_last4` เพื่อการแสดงผล ส่วน PAN ที่จำเป็นให้เก็บเป็น **token** ใน vault ที่แยก scope

---

## 2. บทบาทและความรับผิดชอบ (Roles & Responsibilities)

| บทบาท | ความรับผิดชอบด้าน retention/deletion |
|-------|--------------------------------------|
| **Data Protection Officer (DPO)** | เจ้าของนโยบาย, ตอบคำขอสิทธิเจ้าของข้อมูล (DSAR), อนุมัติการลบตามคำขอ PDPA, ประสาน PDPC |
| **Compliance / MLRO (Money Laundering Reporting Officer)** | กำหนด hold ตาม ปปง./AMLO, legal hold, การรายงาน ธปท. |
| **CISO / DevSecOps** | คุมกระบวนการ crypto-erasure, key destruction, ตรวจสอบ secure deletion, จัดการ HSM/KMS |
| **DBA / SRE** | รัน purge job, จัดการ backup lifecycle, ยืนยันผลการลบ |
| **Head of Engineering** | ดูแล data classification ในโค้ด, TTL, retention config |
| **Internal Audit** | ตรวจสอบการปฏิบัติตาม retention schedule รายไตรมาส |

การลบข้อมูลถาวรทุกครั้งต้องมี **dual control** (ผู้ขอ + ผู้อนุมัติ) และบันทึกลง `audit_log` (append-only)

---

## 3. การจัดชั้นข้อมูล (Data Classification)

| ชั้น | นิยาม | ตัวอย่าง |
|------|-------|---------|
| **C1 — Cardholder Data (CHD)** | ข้อมูลในขอบเขต PCI ที่เก็บได้แบบมีเงื่อนไข | PAN (token/เข้ารหัส), card expiry, cardholder name |
| **C1S — Sensitive Auth Data (SAD)** | ห้ามเก็บหลัง authorization | CVV2/CVC2, full track, PIN/PIN block |
| **C2 — ข้อมูลส่วนบุคคล (PII)** | ตาม PDPA | ชื่อ, อีเมล, เบอร์โทร, เอกสาร KYC ของผู้ถือบัตร/ผู้แทน merchant |
| **C3 — ข้อมูลธุรกรรมและบัญชี** | บันทึกการเงิน/บัญชี | payments, ledger_entries, refunds, settlement |
| **C4 — ข้อมูลระบบ/ความมั่นคงปลอดภัย** | log, audit trail | audit_log, access log, WAF/IDS log |
| **C5 — ข้อมูล KYC/AML** | เอกสารพิสูจน์ตัวตน, screening | CDD/EDD records, sanction screening, STR/SAR |

---

## 4. ตารางการเก็บรักษา (Master Retention Schedule)

> ระยะเวลาที่นานที่สุดในบรรดาหน้าที่ตามกฎหมายเป็นตัวกำหนด (longest-applicable-rule wins) เมื่อพ้นกำหนดต้องลบ/ทำลายภายในรอบ purge

| # | ประเภทข้อมูล | ชั้น | ระยะเก็บ | เริ่มนับจาก | ฐานทางกฎหมาย/เหตุผล | วิธีลบเมื่อครบ |
|---|-------------|------|---------|-------------|----------------------|---------------|
| 1 | **Card token (PAN token) ใน vault** | C1 | **≤ อายุ mandate / สูงสุด 15 เดือน** สำหรับ CIT recurring token; single-use token ลบทันทีหลังใช้ (≤ 24 ชม.) | วันสร้าง token / วันใช้ล่าสุด | PCI-DSS 3.2, EMV 3DS/network tokenization mandate; data minimization | **Crypto-erasure** (ทำลาย DEK) + purge record |
| 2 | **PAN เข้ารหัส (ถ้ามีนอก vault)** | C1 | เท่าที่จำเป็นต่อธุรกรรม ห้ามเกินความจำเป็น | วันธุรกรรม | PCI-DSS 3.2.1 | Crypto-erasure |
| 3 | **Sensitive Auth Data (CVV2/track/PIN)** | C1S | **0 (ห้ามเก็บ)** — ลบทันทีหลัง authorization | — | PCI-DSS 3.3.1 (ห้ามเก็บ SAD หลัง auth) | ไม่ persist; memory-scrub |
| 4 | **payments / refunds (บันทึกธุรกรรม)** | C3 | **10 ปี** | สิ้นปีบัญชีของธุรกรรม | ป.พ.พ. อายุความ + ปปง./AMLO เก็บ 5 ปี + มาตรฐานบัญชี/ภาษี (ปกติ 5 ปี, ใช้เกณฑ์นานสุด) | Soft-delete → hard purge |
| 5 | **ledger_entries (append-only)** | C3 | **10 ปี** | วันลงรายการ | source of truth สำหรับ settlement/reconciliation; audit ธปท. | Archive → purge |
| 6 | **card_last4 / card_brand** | C2 | ตามอายุ payment (สูงสุด 10 ปี) | วันธุรกรรม | แสดงผล/dispute; ไม่ใช่ CHD เต็ม | Purge พร้อม payment |
| 7 | **KYC/CDD/EDD ของ merchant & ผู้แทน** | C5 | **10 ปี หลังยุติความสัมพันธ์** (ขั้นต่ำ 5 ปี ตาม ปปง.) | วันปิดบัญชี merchant | พ.ร.บ. ปปง. ม.22 (เก็บ ≥ 5 ปี); ธปท. outsourcing/AML | Secure delete เอกสาร + record |
| 8 | **STR/SAR & sanction screening logs** | C5 | **10 ปี** | วันรายงาน/วัน screen | ปปง./AMLO | Secure delete |
| 9 | **audit_log (การเปลี่ยนสถานะทุกอย่าง)** | C4 | **≥ 12 เดือน online + รวม 10 ปี archive** (อย่างน้อย 3 เดือนล่าสุดพร้อมสืบค้นทันที) | วันเกิด event | PCI-DSS 10.5.1 (≥12 เดือน, 3 เดือน readily available); ธปท. audit | Immutable → purge |
| 10 | **Access / authentication logs** | C4 | **≥ 12 เดือน** | วันเกิด event | PCI-DSS 10.5.1 | Purge |
| 11 | **Application / infra / WAF / IDS logs** | C4 | 12–18 เดือน | วันเกิด event | security monitoring, forensics | Purge/rotate |
| 12 | **webhook_events** | C3 | 90 วัน (payload) / metadata 10 ปี | วันสร้าง | operational retry + reconciliation | Purge payload |
| 13 | **ข้อมูล 3DS (EMV 3DS auth result, ECI, DS/ACS ref)** | C2/C3 | เก็บ **ผลลัพธ์/หลักฐาน liability shift 13 เดือน+** (ครอบคลุมรอบ chargeback); ไม่เก็บ CAVV ดิบเกินจำเป็น | วัน authenticate | หลักฐาน chargeback/dispute (Visa/Mastercard ~120–540 วัน) | Purge |
| 14 | **Backup / snapshot** | ตามข้อมูลต้นทาง | ≤ 35 วัน (operational) + archive ตามข้อ 4/5 | วันสร้าง backup | RPO ≤ 5 นาที (ARCHITECTURE §8); DR | Crypto-erasure ของ backup key |
| 15 | **PII ที่ merchant/ผู้ใช้ขอลบ (DSAR)** | C2 | ลบภายใน **30 วัน** เว้นแต่มีหน้าที่เก็บตามกฎหมาย | วันรับคำขอ | PDPA ม.33 (สิทธิขอลบ) | Secure delete/anonymize |

> **การชนกันของกฎ (conflict):** หากมีหน้าที่เก็บตามกฎหมาย (เช่น ปปง. หรือ legal hold) คำขอลบตาม PDPA
> จะถูก **ระงับเฉพาะข้อมูลส่วนที่มีหน้าที่เก็บ** และแจ้งเจ้าของข้อมูลถึงเหตุผลตาม PDPA ม.33 วรรคท้าย

---

## 5. เพดานการเก็บ Card Token (Card Token Retention Limits)

1. **Single-use / one-time token** (จากขั้นตอน `POST /v1/payments` ตาม ARCHITECTURE §5.1): ใช้ครั้งเดียวและ
   **หมดอายุอัตโนมัติภายใน 15 นาที** หากไม่ถูกใช้ และถูก **ลบภายใน 24 ชั่วโมง** หลังธุรกรรมจบ
2. **Recurring / card-on-file token (CIT/MIT):** เก็บได้เฉพาะเมื่อมี mandate ของผู้ถือบัตร และผูกกับ
   **network tokenization (Visa VTS / Mastercard MDES)** เมื่อทำได้ เพดานสูงสุด **15 เดือนนับจากการใช้ล่าสุด**
   หากไม่มีการใช้จะถูกทำ crypto-erasure โดยอัตโนมัติ (idle-token expiry job รายวัน)
3. **การยกเลิก mandate / ปิด merchant:** token ทั้งหมดของ scope นั้นถูกทำลายภายใน **72 ชั่วโมง**
4. token ทุกตัวเก็บใน **tokenization vault ที่แยก network segment** (ARCHITECTURE §4, §7) โดย mapping
   token↔PAN เข้ารหัสด้วย **envelope encryption** และคีย์อยู่ใน **HSM/KMS** เท่านั้น
5. ระบบหลัก (operational DB) เก็บเพียง **token reference + card_last4 + card_brand** — ไม่มี PAN

---

## 6. การลบอย่างปลอดภัย (Secure Deletion Procedures)

| วิธี | ใช้กับ | รายละเอียด |
|------|--------|-----------|
| **Crypto-erasure (แนะนำหลัก)** | CHD, token, backup, ข้อมูลเข้ารหัสทั้งหมด | ทำลาย **Data Encryption Key (DEK)** ใน HSM/KMS → ciphertext กู้คืนไม่ได้; สอดคล้อง NIST SP 800-88 (Purge) และ PCI-DSS 3.2/3.5 |
| **Logical purge (hard delete)** | record ใน DB | `DELETE` + `VACUUM`/reclaim, ยืนยันด้วย row-count = 0; ไม่ใช่แค่ soft-delete flag |
| **Anonymization / tokenization** | PII ที่ต้องคง record ธุรกรรมไว้ | แทน PII ด้วยค่า irreversible เพื่อคงความสมบูรณ์ของ ledger โดยไม่ระบุตัวบุคคล |
| **Media sanitization** | disk/SSD ที่ปลดระวาง | NIST SP 800-88 Purge/Destroy; ออก **Certificate of Destruction** เก็บ 10 ปี |
| **Memory scrub** | SAD/PAN ระหว่างประมวลผล | zero-out buffer ทันทีหลังใช้ ไม่ swap ลง disk |

**ขั้นตอนมาตรฐานเมื่อครบกำหนด (purge cycle):**
1. Retention engine (รายวัน) คัดเลือก record ที่พ้นกำหนดตามตารางข้อ 4
2. ตรวจ **hold register** (legal hold / AML hold / คดีความ) — ถ้ามี hold ให้ข้าม
3. Dual-control อนุมัติสำหรับข้อมูล C1/C5
4. รัน crypto-erasure/purge → รวมถึง **backup และ replica** (ตาม RPO/DR)
5. บันทึกผล (จำนวน record, timestamp, ผู้อนุมัติ) ลง `audit_log`
6. Internal Audit สุ่มตรวจรายไตรมาส

---

## 7. Backup, Replica และ Data Residency

- Backup/replica ต้องเข้ารหัส และการลบข้อมูลต้องครอบคลุม backup ผ่าน **crypto-erasure ของ backup key**
  เมื่อ retention ของ backup หมด (≤ 35 วัน operational) โดยไม่ต้องเขียนทับ media รายชิ้น
- ข้อมูลเก็บใน **ประเทศไทย** ตามข้อกำหนด ธปท./PDPA (ARCHITECTURE §8) การโอนข้ามพรมแดนต้องเป็นไปตาม PDPA ม.28–29
- RPO ≤ 5 นาที / RTO ≤ 30 นาที — การออกแบบ purge ต้องไม่กระทบ DR

---

## 8. คำขอสิทธิเจ้าของข้อมูล (DSAR) และ Legal Hold

- คำขอลบ/เข้าถึงตาม PDPA รับผ่านช่องทางที่ประกาศ → DPO ตอบภายใน **30 วัน**
- ข้อมูลที่มีหน้าที่เก็บตามกฎหมาย (ปปง. 5–10 ปี, ธุรกรรม 10 ปี) จะไม่ถูกลบจนพ้นกำหนด และแจ้งเหตุผล
- **Legal hold** ระงับการ purge ทันทีเมื่อมีคดี/คำสั่งทางการ; บันทึกใน hold register พร้อมผู้อนุมัติ

---

## 9. ข้อสมมติและสิ่งที่ยังไม่ยุติ (Assumptions & TODO)

> [!IMPORTANT]
> รายการต่อไปนี้เป็น **สมมติฐาน/TODO** ที่ต้องยืนยันก่อนยื่นจริง — ยังไม่ถือเป็นข้อเท็จจริงที่ล็อกแล้ว:
> - **Sponsor bank / card scheme:** ยังไม่เลือก — เงื่อนไข retention ของ Visa/Mastercard และหน้าต่าง chargeback
>   (ข้อ 13) ต้องปรับตามสัญญาจริง (TODO)
> - **QSA vendor:** ยังไม่ว่าจ้าง — การตีความ PCI-DSS v4.0 บางข้อ (เช่น เพดาน token) ต้องยืนยันกับ QSA ที่เลือก (TODO)
> - **ทุนจดทะเบียนชำระแล้ว:** สมมติ 50 ล้านบาทตามเกณฑ์ Acquiring — ต้องยืนยันตัวเลขจริงที่ชำระแล้ว (TODO)
> - **ระยะเก็บ ปปง. ที่แม่นยำ (5 vs 10 ปี):** ใช้เกณฑ์นานสุด 10 ปีเชิงระมัดระวัง — ยืนยันกับที่ปรึกษากฎหมาย AML (TODO)
> - **ผู้ให้บริการ HSM/KMS และ backup vendor:** ยืนยัน SLA การ crypto-erasure และ certificate of destruction (TODO)

---
---

# Data retention & deletion schedule incl. card token retention limits and secure deletion (English)

> Supporting document for the **Acquiring Payment Service** license application under the Payment Systems
> Act B.E. 2560 · Registered paid-up capital THB 50M · PCI-DSS v4.0 Level 1
> Document **11 — Data Retention & Deletion Policy and Schedule**
> Owner: **[บริษัท / Company]** · Version 1.0 · Issued 2026-07-22 · Reviewed every 12 months
>
> **Note:** This is a technical/operational document, not legal advice. It must be reviewed by legal counsel
> and the QSA before submission. Companion docs: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`.

---

## 1. Purpose & Scope

This policy defines the **retention periods** and **secure deletion methods** for every data category held by
**[บริษัท / Company]** as an Acquiring service provider, applying **data minimization / storage limitation** per
PDPA §37 and PCI-DSS v4.0 Requirement 3.2 (keep cardholder data only as long as required for business/legal need).

Scope: operational DB, tokenization vault, ledger, audit log, system logs, KYC/AML data, merchant contracts,
backups, and data outsourced to third parties.

> **Governing rule throughout this document** (per `ARCHITECTURE.md` §6): **Full PAN, CVV/CVV2/CVC2, PIN/PIN block,
> and full magnetic track (Sensitive Authentication Data) MUST NEVER be stored in the operational DB.** Only
> `card_brand` + `card_last4` are retained for display; any required PAN is stored as a **token** in the
> segmented vault.

---

## 2. Roles & Responsibilities

| Role | Retention/deletion responsibility |
|------|-----------------------------------|
| **Data Protection Officer (DPO)** | Policy owner; handles DSARs; approves PDPA deletions; PDPC liaison |
| **Compliance / MLRO** | Sets AMLO/ปปง. holds, legal holds, BOT (ธปท.) reporting |
| **CISO / DevSecOps** | Owns crypto-erasure, key destruction, secure-deletion verification, HSM/KMS |
| **DBA / SRE** | Runs purge jobs, backup lifecycle, confirms deletion |
| **Head of Engineering** | Data classification in code, TTLs, retention config |
| **Internal Audit** | Quarterly verification against the retention schedule |

Every permanent deletion requires **dual control** (requester + approver) and is recorded in the append-only `audit_log`.

---

## 3. Data Classification

| Class | Definition | Examples |
|-------|-----------|---------|
| **C1 — Cardholder Data (CHD)** | PCI-scope data, conditionally storable | PAN (tokenized/encrypted), card expiry, cardholder name |
| **C1S — Sensitive Auth Data (SAD)** | Must never be stored post-authorization | CVV2/CVC2, full track, PIN/PIN block |
| **C2 — Personal Data (PII)** | Per PDPA | Name, email, phone, KYC docs of cardholder/merchant reps |
| **C3 — Transaction & accounting data** | Financial records | payments, ledger_entries, refunds, settlement |
| **C4 — System / security data** | Logs, audit trail | audit_log, access logs, WAF/IDS logs |
| **C5 — KYC/AML data** | Identity evidence, screening | CDD/EDD records, sanction screening, STR/SAR |

---

## 4. Master Retention Schedule

> The longest applicable legal obligation governs (longest-applicable-rule wins). Once expired, data is deleted/destroyed in the next purge cycle.

| # | Data category | Class | Retention | Clock starts | Legal basis / rationale | Deletion method |
|---|--------------|-------|-----------|--------------|-------------------------|-----------------|
| 1 | **Card token (PAN token) in vault** | C1 | **≤ mandate life / max 15 months** for CIT recurring token; single-use token deleted right after use (≤ 24h) | Token creation / last use | PCI-DSS 3.2, EMV 3DS / network tokenization mandate; data minimization | **Crypto-erasure** (destroy DEK) + purge record |
| 2 | **Encrypted PAN (if any outside vault)** | C1 | Only as long as needed for the transaction | Transaction date | PCI-DSS 3.2.1 | Crypto-erasure |
| 3 | **Sensitive Auth Data (CVV2/track/PIN)** | C1S | **0 (never stored)** — purged immediately after authorization | — | PCI-DSS 3.3.1 (no SAD after auth) | Not persisted; memory scrub |
| 4 | **payments / refunds (transaction records)** | C3 | **10 years** | End of transaction's fiscal year | Civil & Commercial Code limitation + AMLO 5-year + tax/accounting standards (use longest) | Soft-delete → hard purge |
| 5 | **ledger_entries (append-only)** | C3 | **10 years** | Posting date | Source of truth for settlement/reconciliation; BOT audit | Archive → purge |
| 6 | **card_last4 / card_brand** | C2 | Follows payment (max 10 years) | Transaction date | Display/dispute; not full CHD | Purged with payment |
| 7 | **KYC/CDD/EDD (merchant & reps)** | C5 | **10 years after relationship ends** (min 5 yrs per AMLO) | Merchant offboarding date | AMLO Act §22 (retain ≥5 yrs); BOT outsourcing/AML | Secure delete docs + record |
| 8 | **STR/SAR & sanction screening logs** | C5 | **10 years** | Report/screening date | AMLO | Secure delete |
| 9 | **audit_log (all state changes)** | C4 | **≥12 months online + 10 years archived** (last 3 months readily available) | Event time | PCI-DSS 10.5.1 (≥12 mo, 3 mo readily available); BOT audit | Immutable → purge |
| 10 | **Access / authentication logs** | C4 | **≥12 months** | Event time | PCI-DSS 10.5.1 | Purge |
| 11 | **Application / infra / WAF / IDS logs** | C4 | 12–18 months | Event time | Security monitoring, forensics | Purge/rotate |
| 12 | **webhook_events** | C3 | 90 days (payload) / metadata 10 years | Creation | Operational retry + reconciliation | Purge payload |
| 13 | **3DS data (EMV 3DS result, ECI, DS/ACS ref)** | C2/C3 | Keep **liability-shift evidence 13+ months** (covers chargeback window); do not retain raw CAVV beyond need | Authentication date | Chargeback/dispute evidence (Visa/MC ~120–540 days) | Purge |
| 14 | **Backup / snapshot** | Per source | ≤35 days (operational) + archive per rows 4/5 | Backup creation | RPO ≤5 min (ARCHITECTURE §8); DR | Crypto-erasure of backup key |
| 15 | **PII under user/merchant deletion request (DSAR)** | C2 | Delete within **30 days** unless a legal retention duty applies | Request receipt | PDPA §33 (right to erasure) | Secure delete/anonymize |

> **Conflict handling:** Where a legal retention duty exists (e.g. AMLO or legal hold), a PDPA deletion request
> is **suspended only for the portion under that duty**, and the data subject is informed of the reason per PDPA §33.

---

## 5. Card Token Retention Limits

1. **Single-use / one-time token** (from `POST /v1/payments`, ARCHITECTURE §5.1): used once and **auto-expires
   within 15 minutes** if unused; **deleted within 24 hours** after the transaction completes.
2. **Recurring / card-on-file token (CIT/MIT):** stored only with a cardholder mandate and bound to
   **network tokenization (Visa VTS / Mastercard MDES)** where available. Hard cap of **15 months from last use**;
   idle tokens are crypto-erased automatically by a daily idle-token expiry job.
3. **Mandate cancellation / merchant offboarding:** all tokens in that scope are destroyed within **72 hours**.
4. All tokens live in a **network-segmented tokenization vault** (ARCHITECTURE §4, §7); the token↔PAN mapping is
   protected with **envelope encryption** and keys reside **only in HSM/KMS**.
5. The operational DB stores only a **token reference + card_last4 + card_brand** — never the PAN.

---

## 6. Secure Deletion Procedures

| Method | Applies to | Detail |
|--------|-----------|--------|
| **Crypto-erasure (primary)** | CHD, tokens, backups, all encrypted data | Destroy the **Data Encryption Key (DEK)** in HSM/KMS → ciphertext unrecoverable; aligns with NIST SP 800-88 (Purge) and PCI-DSS 3.2/3.5 |
| **Logical purge (hard delete)** | DB records | `DELETE` + `VACUUM`/space reclaim, verified by row-count = 0; not a mere soft-delete flag |
| **Anonymization / tokenization** | PII where the transaction record must remain | Replace PII with irreversible values to preserve ledger integrity without identifying the person |
| **Media sanitization** | Decommissioned disks/SSDs | NIST SP 800-88 Purge/Destroy; issue a **Certificate of Destruction** retained 10 years |
| **Memory scrub** | SAD/PAN during processing | Zero out buffers immediately after use; never swap to disk |

**Standard purge cycle when retention expires:**
1. Retention engine (daily) selects records past their schedule (§4).
2. Check the **hold register** (legal / AML holds, litigation) — skip if held.
3. Dual-control approval for C1/C5 data.
4. Run crypto-erasure/purge — including **backups and replicas** (per RPO/DR).
5. Record the result (record count, timestamp, approver) in `audit_log`.
6. Internal Audit spot-checks quarterly.

---

## 7. Backup, Replica & Data Residency

- Backups/replicas are encrypted; deletion covers backups via **crypto-erasure of the backup key** when backup
  retention expires (≤35 days operational), without per-media overwrite.
- Data resides **in Thailand** per BOT/PDPA requirements (ARCHITECTURE §8); cross-border transfers follow PDPA §28–29.
- RPO ≤5 min / RTO ≤30 min — purge design must not compromise DR.

---

## 8. Data Subject Requests (DSAR) & Legal Hold

- PDPA access/erasure requests arrive via the published channel → DPO responds within **30 days**.
- Data under a legal retention duty (AMLO 5–10 yrs, transactions 10 yrs) is retained until expiry, with reasons given.
- **Legal hold** immediately suspends purge on litigation/authority order; recorded in the hold register with approver.

---

## 9. Assumptions & TODO

> [!IMPORTANT]
> The following are **assumptions/TODOs** to confirm before actual submission — not yet locked facts:
> - **Sponsor bank / card scheme:** not yet selected — Visa/Mastercard retention terms and chargeback windows
>   (row 13) must be adjusted to the actual contract (TODO).
> - **QSA vendor:** not yet engaged — some PCI-DSS v4.0 interpretations (e.g. token retention caps) must be
>   confirmed with the chosen QSA (TODO).
> - **Registered paid-up capital:** assumed THB 50M per the Acquiring threshold — confirm the actual paid-up figure (TODO).
> - **Exact AMLO retention (5 vs 10 years):** conservatively using 10 years — confirm with AML legal counsel (TODO).
> - **HSM/KMS and backup vendor:** confirm crypto-erasure SLA and certificate of destruction (TODO).

---

## References

- Payment Systems Act B.E. 2560 (2017) — Bank of Thailand
- Personal Data Protection Act B.E. 2562 (PDPA) — §28–29, §33, §37
- Anti-Money Laundering Act (AMLO / ปปง.) — §22 (record retention ≥5 years)
- PCI-DSS v4.0 — Req 3.2, 3.3.1, 3.5, 10.5.1
- EMV 3-D Secure (3DS) 2.x; Visa VTS / Mastercard MDES network tokenization
- NIST SP 800-88 Rev.1 — Guidelines for Media Sanitization
- Companion docs: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`
