# การออกแบบ Tokenization และการบริหารกุญแจ HSM/KMS (ไทย)

> เอกสารเลขที่ **21** ในชุดเอกสาร Compliance สำหรับการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)** ภายใต้ **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** กำกับโดย **ธนาคารแห่งประเทศไทย (ธปท.)** และคู่ขนานกับ **PCI-DSS Level 1** (ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท)
>
> **สถานะเอกสาร:** ฉบับร่างเพื่อยื่นขออนุมัติ (submission draft) · เวอร์ชัน 0.1 · วันที่ปรับปรุง 2026-07-22
> **เจ้าของเอกสาร:** ประธานเจ้าหน้าที่ด้านความมั่นคงปลอดภัยสารสนเทศ (CISO) และผู้ดูแลระบบบริหารกุญแจ (Key Custodian Team / Cryptographic Key Management) ของ **[บริษัท / Company]**
>
> เอกสารนี้เป็นนโยบายและมาตรฐานภายในและเอกสารประกอบคำขอใบอนุญาต **มิใช่คำแนะนำทางกฎหมาย** — ต้องผ่านการตรวจทานโดยที่ปรึกษากฎหมาย, CISO และ QSA ก่อนบังคับใช้จริง เนื่องจากประกาศ/หลักเกณฑ์ของ ธปท. และข้อกำหนด PCI-DSS อาจปรับปรุงได้

---

## บทสรุปสำหรับผู้บริหาร (Executive Summary)

[บริษัท / Company] ออกแบบให้ **ข้อมูลบัตร (Cardholder Data — PAN) ไม่ถูกจัดเก็บในระบบปฏิบัติการหลัก (operational database)** โดยเด็ดขาด แต่ถูกแทนที่ด้วย **โทเคน (token)** ที่ไม่มีความสัมพันธ์ทางคณิตศาสตร์กับ PAN และไม่สามารถย้อนกลับได้หากไม่มีสิทธิ์เข้าถึง **Tokenization Vault** ที่แยก network segment ออกจากระบบหลัก แนวทางนี้เป็นการ **ลดขอบเขต PCI (scope minimization)** ให้ระบบหลักเห็นเพียง `token`, `card_brand` และ `card_last4` ตามที่ระบุใน `ARCHITECTURE.md`

การเข้ารหัสข้อมูลใช้สถาปัตยกรรม **Envelope Encryption 3 ชั้น** ได้แก่ Root Key (KEK ระดับสูงสุด) → Key-Encrypting Key (KEK) → Data-Encrypting Key (DEK) โดยกุญแจระดับสูงสุด (Root/Master Key) **ไม่เคยออกจาก HSM** ที่ผ่านการรับรอง **FIPS 140-2 Level 3 (หรือ FIPS 140-3)** การเข้าถึงและการดำเนินการกับกุญแจสำคัญทุกครั้งถูกควบคุมด้วยหลักการ **Dual Control (สองคนควบคุม)** และ **Split Knowledge (ความรู้แยกส่วน)** เพื่อให้ไม่มีบุคคลใดบุคคลหนึ่งสามารถประกอบหรือใช้กุญแจได้โดยลำพัง

เอกสารนี้กำหนดนโยบาย มาตรฐาน และขั้นตอนปฏิบัติ (SOP) ที่ตอบสนอง **PCI-DSS v4.0 Requirement 3 (Protect Stored Account Data) และ Requirement 4 (Protect CHD with Strong Cryptography in Transit)** โดยตรง พร้อมเชื่อมโยงกับข้อกำหนดของ ธปท. ด้าน IT risk / cyber resilience, พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA — กำกับโดย PDPC) และภาระ AML/CDD ภายใต้ ปปง./AMLO

> **ความเชื่อมโยงกับสถาปัตยกรรม:** เอกสารนี้ให้รายละเอียดเชิงลึกของการควบคุมที่อ้างถึงใน `ARCHITECTURE.md` หัวข้อ 7 (Encryption at rest, Key management, Scope minimization) และเอกสาร `13-it-risk-management.md`

---

## 1. ฐานกฎหมายและมาตรฐานอ้างอิง

| กฎหมาย / มาตรฐาน | สาระสำคัญที่เกี่ยวข้อง |
|---|---|
| **พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560** — **ธปท.** | เงื่อนไขใบอนุญาต Acquiring กำหนดให้มีการควบคุมความมั่นคงปลอดภัยของข้อมูลและระบบบริหารความเสี่ยง IT ที่ได้มาตรฐาน |
| **แนวปฏิบัติ ธปท. ด้าน IT Risk & Cyber Resilience** | การเข้ารหัสข้อมูลสำคัญ, การบริหารกุญแจ, การควบคุมการเข้าถึง, data residency ในไทย |
| **PCI-DSS v4.0 — Requirement 3** | การคุ้มครองข้อมูลบัญชีที่จัดเก็บ: เข้ารหัส PAN, บริหารกุญแจ, dual control, split knowledge, key rotation, key retirement (3.5–3.7) |
| **PCI-DSS v4.0 — Requirement 4** | เข้ารหัส CHD ระหว่างการส่งผ่านเครือข่ายสาธารณะด้วย strong cryptography (TLS 1.2+) |
| **PCI-DSS v4.0 — Requirement 8 & 12** | การพิสูจน์ตัวตน (MFA), บทบาทหน้าที่ key custodian และหนังสือรับทราบหน้าที่ (acknowledgement) |
| **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** — **PDPC** | มาตรการรักษาความมั่นคงปลอดภัยที่เหมาะสม (มาตรา 37), การแจ้งเหตุละเมิดภายใน 72 ชั่วโมง; PAN เป็นข้อมูลส่วนบุคคล |
| **พ.ร.บ. ป้องกันและปราบปรามการฟอกเงิน (ปปง./AMLO)** | ต้องคง audit trail ของ detokenization ที่เชื่อมโยงธุรกรรมเพื่อรองรับการสอบสวน/รายงาน STR |
| **EMV 3-D Secure (3DS) 2.x** | โทเคนและกุญแจที่เกี่ยวกับ 3DS device/session ต้องอยู่ในกรอบการบริหารกุญแจเดียวกัน |
| **FIPS 140-2 Level 3 / FIPS 140-3** | มาตรฐานการรับรองโมดูลเข้ารหัส (HSM) |
| **NIST SP 800-57 / SP 800-38 / ISO 11568** | หลักการบริหารวงจรชีวิตกุญแจและ key lengths |

> **[TODO / ข้อสมมติ]** ต้องยืนยันเลขที่/ปีของประกาศ ธปท. ด้าน IT risk & cyber resilience ฉบับล่าสุด ณ วันยื่น และปรับการอ้างอิงให้ตรง โดยที่ปรึกษากฎหมายและ QSA

---

## 2. ขอบเขตและนิยาม (Scope & Definitions)

**ขอบเขต:** ครอบคลุมกุญแจเข้ารหัสทั้งหมดที่คุ้มครองข้อมูลบัตร (CHD) และข้อมูลลับ (Sensitive Authentication Data — SAD ที่ห้ามจัดเก็บหลัง authorization), Tokenization Vault, HSM/KMS, ระบบ payment core (Go/Fiber), ฐานข้อมูล และ backup ทั้งหมดของ [บริษัท / Company]

| คำ | นิยาม |
|---|---|
| **PAN** | Primary Account Number — เลขบัตร 13–19 หลัก |
| **CHD / SAD** | Cardholder Data / Sensitive Authentication Data (CVV2, PIN, full track — **ห้ามจัดเก็บ**) |
| **Token** | ค่าแทน PAN ที่ไม่สามารถย้อนกลับได้ทางคณิตศาสตร์ (irreversible token) สร้างและจับคู่ใน Vault เท่านั้น |
| **Detokenization** | การเรียกคืน PAN จาก token — ทำได้เฉพาะภายใน CDE ผ่าน API ที่พิสูจน์ตัวตนแล้ว |
| **Root/Master Key** | กุญแจระดับสูงสุด สร้างและใช้งานภายใน HSM เท่านั้น ไม่เคย export เป็น plaintext |
| **KEK** | Key-Encrypting Key — กุญแจสำหรับเข้ารหัส DEK |
| **DEK** | Data-Encrypting Key — กุญแจที่ใช้เข้ารหัส PAN โดยตรง (AES-256-GCM) |
| **Dual Control** | ต้องใช้บุคคลอย่างน้อย 2 คนร่วมกันจึงจะดำเนินการกับกุญแจสำคัญได้ |
| **Split Knowledge** | ไม่มีบุคคลใดรู้ส่วนประกอบของกุญแจทั้งหมด แต่ละคนรู้เพียง key component/share เดียว |
| **Key Custodian** | เจ้าหน้าที่ที่ได้รับมอบหมายให้ถือครองและจัดการ key component ตาม PCI Req 3.6/12.3 |
| **CDE** | Cardholder Data Environment ตามนิยาม PCI-DSS |

---

## 3. สถาปัตยกรรม Tokenization

### 3.1 หลักการ

1. **Irreversible tokens** — โทเคนสร้างแบบสุ่ม (random, high-entropy) และจับคู่กับ PAN ใน token vault ผ่านตารางค้นหา (lookup) ที่เข้ารหัส **ไม่ใช้** format-preserving encryption ที่ย้อนกลับได้จาก token โดยตรง
2. **Vault แยก segment** — Token Vault และ HSM อยู่ใน network zone ที่แยกจากระบบหลัก มี firewall/ACL เข้มงวด และเข้าถึงได้เฉพาะผ่าน mTLS + service identity
3. **ระบบหลักไม่แตะ PAN** — payment core เก็บเพียง `token`, `card_brand`, `card_last4`; ไม่มี PAN เต็มในตาราง `payments`, log, หรือ webhook payload
4. **Single-use payment token ฝั่ง client** — ตามข้อ 5.1 ของ `ARCHITECTURE.md` PAN ถูกส่งจาก client ไป Vault โดยตรง ไม่ผ่าน server ของ merchant

### 3.2 การไหลของ Tokenization / Detokenization

| ขั้นตอน | การดำเนินการ | ผู้เกี่ยวข้อง |
|---|---|---|
| **Tokenize** | client ส่ง PAN → Vault → HSM สร้าง token + เข้ารหัส PAN ด้วย DEK → เก็บ ciphertext + mapping → คืน token | client, Vault, HSM |
| **Store** | payment core เก็บเฉพาะ token + last4 + brand | payment core |
| **Detokenize** | เฉพาะเมื่อ authorize/settle/chargeback → payment core เรียก Vault API (mTLS + JWT scope=`detok`) → Vault ตรวจสิทธิ์ → คืน PAN ชั่วคราวในหน่วยความจำเท่านั้น | payment core (ใน CDE), Vault |
| **Audit** | ทุกครั้งของ detokenize เขียน `audit_log` (ใคร/เมื่อไร/token/เหตุผล/txn_id) แบบ append-only | ทุกฝ่าย |

> **ข้อห้าม:** ห้าม detokenize เพื่อแสดงผล PAN เต็มบน UI/รายงาน ยกเว้นกรณีจำเป็นตามกฎหมาย (เช่น การสอบสวน ปปง.) และต้องมีการอนุมัติ Dual Control + บันทึกเหตุผล

---

## 4. Envelope Encryption (การเข้ารหัสแบบซองจดหมาย 3 ชั้น)

### 4.1 ลำดับชั้นกุญแจ (Key Hierarchy)

```
Root/Master Key (RK)         ← สร้าง & ใช้ภายใน HSM เท่านั้น, ไม่เคย export plaintext
        │  (เข้ารหัส/ปลดล็อก)
        ▼
Key-Encrypting Key (KEK)     ← เก็บในรูป wrapped (RK-encrypted), โหลดใน HSM เมื่อใช้งาน
        │  (เข้ารหัส/ปลดล็อก)
        ▼
Data-Encrypting Key (DEK)    ← per-tenant / per-partition, เก็บในรูป KEK-encrypted ข้าง ciphertext
        │  (AES-256-GCM)
        ▼
PAN ciphertext + auth tag    ← จัดเก็บใน Token Vault
```

### 4.2 พารามิเตอร์การเข้ารหัส

| รายการ | ค่ามาตรฐาน |
|---|---|
| อัลกอริทึมข้อมูล (DEK) | **AES-256-GCM** (authenticated encryption) |
| อัลกอริทึม wrap (KEK/RK) | AES-256-KW หรือ RSA-3072/RSA-4096 (ตาม HSM) |
| ความยาวกุญแจขั้นต่ำ | สมมาตร ≥ 256-bit; อสมมาตร RSA ≥ 3072-bit / ECC ≥ P-256 |
| แหล่งสุ่ม | HSM TRNG (FIPS-approved) |
| การจัดเก็บ DEK | เก็บในรูป **KEK-encrypted** ติดกับ ciphertext ไม่เก็บ plaintext DEK ลงดิสก์ |
| การจัดเก็บ RK | ภายใน HSM เท่านั้น (non-exportable) |

### 4.3 การแยกกุญแจ (Key Separation)

- **DEK แยกต่อวัตถุประสงค์** — CHD storage, backup, log encryption, 3DS session ใช้ DEK คนละชุด
- **ไม่ใช้กุญแจ production ร่วมกับ non-production** — dev/staging ใช้กุญแจแยกและข้อมูลปลอม (ห้าม PAN จริงใน non-prod)

---

## 5. การหมุนเวียนกุญแจ (Key Rotation)

### 5.1 กำหนดการหมุนเวียนตามรอบ (Cryptoperiod)

| ประเภทกุญแจ | รอบหมุนเวียนปกติ | วิธีการ |
|---|---|---|
| **DEK (ข้อมูลบัตร)** | อย่างน้อย **ทุก 12 เดือน** หรือเมื่อถึงขีดจำกัดปริมาณข้อมูลที่เข้ารหัส | สร้าง DEK ใหม่, re-encrypt ข้อมูลใหม่ด้วย DEK ใหม่ (lazy หรือ batch) |
| **KEK** | อย่างน้อย **ทุก 12–24 เดือน** | re-wrap DEK ทั้งหมดด้วย KEK ใหม่ ภายใน HSM |
| **Root/Master Key** | ทุก **1–2 ปี** หรือเมื่อมีเหตุ (ceremony) | Key Ceremony ภายใต้ Dual Control + Split Knowledge |
| **TLS/mTLS cert** | ≤ **12 เดือน** (แนะนำ 90 วันหากอัตโนมัติ) | หมุนผ่าน ACME/PKI ภายใน |

### 5.2 การหมุนเวียนแบบฉุกเฉิน (Emergency Rotation / Key Compromise)

หมุนเวียนทันทีเมื่อ: สงสัยว่ากุญแจรั่ว, key custodian ลาออก/เปลี่ยนบทบาท, พบช่องโหว่ในอัลกอริทึม, หรือ QSA/ธปท. สั่งการ

| ขั้นตอน | เวลาเป้าหมาย (SLA) |
|---|---|
| ประกาศเหตุ + เรียก Key Ceremony ฉุกเฉิน | ≤ 4 ชั่วโมง |
| สร้างกุญแจใหม่ + re-wrap/re-encrypt | ≤ 24 ชั่วโมง (ขึ้นกับปริมาณข้อมูล) |
| เพิกถอน (retire) กุญแจเดิม + บันทึก | ทันทีหลัง re-encrypt เสร็จ |
| แจ้ง PDPC/ธปท. (หากเข้าข่ายละเมิดข้อมูล) | ตามกำหนด (PDPA: ภายใน 72 ชม.) |

### 5.3 การเลิกใช้และทำลายกุญแจ (Key Retirement & Destruction)

- กุญแจที่หมุนออกจะเปลี่ยนสถานะเป็น **retired** ใช้ได้เฉพาะ **decrypt ข้อมูลเก่า** จนกว่าจะ re-encrypt ครบ
- เมื่อไม่มีข้อมูลใดอ้างอิงกุญแจแล้ว → **destroy** (crypto-shredding) ภายใน HSM และบันทึกใน key inventory
- ทุกการทำลายต้องมี Dual Control และ audit trail

---

## 6. Dual Control และ Split Knowledge

### 6.1 หลักการ (PCI-DSS Req 3.6 / 3.7 / 12.3)

- **Split Knowledge:** Root/Master Key (หรือ KEK ที่โหลดด้วยมือ) แบ่งเป็น **3 key components (shares)** ด้วย secret sharing (เช่น Shamir k-of-n, กำหนด **2-of-3**) — ไม่มีบุคคลใดรู้กุญแจเต็ม
- **Dual Control:** การประกอบ/ใช้/ทำลายกุญแจต้องมี key custodian อย่างน้อย **2 คน** ปรากฏตัวพร้อมกัน

### 6.2 บทบาท Key Custodian

| บทบาท | ผู้รับผิดชอบ | หน้าที่ |
|---|---|---|
| **Key Custodian A** | Security Engineer (DevSecOps) | ถือ key component / smartcard #1 + PIN |
| **Key Custodian B** | SRE/Infra Lead | ถือ key component / smartcard #2 + PIN |
| **Key Custodian C (backup)** | CISO หรือผู้แทน | ถือ key component #3 (สำรอง, ในตู้เซฟแยก) |
| **Ceremony Witness** | Internal Audit / Compliance | สังเกตการณ์ + ลงนามบันทึก ไม่ถือ component |
| **Vault/HSM Admin** | แยกจากผู้ถือ component | ดำเนินการทางเทคนิค แต่เข้าถึงกุญแจไม่ได้โดยลำพัง |

> **Separation of Duties:** ผู้ที่ถือ key component ต้อง**ไม่ใช่**คนเดียวกับ HSM Admin และต้องลงนามหนังสือรับทราบหน้าที่ key custodian (acknowledgement form) ตาม PCI Req 12.3.1

### 6.3 Key Ceremony (พิธีสร้าง/หมุนกุญแจ)

| ขั้นตอน | รายละเอียด | การควบคุม |
|---|---|---|
| 1. เตรียมการ | จองห้องปลอดภัย, เตรียม HSM, ตรวจ smartcard | Dual Control |
| 2. ยืนยันตัวตน | custodian ≥ 2 คน + witness เข้าห้อง, ลงชื่อ log | MFA + biometric (ถ้ามี) |
| 3. สร้าง/โหลดกุญแจ | HSM สร้าง RK; แจก component ให้ custodian แต่ละคน | Split Knowledge |
| 4. ทดสอบ (KCV) | ตรวจ Key Check Value ยืนยันโหลดถูกต้อง | ทั้งสองฝ่ายยืนยัน |
| 5. จัดเก็บ | เก็บ smartcard/component ในตู้เซฟแยกกัน | tamper-evident bag |
| 6. บันทึก | จัดทำ ceremony log ลงนามทุกคน + witness | เก็บ ≥ 3 ปี / ตาม PCI |

---

## 7. เมทริกซ์บทบาทและสิทธิ์ (RACI สรุป)

| กิจกรรม | CISO | Key Custodian | HSM Admin | Internal Audit | Payment Core |
|---|---|---|---|---|---|
| กำหนดนโยบายกุญแจ | **A/R** | C | C | I | I |
| Key Ceremony | A | **R** (dual) | C | Witness | — |
| Key rotation ปกติ | A | R | R | I | I |
| Emergency rotation | **A** | R | R | I | I |
| Detokenize (runtime) | — | — | — | I (audit) | **R** (มีสิทธิ์เฉพาะ) |
| ทำลายกุญแจ | A | **R** (dual) | R | Witness | — |

(A=Accountable, R=Responsible, C=Consulted, I=Informed)

---

## 8. การควบคุมสนับสนุน (Supporting Controls)

- **Encryption in transit:** TLS 1.2+ ทุกช่องทางภายนอก; mTLS ระหว่าง payment core ↔ Vault (PCI Req 4)
- **Access control:** RBAC + least privilege; MFA บังคับสำหรับผู้เข้าถึง Vault/HSM/KMS (PCI Req 7–8)
- **Logging:** ทุก key operation และ detokenize เขียน `audit_log` แบบ append-only; **ห้าม log PAN/DEK plaintext** (PCI Req 10)
- **Key inventory:** ทะเบียนกุญแจกลาง (key ID, ประเภท, สถานะ, วันสร้าง/หมุน/เพิกถอน, custodian) ทบทวนรายไตรมาส
- **HSM HA/DR:** HSM คู่ (active-active หรือ active-standby) พร้อม secure key backup ข้าม DC ในไทย (data residency) — สอดคล้อง RPO ≤ 5 นาที / RTO ≤ 30 นาที ตาม `ARCHITECTURE.md`
- **การทดสอบ:** ทดสอบขั้นตอน key recovery/DR อย่างน้อยปีละครั้ง; รวมใน scope ASV scan รายไตรมาส + pentest ประจำปี

---

## 9. ข้อสมมติและสิ่งที่ต้องยืนยัน (Assumptions / TODO)

> **[TODO / ข้อสมมติ]** รายการต่อไปนี้ยังขึ้นกับปัจจัยภายนอกที่ยังไม่ยุติ — ระบุเป็นข้อสมมติ ไม่ใช่ข้อเท็จจริงที่ยืนยันแล้ว:
> 1. **ผู้ให้บริการ HSM/KMS จริง** (เช่น cloud KMS + CloudHSM หรือ on-prem HSM แบรนด์ใด) ยังไม่เลือกขั้นสุดท้าย — ต้องยืนยันว่ารองรับ FIPS 140-2 L3/140-3 และ data residency ในไทย
> 2. **QSA vendor** สำหรับ PCI-DSS Level 1 RoC ยังไม่ลงนามสัญญา — พารามิเตอร์ (รอบหมุนกุญแจ, k-of-n) ต้องได้รับการรับรองจาก QSA
> 3. **Sponsor bank / card scheme** (Visa/Mastercard) อาจมีข้อกำหนดเพิ่มเรื่อง token/BIN และ key management ที่ต้อง align
> 4. **ทุนจดทะเบียนชำระแล้ว 50 ล้านบาท** อ้างอิงเกณฑ์ Acquiring ต้องยืนยันสถานะการชำระทุนจริงกับฝ่ายการเงิน/ทะเบียน
> 5. รายละเอียดประกาศ ธปท. ฉบับล่าสุดด้าน IT risk/cyber resilience ต้องยืนยันเลขที่/ปีโดยที่ปรึกษากฎหมาย

---
---

# Tokenization + HSM/KMS key management design: envelope encryption, rotation, dual control, split knowledge (English)

> Document No. **21** in the Compliance document set for the **Full Acquiring** payment service license application under the **Payment Systems Act B.E. 2560 (2017)**, supervised by the **Bank of Thailand (BOT)**, in parallel with **PCI-DSS Level 1** (paid-up registered capital THB 50M).
>
> **Status:** submission draft · v0.1 · last updated 2026-07-22
> **Owner:** Chief Information Security Officer (CISO) and Cryptographic Key Custodian Team of **[บริษัท / Company]**
>
> This is an internal policy/standard and a license-application supporting document; it is **not legal advice** and must be reviewed by legal counsel, the CISO, and the QSA before enforcement, as BOT notifications and PCI-DSS requirements may change.

---

## Executive Summary

[บริษัท / Company] is designed so that **Cardholder Data (PAN) is never stored in the operational database**. PAN is replaced by an **irreversible token** with no mathematical relationship to the PAN, which cannot be reversed without authorized access to the **Tokenization Vault** residing in a network segment isolated from the core system. This achieves **PCI scope minimization**, leaving the core system to hold only `token`, `card_brand`, and `card_last4`, consistent with `ARCHITECTURE.md`.

Encryption uses a **three-tier Envelope Encryption** architecture: Root/Master Key → Key-Encrypting Key (KEK) → Data-Encrypting Key (DEK). The Root/Master Key **never leaves the HSM**, which is certified to **FIPS 140-2 Level 3 (or FIPS 140-3)**. Every sensitive key operation is governed by **Dual Control** and **Split Knowledge**, so no single individual can assemble or use a key alone.

This document defines policy, standards, and SOPs that directly address **PCI-DSS v4.0 Requirement 3 (Protect Stored Account Data)** and **Requirement 4 (Protect CHD in transit with strong cryptography)**, and links to BOT IT risk / cyber-resilience expectations, the Personal Data Protection Act B.E. 2562 (PDPA, supervised by the PDPC), and AML/CDD obligations under AMLO.

> **Architecture link:** this document details the controls referenced in `ARCHITECTURE.md` §7 (Encryption at rest, Key management, Scope minimization) and in `13-it-risk-management.md`.

---

## 1. Legal Basis & Reference Standards

| Law / Standard | Relevance |
|---|---|
| **Payment Systems Act B.E. 2560** — **BOT** | Acquiring license conditions require standard data-security controls and IT risk management |
| **BOT IT Risk & Cyber Resilience guidelines** | Encryption of sensitive data, key management, access control, in-Thailand data residency |
| **PCI-DSS v4.0 — Requirement 3** | Protect stored account data: PAN encryption, key management, dual control, split knowledge, rotation, retirement (3.5–3.7) |
| **PCI-DSS v4.0 — Requirement 4** | Encrypt CHD over public networks with strong cryptography (TLS 1.2+) |
| **PCI-DSS v4.0 — Requirement 8 & 12** | Authentication (MFA), key custodian roles and written acknowledgement |
| **PDPA B.E. 2562** — **PDPC** | Appropriate security measures (§37); breach notification within 72 hours; PAN is personal data |
| **Anti-Money Laundering Act (AMLO)** | Retain detokenization audit trail linked to transactions to support investigation/STR |
| **EMV 3-D Secure (3DS) 2.x** | 3DS device/session tokens and keys fall under the same key-management framework |
| **FIPS 140-2 Level 3 / FIPS 140-3** | HSM cryptographic-module certification |
| **NIST SP 800-57 / SP 800-38 / ISO 11568** | Key lifecycle and key-length principles |

> **[TODO / Assumption]** Confirm the number/year of the latest BOT IT risk & cyber-resilience notification at filing time and align citations, via legal counsel and the QSA.

---

## 2. Scope & Definitions

**Scope:** all cryptographic keys protecting CHD (and SAD, which must never be stored after authorization), the Tokenization Vault, HSM/KMS, the payment core (Go/Fiber), databases, and all backups of [บริษัท / Company].

| Term | Definition |
|---|---|
| **PAN** | Primary Account Number (13–19 digits) |
| **CHD / SAD** | Cardholder Data / Sensitive Authentication Data (CVV2, PIN, full track — **must not be stored**) |
| **Token** | Mathematically irreversible surrogate for PAN, created and mapped only inside the Vault |
| **Detokenization** | Recovering PAN from a token — only inside the CDE via an authenticated API |
| **Root/Master Key** | Top-tier key, generated and used inside the HSM only, never exported in plaintext |
| **KEK** | Key-Encrypting Key — encrypts DEKs |
| **DEK** | Data-Encrypting Key — encrypts PAN directly (AES-256-GCM) |
| **Dual Control** | At least two people are required to perform a sensitive key operation |
| **Split Knowledge** | No single person knows the full key; each holds only one component/share |
| **Key Custodian** | Assigned person holding/managing a key component per PCI Req 3.6/12.3 |
| **CDE** | Cardholder Data Environment (PCI-DSS) |

---

## 3. Tokenization Architecture

### 3.1 Principles

1. **Irreversible tokens** — random, high-entropy tokens mapped to PAN via an encrypted lookup in the vault; not directly reversible from the token.
2. **Segmented vault** — Token Vault and HSM sit in a network zone isolated from the core, with strict firewall/ACLs, reachable only via mTLS + service identity.
3. **Core never touches PAN** — the payment core stores only `token`, `card_brand`, `card_last4`; no full PAN in `payments`, logs, or webhook payloads.
4. **Client-side single-use payment token** — per `ARCHITECTURE.md` §5.1, PAN goes from client directly to the Vault, bypassing the merchant server.

### 3.2 Tokenize / Detokenize Flow

| Step | Action | Actors |
|---|---|---|
| **Tokenize** | client sends PAN → Vault → HSM mints token + encrypts PAN with DEK → stores ciphertext + mapping → returns token | client, Vault, HSM |
| **Store** | core stores only token + last4 + brand | payment core |
| **Detokenize** | only during authorize/settle/chargeback → core calls Vault API (mTLS + JWT scope=`detok`) → Vault authorizes → PAN returned in memory only | payment core (in CDE), Vault |
| **Audit** | every detokenize writes append-only `audit_log` (who/when/token/reason/txn_id) | all |

> **Prohibition:** never detokenize to display full PAN on UI/reports, except where legally required (e.g., AMLO investigation), and only with Dual Control approval and a recorded justification.

---

## 4. Envelope Encryption (three-tier)

### 4.1 Key Hierarchy

```
Root/Master Key (RK)         ← generated & used inside HSM only, never exported in plaintext
        │  (wrap / unwrap)
        ▼
Key-Encrypting Key (KEK)     ← stored RK-wrapped, loaded into HSM at use time
        │  (wrap / unwrap)
        ▼
Data-Encrypting Key (DEK)    ← per-tenant / per-partition, stored KEK-encrypted beside ciphertext
        │  (AES-256-GCM)
        ▼
PAN ciphertext + auth tag    ← stored in Token Vault
```

### 4.2 Cryptographic Parameters

| Item | Standard |
|---|---|
| Data algorithm (DEK) | **AES-256-GCM** (authenticated encryption) |
| Wrap algorithm (KEK/RK) | AES-256-KW or RSA-3072/4096 (per HSM) |
| Minimum key length | Symmetric ≥ 256-bit; asymmetric RSA ≥ 3072-bit / ECC ≥ P-256 |
| Randomness source | HSM TRNG (FIPS-approved) |
| DEK storage | stored **KEK-encrypted** next to ciphertext; never plaintext DEK on disk |
| RK storage | inside HSM only (non-exportable) |

### 4.3 Key Separation

- **Purpose-specific DEKs** — CHD storage, backup, log encryption, and 3DS session each use distinct DEKs.
- **No shared keys across environments** — dev/staging use separate keys and synthetic data (no real PANs in non-prod).

---

## 5. Key Rotation

### 5.1 Scheduled Rotation (Cryptoperiod)

| Key Type | Normal Cadence | Method |
|---|---|---|
| **DEK (CHD)** | at least **every 12 months** or on data-volume threshold | mint new DEK; re-encrypt data (lazy or batch) |
| **KEK** | at least **every 12–24 months** | re-wrap all DEKs with new KEK inside HSM |
| **Root/Master Key** | every **1–2 years** or on event (ceremony) | Key Ceremony under Dual Control + Split Knowledge |
| **TLS/mTLS cert** | ≤ **12 months** (90 days recommended if automated) | rotate via ACME/internal PKI |

### 5.2 Emergency Rotation (Key Compromise)

Rotate immediately when: suspected key leak; a key custodian departs/changes role; algorithm vulnerability; or QSA/BOT directive.

| Step | Target SLA |
|---|---|
| Declare event + convene emergency Key Ceremony | ≤ 4 hours |
| Mint new key + re-wrap/re-encrypt | ≤ 24 hours (data-volume dependent) |
| Retire old key + record | immediately after re-encrypt completes |
| Notify PDPC/BOT (if a data breach) | as required (PDPA: within 72 hours) |

### 5.3 Key Retirement & Destruction

- Rotated-out keys become **retired**, usable only to **decrypt legacy data** until re-encryption completes.
- Once no data references a key → **destroy** (crypto-shredding) inside the HSM and record in inventory.
- Every destruction requires Dual Control and an audit trail.

---

## 6. Dual Control & Split Knowledge

### 6.1 Principles (PCI-DSS Req 3.6 / 3.7 / 12.3)

- **Split Knowledge:** the Root/Master Key (or manually loaded KEK) is split into **3 key components (shares)** via secret sharing (e.g., Shamir **2-of-3**) — no one knows the full key.
- **Dual Control:** assembling/using/destroying a key requires at least **2 key custodians** present together.

### 6.2 Key Custodian Roles

| Role | Owner | Duty |
|---|---|---|
| **Key Custodian A** | Security Engineer (DevSecOps) | holds component / smartcard #1 + PIN |
| **Key Custodian B** | SRE/Infra Lead | holds component / smartcard #2 + PIN |
| **Key Custodian C (backup)** | CISO or delegate | holds component #3 (backup, separate safe) |
| **Ceremony Witness** | Internal Audit / Compliance | observes + signs record; holds no component |
| **Vault/HSM Admin** | separate from component holders | performs technical ops but cannot access keys alone |

> **Separation of Duties:** component holders must **not** be the HSM Admin, and must sign the key-custodian acknowledgement form per PCI Req 12.3.1.

### 6.3 Key Ceremony

| Step | Detail | Control |
|---|---|---|
| 1. Prepare | reserve secure room, ready HSM, check smartcards | Dual Control |
| 2. Authenticate | ≥ 2 custodians + witness enter, sign log | MFA + biometric (if available) |
| 3. Generate/load | HSM generates RK; distribute components | Split Knowledge |
| 4. Verify (KCV) | check Key Check Value confirms correct load | both parties confirm |
| 5. Store | components/smartcards in separate safes | tamper-evident bags |
| 6. Record | ceremony log signed by all + witness | retain ≥ 3 years / per PCI |

---

## 7. Role & Authority Matrix (RACI summary)

| Activity | CISO | Key Custodian | HSM Admin | Internal Audit | Payment Core |
|---|---|---|---|---|---|
| Define key policy | **A/R** | C | C | I | I |
| Key Ceremony | A | **R** (dual) | C | Witness | — |
| Scheduled rotation | A | R | R | I | I |
| Emergency rotation | **A** | R | R | I | I |
| Detokenize (runtime) | — | — | — | I (audit) | **R** (scoped) |
| Key destruction | A | **R** (dual) | R | Witness | — |

(A=Accountable, R=Responsible, C=Consulted, I=Informed)

---

## 8. Supporting Controls

- **Encryption in transit:** TLS 1.2+ on all external channels; mTLS between payment core ↔ Vault (PCI Req 4).
- **Access control:** RBAC + least privilege; MFA mandatory for Vault/HSM/KMS access (PCI Req 7–8).
- **Logging:** every key operation and detokenize writes append-only `audit_log`; **never log plaintext PAN/DEK** (PCI Req 10).
- **Key inventory:** central register (key ID, type, status, created/rotated/retired dates, custodian), reviewed quarterly.
- **HSM HA/DR:** dual HSM (active-active or active-standby) with secure key backup across in-Thailand DCs (data residency) — supporting RPO ≤ 5 min / RTO ≤ 30 min per `ARCHITECTURE.md`.
- **Testing:** exercise key recovery/DR at least annually; include in quarterly ASV scan + annual pentest scope.

---

## 9. Assumptions / TODO

> **[TODO / Assumption]** The following remain dependent on unresolved external factors — stated as assumptions, not confirmed facts:
> 1. **Actual HSM/KMS provider** (cloud KMS + CloudHSM, or on-prem HSM brand) is not finalized — confirm FIPS 140-2 L3/140-3 support and in-Thailand data residency.
> 2. **QSA vendor** for PCI-DSS Level 1 RoC is not yet contracted — parameters (rotation cadence, k-of-n) must be endorsed by the QSA.
> 3. **Sponsor bank / card scheme** (Visa/Mastercard) may impose additional token/BIN and key-management requirements to align.
> 4. **Paid-up capital of THB 50M** reflects the Acquiring threshold — confirm actual capital status with Finance/registration.
> 5. The exact latest BOT IT-risk/cyber-resilience notification number/year must be confirmed by legal counsel.
