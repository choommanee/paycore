# นโยบายความเป็นส่วนตัว (PDPA) (ไทย)

> เอกสารประกอบการยื่นขอใบอนุญาต **การให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (Full Acquiring)**
> ภายใต้ พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 ต่อธนาคารแห่งประเทศไทย (ธปท.) และเป็นเอกสารประกอบการประเมิน PCI-DSS v4.0 Level 1
>
> เอกสารเลขที่: `COMP-09` · เวอร์ชัน 1.0 · เจ้าของเอกสาร: Data Protection Officer (DPO) / Compliance
> เอกสารอ้างอิง: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`
>
> **หมายเหตุ:** เอกสารนี้เป็นเอกสารเชิงนโยบายภายในและฉบับเผยแพร่ต่อเจ้าของข้อมูล ไม่ใช่คำแนะนำทางกฎหมาย ควรผ่านการทบทวนโดยที่ปรึกษากฎหมายก่อนเผยแพร่/ยื่นจริง

---

> ### ⚠️ ข้อสมมติและสิ่งที่ยังต้องยืนยัน (Assumptions / TODO)
> รายการต่อไปนี้ยังขึ้นกับคู่สัญญา/ผู้ให้บริการภายนอกที่ยังไม่สรุป — ห้ามถือเป็นข้อเท็จจริงจนกว่าจะยืนยัน:
> - **[TODO — Sponsor Bank / Acquirer]** ยังไม่ลงนามธนาคารผู้รับเชื่อม (sponsoring bank) และข้อตกลงแบ่งบทบาท controller/processor ระหว่างกัน — ส่งผลต่อฐานการประมวลผลและภาระการแจ้งเหตุละเมิดร่วม
> - **[TODO — QSA]** ยังไม่เลือกผู้ประเมิน PCI-DSS (Qualified Security Assessor) และผู้ให้บริการ ASV — ขอบเขต scope และ vault design ต้องผ่าน QSA
> - **[TODO — ทุนจดทะเบียน]** ทุนจดทะเบียนชำระแล้วเป้าหมาย **50 ล้านบาท** (Full Acquiring) — ต้องยืนยันจำนวนที่ชำระจริงและรักษาไว้ ≥ 75% ตลอดการดำเนินงาน
> - **[TODO — Cross-border sub-processor]** รายชื่อผู้ประมวลผลย่อยในต่างประเทศ (เช่น cloud region, 3DS provider, ผู้ให้บริการ fraud) ต้องสรุปและตรวจสอบมาตรการคุ้มครองตามมาตรา 28/29 PDPA
> - **[TODO — ชื่อบริษัท/ที่อยู่/DPO]** ชื่อ `[บริษัท / Company]`, ที่อยู่จดทะเบียน, ชื่อและช่องทางติดต่อ DPO ต้องเติมค่าจริงก่อนเผยแพร่

---

## 1. บทนำและขอบเขต

`[บริษัท / Company]` (ต่อไปนี้เรียก "บริษัท" หรือ "เรา") ประกอบธุรกิจให้บริการรับชำระเงินด้วยวิธีการทางอิเล็กทรอนิกส์ (payment gateway / acquiring) เราให้ความสำคัญสูงสุดต่อการคุ้มครองข้อมูลส่วนบุคคลตาม **พ.ร.บ. คุ้มครองข้อมูลส่วนบุคคล พ.ศ. 2562 (PDPA)** และประกาศของคณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC) รวมถึงหลักเกณฑ์ของ ธปท. และมาตรฐาน **PCI-DSS v4.0**

นโยบายนี้ครอบคลุมการประมวลผลข้อมูลส่วนบุคคลของ:
- **ผู้ถือบัตร (cardholders)** ที่ทำรายการผ่านร้านค้าที่ใช้บริการของเรา
- **ผู้ค้า/ร้านค้า (merchants)** และผู้ติดต่อ/ผู้มีอำนาจของร้านค้า
- **ผู้เข้าเว็บไซต์ ผู้สมัคร และผู้ติดต่ออื่น ๆ**

### บทบาทตามกฎหมาย (Controller / Processor)
| สถานการณ์ | บทบาทของบริษัท |
|-----------|----------------|
| ข้อมูลผู้ถือบัตรที่ประมวลผลตามคำสั่งร้านค้าเพื่อทำรายการชำระเงิน | โดยหลักเป็น **ผู้ควบคุมข้อมูลร่วม/ผู้ควบคุมข้อมูลของตนเอง** สำหรับวัตถุประสงค์ป้องกันทุจริต, AML และหน้าที่ตามกฎหมาย — และเป็น **ผู้ประมวลผล** เมื่อทำตามคำสั่งร้านค้าล้วน ๆ (ต้องมีข้อตกลง DPA มาตรา 40) **[TODO ยืนยันโครงสร้างกับ sponsor bank]** |
| ข้อมูลผู้ติดต่อของร้านค้า, KYC ของร้านค้า | **ผู้ควบคุมข้อมูล (Controller)** |
| ข้อมูลพนักงาน/ผู้สมัครงาน | **ผู้ควบคุมข้อมูล (Controller)** |

---

## 2. ฐานทางกฎหมายในการประมวลผล (Lawful Basis)

เราประมวลผลข้อมูลส่วนบุคคลบนฐานทางกฎหมายที่ชัดเจน โดย **ไม่พึ่งพาความยินยอมเป็นฐานหลัก** สำหรับกิจกรรมที่จำเป็นต่อการให้บริการชำระเงินและการปฏิบัติตามกฎหมาย (เพื่อหลีกเลี่ยงการถอนความยินยอมที่กระทบการปฏิบัติหน้าที่ตามกฎหมาย)

| กิจกรรมการประมวลผล | ประเภทข้อมูล | ฐานทางกฎหมาย (PDPA) |
|--------------------|-------------|---------------------|
| ประมวลผลรายการชำระเงิน (authorize/capture/refund/void) | token บัตร, `card_last4`, `card_brand`, จำนวนเงิน, merchant, timestamp | **มาตรา 24(3) – ความจำเป็นเพื่อปฏิบัติตามสัญญา** ที่ร้านค้า/ผู้ถือบัตรเป็นคู่สัญญา |
| ป้องกันการทุจริต, risk scoring, velocity check, blacklist, 3DS 2.x | device/IP, ประวัติธุรกรรม, ผลการยืนยันตัวตน | **มาตรา 24(5) – ประโยชน์โดยชอบด้วยกฎหมาย (legitimate interest)** พร้อมทำ LIA (Legitimate Interest Assessment) |
| KYC/CDD ร้านค้า, sanction screening, การรายงานธุรกรรมที่มีเหตุอันควรสงสัย | ชื่อ, เลขบัตรประชาชน, เอกสารจดทะเบียน, ผู้รับผลประโยชน์ | **มาตรา 24(6) – ปฏิบัติตามกฎหมาย** (พ.ร.บ. ปปง./AMLO, กฎ ธปท.) |
| การเก็บ audit log, การเก็บรักษาเอกสารตามกฎหมาย | log, บันทึกธุรกรรม | **มาตรา 24(6) – ปฏิบัติตามกฎหมาย** และ **24(5)** |
| การตลาด/ข่าวสารผลิตภัณฑ์ถึงร้านค้า (ที่ไม่จำเป็นต่อบริการ) | อีเมล, ชื่อผู้ติดต่อ | **มาตรา 19 – ความยินยอม (consent)** ที่ถอนได้ทุกเมื่อ |
| ข้อมูลอ่อนไหว (เช่น ข้อมูลชีวมิติ/ใบหน้าใน KYC ถ้ามี) | biometric | **มาตรา 26 – ความยินยอมโดยชัดแจ้ง (explicit consent)** หรือข้อยกเว้นตามมาตรา 26 เท่านั้น |

> **หลักการ:** สำหรับข้อมูลผู้ถือบัตรที่จำเป็นต่อการทำรายการ ฐานหลักคือ "สัญญา" และ "กฎหมาย/legitimate interest" — ไม่ใช้ consent เพื่อไม่ให้การถอนความยินยอมทำให้เราผิดหน้าที่ตาม AML/ธปท.

---

## 3. ประเภทข้อมูลที่เก็บ และหลักการลดข้อมูลบัตร (Cardholder Data Minimization)

### 3.1 หลักการลดขอบเขตข้อมูลบัตร (สอดคล้อง ARCHITECTURE §2, §6, §7 และ PCI-DSS v4.0)
เราออกแบบระบบให้ **นำ cardholder data ออกจาก scope ให้มากที่สุด**:

- ใช้ **client-side tokenization**: PAN ถูกส่งตรงจาก client ไปยัง Tokenization Vault (network segment แยกใน PCI scope) โดย **ไม่ผ่านเซิร์ฟเวอร์ของร้านค้าและไม่ผ่านระบบหลักของเรา**
- ระบบหลัก (operational DB) เห็นเพียง **token + `card_last4` + `card_brand`** เท่านั้น
- PAN ที่จำเป็นต้องเก็บ เข้ารหัสด้วย **envelope encryption** โดยคีย์อยู่ใน **HSM/KMS** (PCI Req 3) มี key rotation, dual control, split knowledge

### 3.2 ข้อมูลที่ห้ามจัดเก็บโดยเด็ดขาด (PCI-DSS – Sensitive Authentication Data)
ตาม ARCHITECTURE §6 **ห้าม** จัดเก็บใน operational DB:
- **Full PAN** (จัดเก็บเฉพาะใน vault ที่เข้ารหัสใน PCI scope แยก / ปกติเก็บเป็น token เท่านั้น)
- **CVV / CVV2 / CVC2** — ห้ามเก็บหลัง authorization ทุกกรณี
- **PIN / PIN block**
- **Full magnetic stripe / track data**

### 3.3 ตารางข้อมูลที่เก็บ (Data Inventory ระดับสรุป)
| กลุ่ม | ตัวอย่างข้อมูล | ที่จัดเก็บ | สถานะ scope |
|-------|---------------|-----------|-------------|
| ข้อมูลบัตรแบบ tokenized | token, `card_last4`, `card_brand` | operational DB (ในไทย) | นอก PCI scope หลัก |
| PAN (ถ้าจำเป็น) | PAN เข้ารหัส | Tokenization Vault (segment แยก) | ใน PCI scope |
| ข้อมูลธุรกรรม | จำนวนเงิน (minor units), สถานะ, `auth_code`, timestamp | ledger / payments (append-only) | นอก PCI |
| ข้อมูลร้านค้า/KYC | ชื่อ, เลขนิติบุคคล, ผู้มีอำนาจ, เอกสาร | merchant store (เข้ารหัส at rest) | นอก PCI, ใน PDPA/AML |
| Risk/fraud | IP, device fingerprint, ผล 3DS | risk store | legitimate interest |
| Audit | audit_log (append-only) | audit store | บังคับตาม PCI Req 10 / ธปท. |

---

## 4. สิทธิของเจ้าของข้อมูล (Data Subject Rights)

เจ้าของข้อมูลมีสิทธิดังต่อไปนี้ ภายใต้เงื่อนไขและข้อยกเว้นตาม PDPA:

| สิทธิ | มาตรา | หมายเหตุ/ข้อจำกัด |
|-------|-------|-------------------|
| สิทธิได้รับแจ้ง (Right to be informed) | ม.23 | ผ่านนโยบายฉบับนี้และ notice ณ จุดเก็บ |
| สิทธิเข้าถึงและขอสำเนา (Access) | ม.30 | ตอบภายใน 30 วัน |
| สิทธิขอโอนย้ายข้อมูล (Portability) | ม.31 | เฉพาะข้อมูลที่ประมวลผลด้วยระบบอัตโนมัติบนฐาน consent/สัญญา |
| สิทธิคัดค้าน (Object) | ม.32 | เช่น คัดค้านการตลาดตรง — ดำเนินการทันที |
| สิทธิให้ลบ/ทำลาย (Erasure) | ม.33 | **มีข้อจำกัด** — ปฏิเสธได้หากต้องเก็บตามกฎหมาย AML/ธปท. |
| สิทธิให้ระงับการใช้ (Restriction) | ม.34 | ระงับระหว่างตรวจสอบความถูกต้อง |
| สิทธิขอแก้ไขให้ถูกต้อง (Rectification) | ม.35–36 | ให้ข้อมูลถูกต้อง เป็นปัจจุบัน |
| สิทธิถอนความยินยอม (Withdraw consent) | ม.19 | เฉพาะกิจกรรมที่ใช้ฐาน consent (เช่น การตลาด) ไม่กระทบฐานอื่น |
| สิทธิร้องเรียน (Complain) | ม.73 | ต่อบริษัทและต่อ PDPC |

### 4.1 ขั้นตอนและระยะเวลาการใช้สิทธิ (DSAR Procedure)
1. ยื่นคำขอผ่าน `privacy@[บริษัท / Company]` หรือแบบฟอร์มบนเว็บไซต์ **[TODO ยืนยันช่องทาง]**
2. **ยืนยันตัวตน** ผู้ยื่น (ป้องกันการเปิดเผยข้อมูลผิดคน) ภายใน 3 วันทำการ
3. **ดำเนินการและตอบกลับภายใน 30 วัน** นับแต่ได้รับคำขอที่สมบูรณ์ (ขยายได้เมื่อมีเหตุอันควรพร้อมแจ้งเหตุผล)
4. หากปฏิเสธ ต้อง **บันทึกเหตุผลและแจ้งสิทธิร้องเรียนต่อ PDPC**
5. บันทึกคำขอทั้งหมดใน DSAR register (auditable)

---

## 5. การส่ง/โอนข้อมูลไปต่างประเทศ (Cross-Border Transfer)

หลักการ: ตาม ARCHITECTURE §8 เรากำหนด **data residency ในประเทศไทย** สำหรับข้อมูลธุรกรรมและข้อมูลผู้ถือบัตรเป็นค่าตั้งต้น เพื่อสอดคล้องกับข้อกำหนด ธปท./PDPA

การโอนไปต่างประเทศจะเกิดเฉพาะเมื่อจำเป็น (เช่น 3DS directory server, card scheme, cloud sub-processor) และต้องเป็นไปตาม **มาตรา 28/29 PDPA**:

| กลไก | ใช้เมื่อ |
|------|---------|
| ประเทศปลายทางมีมาตรฐานคุ้มครองเพียงพอ (adequacy ที่ PDPC ประกาศ) | ม.28 |
| มี **มาตรการคุ้มครองที่เหมาะสม** เช่น BCR / SCC / ข้อสัญญาที่บังคับได้ (Data Transfer Agreement) | ม.29 |
| ความจำเป็นตามสัญญากับเจ้าของข้อมูล / เพื่อประโยชน์สำคัญ | ข้อยกเว้น ม.28 |

**การควบคุมเพิ่มเติม:** ข้อมูลบัตรที่ออกนอกประเทศต้องเป็น token หรือถูกเข้ารหัสเสมอ; card scheme/3DS โดยธรรมชาติของเครือข่ายอาจประมวลผลข้ามพรมแดน — ต้องระบุใน RoPA และ DPA

> **[TODO — Cross-border]** สรุปรายชื่อ sub-processor ต่างประเทศ, region ของ cloud, และกลไกมาตรา 29 ที่ใช้กับแต่ละราย ก่อนยื่น ธปท.

---

## 6. ระยะเวลาเก็บรักษาข้อมูล (Data Retention)

| ประเภทข้อมูล | ระยะเวลาเก็บ | เหตุผล |
|-------------|-------------|--------|
| บันทึกธุรกรรม/ledger, audit_log | อย่างน้อย **5 ปี** หลังธุรกรรม (สอดคล้อง AML/กฎหมายบัญชี) | ปฏิบัติตามกฎหมาย |
| ข้อมูล KYC/CDD ร้านค้า | อย่างน้อย **5 ปี** หลังสิ้นสุดความสัมพันธ์ | พ.ร.บ. ปปง. |
| PAN/token ในบัตร (สำหรับ recurring) | เท่าที่จำเป็นต่อบริการ / จนถอนหรือสิ้นสัญญา | สัญญา + minimization |
| CVV/CVC | **ห้ามเก็บ** (ลบทันทีหลัง authorization) | PCI-DSS |
| Log ความปลอดภัย (SIEM) | อย่างน้อย **1 ปี** (3 เดือนพร้อมเรียกดูทันที) | PCI Req 10 |
| ข้อมูลการตลาด (consent) | จนกว่าจะถอนความยินยอม | ม.19 |

เมื่อพ้นกำหนด เราจะ **ลบ/ทำให้ไม่สามารถระบุตัวตน (anonymize)** ตามขั้นตอนที่บันทึกได้

---

## 7. มาตรการรักษาความมั่นคงปลอดภัย (Security Measures)

สอดคล้องกับ ARCHITECTURE §7 และ PCI-DSS v4.0:
- **เข้ารหัสระหว่างส่ง**: TLS 1.2+ และ mTLS ระหว่างบริการภายใน
- **เข้ารหัสขณะพัก**: envelope encryption, คีย์ใน HSM/KMS, ไม่มีคีย์ในโค้ด/คอนฟิก
- **การควบคุมการเข้าถึง**: least privilege, RBAC, MFA สำหรับผู้ดูแล (PCI Req 7–8)
- **การแยกส่วนเครือข่าย**: vault/tokenization แยก segment; WAF, IDS/IPS (PCI Req 1, 11)
- **บันทึกและเฝ้าระวัง**: structured log ที่ **ไม่มี card data**, ห้าม log request body, SIEM + alert (PCI Req 10)
- **บริหารช่องโหว่**: `govulncheck`, SAST/DAST ใน CI, **quarterly ASV scan**, **annual penetration test**
- **3DS 2.x / EMV 3DS** เพื่อยืนยันตัวตนผู้ถือบัตรและลดความรับผิด

---

## 8. การแจ้งเหตุละเมิดข้อมูลส่วนบุคคล (Data Breach Notification)

| ขั้นตอน | กำหนดเวลา |
|--------|-----------|
| ตรวจพบ → ประเมินความเสี่ยง (severity/scope) | ทันที (activate incident response) |
| แจ้ง **PDPC** เมื่อมีความเสี่ยงต่อสิทธิเสรีภาพบุคคล | **ภายใน 72 ชั่วโมง** นับแต่ทราบเหตุ (ม.37(4)) |
| แจ้ง **เจ้าของข้อมูล** เมื่อความเสี่ยงสูง | โดยไม่ชักช้า |
| แจ้ง **ธปท.** ตามหลักเกณฑ์ cyber/IT incident | ตามที่ประกาศกำหนด |
| แจ้ง card scheme / acquirer กรณีข้อมูลบัตรรั่ว | ตาม PCI / สัญญา **[TODO ยืนยันกับ sponsor bank]** |

ทุกเหตุการณ์บันทึกใน breach register พร้อมมาตรการเยียวยา

---

## 9. บทบาทและความรับผิดชอบ (Roles)

| บทบาท | หน้าที่ |
|-------|--------|
| **DPO (เจ้าหน้าที่คุ้มครองข้อมูล)** | กำกับการปฏิบัติตาม PDPA, จุดติดต่อ PDPC/เจ้าของข้อมูล, ทบทวน RoPA/LIA/DPA **[TODO ระบุชื่อ]** |
| **Compliance / Legal** | ใบอนุญาต ธปท., AML, ทบทวนนโยบาย |
| **Security / DevSecOps** | PCI, HSM/KMS, network segmentation, breach response |
| **Merchant/Product** | onboarding ตาม data minimization, จัดการ consent |

---

## 10. การติดต่อและการทบทวน

- ติดต่อเรื่องความเป็นส่วนตัว/ใช้สิทธิ: `privacy@[บริษัท / Company]` (DPO) **[TODO เติมช่องทาง/ที่อยู่จริง]**
- ร้องเรียนต่อหน่วยงานกำกับ: สำนักงานคณะกรรมการคุ้มครองข้อมูลส่วนบุคคล (PDPC)
- นโยบายนี้ทบทวนอย่างน้อย **ปีละครั้ง** หรือเมื่อมีการเปลี่ยนแปลงกฎหมาย/บริการที่มีนัยสำคัญ

---
---

# PDPA privacy policy: lawful basis, data subject rights, cross-border transfer, cardholder data minimization (English)

> Supporting document for the license application for **Electronic Payment Acquiring Service (Full Acquiring)** under the Payment Systems Act B.E. 2560, submitted to the Bank of Thailand (BOT), and a supporting artefact for the **PCI-DSS v4.0 Level 1** assessment.
>
> Doc ID: `COMP-09` · Version 1.0 · Owner: Data Protection Officer (DPO) / Compliance
> Related: `COMPLIANCE-TH.md`, `ARCHITECTURE.md`, `ROADMAP.md`
>
> **Note:** This is an internal policy and public-facing notice, not legal advice. It must be reviewed by legal counsel before publication/submission.

---

> ### ⚠️ Assumptions / TODO (unresolved external dependencies)
> The following depend on counterparties/vendors not yet finalized — do not treat as fact until confirmed:
> - **[TODO — Sponsor Bank / Acquirer]** Sponsoring bank not yet signed; the controller/processor split and joint breach-notification duties are undetermined — this affects lawful basis and shared obligations.
> - **[TODO — QSA]** PCI-DSS Qualified Security Assessor and ASV vendor not yet selected — scope and vault design require QSA sign-off.
> - **[TODO — Registered capital]** Target paid-up capital **THB 50M** (Full Acquiring); actual paid-up amount and the requirement to maintain ≥ 75% throughout operations must be confirmed.
> - **[TODO — Cross-border sub-processors]** The list of overseas sub-processors (cloud region, 3DS provider, fraud vendor) and their PDPA s.28/29 safeguards must be finalized.
> - **[TODO — Company/address/DPO]** `[บริษัท / Company]`, registered address, and DPO name/contact must be filled in before publication.

---

## 1. Introduction & Scope

`[บริษัท / Company]` ("we", "us") operates an electronic payment acquiring service (payment gateway). We are committed to protecting personal data in accordance with the **Personal Data Protection Act B.E. 2562 (PDPA)** and PDPC guidance, together with BOT rules and **PCI-DSS v4.0**.

This policy covers processing of personal data of:
- **Cardholders** transacting through merchants that use our service
- **Merchants** and their authorized representatives/contacts
- **Website visitors, applicants, and other contacts**

### Legal role (Controller / Processor)
| Situation | Our role |
|-----------|----------|
| Cardholder data processed on a merchant's instruction to complete a payment | Primarily **controller/joint-controller** for fraud prevention, AML and legal duties; **processor** where acting purely on the merchant's instructions (requires a s.40 DPA). **[TODO confirm structure with sponsor bank]** |
| Merchant contact data, merchant KYC | **Controller** |
| Employee/applicant data | **Controller** |

---

## 2. Lawful Basis for Processing

We rely on clear lawful bases and **do not use consent as the primary basis** for processing that is necessary to deliver the payment service and comply with law (so that withdrawal of consent cannot compromise our legal duties).

| Processing activity | Data | Lawful basis (PDPA) |
|---------------------|------|---------------------|
| Payment processing (authorize/capture/refund/void) | card token, `card_last4`, `card_brand`, amount, merchant, timestamp | **s.24(3) – necessary for contract performance** with the merchant/cardholder |
| Fraud prevention, risk scoring, velocity checks, blacklist, 3DS 2.x | device/IP, transaction history, authentication result | **s.24(5) – legitimate interest**, supported by a documented LIA |
| Merchant KYC/CDD, sanction screening, suspicious-transaction reporting | name, national ID, registration docs, beneficial owner | **s.24(6) – legal obligation** (AML Act / AMLO, BOT rules) |
| Audit logging and statutory record retention | logs, transaction records | **s.24(6) – legal obligation** and **s.24(5)** |
| Marketing/product updates to merchants (non-essential) | email, contact name | **s.19 – consent**, withdrawable at any time |
| Sensitive data (e.g. biometrics/face in KYC, if any) | biometric | **s.26 – explicit consent** or a s.26 exemption only |

> **Principle:** For cardholder data necessary to complete a transaction, the primary bases are "contract" and "legal obligation / legitimate interest" — not consent — so that withdrawal cannot force us to breach AML/BOT duties.

---

## 3. Data Collected & Cardholder Data Minimization

### 3.1 Minimization principle (aligned with ARCHITECTURE §2, §6, §7 and PCI-DSS v4.0)
The system is designed to **take cardholder data out of scope as much as possible**:

- **Client-side tokenization**: the PAN is sent directly from the client to the Tokenization Vault (a separate PCI network segment) and **never touches the merchant's server or our core system**.
- The core (operational DB) sees only **token + `card_last4` + `card_brand`**.
- Any PAN that must be retained is protected by **envelope encryption**, with keys in an **HSM/KMS** (PCI Req 3), plus key rotation, dual control and split knowledge.

### 3.2 Data never stored (PCI-DSS – Sensitive Authentication Data)
Per ARCHITECTURE §6, the operational DB **must never** store:
- **Full PAN** (kept only encrypted inside the vault in a separate PCI scope; normally only a token is retained)
- **CVV / CVV2 / CVC2** — never stored after authorization
- **PIN / PIN block**
- **Full magnetic stripe / track data**

### 3.3 Data inventory (summary)
| Group | Example data | Storage | Scope status |
|-------|--------------|---------|--------------|
| Tokenized card data | token, `card_last4`, `card_brand` | operational DB (in Thailand) | outside core PCI scope |
| PAN (if required) | encrypted PAN | Tokenization Vault (separate segment) | inside PCI scope |
| Transaction data | amount (minor units), status, `auth_code`, timestamp | ledger / payments (append-only) | outside PCI |
| Merchant/KYC | name, entity ID, authorized persons, documents | merchant store (encrypted at rest) | outside PCI, within PDPA/AML |
| Risk/fraud | IP, device fingerprint, 3DS result | risk store | legitimate interest |
| Audit | audit_log (append-only) | audit store | mandatory per PCI Req 10 / BOT |

---

## 4. Data Subject Rights

Subject to PDPA conditions and exemptions, data subjects have the following rights:

| Right | Section | Notes / limits |
|-------|---------|----------------|
| To be informed | s.23 | Via this policy and point-of-collection notices |
| Access & copy | s.30 | Responded within 30 days |
| Data portability | s.31 | Only for automated processing on consent/contract basis |
| Object | s.32 | e.g. object to direct marketing — actioned promptly |
| Erasure | s.33 | **Limited** — may be refused where retention is required by AML/BOT law |
| Restriction | s.34 | Suspended while accuracy is verified |
| Rectification | s.35–36 | Keep data accurate and current |
| Withdraw consent | s.19 | Only for consent-based activities (e.g. marketing); does not affect other bases |
| Complaint | s.73 | To us and to the PDPC |

### 4.1 DSAR procedure & timelines
1. Submit via `privacy@[บริษัท / Company]` or a web form **[TODO confirm channel]**.
2. **Verify the requester's identity** (to avoid wrong-person disclosure) within 3 business days.
3. **Action and respond within 30 days** of a complete request (extendable with justified notice).
4. If refused, **record the reason and inform the right to complain to the PDPC**.
5. Log all requests in an auditable DSAR register.

---

## 5. Cross-Border Transfer

Principle: per ARCHITECTURE §8 we set **data residency in Thailand** by default for transaction and cardholder data, aligned with BOT/PDPA requirements.

Cross-border transfers occur only where necessary (e.g. 3DS directory server, card scheme, cloud sub-processor) and must comply with **PDPA s.28/29**:

| Mechanism | Used when |
|-----------|-----------|
| Destination has adequate protection (PDPC adequacy) | s.28 |
| **Appropriate safeguards** such as BCR / SCC / an enforceable data transfer agreement | s.29 |
| Necessity for a contract with the data subject / substantial interest | s.28 exemption |

**Additional controls:** card data leaving the country must always be tokenized or encrypted; card scheme/3DS networks inherently process cross-border by design — this must be documented in the RoPA and DPA.

> **[TODO — Cross-border]** Finalize the list of overseas sub-processors, cloud regions, and the s.29 mechanism used for each before BOT submission.

---

## 6. Data Retention

| Data type | Retention | Reason |
|-----------|-----------|--------|
| Transaction/ledger records, audit_log | at least **5 years** after the transaction (AML/accounting) | legal obligation |
| Merchant KYC/CDD | at least **5 years** after relationship ends | AML Act |
| PAN/token (for recurring) | only as long as needed / until withdrawal or contract end | contract + minimization |
| CVV/CVC | **never stored** (deleted immediately after authorization) | PCI-DSS |
| Security logs (SIEM) | at least **1 year** (3 months immediately available) | PCI Req 10 |
| Marketing data (consent) | until consent is withdrawn | s.19 |

On expiry we delete or anonymize the data via a documented, auditable procedure.

---

## 7. Security Measures

Aligned with ARCHITECTURE §7 and PCI-DSS v4.0:
- **Encryption in transit**: TLS 1.2+, mTLS between internal services
- **Encryption at rest**: envelope encryption, keys in HSM/KMS, no keys in code/config
- **Access control**: least privilege, RBAC, MFA for admins (PCI Req 7–8)
- **Network segmentation**: vault/tokenization on a separate segment; WAF, IDS/IPS (PCI Req 1, 11)
- **Logging & monitoring**: structured logs with **no card data**, no request-body logging, SIEM + alerting (PCI Req 10)
- **Vulnerability management**: `govulncheck`, SAST/DAST in CI, **quarterly ASV scan**, **annual penetration test**
- **3DS 2.x / EMV 3DS** for cardholder authentication and liability reduction

---

## 8. Data Breach Notification

| Step | Timeline |
|------|----------|
| Detect → assess risk (severity/scope) | immediately (activate incident response) |
| Notify **PDPC** where there is a risk to rights and freedoms | **within 72 hours** of becoming aware (s.37(4)) |
| Notify **data subjects** where high risk | without undue delay |
| Notify **BOT** under cyber/IT incident rules | as prescribed |
| Notify card scheme / acquirer for card data compromise | per PCI / contract **[TODO confirm with sponsor bank]** |

Every incident is recorded in a breach register with remediation measures.

---

## 9. Roles & Responsibilities

| Role | Responsibility |
|------|----------------|
| **DPO** | Oversee PDPA compliance; contact point for PDPC/data subjects; maintain RoPA/LIA/DPA **[TODO name]** |
| **Compliance / Legal** | BOT license, AML, policy review |
| **Security / DevSecOps** | PCI, HSM/KMS, network segmentation, breach response |
| **Merchant/Product** | Onboarding by data minimization, consent management |

---

## 10. Contact & Review

- Privacy / rights requests: `privacy@[บริษัท / Company]` (DPO) **[TODO real channel/address]**
- Regulator complaints: Office of the Personal Data Protection Committee (PDPC)
- This policy is reviewed at least **annually**, or on any material change in law or service.
